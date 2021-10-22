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

package custom

import (
	"net/http"

	restful "github.com/emicklei/go-restful"
)

const (
	// RequestFailedCode starting code for error response
	RequestFailedCode = 4001
	// RequestSuccessCode code for success
	RequestSuccessCode = 0
)

// CustomHTTPResponse : custom http response
func CustomHTTPResponse(response *restful.Response, httpCode int, result bool, msg string, code int, data interface{}) {
	response.WriteHeaderAndEntity(httpCode, APIResponse{
		Result:  result,
		Message: msg,
		Code:    code,
		Data:    data,
	})
}

// CustomServerErrorResponse : custom http response for inner server error
func CustomServerErrorResponse(response *restful.Response, msg string) {
	CustomHTTPResponse(response, http.StatusInternalServerError, false, msg, RequestFailedCode, nil)
}

// CustomSuccessResponse : custom http response for success
func CustomSuccessResponse(response *restful.Response, msg string, data interface{}) {
	CustomHTTPResponse(response, http.StatusOK, true, msg, RequestSuccessCode, data)
}

// CustomSimpleHTTPResponse : simple http response
func CustomSimpleHTTPResponse(response *restful.Response, httpCode int, resp APIResponse) {
	response.WriteHeaderAndEntity(httpCode, resp)
}
