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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaSprint covers sprint CRUD on /api/v2. Sprints live under
// /projects/{project}/sprints, scoped directly to a project (no project view
// involved). Permission model mirrors buckets: Sprint.Can{Create,Update,Delete}
// all delegate to Project.CanUpdate (write access), Sprint.ReadAll only needs
// the project's read access.
//
// Fixture topology (see pkg/db/fixtures/sprints.yml):
//   - project 1 (owned by testuser1): sprints 1, 2.
//   - project 2 (owned by user3, no share to testuser1): sprint 3 — the
//     forbidden / non-member negative.
//   - projects 9/10/11 are owned by user6 and shared to testuser1
//     read/write/admin; they carry sprints 4/5/6 respectively.
func TestHumaSprint(t *testing.T) {
	owned := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/1/sprints",
		idParam:  "sprint",
		t:        t,
	}
	require.NoError(t, owned.ensureEnv())
	forbidden := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/2/sprints",
		idParam:  "sprint",
		t:        t,
		e:        owned.e,
	}
	readShared := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/9/sprints",
		idParam:  "sprint",
		t:        t,
		e:        owned.e,
	}
	writeShared := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/10/sprints",
		idParam:  "sprint",
		t:        t,
		e:        owned.e,
	}
	adminShared := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/11/sprints",
		idParam:  "sprint",
		t:        t,
		e:        owned.e,
	}

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			ids := sprintsFromReadAll(t, rec.Body.Bytes())
			assert.ElementsMatch(t, []int64{1, 2}, ids)
			assert.Contains(t, rec.Body.String(), `"total":2`)
		})
		t.Run("Read-only share can list", func(t *testing.T) {
			rec, err := readShared.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			ids := sprintsFromReadAll(t, rec.Body.Bytes())
			assert.ElementsMatch(t, []int64{4}, ids)
		})
		t.Run("Forbidden", func(t *testing.T) {
			_, err := forbidden.testReadAllWithUser(nil, nil)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testCreateWithUser(nil, nil, `{"title":"New sprint","goal":"Ship it"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"title":"New sprint"`)
			assert.Contains(t, rec.Body.String(), `"status":"planning"`)
			assert.Contains(t, rec.Body.String(), `"project_id":1`)
		})
		t.Run("Write share can create", func(t *testing.T) {
			rec, err := writeShared.testCreateWithUser(nil, nil, `{"title":"Write made"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"project_id":10`)
		})
		t.Run("Read share cannot create", func(t *testing.T) {
			_, err := readShared.testCreateWithUser(nil, nil, `{"title":"Nope"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden", func(t *testing.T) {
			_, err := forbidden.testCreateWithUser(nil, nil, `{"title":"Nope"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Empty title", func(t *testing.T) {
			_, err := owned.testCreateWithUser(nil, nil, `{"title":""}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
		t.Run("Invalid status", func(t *testing.T) {
			_, err := owned.testCreateWithUser(nil, nil, `{"title":"x","status":"bogus"}`)
			require.Error(t, err)
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testUpdateWithUser(nil, map[string]string{"sprint": "1"}, `{"title":"Renamed","status":"active"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Renamed"`)
			assert.Contains(t, rec.Body.String(), `"status":"active"`)
			db.AssertExists(t, "sprints", map[string]interface{}{
				"id":            1,
				"title":         "Renamed",
				"project_id":    1,
				"created_by_id": 1,
			}, false)
		})
		t.Run("Write share can update", func(t *testing.T) {
			rec, err := writeShared.testUpdateWithUser(nil, map[string]string{"sprint": "5"}, `{"title":"Write renamed"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Write renamed"`)
		})
		t.Run("Read share cannot update", func(t *testing.T) {
			_, err := readShared.testUpdateWithUser(nil, map[string]string{"sprint": "4"}, `{"title":"x"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := owned.testUpdateWithUser(nil, map[string]string{"sprint": "9999"}, `{"title":"x"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Forbidden", func(t *testing.T) {
			_, err := forbidden.testUpdateWithUser(nil, map[string]string{"sprint": "3"}, `{"title":"x"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Read share cannot delete", func(t *testing.T) {
			_, err := readShared.testDeleteWithUser(nil, map[string]string{"sprint": "4"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Admin share can delete", func(t *testing.T) {
			rec, err := adminShared.testDeleteWithUser(nil, map[string]string{"sprint": "6"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
			db.AssertMissing(t, "sprints", map[string]interface{}{"id": 6})
		})
		t.Run("Forbidden", func(t *testing.T) {
			_, err := forbidden.testDeleteWithUser(nil, map[string]string{"sprint": "3"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testDeleteWithUser(nil, map[string]string{"sprint": "2"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			db.AssertMissing(t, "sprints", map[string]interface{}{"id": 2})
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := owned.testDeleteWithUser(nil, map[string]string{"sprint": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
	})
}

func sprintsFromReadAll(t *testing.T, body []byte) (ids []int64) {
	t.Helper()
	var resp struct {
		Items []struct {
			ID int64 `json:"id"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(body, &resp), "ReadAll body must be a paginated envelope: %s", string(body))
	ids = make([]int64, 0, len(resp.Items))
	for _, it := range resp.Items {
		ids = append(ids, it.ID)
	}
	return ids
}
