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
 */

// Package token xxx
package token

import (
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	"github.com/dchest/uniuri"
	restful "github.com/emicklei/go-restful/v3"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/errors"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
)

const (
	// NeverExpiredDuration is a very large number that represents a time duration,
	// when user create a never expired token, the token expiration will be 10 years.
	NeverExpiredDuration = time.Hour * 24 * 365 * 10
)

// NeverExpired mean when user's expiration time less than NeverExpired time,
// then the token will never be expired.
var NeverExpired, _ = time.Parse(time.RFC3339, "2032-01-19T03:14:07Z")

// TokenHandler is a restful handler for token.
// nolint
type TokenHandler struct {
	tokenStore sqlstore.TokenStore
	cache      cache.Cache
	jwtClient  jwt.BCSJWTAuthentication
}

// NewTokenHandler is a constructor for TokenHandler.
func NewTokenHandler(tokenStore sqlstore.TokenStore, cache cache.Cache,
	jwtClient jwt.BCSJWTAuthentication) *TokenHandler {
	return &TokenHandler{
		tokenStore: tokenStore,
		cache:      cache,
		jwtClient:  jwtClient,
	}
}

// CreateProjectClientForm is the form of creating project client
type CreateProjectClientForm struct {
	ClientName string `json:"clientName" validate:"required"`
	// Expiration token expiration second, -1: never expire
	Expiration int `json:"expiration" validate:"required"`
}

// CreateProjectClient 创建项目下平台账号
func (t *TokenHandler) CreateProjectClient(request *restful.Request, response *restful.Response) {
	project := utils.GetProjectFromAttribute(request)
	if project == nil {
		utils.ResponseParamsError(response, errors.ErrProjectNotFound)
		return
	}
	form := CreateProjectClientForm{}
	_ = request.ReadEntity(&form)
	err := utils.Validate.Struct(&form)
	if err != nil {
		utils.ResponseParamsError(response, err)
		return
	}

	// check if client name already exist
	exist := t.tokenStore.GetTokenByCondition(&models.BcsUser{Name: form.ClientName, UserType: models.PlainUser})
	if exist != nil {
		utils.ResponseParamsError(response, fmt.Errorf("client %s already exist", form.ClientName))
		return
	}

	// transfer never expire to expiration
	if form.Expiration < 0 {
		form.Expiration = int(NeverExpiredDuration.Seconds())
	}
	// create jwt token
	token := uniuri.NewLen(constant.DefaultTokenLength)
	key := constant.TokenKeyPrefix + token
	expiredAt := time.Now().Add(time.Duration(form.Expiration) * time.Second)
	jwtString, err := t.jwtClient.JWTSign(&jwt.UserInfo{
		SubType:     jwt.User.String(),
		UserName:    form.ClientName,
		ExpiredTime: int64(form.Expiration),
		Issuer:      jwt.JWTIssuer,
	})
	if err != nil {
		utils.ResponseSystemError(response, fmt.Errorf("create jwt token failed, %s", err.Error()))
		return
	}
	_, err = t.cache.Set(request.Request.Context(), key, jwtString, time.Duration(form.Expiration)*time.Second)
	if err != nil {
		blog.Errorf("set client %s token fail", form.ClientName)
		utils.ResponseSystemError(response, fmt.Errorf("set client %s token fail", form.ClientName))
		return
	}

	user := utils.GetUserFromAttribute(request)
	if user == nil {
		utils.ResponseAuthError(response)
		return
	}

	// insert token record in db
	userToken := &models.BcsClientUser{
		ProjectCode: project.ProjectCode,
		Name:        form.ClientName,
		UserToken:   token,
		UserType:    models.PlainUser,
		CreatedBy:   user.Name,
		ExpiresAt:   expiredAt,
	}
	err = t.tokenStore.CreateClientToken(userToken)
	if err != nil {
		blog.Errorf("create client %s token record failed, %s", form.ClientName, err.Error())
		utils.ResponseDBError(response, err)
		return
	}

	utils.ResponseOK(response, nil)
}

// GetProjectClientsResp is the response of get project clients
type GetProjectClientsResp struct {
	ProjectCode   string            `json:"project_code"`
	Name          string            `json:"name"`
	Token         string            `json:"token"`
	Status        utils.TokenStatus `json:"status"`
	CreatedBy     string            `json:"created_by"`
	Manager       []string          `json:"manager"`
	AuthorityUser []string          `json:"authority_user"`
	ExpiredAt     *time.Time        `json:"expired_at"` // nil means never expired
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// GetProjectClients 获取项目下平台账号
func (t *TokenHandler) GetProjectClients(request *restful.Request, response *restful.Response) {
	project := utils.GetProjectFromAttribute(request)
	if project == nil {
		utils.ResponseParamsError(response, errors.ErrProjectNotFound)
		return
	}

	user := utils.GetUserFromAttribute(request)
	if user == nil {
		utils.ResponseAuthError(response)
		return
	}

	tokensInDB := t.tokenStore.GetProjectClients(project.ProjectCode)
	clients := make([]GetProjectClientsResp, 0)
	for _, v := range tokensInDB {
		status := utils.TokenStatusActive
		if v.HasExpired() {
			status = utils.TokenStatusExpired
		}
		expiresAt := &v.ExpiresAt
		// transfer never expired
		if v.ExpiresAt.After(NeverExpired) {
			expiresAt = nil
		}
		token := v.UserToken
		// 无权限不能查看 token
		if !user.IsAdmin() && user.Name != v.CreatedBy && !v.IsManager(user.Name) {
			token = ""
		}
		clients = append(clients, GetProjectClientsResp{
			ProjectCode: project.ProjectCode, Name: v.Name, Token: token, Status: status, CreatedBy: v.CreatedBy,
			Manager: v.ManagerList(), AuthorityUser: v.AuthorityUserList(),
			ExpiredAt: expiresAt, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt})
	}
	utils.ResponseOK(response, clients)
}

// UpdateProjectClientForm is the form of updating project client
type UpdateProjectClientForm struct {
	Manager []string `json:"manager" validate:"required,gte=1"`
}

// UpdateProjectClient 更新项目下平台账号
func (t *TokenHandler) UpdateProjectClient(request *restful.Request, response *restful.Response) {
	project := utils.GetProjectFromAttribute(request)
	if project == nil {
		utils.ResponseParamsError(response, errors.ErrProjectNotFound)
		return
	}

	user := utils.GetUserFromAttribute(request)
	if user == nil {
		utils.ResponseAuthError(response)
		return
	}

	// get client from db
	clientName := request.PathParameter("name")
	clientInDB := t.tokenStore.GetClient(project.ProjectCode, clientName)
	if clientInDB == nil {
		utils.ResponseParamsError(response, fmt.Errorf("client %s not found", clientName))
		return
	}
	if !user.IsAdmin() && user.Name != clientInDB.CreatedBy && !clientInDB.IsManager(user.Name) {
		utils.ResponseAuthError(response)
		return
	}

	// update manager
	form := UpdateProjectClientForm{}
	_ = request.ReadEntity(&form)
	err := utils.Validate.Struct(&form)
	if err != nil {
		utils.ResponseParamsError(response, err)
		return
	}
	manager := strings.Join(form.Manager, ",")
	err = t.tokenStore.UpdateClientToken(project.ProjectCode, clientName,
		&models.BcsClient{
			ProjectCode:   clientInDB.ProjectCode,
			Name:          clientInDB.Name,
			Manager:       &manager,
			AuthorityUser: &clientInDB.AuthorityUser,
			CreatedBy:     clientInDB.CreatedBy,
		})
	if err != nil {
		blog.Errorf("update client %s manager failed, %s", clientName, err.Error())
		utils.ResponseDBError(response, err)
		return
	}

	utils.ResponseOK(response, nil)
}

// DeleteProjectClient 删除项目下平台账号
func (t *TokenHandler) DeleteProjectClient(request *restful.Request, response *restful.Response) {
	project := utils.GetProjectFromAttribute(request)
	if project == nil {
		utils.ResponseParamsError(response, errors.ErrProjectNotFound)
		return
	}

	user := utils.GetUserFromAttribute(request)
	if user == nil {
		utils.ResponseAuthError(response)
		return
	}

	// get client from db
	clientName := request.PathParameter("name")
	clientInDB := t.tokenStore.GetClient(project.ProjectCode, clientName)
	if clientInDB == nil {
		utils.ResponseParamsError(response, fmt.Errorf("client %s not found", clientName))
		return
	}
	if !user.IsAdmin() && user.Name != clientInDB.CreatedBy && !clientInDB.IsManager(user.Name) {
		utils.ResponseAuthError(response)
		return
	}

	// delete client
	err := t.tokenStore.DeleteProjectClient(project.ProjectCode, clientName)
	if err != nil {
		blog.Errorf("delete client %s failed, %s", clientName, err.Error())
		utils.ResponseDBError(response, err)
		return
	}

	_, err = t.cache.Del(request.Request.Context(), constant.TokenKeyPrefix+clientInDB.UserToken)
	if err != nil {
		blog.Errorf("delete client %s token failed, %s", clientName, err.Error())
		utils.ResponseSystemError(response, fmt.Errorf("delete client %s token failed, %s", clientName, err.Error()))
		return
	}
	for _, v := range clientInDB.AuthorityUserList() {
		key := fmt.Sprintf("%sop-%s:%s", constant.TokenKeyPrefix, v, clientInDB.UserToken)
		_, _ = t.cache.Del(request.Request.Context(), key)
	}

	utils.ResponseOK(response, nil)
}

// AuthorizeClientForm is the form of authorizing client
type AuthorizeClientForm struct {
	Username string `json:"username" validate:"required"`
}

// AuthorizeClient 给平台账号授权
func (t *TokenHandler) AuthorizeClient(request *restful.Request, response *restful.Response) {
	project := utils.GetProjectFromAttribute(request)
	if project == nil {
		utils.ResponseParamsError(response, errors.ErrProjectNotFound)
		return
	}

	user := utils.GetUserFromAttribute(request)
	if user == nil {
		utils.ResponseAuthError(response)
		return
	}
	form := AuthorizeClientForm{}
	_ = request.ReadEntity(&form)
	err := utils.Validate.Struct(&form)
	if err != nil {
		utils.ResponseParamsError(response, err)
		return
	}

	// 管理员和本人才能授权
	if !user.IsAdmin() && user.Name != form.Username {
		utils.ResponseAuthError(response)
		return
	}

	// get client from db
	clientName := request.PathParameter("name")
	clientInDB := t.tokenStore.GetClient(project.ProjectCode, clientName)
	if clientInDB == nil {
		utils.ResponseParamsError(response, fmt.Errorf("client %s not found", clientName))
		return
	}

	// create jwt token
	key := fmt.Sprintf("%sop-%s:%s", constant.TokenKeyPrefix, form.Username, clientInDB.UserToken)
	now := time.Now()
	jwtString, err := t.jwtClient.JWTSign(&jwt.UserInfo{
		SubType:     jwt.User.String(),
		UserName:    form.Username,
		ExpiredTime: int64(clientInDB.ExpiresAt.Sub(now).Seconds()),
		Issuer:      jwt.JWTIssuer,
	})
	if err != nil {
		utils.ResponseSystemError(response, fmt.Errorf("create jwt token failed, %s", err.Error()))
		return
	}
	_, err = t.cache.SetNX(request.Request.Context(), key, jwtString, clientInDB.ExpiresAt.Sub(now))
	if err != nil {
		utils.ResponseSystemError(response, fmt.Errorf("authority user %s for client %s fail, %s",
			form.Username, clientName, err.Error()))
		return
	}

	// update authorize
	if clientInDB.IsAuthorityUser(form.Username) {
		utils.ResponseOK(response, nil)
		return
	}
	manager := strings.Join(append(clientInDB.AuthorityUserList(), form.Username), ",")
	err = t.tokenStore.UpdateClientToken(project.ProjectCode, clientName,
		&models.BcsClient{
			ProjectCode:   clientInDB.ProjectCode,
			Name:          clientInDB.Name,
			Manager:       &clientInDB.Manager,
			AuthorityUser: &manager,
			CreatedBy:     clientInDB.CreatedBy,
			CreatedAt:     clientInDB.CreatedAt,
		})
	if err != nil {
		blog.Errorf("update client %s authorize failed, %s", clientName, err.Error())
		utils.ResponseDBError(response, err)
		return
	}
	utils.ResponseOK(response, nil)
}

// DeAuthorizeClient 取消平台账号授权
func (t *TokenHandler) DeAuthorizeClient(request *restful.Request, response *restful.Response) {
	project := utils.GetProjectFromAttribute(request)
	if project == nil {
		utils.ResponseParamsError(response, errors.ErrProjectNotFound)
		return
	}

	// check user permission
	user := utils.GetUserFromAttribute(request)
	if user == nil {
		utils.ResponseAuthError(response)
		return
	}
	form := AuthorizeClientForm{}
	_ = request.ReadEntity(&form)
	err := utils.Validate.Struct(&form)
	if err != nil {
		utils.ResponseParamsError(response, err)
		return
	}

	// 管理员和本人才能解授权
	if !user.IsAdmin() && user.Name != form.Username {
		utils.ResponseAuthError(response)
		return
	}

	// get client from db
	clientName := request.PathParameter("name")
	clientInDB := t.tokenStore.GetClient(project.ProjectCode, clientName)
	if clientInDB == nil {
		utils.ResponseParamsError(response, fmt.Errorf("client %s not found", clientName))
		return
	}

	key := fmt.Sprintf("%sop-%s:%s", constant.TokenKeyPrefix, form.Username, clientInDB.UserToken)
	_, _ = t.cache.Del(request.Request.Context(), key)

	// update authorize
	if !clientInDB.IsAuthorityUser(form.Username) {
		utils.ResponseOK(response, nil)
		return
	}
	authorityUsers := make([]string, 0)
	for _, v := range clientInDB.AuthorityUserList() {
		if v != form.Username {
			authorityUsers = append(authorityUsers, v)
		}
	}
	authorityUser := strings.Join(authorityUsers, ",")
	err = t.tokenStore.UpdateClientToken(project.ProjectCode, clientName,
		&models.BcsClient{
			ProjectCode:   clientInDB.ProjectCode,
			Name:          clientInDB.Name,
			Manager:       &clientInDB.Manager,
			AuthorityUser: &authorityUser,
			CreatedBy:     clientInDB.CreatedBy,
			CreatedAt:     clientInDB.CreatedAt,
		})
	if err != nil {
		blog.Errorf("update client %s authorize failed, %s", clientName, err.Error())
		utils.ResponseDBError(response, err)
		return
	}
	utils.ResponseOK(response, nil)
}
