package webtests

import (
	"net/http"
	"testing"

	apiv1 "code.vikunja.io/api/pkg/routes/api/v1"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskTimeTrackingTimer(t *testing.T) {
	t.Run("Current timer without running timer", func(t *testing.T) {
		rec, err := newTestRequestWithUser(t, http.MethodGet, apiv1.GetCurrentTaskTimeTrackingTimer, &testuser1, "", nil, nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "null\n", rec.Body.String())
	})

	t.Run("Current timer with existing timer", func(t *testing.T) {
		rec, err := newTestRequestWithUser(t, http.MethodGet, apiv1.GetCurrentTaskTimeTrackingTimer, &user.User{
			ID:       5,
			Username: "user5",
			Email:    "user5@example.com",
		}, "", nil, nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"task_id":14`)
		assert.Contains(t, rec.Body.String(), `"status":"snoozed"`)
	})
}
