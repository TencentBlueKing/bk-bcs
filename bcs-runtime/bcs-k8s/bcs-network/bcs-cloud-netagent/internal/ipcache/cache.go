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

package ipcache

import (
	"sync"

	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
)

// Cache cache for eni and ips
type Cache struct {
	sync.Mutex
	eniIPMap map[string]map[string]*pbcommon.IPObject
}

// NewCache create new cache
func NewCache() *Cache {
	return &Cache{
		eniIPMap: make(map[string]map[string]*pbcommon.IPObject),
	}
}

// PutEniIP put eni ip
func (c *Cache) PutEniIP(eniID string, ip *pbcommon.IPObject) {
	c.Lock()
	defer c.Unlock()
	_, ok1 := c.eniIPMap[eniID]
	if !ok1 {
		c.eniIPMap[eniID] = make(map[string]*pbcommon.IPObject)
	}
	c.eniIPMap[eniID][ip.Address] = ip
}

// GetEniIP get eni ip
func (c *Cache) GetEniIP(eniID string, addr string) *pbcommon.IPObject {
	c.Lock()
	defer c.Unlock()
	ipMap, ok1 := c.eniIPMap[eniID]
	if !ok1 {
		return nil
	}
	ip, ok2 := ipMap[addr]
	if !ok2 {
		return nil
	}
	return ip
}

// DeleteEniIP delete eni ip by containerID
func (c *Cache) DeleteEniIPbyContainerID(containerID string) {
	c.Lock()
	defer c.Unlock()
	var delEniID string
	var delIP string
	for eniID, ipMap := range c.eniIPMap {
		for ip, ipObj := range ipMap {
			if ipObj.ContainerID == containerID {
				delEniID = eniID
				delIP = ip
				break
			}
		}
		if delEniID != "" {
			break
		}
	}
	if delEniID != "" {
		delete(c.eniIPMap[delEniID], delIP)
	}
}

// ListEniIP list eni by eni name
func (c *Cache) ListEniIP(eniID string) []*pbcommon.IPObject {
	c.Lock()
	defer c.Unlock()
	ipMap, ok1 := c.eniIPMap[eniID]
	if !ok1 {
		return nil
	}
	var retList []*pbcommon.IPObject
	for _, ip := range ipMap {
		retList = append(retList, ip)
	}
	return retList
}
