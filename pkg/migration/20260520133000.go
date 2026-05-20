package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260520133000",
		Description: "Extend user reports settings and notification state",
		Migrate: func(tx *xorm.Engine) error {
			type columnSpec struct {
				table      string
				column     string
				definition string
			}

			columns := []columnSpec{
				{"user_report_applications", "max_upload_size", "bigint NOT NULL DEFAULT 0"},
				{"user_report_applications", "critical_project_id", "bigint NOT NULL DEFAULT 0"},
				{"user_report_applications", "high_project_id", "bigint NOT NULL DEFAULT 0"},
				{"user_report_applications", "low_project_id", "bigint NOT NULL DEFAULT 0"},
				{"user_report_applications", "critical_priority", "bigint NOT NULL DEFAULT 100"},
				{"user_report_applications", "high_priority", "bigint NOT NULL DEFAULT 75"},
				{"user_report_applications", "low_priority", "bigint NOT NULL DEFAULT 25"},
				{"user_reports", "application_name", "varchar(191) NOT NULL DEFAULT ''"},
				{"user_reports", "report_token_label", "varchar(191) NOT NULL DEFAULT ''"},
				{"user_reports", "notified_done_at", "DATETIME NULL"},
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

			if _, err := tx.Exec("UPDATE user_reports SET application_name = '' WHERE application_name IS NULL"); err != nil {
				return err
			}
			if _, err := tx.Exec("UPDATE user_reports SET report_token_label = '' WHERE report_token_label IS NULL"); err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
