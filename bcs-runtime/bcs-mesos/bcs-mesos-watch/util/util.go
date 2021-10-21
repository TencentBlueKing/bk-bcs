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
 *
 */

package util

//GetHashId get ID for dispatch channel
func GetHashId(s string, maxInt int) int {
	if maxInt <= 1 {
		return 0
	}

	seed := 131
	hash := 0
	char := []byte(s)

	for _, c := range char {
		hash = hash*seed + int(c)
	}

	return (hash & 0x7FFFFFFF) % maxInt
}
