package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type taskTimeTracking20260425120000 struct {
	ID        int64  `xorm:"bigint autoincr not null unique pk"`
	TaskID    int64  `xorm:"bigint not null index"`
	UserID    int64  `xorm:"bigint not null index"`
	TimeSpent int64  `xorm:"bigint not null"`
	TrackedAt string `xorm:"datetime not null 'tracked_at'"`
	IsSnoozed bool   `xorm:"not null default false 'is_snoozed'"`
}

func (taskTimeTracking20260425120000) TableName() string {
	return "task_time_tracking"
}

type taskTimeTrackingTimer20260425120000 struct {
	ID        int64  `xorm:"bigint autoincr not null unique pk"`
	TaskID    int64  `xorm:"bigint not null index"`
	UserID    int64  `xorm:"bigint not null unique index 'user_id'"`
	StartedAt string `xorm:"datetime not null 'started_at'"`
	StoppedAt string `xorm:"datetime null 'stopped_at'"`
	Status    string `xorm:"varchar(20) not null default 'running'"`
}

func (taskTimeTrackingTimer20260425120000) TableName() string {
	return "task_time_tracking_timers"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260425120000",
		Description: "Add task time tracking tables",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(taskTimeTracking20260425120000{}, taskTimeTrackingTimer20260425120000{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
