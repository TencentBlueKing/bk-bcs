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

	"bscp.io/pkg/logs"
)

// ErrorF defines an error with error code and message.
type ErrorF struct {
	// Code is bscp errCode
	Code int32 `json:"code"`
	// Message is error detail
	Message string `json:"message"`
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

// Format the ErrorF error to a string format.
func (e *ErrorF) Format() string {
	if e == nil || e.Code == OK {
		return ""
	}

	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// GRPCStatus implements interface{ GRPCStatus() *Status } , so that it can be recognized by grpc
func (e *ErrorF) GRPCStatus() *status.Status {
	return status.New(codes.Code(e.Code), e.Message)
}

// BSCPErrI is bscp error interface
type BSCPErrI interface {
	BSCPErr() *ErrorF
}

// BSCPErr implements BSCPErrI, so that it can be recognized by bscp itself
func (e *ErrorF) BSCPErr() *ErrorF {
	return e
}

// Errorf 返回自定义封装的bscp错误，包括错误码、错误信息
// bcs-services/bcs-bscp/pkg/rest/response.go中的错误中间件方法GRPCErr会统一进行错误码转换处理
// 需要返回给普通用户看的错误，统一使用该方法返回错误，国际化也以此方法作为提取依据，便于普通用户理解
// 该方法会统一打印错误根因，便于研发排查问题
// 优先使用最底层的bscp错误，越底层的错误越能看出问题根因
func Errorf(err error, code int32, format string, args ...interface{}) error {
	if err == nil {
		err = fmt.Errorf(format, args...)
	}
	// 如果已经是bscp错误，直接返回底层的bscp错误
	if _, ok := err.(BSCPErrI); ok {
		return err
	}
	logs.ErrorDepthf(1, "bscp inner err cause: %v", err)
	return &ErrorF{Code: code, Message: fmt.Sprintf(format, args...)}
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
