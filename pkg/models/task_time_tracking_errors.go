package models

import (
	"fmt"

	"code.vikunja.io/api/pkg/web"
)

type ErrTaskTimeTrackingDoesNotExist struct {
	ID     int64
	TaskID int64
}

func (e ErrTaskTimeTrackingDoesNotExist) Error() string {
	if e.TaskID != 0 {
		return fmt.Sprintf("time tracking entry %d for task %d does not exist", e.ID, e.TaskID)
	}
	return fmt.Sprintf("time tracking entry %d does not exist", e.ID)
}

func (e ErrTaskTimeTrackingDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: 404,
		Message:  e.Error(),
	}
}

type ErrTaskTimeTrackingTimerAlreadyRunning struct {
	TaskID int64
}

func (e ErrTaskTimeTrackingTimerAlreadyRunning) Error() string {
	return fmt.Sprintf("a time tracking timer is already running on task %d", e.TaskID)
}

func (e ErrTaskTimeTrackingTimerAlreadyRunning) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: 409,
		Message:  e.Error(),
	}
}

type ErrTaskTimeTrackingTimerNotRunning struct{}

func (e ErrTaskTimeTrackingTimerNotRunning) Error() string {
	return "no active time tracking timer is running"
}

func (e ErrTaskTimeTrackingTimerNotRunning) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: 404,
		Message:  e.Error(),
	}
}
