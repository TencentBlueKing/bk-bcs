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

package utils

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/server/types"

	"github.com/emicklei/go-restful"
)

// WriteFuncFactory builds WriteXXX shortcut functions
func WriteFuncFactory(statusCode int) func(response *restful.Response, codeName, message string) {
	return func(response *restful.Response, codeName, message string) {
		response.WriteHeaderAndEntity(statusCode, types.ErrorResponse{
			CodeName: codeName,
			Message:  message,
		})
	}
}

// WriteClientError xxx
var WriteClientError = WriteFuncFactory(400)

// WriteUnauthorizedError xxx
var WriteUnauthorizedError = WriteFuncFactory(401)

// WriteForbiddenError xxx
var WriteForbiddenError = WriteFuncFactory(403)

// WriteNotFoundError xxx
var WriteNotFoundError = WriteFuncFactory(404)

// WriteServerError xxx
var WriteServerError = WriteFuncFactory(500)
