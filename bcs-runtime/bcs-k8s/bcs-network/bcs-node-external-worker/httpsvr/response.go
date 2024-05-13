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

package httpsvr

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

var unknownError = errors.New("unknown") // nolint

// APIRespone response for api request
type APIRespone struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

// CreateResponseData common response
func CreateResponseData(err error, msg string, data string) []byte {
	var resp *APIRespone

	if err != nil {
		resp = errResponseDefault(err)
	} else {
		resp = responseData(msg, data)
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		blog.Errorf("marshal failed, resp: %+v, err: %s", resp, err.Error())
		return CreateResponseData(unknownError, "", "")
	}
	// blog.V(3).Infof("createRespone: %s", rpyErr.Error())

	return bytes
}

func responseData(msg string, data string) *APIRespone {
	return &APIRespone{
		Code:    http.StatusOK,
		Message: msg,
		Data:    data,
	}
}

func errResponseDefault(err error) *APIRespone {
	return errResponse(http.StatusInternalServerError, err)
}

func errResponse(code int, err error) *APIRespone {
	return &APIRespone{
		Code:    code,
		Message: err.Error(),
		Data:    "",
	}
}
