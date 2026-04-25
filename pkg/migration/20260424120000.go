// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type milestone20260424120000 struct {
	ID            int64      `xorm:"bigint autoincr not null unique pk"`
	ProjectID     int64      `xorm:"bigint not null index"`
	Name          string     `xorm:"varchar(250) not null"`
	MilestoneDate *time.Time `xorm:"DATETIME null 'milestone_date'"`
	HexColor      string     `xorm:"varchar(6) null"`
	Created       time.Time  `xorm:"created not null"`
	Updated       time.Time  `xorm:"updated not null"`
}

func (milestone20260424120000) TableName() string {
	return "milestones"
}

type milestoneUser20260424120000 struct {
	ID          int64     `xorm:"bigint autoincr not null unique pk"`
	MilestoneID int64     `xorm:"bigint not null index"`
	UserID      int64     `xorm:"bigint not null index"`
	Created     time.Time `xorm:"created not null"`
}

func (milestoneUser20260424120000) TableName() string {
	return "milestone_users"
}

type task20260424120000 struct {
	ID          int64 `xorm:"bigint autoincr not null unique pk"`
	MilestoneID int64 `xorm:"bigint null"`
}

func (task20260424120000) TableName() string {
	return "tasks"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260424120000",
		Description: "add milestones and task milestone relation",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync(
				milestone20260424120000{},
				milestoneUser20260424120000{},
				task20260424120000{},
			)
		},
		Rollback: func(tx *xorm.Engine) error {
			exists, err := columnExists(tx, "tasks", "milestone_id")
			if err != nil {
				return err
			}
			if exists {
				if err := dropTableColum(tx, "tasks", "milestone_id"); err != nil {
					return err
				}
			}

			return tx.DropTables(
				milestoneUser20260424120000{},
				milestone20260424120000{},
			)
		},
	})
}
