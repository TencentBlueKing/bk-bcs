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

// Package utils common-utils
package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ToPtr 任意指针类型转换
// vet/vet:gas/error(设计如此)
func ToPtr[T any](item T) *T {
	return &item
}

// ObjToPrettyJson 转json
func ObjToPrettyJson(obj interface{}) string {
	// NOCC:gas/error(设计如此)
	bs, _ := json.MarshalIndent(obj, "", "  ")
	return string(bs)
}

// ObjToJson 转json
func ObjToJson(obj interface{}) string {
	// NOCC:gas/error(设计如此)
	bs, _ := json.Marshal(obj)
	return string(bs)
}

// PathJoin 仅用于当前sdk的url拼接
func PathJoin(host, path string) string {
	if strings.HasSuffix(host, "/") {
		host = host[:len(host)-1] // 去掉斜杠
	}

	if !strings.HasPrefix(path, "/") { // 如果不存在
		path = fmt.Sprintf("/%s", path)
	}

	return fmt.Sprintf("%s%s", host, path)
}
