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

// Package hash xxx
package hash

import (
	"crypto/md5" // NOCC:gas/crypto(误报)
	"encoding/hex"
)

// MD5Digest 字符串转 MD5
func MD5Digest(key string) string {
	// NOCC:gas/crypto(误报)
	hash := md5.New()
	_, err := hash.Write([]byte(key))
	if err != nil {
		return ""
	}
	return hex.EncodeToString(hash.Sum(nil))
}
