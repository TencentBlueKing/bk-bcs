/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	m "github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"
)

// TokenAuthenticater token auth implementation
type TokenAuthenticater struct {
	req    *http.Request
	config *TokenAuthConfig
}

// TokenAuthConfig configuration for bcs-api auth
type TokenAuthConfig struct {
	SourceBearerEnabled    bool
	sourceBasicAuthEnabled bool
	// Only token in this type will be considered as valid
	ValidTokenType uint
}

// DefaultTokenAuthConfig default configuration
var DefaultTokenAuthConfig = &TokenAuthConfig{
	SourceBearerEnabled:    true,
	sourceBasicAuthEnabled: true,
	ValidTokenType:         m.UserTokenTypeSession,
}

// NewTokenAuthenticater create auth implementation from request
func NewTokenAuthenticater(req *http.Request, config *TokenAuthConfig) *TokenAuthenticater {
	return &TokenAuthenticater{req: req, config: config}
}

// ParseTokenString parses token string from incoming request, currently supports authorization header and basicauth
func (ta *TokenAuthenticater) ParseTokenString() string {
	var token string
	if ta.config.SourceBearerEnabled {
		token = ta.ParseTokenBearer()
	}
	if token == "" && ta.config.sourceBasicAuthEnabled {
		token = ta.ParseTokenBasicAuth()
	}
	return token
}

// ParseTokenBearer parse toke bearer information
func (ta *TokenAuthenticater) ParseTokenBearer() string {
	authHeaderList := ta.req.Header["Authorization"]

	if len(authHeaderList) > 0 {
		authHeader := strings.Split(authHeaderList[0], " ")
		if len(authHeader) == 2 && authHeader[0] == "Bearer" {
			return strings.TrimSpace(authHeader[1])
		}
	}
	return ""
}

// ParseTokenBasicAuth parse token basic auth information
func (ta *TokenAuthenticater) ParseTokenBasicAuth() string {
	_, password, ok := ta.req.BasicAuth()
	if ok && password != "" {
		return password
	}
	return ""
}

// GetUserFromToken returns a user object if the given token is valid
func (ta *TokenAuthenticater) GetUserFromToken(s string) (*m.User, bool) {
	token := sqlstore.GetUserToken(s)

	if token == nil {
		return nil, false
	}

	if token.HasExpired() {
		return sqlstore.GetUser(token.UserId), true
	}

	return sqlstore.GetUser(token.UserId), false
}

// GetUser get user information from specified token
func (ta *TokenAuthenticater) GetUser() (*m.User, bool) {
	tokenString := ta.ParseTokenString()
	blog.Debug(fmt.Sprintf("User token found in request: %s", tokenString))

	user, hasExpired := ta.GetUserFromToken(tokenString)
	if user == nil {
		blog.Warnf("No user can be found by token:%s", tokenString)
		return user, hasExpired
	} else if hasExpired {
		return user, hasExpired
	} else {
		blog.Debug(fmt.Sprintf("User:%s found by token:%s", user.Name, tokenString))
		return user, false
	}
}

// GetUserTokenType check user token type
func (ta *TokenAuthenticater) GetUserTokenType() uint {
	tokenString := ta.ParseTokenString()
	token := sqlstore.GetUserToken(tokenString)
	return token.Type

}
