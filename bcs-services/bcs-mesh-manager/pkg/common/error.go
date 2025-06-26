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

package common

import (
	"fmt"

	commErr "github.com/Tencent/bk-bcs/bcs-common/common"
)

// CodeMessageError 错误码和错误消息
type CodeMessageError struct {
	code    uint32
	message string
	err     error
}

// NewCodeMessageError 创建新的错误
func NewCodeMessageError(code uint32, message string, err error) *CodeMessageError {
	return &CodeMessageError{
		code:    code,
		message: message,
		err:     err,
	}
}

// Error 实现 error 接口
func (c *CodeMessageError) Error() string {
	return fmt.Sprintf("code: %d, message: %s, err: %s", c.code, c.message, c.err)
}

// GetCode 获取错误码
func (c *CodeMessageError) GetCode() uint32 {
	return c.code
}

// GetMessage 获取错误消息
func (c *CodeMessageError) GetMessage() string {
	return c.message
}

// GetErr 获取错误
func (c *CodeMessageError) GetErr() error {
	return c.err
}

// GetMessageWithErr 获取错误消息和错误
func (c *CodeMessageError) GetMessageWithErr() string {
	if c.err == nil {
		return c.message
	}
	return fmt.Sprintf("%s, err: %s", c.message, c.err)
}

// 系统相关错误
var (
	// SuccessCode 正常返回
	SuccessCode = uint32(0)

	// ParamErrorCode 参数校验失败
	ParamErrorCode = uint32(commErr.AdditionErrorCode + 400)
	// ParamError 参数校验失败
	ParamError = &CodeMessageError{
		code:    ParamErrorCode,
		message: "params error",
	}
	// InnerErrorCode 内部服务异常
	InnerErrorCode = uint32(commErr.AdditionErrorCode + 500)
	// InnerError 内部服务异常
	InnerError = &CodeMessageError{
		code:    InnerErrorCode,
		message: "inner error",
	}
	// DBErrorCode DB操作失败
	DBErrorCode = uint32(commErr.AdditionErrorCode + 501)
	// DBError DB操作失败
	DBError = &CodeMessageError{
		code:    DBErrorCode,
		message: "db error",
	}
	// UnauthErrorCode 未认证/认证失败
	UnauthErrorCode = uint32(commErr.AdditionErrorCode + 401)
	// UnauthError 未认证/认证失败
	UnauthError = &CodeMessageError{
		code:    UnauthErrorCode,
		message: "auth error",
	}
	// PermDeniedErrorCode 无权限
	PermDeniedErrorCode = uint32(commErr.AdditionErrorCode + 403)
	// PermDeniedError 无权限
	PermDeniedError = &CodeMessageError{
		code:    PermDeniedErrorCode,
		message: "no permission",
	}

	// NotFoundErrorCode 未找到
	NotFoundErrorCode = uint32(commErr.AdditionErrorCode + 404)
	// NotFoundError 未找到
	NotFoundError = &CodeMessageError{
		code:    NotFoundErrorCode,
		message: "not found",
	}

	// IstioInstallErrorCode istio安装失败
	IstioInstallErrorCode = uint32(commErr.AdditionErrorCode + 502)
	// IstioInstallError istio安装失败
	IstioInstallError = &CodeMessageError{
		code:    IstioInstallErrorCode,
		message: "istio install error",
	}

	// NamespaceExistErrorCode 命名空间已存在
	NamespaceExistErrorCode = uint32(commErr.AdditionErrorCode + 503)
	// NamespaceExistError 命名空间已存在
	NamespaceExistError = &CodeMessageError{
		code:    NamespaceExistErrorCode,
		message: "namespace already exists",
	}

	// InvalidRequestErrorCode 请求参数错误
	InvalidRequestErrorCode = uint32(commErr.AdditionErrorCode + 505)
	// InvalidRequestError 请求参数错误
	InvalidRequestError = &CodeMessageError{
		code:    InvalidRequestErrorCode,
		message: "invalid request",
	}
)

// 业务相关错误
var (
	// InstallIstioErrorCode 安装istio失败
	InstallIstioErrorCode = uint32(commErr.AdditionErrorCode + 506)
)
