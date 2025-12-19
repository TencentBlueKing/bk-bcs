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
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	"github.com/dchest/uniuri"
	restful "github.com/emicklei/go-restful/v3"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
)

const (
	// NeverExpiredDuration is a very large number that represents a time duration,
	// when user create a never expired token, the token expiration will be 10 years.
	NeverExpiredDuration = time.Hour * 24 * 365 * 10

	// seconds, 1000 days
	maxExpiredDuration = 60 * 60 * 24 * 1000
)

// NeverExpired mean when user's expiration time less than NeverExpired time,
// then the token will never be expired.
var NeverExpired, _ = time.Parse(time.RFC3339, "2032-01-19T03:14:07Z")

// TokenHandler is a restful handler for token.
// nolint
type TokenHandler struct {
	tokenStore  sqlstore.TokenStore
	notifyStore sqlstore.TokenNotifyStore
	cache       cache.Cache
	jwtClient   jwt.BCSJWTAuthentication
}

// NewTokenHandler is a constructor for TokenHandler.
func NewTokenHandler(tokenStore sqlstore.TokenStore, notifyStore sqlstore.TokenNotifyStore, cache cache.Cache,
	jwtClient jwt.BCSJWTAuthentication) *TokenHandler {
	return &TokenHandler{
		tokenStore:  tokenStore,
		notifyStore: notifyStore,
		cache:       cache,
		jwtClient:   jwtClient,
	}
}

// token request payload

// CreateTokenForm is a form for create token.
type CreateTokenForm struct {
	UserType uint   `json:"usertype"`
	Username string `json:"username" validate:"required"`
	// token expiration second, -1: never expire
	Expiration int `json:"expiration" validate:"required"`
}

// TokenResp is a response for creating token and other token handler's response data.
// nolint
type TokenResp struct {
	Token     string             `json:"token"`
	TenantID  string             `json:"tenant_id,omitempty"`
	JWT       string             `json:"jwt,omitempty"`
	Status    *utils.TokenStatus `json:"status,omitempty"`
	ExpiredAt *time.Time         `json:"expired_at"` // nil means never expired
}

// UpdateTokenForm is a form for update token.
type UpdateTokenForm struct {
	// token expiration second, 0: never expire
	Expiration int `json:"expiration" validate:"required"`
}

// checkTokenCreateBy xxx
// check user token permission, if user is admin, then allow all.
// if user is not admin, then check the token is belonged to user.
func checkTokenCreateBy(request *restful.Request, targetUser string) (allow, isClient bool, createBy string) {
	currentUser := request.Attribute(constant.CurrentUserAttr)
	var userToken *models.BcsUser
	// nolint
	if v, ok := currentUser.(*models.BcsUser); ok {
		userToken = v
	} else {
		return false, false, ""
	}

	if userToken.IsClient() {
		return true, true, userToken.Name
	}
	if userToken.Name == targetUser {
		return true, false, userToken.Name
	}
	return false, false, ""
}

// CreateToken create a token for user.
// NOCC:golint/fnsize(设计如此)
// nolint
func (t *TokenHandler) CreateToken(request *restful.Request, response *restful.Response) {
	start := time.Now()
	form := CreateTokenForm{}
	_ = request.ReadEntity(&form)
	err := utils.Validate.Struct(&form)
	if err != nil {
		blog.Errorf("formation of creating token request from %s is invalid, %s", request.Request.RemoteAddr, err.Error())
		metrics.ReportRequestAPIMetrics("CreateToken", request.Request.Method, metrics.ErrStatus, start)
		_ = response.WriteHeaderAndEntity(400, utils.FormatValidationError(err))
		return
	}

	if form.Expiration > maxExpiredDuration {
		blog.Infof("user %s create token with %d expiration", form.Username, form.Expiration)
		metrics.ReportRequestAPIMetrics("CreateToken", request.Request.Method, metrics.ErrStatus, start)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, "expiration must <= 1000 days")
		return
	}

	// transfer never expire to expiration
	if form.Expiration < 0 {
		form.Expiration = int(NeverExpiredDuration.Seconds())
	}

	// check token permission
	allow, _, createBy := checkTokenCreateBy(request, form.Username)
	if !allow {
		message := fmt.Sprintf("errcode: %d, not allow to access tokens", common.BcsErrApiUnauthorized)
		metrics.ReportRequestAPIMetrics("CreateToken", request.Request.Method, metrics.ErrStatus, start)
		utils.WriteUnauthorizedError(response, common.BcsErrApiUnauthorized, message)
		return
	}

	// if user has token that not expired, return error
	tokens := t.tokenStore.GetUserTokensByName(form.Username)
	if len(tokens) >= constant.TokenLimits {
		blog.Errorf("user %s token already exists", form.Username)
		metrics.ReportRequestAPIMetrics("CreateToken", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, token already exists", common.BcsErrApiInternalDbError)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	// create jwt token
	token := uniuri.NewLen(constant.DefaultTokenLength)
	key := constant.TokenKeyPrefix + token
	expiredAt := time.Now().Add(time.Duration(form.Expiration) * time.Second)
	jwtString, err := t.jwtClient.JWTSign(&jwt.UserInfo{
		SubType:     jwt.User.String(),
		UserName:    form.Username,
		ExpiredTime: int64(form.Expiration),
		Issuer:      jwt.JWTIssuer,
		TenantId:    utils.GetTenantIDFromAttribute(request),
	})
	if err != nil {
		blog.Errorf("create jwt token failed, %s", err.Error())
		metrics.ReportRequestAPIMetrics("CreateToken", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, create jwt token failed", common.BcsErrApiInternalDbError)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}
	_, err = t.cache.Set(context.TODO(), key, jwtString, time.Duration(form.Expiration)*time.Second)
	if err != nil {
		blog.Errorf("set user %s token fail", form.Username)
		metrics.ReportRequestAPIMetrics("CreateToken", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, creating user [%s] token failed", common.BcsErrApiInternalDbError, form.Username)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	// insert token record in db
	userToken := &models.BcsUser{
		Name:      form.Username,
		UserToken: token,
		UserType:  models.PlainUser,
		CreatedBy: createBy,
		ExpiresAt: expiredAt,
		TenantID:  utils.GetTenantIDFromAttribute(request),
	}
	err = t.tokenStore.CreateToken(userToken)
	if err != nil {
		// delete token from redis when fail to insert token in db
		_, _ = t.cache.Del(context.TODO(), key)
		metrics.ReportRequestAPIMetrics("CreateToken", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("failed to insert user token record [%s]: %s", userToken.Name, err.Error())
		message := fmt.Sprintf("errcode: %d, creating token for user [%s] failed, error: %s",
			common.BcsErrApiInternalDbError, userToken.Name, err)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	resp := &TokenResp{Token: token, TenantID: userToken.TenantID, ExpiredAt: &userToken.ExpiresAt}
	// transfer never expired token
	if resp.ExpiredAt.After(NeverExpired) {
		resp.ExpiredAt = nil
	}
	data := utils.CreateResponseData(nil, "success", *resp)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("CreateToken", request.Request.Method, metrics.SucStatus, start)
}

// GetToken get token
func (t *TokenHandler) GetToken(request *restful.Request, response *restful.Response) {
	start := time.Now()
	username := request.PathParameter("username")

	// check token permission
	allow, _, _ := checkTokenCreateBy(request, username)
	if !allow {
		message := fmt.Sprintf("errcode: %d, not allow to access tokens", common.BcsErrApiUnauthorized)
		metrics.ReportRequestAPIMetrics("GetToken", request.Request.Method, metrics.ErrStatus, start)
		utils.WriteUnauthorizedError(response, common.BcsErrApiUnauthorized, message)
		return
	}

	tokensInDB := t.tokenStore.GetUserTokensByName(username)
	tokens := make([]TokenResp, 0)
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
		tokens = append(tokens, TokenResp{
			Token: v.UserToken, TenantID: v.TenantID, Status: &status, ExpiredAt: expiresAt})
	}
	data := utils.CreateResponseData(nil, "success", tokens)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("GetToken", request.Request.Method, metrics.SucStatus, start)
}

// DeleteToken delete token
func (t *TokenHandler) DeleteToken(request *restful.Request, response *restful.Response) {
	start := time.Now()
	token := request.PathParameter("token")
	force := request.QueryParameter("force") // 是否强制删除

	tokenInDB := t.tokenStore.GetTokenByCondition(&models.BcsUser{UserToken: token})
	if tokenInDB == nil {
		metrics.ReportRequestAPIMetrics("DeleteToken", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("failed to delete token, token [%s] not exists", token)
		message := fmt.Sprintf("errcode: %d, delete user token failed", common.BcsErrApiInternalDbError)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	if tokenInDB.IsAdmin() && force != "true" {
		message := fmt.Sprintf("errcode: %d, not allow to delete admin tokens", common.BcsErrApiBadRequest)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}

	// check token permission
	allow, _, _ := checkTokenCreateBy(request, tokenInDB.Name)
	if !allow {
		message := fmt.Sprintf("errcode: %d, not allow to access tokens", common.BcsErrApiUnauthorized)
		metrics.ReportRequestAPIMetrics("DeleteToken", request.Request.Method, metrics.ErrStatus, start)
		utils.WriteUnauthorizedError(response, common.BcsErrApiUnauthorized, message)
		return
	}

	_, err := t.cache.Del(context.TODO(), constant.TokenKeyPrefix+token)
	if err != nil {
		blog.Errorf("delete token failed, %s", err.Error())
		metrics.ReportRequestAPIMetrics("DeleteToken", request.Request.Method, metrics.ErrStatus, start)
	}

	err = t.tokenStore.DeleteToken(token)
	if err != nil {
		metrics.ReportRequestAPIMetrics("DeleteToken", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("failed to delete token: %s", err.Error())
		message := fmt.Sprintf("errcode: %d, delete user token failed", common.BcsErrApiInternalDbError)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	data := utils.CreateResponseData(nil, "success", nil)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("DeleteToken", request.Request.Method, metrics.SucStatus, start)
}

// UpdateToken update token
// NOCC:golint/fnsize(设计如此)
func (t *TokenHandler) UpdateToken(request *restful.Request, response *restful.Response) {
	start := time.Now()
	token := request.PathParameter("token")
	form := UpdateTokenForm{}
	_ = request.ReadEntity(&form)
	err := utils.Validate.Struct(&form)
	if err != nil {
		blog.Errorf("formation of update token request from %s is invalid, %s", request.Request.RemoteAddr, err.Error())
		metrics.ReportRequestAPIMetrics("UpdateToken", request.Request.Method, metrics.ErrStatus, start)
		_ = response.WriteHeaderAndEntity(400, utils.FormatValidationError(err))
		return
	}

	if form.Expiration > maxExpiredDuration {
		blog.Infof("user create token with %d expiration", form.Expiration)
		metrics.ReportRequestAPIMetrics("UpdateToken", request.Request.Method, metrics.ErrStatus, start)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, "expiration must <= 1000 days")
		return
	}

	tokenInDB := t.tokenStore.GetTokenByCondition(&models.BcsUser{UserToken: token})
	if tokenInDB == nil {
		metrics.ReportRequestAPIMetrics("UpdateToken", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("failed to update token, token [%s] not exists", token)
		message := fmt.Sprintf("errcode: %d, update user token failed", common.BcsErrApiInternalDbError)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	// check token permission
	allow, _, _ := checkTokenCreateBy(request, tokenInDB.Name)
	if !allow {
		message := fmt.Sprintf("errcode: %d, not allow to access tokens", common.BcsErrApiUnauthorized)
		metrics.ReportRequestAPIMetrics("UpdateToken", request.Request.Method, metrics.ErrStatus, start)
		utils.WriteUnauthorizedError(response, common.BcsErrApiUnauthorized, message)
		return
	}

	// if token is expired, then use expiration.
	if form.Expiration < 0 {
		form.Expiration = int(NeverExpiredDuration.Seconds())
	}
	if !tokenInDB.HasExpired() {
		remain := time.Until(tokenInDB.ExpiresAt)
		form.Expiration += int(remain.Seconds())
	}
	expiredAt := time.Now().Add(time.Duration(form.Expiration) * time.Second)

	// create jwt token
	key := constant.TokenKeyPrefix + token
	userInfo := &jwt.UserInfo{
		ExpiredTime: int64(form.Expiration),
		Issuer:      jwt.JWTIssuer,
		TenantId:    tokenInDB.TenantID,
	}
	if tokenInDB.IsClient() {
		userInfo.SubType = jwt.Client.String()
		userInfo.ClientName = tokenInDB.Name
	} else {
		userInfo.SubType = jwt.User.String()
		userInfo.UserName = tokenInDB.Name
	}
	jwtString, err := t.jwtClient.JWTSign(userInfo)
	if err != nil {
		blog.Errorf("recreate jwt token failed, %s", err.Error())
		metrics.ReportRequestAPIMetrics("UpdateToken", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, recreate jwt token failed", common.BcsErrApiInternalDbError)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}
	_, err = t.cache.Set(context.TODO(), key, jwtString, time.Duration(form.Expiration)*time.Second)
	if err != nil {
		blog.Errorf("set user [%s] token fail, err: %s", tokenInDB.Name, err.Error())
		metrics.ReportRequestAPIMetrics("UpdateToken", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, update user [%s] token failed", common.BcsErrApiInternalDbError,
			tokenInDB.Name)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	// update token record in db
	newToken := tokenInDB
	newToken.ExpiresAt = expiredAt
	_ = t.tokenStore.UpdateToken(&models.BcsUser{UserToken: token}, newToken)
	_ = t.notifyStore.DeleteTokenNotify(token)

	resp := &TokenResp{Token: token, ExpiredAt: &newToken.ExpiresAt}
	data := utils.CreateResponseData(nil, "success", *resp)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("UpdateToken", request.Request.Method, metrics.SucStatus, start)
}

// CreateTempToken create temp token
func (t *TokenHandler) CreateTempToken(request *restful.Request, response *restful.Response) {
	start := time.Now()
	form := CreateTokenForm{}
	_ = request.ReadEntity(&form)
	err := utils.Validate.Struct(&form)
	if err != nil {
		blog.Errorf("formation of creating token request from %s is invalid, %s", request.Request.RemoteAddr, err.Error())
		metrics.ReportRequestAPIMetrics("CreateTempToken", request.Request.Method, metrics.ErrStatus, start)
		_ = response.WriteHeaderAndEntity(400, utils.FormatValidationError(err))
		return
	}

	// check token permission
	allow, isClient, createBy := checkTokenCreateBy(request, form.Username)
	if !allow || !isClient {
		message := fmt.Sprintf("errcode: %d, not allow to access tokens", common.BcsErrApiUnauthorized)
		metrics.ReportRequestAPIMetrics("CreateTempToken", request.Request.Method, metrics.ErrStatus, start)
		utils.WriteUnauthorizedError(response, common.BcsErrApiUnauthorized, message)
		return
	}

	// create jwt token
	token := uniuri.NewLen(constant.DefaultTokenLength)
	key := constant.TokenKeyPrefix + token
	expiredAt := time.Now().Add(time.Duration(form.Expiration) * time.Second)
	jwtString, err := t.jwtClient.JWTSign(&jwt.UserInfo{
		SubType:     jwt.User.String(),
		UserName:    form.Username,
		ExpiredTime: int64(form.Expiration),
		Issuer:      jwt.JWTIssuer,
		TenantId:    utils.GetTenantIDFromAttribute(request),
	})
	if err != nil {
		blog.Errorf("create jwt token failed, %s", err.Error())
		metrics.ReportRequestAPIMetrics("CreateTempToken", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, create jwt token failed", common.BcsErrApiInternalDbError)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}
	_, err = t.cache.Set(context.TODO(), key, jwtString, time.Duration(form.Expiration)*time.Second)
	if err != nil {
		blog.Errorf("set user %s token fail", form.Username)
		metrics.ReportRequestAPIMetrics("CreateTempToken", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, creating user [%s] token failed", common.BcsErrApiInternalDbError, form.Username)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	userType := getTempTokenUserType(form.UserType)
	// insert token record in db
	userToken := &models.BcsTempToken{
		Username:  form.Username,
		Token:     token,
		UserType:  userType,
		CreatedBy: createBy,
		ExpiresAt: expiredAt,
		TenantID:  utils.GetTenantIDFromAttribute(request),
	}
	err = t.tokenStore.CreateTemporaryToken(userToken)
	if err != nil {
		// delete token from redis when fail to insert token in db
		_, _ = t.cache.Del(context.TODO(), key)
		metrics.ReportRequestAPIMetrics("CreateTempToken", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("failed to insert user token record [%s]: %s", userToken.Username, err.Error())
		message := fmt.Sprintf("errcode: %d, creating temporary token for user [%s] failed, error: %s",
			common.BcsErrApiInternalDbError, userToken.Username, err)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	resp := &TokenResp{Token: token, TenantID: userToken.TenantID, ExpiredAt: &userToken.ExpiresAt}
	data := utils.CreateResponseData(nil, "success", *resp)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("CreateTempToken", request.Request.Method, metrics.SucStatus, start)
}

func getTempTokenUserType(userType uint) uint {
	switch userType {
	case models.AdminUser, models.SaasUser, models.PlainUser, models.ClientUser:
		return userType
	default:
		userType = models.PlainUser
	}

	return userType
}

// CreateClientTokenForm is the form of creating client token
type CreateClientTokenForm struct {
	// ClientName name
	ClientName string `json:"clientName" validate:"required"`
	// ClientSecret secret
	ClientSecret string `json:"clientSecret"`
	// Expiration token expiration second, -1: never expire
	Expiration int `json:"expiration" validate:"required"`
}

// CreateClientToken create client token
// NOCC:golint/fnsize(设计如此)
// nolint
func (t *TokenHandler) CreateClientToken(request *restful.Request, response *restful.Response) {
	start := time.Now()
	form := CreateClientTokenForm{}
	_ = request.ReadEntity(&form)
	err := utils.Validate.Struct(&form)
	if err != nil {
		blog.Errorf("formation of creating token request from %s is invalid, %s", request.Request.RemoteAddr, err.Error())
		metrics.ReportRequestAPIMetrics("CreateClientToken", request.Request.Method, metrics.ErrStatus, start)
		_ = response.WriteHeaderAndEntity(400, utils.FormatValidationError(err))
		return
	}

	// check token permission
	allow, _, createBy := checkTokenCreateBy(request, form.ClientName)
	if !allow {
		message := fmt.Sprintf("errcode: %d, not allow to create client token", common.BcsErrApiUnauthorized)
		metrics.ReportRequestAPIMetrics("CreateClientToken", request.Request.Method, metrics.ErrStatus, start)
		utils.WriteUnauthorizedError(response, common.BcsErrApiUnauthorized, message)
		return
	}

	// check exist token
	exist := t.tokenStore.GetTokenByCondition(&models.BcsUser{Name: form.ClientName, UserType: models.ClientUser})
	if exist != nil {
		jwtString, _ := t.cache.Get(context.TODO(), constant.TokenKeyPrefix+exist.UserToken)
		resp := &TokenResp{Token: exist.UserToken, ExpiredAt: &exist.ExpiresAt, JWT: jwtString}
		// transfer never expired token
		if resp.ExpiredAt.After(NeverExpired) {
			resp.ExpiredAt = nil
		}
		data := utils.CreateResponseData(nil, "success", *resp)
		_, _ = response.Write([]byte(data))

		metrics.ReportRequestAPIMetrics("CreateClientToken", request.Request.Method, metrics.SucStatus, start)
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
		SubType:      jwt.Client.String(),
		ClientName:   form.ClientName,
		ClientSecret: form.ClientSecret,
		ExpiredTime:  int64(form.Expiration),
		Issuer:       jwt.JWTIssuer,
		TenantId:     utils.GetTenantIDFromAttribute(request),
	})
	if err != nil {
		blog.Errorf("create jwt token failed, %s", err.Error())
		metrics.ReportRequestAPIMetrics("CreateClientToken", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, create jwt token failed", common.BcsErrApiInternalDbError)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}
	_, err = t.cache.Set(context.TODO(), key, jwtString, time.Duration(form.Expiration)*time.Second)
	if err != nil {
		blog.Errorf("set client %s token fail", form.ClientName)
		metrics.ReportRequestAPIMetrics("CreateClientToken", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, creating client [%s] token failed", common.BcsErrApiInternalDbError,
			form.ClientName)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	// insert token record in db
	userToken := &models.BcsUser{
		Name:      form.ClientName,
		UserToken: token,
		UserType:  models.ClientUser,
		CreatedBy: createBy,
		ExpiresAt: expiredAt,
		TenantID:  utils.GetTenantIDFromAttribute(request),
	}
	err = t.tokenStore.CreateToken(userToken)
	if err != nil {
		// delete token from redis when fail to insert token in db
		_, _ = t.cache.Del(context.TODO(), key)
		metrics.ReportRequestAPIMetrics("CreateClientToken", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("failed to insert client token record [%s]: %s", userToken.Name, err.Error())
		message := fmt.Sprintf("errcode: %d, creating client token for client [%s] failed, error: %s",
			common.BcsErrApiInternalDbError, userToken.Name, err)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	resp := &TokenResp{Token: token, ExpiredAt: &userToken.ExpiresAt, JWT: jwtString}
	// transfer never expired token
	if resp.ExpiredAt.After(NeverExpired) {
		resp.ExpiredAt = nil
	}
	data := utils.CreateResponseData(nil, "success", *resp)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("CreateClientToken", request.Request.Method, metrics.SucStatus, start)
}
