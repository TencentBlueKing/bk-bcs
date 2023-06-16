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
	"github.com/hashicorp/go-multierror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"bscp.io/pkg/logs"
)

var (
	// grpcCodeMap 蓝鲸 Code 映射
	grpcCodeMap = map[codes.Code]string{
		codes.Canceled:           "CANCELLED",
		codes.Unknown:            "UNKNOWN",
		codes.InvalidArgument:    "INVALID_ARGUMENT",
		codes.DeadlineExceeded:   "DEADLINE_EXCEEDED",
		codes.NotFound:           "NOT_FOUND",
		codes.AlreadyExists:      "ALREADY_EXISTS",
		codes.PermissionDenied:   "PERMISSION_DENIED",
		codes.ResourceExhausted:  "RESOURCE_EXHAUSTED",
		codes.FailedPrecondition: "FAILED_PRECONDITION",
		codes.Aborted:            "ABORTED",
		codes.OutOfRange:         "OUT_OF_RANGE",
		codes.Unimplemented:      "UNIMPLEMENTED",
		codes.Internal:           "INTERNAL",
		codes.Unavailable:        "UNAVAILABLE",
		codes.DataLoss:           "DATA_LOSS",
		codes.Unauthenticated:    "UNAUTHENTICATED",
	}

	// grpcCodeMap 蓝鲸 status 映射
	grpcHttpStatusMap = map[codes.Code]int{
		codes.Canceled:           http.StatusBadRequest,
		codes.Unknown:            http.StatusBadRequest,
		codes.InvalidArgument:    http.StatusBadRequest,
		codes.DeadlineExceeded:   http.StatusBadRequest,
		codes.NotFound:           http.StatusNotFound,
		codes.AlreadyExists:      http.StatusBadRequest,
		codes.PermissionDenied:   http.StatusForbidden,
		codes.ResourceExhausted:  http.StatusBadRequest,
		codes.FailedPrecondition: http.StatusBadRequest,
		codes.Aborted:            http.StatusBadRequest,
		codes.OutOfRange:         http.StatusBadRequest,
		codes.Unimplemented:      http.StatusBadRequest,
		codes.Internal:           http.StatusBadRequest,
		codes.Unavailable:        http.StatusBadRequest,
		codes.DataLoss:           http.StatusBadRequest,
		codes.Unauthenticated:    http.StatusUnauthorized,
	}
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
	loginURL       string
	loginPlainURL  string
}

// Render go-chi/render Renderer interface implement
func (res *ErrorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	statusCode := res.HTTPStatusCode
	if statusCode == 0 {
		statusCode = http.StatusBadRequest
	}

	switch res.Error.Code {
	case "UNAUTHENTICATED":
		res.Error.Data = &UnauthorizedData{
			LoginURL:      res.loginURL,
			LoginPlainURL: res.loginPlainURL,
		}

	case "PERMISSION_DENIED":
		// 把 detail 中拿出来做鉴权详情
		if len(res.Error.Details) > 0 {
			res.Error.Data = res.Error.Details[0]
		}
		res.Error.Details = []interface{}{}
	}

	render.Status(r, statusCode)
	return nil
}

// UnauthorizedErr rest 未登入返回
func UnauthorizedErr(err error, loginAuthHost, loginAuthPlainHost string) render.Renderer {
	payload := &ErrorPayload{Code: "UNAUTHENTICATED", Message: err.Error(), Details: []interface{}{}}
	if e, ok := err.(*multierror.Error); ok {
		for _, v := range e.Errors {
			payload.Details = append(payload.Details, v.Error())
		}
		payload.Message = "user not logged in"
	}

	return &ErrorResponse{Error: payload, HTTPStatusCode: http.StatusUnauthorized, loginURL: loginAuthHost, loginPlainURL: loginAuthPlainHost}
}

// PermissionDenied 无数据返回
func PermissionDenied(err error, data interface{}) render.Renderer {
	payload := &ErrorPayload{Code: "PERMISSION_DENIED", Message: err.Error(), Data: data}
	return &ErrorResponse{Error: payload, HTTPStatusCode: http.StatusForbidden}
}

// BadRequest rest 通用错误请求
func BadRequest(err error) render.Renderer {
	payload := &ErrorPayload{Code: "INVALID_REQUEST", Message: err.Error()}
	return &ErrorResponse{Error: payload, HTTPStatusCode: http.StatusBadRequest}
}

// GRPCErr GRPC-Gateway 错误
func GRPCErr(err error) render.Renderer {
	s := status.Convert(err)
	code := grpcCodeMap[s.Code()]
	if code == "" {
		code = "INVALID_REQUEST"
	}

	status := grpcHttpStatusMap[s.Code()]
	if status == 0 {
		status = http.StatusBadRequest
	}

	payload := &ErrorPayload{Code: code, Message: s.Message(), Details: s.Details()}
	return &ErrorResponse{Error: payload, HTTPStatusCode: status}
}
