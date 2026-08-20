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
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// isInstanceAdmin gates cross-user access on is_admin. The admin panel is a
// free feature in this fork, so there's no license check here - is_admin is
// re-read from the DB because the auth's flag is claim-derived and stale
// until the JWT expires.
func isInstanceAdmin(s *xorm.Session, a web.Auth) bool {
	u, ok := a.(*user.User)
	if !ok {
		return false
	}
	fresh, err := user.GetUserByID(s, u.ID)
	if err != nil {
		return false
	}
	return fresh.IsAdmin
}
