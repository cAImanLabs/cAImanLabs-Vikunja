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

package models

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTask_Update_CompletedBy(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("marking a task done sets completed_by to the doer", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{ID: 3, Done: true}
		err := task.Update(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		assert.Equal(t, int64(1), task.CompletedByID)
		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":              3,
			"completed_by_id": 1,
		}, false)
	})

	t.Run("reopening a done task clears completed_by", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Task 2 is fixture-seeded as done, with no completed_by_id set - mirroring
		// a task that was done before this field existed.
		task := &Task{ID: 2, Done: false}
		err := task.Update(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		assert.Zero(t, task.CompletedByID)
		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":              2,
			"completed_by_id": 0,
		}, false)
	})

	t.Run("updating an unrelated field on a done task leaves completed_by alone", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{ID: 2, Title: "renamed"}
		err := task.updateSingleTask(s, u, []string{"title"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		assert.Zero(t, task.CompletedByID, "completed_by_id isn't in the field whitelist, so a partial update must not touch it")
	})
}

func TestReassignTaskCompletedBy(t *testing.T) {
	doer := &user.User{ID: 1}

	t.Run("credits the given user on a done task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task, err := ReassignTaskCompletedBy(s, doer, 2, 2)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		assert.Equal(t, int64(2), task.CompletedByID)
		require.NotNil(t, task.CompletedBy)
		assert.Equal(t, int64(2), task.CompletedBy.ID)
		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":              2,
			"completed_by_id": 2,
		}, false)
	})

	t.Run("refuses a task that isn't done", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := ReassignTaskCompletedBy(s, doer, 3, 2)
		require.Error(t, err)
		assert.True(t, IsErrInvalidData(err))
	})

	t.Run("nonexistent task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := ReassignTaskCompletedBy(s, doer, 99999, 2)
		require.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
	})

	t.Run("nonexistent target user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := ReassignTaskCompletedBy(s, doer, 2, 99999)
		require.Error(t, err)
	})
}
