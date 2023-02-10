/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"bscp.io/pkg/logs"
)

// BaseResp http response.
type BaseResp struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

// Response is a http standard response
type Response struct {
	Err            error       `json:"-"` // low-level runtime error
	HTTPStatusCode int         `json:"-"` // http response status code
	Code           int32       `json:"code"`
	Message        string      `json:"message"`
	Data           interface{} `json:"data"`
}

// NewBaseResp new BaseResp.
func NewBaseResp(code int32, msg string) *BaseResp {
	return &BaseResp{
		Code:    code,
		Message: msg,
	}
}

// WriteResp writer response to http.ResponseWriter.
func WriteResp(w http.ResponseWriter, resp interface{}) {
	bytes, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		logs.ErrorDepthf(1, "response marshal failed, err: %v", err)
		return
	}

	_, err = fmt.Fprintf(w, string(bytes))
	if err != nil {
		logs.ErrorDepthf(1, "write resp to ResponseWriter failed, err: %v", err)
		return
	}

	return
}

// Render chi Render 实现
func (res *Response) Render(w http.ResponseWriter, r *http.Request) error {
	statusCode := res.HTTPStatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	render.Status(r, statusCode)
	return nil
}

// UnauthorizedErrRender 未登入返回
func UnauthorizedErrRender(err error) render.Renderer {
	return &Response{Code: 401, Message: err.Error(), HTTPStatusCode: http.StatusUnauthorized}
}

// OKRender 正常返回
func OKRender(data interface{}) render.Renderer {
	return &Response{Data: data}
}
