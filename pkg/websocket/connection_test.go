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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnectionSubscribeUnsubscribe(t *testing.T) {
	conn := &Connection{
		userID:        1,
		authenticated: true,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}

	conn.Subscribe("notification.created")
	assert.True(t, conn.IsSubscribed("notification.created"))

	conn.Unsubscribe("notification.created")
	assert.False(t, conn.IsSubscribed("notification.created"))
}

func TestConnectionIsSubscribedReturnsFalseForUnknownEvent(t *testing.T) {
	conn := &Connection{
		userID:        1,
		authenticated: true,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}

	assert.False(t, conn.IsSubscribed("something"))
}

func TestConnectionAcceptsValidEvent(t *testing.T) {
	hub := NewHub()
	conn := &Connection{
		hub:           hub,
		userID:        1,
		authenticated: true,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}
	hub.Register(conn)

	conn.handleMessage(context.Background(), IncomingMessage{Action: ActionSubscribe, Event: "notification.created"})

	assert.True(t, conn.IsSubscribed("notification.created"))
}

func TestConnectionAcceptsValidKanbanEvent(t *testing.T) {
	hub := NewHub()
	conn := &Connection{
		hub:           hub,
		userID:        1,
		authenticated: true,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}
	hub.Register(conn)

	event := "project.12.view.34.kanban.changed"
	conn.handleMessage(context.Background(), IncomingMessage{Action: ActionSubscribe, Event: event})

	assert.True(t, conn.IsSubscribed(event))
}

func TestConnectionAcceptsValidVisionBoardEvents(t *testing.T) {
	hub := NewHub()
	conn := &Connection{
		hub:           hub,
		userID:        1,
		authenticated: true,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}
	hub.Register(conn)

	changedEvent := "project.12.board.34.changed"
	presenceEvent := "project.12.board.34.presence"
	conn.handleMessage(context.Background(), IncomingMessage{Action: ActionSubscribe, Event: changedEvent})
	conn.handleMessage(context.Background(), IncomingMessage{Action: ActionSubscribe, Event: presenceEvent})

	assert.True(t, conn.IsSubscribed(changedEvent))
	assert.True(t, conn.IsSubscribed(presenceEvent))
}

func TestConnectionRejectsInvalidEvent(t *testing.T) {
	conn := &Connection{
		userID:        1,
		authenticated: true,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}

	conn.handleMessage(context.Background(), IncomingMessage{Action: ActionSubscribe, Event: "notifications"})

	msg := <-conn.send
	assert.Equal(t, "invalid_event", msg.Error)
	assert.False(t, conn.IsSubscribed("notifications"))
}

func TestConnectionRejectsActionsBeforeAuth(t *testing.T) {
	conn := &Connection{
		userID:        0, // not authenticated
		authenticated: false,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}

	// Try to subscribe before auth - should be rejected
	conn.handleMessage(context.Background(), IncomingMessage{Action: ActionSubscribe, Event: "notification.created"})

	// Should have sent an error
	msg := <-conn.send
	assert.Equal(t, "auth_required", msg.Error)
	assert.False(t, conn.IsSubscribed("notification.created"))
}

func TestConnectionRejectsPublishBeforeAuth(t *testing.T) {
	conn := &Connection{
		userID:        0,
		authenticated: false,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}

	conn.handleMessage(context.Background(), IncomingMessage{
		Action: ActionPublish,
		Event:  "project.12.board.34.presence",
		Data:   map[string]any{"sessionId": "abc"},
	})

	msg := <-conn.send
	assert.Equal(t, "auth_required", msg.Error)
}

func TestParseBoardEvent(t *testing.T) {
	projectID, boardID, kind, ok := parseBoardEvent("project.12.board.34.changed")
	assert.True(t, ok)
	assert.EqualValues(t, 12, projectID)
	assert.EqualValues(t, 34, boardID)
	assert.Equal(t, "changed", kind)
}
