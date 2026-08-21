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
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// AdminTaskList overrides ReadAll to return every task on the instance;
// non-ReadAll methods inherit from Task and are gated by RequireInstanceAdmin.
type AdminTaskList struct {
	Task
}

// AdminTaskListEntry is a task plus the project it belongs to, for admin display.
// Task.ReadAll is a no-op stub, so this listing bypasses per-project permission
// scoping entirely - it is authorized purely at the route level.
type AdminTaskListEntry struct {
	Task
	ProjectTitle string `xorm:"-" json:"project_title" doc:"The title of the project this task belongs to."`
}

// ReadAll returns every task on the instance, regardless of project ownership.
// @Summary List tasks (admin)
// @Description Paginated list of every task on the instance, across all projects.
// @tags admin
// @Produce json
// @Security JWTKeyAuth
// @Param page query int false "Page number, defaults to 1."
// @Param per_page query int false "Items per page, defaults to the service setting."
// @Param s query string false "Search tasks by title or description."
// @Success 200 {array} models.AdminTaskListEntry
// @Failure 404 {object} web.HTTPError
// @Router /admin/tasks [get]
func (l *AdminTaskList) ReadAll(s *xorm.Session, _ web.Auth, search string, page, perPage int) (interface{}, int, int64, error) {
	return ListAllTasks(s, search, page, perPage)
}

// ListAllTasks returns every task on the instance, hydrated with its project title and creator.
func ListAllTasks(s *xorm.Session, search string, page, perPage int) (tasks []*AdminTaskListEntry, resultCount int, totalItems int64, err error) {
	rawTasks, resultCount, totalItems, err := getRawTasksUnscoped(s, search, page, perPage)
	if err != nil {
		return nil, 0, 0, err
	}

	projectIDs := make([]int64, 0, len(rawTasks))
	userIDs := make([]int64, 0, len(rawTasks)*2)
	for _, t := range rawTasks {
		projectIDs = append(projectIDs, t.ProjectID)
		userIDs = append(userIDs, t.CreatedByID)
		if t.CompletedByID != 0 {
			userIDs = append(userIDs, t.CompletedByID)
		}
	}

	projects, err := GetProjectsMapByIDs(s, projectIDs)
	if err != nil {
		return nil, 0, 0, err
	}
	users, err := user.GetUsersByIDs(s, userIDs)
	if err != nil {
		return nil, 0, 0, err
	}

	tasks = make([]*AdminTaskListEntry, 0, len(rawTasks))
	for _, t := range rawTasks {
		entry := &AdminTaskListEntry{Task: *t}
		if p, ok := projects[t.ProjectID]; ok {
			entry.ProjectTitle = p.Title
		}
		if c, ok := users[t.CreatedByID]; ok {
			entry.CreatedBy = c
		}
		if c, ok := users[t.CompletedByID]; ok {
			entry.CompletedBy = c
		}
		tasks = append(tasks, entry)
	}

	return tasks, resultCount, totalItems, nil
}

// ReassignTaskCompletedBy credits a different user with completing an
// already-done task, for cases where whoever marked it done wasn't the one
// who actually did the work.
func ReassignTaskCompletedBy(s *xorm.Session, doer *user.User, taskID, completedByID int64) (*Task, error) {
	t, err := GetTaskByIDSimple(s, taskID)
	if err != nil {
		return nil, err
	}
	if !t.Done {
		return nil, ErrInvalidData{Message: "task is not done"}
	}

	completedBy, err := user.GetUserByID(s, completedByID)
	if err != nil {
		return nil, err
	}

	oldCompletedByID := t.CompletedByID
	t.CompletedByID = completedByID
	if _, err := s.ID(t.ID).Cols("completed_by_id").Update(&t); err != nil {
		return nil, err
	}
	t.CompletedBy = completedBy

	events.DispatchOnCommit(s, &AdminTaskCompletedByChangedEvent{
		Task:             &t,
		Doer:             doer,
		OldCompletedByID: oldCompletedByID,
		NewCompletedByID: completedByID,
	})
	return &t, nil
}

func getRawTasksUnscoped(s *xorm.Session, search string, page, perPage int) (tasks []*Task, resultCount int, totalItems int64, err error) {
	limit, start := getLimitFromPageIndex(page, perPage)

	conds := []builder.Cond{}
	if search != "" {
		ids := []int64{}
		for _, val := range strings.Split(search, ",") {
			v, parseErr := strconv.ParseInt(val, 10, 64)
			if parseErr != nil {
				log.Debugf("Task search string part '%s' is not a number: %s", val, parseErr)
				continue
			}
			ids = append(ids, v)
		}
		if len(ids) > 0 {
			conds = append(conds, builder.In("id", ids))
		} else {
			conds = append(conds, db.MultiFieldSearchWithTableAlias(
				[]string{"title", "description"},
				search,
				"",
			))
		}
	}
	var where = builder.Expr("1 = 1")
	if len(conds) > 0 {
		where = builder.And(conds...)
	}

	query := s.Where(where).OrderBy("id DESC")
	if limit > 0 {
		query = query.Limit(limit, start)
	}

	tasks = []*Task{}
	if err = query.Find(&tasks); err != nil {
		return nil, 0, 0, err
	}

	totalItems, err = s.Where(where).Count(&Task{})
	if err != nil {
		return nil, 0, 0, err
	}

	return tasks, len(tasks), totalItems, nil
}
