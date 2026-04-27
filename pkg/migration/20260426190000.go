package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type visionBoard20260426190000 struct {
	ID           int64     `xorm:"bigint autoincr not null unique pk"`
	ProjectID    int64     `xorm:"bigint not null index 'project_id'"`
	TaskID       int64     `xorm:"bigint not null unique index 'task_id'"`
	Title        string    `xorm:"varchar(255) not null"`
	ViewportX    float64   `xorm:"double not null default 0 'viewport_x'"`
	ViewportY    float64   `xorm:"double not null default 0 'viewport_y'"`
	ViewportZoom float64   `xorm:"double not null default 1 'viewport_zoom'"`
	CreatedByID  int64     `xorm:"bigint not null 'created_by_id'"`
	Updated      time.Time `xorm:"updated not null"`
	Created      time.Time `xorm:"created not null"`
}

func (visionBoard20260426190000) TableName() string {
	return "vision_boards"
}

type visionBoardNode20260426190000 struct {
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

func (visionBoardNode20260426190000) TableName() string {
	return "vision_board_nodes"
}

type project20260426190000 struct {
	ID int64 `xorm:"bigint pk"`
}

func (project20260426190000) TableName() string {
	return "projects"
}

type projectView20260426190000 struct {
	ID                      int64     `xorm:"bigint autoincr not null unique pk"`
	Title                   string    `xorm:"varchar(255) not null"`
	ProjectID               int64     `xorm:"not null index 'project_id'"`
	ViewKind                int       `xorm:"not null 'view_kind'"`
	Filter                  string    `xorm:"json null default null"`
	Position                float64   `xorm:"double null"`
	BucketConfigurationMode int       `xorm:"default 0 'bucket_configuration_mode'"`
	DefaultBucketID         int64     `xorm:"bigint INDEX null 'default_bucket_id'"`
	DoneBucketID            int64     `xorm:"bigint INDEX null 'done_bucket_id'"`
	HighBucketID            int64     `xorm:"bigint INDEX null 'high_bucket_id'"`
	MedBucketID             int64     `xorm:"bigint INDEX null 'med_bucket_id'"`
	LowBucketID             int64     `xorm:"bigint INDEX null 'low_bucket_id'"`
	Updated                 time.Time `xorm:"updated not null"`
	Created                 time.Time `xorm:"created not null"`
}

func (projectView20260426190000) TableName() string {
	return "project_views"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260426190000",
		Description: "Add vision boards and default vision board project views",
		Migrate: func(tx *xorm.Engine) error {
			if err := tx.Sync2(visionBoard20260426190000{}, visionBoardNode20260426190000{}); err != nil {
				return err
			}

			projects := []*project20260426190000{}
			if err := tx.Find(&projects); err != nil {
				return err
			}

			for _, project := range projects {
				exists, err := tx.Where("project_id = ? AND view_kind = ?", project.ID, 4).Exist(&projectView20260426190000{})
				if err != nil {
					return err
				}
				if exists {
					continue
				}

				view := &projectView20260426190000{
					Title:     "Vision Board",
					ProjectID: project.ID,
					ViewKind:  4,
					Position:  500,
				}
				if _, err := tx.Insert(view); err != nil {
					return err
				}
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
