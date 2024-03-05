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

// Package portpoolcache 维护端口池中端口的使用情况, 新增Pod时从该缓存中分配端口使用
package portpoolcache

import (
	"fmt"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// CachePool pool of ports
type CachePool struct {
	PoolKey  string
	ItemList []*CachePoolItem
}

// NewCachePool create cache pool
func NewCachePool(key string) *CachePool {
	return &CachePool{
		PoolKey: key,
	}
}

// GetKey get key of caced pool
func (cp *CachePool) GetKey() string {
	return cp.PoolKey
}

// HasItem see if a item with certain key exists in pool cache
func (cp *CachePool) HasItem(poolItemKey string) bool {
	for _, item := range cp.ItemList {
		if item.GetKey() == poolItemKey {
			return true
		}
	}
	return false
}

// AddPoolItem add pool item into port pool cache
func (cp *CachePool) AddPoolItem(itemStatus *networkextensionv1.PortPoolItemStatus) error {
	if itemStatus == nil {
		return fmt.Errorf("item status cannot be empty")
	}
	if cp.HasItem(itemStatus.GetKey()) {
		return fmt.Errorf("item %s of pool %s already exists", itemStatus.GetKey(), cp.GetKey())
	}
	poolItem, err := NewCachePoolItem(cp.GetKey(), itemStatus)
	if err != nil {
		return fmt.Errorf("new cache pool item failed, err %s", err.Error())
	}
	cp.ItemList = append(cp.ItemList, poolItem)
	return nil
}

// DeletePoolItem delete item from cache
func (cp *CachePool) DeletePoolItem(poolItemKey string) {
	for index, item := range cp.ItemList {
		if item.GetKey() == poolItemKey {
			cp.ItemList = append(cp.ItemList[0:index], cp.ItemList[index+1:]...)
			break
		}
	}
}

// SetItemStatus update item status in port pool cache
func (cp *CachePool) SetItemStatus(itemStatus *networkextensionv1.PortPoolItemStatus) error {
	poolItemKey := itemStatus.GetKey()
	for _, item := range cp.ItemList {
		if item.GetKey() == poolItemKey {
			if item.Status != itemStatus.Status {
				item.SetStatus(itemStatus.Status)
			}
			if item.ItemStatus == nil {
				item.SetItemsStatus(itemStatus)
				return nil
			}
			if itemStatus.EndPort > item.ItemStatus.EndPort {
				if err := item.IncreaseEndPort(int(itemStatus.EndPort)); err != nil {
					return err
				}
			}
			item.SetItemsStatus(itemStatus)
			return nil
		}
	}
	return fmt.Errorf("item %s in pool %s not found", poolItemKey, cp.PoolKey)
}

// AllocatePortBinding allocate port by protocol
func (cp *CachePool) AllocatePortBinding(protocol, itemName string) (
	*networkextensionv1.PortPoolItemStatus, AllocatedPortItem, error) {
	for _, item := range cp.ItemList {
		if itemName != "" && item.ItemStatus.ItemName != itemName {
			continue
		}
		retPort := item.Allocate(protocol)
		if retPort != nil {
			return item.ItemStatus, AllocatedPortItem{
				PoolKey:     cp.PoolKey,
				PoolItemKey: item.ItemStatus.GetKey(),
				Protocol:    protocol,
				StartPort:   retPort.StartPort,
				EndPort:     retPort.EndPort,
				IsUsed:      retPort.Used,
			}, nil
		}
	}
	return nil, AllocatedPortItem{}, fmt.Errorf("no available port in pool %s for protocol %s", cp.PoolKey, protocol)
}

// AllocateAllProtocolPortBinding allocate ports with all protocols, if itemName is set,
// only allocate port from the specified item
func (cp *CachePool) AllocateAllProtocolPortBinding(itemName string) (
	*networkextensionv1.PortPoolItemStatus, map[string]AllocatedPortItem, error) {
	for _, item := range cp.ItemList {
		if itemName != "" && item.ItemStatus.ItemName != itemName {
			continue
		}
		retPortMap := item.AllocateAllProtocolPort()
		if retPortMap == nil {
			continue
		}
		retMap := make(map[string]AllocatedPortItem)
		for protocol, retPort := range retPortMap {
			retMap[protocol] = AllocatedPortItem{
				PoolKey:     cp.PoolKey,
				PoolItemKey: item.ItemStatus.GetKey(),
				Protocol:    protocol,
				StartPort:   retPort.StartPort,
				EndPort:     retPort.EndPort,
				IsUsed:      retPort.Used,
			}
		}
		return item.ItemStatus, retMap, nil
	}
	return nil, nil, fmt.Errorf("no suitable item in pool %s", cp.PoolKey)
}

// ReleasePortBinding release port to port pool cache
func (cp *CachePool) ReleasePortBinding(poolItemKey, protocol string, startPort, endPort int) {
	for _, item := range cp.ItemList {
		if item.GetKey() == poolItemKey {
			item.Release(protocol, startPort, endPort)
		}
	}
}

// SetPortBindingUsed occupy port item
func (cp *CachePool) SetPortBindingUsed(poolItemKey, protocol string, startPort, endPort int) {
	for _, item := range cp.ItemList {
		if item.GetKey() == poolItemKey {
			item.SetPortUsed(protocol, startPort, endPort)
		}
	}
}
