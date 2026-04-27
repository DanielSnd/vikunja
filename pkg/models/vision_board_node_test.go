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

func TestVisionBoardNode_CRUD(t *testing.T) {
	u := &user.User{ID: 1}

	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	board := &VisionBoard{
		ProjectID: 1,
		TaskID:    1,
	}
	require.NoError(t, board.Create(s, u))

	node := &VisionBoardNode{
		ProjectID: board.ProjectID,
		BoardID:   board.ID,
		Kind:      VisionBoardNodeKindText,
		Title:     "Idea",
		Content:   "Ship the durable canvas",
		X:         120,
		Y:         80,
	}
	require.NoError(t, node.Create(s, u))

	created := &VisionBoardNode{}
	exists, err := s.Where("id = ? AND board_id = ?", node.ID, board.ID).Get(created)
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, "Idea", created.Title)

	node.Content = "Ship the durable canvas soon"
	node.Width = 320
	require.NoError(t, node.Update(s, u))

	result, count, total, err := (&VisionBoardNode{
		ProjectID: board.ProjectID,
		BoardID:   board.ID,
	}).ReadAll(s, u, "", 0, 0)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	require.EqualValues(t, 1, total)

	nodes := result.([]*VisionBoardNode)
	require.Len(t, nodes, 1)
	assert.Equal(t, "Ship the durable canvas soon", nodes[0].Content)
	assert.EqualValues(t, 320, nodes[0].Width)

	require.NoError(t, (&VisionBoardNode{
		ProjectID: board.ProjectID,
		BoardID:   board.ID,
		ID:        node.ID,
	}).Delete(s, u))

	stillExists, err := s.Where("id = ?", node.ID).Exist(&VisionBoardNode{})
	require.NoError(t, err)
	assert.False(t, stillExists)
}
