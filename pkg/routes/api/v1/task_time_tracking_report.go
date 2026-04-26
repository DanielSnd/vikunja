package v1

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"github.com/jszwedko/go-datemath"
	"github.com/labstack/echo/v5"
)

// TaskTimeTrackingReport returns a time tracking report grouped by project or label.
// @Summary Get a time tracking report
// @Description Returns a time tracking report for the current user across all accessible tasks in the selected time period.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param date_from query string true "Start date, RFC3339 or YYYY-MM-DD"
// @Param date_to query string true "End date, RFC3339 or YYYY-MM-DD"
// @Param group_by query string false "How to group the report items. One of: project, label"
// @Param project_ids query string false "Optional comma-separated project IDs to filter by"
// @Param label_ids query string false "Optional comma-separated label IDs to filter by"
// @Success 200 {object} models.TaskTimeTrackingReport "The report"
// @Failure 400 {object} web.HTTPError "Invalid report query provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /reports/time-tracking [get]
func TaskTimeTrackingReport(c *echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	a, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	dateFrom, err := parseTimeTrackingReportDate(c.QueryParam("date_from"))
	if err != nil {
		return models.InvalidFieldError([]string{"date_from"})
	}
	dateTo, err := parseTimeTrackingReportDate(c.QueryParam("date_to"))
	if err != nil {
		return models.InvalidFieldError([]string{"date_to"})
	}

	report, err := models.GetTaskTimeTrackingReport(s, a, &models.TaskTimeTrackingReportOptions{
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		GroupBy:    models.TaskTimeTrackingReportGroupBy(c.QueryParam("group_by")),
		ProjectIDs: parseTimeTrackingReportIDs(c.QueryParams()["project_ids"]),
		LabelIDs:   parseTimeTrackingReportIDs(c.QueryParams()["label_ids"]),
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, report)
}

func parseTimeTrackingReportDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, strconv.ErrSyntax
	}

	if parsed, err := safeTimeTrackingReportDatemathParse(value); err == nil {
		return parsed.Time(datemath.WithLocation(config.GetTimeZone())).In(config.GetTimeZone()), nil
	}

	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return parsed, nil
	}

	if parsed, err := time.ParseInLocation("2006-01-02 15:04", value, config.GetTimeZone()); err == nil {
		return parsed, nil
	}

	return time.Parse("2006-01-02", value)
}

func safeTimeTrackingReportDatemathParse(value string) (expr datemath.Expression, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = strconv.ErrSyntax
		}
	}()

	return datemath.Parse(value)
}

func parseTimeTrackingReportIDs(values []string) []int64 {
	if len(values) == 0 {
		return nil
	}

	ids := make([]int64, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				continue
			}
			ids = append(ids, id)
		}
	}

	return ids
}
