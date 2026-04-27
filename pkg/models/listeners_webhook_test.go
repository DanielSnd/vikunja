package models

import "testing"

import "github.com/stretchr/testify/assert"

func TestShouldSuppressWebhookEvent(t *testing.T) {
	t.Run("non task updated event", func(t *testing.T) {
		assert.False(t, shouldSuppressWebhookEvent("task.created", map[string]interface{}{
			"change_summary": "Task position updated",
		}))
	})

	t.Run("task position update", func(t *testing.T) {
		assert.True(t, shouldSuppressWebhookEvent("task.updated", map[string]interface{}{
			"change_summary": "Task position updated",
		}))
	})

	t.Run("time tracking update", func(t *testing.T) {
		assert.True(t, shouldSuppressWebhookEvent("task.updated", map[string]interface{}{
			"change_summary": "Time tracking updated",
		}))
	})

	t.Run("time tracking timer update", func(t *testing.T) {
		assert.True(t, shouldSuppressWebhookEvent("task.updated", map[string]interface{}{
			"change_summary": "Time tracking timer started",
		}))
	})

	t.Run("other task update", func(t *testing.T) {
		assert.False(t, shouldSuppressWebhookEvent("task.updated", map[string]interface{}{
			"change_summary": "Marked done",
		}))
	})
}
