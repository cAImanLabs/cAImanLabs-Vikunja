// cAImanDesk is a to-do list application to facilitate your life, forked from Vikunja.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
// Portions Copyright 2026-present cAImanLabs and contributors.
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
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaSprintVelocity covers GET /projects/{project}/sprints/velocity.
// Fixture topology: project 1 (owned by testuser1) carries sprints 1 and 2
// (see pkg/db/fixtures/sprints.yml); project 2 (owned by user3) is used as
// the forbidden negative.
func TestHumaSprintVelocity(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	user1Token := humaTokenFor(t, &testuser1)
	user2Token := humaTokenFor(t, &testuser2)

	t.Run("sums total and completed story points per sprint", func(t *testing.T) {
		s := db.NewSession()
		task1 := &models.Task{Title: "velocity task done", ProjectID: 1, SprintID: 1, StoryPoints: 5}
		require.NoError(t, task1.Create(s, &user.User{ID: 1}))
		_, err := s.ID(task1.ID).Cols("done").Update(&models.Task{Done: true})
		require.NoError(t, err)
		task2 := &models.Task{Title: "velocity task not done", ProjectID: 1, SprintID: 1, StoryPoints: 3}
		require.NoError(t, task2.Create(s, &user.User{ID: 1}))
		require.NoError(t, s.Commit())
		require.NoError(t, s.Close())

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/1/sprints/velocity", "", user1Token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		var points []*models.SprintVelocityPoint
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &points))
		require.Len(t, points, 2)

		bySprintID := map[int64]*models.SprintVelocityPoint{}
		for _, p := range points {
			bySprintID[p.SprintID] = p
		}
		require.Contains(t, bySprintID, int64(1))
		assert.Equal(t, 8, bySprintID[1].TotalPoints)
		assert.Equal(t, 5, bySprintID[1].CompletedPoints)
	})

	t.Run("forbidden without project access", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/1/sprints/velocity", "", user2Token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("empty list for a project with no sprints", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/3/sprints/velocity", "", user2Token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.JSONEq(t, "[]", rec.Body.String())
	})
}

// TestHumaSprintBurndown covers GET /projects/{project}/sprints/{sprint}/burndown.
func TestHumaSprintBurndown(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	user1Token := humaTokenFor(t, &testuser1)
	user2Token := humaTokenFor(t, &testuser2)

	t.Run("missing dates returns 412", func(t *testing.T) {
		// Fixture sprint 1 has no start/end date set.
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/1/sprints/1/burndown", "", user1Token, "")
		assert.Equal(t, http.StatusPreconditionFailed, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("computes the daily remaining series", func(t *testing.T) {
		now := time.Now()
		start := now.AddDate(0, 0, -2)
		end := now.AddDate(0, 0, 2)

		s := db.NewSession()
		_, err := s.ID(int64(2)).Cols("start_date", "end_date").Update(&models.Sprint{StartDate: start, EndDate: end})
		require.NoError(t, err)

		task := &models.Task{Title: "burndown task", ProjectID: 1, SprintID: 2, StoryPoints: 5}
		require.NoError(t, task.Create(s, &user.User{ID: 1}))
		_, err = s.ID(task.ID).Cols("done", "done_at").Update(&models.Task{Done: true, DoneAt: start})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		require.NoError(t, s.Close())

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/1/sprints/2/burndown", "", user1Token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		var burndown models.SprintBurndown
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &burndown))
		assert.Equal(t, int64(2), burndown.SprintID)
		assert.Equal(t, 5, burndown.TotalPoints)
		require.NotEmpty(t, burndown.Series)
		// The only task was done on the sprint's start day, so nothing remains from day 0 on.
		assert.Equal(t, 0, burndown.Series[0].Remaining)
	})

	t.Run("forbidden without project access", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/1/sprints/2/burndown", "", user2Token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("nonexistent sprint", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/1/sprints/9999/burndown", "", user1Token, "")
		assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
	})
}
