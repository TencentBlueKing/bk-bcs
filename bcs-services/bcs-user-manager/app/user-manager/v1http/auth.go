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

package v1http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
	"github.com/emicklei/go-restful"
)

const (
	//CurrentUserAttr user header
	CurrentUserAttr = "current-user"
)

// TokenAuthenticater wrapper for http request
type TokenAuthenticater struct {
	req    *http.Request
	config *TokenAuthConfig
}

// TokenAuthConfig configuration
type TokenAuthConfig struct {
	SourceBearerEnabled    bool
	sourceBasicAuthEnabled bool
	// Only token in this type will be considered as valid
	ValidTokenType uint
}

// GetUser get specified user according token
func (ta *TokenAuthenticater) GetUser() (*models.BcsUser, bool) {
	tokenString := ta.ParseTokenString()
	if tokenString == "" {
		return nil, false
	}

	user, hasExpired := ta.GetUserFromToken(tokenString)
	if user == nil {
		return user, hasExpired
	} else if hasExpired {
		blog.Warnf("usertoken has been expired: %s", tokenString)
		return user, hasExpired
	} else {
		return user, false
	}
}

// GetUserFromToken returns a user object if the given token is valid
func (ta *TokenAuthenticater) GetUserFromToken(s string) (*models.BcsUser, bool) {
	u := models.BcsUser{
		UserToken: s,
	}
	user := sqlstore.GetUserByCondition(&u)

	if user == nil {
		return nil, false
	}

	if user.HasExpired() {
		return user, true
	}

	return user, false
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

// ParseTokenBearer extra token information from http authorization header
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

// ParseTokenBasicAuth extra password information from http header
func (ta *TokenAuthenticater) ParseTokenBasicAuth() string {
	_, password, ok := ta.req.BasicAuth()
	if ok && password != "" {
		return password
	}
	return ""
}

// AdminAuthFunc auth filter
func AdminAuthFunc(rb *restful.RouteBuilder) *restful.RouteBuilder {
	rb.Filter(AdminTokenAuthenticate)
	return rb
}

// AuthFunc token filter
func AuthFunc(rb *restful.RouteBuilder) *restful.RouteBuilder {
	rb.Filter(TokenAuthenticate)
	return rb
}

// AdminTokenAuthenticate admin token verification
func AdminTokenAuthenticate(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	authenticater := newTokenAuthenticater(request.Request, &TokenAuthConfig{
		SourceBearerEnabled: true,
	})
	user, hasExpired := authenticater.GetUser()
	if user != nil && !hasExpired && user.UserType == sqlstore.AdminUser {
		request.SetAttribute(CurrentUserAttr, user)
		chain.ProcessFilter(request, response)
		return
	}

	message := fmt.Sprintf("errcode：%d,  anonymous requests is forbidden, please provide a valid token", common.BcsErrApiUnauthorized)
	utils.WriteUnauthorizedError(response, common.BcsErrApiUnauthorized, message)
	return
}

// TokenAuthenticate uesr token verification
func TokenAuthenticate(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	authenticater := newTokenAuthenticater(request.Request, &TokenAuthConfig{
		SourceBearerEnabled: true,
	})
	user, hasExpired := authenticater.GetUser()
	if user != nil && !hasExpired && (user.UserType == sqlstore.AdminUser || user.UserType == sqlstore.SaasUser) {
		chain.ProcessFilter(request, response)
		return
	}

	message := fmt.Sprintf("errcode：%d,  anonymous requests is forbidden, please provide a valid token", common.BcsErrApiUnauthorized)
	utils.WriteUnauthorizedError(response, common.BcsErrApiUnauthorized, message)
	return
}

func newTokenAuthenticater(req *http.Request, config *TokenAuthConfig) *TokenAuthenticater {
	return &TokenAuthenticater{req: req, config: config}
}

// GetUser get CurrentUser from request object
func GetUser(req *restful.Request) *models.BcsUser {
	user := req.Attribute(CurrentUserAttr)
	ret, ok := user.(*models.BcsUser)
	if ok {
		return ret
	}

	return nil
}
