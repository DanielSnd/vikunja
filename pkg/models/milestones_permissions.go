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

func (m *Milestone) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	project := &Project{ID: m.ProjectID}
	return project.CanWrite(s, a)
}

func (m *Milestone) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	existing, err := getMilestoneSimpleByID(s, m.ID)
	if err != nil {
		return false, 0, err
	}

	project := &Project{ID: existing.ProjectID}
	return project.CanRead(s, a)
}

func (m *Milestone) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return m.canWrite(s, a)
}

func (m *Milestone) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return m.canWrite(s, a)
}

func (m *Milestone) canWrite(s *xorm.Session, a web.Auth) (bool, error) {
	existing, err := getMilestoneSimpleByID(s, m.ID)
	if err != nil {
		return false, err
	}

	project := &Project{ID: existing.ProjectID}
	return project.CanWrite(s, a)
}
