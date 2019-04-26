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
	_, classA, _  = net.ParseCIDR("10.0.0.0/8")
	_, classA1, _ = net.ParseCIDR("9.0.0.0/8")
	_, classAa, _ = net.ParseCIDR("100.64.0.0/10")
	_, classB, _  = net.ParseCIDR("172.16.0.0/12")
	_, classC, _  = net.ParseCIDR("192.168.0.0/16")
)

//GetIPAddress get local usable inner ip address
func GetIPAddress() (addrList []string) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return addrList
	}
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() && ip.IP.To4() != nil {
			if classA.Contains(ip.IP) {
				addrList = append(addrList, ip.IP.String())
				continue
			}
			if classA1.Contains(ip.IP) {
				addrList = append(addrList, ip.IP.String())
				continue
			}
			if classAa.Contains(ip.IP) {
				addrList = append(addrList, ip.IP.String())
				continue
			}
			if classB.Contains(ip.IP) {
				addrList = append(addrList, ip.IP.String())
				continue
			}
			if classC.Contains(ip.IP) {
				addrList = append(addrList, ip.IP.String())
				continue
			}
		}
	}
	return addrList
}
