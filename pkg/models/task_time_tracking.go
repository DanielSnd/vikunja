package models

import (
	"fmt"
	"sort"
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

type TaskTimeTracking struct {
	ID        int64      `xorm:"bigint autoincr not null unique pk" json:"id" param:"timetracking"`
	TaskID    int64      `xorm:"bigint not null index" json:"-" param:"task"`
	UserID    int64      `xorm:"bigint not null index" json:"-"`
	User      *user.User `xorm:"-" json:"user"`
	TimeSpent int64      `xorm:"bigint not null" json:"time_spent"`
	TrackedAt time.Time  `xorm:"datetime not null 'tracked_at'" json:"tracked_at"`
	IsSnoozed bool       `xorm:"not null default false 'is_snoozed'" json:"is_snoozed"`
	OrderBy   string     `xorm:"-" json:"-" query:"order_by"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

type TaskTimeTrackingSummary struct {
	User      *user.User `xorm:"-" json:"user"`
	UserID    int64      `xorm:"-" json:"-"`
	TimeSpent int64      `xorm:"-" json:"time_spent"`
}

func (*TaskTimeTracking) TableName() string {
	return "task_time_tracking"
}

func (t *TaskTimeTracking) validate() error {
	if t.TimeSpent <= 0 {
		return InvalidFieldError([]string{"time_spent"})
	}
	if t.TrackedAt.IsZero() {
		t.TrackedAt = time.Now()
	}
	return nil
}

func (t *TaskTimeTracking) Create(s *xorm.Session, a web.Auth) error {
	if err := t.validate(); err != nil {
		return err
	}

	t.ID = 0
	t.IsSnoozed = false

	if _, err := GetTaskSimple(s, &Task{ID: t.TaskID}); err != nil {
		return err
	}

	u, err := GetUserOrLinkShareUser(s, a)
	if err != nil {
		return err
	}
	t.UserID = u.ID
	t.User = u

	if _, err = s.Insert(t); err != nil {
		return err
	}

	return triggerTaskUpdatedEventForTaskID(s, a, t.TaskID, "Time tracking created")
}

func (t *TaskTimeTracking) ReadOne(s *xorm.Session, a web.Auth) error {
	canRead, _, err := t.CanRead(s, a)
	if err != nil {
		return err
	}
	if !canRead {
		return ErrGenericForbidden{}
	}

	return getTaskTimeTrackingSimple(s, t)
}

func (t *TaskTimeTracking) Update(s *xorm.Session, a web.Auth) error {
	if err := t.validate(); err != nil {
		return err
	}

	saved := &TaskTimeTracking{ID: t.ID, TaskID: t.TaskID}
	if err := getTaskTimeTrackingSimple(s, saved); err != nil {
		return err
	}

	t.UserID = saved.UserID
	t.IsSnoozed = false

	updated, err := s.
		ID(t.ID).
		Cols("time_spent", "tracked_at", "is_snoozed").
		Update(t)
	if err != nil {
		return err
	}
	if updated == 0 {
		return ErrTaskTimeTrackingDoesNotExist{ID: t.ID, TaskID: t.TaskID}
	}

	return triggerTaskUpdatedEventForTaskID(s, a, t.TaskID, "Time tracking updated")
}

func (t *TaskTimeTracking) Delete(s *xorm.Session, a web.Auth) error {
	if err := getTaskTimeTrackingSimple(s, t); err != nil {
		return err
	}

	deleted, err := s.ID(t.ID).NoAutoCondition().Delete(t)
	if err != nil {
		return err
	}
	if deleted == 0 {
		return ErrTaskTimeTrackingDoesNotExist{ID: t.ID, TaskID: t.TaskID}
	}

	return triggerTaskUpdatedEventForTaskID(s, a, t.TaskID, "Time tracking deleted")
}

func (t *TaskTimeTracking) ReadAll(s *xorm.Session, a web.Auth, _ string, page int, perPage int) (interface{}, int, int64, error) {
	canRead, _, err := t.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canRead {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	return getAllTimeTrackingForTasksWithoutPermissionCheck(s, []int64{t.TaskID}, page, perPage, t.OrderBy)
}

func getTaskTimeTrackingSimple(s *xorm.Session, t *TaskTimeTracking) error {
	query := s.Where("id = ?", t.ID).NoAutoCondition()
	if t.TaskID != 0 {
		query = query.And("task_id = ?", t.TaskID)
	}

	exists, err := query.Get(t)
	if err != nil {
		return err
	}
	if !exists {
		return ErrTaskTimeTrackingDoesNotExist{ID: t.ID, TaskID: t.TaskID}
	}

	users, err := getUsersOrLinkSharesFromIDs(s, []int64{t.UserID})
	if err != nil {
		return err
	}
	t.User = users[t.UserID]
	return nil
}

func getAllTimeTrackingForTasksWithoutPermissionCheck(s *xorm.Session, taskIDs []int64, page int, perPage int, orderBy string) ([]*TaskTimeTracking, int, int64, error) {
	order := "desc"
	if orderBy == "asc" {
		order = "asc"
	}

	limit, start := getLimitFromPageIndex(page, perPage)
	entries := []*TaskTimeTracking{}

	query := s.
		Where(builder.In("task_id", taskIDs)).
		OrderBy(fmt.Sprintf("tracked_at %s, id %s", order, order))
	if limit > 0 {
		query = query.Limit(limit, start)
	}

	if err := query.Find(&entries); err != nil {
		return nil, 0, 0, err
	}

	userIDs := make([]int64, 0, len(entries))
	for _, entry := range entries {
		userIDs = append(userIDs, entry.UserID)
	}
	users, err := getUsersOrLinkSharesFromIDs(s, userIDs)
	if err != nil {
		return nil, 0, 0, err
	}
	for _, entry := range entries {
		entry.User = users[entry.UserID]
	}

	total, err := s.In("task_id", taskIDs).Count(&TaskTimeTracking{})
	if err != nil {
		return nil, 0, 0, err
	}

	return entries, len(entries), total, nil
}

func addTimeTrackingSummaryToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*Task) error {
	if len(taskIDs) == 0 {
		return nil
	}

	type summaryRow struct {
		TaskID    int64 `xorm:"task_id"`
		UserID    int64 `xorm:"user_id"`
		TimeSpent int64 `xorm:"time_spent"`
	}

	rows := []summaryRow{}
	if err := s.
		Table("task_time_tracking").
		Select("task_id, user_id, SUM(time_spent) AS time_spent").
		Where(builder.In("task_id", taskIDs)).
		GroupBy("task_id, user_id").
		Find(&rows); err != nil {
		return err
	}

	userIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.UserID)
	}
	users, err := getUsersOrLinkSharesFromIDs(s, userIDs)
	if err != nil {
		return err
	}

	zero := int64(0)
	for _, taskID := range taskIDs {
		if task, ok := taskMap[taskID]; ok {
			task.TimeTrackingTotal = &zero
			task.TimeTrackingSummary = []*TaskTimeTrackingSummary{}
		}
	}

	for _, row := range rows {
		task, ok := taskMap[row.TaskID]
		if !ok {
			continue
		}
		task.TimeTrackingSummary = append(task.TimeTrackingSummary, &TaskTimeTrackingSummary{
			UserID:    row.UserID,
			User:      users[row.UserID],
			TimeSpent: row.TimeSpent,
		})
		total := *task.TimeTrackingTotal + row.TimeSpent
		task.TimeTrackingTotal = &total
	}

	for _, taskID := range taskIDs {
		task, ok := taskMap[taskID]
		if !ok || len(task.TimeTrackingSummary) == 0 {
			continue
		}
		sort.Slice(task.TimeTrackingSummary, func(i, j int) bool {
			left := ""
			right := ""
			if task.TimeTrackingSummary[i].User != nil {
				left = task.TimeTrackingSummary[i].User.GetName()
			}
			if task.TimeTrackingSummary[j].User != nil {
				right = task.TimeTrackingSummary[j].User.GetName()
			}
			return left < right
		})
	}

	return nil
}
