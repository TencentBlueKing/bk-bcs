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

package middleware

import (
	"context"

	restful "github.com/emicklei/go-restful/v3"
	"github.com/google/uuid"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/utils"
)

// RequestIDFilter set request id to context
func RequestIDFilter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	requestID := request.Request.Header.Get(string(utils.ContextValueKeyRequestID))
	if len(requestID) == 0 {
		requestID = uuid.New().String()
	}
	// 新增泳道特性
	ctx := utils.WithLaneIdCtx(request.Request.Context(), request.Request.Header)
	ctx = context.WithValue(ctx, utils.ContextValueKeyRequestID, requestID)
	request.Request = request.Request.WithContext(ctx)
	chain.ProcessFilter(request, response)
}
