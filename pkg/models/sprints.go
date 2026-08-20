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

package models

import (
	"encoding/json"
	"fmt"
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"

	"github.com/danielgtaylor/huma/v2"
	"xorm.io/xorm"
)

// SprintStatus is the status of a sprint in its lifecycle.
type SprintStatus int

// NOTE: When adding or changing enum values for SprintStatus,
// make sure to update the corresponding `enum` tag on Sprint.Status
// to keep the OpenAPI documentation in sync.
const (
	SprintStatusPlanning SprintStatus = iota
	SprintStatusActive
	SprintStatusCompleted
)

func (s *SprintStatus) MarshalJSON() ([]byte, error) {
	switch *s {
	case SprintStatusPlanning:
		return []byte(`"planning"`), nil
	case SprintStatusActive:
		return []byte(`"active"`), nil
	case SprintStatusCompleted:
		return []byte(`"completed"`), nil
	}

	return []byte(`null`), nil
}

func (s *SprintStatus) UnmarshalJSON(bytes []byte) error {
	var value string
	err := json.Unmarshal(bytes, &value)
	if err != nil {
		return err
	}

	switch value {
	case "planning":
		*s = SprintStatusPlanning
	case "active":
		*s = SprintStatusActive
	case "completed":
		*s = SprintStatusCompleted
	default:
		return fmt.Errorf("unknown sprint status: %s", value)
	}

	return nil
}

// Schema lets Huma (/api/v2) reflect this type as a string enum; see the note
// on ProjectViewKind.Schema for why this is needed.
func (*SprintStatus) Schema(_ huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: "string",
		Enum: []any{"planning", "active", "completed"},
	}
}

// Sprint represents an agile sprint iteration scoped to a project.
type Sprint struct {
	// The unique, numeric id of this sprint.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"sprint" readOnly:"true" doc:"The unique, numeric id of this sprint. Set by the server."`
	// The title of this sprint, e.g. "Sprint 12".
	Title string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250" doc:"The title of this sprint, e.g. \"Sprint 12\"."`
	// The goal or focus of this sprint.
	Goal string `xorm:"longtext null" json:"goal" doc:"The goal or focus of this sprint."`
	// When this sprint starts.
	StartDate time.Time `xorm:"DATETIME null 'start_date'" json:"start_date" doc:"When this sprint starts. Optional."`
	// When this sprint ends.
	EndDate time.Time `xorm:"DATETIME null 'end_date'" json:"end_date" doc:"When this sprint ends. Optional."`
	// The status of this sprint. Can be `planning`, `active` or `completed`.
	Status SprintStatus `xorm:"not null default 0" json:"status" swaggertype:"string" enums:"planning,active,completed" doc:"The status of this sprint. One of planning, active or completed."`
	// The sprint color in hex
	HexColor string `xorm:"varchar(6) null" json:"hex_color" valid:"runelength(0|7)" maxLength:"7" doc:"The sprint color as a hex string without the leading '#'."`
	// The project this sprint belongs to.
	ProjectID int64 `xorm:"bigint not null index" json:"project_id" param:"project" readOnly:"true" doc:"The project this sprint belongs to. Taken from the URL path; ignored on write."`

	CreatedByID int64      `xorm:"bigint not null" json:"-"`
	CreatedBy   *user.User `xorm:"-" json:"created_by" readOnly:"true" doc:"The user who created this sprint."`

	// A timestamp when this sprint was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created" readOnly:"true" doc:"A timestamp when this sprint was created. You cannot change this value."`
	// A timestamp when this sprint was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated" readOnly:"true" doc:"A timestamp when this sprint was last updated. You cannot change this value."`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (*Sprint) TableName() string {
	return "sprints"
}

func getSprintByID(s *xorm.Session, id int64) (sprint *Sprint, err error) {
	sprint = &Sprint{}
	exists, err := s.Where("id = ?", id).Get(sprint)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrSprintDoesNotExist{SprintID: id}
	}
	return sprint, nil
}

// ReadAll returns the sprints of a project. Requires read access to the project.
// There is intentionally no ReadOne (mirroring the Bucket model): the frontend
// only ever lists a project's sprints, it never fetches a single one by id.
func (sp *Sprint) ReadAll(s *xorm.Session, a web.Auth, _ string, _ int, _ int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	p := &Project{ID: sp.ProjectID}
	can, _, err := p.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !can {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	sprints := []*Sprint{}
	err = s.
		Where("project_id = ?", sp.ProjectID).
		OrderBy("start_date asc, id asc").
		Find(&sprints)
	if err != nil {
		return nil, 0, 0, err
	}

	userIDs := make([]int64, 0, len(sprints))
	for _, spr := range sprints {
		userIDs = append(userIDs, spr.CreatedByID)
	}
	users, err := getUsersOrLinkSharesFromIDs(s, userIDs)
	if err != nil {
		return nil, 0, 0, err
	}
	for _, spr := range sprints {
		if createdBy, has := users[spr.CreatedByID]; has {
			spr.CreatedBy = createdBy
		}
	}

	return sprints, len(sprints), int64(len(sprints)), nil
}

// Create creates a new sprint. Requires write access to the project.
func (sp *Sprint) Create(s *xorm.Session, a web.Auth) (err error) {
	sp.CreatedBy, err = GetUserOrLinkShareUser(s, a)
	if err != nil {
		return err
	}
	sp.CreatedByID = sp.CreatedBy.ID

	sp.HexColor = utils.NormalizeHex(sp.HexColor)
	sp.ID = 0
	_, err = s.Insert(sp)
	return err
}

// Update updates an existing sprint's title, goal, dates, status and color. Requires write access to the project.
func (sp *Sprint) Update(s *xorm.Session, _ web.Auth) (err error) {
	sp.HexColor = utils.NormalizeHex(sp.HexColor)
	_, err = s.
		Where("id = ?", sp.ID).
		Cols("title", "goal", "start_date", "end_date", "status", "hex_color").
		Update(sp)
	return
}

// Delete removes a sprint. Requires write access to the project. Tasks in the
// sprint are not deleted; they simply lose their sprint assignment.
func (sp *Sprint) Delete(s *xorm.Session, _ web.Auth) (err error) {
	_, err = s.Where("id = ?", sp.ID).Delete(&Sprint{})
	if err != nil {
		return err
	}
	_, err = s.Where("sprint_id = ?", sp.ID).Cols("sprint_id").Update(&Task{SprintID: 0})
	return err
}

// CanCreate checks if a user can create a new sprint in a project.
func (sp *Sprint) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	p := &Project{ID: sp.ProjectID}
	return p.CanUpdate(s, a)
}

// CanUpdate checks if a user can update an existing sprint.
func (sp *Sprint) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return sp.canDoSprint(s, a)
}

// CanDelete checks if a user can delete an existing sprint.
func (sp *Sprint) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return sp.canDoSprint(s, a)
}

// canDoSprint checks the sprint exists, belongs to the project given in the
// URL, and the user has write access to that project.
func (sp *Sprint) canDoSprint(s *xorm.Session, a web.Auth) (bool, error) {
	existing, err := getSprintByID(s, sp.ID)
	if err != nil {
		return false, err
	}
	if sp.ProjectID != 0 && sp.ProjectID != existing.ProjectID {
		return false, nil
	}

	p := &Project{ID: existing.ProjectID}
	return p.CanUpdate(s, a)
}
