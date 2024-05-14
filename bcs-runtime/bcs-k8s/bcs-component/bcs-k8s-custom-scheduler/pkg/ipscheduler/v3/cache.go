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

package v3

import (
	"sync"
)

// PoolCache cache of pools
type PoolCache struct {
	lock        sync.Mutex
	ipNumberMap map[string]int
}

// NewPoolCache creates new pool cache
func NewPoolCache() *PoolCache {
	return &PoolCache{
		lock:        sync.Mutex{},
		ipNumberMap: make(map[string]int),
	}
}

// UpdatePool update pool ip number
func (pc *PoolCache) UpdatePool(poolName string, ipNum int) {
	pc.lock.Lock()
	defer pc.lock.Unlock()
	pc.ipNumberMap[poolName] = ipNum
}

// AssumeOne assume pool ip number
func (pc *PoolCache) AssumeOne(poolName string) {
	pc.lock.Lock()
	defer pc.lock.Unlock()
	if ipNum, ok := pc.ipNumberMap[poolName]; ok {
		pc.ipNumberMap[poolName] = ipNum - 1
	}
}

// GetAvailablePoolNameList get available pool name list
func (pc *PoolCache) GetAvailablePoolNameList() []string {
	pc.lock.Lock()
	defer pc.lock.Unlock()
	var retList []string
	for p, num := range pc.ipNumberMap {
		if num > 0 {
			retList = append(retList, p)
		}
	}
	return retList
}

// DeletePool delete pool
func (pc *PoolCache) DeletePool(poolName string) {
	pc.lock.Lock()
	defer pc.lock.Unlock()
	delete(pc.ipNumberMap, poolName)
}
