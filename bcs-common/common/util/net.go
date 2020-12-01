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

package util

import "net"

var (
	_, classA, _ = net.ParseCIDR("10.0.0.0/8")
	_, classB, _ = net.ParseCIDR("172.16.0.0/12")
	_, classC, _ = net.ParseCIDR("192.168.0.0/16")
)

//GetIPAddress get local usable inner ip address
// check eth1 and eth0 first, if no IP address from eth1 and eth0,
// try to filter all network interface with private address
func GetIPAddress() string {
	//try eth1 first
	eth1Addr := getInterfaceIPv4Addr("eth1")
	if len(eth1Addr) != 0 {
		return eth1Addr
	}
	//try eth0
	eth0Addr := getInterfaceIPv4Addr("eth0")
	if len(eth0Addr) != 0 {
		return eth0Addr
	}
	//try all private network address
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() && ip.IP.To4() != nil {
			if classA.Contains(ip.IP) {
				return ip.IP.String()
			}
			if classB.Contains(ip.IP) {
				return ip.IP.String()
			}
			if classC.Contains(ip.IP) {
				return ip.IP.String()
			}
		}
	}
	return ""
}

//getInterfaceIPv4Addr get specified network interface IPv4 address
// if interface has multiple available IP addresses and already UP
// just return the first one
func getInterfaceIPv4Addr(name string) string {
	itf, err := net.InterfaceByName(name)
	if err != nil {
		return ""
	}
	if (itf.Flags & net.FlagUp) == 0 {
		return ""
	}
	addrs, err := itf.Addrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() && ip.IP.To4() != nil {
			return ip.IP.String()
		}
	}
	return ""
}
