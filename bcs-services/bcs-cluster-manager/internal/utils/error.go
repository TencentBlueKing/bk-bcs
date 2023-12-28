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

package utils

import (
	"strings"
)

// MultiError aggregate error
type MultiError struct {
	errors []error
}

// NewMultiError init multiError client
func NewMultiError() *MultiError {
	return &MultiError{errors: make([]error, 0)}
}

// Error 实现了 error 接口，将多个错误格式化为一个字符串
func (m *MultiError) Error() string {
	errorStrings := make([]string, len(m.errors))

	for i, err := range m.errors {
		errorStrings[i] = err.Error()
	}

	return strings.Join(errorStrings, "; ")
}

// Append 添加一个错误到 MultiError 中
func (m *MultiError) Append(err error) {
	if err != nil {
		m.errors = append(m.errors, err)
	}
}

// HasErrors 检查 MultiError 是否包含错误
func (m *MultiError) HasErrors() bool {
	return len(m.errors) > 0
}
