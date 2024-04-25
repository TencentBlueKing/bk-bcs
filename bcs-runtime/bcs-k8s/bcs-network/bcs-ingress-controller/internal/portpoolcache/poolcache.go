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

package portpoolcache

import (
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// Cache for ports in port pools
type Cache struct {
	sync.Mutex
	portPoolMap map[string]*CachePool
}

// NewCache create new cache
func NewCache() *Cache {
	return &Cache{
		portPoolMap: make(map[string]*CachePool),
	}
}

// Start start report metric
func (c *Cache) Start() {
	ticker := time.NewTicker(metricCollectInterval)
	for {
		select {
		case <-ticker.C:
			portPoolCapacityMetric.Reset()
			portPoolAllocatedMetric.Reset()
			blog.V(4).Infof("pool cache info: %s", common.ToJsonString(c.portPoolMap))
			for poolKey, pool := range c.portPoolMap {
				for _, item := range pool.ItemList {
					for protocol, list := range item.PortListMap {
						portPoolCapacityMetric.WithLabelValues(poolKey, item.ItemStatus.ItemName, protocol).
							Set(float64(list.GetAvailabePortNum()))
						portPoolAllocatedMetric.WithLabelValues(poolKey, item.ItemStatus.ItemName, protocol).
							Set(float64(list.GetAllocatedPortNum()))
					}
				}
			}
		}
	}
}

// IsItemExisted if an item with certain key in port pool
func (c *Cache) IsItemExisted(poolKey, poolItemKey string) bool {
	for _, item := range c.portPoolMap {
		if item.GetKey() == poolKey && item.HasItem(poolItemKey) {
			return true
		}
	}
	return false
}

// AddPortPoolItem add port pool item to port pool
func (c *Cache) AddPortPoolItem(poolKey string, allocatePolicy string, itemStatus *networkextensionv1.
	PortPoolItemStatus) error {
	// if itemStatus.Status != constant.PortPoolItemStatusReady {
	// 	return fmt.Errorf("item %s in pool %s is not ready, cannot add to cache", itemStatus.GetKey(), poolKey)
	// }
	if _, ok := c.portPoolMap[poolKey]; !ok {
		c.portPoolMap[poolKey] = NewCachePool(poolKey, allocatePolicy)
	}
	if c.portPoolMap[poolKey].HasItem(itemStatus.GetKey()) {
		return fmt.Errorf("item %s in pool %s already exists", itemStatus.GetKey(), poolKey)
	}
	if err := c.portPoolMap[poolKey].AddPoolItem(itemStatus); err != nil {
		return fmt.Errorf("add pool item failed, err %s", err.Error())
	}
	return nil
}

// DeletePortPoolItem delete port pool item from pool
func (c *Cache) DeletePortPoolItem(poolKey, poolItemKey string) {
	pool, ok := c.portPoolMap[poolKey]
	if !ok {
		return
	}
	if !pool.HasItem(poolItemKey) {
		return
	}
	pool.DeletePoolItem(poolItemKey)
	if len(pool.ItemList) == 0 {
		delete(c.portPoolMap, poolKey)
	}
}

// SetPortPoolItemStatus update status of item in port pool
func (c *Cache) SetPortPoolItemStatus(poolKey string, itemStatus *networkextensionv1.PortPoolItemStatus) error {
	pool, ok := c.portPoolMap[poolKey]
	if !ok {
		return fmt.Errorf("pool %s not found in cache", poolKey)
	}
	return pool.SetItemStatus(itemStatus)
}

// AllocatePortBinding allocate port from pool for protocol
func (c *Cache) AllocatePortBinding(poolKey, protocol, itemName string) (
	*networkextensionv1.PortPoolItemStatus, AllocatedPortItem, error) {
	pool, ok := c.portPoolMap[poolKey]
	if !ok {
		return nil, AllocatedPortItem{}, fmt.Errorf("pool %s not found in cache", poolKey)
	}
	return pool.AllocatePortBinding(protocol, itemName)
}

// AllocateAllProtocolPortBinding allocate ports with all protocols,  if itemName is set,
// // only allocate port from the specified item
func (c *Cache) AllocateAllProtocolPortBinding(poolKey, itemName string) (
	*networkextensionv1.PortPoolItemStatus, map[string]AllocatedPortItem, error) {
	pool, ok := c.portPoolMap[poolKey]
	if !ok {
		return nil, nil, fmt.Errorf("pool %s not found in cache", poolKey)
	}
	return pool.AllocateAllProtocolPortBinding(itemName)
}

// ReleasePortBinding release port binding
func (c *Cache) ReleasePortBinding(poolKey, poolItemKey, protocol string, startPort, endPort int) {
	pool, ok := c.portPoolMap[poolKey]
	if !ok {
		return
	}
	pool.ReleasePortBinding(poolItemKey, protocol, startPort, endPort)
}

// SetPortBindingUsed set certain port status to used
func (c *Cache) SetPortBindingUsed(poolKey, poolItemKey, protocol string, startPort, endPort int) {
	pool, ok := c.portPoolMap[poolKey]
	if !ok {
		return
	}
	pool.SetPortBindingUsed(poolItemKey, protocol, startPort, endPort)
}
