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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The error body shape is covered by TestHuma_ErrorShapeIsRFC9457; this test
// only asserts gate status codes (404 on failure, matching v1).
func TestHumaAdminProjects(t *testing.T) {
	t.Run("non-admin user gets 404", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, 1)
		require.NoError(t, err)
		require.False(t, u.IsAdmin, "fixture precondition: user1 is not an admin")

		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/projects", u, "")
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("admin with the feature sees every project", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/projects", admin, "")
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())

		var envelope struct {
			Items []struct {
				ID int64 `json:"id"`
			} `json:"items"`
			Total int64 `json:"total"`
		}
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &envelope))

		ids := make(map[int64]bool, len(envelope.Items))
		for _, item := range envelope.Items {
			ids[item.ID] = true
		}
		// Project 6 (owned by user6, not shared with user1) proves the list ignores ownership.
		assert.True(t, ids[6], "expected project 6 in the admin list, got items %v", ids)
		// Project 22 is archived, proving the list includes archived projects.
		assert.True(t, ids[22], "expected archived project 22 in the admin list, got items %v", ids)

		// Ported from v1 TestAdmin_ListProjects (admin_test.go:222-226): the
		// response body must carry project fields and a hydrated owner.
		body := res.Body.String()
		assert.Contains(t, body, `"id":`)
		assert.Contains(t, body, `"title":`)
		// Owner is xorm:"-" and must be hydrated explicitly (project 1 is owned by user1).
		assert.Contains(t, body, `"username":"user1"`)
		assert.NotContains(t, body, `"owner":null`)
	})

	t.Run("unauthenticated caller gets 401", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		// The token middleware rejects with 401 before the gate runs, matching v1.
		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/projects", nil, "")
		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})
}

func TestHumaAdminTasks(t *testing.T) {
	t.Run("non-admin user gets 404", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, 1)
		require.NoError(t, err)
		require.False(t, u.IsAdmin, "fixture precondition: user1 is not an admin")

		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/tasks", u, "")
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("admin sees every task", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/tasks", admin, "")
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())

		var envelope struct {
			Items []struct {
				ID    int64  `json:"id"`
				Title string `json:"title"`
			} `json:"items"`
			Total int64 `json:"total"`
		}
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &envelope))

		titles := make(map[string]bool, len(envelope.Items))
		for _, item := range envelope.Items {
			titles[item.Title] = true
		}
		// Task 15 belongs to project 6, owned by user6 and not shared with user1.
		assert.True(t, titles["task #15"], "expected task #15 in the admin list, got %v", titles)

		body := res.Body.String()
		assert.Contains(t, body, `"project_title":`)
	})

	t.Run("unauthenticated caller gets 401", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/tasks", nil, "")
		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})
}

func TestHumaAdminPatchTaskCompletedBy(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	t.Run("credits a different user", func(t *testing.T) {
		// Task 2 is fixture-seeded as done.
		res := adminReq(t, e, http.MethodPatch, "/api/v2/admin/tasks/2/completed-by", admin, `{"completed_by_id":2}`)
		assert.Equal(t, http.StatusOK, res.Code, res.Body.String())

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":              2,
			"completed_by_id": 2,
		}, false)
	})

	t.Run("rejects a task that isn't done", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v2/admin/tasks/3/completed-by", admin, `{"completed_by_id":2}`)
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("nonexistent task returns 404", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v2/admin/tasks/99999/completed-by", admin, `{"completed_by_id":2}`)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("non-admin caller gets 404", func(t *testing.T) {
		s := db.NewSession()
		u, err := user.GetUserByID(s, 2)
		require.NoError(t, err)
		s.Close()

		res := adminReq(t, e, http.MethodPatch, "/api/v2/admin/tasks/2/completed-by", u, `{"completed_by_id":2}`)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})
}

func TestHumaAdminTaskCompletionStats(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	s := db.NewSession()
	_, err = s.ID(int64(2)).Cols("completed_by_id", "story_points").Update(&models.Task{CompletedByID: 1, StoryPoints: 3})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()

	res := adminReq(t, e, http.MethodGet, "/api/v2/admin/tasks/completion-stats", admin, "")
	assert.Equal(t, http.StatusOK, res.Code, res.Body.String())
	body := res.Body.String()
	assert.Contains(t, body, `"user_id":1`)
	assert.Contains(t, body, `"username":"user1"`)
	assert.Contains(t, body, `"completed":1`)
	assert.Contains(t, body, `"story_points":3`)
}
