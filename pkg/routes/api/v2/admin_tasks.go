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

package apiv2

import (
	"context"
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
)

type adminTaskListBody struct {
	Body Paginated[*models.AdminTaskListEntry]
}

type adminTaskBody struct {
	Body *models.Task
}

type adminTaskCompletionStatsBody struct {
	Body []*models.UserTaskCompletionStat
}

// adminCompletedByPatchBody credits a different user with completing a task.
type adminCompletedByPatchBody struct {
	CompletedByID int64 `json:"completed_by_id" minimum:"1" doc:"The numeric ID of the user who should be credited with completing the task."`
}

// Permissions are enforced by the gateV2AdminRoutes path middleware, not per-handler.
func RegisterAdminTaskRoutes(api huma.API) {
	tags := []string{"admin"}

	Register(api, huma.Operation{
		OperationID: "admin-tasks-list",
		Summary:     "List all tasks (admin)",
		Description: "Returns every task on the instance, across every project regardless of ownership. Restricted to instance admins; non-admin callers get a 404, making the endpoint indistinguishable from one that is not registered.",
		Method:      http.MethodGet,
		Path:        "/admin/tasks",
		Tags:        tags,
	}, adminTasksList)

	Register(api, huma.Operation{
		OperationID: "admin-tasks-completion-stats",
		Summary:     "Task completion stats by user (admin)",
		Description: "Returns, for every user credited with completing at least one task, their completed-task count and summed story points. Ordered by completed count descending.",
		Method:      http.MethodGet,
		Path:        "/admin/tasks/completion-stats",
		Tags:        tags,
	}, adminTasksCompletionStats)

	Register(api, huma.Operation{
		OperationID: "admin-tasks-patch-completed-by",
		Summary:     "Reassign a task's completed-by credit (admin)",
		Description: "Credits a different user with completing an already-done task, for cases where whoever marked it done wasn't the one who actually did the work. The task must currently be done. Restricted to instance admins.",
		Method:      http.MethodPatch,
		Path:        "/admin/tasks/{id}/completed-by",
		Tags:        tags,
	}, adminTasksPatchCompletedBy)
}

func init() { AddRouteRegistrar(RegisterAdminTaskRoutes) }

func adminTasksList(ctx context.Context, in *ListParams) (*adminTaskListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.AdminTaskList{}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.AdminTaskListEntry)
	if !ok {
		return nil, fmt.Errorf("AdminTaskList.ReadAll returned unexpected type %T (expected []*models.AdminTaskListEntry)", result)
	}
	return &adminTaskListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

func adminTasksCompletionStats(_ context.Context, _ *struct{}) (*adminTaskCompletionStatsBody, error) {
	s := db.NewSession()
	defer s.Close()

	stats, err := models.GetTaskCompletionStatsByUser(s)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &adminTaskCompletionStatsBody{Body: stats}, nil
}

func adminTasksPatchCompletedBy(ctx context.Context, in *struct {
	ID   int64 `path:"id" doc:"The numeric ID of the task."`
	Body adminCompletedByPatchBody
}) (*adminTaskBody, error) {
	if in.ID < 1 {
		return nil, translateDomainError(models.ErrTaskDoesNotExist{ID: in.ID})
	}
	if in.Body.CompletedByID < 1 {
		return nil, translateDomainError(models.ErrInvalidData{Message: "invalid body"})
	}

	doer, err := adminDoerFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	t, err := models.ReassignTaskCompletedBy(s, doer, in.ID, in.Body.CompletedByID)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		events.CleanupPending(s)
		return nil, translateDomainError(err)
	}
	events.DispatchPending(ctx, s)
	return &adminTaskBody{Body: t}, nil
}
