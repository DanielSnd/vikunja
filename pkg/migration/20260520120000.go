package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type userReportApplication20260520120000 struct {
	ID                 int64     `xorm:"bigint autoincr not null unique pk"`
	Name               string    `xorm:"not null"`
	ProjectID          int64     `xorm:"bigint not null index"`
	OwnerID            int64     `xorm:"bigint not null index"`
	AccessKeySalt      string    `xorm:"not null"`
	AccessKeyHash      string    `xorm:"not null unique"`
	AccessKeyLastEight string    `xorm:"not null index varchar(8)"`
	Created            time.Time `xorm:"created not null"`
	Updated            time.Time `xorm:"updated not null"`
}

func (userReportApplication20260520120000) TableName() string {
	return "user_report_applications"
}

type userReportToken20260520120000 struct {
	ID             int64     `xorm:"bigint autoincr not null unique pk"`
	ApplicationID  int64     `xorm:"bigint not null index"`
	Label          string    `xorm:"not null"`
	TokenSalt      string    `xorm:"not null"`
	TokenHash      string    `xorm:"not null unique"`
	TokenLastEight string    `xorm:"not null index varchar(8)"`
	IsEnabled      bool      `xorm:"not null default true"`
	CreatedByID    int64     `xorm:"bigint not null"`
	Created        time.Time `xorm:"created not null"`
}

func (userReportToken20260520120000) TableName() string {
	return "user_report_tokens"
}

type userReport20260520120000 struct {
	ID            int64     `xorm:"bigint autoincr not null unique pk"`
	TaskID        int64     `xorm:"bigint not null unique index"`
	ApplicationID int64     `xorm:"bigint not null index"`
	ReportTokenID int64     `xorm:"bigint not null index"`
	Severity      string    `xorm:"varchar(16) not null"`
	ReporterEmail string    `xorm:"varchar(191)"`
	Created       time.Time `xorm:"created not null"`
}

func (userReport20260520120000) TableName() string {
	return "user_reports"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260520120000",
		Description: "Add user reports",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(
				userReportApplication20260520120000{},
				userReportToken20260520120000{},
				userReport20260520120000{},
			)
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
