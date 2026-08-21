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
	"sort"

	"code.vikunja.io/api/pkg/user"

	"xorm.io/xorm"
)

// UserTaskCompletionStat is one user's task-completion totals, the data
// behind the admin overview's per-user completion charts.
type UserTaskCompletionStat struct {
	UserID      int64  `json:"user_id" doc:"The id of the user credited with these completions."`
	Username    string `json:"username" doc:"The username of the user credited with these completions."`
	Completed   int64  `json:"completed" doc:"The number of tasks credited as completed by this user."`
	StoryPoints int64  `json:"story_points" doc:"The sum of story points across the tasks credited as completed by this user."`
}

// GetTaskCompletionStatsByUser returns, for every user credited with
// completing at least one task, their total completed-task count and summed
// story points. Ordered by completed count descending, ties broken by
// username for a stable chart order.
func GetTaskCompletionStatsByUser(s *xorm.Session) ([]*UserTaskCompletionStat, error) {
	tasks := []*Task{}
	err := s.
		Where("done = ? AND completed_by_id > 0", true).
		Cols("completed_by_id", "story_points").
		Find(&tasks)
	if err != nil {
		return nil, err
	}

	completed := map[int64]int64{}
	storyPoints := map[int64]int64{}
	userIDs := make([]int64, 0, len(tasks))
	seen := map[int64]bool{}
	for _, t := range tasks {
		completed[t.CompletedByID]++
		storyPoints[t.CompletedByID] += int64(t.StoryPoints)
		if !seen[t.CompletedByID] {
			seen[t.CompletedByID] = true
			userIDs = append(userIDs, t.CompletedByID)
		}
	}

	users, err := user.GetUsersByIDs(s, userIDs)
	if err != nil {
		return nil, err
	}

	result := make([]*UserTaskCompletionStat, 0, len(userIDs))
	for _, id := range userIDs {
		username := ""
		if u, ok := users[id]; ok {
			username = u.Username
		}
		result = append(result, &UserTaskCompletionStat{
			UserID:      id,
			Username:    username,
			Completed:   completed[id],
			StoryPoints: storyPoints[id],
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Completed != result[j].Completed {
			return result[i].Completed > result[j].Completed
		}
		return result[i].Username < result[j].Username
	})

	return result, nil
}
