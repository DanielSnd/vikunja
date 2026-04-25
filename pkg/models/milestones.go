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
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// Milestone represents a project milestone a task can be assigned to.
type Milestone struct {
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"milestone"`

	ProjectID int64  `xorm:"bigint not null index" json:"project_id" param:"project"`
	Name      string `xorm:"varchar(250) not null" json:"name" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250"`
	// IncludeParents controls whether parent project milestones should be returned when listing milestones.
	IncludeParents bool `xorm:"-" json:"-" query:"include_parents"`

	MilestoneDate *time.Time `xorm:"DATETIME null 'milestone_date'" json:"milestone_date"`
	HexColor      string     `xorm:"varchar(6) null" json:"hex_color" valid:"runelength(0|7)" maxLength:"7"`

	Users []*user.User `xorm:"-" json:"users"`

	Created time.Time `xorm:"created not null" json:"created"`
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (*Milestone) TableName() string {
	return "milestones"
}

// MilestoneUser stores a milestone <-> user relation.
type MilestoneUser struct {
	ID          int64     `xorm:"bigint autoincr not null unique pk" json:"-"`
	MilestoneID int64     `xorm:"bigint not null index" json:"-"`
	UserID      int64     `xorm:"bigint not null index" json:"user_id"`
	Created     time.Time `xorm:"created not null" json:"created"`
}

func (*MilestoneUser) TableName() string {
	return "milestone_users"
}

type milestoneUserWithUser struct {
	MilestoneID int64
	user.User   `xorm:"extends"`
}

func getMilestoneSimpleByID(s *xorm.Session, milestoneID int64) (*Milestone, error) {
	if milestoneID < 1 {
		return nil, ErrMilestoneDoesNotExist{ID: milestoneID}
	}

	milestone := &Milestone{ID: milestoneID}
	has, err := s.Get(milestone)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrMilestoneDoesNotExist{ID: milestoneID}
	}

	return milestone, nil
}

func getMilestoneUsersForMilestones(s *xorm.Session, milestoneIDs []int64) ([]*milestoneUserWithUser, error) {
	if len(milestoneIDs) == 0 {
		return []*milestoneUserWithUser{}, nil
	}

	users := []*milestoneUserWithUser{}
	err := s.Table("milestone_users").
		Select("milestone_users.milestone_id, users.*").
		In("milestone_users.milestone_id", milestoneIDs).
		Join("INNER", "users", "milestone_users.user_id = users.id").
		Find(&users)

	return users, err
}

func addUsersToMilestones(s *xorm.Session, milestones []*Milestone) error {
	if len(milestones) == 0 {
		return nil
	}

	milestoneMap := make(map[int64]*Milestone, len(milestones))
	milestoneIDs := make([]int64, 0, len(milestones))
	for _, milestone := range milestones {
		milestoneMap[milestone.ID] = milestone
		milestoneIDs = append(milestoneIDs, milestone.ID)
	}

	users, err := getMilestoneUsersForMilestones(s, milestoneIDs)
	if err != nil {
		return err
	}

	for i := range users {
		users[i].Email = ""
		milestone := milestoneMap[users[i].MilestoneID]
		if milestone != nil {
			milestone.Users = append(milestone.Users, &users[i].User)
		}
	}

	return nil
}

func (m *Milestone) syncUsers(s *xorm.Session, users []*user.User) error {
	currentUsers, err := getMilestoneUsersForMilestones(s, []int64{m.ID})
	if err != nil {
		return err
	}

	existing := make(map[int64]bool, len(currentUsers))
	for _, currentUser := range currentUsers {
		existing[currentUser.ID] = true
	}

	nextUsers := make(map[int64]*user.User, len(users))
	for _, milestoneUser := range users {
		if milestoneUser == nil || milestoneUser.ID == 0 {
			return InvalidFieldError([]string{"users"})
		}

		hasAccess, err := canUserAccessProject(s, milestoneUser.ID, m.ProjectID)
		if err != nil {
			return err
		}
		if !hasAccess {
			return ErrUserDoesNotHaveAccessToProject{ProjectID: m.ProjectID, UserID: milestoneUser.ID}
		}

		nextUsers[milestoneUser.ID] = milestoneUser
	}

	for userID := range existing {
		if nextUsers[userID] != nil {
			continue
		}

		_, err = s.Where("milestone_id = ? AND user_id = ?", m.ID, userID).Delete(&MilestoneUser{})
		if err != nil {
			return err
		}
	}

	for userID := range nextUsers {
		if existing[userID] {
			continue
		}

		_, err = s.Insert(&MilestoneUser{
			MilestoneID: m.ID,
			UserID:      userID,
		})
		if err != nil {
			return err
		}
	}

	m.Users = users
	return nil
}

func canUserAccessProject(s *xorm.Session, userID, projectID int64) (bool, error) {
	project := &Project{ID: projectID}
	canRead, _, err := project.CanRead(s, &user.User{ID: userID})
	return canRead, err
}

func getProjectAndParentProjectIDs(s *xorm.Session, projectID int64) ([]int64, error) {
	projects, err := GetAllParentProjects(s, projectID)
	if err != nil {
		return nil, err
	}

	projectIDs := make([]int64, 0, len(projects))
	for id := range projects {
		projectIDs = append(projectIDs, id)
	}

	return projectIDs, nil
}

// Create creates a new milestone.
func (m *Milestone) Create(s *xorm.Session, _ web.Auth) error {
	if _, err := GetProjectSimpleByID(s, m.ProjectID); err != nil {
		return err
	}

	m.ID = 0
	m.HexColor = utils.NormalizeHex(m.HexColor)

	_, err := s.Insert(m)
	if err != nil {
		return err
	}

	if err := m.syncUsers(s, m.Users); err != nil {
		return err
	}

	if err := updateProjectLastUpdated(s, &Project{ID: m.ProjectID}); err != nil {
		return err
	}

	return m.ReadOne(s, nil)
}

// Update updates an existing milestone.
func (m *Milestone) Update(s *xorm.Session, _ web.Auth) error {
	existing, err := getMilestoneSimpleByID(s, m.ID)
	if err != nil {
		return err
	}

	if m.ProjectID == 0 {
		m.ProjectID = existing.ProjectID
	}
	if m.ProjectID != existing.ProjectID {
		return InvalidFieldError([]string{"project_id"})
	}

	m.HexColor = utils.NormalizeHex(m.HexColor)

	_, err = s.ID(m.ID).
		Cols("name", "milestone_date", "hex_color").
		Update(m)
	if err != nil {
		return err
	}

	if err := m.syncUsers(s, m.Users); err != nil {
		return err
	}

	if err := updateProjectLastUpdated(s, &Project{ID: m.ProjectID}); err != nil {
		return err
	}

	return m.ReadOne(s, nil)
}

// Delete deletes a milestone and unassigns it from all tasks.
func (m *Milestone) Delete(s *xorm.Session, _ web.Auth) error {
	existing, err := getMilestoneSimpleByID(s, m.ID)
	if err != nil {
		return err
	}
	if m.ProjectID == 0 {
		m.ProjectID = existing.ProjectID
	}

	_, err = s.Where("milestone_id = ?", m.ID).
		Cols("milestone_id").
		Update(&Task{MilestoneID: 0})
	if err != nil {
		return err
	}

	_, err = s.Where("milestone_id = ?", m.ID).Delete(&MilestoneUser{})
	if err != nil {
		return err
	}

	_, err = s.ID(m.ID).Delete(&Milestone{})
	if err != nil {
		return err
	}

	return updateProjectLastUpdated(s, &Project{ID: m.ProjectID})
}

// ReadAll returns all milestones for a project.
func (m *Milestone) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	project := &Project{ID: m.ProjectID}
	canRead, _, err := project.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canRead {
		return nil, 0, 0, ErrNeedToHaveProjectReadAccess{ProjectID: m.ProjectID, UserID: a.GetID()}
	}

	limit, start := getLimitFromPageIndex(page, perPage)
	milestones := []*Milestone{}

	projectIDs := []int64{m.ProjectID}
	if m.IncludeParents {
		projectIDs, err = getProjectAndParentProjectIDs(s, m.ProjectID)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	query := s.In("project_id", projectIDs).
		Where(db.ILIKE("name", search)).
		OrderBy("milestone_date ASC, id ASC")
	if limit > 0 {
		query = query.Limit(limit, start)
	}

	if err = query.Find(&milestones); err != nil {
		return nil, 0, 0, err
	}

	if err = addUsersToMilestones(s, milestones); err != nil {
		return nil, 0, 0, err
	}

	totalItems, err = s.In("project_id", projectIDs).
		Where(db.ILIKE("name", search)).
		Count(&Milestone{})
	if err != nil {
		return nil, 0, 0, err
	}

	return milestones, len(milestones), totalItems, nil
}

// ReadOne returns a milestone.
func (m *Milestone) ReadOne(s *xorm.Session, _ web.Auth) error {
	milestone, err := getMilestoneSimpleByID(s, m.ID)
	if err != nil {
		return err
	}

	*m = *milestone
	return addUsersToMilestones(s, []*Milestone{m})
}
