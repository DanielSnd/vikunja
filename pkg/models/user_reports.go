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
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"

	"xorm.io/xorm"
)

const (
	UserReportAccessKeyPrefix  = "ura_"
	UserReportTokenPrefix      = "urt_"
	defaultUserReportLabel     = "Default"
	userReportPriorityLow      = 1
	userReportPriorityHigh     = 4
	userReportPriorityCritical = 5
)

var (
	ErrUserReportAccessKeyInvalid = errors.New("invalid user report access key")
	ErrUserReportTokenInvalid     = errors.New("invalid user report token")
	ErrUserReportUploadInvalid    = errors.New("invalid user report upload token")
)

type UserReportApplication struct {
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id"`

	Name      string `xorm:"not null" json:"name"`
	ProjectID int64  `xorm:"bigint not null index" json:"project_id"`
	OwnerID   int64  `xorm:"bigint not null index" json:"-"`

	MaxUploadSize int64 `xorm:"bigint not null default 0" json:"max_upload_size"`

	BucketID int64 `xorm:"bigint not null default 0" json:"bucket_id"`

	CriticalProjectID int64 `xorm:"bigint not null default 0" json:"critical_project_id"`
	HighProjectID     int64 `xorm:"bigint not null default 0" json:"high_project_id"`
	LowProjectID      int64 `xorm:"bigint not null default 0" json:"low_project_id"`

	CriticalBucketID int64 `xorm:"bigint not null default 0" json:"critical_bucket_id"`
	HighBucketID     int64 `xorm:"bigint not null default 0" json:"high_bucket_id"`
	LowBucketID      int64 `xorm:"bigint not null default 0" json:"low_bucket_id"`

	CriticalPriority int64 `xorm:"bigint not null default 100" json:"critical_priority"`
	HighPriority     int64 `xorm:"bigint not null default 75" json:"high_priority"`
	LowPriority      int64 `xorm:"bigint not null default 25" json:"low_priority"`

	AccessKey          string `xorm:"-" json:"access_key,omitempty"`
	AccessKeySalt      string `xorm:"not null" json:"-"`
	AccessKeyHash      string `xorm:"not null unique" json:"-"`
	AccessKeyLastEight string `xorm:"not null index varchar(8)" json:"-"`

	Created time.Time `xorm:"created not null" json:"created"`
	Updated time.Time `xorm:"updated not null" json:"updated"`
}

func (*UserReportApplication) TableName() string {
	return "user_report_applications"
}

type UserReportToken struct {
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id"`

	ApplicationID int64  `xorm:"bigint not null index" json:"application_id"`
	Label         string `xorm:"not null" json:"label"`

	Token          string `xorm:"-" json:"token,omitempty"`
	TokenSalt      string `xorm:"not null" json:"-"`
	TokenHash      string `xorm:"not null unique" json:"-"`
	TokenLastEight string `xorm:"not null index varchar(8)" json:"-"`

	IsEnabled   bool      `xorm:"not null default true" json:"is_enabled"`
	CreatedByID int64     `xorm:"bigint not null" json:"-"`
	Created     time.Time `xorm:"created not null" json:"created"`
}

func (*UserReportToken) TableName() string {
	return "user_report_tokens"
}

type UserReport struct {
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id"`

	TaskID           int64  `xorm:"bigint not null unique index" json:"task_id"`
	ApplicationID    int64  `xorm:"bigint not null index" json:"application_id"`
	ApplicationName  string `xorm:"not null" json:"application_name"`
	ReportTokenID    int64  `xorm:"bigint not null index" json:"report_token_id"`
	ReportTokenLabel string `xorm:"not null" json:"report_token_label"`
	Severity         string `xorm:"varchar(16) not null" json:"severity"`
	ReporterEmail    string `xorm:"varchar(191)" json:"reporter_email"`

	NotifiedDoneAt time.Time `xorm:"DATETIME null" json:"notified_done_at"`
	Created        time.Time `xorm:"created not null" json:"created"`
}

func (*UserReport) TableName() string {
	return "user_reports"
}

type UserReportApplicationCreateRequest struct {
	Name      string `json:"name"`
	ProjectID int64  `json:"project_id"`
}

type UserReportApplicationUpdateRequest struct {
	Name              string `json:"name"`
	ProjectID         int64  `json:"project_id"`
	MaxUploadSize     int64  `json:"max_upload_size"`
	BucketID          int64  `json:"bucket_id"`
	CriticalProjectID int64  `json:"critical_project_id"`
	HighProjectID     int64  `json:"high_project_id"`
	LowProjectID      int64  `json:"low_project_id"`
	CriticalBucketID  int64  `json:"critical_bucket_id"`
	HighBucketID      int64  `json:"high_bucket_id"`
	LowBucketID       int64  `json:"low_bucket_id"`
	CriticalPriority  int64  `json:"critical_priority"`
	HighPriority      int64  `json:"high_priority"`
	LowPriority       int64  `json:"low_priority"`
}

type UserReportApplicationCreateResponse struct {
	Application        *UserReportApplication `json:"application"`
	AccessKey          string                 `json:"accessKey"`
	DefaultReportToken *UserReportToken       `json:"defaultReportToken"`
}

type UserReportCreateTokenRequest struct {
	Label string `json:"label"`
}

type UserReportTokenUpdateRequest struct {
	IsEnabled bool `json:"is_enabled"`
}

type UserReportCreateTokenResponse struct {
	OK    bool   `json:"ok"`
	Token string `json:"token"`
}

type UserReportCreateRequest struct {
	Content   string   `json:"content"`
	Severity  string   `json:"severity"`
	FileNames []string `json:"fileNames"`
	UserEmail string   `json:"userEmail"`
}

type UserReportUploadURL struct {
	FileName string            `json:"fileName"`
	URL      string            `json:"url"`
	Fields   map[string]string `json:"fields"`
}

type UserReportCreateResponse struct {
	OK         bool                  `json:"ok"`
	CardID     string                `json:"cardId"`
	UploadURLs []UserReportUploadURL `json:"uploadUrls"`
}

type userReportUploadTokenPayload struct {
	ReportID int64  `json:"reportId"`
	FileName string `json:"fileName"`
	Expires  int64  `json:"expires"`
}

type userReportSignedUploadToken struct {
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
}

type userReportDoneMailEntry struct {
	TaskID           int64
	TaskTitle        string
	ApplicationName  string
	ReportTokenLabel string
}

func CreateUserReportApplication(s *xorm.Session, auth *user.User, req *UserReportApplicationCreateRequest) (*UserReportApplicationCreateResponse, error) {
	if req.ProjectID == 0 {
		return nil, fmt.Errorf("project id is required")
	}

	app := &UserReportApplication{
		Name:             strings.TrimSpace(req.Name),
		ProjectID:        req.ProjectID,
		OwnerID:          auth.ID,
		CriticalPriority: userReportPriorityCritical,
		HighPriority:     userReportPriorityHigh,
		LowPriority:      userReportPriorityLow,
	}
	if err := validateUserReportApplicationSettings(s, auth, app); err != nil {
		return nil, err
	}

	accessKey, salt, hash, lastEight, err := generateUserReportSecret(UserReportAccessKeyPrefix)
	if err != nil {
		return nil, err
	}

	app.AccessKey = accessKey
	app.AccessKeySalt = salt
	app.AccessKeyHash = hash
	app.AccessKeyLastEight = lastEight

	if _, err = s.Insert(app); err != nil {
		return nil, err
	}

	token, err := createUserReportToken(s, app, defaultUserReportLabel, auth.ID)
	if err != nil {
		return nil, err
	}

	return &UserReportApplicationCreateResponse{
		Application:        app,
		AccessKey:          accessKey,
		DefaultReportToken: token,
	}, nil
}

func ListUserReportApplicationsForUser(s *xorm.Session, auth *user.User) ([]*UserReportApplication, error) {
	apps := []*UserReportApplication{}
	return apps, s.Where("owner_id = ?", auth.ID).Asc("id").Find(&apps)
}

func GetUserReportApplicationForOwner(s *xorm.Session, id, ownerID int64) (*UserReportApplication, error) {
	app := &UserReportApplication{}
	query := s.Where("id = ?", id)
	if ownerID != 0 {
		query = query.And("owner_id = ?", ownerID)
	}

	exists, err := query.Get(app)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrProjectDoesNotExist{ID: id}
	}
	return app, nil
}

func UpdateUserReportApplication(s *xorm.Session, applicationID int64, auth *user.User, req *UserReportApplicationUpdateRequest) (*UserReportApplication, error) {
	app, err := GetUserReportApplicationForOwner(s, applicationID, auth.ID)
	if err != nil {
		return nil, err
	}

	app.Name = strings.TrimSpace(req.Name)
	app.ProjectID = req.ProjectID
	app.MaxUploadSize = req.MaxUploadSize
	app.BucketID = req.BucketID
	app.CriticalProjectID = req.CriticalProjectID
	app.HighProjectID = req.HighProjectID
	app.LowProjectID = req.LowProjectID
	app.CriticalBucketID = req.CriticalBucketID
	app.HighBucketID = req.HighBucketID
	app.LowBucketID = req.LowBucketID
	app.CriticalPriority = req.CriticalPriority
	app.HighPriority = req.HighPriority
	app.LowPriority = req.LowPriority

	if err = validateUserReportApplicationSettings(s, auth, app); err != nil {
		return nil, err
	}

	_, err = s.ID(app.ID).Cols(
		"name",
		"project_id",
		"max_upload_size",
		"bucket_id",
		"critical_project_id",
		"high_project_id",
		"low_project_id",
		"critical_bucket_id",
		"high_bucket_id",
		"low_bucket_id",
		"critical_priority",
		"high_priority",
		"low_priority",
		"updated",
	).Update(app)
	if err != nil {
		return nil, err
	}

	return app, nil
}

func DeleteUserReportApplication(s *xorm.Session, applicationID int64, auth *user.User) error {
	app, err := GetUserReportApplicationForOwner(s, applicationID, auth.ID)
	if err != nil {
		return err
	}

	if _, err = s.Where("application_id = ?", app.ID).Delete(&UserReportToken{}); err != nil {
		return err
	}
	_, err = s.ID(app.ID).Delete(&UserReportApplication{})
	return err
}

func RegenerateUserReportAccessKey(s *xorm.Session, applicationID int64, auth *user.User) (string, error) {
	app, err := GetUserReportApplicationForOwner(s, applicationID, auth.ID)
	if err != nil {
		return "", err
	}

	accessKey, salt, hash, lastEight, err := generateUserReportSecret(UserReportAccessKeyPrefix)
	if err != nil {
		return "", err
	}

	app.AccessKeySalt = salt
	app.AccessKeyHash = hash
	app.AccessKeyLastEight = lastEight

	_, err = s.ID(app.ID).Cols("access_key_salt", "access_key_hash", "access_key_last_eight", "updated").Update(app)
	return accessKey, err
}

func ListUserReportTokensForApplication(s *xorm.Session, applicationID int64, auth *user.User) ([]*UserReportToken, error) {
	if _, err := GetUserReportApplicationForOwner(s, applicationID, auth.ID); err != nil {
		return nil, err
	}

	tokens := []*UserReportToken{}
	return tokens, s.Where("application_id = ?", applicationID).Asc("id").Find(&tokens)
}

func CreateUserReportTokenForApplication(s *xorm.Session, applicationID int64, auth *user.User, req *UserReportCreateTokenRequest) (*UserReportToken, error) {
	app, err := GetUserReportApplicationForOwner(s, applicationID, auth.ID)
	if err != nil {
		return nil, err
	}

	req.Label = strings.TrimSpace(req.Label)
	if req.Label == "" {
		return nil, fmt.Errorf("token label is required")
	}

	return createUserReportToken(s, app, req.Label, auth.ID)
}

func UpdateUserReportToken(s *xorm.Session, applicationID, tokenID int64, auth *user.User, req *UserReportTokenUpdateRequest) (*UserReportToken, error) {
	if _, err := GetUserReportApplicationForOwner(s, applicationID, auth.ID); err != nil {
		return nil, err
	}

	token := &UserReportToken{}
	exists, err := s.Where("id = ? AND application_id = ?", tokenID, applicationID).Get(token)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrTaskDoesNotExist{ID: tokenID}
	}

	token.IsEnabled = req.IsEnabled
	_, err = s.ID(token.ID).Cols("is_enabled").Update(token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func CreateUserReportTokenFromAccessKey(s *xorm.Session, accessKey string, req *UserReportCreateTokenRequest) (*UserReportToken, error) {
	req.Label = strings.TrimSpace(req.Label)
	if req.Label == "" {
		return nil, fmt.Errorf("token label is required")
	}

	app, owner, err := ValidateUserReportAccessKey(s, accessKey)
	if err != nil {
		return nil, err
	}

	return createUserReportToken(s, app, req.Label, owner.ID)
}

func CreateUserReport(s *xorm.Session, rawToken string, req *UserReportCreateRequest) (*UserReportCreateResponse, error) {
	token, app, owner, err := ValidateUserReportToken(s, rawToken)
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, fmt.Errorf("report content is required")
	}

	fileNames := make([]string, 0, len(req.FileNames))
	for _, fileName := range req.FileNames {
		fileName = strings.TrimSpace(fileName)
		if fileName == "" {
			continue
		}
		fileNames = append(fileNames, fileName)
	}
	if len(fileNames) > 0 && !config.ServiceEnableTaskAttachments.GetBool() {
		return nil, fmt.Errorf("task attachments are disabled")
	}

	severity := normalizeUserReportSeverity(req.Severity)
	projectID, bucketID, priority := resolveUserReportTarget(app, severity)
	task := &Task{
		ProjectID:   projectID,
		BucketID:    bucketID,
		Title:       userReportTitleFromContent(content),
		Description: buildUserReportDescription(content, app.Name, token.Label, severity, req.UserEmail),
		Priority:    priority,
	}
	if task.Title == "" {
		return nil, fmt.Errorf("report title is required")
	}

	if err = createTask(s, task, owner, false, true); err != nil {
		return nil, err
	}

	report := &UserReport{
		TaskID:           task.ID,
		ApplicationID:    app.ID,
		ApplicationName:  app.Name,
		ReportTokenID:    token.ID,
		ReportTokenLabel: token.Label,
		Severity:         severity,
		ReporterEmail:    normalizeUserReportEmail(req.UserEmail),
	}
	if _, err = s.Insert(report); err != nil {
		return nil, err
	}

	response := &UserReportCreateResponse{
		OK:         true,
		CardID:     fmt.Sprintf("%d", task.ID),
		UploadURLs: make([]UserReportUploadURL, 0, len(fileNames)),
	}

	for _, fileName := range fileNames {
		uploadToken, err := CreateUserReportUploadToken(report.ID, fileName, time.Now().Add(15*time.Minute))
		if err != nil {
			return nil, err
		}

		response.UploadURLs = append(response.UploadURLs, UserReportUploadURL{
			FileName: fileName,
			URL:      buildUserReportUploadURL(uploadToken),
			Fields:   map[string]string{},
		})
	}

	return response, nil
}

func CreateUserReportUploadToken(reportID int64, fileName string, expiresAt time.Time) (string, error) {
	payloadBytes, err := json.Marshal(userReportUploadTokenPayload{
		ReportID: reportID,
		FileName: fileName,
		Expires:  expiresAt.Unix(),
	})
	if err != nil {
		return "", err
	}

	signature := signUserReportUploadPayload(payloadBytes)
	wrapperBytes, err := json.Marshal(userReportSignedUploadToken{
		Payload:   base64.RawURLEncoding.EncodeToString(payloadBytes),
		Signature: hex.EncodeToString(signature),
	})
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(wrapperBytes), nil
}

func ValidateUserReportUploadToken(uploadToken string) (*UserReport, string, error) {
	wrapperBytes, err := base64.RawURLEncoding.DecodeString(uploadToken)
	if err != nil {
		return nil, "", ErrUserReportUploadInvalid
	}

	wrapper := &userReportSignedUploadToken{}
	if err = json.Unmarshal(wrapperBytes, wrapper); err != nil {
		return nil, "", ErrUserReportUploadInvalid
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(wrapper.Payload)
	if err != nil {
		return nil, "", ErrUserReportUploadInvalid
	}

	expectedSignature := signUserReportUploadPayload(payloadBytes)
	gotSignature, err := hex.DecodeString(wrapper.Signature)
	if err != nil {
		return nil, "", ErrUserReportUploadInvalid
	}
	if subtle.ConstantTimeCompare(expectedSignature, gotSignature) != 1 {
		return nil, "", ErrUserReportUploadInvalid
	}

	payload := &userReportUploadTokenPayload{}
	if err = json.Unmarshal(payloadBytes, payload); err != nil {
		return nil, "", ErrUserReportUploadInvalid
	}
	if payload.ReportID == 0 || payload.FileName == "" || time.Now().Unix() > payload.Expires {
		return nil, "", ErrUserReportUploadInvalid
	}

	return &UserReport{ID: payload.ReportID}, payload.FileName, nil
}

func ValidateUserReportAccessKey(s *xorm.Session, rawAccessKey string) (*UserReportApplication, *user.User, error) {
	app, err := getUserReportApplicationByAccessKey(s, rawAccessKey)
	if err != nil {
		return nil, nil, err
	}

	owner, err := user.GetUserByID(s, app.OwnerID)
	if err != nil {
		if user.IsErrUserStatusError(err) {
			return nil, nil, ErrUserReportAccessKeyInvalid
		}
		return nil, nil, err
	}

	return app, owner, nil
}

func ValidateUserReportToken(s *xorm.Session, rawToken string) (*UserReportToken, *UserReportApplication, *user.User, error) {
	token, err := getUserReportTokenByRawToken(s, rawToken)
	if err != nil {
		return nil, nil, nil, err
	}
	if !token.IsEnabled {
		return nil, nil, nil, ErrUserReportTokenInvalid
	}

	app, err := GetUserReportApplicationForOwner(s, token.ApplicationID, 0)
	if err != nil {
		return nil, nil, nil, err
	}

	owner, err := user.GetUserByID(s, app.OwnerID)
	if err != nil {
		if user.IsErrUserStatusError(err) {
			return nil, nil, nil, ErrUserReportTokenInvalid
		}
		return nil, nil, nil, err
	}

	return token, app, owner, nil
}

func GetUserReportByID(s *xorm.Session, id int64) (*UserReport, error) {
	report := &UserReport{}
	exists, err := s.Where("id = ?", id).Get(report)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrTaskDoesNotExist{ID: id}
	}
	return report, nil
}

func createUserReportToken(s *xorm.Session, app *UserReportApplication, label string, createdByID int64) (*UserReportToken, error) {
	rawToken, salt, hash, lastEight, err := generateUserReportSecret(UserReportTokenPrefix)
	if err != nil {
		return nil, err
	}

	token := &UserReportToken{
		ApplicationID:  app.ID,
		Label:          label,
		Token:          rawToken,
		TokenSalt:      salt,
		TokenHash:      hash,
		TokenLastEight: lastEight,
		IsEnabled:      true,
		CreatedByID:    createdByID,
	}
	if _, err = s.Insert(token); err != nil {
		return nil, err
	}

	return token, nil
}

func generateUserReportSecret(prefix string) (raw, salt, hash, lastEight string, err error) {
	salt, err = utils.CryptoRandomString(10)
	if err != nil {
		return
	}

	tokenBytes, err := utils.CryptoRandomBytes(20)
	if err != nil {
		return
	}

	raw = prefix + hex.EncodeToString(tokenBytes)
	hash = HashToken(raw, salt)
	lastEight = raw[len(raw)-8:]
	return
}

func getUserReportApplicationByAccessKey(s *xorm.Session, rawAccessKey string) (*UserReportApplication, error) {
	if !strings.HasPrefix(rawAccessKey, UserReportAccessKeyPrefix) || len(rawAccessKey) < 8 {
		return nil, ErrUserReportAccessKeyInvalid
	}

	apps := []*UserReportApplication{}
	lastEight := rawAccessKey[len(rawAccessKey)-8:]
	if err := s.Where("access_key_last_eight = ?", lastEight).Find(&apps); err != nil {
		return nil, err
	}

	for _, candidate := range apps {
		expectedHash := HashToken(rawAccessKey, candidate.AccessKeySalt)
		if subtle.ConstantTimeCompare([]byte(candidate.AccessKeyHash), []byte(expectedHash)) == 1 {
			return candidate, nil
		}
	}

	return nil, ErrUserReportAccessKeyInvalid
}

func getUserReportTokenByRawToken(s *xorm.Session, rawToken string) (*UserReportToken, error) {
	if !strings.HasPrefix(rawToken, UserReportTokenPrefix) || len(rawToken) < 8 {
		return nil, ErrUserReportTokenInvalid
	}

	tokens := []*UserReportToken{}
	lastEight := rawToken[len(rawToken)-8:]
	if err := s.Where("token_last_eight = ?", lastEight).Find(&tokens); err != nil {
		return nil, err
	}

	for _, candidate := range tokens {
		expectedHash := HashToken(rawToken, candidate.TokenSalt)
		if subtle.ConstantTimeCompare([]byte(candidate.TokenHash), []byte(expectedHash)) == 1 {
			return candidate, nil
		}
	}

	return nil, ErrUserReportTokenInvalid
}

func normalizeUserReportSeverity(severity string) string {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "critical":
		return "critical"
	case "high":
		return "high"
	case "low":
		return "low"
	default:
		return ""
	}
}

func resolveUserReportTarget(app *UserReportApplication, severity string) (projectID int64, bucketID int64, priority int64) {
	projectID = app.ProjectID
	bucketID = app.BucketID
	priority = 0

	switch severity {
	case "critical":
		if app.CriticalProjectID != 0 {
			projectID = app.CriticalProjectID
		}
		if app.CriticalBucketID != 0 {
			bucketID = app.CriticalBucketID
		}
		priority = app.CriticalPriority
	case "high":
		if app.HighProjectID != 0 {
			projectID = app.HighProjectID
		}
		if app.HighBucketID != 0 {
			bucketID = app.HighBucketID
		}
		priority = app.HighPriority
	case "low":
		if app.LowProjectID != 0 {
			projectID = app.LowProjectID
		}
		if app.LowBucketID != 0 {
			bucketID = app.LowBucketID
		}
		priority = app.LowPriority
	}

	return
}

func userReportTitleFromContent(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

func buildUserReportDescription(content, applicationName, label, severity, userEmail string) string {
	metadata := []string{
		"---",
		"User report metadata",
		"Application: " + applicationName,
		"Build: " + label,
	}
	if severity != "" {
		metadata = append(metadata, "Severity: "+severity)
	}
	if email := normalizeUserReportEmail(userEmail); email != "" {
		metadata = append(metadata, "Reporter Email: "+email)
	}

	return strings.TrimSpace(content) + "\n\n" + strings.Join(metadata, "\n")
}

func normalizeUserReportEmail(email string) string {
	email = strings.TrimSpace(email)
	if email == "" || !strings.Contains(email, "@") {
		return ""
	}
	return email
}

func buildUserReportUploadURL(uploadToken string) string {
	base := strings.TrimRight(config.ServicePublicURL.GetString(), "/")
	return base + "/api/v1/user-reports/v1/upload?uploadToken=" + uploadToken
}

func signUserReportUploadPayload(payload []byte) []byte {
	mac := hmac.New(sha256.New, []byte(config.ServiceSecret.GetString()))
	_, _ = mac.Write(payload)
	return mac.Sum(nil)
}

func validateUserReportApplicationSettings(s *xorm.Session, auth *user.User, app *UserReportApplication) error {
	app.Name = strings.TrimSpace(app.Name)
	if app.Name == "" {
		return fmt.Errorf("application name is required")
	}
	if app.ProjectID == 0 {
		return fmt.Errorf("project id is required")
	}
	if app.MaxUploadSize < 0 {
		return fmt.Errorf("maximum upload size must not be negative")
	}

	for _, priority := range []int64{app.CriticalPriority, app.HighPriority, app.LowPriority} {
		if priority < 0 || priority > userReportPriorityCritical {
			return fmt.Errorf("priorities must be between 0 and 5")
		}
	}

	projectIDs := []int64{
		app.ProjectID,
		app.CriticalProjectID,
		app.HighProjectID,
		app.LowProjectID,
	}
	for _, projectID := range projectIDs {
		if projectID == 0 {
			continue
		}
		project := &Project{ID: projectID}
		canWrite, err := project.CanWrite(s, auth)
		if err != nil {
			return err
		}
		if !canWrite {
			return ErrGenericForbidden{}
		}
	}

	bucketTargets := []struct {
		bucketID  int64
		projectID int64
	}{
		{bucketID: app.BucketID, projectID: app.ProjectID},
		{bucketID: app.CriticalBucketID, projectID: firstNonZeroInt64(app.CriticalProjectID, app.ProjectID)},
		{bucketID: app.HighBucketID, projectID: firstNonZeroInt64(app.HighProjectID, app.ProjectID)},
		{bucketID: app.LowBucketID, projectID: firstNonZeroInt64(app.LowProjectID, app.ProjectID)},
	}
	for _, target := range bucketTargets {
		if target.bucketID == 0 {
			continue
		}

		bucket, err := getBucketByID(s, target.bucketID)
		if err != nil {
			return err
		}

		view, err := GetProjectViewByID(s, bucket.ProjectViewID)
		if err != nil {
			return err
		}

		if view.ProjectID != target.projectID {
			return ErrBucketDoesNotBelongToProjectView{
				BucketID:      bucket.ID,
				ProjectViewID: view.ID,
			}
		}
		if view.ViewKind != ProjectViewKindKanban || view.BucketConfigurationMode != BucketConfigurationModeManual {
			return ErrBucketDoesNotBelongToProjectView{
				BucketID:      bucket.ID,
				ProjectViewID: view.ID,
			}
		}
	}

	return nil
}

func firstNonZeroInt64(values ...int64) int64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}

	return 0
}
