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
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

func (n *VisionBoardNode) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	board, err := n.resolveBoard(s)
	if err != nil {
		return false, 0, err
	}

	return board.CanRead(s, a)
}

func (n *VisionBoardNode) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	board, err := n.resolveBoard(s)
	if err != nil {
		return false, err
	}

	can, err := board.CanUpdate(s, a)
	if err != nil || !can {
		return can, err
	}

	if n.TaskID == 0 {
		return true, nil
	}

	task := &Task{ID: n.TaskID}
	taskCanRead, _, taskErr := task.CanRead(s, a)
	return taskCanRead, taskErr
}

func (n *VisionBoardNode) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return n.CanCreate(s, a)
}

func (n *VisionBoardNode) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	board, err := n.resolveBoard(s)
	if err != nil {
		return false, err
	}

	return board.CanUpdate(s, a)
}
