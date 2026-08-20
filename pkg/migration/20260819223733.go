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
	"xorm.io/xorm/schemas"
)

// Partial struct: tasks already exists, so this must go through partialSync
// (see migration.go) rather than tx.Sync2, which would drop every index on
// tasks the struct doesn't declare.
type taskKindIndexField20260819223733 struct {
	ID        int64 `xorm:"bigint autoincr not null unique pk"`
	Kind      int64 `xorm:"not null default 0 'kind'"`
	KindIndex int64 `xorm:"bigint not null default 0 'kind_index'"`
}

func (taskKindIndexField20260819223733) TableName() string {
	return "tasks"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260819223733",
		Description: "Add a per-kind sequential ticket number to tasks (e.g. ST-01 for stories)",
		Migrate:     addTaskKindIndex20260819223733,
		Rollback: func(_ *xorm.Engine) error {
			// Not reversible: partialSync only ever adds columns, never drops
			// them, and by the time a rollback runs kind_index may already be
			// referenced by ticket numbers users have started relying on.
			return nil
		},
	})
}

func addTaskKindIndex20260819223733(tx *xorm.Engine) error {
	if err := partialSync(tx, taskKindIndexField20260819223733{}); err != nil {
		return err
	}

	s := tx.NewSession()
	defer s.Close()

	if err := s.Begin(); err != nil {
		return err
	}

	// Backfill every existing task with a sequential kind_index, counted per
	// kind and ordered by id so numbering is deterministic across databases.
	// All rows currently share kind_index 0, so this must run before the
	// unique index below can be created.
	tasks := []*taskKindIndexField20260819223733{}
	if err := s.OrderBy("kind ASC, id ASC").Find(&tasks); err != nil {
		_ = s.Rollback()
		return err
	}

	counters := map[int64]int64{}
	for _, t := range tasks {
		counters[t.Kind]++
		t.KindIndex = counters[t.Kind]
		if _, err := s.ID(t.ID).Cols("kind_index").Update(t); err != nil {
			_ = s.Rollback()
			return err
		}
	}

	if err := s.Commit(); err != nil {
		return err
	}

	var query string
	switch tx.Dialect().URI().DBType {
	case schemas.MYSQL:
		query = "CREATE UNIQUE INDEX UQE_tasks_kind_kind_index ON tasks (kind, kind_index)"
	default:
		query = "CREATE UNIQUE INDEX IF NOT EXISTS UQE_tasks_kind_kind_index ON tasks (kind, kind_index)"
	}
	_, err := tx.Exec(query)
	return err
}
