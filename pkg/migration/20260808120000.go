// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// Partial struct: users already exists, so this must go through partialSync
// (see migration.go) rather than tx.Sync2, which would drop every index on
// users the struct doesn't declare.
type userColorField20260808120000 struct {
	ID    int64  `xorm:"bigint autoincr not null unique pk"`
	Color string `xorm:"varchar(7) null 'color'"`
}

func (userColorField20260808120000) TableName() string {
	return "users"
}

// Partial struct: tasks already exists, same partialSync reasoning as above.
type taskKindField20260808120000 struct {
	ID int64 `xorm:"bigint autoincr not null unique pk"`
	// Mirrors models.TaskKind (0=task, 1=epic, 2=story, 3=bug, 4=subtask, 5=feature).
	Kind int64 `xorm:"not null default 0 'kind'"`
}

func (taskKindField20260808120000) TableName() string {
	return "tasks"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260808120000",
		Description: "Add color field to users and kind field to tasks",
		Migrate: func(tx *xorm.Engine) error {
			return partialSync(tx, userColorField20260808120000{}, taskKindField20260808120000{})
		},
		Rollback: func(tx *xorm.Engine) error {
			// Not reversible: partialSync only ever adds columns, never drops them,
			// and by the time a rollback runs these columns may hold real user data
			// (colors, task kinds) that a DROP COLUMN would destroy.
			return nil
		},
	})
}
