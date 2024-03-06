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

package cloudcollector

import (
	"sync"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
)

// StatusCache health status cache
type StatusCache struct {
	cache map[string][]*cloud.BackendHealthStatus
	mutex sync.Mutex
}

// NewStatusCache create cache object for health status
func NewStatusCache() StatusCache {
	return StatusCache{
		cache: make(map[string][]*cloud.BackendHealthStatus),
	}
}

// UpdateCache update health status
func (sc *StatusCache) UpdateCache(newData map[string][]*cloud.BackendHealthStatus) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	// clear old data
	sc.cache = make(map[string][]*cloud.BackendHealthStatus)
	// update new data
	for k, v := range newData {
		sc.cache[k] = v
	}
}

// Get get health status from cache
func (sc *StatusCache) Get() map[string][]*cloud.BackendHealthStatus {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	return sc.cache
}
