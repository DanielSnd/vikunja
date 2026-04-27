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

	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

type VisionBoardNodeKind string

const (
	VisionBoardNodeKindText      VisionBoardNodeKind = "text"
	VisionBoardNodeKindImage     VisionBoardNodeKind = "image"
	VisionBoardNodeKindVideo     VisionBoardNodeKind = "video"
	VisionBoardNodeKindCard      VisionBoardNodeKind = "card"
	VisionBoardNodeKindContainer VisionBoardNodeKind = "container"
)

type VisionBoardNode struct {
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"node"`

	ProjectID int64 `xorm:"-" json:"project_id" param:"project"`
	BoardID   int64 `xorm:"bigint not null index" json:"board_id" param:"board"`

	ParentNodeID int64 `xorm:"bigint null index" json:"parent_node_id"`
	TaskID       int64 `xorm:"bigint null index" json:"task_id"`
	AttachmentID int64 `xorm:"bigint null index" json:"attachment_id"`

	Kind VisionBoardNodeKind `xorm:"varchar(32) not null" json:"kind"`

	Title   string `xorm:"varchar(255) not null default ''" json:"title"`
	Content string `xorm:"text null" json:"content"`
	URL     string `xorm:"text null" json:"url"`

	X      float64 `xorm:"double not null default 0" json:"x"`
	Y      float64 `xorm:"double not null default 0" json:"y"`
	Width  float64 `xorm:"double not null default 240" json:"width"`
	Height float64 `xorm:"double not null default 160" json:"height"`

	Color  string `xorm:"varchar(32) not null default ''" json:"color"`
	ZIndex int64  `xorm:"bigint not null default 0" json:"z_index"`

	Version int64 `xorm:"bigint not null default 1" json:"version"`

	CreatedByID int64 `xorm:"bigint not null" json:"created_by_id"`
	UpdatedByID int64 `xorm:"bigint not null" json:"updated_by_id"`

	Updated time.Time `xorm:"updated not null" json:"updated"`
	Created time.Time `xorm:"created not null" json:"created"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (*VisionBoardNode) TableName() string {
	return "vision_board_nodes"
}

func (n *VisionBoardNode) resolveBoard(s *xorm.Session) (*VisionBoard, error) {
	if n.BoardID == 0 && n.ID != 0 {
		current, err := getVisionBoardNodeByIDAndBoard(s, n.ID, 0)
		if err != nil {
			return nil, err
		}
		n.BoardID = current.BoardID
	}

	board, err := getVisionBoardByIDAndProject(s, n.BoardID, n.ProjectID)
	if err == nil {
		return board, nil
	}

	if n.ProjectID != 0 {
		return nil, err
	}

	board = &VisionBoard{}
	exists, getErr := s.ID(n.BoardID).Get(board)
	if getErr != nil {
		return nil, getErr
	}
	if !exists {
		return nil, ErrVisionBoardDoesNotExist{VisionBoardID: n.BoardID}
	}
	n.ProjectID = board.ProjectID
	return board, nil
}

func getVisionBoardNodeByIDAndBoard(s *xorm.Session, id, boardID int64) (*VisionBoardNode, error) {
	node := &VisionBoardNode{}
	query := s.Where("id = ?", id)
	if boardID > 0 {
		query = query.And("board_id = ?", boardID)
	}

	exists, err := query.Get(node)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrVisionBoardNodeDoesNotExist{VisionBoardNodeID: id}
	}

	return node, nil
}

func getVisionBoardNodesForBoard(s *xorm.Session, boardID int64) ([]*VisionBoardNode, error) {
	nodes := []*VisionBoardNode{}
	err := s.Where("board_id = ?", boardID).Asc("z_index").Asc("id").Find(&nodes)
	return nodes, err
}

func (n *VisionBoardNode) applyDefaults() {
	if n.Width > 0 && n.Height > 0 {
		return
	}

	switch n.Kind {
	case VisionBoardNodeKindContainer:
		if n.Width == 0 {
			n.Width = 320
		}
		if n.Height == 0 {
			n.Height = 240
		}
	case VisionBoardNodeKindCard:
		if n.Width == 0 {
			n.Width = 280
		}
		if n.Height == 0 {
			n.Height = 160
		}
	case VisionBoardNodeKindImage, VisionBoardNodeKindVideo:
		if n.Width == 0 {
			n.Width = 260
		}
		if n.Height == 0 {
			n.Height = 180
		}
	default:
		if n.Width == 0 {
			n.Width = 240
		}
		if n.Height == 0 {
			n.Height = 160
		}
	}
}

func (n *VisionBoardNode) validateTaskReference(s *xorm.Session) error {
	if n.TaskID == 0 {
		return nil
	}

	_, err := GetTaskSimple(s, &Task{ID: n.TaskID})
	return err
}

func (n *VisionBoardNode) ReadAll(s *xorm.Session, a web.Auth, _ string, _ int, _ int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	board, err := n.resolveBoard(s)
	if err != nil {
		return nil, 0, 0, err
	}

	can, _, err := board.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !can {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	nodes, err := getVisionBoardNodesForBoard(s, board.ID)
	if err != nil {
		return nil, 0, 0, err
	}

	return nodes, len(nodes), int64(len(nodes)), nil
}

func (n *VisionBoardNode) ReadOne(s *xorm.Session, _ web.Auth) error {
	current, err := getVisionBoardNodeByIDAndBoard(s, n.ID, n.BoardID)
	if err != nil {
		return err
	}
	*n = *current
	return nil
}

func (n *VisionBoardNode) Create(s *xorm.Session, a web.Auth) error {
	board, err := n.resolveBoard(s)
	if err != nil {
		return err
	}
	n.BoardID = board.ID
	n.ProjectID = board.ProjectID

	if err = n.validateTaskReference(s); err != nil {
		return err
	}

	n.applyDefaults()
	n.ID = 0
	n.Version = 1
	n.CreatedByID = a.GetID()
	n.UpdatedByID = a.GetID()

	_, err = s.Insert(n)
	return err
}

func (n *VisionBoardNode) Update(s *xorm.Session, a web.Auth) error {
	current, err := getVisionBoardNodeByIDAndBoard(s, n.ID, n.BoardID)
	if err != nil {
		return err
	}

	current.ProjectID = n.ProjectID
	if err = n.validateTaskReference(s); err != nil {
		return err
	}

	current.ParentNodeID = n.ParentNodeID
	if n.TaskID != 0 || current.Kind == VisionBoardNodeKindCard {
		current.TaskID = n.TaskID
	}
	current.AttachmentID = n.AttachmentID
	if n.Kind != "" {
		current.Kind = n.Kind
	}
	current.Title = n.Title
	current.Content = n.Content
	current.URL = n.URL
	current.X = n.X
	current.Y = n.Y
	current.Width = n.Width
	current.Height = n.Height
	current.Color = n.Color
	current.ZIndex = n.ZIndex
	current.Version++
	current.UpdatedByID = a.GetID()
	current.applyDefaults()

	_, err = s.ID(current.ID).Cols(
		"parent_node_id",
		"task_id",
		"attachment_id",
		"kind",
		"title",
		"content",
		"url",
		"x",
		"y",
		"width",
		"height",
		"color",
		"z_index",
		"version",
		"updated_by_id",
	).Update(current)
	if err != nil {
		return err
	}

	*n = *current
	return nil
}

func (n *VisionBoardNode) Delete(s *xorm.Session, _ web.Auth) error {
	_, err := s.Where("id = ? AND board_id = ?", n.ID, n.BoardID).Delete(&VisionBoardNode{})
	return err
}
