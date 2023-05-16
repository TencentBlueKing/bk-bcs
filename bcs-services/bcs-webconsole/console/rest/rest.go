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

// Package rest xxx
package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Result 返回的标准结构
type Result struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	RequestId string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// AbortWithBadRequestError 错误返回
func AbortWithBadRequestError(c *gin.Context, err error) {
	result := Result{Code: 1400, Message: err.Error(), RequestId: ""}
	c.AbortWithStatusJSON(http.StatusBadRequest, result)
}
