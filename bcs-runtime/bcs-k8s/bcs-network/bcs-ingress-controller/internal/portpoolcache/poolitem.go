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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// CachePoolItem pool item in cache
type CachePoolItem struct {
	// PoolKey key of port pool which this item belongs to
	PoolKey      string
	ProtocolList []string
	Status       string
	// ItemStatus is port pool item in original PortPool applied by user
	ItemStatus *networkextensionv1.PortPoolItemStatus
	// key: protocol, value: port list
	PortListMap map[string]*CachePortList
}

// NewCachePoolItem create pool item object
func NewCachePoolItem(
	poolKey string, itemStatus *networkextensionv1.PortPoolItemStatus) (*CachePoolItem, error) {
	if itemStatus == nil {
		return nil, fmt.Errorf("original item status cannot be empty")
	}
	if len(itemStatus.Protocol) == 0 {
		return nil, fmt.Errorf("protocol list cannot be empty")
	}
	newItem := &CachePoolItem{
		PoolKey:      poolKey,
		ProtocolList: itemStatus.Protocol,
		ItemStatus:   itemStatus,
		Status:       itemStatus.Status,
		PortListMap:  make(map[string]*CachePortList),
	}
	for _, protocol := range itemStatus.Protocol {
		newList, err := NewCachePortList(protocol,
			int(itemStatus.StartPort), int(itemStatus.EndPort), int(itemStatus.SegmentLength))
		if err != nil {
			return nil, fmt.Errorf("create port list with protocol %s, item status %v", protocol, itemStatus)
		}
		newItem.PortListMap[protocol] = newList
	}
	return newItem, nil
}

// GetKey get key of item
func (cpi *CachePoolItem) GetKey() string {
	return cpi.ItemStatus.GetKey()
}

// SetStatus set status of pool cache status
func (cpi *CachePoolItem) SetStatus(status string) {
	cpi.Status = status
}

// SetItemsStatus set itemStatus of pool cache status
func (cpi *CachePoolItem) SetItemsStatus(itemStatus *networkextensionv1.PortPoolItemStatus) {
	cpi.ItemStatus = itemStatus
}

// IncreaseEndPort increase end port
func (cpi *CachePoolItem) IncreaseEndPort(endPort int) error {
	for protocol, list := range cpi.PortListMap {
		if err := list.IncreaseEndPort(endPort); err != nil {
			return fmt.Errorf("increase end port for protocol %s failed, err %s", protocol, err.Error())
		}
	}
	cpi.ItemStatus.EndPort = uint32(endPort)
	return nil
}

// AllocateAllProtocolPort allocate one port with all protocol
func (cpi *CachePoolItem) AllocateAllProtocolPort() map[string]*CachePort {
	retMap := make(map[string]*CachePort)
	if len(cpi.ProtocolList) < 2 {
		// Note: AllocateAllProtocolPort方法需要同时分配TCP、UDP协议端口
		// 目前portpool只处理TCP、UDP两种协议， 当协议小于2时，认为无法分配
		return nil
	}
	protocol := cpi.ProtocolList[0]
	list := cpi.PortListMap[protocol]
	for _, port := range list.Ports {
		if port.IsUsed() {
			continue
		}
		allFree := true
		for index := 1; index < len(cpi.ProtocolList); index++ {
			tmpProtocol := cpi.ProtocolList[index]
			if !cpi.PortListMap[tmpProtocol].IsPortFree(port.StartPort, port.EndPort) {
				allFree = false
				break
			}
		}
		if allFree {
			for index := 0; index < len(cpi.ProtocolList); index++ {
				tmpProtocol := cpi.ProtocolList[index]
				tmpPort := cpi.PortListMap[tmpProtocol].Allocate(port.StartPort, port.EndPort)
				retMap[tmpProtocol] = tmpPort
			}
			return retMap
		}
	}
	return nil
}

// Allocate allocate one port with given protocol
func (cpi *CachePoolItem) Allocate(protocol string) *CachePort {
	list, ok := cpi.PortListMap[protocol]
	if !ok {
		return nil
	}
	return list.AllocateOne()
}

// Release release one port with given protocol, start port and end port
func (cpi *CachePoolItem) Release(protocol string, startPort, endPort int) {
	list, ok := cpi.PortListMap[protocol]
	if !ok {
		blog.Warnf("protocol %s not found when release port", protocol)
		return
	}
	list.Release(startPort, endPort)
}

// SetPortUsed set port used
func (cpi *CachePoolItem) SetPortUsed(protocol string, startPort, endPort int) {
	list, ok := cpi.PortListMap[protocol]
	if !ok {
		blog.Warnf("protocol %s not found when set port used", protocol)
		return
	}
	list.SetPortUsed(startPort, endPort)
}
