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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTaskCompletionStatsByUser(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// Task 2 is already done in fixtures; tasks 3 and 6 are forced done here
	// so both user1 and user2 show up with distinct, known totals.
	_, err := s.ID(int64(2)).Cols("done", "completed_by_id", "story_points").Update(&Task{Done: true, CompletedByID: 1, StoryPoints: 3})
	require.NoError(t, err)
	_, err = s.ID(int64(3)).Cols("done", "completed_by_id", "story_points").Update(&Task{Done: true, CompletedByID: 1, StoryPoints: 2})
	require.NoError(t, err)
	_, err = s.ID(int64(6)).Cols("done", "completed_by_id", "story_points").Update(&Task{Done: true, CompletedByID: 2, StoryPoints: 5})
	require.NoError(t, err)

	stats, err := GetTaskCompletionStatsByUser(s)
	require.NoError(t, err)

	byUser := make(map[int64]*UserTaskCompletionStat, len(stats))
	for _, stat := range stats {
		byUser[stat.UserID] = stat
	}

	require.Contains(t, byUser, int64(1))
	assert.Equal(t, int64(2), byUser[1].Completed)
	assert.Equal(t, int64(5), byUser[1].StoryPoints)
	assert.Equal(t, "user1", byUser[1].Username)

	require.Contains(t, byUser, int64(2))
	assert.Equal(t, int64(1), byUser[2].Completed)
	assert.Equal(t, int64(5), byUser[2].StoryPoints)

	// Ordered by completed count descending.
	require.GreaterOrEqual(t, len(stats), 2)
	assert.Equal(t, int64(1), stats[0].UserID, "user1 has the most completions and sorts first")
}
