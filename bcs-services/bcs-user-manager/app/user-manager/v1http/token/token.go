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

package token

import (
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/constant"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
	"github.com/dchest/uniuri"
	"github.com/emicklei/go-restful"
)

const (
	NeverExpiredDuration = time.Hour * 24 * 365 * 10
)

var NeverExpired, _ = time.Parse(time.RFC3339, "2032-01-19T03:14:07Z")

type TokenHandler struct {
	tokenStore  sqlstore.TokenStore
	notifyStore sqlstore.TokenNotifyStore
	cache       cache.Cache
	jwtClient   jwt.BCSJWTAuthentication
}

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

type CreateTokenForm struct {
	Username string `json:"username" validate:"required"`
	// token expiration second, -1: never expire
	Expiration int `json:"expiration" validate:"required"`
}

type TokenStatus uint8

const (
	TokenStatusExpired TokenStatus = iota
	TokenStatusActive
)

type TokenResp struct {
	Token     string       `json:"token"`
	Status    *TokenStatus `json:"status,omitempty"`
	ExpiredAt *time.Time   `json:"expired_at"` // nil means never expired
}

type UpdateTokenForm struct {
	// token expiration second, 0: never expire
	Expiration int `json:"expiration" validate:"required"`
}

// check user token permission, if user is admin, then allow all.
// if user is not admin, then check the token is belonged to user.
func checkTokenCreateBy(request *restful.Request, targetUser string) (allow bool, createBy string) {
	currentUser := request.Attribute(constant.CurrentUserAttr)
	var userToken *models.BcsToken
	if v, ok := currentUser.(*models.BcsToken); ok {
		userToken = v
	} else {
		return false, ""
	}

	if userToken.UserType == sqlstore.AdminUser || userToken.UserType == sqlstore.SaasUser ||
		userToken.UserType == sqlstore.ClientUser {
		return true, userToken.Username
	}
	if userToken.Username == targetUser {
		return true, userToken.Username
	}
	return false, ""
}

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

	// transfer never expire to expiration
	if form.Expiration < 0 {
		form.Expiration = int(NeverExpiredDuration.Seconds())
	}

	// check token permission
	allow, createBy := checkTokenCreateBy(request, form.Username)
	if !allow {
		message := fmt.Sprintf("errcode: %d, not allow to access tokens", common.BcsErrApiUnauthorized)
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
	})
	if err != nil {
		blog.Errorf("create jwt token failed, %s", err.Error())
		metrics.ReportRequestAPIMetrics("CreateToken", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, create jwt token failed", common.BcsErrApiInternalDbError)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}
	_, err = t.cache.Set(key, jwtString, time.Duration(form.Expiration)*time.Second)
	if err != nil {
		blog.Errorf("set user %s token fail", form.Username)
		metrics.ReportRequestAPIMetrics("CreateToken", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, creating user [%s] token failed", common.BcsErrApiInternalDbError, form.Username)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	// insert token record in db
	userToken := &models.BcsToken{
		Username:  form.Username,
		Token:     token,
		UserType:  sqlstore.PlainUser,
		CreatedBy: createBy,
		ExpiresAt: expiredAt,
	}
	err = t.tokenStore.CreateToken(userToken)
	if err != nil {
		// delete token from redis when fail to insert token in db
		_, _ = t.cache.Del(key)
		metrics.ReportRequestAPIMetrics("CreateToken", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("failed to insert user token record [%s]: %s", userToken.Username, err.Error())
		message := fmt.Sprintf("errcode: %d, creating token for user [%s] failed, error: %s",
			common.BcsErrApiInternalDbError, userToken.Username, err)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	resp := &TokenResp{Token: token, ExpiredAt: &userToken.ExpiresAt}
	// transfer never expired token
	if resp.ExpiredAt.After(NeverExpired) {
		resp.ExpiredAt = nil
	}
	data := utils.CreateResponseData(nil, "success", *resp)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("CreateToken", request.Request.Method, metrics.SucStatus, start)
}

func (t *TokenHandler) GetToken(request *restful.Request, response *restful.Response) {
	start := time.Now()
	username := request.PathParameter("username")

	// check token permission
	allow, _ := checkTokenCreateBy(request, username)
	if !allow {
		message := fmt.Sprintf("errcode: %d, not allow to access tokens", common.BcsErrApiUnauthorized)
		utils.WriteUnauthorizedError(response, common.BcsErrApiUnauthorized, message)
		return
	}

	tokensInDB := t.tokenStore.GetUserTokensByName(username)
	tokens := make([]TokenResp, 0)
	for _, v := range tokensInDB {
		status := TokenStatusActive
		if v.HasExpired() {
			status = TokenStatusExpired
		}
		expiresAt := &v.ExpiresAt
		// transfer never expired
		if v.ExpiresAt.After(NeverExpired) {
			expiresAt = nil
		}
		tokens = append(tokens, TokenResp{Token: v.Token, Status: &status, ExpiredAt: expiresAt})
	}
	data := utils.CreateResponseData(nil, "success", tokens)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("GetToken", request.Request.Method, metrics.SucStatus, start)
}

func (t *TokenHandler) DeleteToken(request *restful.Request, response *restful.Response) {
	start := time.Now()
	token := request.PathParameter("token")

	tokenInDB := t.tokenStore.GetTokenByCondition(&models.BcsToken{Token: token})
	if tokenInDB == nil {
		metrics.ReportRequestAPIMetrics("DeleteToken", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("failed to delete token, token [%s] not exists", token)
		message := fmt.Sprintf("errcode: %d, delete user token failed", common.BcsErrApiInternalDbError)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	// check token permission
	allow, _ := checkTokenCreateBy(request, tokenInDB.Username)
	if !allow {
		message := fmt.Sprintf("errcode: %d, not allow to access tokens", common.BcsErrApiUnauthorized)
		utils.WriteUnauthorizedError(response, common.BcsErrApiUnauthorized, message)
		return
	}

	_, err := t.cache.Del(constant.TokenKeyPrefix + token)
	if err != nil {
		blog.Errorf("delete token failed")
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

	tokenInDB := t.tokenStore.GetTokenByCondition(&models.BcsToken{Token: token})
	if tokenInDB == nil {
		metrics.ReportRequestAPIMetrics("UpdateToken", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("failed to update token, token [%s] not exists", token)
		message := fmt.Sprintf("errcode: %d, update user token failed", common.BcsErrApiInternalDbError)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	// check token permission
	allow, _ := checkTokenCreateBy(request, tokenInDB.Username)
	if !allow {
		message := fmt.Sprintf("errcode: %d, not allow to access tokens", common.BcsErrApiUnauthorized)
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
	jwtString, err := t.jwtClient.JWTSign(&jwt.UserInfo{
		SubType:     jwt.User.String(),
		UserName:    tokenInDB.Username,
		ExpiredTime: int64(form.Expiration),
		Issuer:      jwt.JWTIssuer,
	})
	if err != nil {
		blog.Errorf("recreate jwt token failed, %s", err.Error())
		metrics.ReportRequestAPIMetrics("UpdateToken", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, recreate jwt token failed", common.BcsErrApiInternalDbError)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}
	_, err = t.cache.Set(key, jwtString, time.Duration(form.Expiration)*time.Second)
	if err != nil {
		blog.Errorf("set user [%s] token fail, err: %s", tokenInDB.Username, err.Error())
		metrics.ReportRequestAPIMetrics("UpdateToken", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, update user [%s] token failed", common.BcsErrApiInternalDbError,
			tokenInDB.Username)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	// update token record in db
	newToken := tokenInDB
	newToken.ExpiresAt = expiredAt
	_ = t.tokenStore.UpdateToken(&models.BcsToken{Token: token}, newToken)
	_ = t.notifyStore.DeleteTokenNotify(token)

	resp := &TokenResp{Token: token, ExpiredAt: &newToken.ExpiresAt}
	data := utils.CreateResponseData(nil, "success", *resp)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("UpdateToken", request.Request.Method, metrics.SucStatus, start)
}

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
	allow, createBy := checkTokenCreateBy(request, form.Username)
	if !allow || form.Username == createBy {
		message := fmt.Sprintf("errcode: %d, not allow to access tokens", common.BcsErrApiUnauthorized)
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
	})
	if err != nil {
		blog.Errorf("create jwt token failed, %s", err.Error())
		metrics.ReportRequestAPIMetrics("CreateTempToken", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, create jwt token failed", common.BcsErrApiInternalDbError)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}
	_, err = t.cache.Set(key, jwtString, time.Duration(form.Expiration)*time.Second)
	if err != nil {
		blog.Errorf("set user %s token fail", form.Username)
		metrics.ReportRequestAPIMetrics("CreateTempToken", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, creating user [%s] token failed", common.BcsErrApiInternalDbError, form.Username)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	// insert token record in db
	userToken := &models.BcsTempToken{
		Username:  form.Username,
		Token:     token,
		UserType:  sqlstore.PlainUser,
		CreatedBy: createBy,
		ExpiresAt: expiredAt,
	}
	err = t.tokenStore.CreateTemporaryToken(userToken)
	if err != nil {
		// delete token from redis when fail to insert token in db
		_, _ = t.cache.Del(key)
		metrics.ReportRequestAPIMetrics("CreateTempToken", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("failed to insert user token record [%s]: %s", userToken.Username, err.Error())
		message := fmt.Sprintf("errcode: %d, creating temporary token for user [%s] failed, error: %s",
			common.BcsErrApiInternalDbError, userToken.Username, err)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	resp := &TokenResp{Token: token, ExpiredAt: &userToken.ExpiresAt}
	data := utils.CreateResponseData(nil, "success", *resp)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("CreateTempToken", request.Request.Method, metrics.SucStatus, start)
}
