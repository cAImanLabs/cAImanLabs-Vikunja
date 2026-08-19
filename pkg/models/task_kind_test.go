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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskKind_MarshalUnmarshalJSON(t *testing.T) {
	cases := []struct {
		kind TaskKind
		json string
	}{
		{TaskKindTask, `"task"`},
		{TaskKindEpic, `"epic"`},
		{TaskKindStory, `"story"`},
		{TaskKindBug, `"bug"`},
		{TaskKindSubtask, `"subtask"`},
		{TaskKindFeature, `"feature"`},
		{TaskKindInitiative, `"initiative"`},
	}

	for _, c := range cases {
		t.Run(c.json, func(t *testing.T) {
			b, err := json.Marshal(&c.kind)
			require.NoError(t, err)
			assert.Equal(t, c.json, string(b))

			var got TaskKind
			err = json.Unmarshal([]byte(c.json), &got)
			require.NoError(t, err)
			assert.Equal(t, c.kind, got)
		})
	}

	t.Run("unknown value", func(t *testing.T) {
		var got TaskKind
		err := json.Unmarshal([]byte(`"not-a-kind"`), &got)
		require.Error(t, err)
	})
}

func TestTaskKind_TaskKindLevel(t *testing.T) {
	cases := []struct {
		kind  TaskKind
		level int
	}{
		{TaskKindSubtask, 0},
		{TaskKindTask, 1},
		{TaskKindStory, 1},
		{TaskKindBug, 1},
		{TaskKindEpic, 2},
		{TaskKindFeature, 3},
		{TaskKindInitiative, 4},
	}

	for _, c := range cases {
		assert.Equal(t, c.level, c.kind.TaskKindLevel(), "kind %v", c.kind)
	}
}
