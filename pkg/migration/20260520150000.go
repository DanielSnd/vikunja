package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260520150000",
		Description: "Add bucket routing to user report applications",
		Migrate: func(tx *xorm.Engine) error {
			type columnSpec struct {
				table      string
				column     string
				definition string
			}

			columns := []columnSpec{
				{"user_report_applications", "bucket_id", "bigint NOT NULL DEFAULT 0"},
				{"user_report_applications", "critical_bucket_id", "bigint NOT NULL DEFAULT 0"},
				{"user_report_applications", "high_bucket_id", "bigint NOT NULL DEFAULT 0"},
				{"user_report_applications", "low_bucket_id", "bigint NOT NULL DEFAULT 0"},
			}

			for _, column := range columns {
				exists, err := columnExists(tx, column.table, column.column)
				if err != nil {
					return err
				}
				if exists {
					continue
				}

				if _, err = tx.Exec("ALTER TABLE " + column.table + " ADD COLUMN " + column.column + " " + column.definition); err != nil {
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
