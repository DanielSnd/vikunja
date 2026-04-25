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

package models

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMilestone_CRUD(t *testing.T) {
	u := &user.User{ID: 1, Username: "user1"}

	t.Run("create", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		milestoneDate := time.Date(2026, time.April, 24, 12, 0, 0, 0, time.UTC)
		milestone := &Milestone{
			ProjectID:     1,
			Name:          "Release 1.0",
			MilestoneDate: &milestoneDate,
			HexColor:      "#ff0000",
			Users: []*user.User{
				{ID: 1},
			},
		}

		err := milestone.Create(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertExists(t, "milestones", map[string]interface{}{
			"id":         milestone.ID,
			"project_id": milestone.ProjectID,
			"name":       milestone.Name,
			"hex_color":  "ff0000",
		}, false)
		db.AssertExists(t, "milestone_users", map[string]interface{}{
			"milestone_id": milestone.ID,
			"user_id":      1,
		}, false)
	})

	t.Run("update", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		milestone := &Milestone{
			ProjectID: 1,
			Name:      "Release 1.0",
		}
		require.NoError(t, milestone.Create(s, u))

		milestone.Name = "Release 1.1"
		milestone.HexColor = "#00ff00"
		milestone.Users = []*user.User{}

		err := milestone.Update(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertExists(t, "milestones", map[string]interface{}{
			"id":        milestone.ID,
			"name":      "Release 1.1",
			"hex_color": "00ff00",
		}, false)
		db.AssertMissing(t, "milestone_users", map[string]interface{}{
			"milestone_id": milestone.ID,
			"user_id":      1,
		})
	})

	t.Run("update persists milestone date", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		milestone := &Milestone{
			ProjectID: 1,
			Name:      "Release 1.0",
		}
		require.NoError(t, milestone.Create(s, u))

		expectedDate := time.Date(2026, time.April, 30, 15, 45, 0, 0, time.UTC)
		milestone.MilestoneDate = &expectedDate

		err := milestone.Update(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		read := &Milestone{ID: milestone.ID}
		err = read.ReadOne(s, u)
		require.NoError(t, err)
		require.NotNil(t, read.MilestoneDate)
		assert.Equal(t, expectedDate.Unix(), read.MilestoneDate.Unix())
	})

	t.Run("delete clears task relation", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		milestone := &Milestone{
			ProjectID: 1,
			Name:      "Release 1.0",
		}
		require.NoError(t, milestone.Create(s, u))

		_, err := s.Where("id = ?", 1).
			Cols("milestone_id").
			Update(&Task{MilestoneID: milestone.ID})
		require.NoError(t, err)

		err = milestone.Delete(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertMissing(t, "milestones", map[string]interface{}{
			"id": milestone.ID,
		})
		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":           1,
			"milestone_id": 0,
		}, false)
	})

	t.Run("read one populates users", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		milestone := &Milestone{
			ProjectID: 1,
			Name:      "Release 1.0",
			Users: []*user.User{
				{ID: 1},
			},
		}
		require.NoError(t, milestone.Create(s, u))

		read := &Milestone{ID: milestone.ID}
		err := read.ReadOne(s, u)
		require.NoError(t, err)
		assert.Len(t, read.Users, 1)
		assert.Equal(t, int64(1), read.Users[0].ID)
	})
}
