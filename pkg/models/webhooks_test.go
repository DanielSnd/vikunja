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
	"encoding/json"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalWebhookPayloadDiscord(t *testing.T) {
	config.ServicePublicURL.Set("https://vikunja.example/")

	payload, err := marshalWebhookPayload("https://discord.com/api/webhooks/123/token", &WebhookPayload{
		EventName: "task.updated",
		Time:      time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC),
		Data: map[string]interface{}{
			"change_summary": "Status changed: Started -> For Review",
			"task": map[string]interface{}{
				"id":          42,
				"title":       "Discord webhook test",
				"description": "Updated from a unit test",
				"identifier":  "TEST-42",
			},
			"project": map[string]interface{}{
				"title": "Test Project",
			},
			"doer": map[string]interface{}{
				"username": "demo",
			},
			"bucket": map[string]interface{}{
				"title": "In Progress",
			},
		},
	})
	require.NoError(t, err)

	var discordPayload map[string]interface{}
	require.NoError(t, json.Unmarshal(payload, &discordPayload))

	assert.Equal(t, "Task updated: **Discord webhook test**", discordPayload["content"])

	embeds, ok := discordPayload["embeds"].([]interface{})
	require.True(t, ok)
	require.Len(t, embeds, 1)

	embed, ok := embeds[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Discord webhook test", embed["title"])
	assert.Equal(t, "Status changed: Started -> For Review", embed["description"])
	assert.Equal(t, "https://vikunja.example/tasks/42", embed["url"])
	assert.Equal(t, "2026-04-24T12:00:00Z", embed["timestamp"])

	fields, ok := embed["fields"].([]interface{})
	require.True(t, ok)

	fieldValues := map[string]string{}
	for _, field := range fields {
		f, ok := field.(map[string]interface{})
		require.True(t, ok)
		fieldValues[f["name"].(string)] = f["value"].(string)
	}

	assert.Equal(t, "Status changed: Started -> For Review", fieldValues["Update"])
	assert.Equal(t, "task.updated", fieldValues["Event"])
	assert.Equal(t, "Test Project", fieldValues["Project"])
	assert.Equal(t, "In Progress", fieldValues["Bucket"])
	assert.Equal(t, "demo", fieldValues["By"])
	assert.Equal(t, "TEST-42", fieldValues["Task"])
}

func TestMarshalWebhookPayloadGeneric(t *testing.T) {
	payload, err := marshalWebhookPayload("https://example.com/webhook", &WebhookPayload{
		EventName: "task.updated",
		Time:      time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC),
		Data: map[string]interface{}{
			"task": map[string]interface{}{
				"title": "Generic webhook test",
			},
		},
	})
	require.NoError(t, err)

	var genericPayload map[string]interface{}
	require.NoError(t, json.Unmarshal(payload, &genericPayload))

	assert.Equal(t, "task.updated", genericPayload["event_name"])
	assert.NotContains(t, genericPayload, "content")
	assert.NotContains(t, genericPayload, "embeds")
}
