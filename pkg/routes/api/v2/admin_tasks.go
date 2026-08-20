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

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
)

type adminTaskListBody struct {
	Body Paginated[*models.AdminTaskListEntry]
}

// Permissions are enforced by the gateV2AdminRoutes path middleware, not per-handler.
func RegisterAdminTaskRoutes(api huma.API) {
	Register(api, huma.Operation{
		OperationID: "admin-tasks-list",
		Summary:     "List all tasks (admin)",
		Description: "Returns every task on the instance, across every project regardless of ownership. Restricted to instance admins; non-admin callers get a 404, making the endpoint indistinguishable from one that is not registered.",
		Method:      http.MethodGet,
		Path:        "/admin/tasks",
		Tags:        []string{"admin"},
	}, adminTasksList)
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
