// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package models

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// createBurndownTask inserts a task directly (bypassing Task.Update's
// server-managed done_at) so tests can simulate a task having been done on a
// specific past date.
func createBurndownTask(t *testing.T, s *xorm.Session, sprintID int64, title string, storyPoints int, done bool, doneAt time.Time) {
	t.Helper()
	task := &Task{
		Title:       title,
		ProjectID:   1,
		SprintID:    sprintID,
		StoryPoints: storyPoints,
	}
	require.NoError(t, task.Create(s, &user.User{ID: 1}))

	if done {
		_, err := s.ID(task.ID).Cols("done", "done_at").Update(&Task{Done: true, DoneAt: doneAt})
		require.NoError(t, err)
	}
}

func TestGetSprintBurndown(t *testing.T) {
	t.Run("computes a daily remaining series from task done_at", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		now := time.Now()
		start := now.AddDate(0, 0, -3)
		end := now.AddDate(0, 0, 3)

		_, err := s.ID(int64(1)).Cols("start_date", "end_date").Update(&Sprint{StartDate: start, EndDate: end})
		require.NoError(t, err)

		createBurndownTask(t, s, 1, "done day 0", 5, true, start)
		createBurndownTask(t, s, 1, "done day 2", 3, true, start.AddDate(0, 0, 2))
		createBurndownTask(t, s, 1, "not done", 2, false, time.Time{})

		require.NoError(t, s.Commit())
		s2 := db.NewSession()
		defer s2.Close()

		burndown, err := GetSprintBurndown(s2, 1, 1)
		require.NoError(t, err)

		assert.Equal(t, 10, burndown.TotalPoints)
		require.NotEmpty(t, burndown.Series)

		// Day 0: only the first task (5 points) is done -> 10-5=5 remaining.
		assert.Equal(t, 5, burndown.Series[0].Remaining)
		// By day 2, both done tasks (5+3=8) are done -> 10-8=2 remaining.
		require.Greater(t, len(burndown.Series), 2)
		assert.Equal(t, 2, burndown.Series[2].Remaining)
		// The series never goes past today, even though the sprint ends later.
		assert.False(t, burndown.Series[len(burndown.Series)-1].Date.After(truncateToDay(now)))
	})

	t.Run("errors when the sprint has no dates", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Fixture sprint 3 (project 2) has no start/end date set.
		_, err := GetSprintBurndown(s, 2, 3)
		require.Error(t, err)
		assert.True(t, IsErrSprintMissingDates(err))
	})

	t.Run("errors when the sprint belongs to a different project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := GetSprintBurndown(s, 999, 1)
		require.Error(t, err)
		assert.True(t, IsErrSprintDoesNotExist(err))
	})
}

func TestGetProjectVelocity(t *testing.T) {
	t.Run("sums total and completed points per sprint", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		createBurndownTask(t, s, 1, "v-done", 5, true, time.Now())
		createBurndownTask(t, s, 1, "v-not-done", 3, false, time.Time{})
		createBurndownTask(t, s, 2, "v2-done", 8, true, time.Now())

		require.NoError(t, s.Commit())
		s2 := db.NewSession()
		defer s2.Close()

		velocity, err := GetProjectVelocity(s2, 1)
		require.NoError(t, err)
		require.Len(t, velocity, 2)

		bySprintID := map[int64]*SprintVelocityPoint{}
		for _, v := range velocity {
			bySprintID[v.SprintID] = v
		}

		assert.Equal(t, 8, bySprintID[1].TotalPoints)
		assert.Equal(t, 5, bySprintID[1].CompletedPoints)
		assert.Equal(t, 8, bySprintID[2].TotalPoints)
		assert.Equal(t, 8, bySprintID[2].CompletedPoints)
	})

	t.Run("empty when the project has no sprints", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		velocity, err := GetProjectVelocity(s, 3)
		require.NoError(t, err)
		assert.Empty(t, velocity)
	})
}
