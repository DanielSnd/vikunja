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
	"fmt"
	"time"

	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

type VisionBoard struct {
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"board"`

	ProjectID int64 `xorm:"bigint not null index" json:"project_id" param:"project"`
	TaskID    int64 `xorm:"bigint not null unique index" json:"task_id" query:"task_id"`

	Title string `xorm:"varchar(255) not null" json:"title" valid:"runelength(0|250)"`

	ViewportX    float64 `xorm:"double not null default 0" json:"viewport_x"`
	ViewportY    float64 `xorm:"double not null default 0" json:"viewport_y"`
	ViewportZoom float64 `xorm:"double not null default 1" json:"viewport_zoom"`

	TaskTitle string             `xorm:"-" json:"task_title"`
	Nodes     []*VisionBoardNode `xorm:"-" json:"nodes,omitempty"`
	Edges     []*VisionBoardEdge `xorm:"-" json:"edges,omitempty"`

	CreatedByID int64 `xorm:"bigint not null" json:"created_by_id"`

	Updated time.Time `xorm:"updated not null" json:"updated"`
	Created time.Time `xorm:"created not null" json:"created"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (*VisionBoard) TableName() string {
	return "vision_boards"
}

func (vb *VisionBoard) ensureTaskForProject(s *xorm.Session) (task Task, err error) {
	task, err = GetTaskSimple(s, &Task{ID: vb.TaskID})
	if err != nil {
		return
	}

	if task.ProjectID != vb.ProjectID {
		return Task{}, ErrVisionBoardTaskDoesNotBelongToProject{
			TaskID:    vb.TaskID,
			ProjectID: vb.ProjectID,
		}
	}

	return
}

func (vb *VisionBoard) populateTaskTitles(s *xorm.Session, boards []*VisionBoard) error {
	if len(boards) == 0 {
		return nil
	}

	taskIDs := make([]int64, 0, len(boards))
	for _, board := range boards {
		taskIDs = append(taskIDs, board.TaskID)
	}

	tasks, err := GetTasksSimpleByIDs(s, taskIDs)
	if err != nil {
		return err
	}

	titles := make(map[int64]string, len(tasks))
	for _, task := range tasks {
		titles[task.ID] = task.Title
	}

	for _, board := range boards {
		board.TaskTitle = titles[board.TaskID]
		if board.Title == "" {
			board.Title = board.TaskTitle
		}
	}

	return nil
}

func getVisionBoardByIDAndProject(s *xorm.Session, id, projectID int64) (board *VisionBoard, err error) {
	board = &VisionBoard{}
	exists, err := s.Where("id = ? AND project_id = ?", id, projectID).Get(board)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrVisionBoardDoesNotExist{VisionBoardID: id}
	}

	return board, nil
}

// ReadAll gets all vision boards for a project.
func (vb *VisionBoard) ReadAll(s *xorm.Session, a web.Auth, search string, _ int, _ int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	project := &Project{ID: vb.ProjectID}
	can, _, err := project.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !can {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	query := s.Where("project_id = ?", vb.ProjectID)
	if vb.TaskID > 0 {
		query = query.And("task_id = ?", vb.TaskID)
	}

	boards := []*VisionBoard{}
	if search != "" {
		taskIDs := make([]int64, 0)
		tasks := []*Task{}
		err = s.Where("project_id = ? AND title LIKE ?", vb.ProjectID, "%"+search+"%").Find(&tasks)
		if err != nil {
			return nil, 0, 0, err
		}
		for _, task := range tasks {
			taskIDs = append(taskIDs, task.ID)
		}
		if len(taskIDs) == 0 {
			return []*VisionBoard{}, 0, 0, nil
		}
		query = query.In("task_id", taskIDs)
	}

	err = query.OrderBy("created desc").Find(&boards)
	if err != nil {
		return nil, 0, 0, err
	}

	if err = vb.populateTaskTitles(s, boards); err != nil {
		return nil, 0, 0, err
	}

	return boards, len(boards), int64(len(boards)), nil
}

// ReadOne gets a single vision board.
func (vb *VisionBoard) ReadOne(s *xorm.Session, _ web.Auth) error {
	board, err := getVisionBoardByIDAndProject(s, vb.ID, vb.ProjectID)
	if err != nil {
		return err
	}

	*vb = *board
	if err = vb.populateTaskTitles(s, []*VisionBoard{vb}); err != nil {
		return err
	}

	vb.Nodes, err = getVisionBoardNodesForBoard(s, vb.ID)
	if err != nil {
		return err
	}

	vb.Edges, err = getVisionBoardEdgesForBoard(s, vb.ID)
	return err
}

// Create creates a new vision board for a task.
func (vb *VisionBoard) Create(s *xorm.Session, a web.Auth) error {
	task, err := vb.ensureTaskForProject(s)
	if err != nil {
		return err
	}

	exists, err := s.Where("task_id = ?", vb.TaskID).Exist(&VisionBoard{})
	if err != nil {
		return err
	}
	if exists {
		return ErrVisionBoardAlreadyExistsForTask{TaskID: vb.TaskID}
	}

	if vb.Title == "" {
		vb.Title = task.Title
	}
	if vb.ViewportZoom == 0 {
		vb.ViewportZoom = 1
	}
	vb.ID = 0
	vb.CreatedByID = a.GetID()

	_, err = s.Insert(vb)
	if err != nil {
		return err
	}

	vb.TaskTitle = task.Title
	return nil
}

// Update updates vision board metadata.
func (vb *VisionBoard) Update(s *xorm.Session, _ web.Auth) error {
	current, err := getVisionBoardByIDAndProject(s, vb.ID, vb.ProjectID)
	if err != nil {
		return err
	}

	update := &VisionBoard{
		Title:        current.Title,
		ViewportX:    current.ViewportX,
		ViewportY:    current.ViewportY,
		ViewportZoom: current.ViewportZoom,
	}

	if vb.Title != "" {
		update.Title = vb.Title
	}
	update.ViewportX = vb.ViewportX
	update.ViewportY = vb.ViewportY
	if vb.ViewportZoom != 0 {
		update.ViewportZoom = vb.ViewportZoom
	}

	_, err = s.ID(vb.ID).Cols("title", "viewport_x", "viewport_y", "viewport_zoom").Update(update)
	if err != nil {
		return err
	}

	updated, err := getVisionBoardByIDAndProject(s, vb.ID, vb.ProjectID)
	if err != nil {
		return err
	}
	*vb = *updated
	return vb.populateTaskTitles(s, []*VisionBoard{vb})
}

// Delete deletes a vision board.
func (vb *VisionBoard) Delete(s *xorm.Session, _ web.Auth) error {
	_, err := s.Where("id = ? AND project_id = ?", vb.ID, vb.ProjectID).Delete(&VisionBoard{})
	return err
}

func (vb *VisionBoard) String() string {
	return fmt.Sprintf("VisionBoard{ID:%d, ProjectID:%d, TaskID:%d}", vb.ID, vb.ProjectID, vb.TaskID)
}
