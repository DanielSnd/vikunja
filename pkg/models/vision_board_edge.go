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

type VisionBoardEdge struct {
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"edge"`

	ProjectID int64 `xorm:"-" json:"project_id" param:"project"`
	BoardID   int64 `xorm:"bigint not null index" json:"board_id" param:"board"`

	SourceNodeID int64  `xorm:"bigint not null index" json:"source_node_id"`
	TargetNodeID int64  `xorm:"bigint not null index" json:"target_node_id"`
	SourceHandle string `xorm:"varchar(16) not null default ''" json:"source_handle"`
	TargetHandle string `xorm:"varchar(16) not null default ''" json:"target_handle"`

	Label string `xorm:"varchar(255) not null default ''" json:"label"`
	Color string `xorm:"varchar(32) not null default ''" json:"color"`

	Version int64 `xorm:"bigint not null default 1" json:"version"`

	Updated time.Time `xorm:"updated not null" json:"updated"`
	Created time.Time `xorm:"created not null" json:"created"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (*VisionBoardEdge) TableName() string {
	return "vision_board_edges"
}

func getVisionBoardEdgesForBoard(s *xorm.Session, boardID int64) ([]*VisionBoardEdge, error) {
	edges := []*VisionBoardEdge{}
	err := s.Where("board_id = ?", boardID).Asc("id").Find(&edges)
	return edges, err
}

func getVisionBoardEdgeByIDAndBoard(s *xorm.Session, id, boardID int64) (*VisionBoardEdge, error) {
	edge := &VisionBoardEdge{}
	query := s.Where("id = ?", id)
	if boardID > 0 {
		query = query.And("board_id = ?", boardID)
	}

	exists, err := query.Get(edge)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrVisionBoardEdgeDoesNotExist{VisionBoardEdgeID: id}
	}

	return edge, nil
}

func (e *VisionBoardEdge) resolveBoard(s *xorm.Session) (*VisionBoard, error) {
	if e.BoardID == 0 && e.ID != 0 {
		current, err := getVisionBoardEdgeByIDAndBoard(s, e.ID, 0)
		if err != nil {
			return nil, err
		}
		e.BoardID = current.BoardID
	}

	board, err := getVisionBoardByIDAndProject(s, e.BoardID, e.ProjectID)
	if err == nil {
		return board, nil
	}

	if e.ProjectID != 0 {
		return nil, err
	}

	board = &VisionBoard{}
	exists, getErr := s.ID(e.BoardID).Get(board)
	if getErr != nil {
		return nil, getErr
	}
	if !exists {
		return nil, ErrVisionBoardDoesNotExist{VisionBoardID: e.BoardID}
	}

	e.ProjectID = board.ProjectID
	return board, nil
}

func (e *VisionBoardEdge) validateNodes(s *xorm.Session) error {
	source, err := getVisionBoardNodeByIDAndBoard(s, e.SourceNodeID, e.BoardID)
	if err != nil {
		return err
	}
	if source.BoardID != e.BoardID {
		return ErrVisionBoardNodeDoesNotExist{VisionBoardNodeID: e.SourceNodeID}
	}

	target, err := getVisionBoardNodeByIDAndBoard(s, e.TargetNodeID, e.BoardID)
	if err != nil {
		return err
	}
	if target.BoardID != e.BoardID {
		return ErrVisionBoardNodeDoesNotExist{VisionBoardNodeID: e.TargetNodeID}
	}

	return nil
}

func (e *VisionBoardEdge) ReadAll(s *xorm.Session, a web.Auth, _ string, _ int, _ int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	board, err := e.resolveBoard(s)
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

	edges, err := getVisionBoardEdgesForBoard(s, board.ID)
	if err != nil {
		return nil, 0, 0, err
	}

	return edges, len(edges), int64(len(edges)), nil
}

func (e *VisionBoardEdge) ReadOne(s *xorm.Session, _ web.Auth) error {
	current, err := getVisionBoardEdgeByIDAndBoard(s, e.ID, e.BoardID)
	if err != nil {
		return err
	}

	*e = *current
	return nil
}

func (e *VisionBoardEdge) Create(s *xorm.Session, _ web.Auth) error {
	board, err := e.resolveBoard(s)
	if err != nil {
		return err
	}

	e.BoardID = board.ID
	e.ProjectID = board.ProjectID
	if err = e.validateNodes(s); err != nil {
		return err
	}

	e.ID = 0
	e.Version = 1
	_, err = s.Insert(e)
	return err
}

func (e *VisionBoardEdge) Update(s *xorm.Session, _ web.Auth) error {
	current, err := getVisionBoardEdgeByIDAndBoard(s, e.ID, e.BoardID)
	if err != nil {
		return err
	}

	current.SourceNodeID = e.SourceNodeID
	current.TargetNodeID = e.TargetNodeID
	current.SourceHandle = e.SourceHandle
	current.TargetHandle = e.TargetHandle
	current.Label = e.Label
	current.Color = e.Color
	current.Version++

	if err = current.validateNodes(s); err != nil {
		return err
	}

	_, err = s.ID(current.ID).Cols(
		"source_node_id",
		"target_node_id",
		"source_handle",
		"target_handle",
		"label",
		"color",
		"version",
	).Update(current)
	if err != nil {
		return err
	}

	*e = *current
	return nil
}

func (e *VisionBoardEdge) Delete(s *xorm.Session, _ web.Auth) error {
	_, err := s.Where("id = ? AND board_id = ?", e.ID, e.BoardID).Delete(&VisionBoardEdge{})
	return err
}
