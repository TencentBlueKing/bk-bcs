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

package resthdrs

import (
	"fmt"
	"reflect"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/metric"
	m "github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"
	"github.com/emicklei/go-restful"
	"github.com/iancoleman/strcase"
	"strconv"
	"time"
)

const (
	PlainBCSUserType = "plain"
)

// CreateBCSUserForm
type BCSUserForm struct {
	UserName string `json:"user_name" validate:"required"`
}

func CreateUser(request *restful.Request, response *restful.Response) {

	start := time.Now()

	blog.Debug("CreateBCSUser begin")
	form := BCSUserForm{}
	request.ReadEntity(&form)

	err := validate.Struct(&form)
	if err != nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Debug(fmt.Sprintf("CreateBCSUser form validate failed, %s", err))
		response.WriteEntity(FormatValidationError(err))
		return
	}

	user := &m.User{
		Name:        fmt.Sprintf("%s:%s", PlainBCSUserType, form.UserName),
		IsSuperUser: false,
	}
	// Query if user already exists
	userInDb := sqlstore.GetUserByCondition(&m.User{Name: user.Name})
	if userInDb != nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		message := fmt.Sprintf("errcode: %d, create failed, user with this username already exists", common.BcsErrApiBadRequest)
		WriteClientError(response, "USER_ALREADY_EXISTS", message)
		return
	}

	err = sqlstore.CreateUser(user)
	errorCode := strcase.ToScreamingSnake(fmt.Sprint(reflect.TypeOf(user)))
	if err != nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		message := fmt.Sprintf("errcode: %d, can not create user, error: %s", common.BcsErrApiInternalDbError, err)
		WriteServerError(response, errorCode, message)
		return
	}

	response.WriteEntity(*user)

	metric.RequestCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
	metric.RequestLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
}

func QueryBCSUserByName(request *restful.Request, response *restful.Response) {

	start := time.Now()

	userName := request.PathParameter("user_name")

	// get user
	user := sqlstore.GetUserByCondition(&m.User{Name: fmt.Sprintf("%s:%s", PlainBCSUserType, userName)})
	if user == nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		message := fmt.Sprintf("errcode: %d, user with user_name=%s not found", common.BcsErrApiBadRequest, userName)
		blog.Warnf(message)
		WriteNotFoundError(response, "USER_NOT_FOUND", message)
		return
	}

	response.WriteEntity(*user)

	metric.RequestCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
	metric.RequestLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
}

func CreateUserToken(request *restful.Request, response *restful.Response) {

	start := time.Now()

	userID := request.PathParameter("user_id")
	idStr, err := strconv.Atoi(userID)
	if err != nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		message := fmt.Sprintf("errcode: %d, error parsing user_id to uint: %s", common.BcsErrApiBadRequest, err.Error())
		blog.Warnf(message)
		WriteClientError(response, "USER_ID_INVALID", message)
		return
	}
	// get user
	user := sqlstore.GetUserByCondition(&m.User{ID: uint(idStr)})
	if user == nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		message := fmt.Sprintf("errcode: %d, user with user_id=%d not found", common.BcsErrApiBadRequest, idStr)
		blog.Warnf(message)
		WriteNotFoundError(response, "USER_NOT_FOUND", message)
		return
	}

	// Create a user token if not exists
	userToken, err := sqlstore.GetOrCreateUserToken(user, m.UserTokenTypeKubeConfigPlain, "")
	if err != nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Warnf("Unable to create user token of type UserTokenTypeKubeConfigPlain for user %s: %s", user.Name, err.Error())
		message := fmt.Sprintf("errcode: %d, can not create user token: %s", common.BcsErrApiInternalDbError, err.Error())
		WriteServerError(response, "CANNOT_CREATE_USER_RTOKEN", message)
		return
	}

	response.WriteEntity(*userToken)

	metric.RequestCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
	metric.RequestLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
}
