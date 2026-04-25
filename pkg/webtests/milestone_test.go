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
	"net/url"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMilestone(t *testing.T) {
	testHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.Milestone{}
		},
		t: t,
	}

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Include parent milestones for child projects", func(t *testing.T) {
			s := db.NewSession()
			defer s.Close()

			milestone := &models.Milestone{
				ProjectID: 28,
				Name:      "Shared parent milestone",
			}
			require.NoError(t, milestone.Create(s, &testuser1))
			require.NoError(t, s.Commit())

			rec, err := testHandler.testReadAllWithUser(url.Values{"include_parents": []string{"true"}}, map[string]string{"project": "13"})
			require.NoError(t, err)

			assert.Contains(t, rec.Body.String(), `"name":"Shared parent milestone"`)
			assert.Contains(t, rec.Body.String(), `"project_id":28`)
		})
	})
}
