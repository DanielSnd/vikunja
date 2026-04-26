package v1

import (
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"github.com/labstack/echo/v5"
)

type taskRouteParam struct {
	TaskID int64 `param:"task"`
}

type adjustTaskTimeTrackingTimerRequest struct {
	DeltaSeconds int64 `json:"delta_seconds"`
}

func GetCurrentTaskTimeTrackingTimer(c *echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	a, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	timer, err := models.GetCurrentTaskTimeTrackingTimer(s, a)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, timer)
}

func StartTaskTimeTrackingTimer(c *echo.Context) error {
	var task taskRouteParam
	if err := c.Bind(&task); err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	a, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	timer, err := models.StartTaskTimeTrackingTimer(s, a, task.TaskID)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	if err := s.Commit(); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, timer)
}

func StopTaskTimeTrackingTimer(c *echo.Context) error {
	var task taskRouteParam
	if err := c.Bind(&task); err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	a, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	entry, err := models.StopTaskTimeTrackingTimer(s, a, task.TaskID)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	if err := s.Commit(); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, entry)
}

func AdjustCurrentTaskTimeTrackingTimer(c *echo.Context) error {
	req := &adjustTaskTimeTrackingTimerRequest{}
	if err := c.Bind(req); err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	a, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	timer, err := models.AdjustCurrentTaskTimeTrackingTimer(s, a, req.DeltaSeconds)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	if err := s.Commit(); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, timer)
}
