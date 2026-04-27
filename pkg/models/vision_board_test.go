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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVisionBoard_Create(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("creates board for task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		board := &VisionBoard{
			ProjectID: 1,
			TaskID:    1,
		}

		err := board.Create(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertExists(t, "vision_boards", map[string]interface{}{
			"id":         board.ID,
			"project_id": int64(1),
			"task_id":    int64(1),
		}, false)
		assert.NotEmpty(t, board.Title)
		assert.Equal(t, 1.0, board.ViewportZoom)
	})

	t.Run("prevents duplicate board per task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		first := &VisionBoard{
			ProjectID: 1,
			TaskID:    1,
		}
		err := first.Create(s, u)
		require.NoError(t, err)

		second := &VisionBoard{
			ProjectID: 1,
			TaskID:    1,
		}
		err = second.Create(s, u)
		require.Error(t, err)
		assert.True(t, IsErrVisionBoardAlreadyExistsForTask(err))
	})
}

func TestVisionBoard_ReadAll(t *testing.T) {
	u := &user.User{ID: 1}

	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	board := &VisionBoard{
		ProjectID: 1,
		TaskID:    1,
	}
	require.NoError(t, board.Create(s, u))

	result, count, total, err := (&VisionBoard{ProjectID: 1}).ReadAll(s, u, "", 0, 0)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	require.EqualValues(t, 1, total)

	boards := result.([]*VisionBoard)
	require.Len(t, boards, 1)
	assert.Equal(t, int64(1), boards[0].TaskID)
	assert.NotEmpty(t, boards[0].TaskTitle)
}
