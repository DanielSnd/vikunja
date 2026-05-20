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

package webtests

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserReportsFlow(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	authToken, err := auth.NewUserJWTAuthtoken(&testuser1, "test-session-id")
	require.NoError(t, err)

	createAppReq := httptest.NewRequest(http.MethodPut, "/api/v1/user-reports/applications", bytes.NewBufferString(`{"name":"Godot Build","project_id":11}`))
	createAppReq.Header.Set("Authorization", "Bearer "+authToken)
	createAppReq.Header.Set("Content-Type", "application/json")
	createAppRes := httptest.NewRecorder()
	e.ServeHTTP(createAppRes, createAppReq)
	require.Equal(t, http.StatusCreated, createAppRes.Code)

	var createAppResp models.UserReportApplicationCreateResponse
	require.NoError(t, json.Unmarshal(createAppRes.Body.Bytes(), &createAppResp))
	require.NotNil(t, createAppResp.Application)
	require.NotNil(t, createAppResp.DefaultReportToken)
	require.NotEmpty(t, createAppResp.AccessKey)
	require.NotEmpty(t, createAppResp.DefaultReportToken.Token)

	updateReq := httptest.NewRequest(http.MethodPost, "/api/v1/user-reports/applications/"+jsonNumber(createAppResp.Application.ID), bytes.NewBufferString(`{
		"name":"Godot Build",
		"project_id":11,
		"max_upload_size":1024,
		"bucket_id":11,
		"high_project_id":10,
		"high_bucket_id":10,
		"high_priority":4,
		"critical_priority":5,
		"low_priority":1
	}`))
	updateReq.Header.Set("Authorization", "Bearer "+authToken)
	updateReq.Header.Set("Content-Type", "application/json")
	updateRes := httptest.NewRecorder()
	e.ServeHTTP(updateRes, updateReq)
	require.Equal(t, http.StatusOK, updateRes.Code)

	createTokenReq := httptest.NewRequest(http.MethodPost, "/api/v1/user-reports/v1/create-report-token?accessKey="+createAppResp.AccessKey, bytes.NewBufferString(`{"label":"v1.2.3"}`))
	createTokenReq.Header.Set("Content-Type", "application/json")
	createTokenRes := httptest.NewRecorder()
	e.ServeHTTP(createTokenRes, createTokenReq)
	require.Equal(t, http.StatusOK, createTokenRes.Code)

	var createTokenResp models.UserReportCreateTokenResponse
	require.NoError(t, json.Unmarshal(createTokenRes.Body.Bytes(), &createTokenResp))
	require.True(t, createTokenResp.OK)
	require.NotEmpty(t, createTokenResp.Token)

	listTokensReq := httptest.NewRequest(http.MethodGet, "/api/v1/user-reports/applications/"+jsonNumber(createAppResp.Application.ID)+"/tokens", nil)
	listTokensReq.Header.Set("Authorization", "Bearer "+authToken)
	listTokensRes := httptest.NewRecorder()
	e.ServeHTTP(listTokensRes, listTokensReq)
	require.Equal(t, http.StatusOK, listTokensRes.Code)

	authCreateTokenReq := httptest.NewRequest(http.MethodPut, "/api/v1/user-reports/applications/"+jsonNumber(createAppResp.Application.ID)+"/tokens", bytes.NewBufferString(`{"label":"canary-build"}`))
	authCreateTokenReq.Header.Set("Authorization", "Bearer "+authToken)
	authCreateTokenReq.Header.Set("Content-Type", "application/json")
	authCreateTokenRes := httptest.NewRecorder()
	e.ServeHTTP(authCreateTokenRes, authCreateTokenReq)
	require.Equal(t, http.StatusCreated, authCreateTokenRes.Code)

	var authTokenResp models.UserReportToken
	require.NoError(t, json.Unmarshal(authCreateTokenRes.Body.Bytes(), &authTokenResp))
	require.NotEmpty(t, authTokenResp.Token)

	disableTokenReq := httptest.NewRequest(http.MethodPost, "/api/v1/user-reports/applications/"+jsonNumber(createAppResp.Application.ID)+"/tokens/"+jsonNumber(authTokenResp.ID), bytes.NewBufferString(`{"is_enabled":false}`))
	disableTokenReq.Header.Set("Authorization", "Bearer "+authToken)
	disableTokenReq.Header.Set("Content-Type", "application/json")
	disableTokenRes := httptest.NewRecorder()
	e.ServeHTTP(disableTokenRes, disableTokenReq)
	require.Equal(t, http.StatusOK, disableTokenRes.Code)

	disabledReportReq := httptest.NewRequest(http.MethodPost, "/api/v1/user-reports/v1/create-report?token="+authTokenResp.Token, bytes.NewBufferString(`{"content":"Should fail"}`))
	disabledReportReq.Header.Set("Content-Type", "application/json")
	disabledReportRes := httptest.NewRecorder()
	e.ServeHTTP(disabledReportRes, disabledReportReq)
	require.Equal(t, http.StatusUnauthorized, disabledReportRes.Code)

	reportReq := httptest.NewRequest(http.MethodPost, "/api/v1/user-reports/v1/create-report?token="+createTokenResp.Token, bytes.NewBufferString(`{
		"content":"Crash on launch\n\nThe app closes immediately after the splash screen.",
		"severity":"high",
		"fileNames":["logs.txt", "screenshot.png"],
		"userEmail":"player@example.com"
	}`))
	reportReq.Header.Set("Content-Type", "application/json")
	reportRes := httptest.NewRecorder()
	e.ServeHTTP(reportRes, reportReq)
	require.Equal(t, http.StatusOK, reportRes.Code)

	var reportResp models.UserReportCreateResponse
	require.NoError(t, json.Unmarshal(reportRes.Body.Bytes(), &reportResp))
	require.True(t, reportResp.OK)
	require.Len(t, reportResp.UploadURLs, 2)
	assert.NotEmpty(t, reportResp.CardID)
	assert.Equal(t, "logs.txt", reportResp.UploadURLs[0].FileName)
	assert.Equal(t, "screenshot.png", reportResp.UploadURLs[1].FileName)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "logs.txt")
	require.NoError(t, err)
	_, err = part.Write([]byte("stack trace"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	uploadReq := httptest.NewRequest(http.MethodPost, reportResp.UploadURLs[0].URL, body)
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())
	uploadRes := httptest.NewRecorder()
	e.ServeHTTP(uploadRes, uploadReq)
	require.Equal(t, http.StatusOK, uploadRes.Code)

	imageBody := &bytes.Buffer{}
	imageWriter := multipart.NewWriter(imageBody)
	imagePart, err := imageWriter.CreateFormFile("file", "screenshot.png")
	require.NoError(t, err)
	_, err = imagePart.Write([]byte("fake png bytes"))
	require.NoError(t, err)
	require.NoError(t, imageWriter.Close())

	imageUploadReq := httptest.NewRequest(http.MethodPost, reportResp.UploadURLs[1].URL, imageBody)
	imageUploadReq.Header.Set("Content-Type", imageWriter.FormDataContentType())
	imageUploadRes := httptest.NewRecorder()
	e.ServeHTTP(imageUploadRes, imageUploadReq)
	require.Equal(t, http.StatusOK, imageUploadRes.Code)

	s := db.NewSession()
	defer s.Close()

	reports := []*models.UserReport{}
	require.NoError(t, s.Find(&reports))
	require.Len(t, reports, 1)
	assert.Equal(t, "high", reports[0].Severity)
	assert.Equal(t, "player@example.com", reports[0].ReporterEmail)
	assert.Equal(t, "Godot Build", reports[0].ApplicationName)
	assert.Equal(t, "v1.2.3", reports[0].ReportTokenLabel)

	task, err := models.GetTaskByIDSimple(s, reports[0].TaskID)
	require.NoError(t, err)
	assert.Equal(t, "Crash on launch", task.Title)
	assert.Contains(t, task.Description, "Build: v1.2.3")
	assert.Equal(t, int64(10), task.ProjectID)
	assert.Equal(t, int64(4), task.Priority)
	db.AssertExists(t, "task_buckets", map[string]interface{}{
		"task_id":         reports[0].TaskID,
		"project_view_id": int64(40),
		"bucket_id":       int64(10),
	}, false)

	attachments, _, _, err := (&models.TaskAttachment{TaskID: reports[0].TaskID}).ReadAll(s, &testuser1, "", 0, 0)
	require.NoError(t, err)
	taskAttachments := attachments.([]*models.TaskAttachment)
	require.Len(t, taskAttachments, 2)
	var latestAttachmentID int64
	for _, attachment := range taskAttachments {
		if attachment.ID > latestAttachmentID {
			latestAttachmentID = attachment.ID
		}
	}
	assert.NotZero(t, task.CoverImageAttachmentID)
	assert.Equal(t, latestAttachmentID, task.CoverImageAttachmentID)
}

func jsonNumber(id int64) string {
	return strconv.FormatInt(id, 10)
}
