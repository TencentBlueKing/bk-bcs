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

package cidrtree

import (
	"errors"
	"math"
	"net"
)

// CidrManager define cidr operate
type CidrManager interface {
	// AllocateSubnet allocate directrouter subnet
	AllocateSubnet(
		vpcId, zone string, mask int, subnetName string, cidrBlocks, resevedBlocks []*net.IPNet) (*Subnet, error)
	DeleteSubnet(subnetId string) error
	// GetSubnetInfo return subnetinfo by subnetId
	GetSubnetInfo(subnetId string) (*Subnet, error)
	// ListSubnetInfo return a list of subnetinfo by vpcId
	ListSubnetInfo(vpcId string) ([]*Subnet, error)
}

// Cidr container info for vpc cidr
type Cidr struct {
	IPNet *net.IPNet `json:"ipNet,omitempty"`
	Type  string     `json:"type,omitempty"`
}

// String return cidr string
func (c *Cidr) String() string {
	return c.IPNet.String()
}

// GetIPNum returns the number of IP addresses of the cidr block, currently only ipv4
func (c *Cidr) GetIPNum() (ipnum uint32, err error) {
	prefixSize, totalSize := c.IPNet.Mask.Size()
	if totalSize > 32 {
		ipnum = 0
		err = errors.New("currently only ipv4 cidr is supported")
		return
	}
	ipnum = uint32(math.Pow(2, float64(totalSize-prefixSize)))
	return ipnum, nil
}
