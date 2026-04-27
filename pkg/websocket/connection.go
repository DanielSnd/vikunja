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
	"encoding/json"
	"regexp"
	"strconv"
	"sync"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/coder/websocket"
)

const (
	writeTimeout = 10 * time.Second
	pingInterval = 30 * time.Second
	authTimeout  = 30 * time.Second
	sendBufSize  = 64
)

// Connection wraps a single WebSocket connection.
type Connection struct {
	ws  *websocket.Conn
	hub *Hub

	mu            sync.RWMutex
	userID        int64
	authenticated bool
	subscriptions map[string]bool

	send chan OutgoingMessage
}

type publishedEventEnvelope struct {
	Sender publishedEventSender `json:"sender"`
	Data   any                  `json:"data,omitempty"`
}

type publishedEventSender struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

// NewConnection creates a new unauthenticated Connection.
func NewConnection(ws *websocket.Conn, hub *Hub) *Connection {
	return &Connection{
		ws:            ws,
		hub:           hub,
		authenticated: false,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, sendBufSize),
	}
}

// Subscribe adds an event subscription.
func (c *Connection) Subscribe(event string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subscriptions[event] = true
}

// Unsubscribe removes an event subscription.
func (c *Connection) Unsubscribe(event string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subscriptions, event)
}

// IsSubscribed checks if the connection is subscribed to an event.
func (c *Connection) IsSubscribed(event string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.subscriptions[event]
}

// IsAuthenticated returns whether the connection is authenticated.
func (c *Connection) IsAuthenticated() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.authenticated
}

// UserID returns the authenticated user's ID.
func (c *Connection) UserID() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.userID
}

// ReadLoop reads messages from the WebSocket and handles auth/subscribe/unsubscribe.
func (c *Connection) ReadLoop(ctx context.Context, cancel context.CancelFunc) {
	defer func() {
		cancel()
		if c.IsAuthenticated() {
			c.hub.Unregister(c)
		}
		c.ws.Close(websocket.StatusNormalClosure, "")
	}()

	// Close the connection if auth doesn't happen within the timeout
	authTimer := time.AfterFunc(authTimeout, func() {
		if !c.IsAuthenticated() {
			log.Debugf("WebSocket: closing unauthenticated connection after timeout")
			c.ws.Close(websocket.StatusPolicyViolation, "auth timeout")
		}
	})
	defer authTimer.Stop()

	for {
		_, data, err := c.ws.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				log.Debugf("WebSocket: connection closed normally for user %d", c.UserID())
			} else {
				log.Debugf("WebSocket: read error for user %d: %v", c.UserID(), err)
			}
			return
		}

		var msg IncomingMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Warningf("WebSocket: invalid message: %v", err)
			continue
		}

		if !c.handleMessage(ctx, msg) {
			return // close connection
		}
	}
}

// handleMessage processes an incoming message. Returns false if connection should be closed.
func (c *Connection) handleMessage(ctx context.Context, msg IncomingMessage) bool {
	switch msg.Action {
	case ActionAuth:
		return c.handleAuth(ctx, msg.Token)
	case ActionSubscribe:
		if !c.IsAuthenticated() {
			c.sendError("auth_required", "")
			return true
		}
		if !isValidEvent(msg.Event) {
			c.sendError("invalid_event", msg.Event)
			return true
		}
		c.Subscribe(msg.Event)
		log.Debugf("WebSocket: user %d subscribed to %s", c.UserID(), msg.Event)
	case ActionUnsubscribe:
		if !c.IsAuthenticated() {
			c.sendError("auth_required", "")
			return true
		}
		c.Unsubscribe(msg.Event)
		log.Debugf("WebSocket: user %d unsubscribed from %s", c.UserID(), msg.Event)
	case ActionPublish:
		if !c.IsAuthenticated() {
			c.sendError("auth_required", "")
			return true
		}
		if !isValidEvent(msg.Event) {
			c.sendError("invalid_event", msg.Event)
			return true
		}
		if err := c.publishEvent(msg.Event, msg.Data); err != nil {
			log.Warningf("WebSocket: user %d cannot publish %s: %v", c.UserID(), msg.Event, err)
			c.sendError("forbidden", msg.Event)
		}
	default:
		log.Warningf("WebSocket: unknown action %q", msg.Action)
	}
	return true
}

func (c *Connection) handleAuth(ctx context.Context, token string) bool {
	if c.IsAuthenticated() {
		c.sendError("already_authenticated", "")
		return true
	}

	userID, err := auth.GetUserIDFromToken(token)
	if err != nil {
		log.Debugf("WebSocket: auth failed: %v", err)
		// Write the error directly to the websocket since ReadLoop will close the
		// connection immediately after we return false, before WriteLoop can drain the channel.
		c.writeMessageDirect(ctx, OutgoingMessage{Error: "invalid_token"})
		return false
	}

	c.mu.Lock()
	c.userID = userID
	c.authenticated = true
	c.mu.Unlock()

	c.hub.Register(c)

	// Send auth success
	select {
	case c.send <- OutgoingMessage{Action: ActionAuthSuccess, Success: true}:
	default:
		log.Warningf("WebSocket: send buffer full for user %d", userID)
	}

	log.Debugf("WebSocket: user %d authenticated", userID)
	return true
}

// writeMessageDirect writes a message directly to the websocket, bypassing the send channel.
// Use this when the message must be sent before the connection is closed.
func (c *Connection) writeMessageDirect(ctx context.Context, msg OutgoingMessage) {
	writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	data, err := json.Marshal(msg)
	if err != nil {
		log.Errorf("WebSocket: marshal error: %v", err)
		return
	}
	if err := c.ws.Write(writeCtx, websocket.MessageText, data); err != nil {
		log.Debugf("WebSocket: direct write error: %v", err)
	}
}

func (c *Connection) sendError(errMsg, event string) {
	select {
	case c.send <- OutgoingMessage{Error: errMsg, Event: event}:
	default:
		log.Warningf("WebSocket: send buffer full, dropping error")
	}
}

// WriteLoop drains the send channel and writes messages to the WebSocket.
// It also sends periodic pings.
func (c *Connection) WriteLoop(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			data, err := json.Marshal(msg)
			if err != nil {
				cancel()
				log.Errorf("WebSocket: marshal error: %v", err)
				continue
			}
			err = c.ws.Write(writeCtx, websocket.MessageText, data)
			cancel()
			if err != nil {
				log.Debugf("WebSocket: write error for user %d: %v", c.UserID(), err)
				return
			}
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.ws.Ping(pingCtx)
			cancel()
			if err != nil {
				log.Debugf("WebSocket: ping error for user %d: %v", c.UserID(), err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// validEvents is the set of event names clients are allowed to subscribe to.
var validEvents = map[string]bool{
	"notification.created": true,
}

var dynamicEventPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^project\.-?\d+\.view\.\d+\.kanban\.changed$`),
	regexp.MustCompile(`^project\.-?\d+\.task\.\d+\.changed$`),
	regexp.MustCompile(`^project\.\d+\.board\.\d+\.changed$`),
	regexp.MustCompile(`^project\.\d+\.board\.\d+\.presence$`),
}

var boardEventPattern = regexp.MustCompile(`^project\.(\d+)\.board\.(\d+)\.(changed|presence)$`)

func isValidEvent(event string) bool {
	if validEvents[event] {
		return true
	}

	for _, pattern := range dynamicEventPatterns {
		if pattern.MatchString(event) {
			return true
		}
	}

	return false
}

func (c *Connection) publishEvent(event string, data any) error {
	projectID, boardID, kind, ok := parseBoardEvent(event)
	if !ok {
		return models.ErrGenericForbidden{}
	}

	s := db.NewSession()
	defer s.Close()

	board := &models.VisionBoard{
		ID:        boardID,
		ProjectID: projectID,
	}

	currentUser, err := user.GetUserByID(s, c.UserID())
	if err != nil {
		return err
	}

	switch kind {
	case "changed":
		can, err := board.CanUpdate(s, currentUser)
		if err != nil {
			return err
		}
		if !can {
			return models.ErrGenericForbidden{}
		}
	case "presence":
		can, _, err := board.CanRead(s, currentUser)
		if err != nil {
			return err
		}
		if !can {
			return models.ErrGenericForbidden{}
		}
	default:
		return models.ErrGenericForbidden{}
	}

	users, hub, err := getProjectUsersForPublish(s, projectID)
	if err != nil {
		return err
	}
	if hub == nil {
		return nil
	}

	payload := publishedEventEnvelope{
		Sender: publishedEventSender{
			ID:       currentUser.ID,
			Name:     currentUser.Name,
			Username: currentUser.Username,
		},
		Data: data,
	}

	for _, u := range users {
		hub.PublishForUser(u.ID, event, payload)
	}

	return nil
}

func parseBoardEvent(event string) (projectID, boardID int64, kind string, ok bool) {
	matches := boardEventPattern.FindStringSubmatch(event)
	if len(matches) != 4 {
		return 0, 0, "", false
	}

	projectID, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, 0, "", false
	}

	boardID, err = strconv.ParseInt(matches[2], 10, 64)
	if err != nil {
		return 0, 0, "", false
	}

	return projectID, boardID, matches[3], true
}
