package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260427100000",
		Description: "Add vision board edge handles",
		Migrate: func(tx *xorm.Engine) error {
			columns := []string{"source_handle", "target_handle"}
			for _, col := range columns {
				exists, err := columnExists(tx, "vision_board_edges", col)
				if err != nil {
					return err
				}
				if exists {
					continue
				}

				if _, err = tx.Exec("ALTER TABLE vision_board_edges ADD COLUMN " + col + " varchar(16) NOT NULL DEFAULT ''"); err != nil {
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
