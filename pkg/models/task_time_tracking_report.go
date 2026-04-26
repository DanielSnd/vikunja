package models

import (
	"database/sql"
	"slices"
	"sort"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

type TaskTimeTrackingReportGroupBy string

const (
	TaskTimeTrackingReportGroupByProject TaskTimeTrackingReportGroupBy = "project"
	TaskTimeTrackingReportGroupByLabel   TaskTimeTrackingReportGroupBy = "label"
)

type TaskTimeTrackingReportOptions struct {
	DateFrom   time.Time
	DateTo     time.Time
	GroupBy    TaskTimeTrackingReportGroupBy
	ProjectIDs []int64
	LabelIDs   []int64
}

type TaskTimeTrackingReport struct {
	GroupBy      TaskTimeTrackingReportGroupBy `json:"group_by"`
	DateFrom     time.Time                     `json:"date_from"`
	DateTo       time.Time                     `json:"date_to"`
	TotalSeconds int64                         `json:"total_seconds"`
	Days         []*TaskTimeTrackingReportDay  `json:"days"`
	Items        []*TaskTimeTrackingReportItem `json:"items"`
}

type TaskTimeTrackingReportDay struct {
	Date         string `json:"date"`
	TotalSeconds int64  `json:"total_seconds"`
}

type TaskTimeTrackingReportItem struct {
	ID           int64   `json:"id"`
	Title        string  `json:"title"`
	Color        string  `json:"color"`
	TotalSeconds int64   `json:"total_seconds"`
	DailySeconds []int64 `json:"daily_seconds"`
}

type taskTimeTrackingReportRow struct {
	TrackedAt  time.Time      `xorm:"tracked_at"`
	TimeSpent  int64          `xorm:"time_spent"`
	GroupID    sql.NullInt64  `xorm:"group_id"`
	GroupTitle sql.NullString `xorm:"group_title"`
	GroupColor sql.NullString `xorm:"group_color"`
}

func (o *TaskTimeTrackingReportOptions) validate() error {
	if o.DateFrom.IsZero() || o.DateTo.IsZero() {
		return InvalidFieldError([]string{"date_from", "date_to"})
	}
	if o.GroupBy == "" {
		o.GroupBy = TaskTimeTrackingReportGroupByProject
	}
	if o.GroupBy != TaskTimeTrackingReportGroupByProject && o.GroupBy != TaskTimeTrackingReportGroupByLabel {
		return InvalidFieldErrorWithMessage([]string{"group_by"}, "group_by must be either project or label")
	}
	if o.DateTo.Before(o.DateFrom) {
		return InvalidFieldErrorWithMessage([]string{"date_to"}, "date_to must be greater than or equal to date_from")
	}
	return nil
}

func GetTaskTimeTrackingReport(s *xorm.Session, a web.Auth, opts *TaskTimeTrackingReportOptions) (*TaskTimeTrackingReport, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	tz := config.GetTimeZone()
	from := time.Date(opts.DateFrom.In(tz).Year(), opts.DateFrom.In(tz).Month(), opts.DateFrom.In(tz).Day(), 0, 0, 0, 0, tz)
	to := time.Date(opts.DateTo.In(tz).Year(), opts.DateTo.In(tz).Month(), opts.DateTo.In(tz).Day(), 0, 0, 0, 0, tz)
	toExclusive := to.AddDate(0, 0, 1)

	rows, err := getTaskTimeTrackingReportRows(s, a, opts, from, toExclusive)
	if err != nil {
		return nil, err
	}

	report := &TaskTimeTrackingReport{
		GroupBy:  opts.GroupBy,
		DateFrom: from,
		DateTo:   to,
		Days:     []*TaskTimeTrackingReportDay{},
		Items:    []*TaskTimeTrackingReportItem{},
	}

	dayIndexByKey := make(map[string]int)
	for current := from; !current.After(to); current = current.AddDate(0, 0, 1) {
		key := current.Format("2006-01-02")
		dayIndexByKey[key] = len(report.Days)
		report.Days = append(report.Days, &TaskTimeTrackingReportDay{
			Date: key,
		})
	}

	itemIndexByID := make(map[int64]int)
	for _, row := range rows {
		dayKey := row.TrackedAt.In(tz).Format("2006-01-02")
		dayIdx, hasDay := dayIndexByKey[dayKey]
		if !hasDay {
			continue
		}

		groupID := row.GroupID.Int64
		groupTitle := row.GroupTitle.String
		groupColor := row.GroupColor.String
		if opts.GroupBy == TaskTimeTrackingReportGroupByLabel && !row.GroupID.Valid {
			groupID = 0
			groupTitle = "No Label"
			groupColor = ""
		}
		if groupTitle == "" {
			continue
		}

		itemIdx, hasItem := itemIndexByID[groupID]
		if !hasItem {
			itemIdx = len(report.Items)
			itemIndexByID[groupID] = itemIdx
			report.Items = append(report.Items, &TaskTimeTrackingReportItem{
				ID:           groupID,
				Title:        groupTitle,
				Color:        groupColor,
				DailySeconds: make([]int64, len(report.Days)),
			})
		}

		report.TotalSeconds += row.TimeSpent
		report.Days[dayIdx].TotalSeconds += row.TimeSpent
		report.Items[itemIdx].TotalSeconds += row.TimeSpent
		report.Items[itemIdx].DailySeconds[dayIdx] += row.TimeSpent
	}

	sort.Slice(report.Items, func(i, j int) bool {
		if report.Items[i].TotalSeconds == report.Items[j].TotalSeconds {
			return report.Items[i].Title < report.Items[j].Title
		}
		return report.Items[i].TotalSeconds > report.Items[j].TotalSeconds
	})

	return report, nil
}

func getTaskTimeTrackingReportRows(s *xorm.Session, a web.Auth, opts *TaskTimeTrackingReportOptions, from time.Time, toExclusive time.Time) ([]*taskTimeTrackingReportRow, error) {
	rows := []*taskTimeTrackingReportRow{}

	query := s.
		Table("task_time_tracking").
		Join("INNER", "tasks", "tasks.id = task_time_tracking.task_id").
		Where(accessibleProjectIDsSubquery(a, "tasks.project_id")).
		And("task_time_tracking.tracked_at >= ?", from).
		And("task_time_tracking.tracked_at < ?", toExclusive)

	if len(opts.ProjectIDs) > 0 {
		projectIDs := slices.Clone(opts.ProjectIDs)
		query = query.And(builder.In("tasks.project_id", projectIDs))
	}

	if opts.GroupBy == TaskTimeTrackingReportGroupByProject {
		query = query.
			Select("task_time_tracking.tracked_at, task_time_tracking.time_spent, projects.id AS group_id, projects.title AS group_title, projects.hex_color AS group_color").
			Join("INNER", "projects", "projects.id = tasks.project_id")

		if len(opts.LabelIDs) > 0 {
			labelIDs := slices.Clone(opts.LabelIDs)
			query = query.And(builder.In("task_time_tracking.task_id",
				builder.
					Select("task_id").
					From("label_tasks").
					Where(builder.In("label_id", labelIDs)),
			))
		}
	} else {
		query = query.Select("task_time_tracking.tracked_at, task_time_tracking.time_spent, labels.id AS group_id, labels.title AS group_title, labels.hex_color AS group_color")

		if len(opts.LabelIDs) > 0 {
			labelIDs := slices.Clone(opts.LabelIDs)
			query = query.
				Join("INNER", "label_tasks", "label_tasks.task_id = tasks.id").
				Join("INNER", "labels", "labels.id = label_tasks.label_id").
				And(builder.In("label_tasks.label_id", labelIDs))
		} else {
			query = query.
				Join("LEFT", "label_tasks", "label_tasks.task_id = tasks.id").
				Join("LEFT", "labels", "labels.id = label_tasks.label_id")
		}
	}

	if err := query.Find(&rows); err != nil {
		return nil, err
	}

	return rows, nil
}
