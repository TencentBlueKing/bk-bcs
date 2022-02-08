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
 *
 */

package utils

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"net"
)

// CIDR Info
type CIDR struct {
	ip    net.IP
	ipnet *net.IPNet
}

// ParseCIDR parse cidr
func ParseCIDR(s string) (*CIDR, error) {
	i, n, err := net.ParseCIDR(s)
	if err != nil {
		return nil, err
	}
	return &CIDR{ip: i, ipnet: n}, nil
}

// Equal check cidr equal
func (c CIDR) Equal(ns string) bool {
	c2, err := ParseCIDR(ns)
	if err != nil {
		return false
	}
	return c.ipnet.IP.Equal(c2.ipnet.IP)
}

// IsIPv4 check
func (c CIDR) IsIPv4() bool {
	_, bits := c.ipnet.Mask.Size()
	return bits/8 == net.IPv4len
}

// IsIPv6 check
func (c CIDR) IsIPv6() bool {
	_, bits := c.ipnet.Mask.Size()
	return bits/8 == net.IPv6len
}

// Contains cidr include ip
func (c CIDR) Contains(ip string) bool {
	return c.ipnet.Contains(net.ParseIP(ip))
}

// CIDR format
func (c CIDR) CIDR() string {
	return c.ipnet.String()
}

// IP segment for cidr
func (c CIDR) IP() string {
	return c.ip.String()
}

// Network network
func (c CIDR) Network() string {
	return c.ipnet.IP.String()
}

// MaskSize size
func (c CIDR) MaskSize() (ones, bits int) {
	ones, bits = c.ipnet.Mask.Size()
	return
}

// Mask xxx
func (c CIDR) Mask() string {
	mask, err := hex.DecodeString(c.ipnet.Mask.String())
	if err != nil {
		return ""
	}
	return net.IP([]byte(mask)).String()
}

// Gateway address
func (c CIDR) Gateway() string {
	return ""
}

// Broadcast address
func (c CIDR) Broadcast() string {
	mask := c.ipnet.Mask
	bcst := make(net.IP, len(c.ipnet.IP))
	copy(bcst, c.ipnet.IP)
	for i := 0; i < len(mask); i++ {
		ipIdx := len(bcst) - i - 1
		bcst[ipIdx] = c.ipnet.IP[ipIdx] | ^mask[len(mask)-i-1]
	}
	return bcst.String()
}

// IPRange begin-end IP
func (c CIDR) IPRange() (start, end string) {
	return c.Network(), c.Broadcast()
}

// IPCount IP num
func (c CIDR) IPCount() *big.Int {
	ones, bits := c.ipnet.Mask.Size()
	return big.NewInt(0).Lsh(big.NewInt(1), uint(bits-ones))
}

// ForEachIP iterator each ip
func (c CIDR) ForEachIP(iterator func(ip string) error) error {
	next := make(net.IP, len(c.ipnet.IP))
	copy(next, c.ipnet.IP)
	for c.ipnet.Contains(next) {
		if err := iterator(next.String()); err != nil {
			return err
		}
		IncrIP(next)
	}
	return nil
}

// ForEachIPBeginWith from beginIP
func (c CIDR) ForEachIPBeginWith(beginIP string, iterator func(ip string) error) error {
	next := net.ParseIP(beginIP)
	for c.ipnet.Contains(next) {
		if err := iterator(next.String()); err != nil {
			return err
		}
		IncrIP(next)
	}
	return nil
}

// IncrIP xxx
func IncrIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
}

// DecrIP xxx
func DecrIP(ip net.IP) {
	length := len(ip)
	for i := length - 1; i >= 0; i-- {
		ip[length-1]--
		if ip[length-1] < 0xFF {
			break
		}
		for j := 1; j < length; j++ {
			ip[length-j-1]--
			if ip[length-j-1] < 0xFF {
				return
			}
		}
	}
}

// Compare IP a&b; a == b return 0; a > b return 1; a < b return -1
func Compare(a, b net.IP) int {
	return bytes.Compare(a, b)
}
