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
	"math/rand"

	"github.com/google/uuid"
)

// GenUUID generate a uuid
func GenUUID() string {
	return uuid.New().String()
}

// GenMeshID 生成mesh id
// 格式：mesh-bcs-xxxx
func GenMeshID() string {
	meshID := "bcs-mesh-"
	meshID += RandString(8)
	return meshID
}

// GenNetworkID 生成network id
// 格式：network-bcs-xxxx
func GenNetworkID() string {
	networkID := "bcs-network-"
	networkID += RandString(8)
	return networkID
}

// RandString 随机生成n位字符串（数字+小写字母）
// #nosec G404 -- RandString 仅用于非安全场景
func RandString(n int) string {
	letters := "0123456789abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
