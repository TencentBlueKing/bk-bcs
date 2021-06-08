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

package ipcache

import (
	"cryto/math"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/types"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/utils"
)

// Cache cache for ip object
type Cache struct {
	ipCache       map[string]map[string]map[string]bool
	subnetLockMap map[string]*sync.Mutex
	lock          sync.Mutex
}

// AddSubnet add subnet
func (c *Cache) AddSubnet(subnetCidr string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if isValid, _ := utils.ValidateIPv4Cidr(subnetCidr); !isValid {
		return fmt.Errorf("subnetCidr %s is invalid", subnetCidr)
	}
	if c.ipCache == nil {
		c.ipCache = make(map[string]map[string]map[string]bool)
		c.subnetLockMap = make(map[string]*sync.Mutex)
	}
	if _, ok := c.ipCache[subnetCidr]; ok {
		return fmt.Errorf("subnetCidr %s already exists", subnetCidr)
	}
	c.ipCache[subnetCidr] = make(map[string]map[string]bool)
	c.ipCache[subnetCidr][types.IPStatusActive] = make(map[string]bool)
	c.ipCache[subnetCidr][types.IPStatusAvailable] = make(map[string]bool)
	c.ipCache[subnetCidr][types.IPStatusReserved] = make(map[string]bool)
	return nil
}

// DeleteSubnet delete subnet
func (c *Cache) DeleteSubnet(subnetCidr string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if isValid, _ := utils.ValidateIPv4Cidr(subnetCidr); !isValid {
		return fmt.Errorf("subnetCidr %s is invalid", subnetCidr)
	}
	if c.ipCache == nil {
		return fmt.Errorf("cache is empty")
	}
	if _, ok := c.ipCache[subnetCidr]; !ok {
		return fmt.Errorf("subnetCidr %s does not exist", subnetCidr)
	}
	delete(c.ipCache, subnetCidr)
	delete(c.subnetLockMap, subnetCidr)
	return nil
}

// GetStatusIP get one ip by status
func (c *Cache) GetStatusIP(subnetCidr, status string) (string, error) {
	if _, ok := c.ipCache[subnetCidr]; !ok {
		return "", fmt.Errorf("subnetCidr %s does not exist", subnetCidr)
	}
	c.subnetLockMap[subnetCidr].Lock()
	defer c.subnetLockMap[subnetCidr].Unlock()
	k := rand.Intn(c.ipCache[subnetCidr])
	i := 0

	for ip := range c.ipCache[subnetCidr] {
		if i == k {
			return ip, nil
		}
		i++
	}
	return "", fmt.Errorf("never reach this code")
}

// TransIPStatus trans ip status
func (c *Cache) TransIPStatus(subnetCidr, ip, srcStatus, destStatus string) {
	if _, ok := c.ipCache[subnetCidr]; !ok {
		return "", fmt.Errorf("subnetCidr %s does not exist", subnetCidr)
	}
	if _, ok := c.ipCache[subnetCidr][srcStatus]; !ok {
		return "", fmt.Errorf("subnetCidr %s status %s ip does not exist", subnetCidr, srcStatus)
	}
	if _, ok := c.ipCache[subnetCidr][destStatus]; !ok {
		return "", fmt.Errorf("subnetCidr %s status %s ip does not exist", subnetCidr, destStatus)
	}
	
}
