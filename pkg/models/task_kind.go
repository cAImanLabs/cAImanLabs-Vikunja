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

	"github.com/danielgtaylor/huma/v2"
)

// TaskKind categorizes a task the way an issue tracker does (epic, story, bug, ...).
//
// int64, not int: the task filter query engine (getValueForField in
// task_collection_filter.go) switches on reflect.Kind() and only has a case
// for Int64 (matching Task.Priority), not Int - an int-kinded field there
// falls into the panic default case the first time someone filters on it.
type TaskKind int64

// NOTE: When adding or changing enum values for TaskKind, make sure to update
// the corresponding `enum` tag on Task.Kind to keep the OpenAPI documentation
// in sync. New values must be appended (never inserted) - the int64 is the
// value actually persisted in the `kind` column, so reordering would silently
// reinterpret every already-stored task's kind.
const (
	TaskKindTask TaskKind = iota
	TaskKindEpic
	TaskKindStory
	TaskKindBug
	TaskKindSubtask
	TaskKindFeature
	TaskKindInitiative
)

func (k *TaskKind) MarshalJSON() ([]byte, error) {
	switch *k {
	case TaskKindTask:
		return []byte(`"task"`), nil
	case TaskKindEpic:
		return []byte(`"epic"`), nil
	case TaskKindStory:
		return []byte(`"story"`), nil
	case TaskKindBug:
		return []byte(`"bug"`), nil
	case TaskKindSubtask:
		return []byte(`"subtask"`), nil
	case TaskKindFeature:
		return []byte(`"feature"`), nil
	case TaskKindInitiative:
		return []byte(`"initiative"`), nil
	}

	return []byte(`null`), nil
}

func (k *TaskKind) UnmarshalJSON(bytes []byte) error {
	var value string
	err := json.Unmarshal(bytes, &value)
	if err != nil {
		return err
	}

	switch value {
	case "task":
		*k = TaskKindTask
	case "epic":
		*k = TaskKindEpic
	case "story":
		*k = TaskKindStory
	case "bug":
		*k = TaskKindBug
	case "subtask":
		*k = TaskKindSubtask
	case "feature":
		*k = TaskKindFeature
	case "initiative":
		*k = TaskKindInitiative
	default:
		return fmt.Errorf("unknown task kind: %s", value)
	}

	return nil
}

// Schema lets Huma (/api/v2) reflect this type as a string enum; see the note
// on ProjectViewKind.Schema for why this is needed.
func (*TaskKind) Schema(_ huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: "string",
		Enum: []any{"task", "epic", "story", "bug", "subtask", "feature", "initiative"},
	}
}

// TaskKindLevel places a kind in the 5-level SAFe-style hierarchy this app
// enforces relations against (see updateSingleTask's kind-relation checks):
//
//	4 Initiative  - strategic, cross-team goals spanning quarters
//	3 Feature     - major deliverables under an initiative, spanning sprints
//	2 Epic        - a targeted set of work under a feature, one team's backlog
//	1 Story/Task/Bug - the day-to-day work done in a single sprint
//	0 Subtask     - granular steps to complete a single base issue
func (k *TaskKind) TaskKindLevel() int {
	switch *k {
	case TaskKindSubtask:
		return 0
	case TaskKindEpic:
		return 2
	case TaskKindFeature:
		return 3
	case TaskKindInitiative:
		return 4
	case TaskKindTask, TaskKindStory, TaskKindBug:
		return 1
	}
	return 1
}

// Prefix returns the short, uppercase prefix used to build this kind's
// per-kind ticket identifier (e.g. "ST" for a Story, so the 5th story
// created becomes "ST-05"). See Task.KindIndex/KindIdentifier.
func (k *TaskKind) Prefix() string {
	switch *k {
	case TaskKindEpic:
		return "EPIC"
	case TaskKindStory:
		return "ST"
	case TaskKindBug:
		return "BUG"
	case TaskKindSubtask:
		return "SUB"
	case TaskKindFeature:
		return "FEAT"
	case TaskKindInitiative:
		return "INIT"
	case TaskKindTask:
		return "TASK"
	}
	return "TASK"
}
