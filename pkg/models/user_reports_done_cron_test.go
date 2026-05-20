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
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/require"
)

func TestNotifyDoneUserReportsAt(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	s := db.NewSession()
	defer s.Close()

	report := &UserReport{
		TaskID:           1,
		ApplicationID:    1,
		ApplicationName:  "Desktop App",
		ReportTokenID:    1,
		ReportTokenLabel: "v1.0.0",
		Severity:         "high",
		ReporterEmail:    "reporter@example.com",
	}
	_, err := s.Insert(report)
	require.NoError(t, err)

	_, err = s.ID(1).Cols("done", "done_at").Update(&Task{Done: true, DoneAt: time.Now()})
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	require.NoError(t, notifyDoneUserReportsAt(time.Now()))

	s = db.NewSession()
	defer s.Close()

	storedReport, err := GetUserReportByID(s, report.ID)
	require.NoError(t, err)
	require.False(t, storedReport.NotifiedDoneAt.IsZero())
}
