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

func completedByIDsByID20260821010743(t *testing.T, x *xorm.Engine) map[int64]int64 {
	t.Helper()
	tasks := []*models.Task{}
	require.NoError(t, x.Find(&tasks))
	out := make(map[int64]int64, len(tasks))
	for _, tsk := range tasks {
		out[tsk.ID] = tsk.CompletedByID
	}
	return out
}

func TestAddTaskCompletedByField20260821010743(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, x.Sync2(&models.Project{}, &user.User{}))
	// partialSync (the same helper the migration itself uses) creates every
	// Task column, including completed_by_id.
	require.NoError(t, partialSync(x, &models.Task{}))
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(&models.Project{}, &user.User{}, &models.Task{}))
	})

	owner := &user.User{Username: "addtaskcompletedby20260821010743", Email: "addtaskcompletedby20260821010743@example.com"}
	_, err = x.Insert(owner)
	require.NoError(t, err)

	project := &models.Project{Title: "p", Identifier: "P20260821010743", OwnerID: owner.ID}
	_, err = x.Insert(project)
	require.NoError(t, err)

	done := &models.Task{Title: "done", ProjectID: project.ID, CreatedByID: owner.ID, Done: true}
	notDone := &models.Task{Title: "not done", ProjectID: project.ID, CreatedByID: owner.ID, Done: false}
	for _, tsk := range []*models.Task{done, notDone} {
		_, err = x.Insert(tsk)
		require.NoError(t, err)
	}

	// ADD COLUMN leaves completed_by_id actual SQL NULL on every pre-existing
	// row in a real upgrade, not the Go zero value 0 that a fresh ORM Insert
	// produces - force both rows into that state so the backfill query is
	// exercised against the condition it actually has to handle in production.
	_, err = x.Table("tasks").In("id", done.ID, notDone.ID).Update(map[string]interface{}{"completed_by_id": nil})
	require.NoError(t, err)

	require.NoError(t, addTaskCompletedByField20260821010743(x))

	byID := completedByIDsByID20260821010743(t, x)
	require.Equal(t, owner.ID, byID[done.ID], "an already-done task with a NULL completed_by_id is backfilled with its creator")
	require.Zero(t, byID[notDone.ID], "a not-done task is left without a completed_by_id")
}
