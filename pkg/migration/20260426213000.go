package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type visionBoardEdge20260426213000 struct {
	ID           int64     `xorm:"bigint autoincr not null unique pk"`
	BoardID      int64     `xorm:"bigint not null index 'board_id'"`
	SourceNodeID int64     `xorm:"bigint not null index 'source_node_id'"`
	TargetNodeID int64     `xorm:"bigint not null index 'target_node_id'"`
	SourceHandle string    `xorm:"varchar(16) not null default ''"`
	TargetHandle string    `xorm:"varchar(16) not null default ''"`
	Label        string    `xorm:"varchar(255) not null default ''"`
	Color        string    `xorm:"varchar(32) not null default ''"`
	Version      int64     `xorm:"bigint not null default 1"`
	Updated      time.Time `xorm:"updated not null"`
	Created      time.Time `xorm:"created not null"`
}

func (visionBoardEdge20260426213000) TableName() string {
	return "vision_board_edges"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260426213000",
		Description: "Add vision board edges table",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(visionBoardEdge20260426213000{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
