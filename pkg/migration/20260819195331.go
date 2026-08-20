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

// Partial struct: sprints already exists, so this must go through partialSync
// (see migration.go) rather than tx.Sync2, which would drop every index on
// sprints the struct doesn't declare.
type sprintHexColorField20260819195331 struct {
	ID       int64  `xorm:"bigint autoincr not null unique pk"`
	HexColor string `xorm:"varchar(6) null 'hex_color'"`
}

func (sprintHexColorField20260819195331) TableName() string {
	return "sprints"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260819195331",
		Description: "Add hex_color field to sprints",
		Migrate: func(tx *xorm.Engine) error {
			return partialSync(tx, sprintHexColorField20260819195331{})
		},
		Rollback: func(tx *xorm.Engine) error {
			// Not reversible: partialSync only ever adds columns, never drops them,
			// and by the time a rollback runs this column may hold real user data
			// that a DROP COLUMN would destroy.
			return nil
		},
	})
}
