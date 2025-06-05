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

// Package util xxx
package util

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var (
	DefaultCache = cache.New(20*time.Minute, 20*time.Minute)
)

// GetCache get cache interface struct by key
func GetCache(key string) (interface{}, bool) {
	result, exist := DefaultCache.Get(key)
	return result, exist
}

// SetCache set cache by key
func SetCache(key string, value interface{}) {
	DefaultCache.Set(key, value, time.Hour)
}

// SetCacheWithTimeout set cache by key with timeout
func SetCacheWithTimeout(key string, value interface{}, duration time.Duration) {
	DefaultCache.Set(key, value, duration)
}
