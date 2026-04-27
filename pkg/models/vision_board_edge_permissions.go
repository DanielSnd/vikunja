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

func (e *VisionBoardEdge) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	board, err := e.resolveBoard(s)
	if err != nil {
		return false, 0, err
	}

	return board.CanRead(s, a)
}

func (e *VisionBoardEdge) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	board, err := e.resolveBoard(s)
	if err != nil {
		return false, err
	}

	return board.CanUpdate(s, a)
}

func (e *VisionBoardEdge) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return e.CanCreate(s, a)
}

func (e *VisionBoardEdge) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return e.CanCreate(s, a)
}
