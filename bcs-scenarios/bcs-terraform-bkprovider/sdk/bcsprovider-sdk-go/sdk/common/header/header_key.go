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

// Package header bcs-header
package header

import (
	"strings"

	"github.com/google/uuid"
)

const (
	// RequestIDKey 请求ID
	RequestIDKey = "X-Request-Id"

	// UsernameKey 请求用户
	UsernameKey = "X-Project-Username"

	// Authorization 身份
	Authorization = "Authorization"
)

// GenUUID 生成请求ID，长度为32
func GenUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}
