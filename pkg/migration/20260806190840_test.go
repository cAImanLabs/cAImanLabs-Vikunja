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

package migration

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func sprintViewKindsForProject20260806190840(t *testing.T, x *xorm.Engine, projectID int64) []int {
	t.Helper()
	views := []*projectView20260806190840{}
	require.NoError(t, x.Where("project_id = ?", projectID).Find(&views))
	kinds := make([]int, 0, len(views))
	for _, v := range views {
		kinds = append(kinds, v.ViewKind)
	}
	return kinds
}

// Syncing against the real models.Project/models.ProjectView tables (not this
// migration's own mirror structs) is deliberate: it is what caught this
// migration's mirror struct missing project_views.created/updated (both
// NOT NULL on the real table), which a self-consistent Sync2 of the mirror
// alone could never surface.
func TestBackfillSprintViews20260806190840(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, x.Sync2(&models.Project{}, &models.ProjectView{}, &user.User{}))
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(&models.Project{}, &models.ProjectView{}, &user.User{}))
	})

	owner := &user.User{Username: "backfillsprintviews20260806190840", Email: "backfillsprintviews20260806190840@example.com"}
	_, err = x.Insert(owner)
	require.NoError(t, err)

	// Project 101: pre-existing, no sprint view yet (created before the default-view fix).
	_, err = x.Insert(&models.Project{ID: 101, Title: "p101", Identifier: "P101B20260806190840", OwnerID: owner.ID})
	require.NoError(t, err)
	_, err = x.Insert(&models.ProjectView{ProjectID: 101, Title: "List", ViewKind: models.ProjectViewKindList, Position: 100})
	require.NoError(t, err)

	// Project 102: already has a sprint view (e.g. added manually by the user).
	_, err = x.Insert(&models.Project{ID: 102, Title: "p102", Identifier: "P102B20260806190840", OwnerID: owner.ID})
	require.NoError(t, err)
	_, err = x.Insert(&models.ProjectView{ProjectID: 102, Title: "My Sprints", ViewKind: models.ProjectViewKindSprint, Position: 500})
	require.NoError(t, err)

	require.NoError(t, backfillSprintViews20260806190840(x))

	require.ElementsMatch(t, []int{0, projectViewKindSprint20260806190840}, sprintViewKindsForProject20260806190840(t, x, 101),
		"project 101 must get a backfilled sprint view")

	views102 := sprintViewKindsForProject20260806190840(t, x, 102)
	require.Len(t, views102, 1, "project 102 already had a sprint view; the migration must not add a second one")

	title := &models.ProjectView{}
	exists, err := x.Where("project_id = ? AND view_kind = ?", 102, projectViewKindSprint20260806190840).Get(title)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, "My Sprints", title.Title, "the migration must not touch an existing sprint view's data")

	// Idempotent: running it again must not create duplicates.
	require.NoError(t, backfillSprintViews20260806190840(x))
	require.Len(t, sprintViewKindsForProject20260806190840(t, x, 101), 2)
}
