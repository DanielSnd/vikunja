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

func (vb *VisionBoard) resolveTask(s *xorm.Session) (*Task, error) {
	if vb.TaskID == 0 && vb.ID != 0 {
		board, err := getVisionBoardByIDAndProject(s, vb.ID, vb.ProjectID)
		if err != nil {
			return nil, err
		}
		vb.TaskID = board.TaskID
		vb.ProjectID = board.ProjectID
	}

	task := &Task{ID: vb.TaskID}
	return task, nil
}

func (vb *VisionBoard) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	task, err := vb.resolveTask(s)
	if err != nil {
		return false, 0, err
	}

	return task.CanRead(s, a)
}

func (vb *VisionBoard) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	task, err := vb.resolveTask(s)
	if err != nil {
		return false, err
	}

	return task.CanUpdate(s, a)
}

func (vb *VisionBoard) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	task, err := vb.resolveTask(s)
	if err != nil {
		return false, err
	}

	return task.CanUpdate(s, a)
}

func (vb *VisionBoard) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	task, err := vb.resolveTask(s)
	if err != nil {
		return false, err
	}

	return task.CanUpdate(s, a)
}
