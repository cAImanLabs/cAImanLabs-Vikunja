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
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
)

type sprintListBody struct {
	Body Paginated[*models.Sprint]
}

type sprintVelocityBody struct {
	Body []*models.SprintVelocityPoint
}

// RegisterSprintRoutes wires project-scoped sprint CRUD onto the Huma API.
// Sprints live under /projects/{project}/sprints; every operation binds
// {project} -> ProjectID and the write operations additionally {sprint} -> ID.
// There is intentionally no read-one route (mirroring buckets.go: the Sprint
// model has no ReadOne/CanRead), so AutoPatch synthesises no PATCH either.
func RegisterSprintRoutes(api huma.API) {
	tags := []string{"projects"}

	Register(api, huma.Operation{
		OperationID: "sprints-list",
		Summary:     "List the sprints of a project",
		Description: "Returns all sprints of a project, ordered by start date. Requires read access to the project.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/sprints",
		Tags:        tags,
	}, sprintsList)

	Register(api, huma.Operation{
		OperationID: "sprints-create",
		Summary:     "Create a sprint in a project",
		Description: "Creates a sprint in the given project. The project comes from the URL, not the body. Requires write access to the project.",
		Method:      http.MethodPost,
		Path:        "/projects/{project}/sprints",
		Tags:        tags,
	}, sprintsCreate)

	Register(api, huma.Operation{
		OperationID: "sprints-update",
		Summary:     "Update a sprint",
		Description: "Replaces a sprint's title, goal, dates and status. The sprint is identified by the URL, which also scopes it to the project. Requires write access to the project.",
		Method:      http.MethodPut,
		Path:        "/projects/{project}/sprints/{sprint}",
		Tags:        tags,
	}, sprintsUpdate)

	Register(api, huma.Operation{
		OperationID: "sprints-delete",
		Summary:     "Delete a sprint",
		Description: "Deletes a sprint. Tasks assigned to it are not deleted, they just lose their sprint assignment. Requires write access to the project.",
		Method:      http.MethodDelete,
		Path:        "/projects/{project}/sprints/{sprint}",
		Tags:        tags,
	}, sprintsDelete)

	Register(api, huma.Operation{
		OperationID: "sprints-velocity",
		Summary:     "Get a project's sprint velocity",
		Description: "Returns the total and completed story points for every sprint in the project, ordered by start date, for plotting a velocity chart across sprints. Requires read access to the project.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/sprints/velocity",
		Tags:        tags,
	}, sprintsVelocity)

	Register(api, huma.Operation{
		OperationID: "sprints-burndown",
		Summary:     "Get a sprint's burndown chart data",
		Description: "Returns one remaining-story-points data point per day from the sprint's start date through today (or its end date, whichever is earlier), plus the total points needed to draw the ideal line. Requires read access to the project. Returns 412 if the sprint has no start or end date set.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/sprints/{sprint}/burndown",
		Tags:        tags,
	}, sprintsBurndown)
}

func init() { AddRouteRegistrar(RegisterSprintRoutes) }

func sprintsList(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ListParams
}) (*sprintListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.Sprint{ProjectID: in.ProjectID}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	sprints, ok := result.([]*models.Sprint)
	if !ok {
		return nil, fmt.Errorf("sprints.ReadAll returned unexpected type %T (expected []*models.Sprint)", result)
	}
	return &sprintListBody{Body: NewPaginated(sprints, total, in.Page, in.PerPage)}, nil
}

func sprintsCreate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	Body      models.Sprint
}) (*singleBody[models.Sprint], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	sp := &in.Body
	sp.ProjectID = in.ProjectID // URL wins over body
	if err := handler.DoCreate(ctx, sp, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.Sprint]{Body: sp}, nil
}

func sprintsUpdate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	SprintID  int64 `path:"sprint"`
	Body      models.Sprint
}) (*singleBody[models.Sprint], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	sp := &in.Body
	sp.ID = in.SprintID         // URL wins over body
	sp.ProjectID = in.ProjectID // URL wins over body
	if err := handler.DoUpdate(ctx, sp, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.Sprint]{Body: sp}, nil
}

func sprintsDelete(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	SprintID  int64 `path:"sprint"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.Sprint{ID: in.SprintID, ProjectID: in.ProjectID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

// sprintsVelocity and sprintsBurndown are read-only aggregates, not plain CRUD,
// so there's no handler.Do* to lean on: permission checks and session
// lifecycle are this handler's own responsibility (see the api-v2-routes skill).
func sprintsVelocity(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
}) (*sprintVelocityBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	p := &models.Project{ID: in.ProjectID}
	can, _, err := p.CanRead(s, a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !can {
		_ = s.Rollback()
		return nil, huma.Error403Forbidden("forbidden")
	}

	velocity, err := models.GetProjectVelocity(s, in.ProjectID)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &sprintVelocityBody{Body: velocity}, nil
}

func sprintsBurndown(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	SprintID  int64 `path:"sprint"`
}) (*singleBody[models.SprintBurndown], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	p := &models.Project{ID: in.ProjectID}
	can, _, err := p.CanRead(s, a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !can {
		_ = s.Rollback()
		return nil, huma.Error403Forbidden("forbidden")
	}

	burndown, err := models.GetSprintBurndown(s, in.ProjectID, in.SprintID)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &singleBody[models.SprintBurndown]{Body: burndown}, nil
}
