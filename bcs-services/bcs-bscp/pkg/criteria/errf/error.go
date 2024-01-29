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

package errf

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n/localizer"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// ErrorF defines an error with error code and message.
type ErrorF struct {
	// Kit is bscp kit
	Kit *kit.Kit
	// Code is bscp errCode
	Code int32 `json:"code"`
	// Message is error detail
	Message string `json:"message"`
}

// Errorf 返回自定义封装的bscp错误，包括错误码、错误信息
// bcs-services/bcs-bscp/pkg/rest/response.go中的错误中间件方法GRPCErr会统一进行错误码转换处理
// 需要返回给普通用户看的错误，统一使用该方法返回错误，且对错误信息进行国际化处理，便于普通用户理解
func Errorf(kit *kit.Kit, code int32, format string, args ...interface{}) *ErrorF {
	return &ErrorF{
		Kit:  kit,
		Code: code,
		// 错误信息国际化
		Message: localizer.Get(kit.Lang).Translate(format, args...),
	}
}

// Error implement the golang's basic error interface
func (e *ErrorF) Error() string {
	if e == nil || e.Code == OK {
		return "nil"
	}

	// return with a json format string error, so that the upper service
	// can use Wrap to decode it.
	return fmt.Sprintf(`{"code": %d, "message": "%s"}`, e.Code, e.Message)
}

// WithCause 打印根因错误，有底层错误需要暴露时调用该方法，便于研发排查问题
func (e *ErrorF) WithCause(cause error) *ErrorF {
	if cause == nil {
		return e
	}

	// 如果底层根因错误已经是bscp错误，直接使用该根因错误
	if c, ok := cause.(*ErrorF); ok {
		return c
	}
	// 打印其他错误根因日志
	logs.ErrorDepthf(1, "bscp inner err cause: %v, rid: %s", cause, e.Kit.Rid)

	return e
}

// GRPCStatus implements interface{ GRPCStatus() *Status } , so that it can be recognized by grpc
func (e *ErrorF) GRPCStatus() *status.Status {
	return status.New(codes.Code(e.Code), e.Message)
}

// Format the ErrorF error to a string format.
func (e *ErrorF) Format() string {
	if e == nil || e.Code == OK {
		return ""
	}

	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// AssignResp used only to assign the values of the Code and Message
// fields of ErrorF to the Code and Message fields of the response.
// Node: resp must be a *struct.

// New an error with error code and message.
func New(code int32, message string) error {
	return &ErrorF{Code: code, Message: message}
}

// Newf create an error with error code and formatted message.
func Newf(code int32, format string, args ...interface{}) error {
	return &ErrorF{Code: code, Message: fmt.Sprintf(format, args...)}
}

const grpcErrPrefix = "rpc error: code = Unknown desc = "

// Error try to convert the error to ErrorF if possible.
// it is used by the RPC client to wrap the response error response
// by the RPC server to the ErrorF, user can use this ErrorF to test
// if an error is returned or not, if yes, then use the ErrorF to
// response with error code and message.
func Error(err error) *ErrorF {
	if err == nil {
		return nil
	}

	ef, ok := err.(*ErrorF)
	if ok {
		// if this error is already is ErrorF, then return directly.
		ef.Message = strings.TrimPrefix(ef.Message, grpcErrPrefix)
		return ef
	}

	// try to parse this error to ErrorF with to kind of strategy if possible
	s := strings.TrimPrefix(strings.TrimSpace(err.Error()), grpcErrPrefix)

	// test if the error is a json error,
	// if not, then this is an error without error code.
	if !strings.HasPrefix(s, "{") {
		return &ErrorF{
			Code:    Unknown,
			Message: s,
		}
	}

	// this is a standard error format, then decode it directly.
	ef = new(ErrorF)
	if err := json.Unmarshal([]byte(s), ef); err != nil {
		return &ErrorF{
			Code:    Unknown,
			Message: s,
		}
	}

	return ef
}

// RPCAborted 通过 msg 构建通用 rpc 退出错误
func RPCAborted(format string, a ...interface{}) error {
	return status.Errorf(codes.Aborted, format, a...)
}

// RPCAbortedErr 通过 err 构建通用 rpc 退出错误
func RPCAbortedErr(err error) error {
	return status.Errorf(codes.Aborted, err.Error())
}

// PRCPermissionDenied 无权限错误
func PRCPermissionDenied() *status.Status {
	return status.New(codes.PermissionDenied, ErrPermissionDenied.Error())
}

// GetErrMsg 获取 err wrap 的 msg
func GetErrMsg(err error) string {
	// 去除后缀字符规则 https://github.com/pkg/errors/blob/master/errors.go#L244
	msg := strings.TrimSuffix(err.Error(), ": "+errors.Cause(err).Error())
	return msg
}
