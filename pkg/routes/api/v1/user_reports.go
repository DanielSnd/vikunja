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

package v1

import (
	"errors"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	auth2 "code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
)

func CreateUserReportApplication(c *echo.Context) error {
	currentUser, err := getUserReportCurrentUser(c)
	if err != nil {
		return err
	}

	req := &models.UserReportApplicationCreateRequest{}
	if err = c.Bind(req); err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	response, err := models.CreateUserReportApplication(s, currentUser, req)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	if err = s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusCreated, response)
}

func ListUserReportApplications(c *echo.Context) error {
	currentUser, err := getUserReportCurrentUser(c)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	apps, err := models.ListUserReportApplicationsForUser(s, currentUser)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, apps)
}

func UpdateUserReportApplication(c *echo.Context) error {
	currentUser, err := getUserReportCurrentUser(c)
	if err != nil {
		return err
	}

	applicationID, err := strconv.ParseInt(c.Param("application"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid application id").Wrap(err)
	}

	req := &models.UserReportApplicationUpdateRequest{}
	if err = c.Bind(req); err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	app, err := models.UpdateUserReportApplication(s, applicationID, currentUser, req)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	if err = s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, app)
}

func DeleteUserReportApplication(c *echo.Context) error {
	currentUser, err := getUserReportCurrentUser(c)
	if err != nil {
		return err
	}

	applicationID, err := strconv.ParseInt(c.Param("application"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid application id").Wrap(err)
	}

	s := db.NewSession()
	defer s.Close()

	if err = models.DeleteUserReportApplication(s, applicationID, currentUser); err != nil {
		_ = s.Rollback()
		return err
	}
	if err = s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, map[string]any{"message": "Successfully deleted."})
}

func RegenerateUserReportAccessKey(c *echo.Context) error {
	currentUser, err := getUserReportCurrentUser(c)
	if err != nil {
		return err
	}

	applicationID, err := strconv.ParseInt(c.Param("application"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid application id").Wrap(err)
	}

	s := db.NewSession()
	defer s.Close()

	accessKey, err := models.RegenerateUserReportAccessKey(s, applicationID, currentUser)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	if err = s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, map[string]any{
		"ok":        true,
		"accessKey": accessKey,
	})
}

func ListUserReportTokens(c *echo.Context) error {
	currentUser, err := getUserReportCurrentUser(c)
	if err != nil {
		return err
	}

	applicationID, err := strconv.ParseInt(c.Param("application"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid application id").Wrap(err)
	}

	s := db.NewSession()
	defer s.Close()

	tokens, err := models.ListUserReportTokensForApplication(s, applicationID, currentUser)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, tokens)
}

func CreateUserReportApplicationToken(c *echo.Context) error {
	currentUser, err := getUserReportCurrentUser(c)
	if err != nil {
		return err
	}

	applicationID, err := strconv.ParseInt(c.Param("application"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid application id").Wrap(err)
	}

	req := &models.UserReportCreateTokenRequest{}
	if err = c.Bind(req); err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	token, err := models.CreateUserReportTokenForApplication(s, applicationID, currentUser, req)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	if err = s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusCreated, token)
}

func UpdateUserReportToken(c *echo.Context) error {
	currentUser, err := getUserReportCurrentUser(c)
	if err != nil {
		return err
	}

	applicationID, err := strconv.ParseInt(c.Param("application"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid application id").Wrap(err)
	}
	tokenID, err := strconv.ParseInt(c.Param("token"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid token id").Wrap(err)
	}

	req := &models.UserReportTokenUpdateRequest{}
	if err = c.Bind(req); err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	token, err := models.UpdateUserReportToken(s, applicationID, tokenID, currentUser, req)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	if err = s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, token)
}

func CreateUserReportToken(c *echo.Context) error {
	req := &models.UserReportCreateTokenRequest{}
	if err := c.Bind(req); err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	token, err := models.CreateUserReportTokenFromAccessKey(s, c.QueryParam("accessKey"), req)
	if err != nil {
		if errors.Is(err, models.ErrUserReportAccessKeyInvalid) {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}
		_ = s.Rollback()
		return err
	}
	if err = s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, &models.UserReportCreateTokenResponse{
		OK:    true,
		Token: token.Token,
	})
}

func CreateUserReport(c *echo.Context) error {
	req := &models.UserReportCreateRequest{}
	if err := c.Bind(req); err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	response, err := models.CreateUserReport(s, c.QueryParam("token"), req)
	if err != nil {
		if errors.Is(err, models.ErrUserReportTokenInvalid) {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}
		_ = s.Rollback()
		return err
	}
	if err = s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, response)
}

func UploadUserReportAttachment(c *echo.Context) error {
	if !config.ServiceEnableTaskAttachments.GetBool() {
		return echo.NewHTTPError(http.StatusBadRequest, "Task attachments are disabled")
	}

	reportRef, expectedFileName, err := models.ValidateUserReportUploadToken(c.QueryParam("uploadToken"))
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	files := form.File["file"]
	if len(files) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "No file uploaded")
	}

	fileHeader := files[0]
	if fileHeader.Filename != expectedFileName {
		return echo.NewHTTPError(http.StatusBadRequest, "Uploaded file does not match the expected file name")
	}

	s := db.NewSession()
	defer s.Close()

	report, err := models.GetUserReportByID(s, reportRef.ID)
	if err != nil {
		return err
	}

	application, err := models.GetUserReportApplicationForOwner(s, report.ApplicationID, 0)
	if err != nil {
		return err
	}

	if application.MaxUploadSize > 0 && fileHeader.Size > application.MaxUploadSize {
		return echo.NewHTTPError(http.StatusRequestEntityTooLarge, "Uploaded file exceeds the configured maximum upload size for this application")
	}

	owner, err := user.GetUserByID(s, application.OwnerID)
	if err != nil {
		return err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	attachment := &models.TaskAttachment{TaskID: report.TaskID}
	if err = attachment.NewAttachment(s, file, fileHeader.Filename, uint64(fileHeader.Size), owner); err != nil {
		_ = s.Rollback()
		return err
	}

	if isUserReportImageUpload(fileHeader) {
		if _, err = s.ID(report.TaskID).
			Cols("cover_image_attachment_id").
			Update(&models.Task{
				ID:                     report.TaskID,
				CoverImageAttachmentID: attachment.ID,
			}); err != nil {
			_ = s.Rollback()
			return err
		}
	}
	if err = s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, map[string]any{"ok": true})
}

func getUserReportCurrentUser(c *echo.Context) (*user.User, error) {
	auth, err := auth2.GetAuthFromClaims(c)
	if err != nil {
		return nil, err
	}

	currentUser, ok := auth.(*user.User)
	if !ok {
		return nil, echo.ErrForbidden
	}

	return currentUser, nil
}

func isUserReportImageUpload(fileHeader *multipart.FileHeader) bool {
	contentType := strings.TrimSpace(fileHeader.Header.Get("Content-Type"))
	if strings.HasPrefix(strings.ToLower(contentType), "image/") {
		return true
	}

	return strings.HasPrefix(strings.ToLower(mime.TypeByExtension(filepath.Ext(fileHeader.Filename))), "image/")
}
