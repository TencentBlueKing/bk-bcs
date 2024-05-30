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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// CachePort port entry in port list
type CachePort struct {
	StartPort int
	EndPort   int
	Used      bool

	// if used, set by portBinding
	RefName      string
	RefNamespace string
	RefType      string
	RefStartTime time.Time
}

// NewCachePort create cache port
func NewCachePort(startPort, endPort int) *CachePort {
	return &CachePort{
		StartPort: startPort,
		EndPort:   endPort,
	}
}

// GetKey get key for cache port
func (cp *CachePort) GetKey() string {
	return fmt.Sprintf("%d-%d", cp.StartPort, cp.EndPort)
}

// IsUsed is port used
func (cp *CachePort) IsUsed() bool {
	return cp.Used
}

// SetUsed set port in used status
func (cp *CachePort) SetUsed() {
	cp.Used = true
	cp.RefStartTime = time.Now()
}

// SetUnused set port in unused status
func (cp *CachePort) SetUnused() {
	cp.Used = false

	cp.RefType = ""
	cp.RefNamespace = ""
	cp.RefName = ""
	cp.RefStartTime = time.Time{}
}

// SetResource set port referenced resource
func (cp *CachePort) SetResource(refType, refNamespace, refName string) {
	if cp.RefType == "" && cp.RefNamespace == "" && cp.RefName == "" {
		cp.RefStartTime = time.Now()
	}
	cp.RefType = refType
	cp.RefNamespace = refNamespace
	cp.RefName = refName
}

// IsMatched if startPort and endPort are matched
func (cp *CachePort) IsMatched(startPort, endPort int) bool {
	return cp.StartPort == startPort && cp.EndPort == endPort
}

// CachePortList port list with certain protocol for pool item
type CachePortList struct {
	Protocol         string
	StartPort        int
	EndPort          int
	SegmentLength    int
	AvailablePortNum int // 可用的端口总数（注意这个数量不会变化，代表已分配量+未分配量）
	AllocatedPortNum int
	Ports            []*CachePort
}

// NewCachePortList create cache port list
func NewCachePortList(protocol string, startPort, endPort, segLen int) (*CachePortList, error) {
	if endPort <= startPort {
		return nil, fmt.Errorf("invalid start port %d and end port %d", startPort, endPort)
	}
	segmentLength := 1
	if segLen > 0 {
		segmentLength = segLen
	}
	list := &CachePortList{
		Protocol:      protocol,
		StartPort:     startPort,
		EndPort:       endPort,
		SegmentLength: segmentLength,
	}

	availableItem := (endPort - startPort) / segmentLength
	for index := 0; index < availableItem; index++ {
		tmpStartPort := startPort + index*segmentLength
		tmpEndPort := 0
		if segmentLength > 1 {
			tmpEndPort = startPort + (index+1)*segmentLength - 1
		}
		newPort := NewCachePort(tmpStartPort, tmpEndPort)
		list.Ports = append(list.Ports, newPort)
	}
	list.AvailablePortNum = availableItem
	list.AllocatedPortNum = 0
	return list, nil
}

func (cpl *CachePortList) getName() string {
	return fmt.Sprintf("cachelist (protocol %s, start %d, end %d, seg %d)",
		cpl.Protocol, cpl.StartPort, cpl.EndPort, cpl.SegmentLength)
}

func (cpl *CachePortList) hasEnoughPort() bool {
	return cpl.AvailablePortNum-cpl.AllocatedPortNum > 0
}

// GetAvailabePortNum get available port number
func (cpl *CachePortList) GetAvailabePortNum() int {
	return cpl.AvailablePortNum
}

// GetAllocatedPortNum get allocated port number
func (cpl *CachePortList) GetAllocatedPortNum() int {
	return cpl.AllocatedPortNum
}

// IsPortFree to see if port is free with given start port and end port
func (cpl *CachePortList) IsPortFree(startPort, endPort int) bool {
	for _, port := range cpl.Ports {
		if port.IsMatched(startPort, endPort) {
			if !port.IsUsed() {
				return true
			}
			return false
		}
	}
	return false
}

// Allocate allocate one cache port by start port and end port
func (cpl *CachePortList) Allocate(startPort, endPort int) *CachePort {
	if !cpl.hasEnoughPort() {
		blog.Warnf("%s has no enough ports", cpl.getName())
		return nil
	}
	for _, port := range cpl.Ports {
		if port.IsMatched(startPort, endPort) && !port.IsUsed() {
			port.SetUsed()
			cpl.AllocatedPortNum++
			return &CachePort{
				StartPort: port.StartPort,
				EndPort:   port.EndPort,
				Used:      port.Used,
			}
		}
	}
	return nil
}

// AllocateOne allocate one cache port without given start port and end port
func (cpl *CachePortList) AllocateOne() *CachePort {
	if !cpl.hasEnoughPort() {
		blog.Warnf("%s has no enough ports", cpl.getName())
		return nil
	}
	for _, port := range cpl.Ports {
		if !port.IsUsed() {
			port.SetUsed()
			cpl.AllocatedPortNum++
			return &CachePort{
				StartPort: port.StartPort,
				EndPort:   port.EndPort,
				Used:      port.Used,
			}
		}
	}
	return nil
}

// Release set cache port status to unused
func (cpl *CachePortList) Release(startPort, endPort int) {
	for _, port := range cpl.Ports {
		if port.IsMatched(startPort, endPort) && port.IsUsed() {
			port.SetUnused()
			cpl.AllocatedPortNum--
			return
		}
	}
	blog.Warnf("port %d-%d not found for %s when release", startPort, endPort, cpl.getName())
}

// SetPortUsed set port used
func (cpl *CachePortList) SetPortUsed(startPort, endPort int, refType, refNamespace, refName string) {
	for _, port := range cpl.Ports {
		if port.IsMatched(startPort, endPort) {
			if !port.IsUsed() {
				cpl.AllocatedPortNum++
			}

			port.SetUsed()
			port.SetResource(refType, refNamespace, refName)
			return
		}
	}
	blog.Warnf("port %d-%d not found for %s when set used", startPort, endPort, cpl.getName())
}

// IncreaseEndPort increase end port
func (cpl *CachePortList) IncreaseEndPort(endPort int) error {
	if endPort <= cpl.EndPort {
		return fmt.Errorf("endPort %d is smaller than current endPort %d", endPort, cpl.EndPort)
	}
	availableItem := (endPort - cpl.StartPort) / cpl.SegmentLength
	for index := cpl.AvailablePortNum; index < availableItem; index++ {
		tmpStartPort := cpl.StartPort + index*cpl.SegmentLength
		tmpEndPort := 0
		if cpl.SegmentLength > 1 {
			tmpEndPort = cpl.StartPort + (index+1)*cpl.SegmentLength - 1
		}
		newPort := NewCachePort(tmpStartPort, tmpEndPort)
		cpl.Ports = append(cpl.Ports, newPort)
	}
	cpl.AvailablePortNum = availableItem
	cpl.EndPort = endPort
	return nil
}
