/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package errorx

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
)

// BaseError ClusterResources 模块基础 Error
type BaseError struct {
	code int
	err  error
}

// Error ...
func (e *BaseError) Error() string {
	return e.err.Error()
}

// Code ...
func (e *BaseError) Code() int {
	return e.code
}

// New ...
func New(code int, msg string, vars ...interface{}) error {
	// 若没有指定错误码，则使用默认错误码
	if code == 0 {
		code = errcode.DefaultErrCode
	}
	return &BaseError{code: code, err: fmt.Errorf(msg, vars...)}
}
