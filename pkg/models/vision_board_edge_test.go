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

func TestVisionBoardEdge_CRUD(t *testing.T) {
	u := &user.User{ID: 1}

	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	board := &VisionBoard{ProjectID: 1, TaskID: 1}
	require.NoError(t, board.Create(s, u))

	source := &VisionBoardNode{
		ProjectID: board.ProjectID,
		BoardID:   board.ID,
		Kind:      VisionBoardNodeKindText,
		Title:     "Source",
	}
	target := &VisionBoardNode{
		ProjectID: board.ProjectID,
		BoardID:   board.ID,
		Kind:      VisionBoardNodeKindText,
		Title:     "Target",
		X:         100,
	}
	require.NoError(t, source.Create(s, u))
	require.NoError(t, target.Create(s, u))

	edge := &VisionBoardEdge{
		ProjectID:    board.ProjectID,
		BoardID:      board.ID,
		SourceNodeID: source.ID,
		TargetNodeID: target.ID,
		SourceHandle: "right",
		TargetHandle: "left",
		Label:        "depends on",
	}
	require.NoError(t, edge.Create(s, u))

	edge.Label = "blocks"
	require.NoError(t, edge.Update(s, u))

	result, count, total, err := (&VisionBoardEdge{
		ProjectID: board.ProjectID,
		BoardID:   board.ID,
	}).ReadAll(s, u, "", 0, 0)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	require.EqualValues(t, 1, total)

	edges := result.([]*VisionBoardEdge)
	require.Len(t, edges, 1)
	assert.Equal(t, "blocks", edges[0].Label)
	assert.Equal(t, "right", edges[0].SourceHandle)
	assert.Equal(t, "left", edges[0].TargetHandle)

	require.NoError(t, (&VisionBoardEdge{
		ProjectID: board.ProjectID,
		BoardID:   board.ID,
		ID:        edge.ID,
	}).Delete(s, u))
}
