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
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// Partial struct: tasks already exists, so this must go through partialSync
// (see migration.go) rather than tx.Sync2, which would drop every index on
// tasks the struct doesn't declare.
type taskCompletedByField20260821010743 struct {
	ID            int64 `xorm:"bigint autoincr not null unique pk"`
	Done          bool  `xorm:"INDEX null"`
	CreatedByID   int64 `xorm:"bigint not null"`
	CompletedByID int64 `xorm:"bigint INDEX null 'completed_by_id'"`
}

func (taskCompletedByField20260821010743) TableName() string {
	return "tasks"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260821010743",
		Description: "Add completed_by_id to tasks, tracking who actually completed the work (may differ from who marked it done)",
		Migrate:     addTaskCompletedByField20260821010743,
		Rollback: func(_ *xorm.Engine) error {
			// Not reversible: partialSync only ever adds columns, never drops them.
			return nil
		},
	})
}

func addTaskCompletedByField20260821010743(tx *xorm.Engine) error {
	if err := partialSync(tx, taskCompletedByField20260821010743{}); err != nil {
		return err
	}

	s := tx.NewSession()
	defer s.Close()

	if err := s.Begin(); err != nil {
		return err
	}

	// Backfill already-done tasks with their creator as a reasonable default -
	// there's no record of who actually did the work before this column
	// existed. An admin can correct individual tasks via the reassignment
	// endpoint once real attribution matters.
	// completed_by_id is a NEW nullable column: ADD COLUMN leaves it actual SQL
	// NULL on every pre-existing row, not 0, so both must be matched here.
	var tasks []*taskCompletedByField20260821010743
	if err := s.Where("done = ? AND (completed_by_id = 0 OR completed_by_id IS NULL)", true).Find(&tasks); err != nil {
		_ = s.Rollback()
		return err
	}

	for _, t := range tasks {
		t.CompletedByID = t.CreatedByID
		if _, err := s.ID(t.ID).Cols("completed_by_id").Update(t); err != nil {
			_ = s.Rollback()
			return err
		}
	}

	return s.Commit()
}
