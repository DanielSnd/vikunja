package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type visionBoardNode20260426210000 struct {
	ID           int64     `xorm:"bigint autoincr not null unique pk"`
	BoardID      int64     `xorm:"bigint not null index 'board_id'"`
	ParentNodeID int64     `xorm:"bigint null index 'parent_node_id'"`
	TaskID       int64     `xorm:"bigint null index 'task_id'"`
	AttachmentID int64     `xorm:"bigint null index 'attachment_id'"`
	Kind         string    `xorm:"varchar(32) not null"`
	Title        string    `xorm:"varchar(255) not null default ''"`
	Content      string    `xorm:"text null"`
	URL          string    `xorm:"text null"`
	X            float64   `xorm:"double not null default 0"`
	Y            float64   `xorm:"double not null default 0"`
	Width        float64   `xorm:"double not null default 240"`
	Height       float64   `xorm:"double not null default 160"`
	Color        string    `xorm:"varchar(32) not null default ''"`
	ZIndex       int64     `xorm:"bigint not null default 0 'z_index'"`
	Version      int64     `xorm:"bigint not null default 1"`
	CreatedByID  int64     `xorm:"bigint not null 'created_by_id'"`
	UpdatedByID  int64     `xorm:"bigint not null 'updated_by_id'"`
	Updated      time.Time `xorm:"updated not null"`
	Created      time.Time `xorm:"created not null"`
}

func (visionBoardNode20260426210000) TableName() string {
	return "vision_board_nodes"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260426210000",
		Description: "Add vision board nodes table for existing vision board installs",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(visionBoardNode20260426210000{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
