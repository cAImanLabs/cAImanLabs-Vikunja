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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixture topology (pkg/db/fixtures/sprints.yml), mirroring the bucket tests:
//   - project 1, owned by user 1 (testuser1): sprints 1, 2.
//   - project 2, owned by user 3, no share to user 1: sprint 3.
//   - project 9 (read share to user 1): sprint 4.
//   - project 10 (write share to user 1): sprint 5.
//   - project 11 (admin share to user 1): sprint 6.
func TestSprint_ReadAll(t *testing.T) {
	t.Run("owner can list", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		result, resultCount, total, err := (&Sprint{ProjectID: 1}).ReadAll(s, &user.User{ID: 1}, "", 1, 50)
		require.NoError(t, err)
		sprints, ok := result.([]*Sprint)
		require.True(t, ok)
		assert.Equal(t, 2, resultCount)
		assert.EqualValues(t, 2, total)
		ids := []int64{sprints[0].ID, sprints[1].ID}
		assert.ElementsMatch(t, []int64{1, 2}, ids)
	})

	t.Run("read share can list", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		result, resultCount, _, err := (&Sprint{ProjectID: 9}).ReadAll(s, &user.User{ID: 1}, "", 1, 50)
		require.NoError(t, err)
		sprints, ok := result.([]*Sprint)
		require.True(t, ok)
		assert.Equal(t, 1, resultCount)
		assert.EqualValues(t, 4, sprints[0].ID)
	})

	t.Run("forbidden without project access", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, _, _, err := (&Sprint{ProjectID: 2}).ReadAll(s, &user.User{ID: 1}, "", 1, 50)
		require.Error(t, err)
		assert.True(t, IsErrGenericForbidden(err))
	})
}

func TestSprint_CanCreate(t *testing.T) {
	t.Run("owner can create", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		can, err := (&Sprint{ProjectID: 1}).CanCreate(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("write share can create", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		can, err := (&Sprint{ProjectID: 10}).CanCreate(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("read share cannot create", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		can, err := (&Sprint{ProjectID: 9}).CanCreate(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("no access cannot create", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		can, err := (&Sprint{ProjectID: 2}).CanCreate(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestSprint_CanUpdate(t *testing.T) {
	t.Run("owner can update", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		can, err := (&Sprint{ID: 1, ProjectID: 1}).CanUpdate(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("write share can update", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		can, err := (&Sprint{ID: 5, ProjectID: 10}).CanUpdate(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("read share cannot update", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		can, err := (&Sprint{ID: 4, ProjectID: 9}).CanUpdate(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("no access cannot update", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		can, err := (&Sprint{ID: 3, ProjectID: 2}).CanUpdate(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("nonexistent sprint errors", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := (&Sprint{ID: 9999}).CanUpdate(s, &user.User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrSprintDoesNotExist(err))
	})
}

func TestSprint_CanDelete(t *testing.T) {
	t.Run("admin share can delete", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		can, err := (&Sprint{ID: 6, ProjectID: 11}).CanDelete(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("read share cannot delete", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		can, err := (&Sprint{ID: 4, ProjectID: 9}).CanDelete(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestSprint_CreateDefaultsToPlanning(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sp := &Sprint{ProjectID: 1, Title: "New sprint"}
	require.NoError(t, sp.Create(s, &user.User{ID: 1}))
	assert.Equal(t, SprintStatusPlanning, sp.Status)
	assert.NotZero(t, sp.ID)
	assert.Equal(t, int64(1), sp.CreatedByID)
}
