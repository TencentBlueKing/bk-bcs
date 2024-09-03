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

// Package cache xxx
package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var cacheData *cache.Cache

// InitCache init cache data
func InitCache() {
	// Create a cache with a default expiration time of 10 minutes, and which
	// purges expired items every 1 hour
	cacheData = cache.New(5*time.Minute, 60*time.Minute)
}

// GetCache get cache data
func GetCache() *cache.Cache {
	return cacheData
}
