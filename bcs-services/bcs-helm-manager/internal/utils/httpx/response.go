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

package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"

	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
)

// BaseResponse base response
type BaseResponse struct {
	Code           int                        `json:"code"`
	Message        string                     `json:"message"`
	RequestID      string                     `json:"requestID"`
	Data           interface{}                `json:"data"`
	WebAnnotations *authutils.PermDeniedError `json:"web_annotations,omitempty"`
}

// ResponseAuthError response auth error
func ResponseAuthError(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusUnauthorized)
	returnJSON(w, BaseResponse{Code: http.StatusUnauthorized, Message: err.Error(), RequestID: getRequestID(r)})
}

// ResponsePermissionError response permission error
func ResponsePermissionError(w http.ResponseWriter, r *http.Request, err *authutils.PermDeniedError) {
	w.WriteHeader(http.StatusUnauthorized)
	returnJSON(w, BaseResponse{
		Code: http.StatusUnauthorized, Message: err.Error(), RequestID: getRequestID(r),
		WebAnnotations: err,
	})
}

// ResponseSystemError response system error
func ResponseSystemError(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	returnJSON(w, BaseResponse{Code: http.StatusInternalServerError, Message: err.Error(),
		RequestID: getRequestID(r)})
}

// ResponseOK response ok
func ResponseOK(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.WriteHeader(http.StatusOK)
	returnJSON(w, BaseResponse{Code: 0, Message: "success", Data: data, RequestID: getRequestID(r)})
}

func returnJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	result, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err.Error()) // nolint:errcheck
		return
	}
	fmt.Fprintln(w, string(result)) // nolint:errcheck
}
