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

	"github.com/apparentlymart/go-cidr/cidr"
)

// Subnet container info for vpc subnet
type Subnet struct {
	ID           string     `json:"id,omitempty"`
	IPNet        *net.IPNet `json:"ipNet,omitempty"`
	Name         string     `json:"name,omitempty"`
	Zone         string     `json:"zone,omitempty"`
	ZoneName     string     `json:"zonename,omitempty"`
	VpcID        string     `json:"vpcId,omitempty"`
	CreatedTime  string     `json:"createdTime,omitempty"`
	AvailableIps uint64     `json:"availableIps,omitempty"`
	TotalIps     uint64     `json:"totalIps,omitempty"`
}

// InSlice report wether a in s
func InSlice(a *net.IPNet, s []*net.IPNet) bool {
	for _, v := range s {
		if IsIPnetEqual(a, v) {
			return true
		}
	}
	return false
}

// StringToCidr convert string to cidr
func StringToCidr(cidrstr string) (cidr *Cidr, err error) {
	_, ipnet, err := net.ParseCIDR(cidrstr)
	if err != nil {
		return nil, err
	}
	cidr = &Cidr{
		IPNet: ipnet,
	}
	return cidr, err
}

// GetFreeIPNets get cidr block free cidrs
func GetFreeIPNets(allBlocks, reservedBlocks, allExistingSubnets []*net.IPNet) []*net.IPNet {
	var allFrees []*net.IPNet

	for _, block := range allBlocks {
		exsits := filterSubnet(block, allExistingSubnets)

		man := NewCidrManager(block, exsits)
		// nolint
		for _, free := range man.GetFrees() {
			if !inReserved(free, reservedBlocks) {
				allFrees = append(allFrees, free)
			}
		}
	}
	return allFrees
}

// filterSubnet cidr filter allocated subnets
func filterSubnet(cidrBlock *net.IPNet, subnets []*net.IPNet) []*net.IPNet {
	var filtered []*net.IPNet
	for _, subnet := range subnets {
		if CidrContains(cidrBlock, subnet) {
			filtered = append(filtered, subnet)
		}
	}
	return filtered
}

// subnet exist reserved cidrs
// nolint
func inReserved(subnet *net.IPNet, reservedBlocks []*net.IPNet) bool {
	for _, r := range reservedBlocks {
		if CidrContains(r, subnet) {
			return true
		}
	}
	return false
}

// CidrContains cidr(a) contain cidr(b)
func CidrContains(a, b *net.IPNet) bool {
	first, last := cidr.AddressRange(b)
	if a.Contains(first) && a.Contains(last) {
		return true
	}
	return false
}

// GetIPNumByCidr get ip num by cidr
func GetIPNumByCidr(cidr string) (ipnum uint32, err error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return 0, err
	}
	prefixSize, totalSize := ipnet.Mask.Size()
	if totalSize > 32 {
		ipnum = 0
		err = errors.New("currently only ipv4 cidr is supported")
		return
	}
	ipnum = uint32(math.Pow(2, float64(totalSize-prefixSize)))
	return ipnum, nil
}

// GetIPNum get ip num
func GetIPNum(ipnet *net.IPNet) (ipnum uint32, err error) {
	prefixSize, totalSize := ipnet.Mask.Size()
	if totalSize > 32 {
		ipnum = 0
		err = errors.New("currently only ipv4 cidr is supported")
		return
	}
	ipnum = uint32(math.Pow(2, float64(totalSize-prefixSize)))
	return ipnum, nil
}

// GetIPNetsNum get ip nets num
func GetIPNetsNum(frees []*net.IPNet) (uint32, error) {
	ipSurplus := uint32(0)
	for _, free := range frees {
		freeIPNum, errLocal := GetIPNum(free)
		if errLocal != nil {
			return 0, errLocal
		}
		ipSurplus += freeIPNum
	}

	return ipSurplus, nil
}

// VpcInfo vpc info
type VpcInfo struct {
	AvailableIpAddressCount uint32         `json:"availableIpAddressCount,omitempty"`
	TotalIpAddressCount     uint32         `json:"totalIpAddressCount,omitempty"`
	AvailableCidrBlock      []string       `json:"availableCidrBlock,omitempty"`
	CidrBlock               []string       `json:"cidrBlock,omitempty"`
	SubnetIPCidr            []SubnetIPCidr `json:"subnetIpcidr,omitempty"`
}

// SubnetIPCidr subnet ip cidr
type SubnetIPCidr struct {
	IPCidr string `json:"ipCidr,omitempty"`
	IPNum  uint32 `json:"ipNum,omitempty"`
}
