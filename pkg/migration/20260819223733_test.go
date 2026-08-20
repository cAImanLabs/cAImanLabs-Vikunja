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

package migration

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func kindIndexesByID20260819223733(t *testing.T, x *xorm.Engine) map[int64]int64 {
	t.Helper()
	tasks := []*models.Task{}
	require.NoError(t, x.Find(&tasks))
	out := make(map[int64]int64, len(tasks))
	for _, tsk := range tasks {
		out[tsk.ID] = tsk.KindIndex
	}
	return out
}

func TestAddTaskKindIndex20260819223733(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, x.Sync2(&models.Project{}, &user.User{}))
	// partialSync (the same helper the migration itself uses) creates every
	// Task column but - since it ignores constraints - deliberately skips the
	// unique(tasks_kind_index) index, matching the real pre-migration schema
	// closely enough to seed several rows sharing kind_index 0 below.
	require.NoError(t, partialSync(x, &models.Task{}))
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(&models.Project{}, &user.User{}, &models.Task{}))
	})

	owner := &user.User{Username: "addtaskkindindex20260819223733", Email: "addtaskkindindex20260819223733@example.com"}
	_, err = x.Insert(owner)
	require.NoError(t, err)

	project := &models.Project{Title: "p", Identifier: "P20260819223733", OwnerID: owner.ID}
	_, err = x.Insert(project)
	require.NoError(t, err)

	// All pre-migration rows share kind_index 0 - two stories, a bug and a
	// plain task, inserted out of id/kind order to prove the backfill sorts
	// deterministically rather than relying on insertion order.
	bug := &models.Task{Title: "bug", ProjectID: project.ID, Kind: models.TaskKindBug}
	storyB := &models.Task{Title: "story b", ProjectID: project.ID, Kind: models.TaskKindStory}
	plain := &models.Task{Title: "task", ProjectID: project.ID, Kind: models.TaskKindTask}
	storyA := &models.Task{Title: "story a", ProjectID: project.ID, Kind: models.TaskKindStory}
	for _, tsk := range []*models.Task{bug, storyB, plain, storyA} {
		_, err = x.Insert(tsk)
		require.NoError(t, err)
	}

	require.NoError(t, addTaskKindIndex20260819223733(x))

	byID := kindIndexesByID20260819223733(t, x)
	require.Equal(t, int64(1), byID[bug.ID], "the only bug gets kind_index 1")
	require.Equal(t, int64(1), byID[plain.ID], "the only plain task gets kind_index 1")
	// storyB has the lower id, so ordering by (kind, id) assigns it first.
	require.Equal(t, int64(1), byID[storyB.ID], "the first story by id gets kind_index 1")
	require.Equal(t, int64(2), byID[storyA.ID], "the second story by id gets kind_index 2")

	// The unique index must actually reject a duplicate (kind, kind_index) pair.
	dup := &models.Task{Title: "dup story", ProjectID: project.ID, Kind: models.TaskKindStory, KindIndex: 1}
	_, err = x.Insert(dup)
	require.Error(t, err, "inserting a second story with kind_index 1 must violate the unique index")
}
