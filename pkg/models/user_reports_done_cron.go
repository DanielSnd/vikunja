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
	"fmt"
	"html"
	"sort"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/notifications"

	"xorm.io/builder"
	"xorm.io/xorm"
)

type userReportDoneNotificationCandidate struct {
	ID               int64     `xorm:"'id'"`
	TaskID           int64     `xorm:"'task_id'"`
	ApplicationName  string    `xorm:"'application_name'"`
	ReportTokenLabel string    `xorm:"'report_token_label'"`
	ReporterEmail    string    `xorm:"'reporter_email'"`
	NotifiedDoneAt   time.Time `xorm:"'notified_done_at'"`
	TaskTitle        string    `xorm:"'task_title'"`
}

func RegisterUserReportDoneNotificationCron() {
	err := cron.Schedule("0 9 * * *", func() {
		if err := notifyDoneUserReportsAt(time.Now()); err != nil {
			log.Errorf("Could not send done user report notifications: %s", err)
		}
	})
	if err != nil {
		log.Fatalf("Could not register user report done notification cron: %s", err)
	}
}

func notifyDoneUserReportsAt(_ time.Time) error {
	s := db.NewSession()
	defer s.Close()

	candidates, err := getDoneUserReportNotificationCandidates(s)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return nil
	}

	grouped := map[string][]userReportDoneNotificationCandidate{}
	for _, candidate := range candidates {
		grouped[candidate.ReporterEmail] = append(grouped[candidate.ReporterEmail], candidate)
	}

	for email, reports := range grouped {
		if err := sendDoneUserReportMail(email, reports); err != nil {
			return err
		}

		reportIDs := make([]int64, 0, len(reports))
		for _, report := range reports {
			reportIDs = append(reportIDs, report.ID)
		}

		_, err = s.In("id", reportIDs).Cols("notified_done_at").Update(&UserReport{NotifiedDoneAt: time.Now()})
		if err != nil {
			return err
		}
	}

	return s.Commit()
}

func getDoneUserReportNotificationCandidates(s *xorm.Session) ([]userReportDoneNotificationCandidate, error) {
	candidates := []userReportDoneNotificationCandidate{}

	err := s.
		Table("user_reports").
		Select("user_reports.id, user_reports.task_id, user_reports.application_name, user_reports.report_token_label, user_reports.reporter_email, user_reports.notified_done_at, tasks.title as task_title").
		Join("INNER", "tasks", "tasks.id = user_reports.task_id").
		Where(builder.And(
			builder.Expr("user_reports.reporter_email <> ''"),
			builder.Expr("user_reports.notified_done_at IS NULL"),
			builder.Eq{"tasks.done": true},
		)).
		Find(&candidates)

	return candidates, err
}

func sendDoneUserReportMail(email string, reports []userReportDoneNotificationCandidate) error {
	sort.Slice(reports, func(i, j int) bool {
		if reports[i].ApplicationName == reports[j].ApplicationName {
			return reports[i].TaskTitle < reports[j].TaskTitle
		}
		return reports[i].ApplicationName < reports[j].ApplicationName
	})

	introKey := "notifications.user_report.done.message.single"
	subjectKey := "notifications.user_report.done.subject.single"
	if len(reports) > 1 {
		introKey = "notifications.user_report.done.message.multiple"
		subjectKey = "notifications.user_report.done.subject.multiple"
	}

	firstAppName := reports[0].ApplicationName
	subject := i18n.T("en", subjectKey, firstAppName)

	body := strings.Builder{}
	body.WriteString("<ul>")
	for _, report := range reports {
		taskURL := config.ServicePublicURL.GetString() + "tasks/" + fmt.Sprintf("%d", report.TaskID)
		body.WriteString("<li>")
		body.WriteString(`<a href="` + taskURL + `">` + html.EscapeString(report.TaskTitle) + `</a>`)
		body.WriteString(" (" + html.EscapeString(report.ApplicationName))
		if report.ReportTokenLabel != "" {
			body.WriteString(", " + html.EscapeString(report.ReportTokenLabel))
		}
		body.WriteString(")</li>")
	}
	body.WriteString("</ul>")

	m := notifications.NewMail().
		To(email).
		Subject(subject).
		Line(i18n.T("en", introKey, firstAppName)).
		HTML(body.String()).
		FooterLine(i18n.T("en", "notifications.user_report.done.footer"))

	return notifications.SendMail(m, "en")
}
