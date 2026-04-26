package models

import (
	"time"

	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

const (
	TaskTimeTrackingTimerStatusRunning = "running"
	TaskTimeTrackingTimerStatusSnoozed = "snoozed"
	TaskTimeTrackingTimerMaxDuration   = 8 * time.Hour
)

type TaskTimeTrackingTimer struct {
	ID        int64      `xorm:"bigint autoincr not null unique pk" json:"id"`
	TaskID    int64      `xorm:"bigint not null index" json:"task_id"`
	Task      *Task      `xorm:"-" json:"task,omitempty"`
	UserID    int64      `xorm:"bigint not null unique index 'user_id'" json:"-"`
	User      *user.User `xorm:"-" json:"user,omitempty"`
	StartedAt time.Time  `xorm:"datetime not null 'started_at'" json:"started_at"`
	StoppedAt time.Time  `xorm:"datetime null 'stopped_at'" json:"stopped_at"`
	Status    string     `xorm:"varchar(20) not null default 'running'" json:"status"`
}

func (*TaskTimeTrackingTimer) TableName() string {
	return "task_time_tracking_timers"
}

func getCurrentTaskTimeTrackingTimer(s *xorm.Session, userID int64) (*TaskTimeTrackingTimer, error) {
	timer := &TaskTimeTrackingTimer{}
	exists, err := s.Where("user_id = ?", userID).NoAutoCondition().Get(timer)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	task, err := GetTaskByIDSimple(s, timer.TaskID)
	if err != nil {
		return nil, err
	}
	timer.Task = &task
	return timer, nil
}

func GetCurrentTaskTimeTrackingTimer(s *xorm.Session, a web.Auth) (*TaskTimeTrackingTimer, error) {
	if _, is := a.(*LinkSharing); is {
		return nil, ErrGenericForbidden{}
	}
	return getCurrentTaskTimeTrackingTimer(s, a.GetID())
}

func StartTaskTimeTrackingTimer(s *xorm.Session, a web.Auth, taskID int64) (*TaskTimeTrackingTimer, error) {
	task := &Task{ID: taskID}
	canWrite, err := task.CanWrite(s, a)
	if err != nil {
		return nil, err
	}
	if !canWrite {
		return nil, ErrGenericForbidden{}
	}
	if _, err = GetTaskSimple(s, task); err != nil {
		return nil, err
	}

	current, err := GetCurrentTaskTimeTrackingTimer(s, a)
	if err != nil {
		return nil, err
	}
	if current != nil && current.Status == TaskTimeTrackingTimerStatusRunning && current.TaskID != taskID {
		return nil, ErrTaskTimeTrackingTimerAlreadyRunning{TaskID: current.TaskID}
	}
	if current != nil && current.Status == TaskTimeTrackingTimerStatusRunning && current.TaskID == taskID {
		return current, nil
	}
	if current != nil {
		if _, err := s.ID(current.ID).Delete(&TaskTimeTrackingTimer{}); err != nil {
			return nil, err
		}
	}

	timer := &TaskTimeTrackingTimer{
		TaskID:    taskID,
		UserID:    a.GetID(),
		StartedAt: time.Now(),
		Status:    TaskTimeTrackingTimerStatusRunning,
	}
	if _, err := s.Insert(timer); err != nil {
		return nil, err
	}

	timer.Task = task
	return timer, triggerTaskUpdatedEventForTaskID(s, a, taskID, "Time tracking timer started")
}

func StopTaskTimeTrackingTimer(s *xorm.Session, a web.Auth, taskID int64) (*TaskTimeTracking, error) {
	current, err := GetCurrentTaskTimeTrackingTimer(s, a)
	if err != nil {
		return nil, err
	}
	if current == nil || current.Status != TaskTimeTrackingTimerStatusRunning {
		return nil, ErrTaskTimeTrackingTimerNotRunning{}
	}
	if taskID != 0 && current.TaskID != taskID {
		return nil, ErrTaskTimeTrackingTimerNotRunning{}
	}

	now := time.Now()
	timeSpent := int64(now.Sub(current.StartedAt).Seconds())
	if timeSpent <= 0 {
		timeSpent = 1
	}

	entry := &TaskTimeTracking{
		TaskID:    current.TaskID,
		UserID:    a.GetID(),
		TimeSpent: timeSpent,
		TrackedAt: current.StartedAt,
	}
	if _, err := s.Insert(entry); err != nil {
		return nil, err
	}
	if _, err := s.ID(current.ID).Delete(&TaskTimeTrackingTimer{}); err != nil {
		return nil, err
	}

	return entry, triggerTaskUpdatedEventForTaskID(s, a, current.TaskID, "Time tracking timer stopped")
}

func AdjustCurrentTaskTimeTrackingTimer(s *xorm.Session, a web.Auth, deltaSeconds int64) (*TaskTimeTrackingTimer, error) {
	current, err := GetCurrentTaskTimeTrackingTimer(s, a)
	if err != nil {
		return nil, err
	}
	if current == nil || current.Status != TaskTimeTrackingTimerStatusRunning {
		return nil, ErrTaskTimeTrackingTimerNotRunning{}
	}

	current.StartedAt = current.StartedAt.Add(-time.Duration(deltaSeconds) * time.Second)
	maxStart := time.Now().Add(-1 * time.Second)
	if current.StartedAt.After(maxStart) {
		current.StartedAt = maxStart
	}

	if _, err := s.ID(current.ID).Cols("started_at").Update(current); err != nil {
		return nil, err
	}

	return current, triggerTaskUpdatedEventForTaskID(s, a, current.TaskID, "Time tracking timer adjusted")
}

func autoStopTaskTimeTrackingTimers(now time.Time) error {
	s := db.NewSession()
	defer s.Close()

	timers := []*TaskTimeTrackingTimer{}
	if err := s.
		Where("status = ? AND started_at <= ?", TaskTimeTrackingTimerStatusRunning, now.Add(-TaskTimeTrackingTimerMaxDuration)).
		Find(&timers); err != nil {
		return err
	}

	for _, timer := range timers {
		stoppedAt := timer.StartedAt.Add(TaskTimeTrackingTimerMaxDuration)
		entry := &TaskTimeTracking{
			TaskID:    timer.TaskID,
			UserID:    timer.UserID,
			TimeSpent: int64(TaskTimeTrackingTimerMaxDuration.Seconds()),
			TrackedAt: timer.StartedAt,
			IsSnoozed: true,
		}
		if _, err := s.Insert(entry); err != nil {
			return err
		}

		timer.Status = TaskTimeTrackingTimerStatusSnoozed
		timer.StoppedAt = stoppedAt
		if _, err := s.ID(timer.ID).Cols("status", "stopped_at").Update(timer); err != nil {
			return err
		}

		u, err := user.GetUserByID(s, timer.UserID)
		if err != nil {
			return err
		}
		task, err := GetTaskByIDSimple(s, timer.TaskID)
		if err != nil {
			return err
		}
		if err := notifications.Notify(u, &TaskTimerAutoStoppedNotification{
			Task:  &task,
			Timer: timer,
		}, s); err != nil {
			return err
		}
	}

	return s.Commit()
}

func RegisterTaskTimeTrackingTimerAutoStopCron() {
	err := cron.Schedule("* * * * *", func() {
		if err := autoStopTaskTimeTrackingTimers(time.Now()); err != nil {
			log.Errorf("Could not auto-stop task time tracking timers: %s", err)
		}
	})
	if err != nil {
		log.Fatalf("Could not register task time tracking timer auto-stop cron: %s", err)
	}
}
