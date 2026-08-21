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

package admin

import (
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
)

type CompletedByPatch struct {
	CompletedByID int64 `json:"completed_by_id"`
}

// PatchTaskCompletedBy credits a different user with completing a task.
// @Summary Reassign a task's completed-by credit (admin)
// @Description Credits a different user with completing an already-done task, for cases where whoever marked it done wasn't the one who actually did the work.
// @tags admin
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Task ID"
// @Param body body admin.CompletedByPatch true "The user to credit"
// @Success 200 {object} models.Task
// @Failure 400 {object} web.HTTPError
// @Failure 404 {object} web.HTTPError
// @Router /admin/tasks/{id}/completed-by [patch]
func PatchTaskCompletedBy(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return models.ErrTaskDoesNotExist{ID: id}
	}

	body := &CompletedByPatch{}
	if err := c.Bind(body); err != nil || body.CompletedByID < 1 {
		return models.ErrInvalidData{Message: "invalid body"}
	}

	doer, err := user.GetCurrentUser(c)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	t, err := models.ReassignTaskCompletedBy(s, doer, id, body.CompletedByID)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return err
	}
	if err := s.Commit(); err != nil {
		events.CleanupPending(s)
		return err
	}
	events.DispatchPending(c.Request().Context(), s)
	return c.JSON(http.StatusOK, t)
}
