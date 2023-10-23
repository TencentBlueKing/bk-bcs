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

// Package rest xxx
package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"
)

// Result 返回的标准结构
type Result struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	RequestId string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// APIError 错误返回，兼容国际化
func APIError(c *gin.Context, err string) {
	requestId := route.MustGetAuthContext(c).RequestId
	result := Result{Code: types.ApiErrorCode, Message: err, RequestId: requestId}
	c.AbortWithStatusJSON(http.StatusBadRequest, result)
}

// APIOK 正常返回
func APIOK(c *gin.Context, message string, data interface{}) {
	requestId := route.MustGetAuthContext(c).RequestId
	result := Result{Code: types.NoError, Message: message, RequestId: requestId, Data: data}
	c.AbortWithStatusJSON(http.StatusOK, result)
}
