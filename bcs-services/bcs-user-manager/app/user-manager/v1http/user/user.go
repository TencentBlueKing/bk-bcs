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

package user

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/constant"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"

	"github.com/dchest/uniuri"
	"github.com/emicklei/go-restful"
)

// CreateAdminUser create a admin user
func CreateAdminUser(request *restful.Request, response *restful.Response) {
	start := time.Now()

	userName := request.PathParameter("user_name")
	user := &models.BcsUser{
		Name:     userName,
		UserType: sqlstore.AdminUser,
	}
	// if this user already exist
	userInDb := sqlstore.GetUserByCondition(user)
	if userInDb != nil {
		metrics.ReportRequestAPIMetrics("CreateAdminUser", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, user [%s] already exist", common.BcsErrApiBadRequest, userName)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}
	user.UserToken = uniuri.NewLen(constant.DefaultTokenLength)
	user.ExpiresAt = time.Now().Add(sqlstore.AdminSaasUserExpiredTime)

	// create this user and save to db
	err := sqlstore.CreateUser(user)
	if err != nil {
		metrics.ReportRequestAPIMetrics("CreateAdminUser", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("failed to create user [%s]: %s", user.Name, err.Error())
		message := fmt.Sprintf("errcode: %d, creating user [%s] failed, error: %s", common.BcsErrApiInternalDbError, user.Name, err)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	data := utils.CreateResponseData(nil, "success", *user)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("CreateAdminUser", request.Request.Method, metrics.SucStatus, start)
}

// GetAdminUser get an admin user and usertoken information
func GetAdminUser(request *restful.Request, response *restful.Response) {
	start := time.Now()

	userName := request.PathParameter("user_name")
	user := sqlstore.GetUserByCondition(&models.BcsUser{Name: userName, UserType: sqlstore.AdminUser})
	if user == nil {
		metrics.ReportRequestAPIMetrics("GetAdminUser", request.Request.Method, metrics.ErrStatus, start)
		blog.Warnf("user [%s] not found in db", userName)
		message := fmt.Sprintf("errcode: %d, user with user_name=%s not found", common.BcsErrApiBadRequest, userName)
		utils.WriteNotFoundError(response, common.BcsErrApiBadRequest, message)
		return
	}

	data := utils.CreateResponseData(nil, "success", *user)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("GetAdminUser", request.Request.Method, metrics.SucStatus, start)
}

// CreateSaasUser create a saas user
func CreateSaasUser(request *restful.Request, response *restful.Response) {
	start := time.Now()

	userName := request.PathParameter("user_name")
	user := &models.BcsUser{
		Name:     userName,
		UserType: sqlstore.SaasUser,
	}
	// if this user already exist
	userInDb := sqlstore.GetUserByCondition(user)
	if userInDb != nil {
		metrics.ReportRequestAPIMetrics("CreateSaasUser", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, user [%s] already exist", common.BcsErrApiBadRequest, userName)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}

	user.UserToken = uniuri.NewLen(constant.DefaultTokenLength)
	user.ExpiresAt = time.Now().Add(sqlstore.AdminSaasUserExpiredTime)

	// create this user and save to db
	err := sqlstore.CreateUser(user)
	if err != nil {
		metrics.ReportRequestAPIMetrics("CreateSaasUser", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("failed to create user [%s]: %s", user.Name, err.Error())
		message := fmt.Sprintf("errcode: %d, creating user [%s] failed, error: %s", common.BcsErrApiInternalDbError, user.Name, err)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	data := utils.CreateResponseData(nil, "success", *user)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("CreateSaasUser", request.Request.Method, metrics.SucStatus, start)
}

// GetSaasUser get an saas user and usertoken information
func GetSaasUser(request *restful.Request, response *restful.Response) {
	start := time.Now()

	userName := request.PathParameter("user_name")
	user := sqlstore.GetUserByCondition(&models.BcsUser{Name: userName, UserType: sqlstore.SaasUser})
	if user == nil {
		metrics.ReportRequestAPIMetrics("GetSaasUser", request.Request.Method, metrics.ErrStatus, start)
		blog.Warnf("user [%s] not found in db", userName)
		message := fmt.Sprintf("errcode: %d, user with user_name=%s not found", common.BcsErrApiBadRequest, userName)
		utils.WriteNotFoundError(response, common.BcsErrApiBadRequest, message)
		return
	}

	data := utils.CreateResponseData(nil, "success", *user)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("GetSaasUser", request.Request.Method, metrics.SucStatus, start)
}

// CreatePlainUser create a plain user
func CreatePlainUser(request *restful.Request, response *restful.Response) {
	start := time.Now()

	userName := request.PathParameter("user_name")
	user := &models.BcsUser{
		Name:     userName,
		UserType: sqlstore.PlainUser,
	}
	// if this user already exist
	userInDb := sqlstore.GetUserByCondition(user)
	if userInDb != nil {
		metrics.ReportRequestAPIMetrics("CreatePlainUser", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, user [%s] already exist", common.BcsErrApiBadRequest, userName)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}

	user.UserToken = uniuri.NewLen(constant.DefaultTokenLength)
	user.ExpiresAt = time.Now().Add(sqlstore.PlainUserExpiredTime)

	// create this user and save to db
	err := sqlstore.CreateUser(user)
	if err != nil {
		metrics.ReportRequestAPIMetrics("CreatePlainUser", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("failed to create user [%s]: %s", user.Name, err.Error())
		message := fmt.Sprintf("errcode: %d, creating user [%s] failed, error: %s", common.BcsErrApiInternalDbError, user.Name, err)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	data := utils.CreateResponseData(nil, "success", *user)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("CreatePlainUser", request.Request.Method, metrics.SucStatus, start)
}

// GetPlainUser get a plain user and usertoken information
func GetPlainUser(request *restful.Request, response *restful.Response) {
	start := time.Now()

	userName := request.PathParameter("user_name")
	user := sqlstore.GetUserByCondition(&models.BcsUser{Name: userName, UserType: sqlstore.PlainUser})
	if user == nil {
		metrics.ReportRequestAPIMetrics("GetPlainUser", request.Request.Method, metrics.ErrStatus, start)
		blog.Warnf("failed to get user, user [%s] not found in db", userName)
		message := fmt.Sprintf("errcode: %d, user with user_name=%s not found", common.BcsErrApiBadRequest, userName)
		utils.WriteNotFoundError(response, common.BcsErrApiBadRequest, message)
		return
	}

	data := utils.CreateResponseData(nil, "success", *user)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("GetPlainUser", request.Request.Method, metrics.SucStatus, start)
}

// RefreshPlainToken refresh usertoken for a plain user
func RefreshPlainToken(request *restful.Request, response *restful.Response) {
	start := time.Now()

	userName := request.PathParameter("user_name")
	expireDays := request.PathParameter("expire_time")
	expireDaysInt, err := strconv.Atoi(expireDays)
	if err != nil {
		metrics.ReportRequestAPIMetrics("RefreshPlainToken", request.Request.Method, metrics.ErrStatus, start)
		blog.Warnf("invalid expire_time, failed to atoi: %s", err.Error())
		message := fmt.Sprintf("errcode: %d, invalid expire_time, failed to atoi: %s", common.BcsErrApiBadRequest, err.Error())
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}
	if expireDaysInt < 0 {
		metrics.ReportRequestAPIMetrics("RefreshPlainToken", request.Request.Method, metrics.ErrStatus, start)
		blog.Warnf("invalid expire_time: %d", expireDaysInt)
		message := fmt.Sprintf("errcode: %d, invalid expire_time", common.BcsErrApiBadRequest)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}

	user := sqlstore.GetUserByCondition(&models.BcsUser{Name: userName, UserType: sqlstore.PlainUser})
	if user == nil {
		metrics.ReportRequestAPIMetrics("RefreshPlainToken", request.Request.Method, metrics.ErrStatus, start)
		blog.Warnf("failed to refresh token, user [%s] not found in db", userName)
		message := fmt.Sprintf("errcode: %d, user with user_name=%s not found", common.BcsErrApiBadRequest, userName)
		utils.WriteNotFoundError(response, common.BcsErrApiBadRequest, message)
		return
	}

	expireTime := time.Duration(expireDaysInt) * sqlstore.PlainUserExpiredTime
	updatedUser := user
	// if usertoken has been expired, refresh the usertoken
	// or just refresh the expiresTime and return the same token
	if time.Now().After(user.ExpiresAt) {
		updatedUser.UserToken = uniuri.NewLen(constant.DefaultTokenLength)
		updatedUser.ExpiresAt = time.Now().Add(expireTime)
	} else {
		updatedUser.ExpiresAt = time.Now().Add(expireTime)
	}

	// update and save to db
	// if update failed, it's better to refresh by client
	err = sqlstore.UpdateUser(user, updatedUser)
	if err != nil {
		metrics.ReportRequestAPIMetrics("RefreshPlainToken", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("failed to refresh usertoken [%s]: %s", user.Name, err.Error())
		message := fmt.Sprintf("errcode: %d, failed to refresh usertoken [%s], error: %s", common.BcsErrApiInternalDbError, userName, err)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	data := utils.CreateResponseData(nil, "success", *user)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("RefreshPlainToken", request.Request.Method, metrics.SucStatus, start)
}

// RefreshSaasToken refresh usertoken for a saas user
func RefreshSaasToken(request *restful.Request, response *restful.Response) {
	start := time.Now()

	userName := request.PathParameter("user_name")
	user := sqlstore.GetUserByCondition(&models.BcsUser{Name: userName, UserType: sqlstore.SaasUser})
	if user == nil {
		metrics.ReportRequestAPIMetrics("RefreshSaasToken", request.Request.Method, metrics.ErrStatus, start)
		blog.Warnf("failed to refresh token, user [%s] not found in db", userName)
		message := fmt.Sprintf("errcode: %d, user with user_name=%s not found", common.BcsErrApiBadRequest, userName)
		utils.WriteNotFoundError(response, common.BcsErrApiBadRequest, message)
		return
	}

	// refresh the usertoken
	updatedUser := user
	updatedUser.UserToken = uniuri.NewLen(constant.DefaultTokenLength)
	updatedUser.ExpiresAt = time.Now().Add(sqlstore.AdminSaasUserExpiredTime)

	// update and save to db
	// if update failed, it's better to refresh by client
	err := sqlstore.UpdateUser(user, updatedUser)
	if err != nil {
		metrics.ReportRequestAPIMetrics("RefreshSaasToken", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("failed to refresh usertoken [%s]: %s", user.Name, err.Error())
		message := fmt.Sprintf("errcode: %d, failed to refresh usertoken [%s], error: %s", common.BcsErrApiInternalDbError, userName, err)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	data := utils.CreateResponseData(nil, "success", *user)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("RefreshSaasToken", request.Request.Method, metrics.SucStatus, start)
}
