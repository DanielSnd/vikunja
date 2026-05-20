package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260520154000",
		Description: "Normalize user report priorities to Vikunja priority scale",
		Migrate: func(tx *xorm.Engine) error {
			statements := []string{
				"UPDATE user_report_applications SET critical_priority = 5 WHERE critical_priority = 100",
				"UPDATE user_report_applications SET high_priority = 4 WHERE high_priority = 75",
				"UPDATE user_report_applications SET low_priority = 1 WHERE low_priority = 25",
				"UPDATE user_report_applications SET critical_priority = 5 WHERE critical_priority > 5",
				"UPDATE user_report_applications SET high_priority = 5 WHERE high_priority > 5",
				"UPDATE user_report_applications SET low_priority = 5 WHERE low_priority > 5",
				"UPDATE user_report_applications SET critical_priority = 0 WHERE critical_priority < 0",
				"UPDATE user_report_applications SET high_priority = 0 WHERE high_priority < 0",
				"UPDATE user_report_applications SET low_priority = 0 WHERE low_priority < 0",
			}

			for _, statement := range statements {
				if _, err := tx.Exec(statement); err != nil {
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
