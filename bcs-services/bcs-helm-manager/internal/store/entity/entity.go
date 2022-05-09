/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package entity

// M used for incremental update
type M map[string]interface{}

// GetString return string value of interface
func (m M) GetString(key string) string {
	val, ok := m[key]
	if ok {
		if res, ok := val.(string); ok {
			return res
		}
	}
	return ""
}

// Update k-v
func (m M) Update(k string, v interface{}) M {
	r := m
	if r == nil {
		r = make(M)
	}

	r[k] = v
	return r
}

// Updates k-v pairs
func (m M) Updates(um M) M {
	r := m
	if r == nil {
		r = make(M)
	}

	for k, v := range um {
		r[k] = v
	}
	return r
}
