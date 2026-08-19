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

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// Copy of models.ProjectViewKindSprint, frozen here so this migration doesn't
// move if the enum ever changes.
const projectViewKindSprint20260806190840 = 4

type project20260806190840 struct {
	ID int64 `xorm:"autoincr not null unique pk"`
}

func (project20260806190840) TableName() string {
	return "projects"
}

type projectView20260806190840 struct {
	ID        int64     `xorm:"autoincr not null unique pk"`
	Title     string    `xorm:"varchar(255) not null"`
	ProjectID int64     `xorm:"not null index"`
	ViewKind  int       `xorm:"not null"`
	Position  float64   `xorm:"double null"`
	Created   time.Time `xorm:"created not null"`
	Updated   time.Time `xorm:"updated not null"`
}

func (projectView20260806190840) TableName() string {
	return "project_views"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260806190840",
		Description: "Backfill a Sprints view onto projects created before it became a default view",
		Migrate:     backfillSprintViews20260806190840,
		Rollback: func(tx *xorm.Engine) error {
			// Not reversible: by the time a rollback runs, some of these
			// sprint views may be real, user-created data (their own
			// sprints attached), not just the backfilled empty ones.
			return nil
		},
	})
}

func backfillSprintViews20260806190840(tx *xorm.Engine) error {
	projects := []*project20260806190840{}
	if err := tx.Find(&projects); err != nil {
		return err
	}

	existing := []*projectView20260806190840{}
	if err := tx.Where("view_kind = ?", projectViewKindSprint20260806190840).Find(&existing); err != nil {
		return err
	}
	hasSprintView := make(map[int64]bool, len(existing))
	for _, v := range existing {
		hasSprintView[v.ProjectID] = true
	}

	for _, p := range projects {
		if hasSprintView[p.ID] {
			continue
		}
		view := &projectView20260806190840{
			Title:     "Sprints",
			ProjectID: p.ID,
			ViewKind:  projectViewKindSprint20260806190840,
			Position:  500,
		}
		if _, err := tx.Insert(view); err != nil {
			return err
		}
	}

	return nil
}
