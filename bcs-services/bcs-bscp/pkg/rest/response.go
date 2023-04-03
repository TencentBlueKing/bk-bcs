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

	"bscp.io/pkg/cc"
	"bscp.io/pkg/components/bkpaas"
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

// OKResponse is a http standard response
type OKResponse struct {
	Err            error       `json:"-"` // low-level runtime error
	HTTPStatusCode int         `json:"-"` // http response status code
	Data           interface{} `json:"data"`
}

// Render chi Render 实现
func (res *OKResponse) Render(w http.ResponseWriter, r *http.Request) error {
	statusCode := res.HTTPStatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	render.Status(r, statusCode)
	return nil
}

// OKRender 正常返回
func OKRender(data interface{}) render.Renderer {
	return &OKResponse{Data: data}
}

// UnauthorizedData 登入错误返回
type UnauthorizedData struct {
	LoginURL      string `json:"login_url"`
	LoginPlainURL string `json:"login_plain_url"`
}

// PermissionData 没有权限返回
type PermissionData struct {
	System     string   `json:"system"`
	SystemName string   `json:"system_name"`
	Action     []string `json:"action"`
}

// ErrorPayload 错误详情
type ErrorPayload struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Data    interface{}   `json:"data"`
	Details []interface{} `json:"details"`
}

// Error 实现error接口
func (ep ErrorPayload) Error() string {
	return fmt.Sprintf("error code:%s, message:%s", ep.Code, ep.Message)
}

// ErrorResponse 错误返回
type ErrorResponse struct {
	Err            error         `json:"-"` // low-level runtime error
	HTTPStatusCode int           `json:"-"` // http response status code
	Error          *ErrorPayload `json:"error"`
}

// Render
func (res *ErrorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	statusCode := res.HTTPStatusCode
	if statusCode == 0 {
		statusCode = http.StatusBadRequest
	}

	if res.Error.Code == "UNAUTHENTICATED" {
		res.Error.Data = &UnauthorizedData{
			LoginURL:      bkpaas.BuildLoginURL(r, cc.ApiServer().LoginAuth.Host),
			LoginPlainURL: bkpaas.BuildLoginPlainURL(r, cc.ApiServer().LoginAuth.Host),
		}
	}

	res.Error.Details = []interface{}{}
	render.Status(r, statusCode)
	return nil
}

// UnauthorizedErr rest 未登入返回
func UnauthorizedErr(err error) render.Renderer {
	payload := &ErrorPayload{Code: "UNAUTHENTICATED", Message: err.Error()}
	return &ErrorResponse{Error: payload, HTTPStatusCode: http.StatusUnauthorized}
}

// PermissionDenied 无数据返回
func PermissionDenied(err error) render.Renderer {
	payload := &ErrorPayload{Code: "PERMISSION_DENIED", Message: err.Error()}
	return &ErrorResponse{Error: payload, HTTPStatusCode: http.StatusForbidden}
}

// BadRequest rest 通用错误请求
func BadRequest(err error) render.Renderer {
	payload := &ErrorPayload{Code: "INVALID_REQUEST", Message: err.Error()}
	return &ErrorResponse{Error: payload, HTTPStatusCode: http.StatusBadRequest}
}
