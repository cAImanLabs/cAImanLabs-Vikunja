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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v5"
)

// GetTaskCompletionStats returns per-user task-completion totals for the
// admin overview's completion charts.
// @Summary Task completion stats by user (admin)
// @Description Returns, for every user credited with completing at least one task, their completed-task count and summed story points.
// @tags admin
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {array} models.UserTaskCompletionStat
// @Failure 404 {object} web.HTTPError
// @Router /admin/tasks/completion-stats [get]
func GetTaskCompletionStats(c *echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	stats, err := models.GetTaskCompletionStatsByUser(s)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, stats)
}
