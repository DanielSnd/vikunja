package webtests

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	apiv1 "code.vikunja.io/api/pkg/routes/api/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskTimeTrackingReportProjectGrouping(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	s := db.NewSession()
	_, err = s.Insert(&models.TaskTimeTracking{
		TaskID:    2,
		UserID:    1,
		TimeSpent: 3600,
		TrackedAt: time.Date(2024, 1, 11, 12, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	_, err = s.Insert(&models.TaskTimeTracking{
		TaskID:    3,
		UserID:    1,
		TimeSpent: 900,
		TrackedAt: time.Date(2024, 1, 11, 13, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	require.NoError(t, s.Close())

	query := url.Values{
		"date_from": {"2024-01-10"},
		"date_to":   {"2024-01-11"},
		"group_by":  {"project"},
	}

	c, rec := createRequest(e, http.MethodGet, "", query, nil)
	addUserTokenToContext(t, &testuser1, c)

	err = apiv1.TaskTimeTrackingReport(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"group_by":"project"`)
	assert.Contains(t, rec.Body.String(), `"total_seconds":9000`)
	assert.Contains(t, rec.Body.String(), `"title":"Test1"`)
	assert.Contains(t, rec.Body.String(), `"daily_seconds":[4500,4500]`)
	assert.NotContains(t, rec.Body.String(), `28800`)
}

func TestTaskTimeTrackingReportLabelGrouping(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	s := db.NewSession()
	_, err = s.Insert(&models.TaskTimeTracking{
		TaskID:    2,
		UserID:    1,
		TimeSpent: 3600,
		TrackedAt: time.Date(2024, 1, 11, 12, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	_, err = s.Insert(&models.TaskTimeTracking{
		TaskID:    3,
		UserID:    1,
		TimeSpent: 900,
		TrackedAt: time.Date(2024, 1, 11, 13, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	require.NoError(t, s.Close())

	query := url.Values{
		"date_from": {"2024-01-10"},
		"date_to":   {"2024-01-11"},
		"group_by":  {"label"},
	}

	c, rec := createRequest(e, http.MethodGet, "", query, nil)
	addUserTokenToContext(t, &testuser1, c)

	err = apiv1.TaskTimeTrackingReport(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"group_by":"label"`)
	assert.Contains(t, rec.Body.String(), `"title":"Label #4 - visible via other task"`)
	assert.Contains(t, rec.Body.String(), `"total_seconds":8100`)
	assert.Contains(t, rec.Body.String(), `"title":"No Label"`)
	assert.Contains(t, rec.Body.String(), `"total_seconds":900`)
}
