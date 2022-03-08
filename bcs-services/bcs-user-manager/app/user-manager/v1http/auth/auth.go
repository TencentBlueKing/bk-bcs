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
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/constant"
	jwt2 "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/jwt"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
	"github.com/emicklei/go-restful"
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
func (ta *TokenAuthenticater) GetUser() *models.BcsUser {
	tokenString := ta.ParseTokenString()
	if tokenString == "" {
		return nil
	}

	// get user from 32 bytes token
	if len(tokenString) == constant.DefaultTokenLength {
		return ta.GetUserFromToken(tokenString)
	}
	// get user from jwt
	return ta.GetJWTUser()
}

// GetUserFromToken returns a user object if the given token is valid
func (ta *TokenAuthenticater) GetUserFromToken(s string) *models.BcsUser {
	u := models.BcsUser{
		UserToken: s,
	}
	return sqlstore.GetUserByCondition(&u)
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

// GetJWTUser get specified user according jwt token
func (ta *TokenAuthenticater) GetJWTUser() *models.BcsUser {
	// resolve jwt token
	tokenString := ta.ParseTokenBearer()
	if tokenString == "" {
		return nil
	}

	jwtUser, err := jwt2.JWTClient.JWTDecode(tokenString)
	if err != nil {
		blog.Errorf("decode jwt user failed: %s", err.Error())
		return nil
	}

	// check expired time
	if time.Now().Unix() > jwtUser.ExpiresAt {
		return nil
	}

	// normal user or client
	var username string
	switch jwtUser.SubType {
	case jwt.User.String():
		if jwtUser.UserName == "" {
			blog.Errorf("invalid jwt user: %v", jwtUser)
			return nil
		}
		username = jwtUser.UserName
	case jwt.Client.String():
		if jwtUser.ClientID == "" {
			blog.Errorf("invalid jwt user: %v", jwtUser)
			return nil
		}
		username = jwtUser.ClientID
	default:
		blog.Errorf("invalid jwt user: %v", jwtUser)
		return nil
	}

	// get user from db
	u := models.BcsUser{
		Name: username,
	}
	user := sqlstore.GetUserByCondition(&u)
	// user is not exist in db, it means the jwt user is from browser client.
	// we need to create a new plain user.
	if user == nil {
		user = &models.BcsUser{
			Name:      username,
			UserType:  sqlstore.PlainUser,
			ExpiresAt: time.Unix(jwtUser.ExpiresAt, 0),
		}
	}
	return user
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

// TokenAuthFunc user token auth filter
func TokenAuthFunc(rb *restful.RouteBuilder) *restful.RouteBuilder {
	rb.Filter(TokenAuthAuthenticate)
	return rb
}

// AdminTokenAuthenticate admin token verification
func AdminTokenAuthenticate(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	authenticater := newTokenAuthenticater(request.Request, &TokenAuthConfig{
		SourceBearerEnabled: true,
	})
	user := authenticater.GetUser()
	if user != nil && !user.HasExpired() && user.UserType == sqlstore.AdminUser {
		request.SetAttribute(constant.CurrentUserAttr, user)
		chain.ProcessFilter(request, response)
		return
	}

	message := fmt.Sprintf("errcode: %d,  anonymous requests is forbidden, please provide a valid token", common.BcsErrApiUnauthorized)
	utils.WriteUnauthorizedError(response, common.BcsErrApiUnauthorized, message)
	return
}

// TokenAuthenticate uesr token verification
func TokenAuthenticate(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	authenticater := newTokenAuthenticater(request.Request, &TokenAuthConfig{
		SourceBearerEnabled: true,
	})
	user := authenticater.GetUser()
	if user != nil && !user.HasExpired() && (user.UserType == sqlstore.AdminUser || user.UserType == sqlstore.SaasUser) {
		chain.ProcessFilter(request, response)
		return
	}

	message := fmt.Sprintf("errcode: %d,  anonymous requests is forbidden, please provide a valid token", common.BcsErrApiUnauthorized)
	utils.WriteUnauthorizedError(response, common.BcsErrApiUnauthorized, message)
	return
}

// TokenAuthAuthenticate verify user token permission, when user is admin, it will bypass this filter.
// if user is not admin, it will check whether user has permission to access token service.
func TokenAuthAuthenticate(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	authenticater := newTokenAuthenticater(request.Request, &TokenAuthConfig{
		SourceBearerEnabled: true,
	})
	user := authenticater.GetUser()
	if user == nil || user.HasExpired() {
		message := fmt.Sprintf("errcode: %d,  anonymous requests is forbidden, please provide a valid token", common.BcsErrApiUnauthorized)
		utils.WriteUnauthorizedError(response, common.BcsErrApiUnauthorized, message)
		return
	}

	request.SetAttribute(constant.CurrentUserAttr, user)
	chain.ProcessFilter(request, response)
}

func newTokenAuthenticater(req *http.Request, config *TokenAuthConfig) *TokenAuthenticater {
	return &TokenAuthenticater{req: req, config: config}
}

// GetUser get CurrentUser from request object
func GetUser(req *restful.Request) *models.BcsUser {
	user := req.Attribute(constant.CurrentUserAttr)
	ret, ok := user.(*models.BcsUser)
	if ok {
		return ret
	}

	return nil
}
