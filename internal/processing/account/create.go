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

package account

import (
	"context"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/oauth2/v4"
)

func (p *processor) Create(ctx context.Context, applicationToken oauth2.TokenInfo, application *gtsmodel.Application, form *apimodel.AccountCreateRequest) (*apimodel.Token, error) {
	l := p.log.WithField("func", "accountCreate")

	emailAvailable, err := p.db.IsEmailAvailable(ctx, form.Email)
	if err != nil {
		return nil, err
	}
	if !emailAvailable {
		return nil, fmt.Errorf("email address %s in use", form.Email)
	}

	usernameAvailable, err := p.db.IsUsernameAvailable(ctx, form.Username)
	if err != nil {
		return nil, err
	}
	if !usernameAvailable {
		return nil, fmt.Errorf("username %s in use", form.Username)
	}

	// don't store a reason if we don't require one
	reason := form.Reason
	if !p.config.AccountsConfig.ReasonRequired {
		reason = ""
	}

	l.Trace("creating new username and account")
	user, err := p.db.NewSignup(ctx, form.Username, text.RemoveHTML(reason), p.config.AccountsConfig.RequireApproval, form.Email, form.Password, form.IP, form.Locale, application.ID, false, false)
	if err != nil {
		return nil, fmt.Errorf("error creating new signup in the database: %s", err)
	}

	l.Tracef("generating a token for user %s with account %s and application %s", user.ID, user.AccountID, application.ID)
	accessToken, err := p.oauthServer.GenerateUserAccessToken(applicationToken, application.ClientSecret, user.ID)
	if err != nil {
		return nil, fmt.Errorf("error creating new access token for user %s: %s", user.ID, err)
	}

	return &apimodel.Token{
		AccessToken: accessToken.GetAccess(),
		TokenType:   "Bearer",
		Scope:       accessToken.GetScope(),
		CreatedAt:   accessToken.GetAccessCreateAt().Unix(),
	}, nil
}
