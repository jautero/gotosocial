/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package followrequest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// FollowRequestGETHandler allows clients to get a list of their incoming follow requests.
func (m *Module) FollowRequestGETHandler(c *gin.Context) {
	l := m.log.WithField("func", "statusCreatePOSTHandler")
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if authed.User.Disabled || !authed.User.Approved || !authed.Account.SuspendedAt.IsZero() {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled, not yet approved, or suspended"})
		return
	}

	accts, errWithCode := m.processor.FollowRequestsGet(c.Request.Context(), authed)
	if errWithCode != nil {
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	c.JSON(http.StatusOK, accts)
}
