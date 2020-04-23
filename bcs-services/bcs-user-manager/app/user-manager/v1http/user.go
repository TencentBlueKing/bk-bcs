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
	"time"

	"bk-bcs/bcs-common/common"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-services/bcs-user-manager/app/metrics"
	"bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"bk-bcs/bcs-services/bcs-user-manager/app/user-manager/utils"
	"github.com/dchest/uniuri"
	"github.com/emicklei/go-restful"
)

const DefaultTokenLength = 32

func CreateAdminUser(request *restful.Request, response *restful.Response) {
	start := time.Now()

	userName := request.PathParameter("user_name")
	user := &models.BcsUser{
		Name:     userName,
		UserType: sqlstore.AdminUser,
	}
	userInDb := sqlstore.GetUserByCondition(user)
	if userInDb != nil {
		metrics.RequestErrorCount.WithLabelValues("user", request.Request.Method).Inc()
		metrics.RequestErrorLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
		message := fmt.Sprintf("errcode: %d, user [%s] already exist", common.BcsErrApiBadRequest, userName)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}

	user.UserToken = uniuri.NewLen(DefaultTokenLength)
	user.ExpiresAt = time.Now().Add(sqlstore.AdminSaasUserExpiredTime)

	err := sqlstore.CreateUser(user)
	if err != nil {
		metrics.RequestErrorCount.WithLabelValues("user", request.Request.Method).Inc()
		metrics.RequestErrorLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Errorf("failed to create user [%s]: %s", user.Name, err.Error())
		message := fmt.Sprintf("errcode: %d, creating user [%s] failed, error: %s", common.BcsErrApiInternalDbError, user.Name, err)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	data := utils.CreateResponeData(nil, "success", *user)
	_, _ = response.Write([]byte(data))

	metrics.RequestCount.WithLabelValues("user", request.Request.Method).Inc()
	metrics.RequestLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
}

func GetAdminUser(request *restful.Request, response *restful.Response) {
	start := time.Now()

	userName := request.PathParameter("user_name")
	user := sqlstore.GetUserByCondition(&models.BcsUser{Name: userName, UserType: sqlstore.AdminUser})
	if user == nil {
		metrics.RequestErrorCount.WithLabelValues("user", request.Request.Method).Inc()
		metrics.RequestErrorLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Warnf("user [%s] not found in db", userName)
		message := fmt.Sprintf("errcode: %d, user with user_name=%s not found", common.BcsErrApiBadRequest, userName)
		utils.WriteNotFoundError(response, common.BcsErrApiBadRequest, message)
		return
	}

	data := utils.CreateResponeData(nil, "success", *user)
	_, _ = response.Write([]byte(data))

	metrics.RequestCount.WithLabelValues("user", request.Request.Method).Inc()
	metrics.RequestLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
}

func CreateSaasUser(request *restful.Request, response *restful.Response) {
	start := time.Now()

	userName := request.PathParameter("user_name")
	user := &models.BcsUser{
		Name:     userName,
		UserType: sqlstore.SaasUser,
	}
	userInDb := sqlstore.GetUserByCondition(user)
	if userInDb != nil {
		metrics.RequestErrorCount.WithLabelValues("user", request.Request.Method).Inc()
		metrics.RequestErrorLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
		message := fmt.Sprintf("errcode: %d, user [%s] already exist", common.BcsErrApiBadRequest, userName)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}

	user.UserToken = uniuri.NewLen(DefaultTokenLength)
	user.ExpiresAt = time.Now().Add(sqlstore.AdminSaasUserExpiredTime)

	err := sqlstore.CreateUser(user)
	if err != nil {
		metrics.RequestErrorCount.WithLabelValues("user", request.Request.Method).Inc()
		metrics.RequestErrorLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Errorf("failed to create user [%s]: %s", user.Name, err.Error())
		message := fmt.Sprintf("errcode: %d, creating user [%s] failed, error: %s", common.BcsErrApiInternalDbError, user.Name, err)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	data := utils.CreateResponeData(nil, "success", *user)
	_, _ = response.Write([]byte(data))

	metrics.RequestCount.WithLabelValues("user", request.Request.Method).Inc()
	metrics.RequestLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
}

func GetSaasUser(request *restful.Request, response *restful.Response) {
	start := time.Now()

	userName := request.PathParameter("user_name")
	user := sqlstore.GetUserByCondition(&models.BcsUser{Name: userName, UserType: sqlstore.SaasUser})
	if user == nil {
		metrics.RequestErrorCount.WithLabelValues("user", request.Request.Method).Inc()
		metrics.RequestErrorLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Warnf("user [%s] not found in db", userName)
		message := fmt.Sprintf("errcode: %d, user with user_name=%s not found", common.BcsErrApiBadRequest, userName)
		utils.WriteNotFoundError(response, common.BcsErrApiBadRequest, message)
		return
	}

	data := utils.CreateResponeData(nil, "success", *user)
	_, _ = response.Write([]byte(data))

	metrics.RequestCount.WithLabelValues("user", request.Request.Method).Inc()
	metrics.RequestLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
}

func CreatePlainUser(request *restful.Request, response *restful.Response) {
	start := time.Now()

	userName := request.PathParameter("user_name")
	user := &models.BcsUser{
		Name:     userName,
		UserType: sqlstore.PlainUser,
	}
	userInDb := sqlstore.GetUserByCondition(user)
	if userInDb != nil {
		metrics.RequestErrorCount.WithLabelValues("user", request.Request.Method).Inc()
		metrics.RequestErrorLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
		message := fmt.Sprintf("errcode: %d, user [%s] already exist", common.BcsErrApiBadRequest, userName)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}

	user.UserToken = uniuri.NewLen(DefaultTokenLength)
	user.ExpiresAt = time.Now().Add(sqlstore.PlainUserExpiredTime)

	err := sqlstore.CreateUser(user)
	if err != nil {
		metrics.RequestErrorCount.WithLabelValues("user", request.Request.Method).Inc()
		metrics.RequestErrorLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Errorf("failed to create user [%s]: %s", user.Name, err.Error())
		message := fmt.Sprintf("errcode: %d, creating user [%s] failed, error: %s", common.BcsErrApiInternalDbError, user.Name, err)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	data := utils.CreateResponeData(nil, "success", *user)
	_, _ = response.Write([]byte(data))

	metrics.RequestCount.WithLabelValues("user", request.Request.Method).Inc()
	metrics.RequestLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
}

func GetPlainUser(request *restful.Request, response *restful.Response) {
	start := time.Now()

	userName := request.PathParameter("user_name")
	user := sqlstore.GetUserByCondition(&models.BcsUser{Name: userName, UserType: sqlstore.PlainUser})
	if user == nil {
		metrics.RequestErrorCount.WithLabelValues("user", request.Request.Method).Inc()
		metrics.RequestErrorLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Warnf("failed to get user, user [%s] not found in db", userName)
		message := fmt.Sprintf("errcode: %d, user with user_name=%s not found", common.BcsErrApiBadRequest, userName)
		utils.WriteNotFoundError(response, common.BcsErrApiBadRequest, message)
		return
	}

	data := utils.CreateResponeData(nil, "success", *user)
	_, _ = response.Write([]byte(data))

	metrics.RequestCount.WithLabelValues("user", request.Request.Method).Inc()
	metrics.RequestLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
}

func RefreshPlainToken(request *restful.Request, response *restful.Response) {
	start := time.Now()

	userName := request.PathParameter("user_name")
	user := sqlstore.GetUserByCondition(&models.BcsUser{Name: userName, UserType: sqlstore.PlainUser})
	if user == nil {
		metrics.RequestErrorCount.WithLabelValues("user", request.Request.Method).Inc()
		metrics.RequestErrorLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Warnf("failed to refresh token, user [%s] not found in db", userName)
		message := fmt.Sprintf("errcode: %d, user with user_name=%s not found", common.BcsErrApiBadRequest, userName)
		utils.WriteNotFoundError(response, common.BcsErrApiBadRequest, message)
		return
	}

	updatedUser := user
	if time.Now().After(user.ExpiresAt) {
		updatedUser.UserToken = uniuri.NewLen(DefaultTokenLength)
		updatedUser.ExpiresAt = time.Now().Add(sqlstore.PlainUserExpiredTime)
	} else {
		updatedUser.ExpiresAt = time.Now().Add(sqlstore.PlainUserExpiredTime)
	}

	err := sqlstore.UpdateUser(user, updatedUser)
	if err != nil {
		metrics.RequestErrorCount.WithLabelValues("user", request.Request.Method).Inc()
		metrics.RequestErrorLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Errorf("failed to refresh usertoken [%s]: %s", user.Name, err.Error())
		message := fmt.Sprintf("errcode: %d, failed to refresh usertoken [%s], error: %s", common.BcsErrApiInternalDbError, userName, err)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	data := utils.CreateResponeData(nil, "success", *user)
	_, _ = response.Write([]byte(data))

	metrics.RequestCount.WithLabelValues("user", request.Request.Method).Inc()
	metrics.RequestLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
}

func RefreshSaasToken(request *restful.Request, response *restful.Response) {
	start := time.Now()

	userName := request.PathParameter("user_name")
	user := sqlstore.GetUserByCondition(&models.BcsUser{Name: userName, UserType: sqlstore.SaasUser})
	if user == nil {
		metrics.RequestErrorCount.WithLabelValues("user", request.Request.Method).Inc()
		metrics.RequestErrorLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Warnf("failed to refresh token, user [%s] not found in db", userName)
		message := fmt.Sprintf("errcode: %d, user with user_name=%s not found", common.BcsErrApiBadRequest, userName)
		utils.WriteNotFoundError(response, common.BcsErrApiBadRequest, message)
		return
	}

	updatedUser := user
	updatedUser.UserToken = uniuri.NewLen(DefaultTokenLength)
	updatedUser.ExpiresAt = time.Now().Add(sqlstore.AdminSaasUserExpiredTime)

	err := sqlstore.UpdateUser(user, updatedUser)
	if err != nil {
		metrics.RequestErrorCount.WithLabelValues("user", request.Request.Method).Inc()
		metrics.RequestErrorLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Errorf("failed to refresh usertoken [%s]: %s", user.Name, err.Error())
		message := fmt.Sprintf("errcode: %d, failed to refresh usertoken [%s], error: %s", common.BcsErrApiInternalDbError, userName, err)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	data := utils.CreateResponeData(nil, "success", *user)
	_, _ = response.Write([]byte(data))

	metrics.RequestCount.WithLabelValues("user", request.Request.Method).Inc()
	metrics.RequestLatency.WithLabelValues("user", request.Request.Method).Observe(time.Since(start).Seconds())
}
