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

package websocket

import (
	"encoding/json"
	"fmt"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/ThreeDotsLabs/watermill/message"
	"xorm.io/xorm"
)

// NotificationListener pushes new notifications to WebSocket clients.
type NotificationListener struct{}

type kanbanChangedPayload struct {
	ProjectID int64 `json:"project_id"`
	ViewID    int64 `json:"view_id"`
}

type taskChangedPayload struct {
	ProjectID int64 `json:"project_id"`
	TaskID    int64 `json:"task_id"`
}

type TaskKanbanListener struct{}

type TaskPositionsKanbanListener struct{}

type KanbanViewChangedListener struct{}

// Name returns the listener name.
func (n *NotificationListener) Name() string {
	return "websocket.notification.push"
}

// Handle processes a notification created event, reloads the notification
// from the database (to get accurate timestamps), and pushes it to the
// relevant WebSocket connections.
func (n *NotificationListener) Handle(msg *message.Message) error {
	var event notifications.NotificationCreatedEvent
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		return err
	}

	hub := GetHub()
	if hub == nil {
		log.Warningf("WebSocket: hub not initialized, skipping notification push")
		return nil
	}

	s := db.NewSession()
	defer s.Close()

	dbNotification, err := notifications.GetNotificationByID(s, event.NotificationID)
	if err != nil {
		log.Errorf("WebSocket: failed to load notification %d: %v", event.NotificationID, err)
		return nil
	}
	if dbNotification == nil {
		log.Warningf("WebSocket: notification %d not found, skipping push", event.NotificationID)
		return nil
	}

	hub.PublishForUser(event.UserID, "notification.created", dbNotification)
	return nil
}

func kanbanChangedEventName(projectID, viewID int64) string {
	return fmt.Sprintf("project.%d.view.%d.kanban.changed", projectID, viewID)
}

func taskChangedEventName(projectID, taskID int64) string {
	return fmt.Sprintf("project.%d.task.%d.changed", projectID, taskID)
}

func publishKanbanChanged(s *xorm.Session, projectID, viewID int64) error {
	users, hub, err := getProjectUsersForPublish(s, projectID)
	if err != nil {
		return err
	}
	if hub == nil {
		return nil
	}
	eventName := kanbanChangedEventName(projectID, viewID)
	payload := kanbanChangedPayload{
		ProjectID: projectID,
		ViewID:    viewID,
	}

	for _, u := range users {
		hub.PublishForUser(u.ID, eventName, payload)
	}

	return nil
}

func publishTaskChanged(s *xorm.Session, projectID, taskID int64) error {
	users, hub, err := getProjectUsersForPublish(s, projectID)
	if err != nil {
		return err
	}
	if hub == nil {
		return nil
	}

	eventName := taskChangedEventName(projectID, taskID)
	payload := taskChangedPayload{
		ProjectID: projectID,
		TaskID:    taskID,
	}

	for _, u := range users {
		hub.PublishForUser(u.ID, eventName, payload)
	}

	return nil
}

func getProjectUsersForPublish(s *xorm.Session, projectID int64) ([]*user.User, *Hub, error) {
	hub := GetHub()
	if hub == nil {
		log.Warningf("WebSocket: hub not initialized, skipping realtime push")
		return nil, nil, nil
	}

	project, err := models.GetProjectSimpleByID(s, projectID)
	if err != nil {
		return nil, nil, err
	}

	users, err := models.ListUsersFromProject(s, project, &user.User{ID: project.OwnerID}, "")
	if err != nil {
		return nil, nil, err
	}

	return users, hub, nil
}

func publishKanbanChangedForTaskProject(s *xorm.Session, projectID int64) error {
	views := []*models.ProjectView{}
	err := s.
		Where("project_id = ? AND view_kind = ?", projectID, models.ProjectViewKindKanban).
		Find(&views)
	if err != nil {
		return err
	}

	for _, view := range views {
		err = publishKanbanChanged(s, projectID, view.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *TaskKanbanListener) Name() string {
	return "websocket.kanban.task.push"
}

func (n *TaskKanbanListener) Handle(msg *message.Message) error {
	var event struct {
		Task *models.Task `json:"task"`
	}
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		return err
	}
	if event.Task == nil || event.Task.ProjectID == 0 {
		return nil
	}

	s := db.NewSession()
	defer s.Close()

	if err := publishKanbanChangedForTaskProject(s, event.Task.ProjectID); err != nil {
		log.Errorf("WebSocket: failed to publish kanban task change for project %d: %v", event.Task.ProjectID, err)
	}
	if err := publishTaskChanged(s, event.Task.ProjectID, event.Task.ID); err != nil {
		log.Errorf("WebSocket: failed to publish task change for task %d: %v", event.Task.ID, err)
	}

	return nil
}

func (n *TaskPositionsKanbanListener) Name() string {
	return "websocket.kanban.positions.push"
}

func (n *TaskPositionsKanbanListener) Handle(msg *message.Message) error {
	var event models.TaskPositionsRecalculatedEvent
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	viewIDs := make(map[int64]bool, len(event.NewTaskPositions))
	for _, position := range event.NewTaskPositions {
		if position == nil {
			continue
		}
		viewIDs[position.ProjectViewID] = true
	}

	for viewID := range viewIDs {
		view, err := models.GetProjectViewByID(s, viewID)
		if err != nil {
			log.Errorf("WebSocket: failed to load project view %d for kanban refresh: %v", viewID, err)
			continue
		}

		if view.ViewKind != models.ProjectViewKindKanban {
			continue
		}

		if err := publishKanbanChanged(s, view.ProjectID, view.ID); err != nil {
			log.Errorf("WebSocket: failed to publish kanban position change for view %d: %v", view.ID, err)
		}
	}

	return nil
}

func (n *KanbanViewChangedListener) Name() string {
	return "websocket.kanban.view.push"
}

func (n *KanbanViewChangedListener) Handle(msg *message.Message) error {
	var event models.KanbanViewChangedEvent
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		return err
	}
	if event.ProjectID == 0 || event.ViewID == 0 {
		return nil
	}

	s := db.NewSession()
	defer s.Close()

	if err := publishKanbanChanged(s, event.ProjectID, event.ViewID); err != nil {
		log.Errorf("WebSocket: failed to publish kanban view change for project %d view %d: %v", event.ProjectID, event.ViewID, err)
	}

	return nil
}

// RegisterListeners registers WebSocket event listeners.
func RegisterListeners() {
	events.RegisterListener(
		(&notifications.NotificationCreatedEvent{}).Name(),
		&NotificationListener{},
	)
	events.RegisterListener(
		(&models.TaskCreatedEvent{}).Name(),
		&TaskKanbanListener{},
	)
	events.RegisterListener(
		(&models.TaskUpdatedEvent{}).Name(),
		&TaskKanbanListener{},
	)
	events.RegisterListener(
		(&models.TaskDeletedEvent{}).Name(),
		&TaskKanbanListener{},
	)
	events.RegisterListener(
		(&models.TaskPositionsRecalculatedEvent{}).Name(),
		&TaskPositionsKanbanListener{},
	)
	events.RegisterListener(
		(&models.KanbanViewChangedEvent{}).Name(),
		&KanbanViewChangedListener{},
	)
}
