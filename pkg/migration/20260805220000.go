// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// Partial struct: tasks already exists, so this must go through partialSync
// (see migration.go) rather than tx.Sync2, which would drop every index on
// tasks the struct doesn't declare.
type taskSprintFields20260805220000 struct {
	ID          int64 `xorm:"bigint autoincr not null unique pk"`
	StoryPoints int   `xorm:"int null 'story_points'"`
	SprintID    int64 `xorm:"bigint INDEX null 'sprint_id'"`
}

func (taskSprintFields20260805220000) TableName() string {
	return "tasks"
}

type sprint20260805220000 struct {
	ID        int64     `xorm:"bigint autoincr not null unique pk"`
	Title     string    `xorm:"varchar(250) not null"`
	Goal      string    `xorm:"longtext null"`
	StartDate time.Time `xorm:"DATETIME null 'start_date'"`
	EndDate   time.Time `xorm:"DATETIME null 'end_date'"`
	// Mirrors models.SprintStatus (0=planning, 1=active, 2=completed).
	Status      int   `xorm:"not null default 0"`
	ProjectID   int64 `xorm:"bigint not null INDEX"`
	CreatedByID int64 `xorm:"bigint not null"`

	Created time.Time `xorm:"created not null"`
	Updated time.Time `xorm:"updated not null"`
}

func (sprint20260805220000) TableName() string {
	return "sprints"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260805220000",
		Description: "Add story points/sprint fields to tasks and create the sprints table",
		Migrate: func(tx *xorm.Engine) error {
			return partialSync(tx, taskSprintFields20260805220000{}, sprint20260805220000{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(new(sprint20260805220000))
		},
	})
}
