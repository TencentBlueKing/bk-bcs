/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package tools

import (
	"net"
	"os"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/util"
)

const (
	podIPsEnv     = "POD_IPs"        // 双栈监听环境变量
	ipv6Interface = "IPV6_INTERFACE" // ipv6本地网关地址
)

// GetIPv6AddrFromEnv 解析ipv6
func GetIPv6AddrFromEnv() string {
	podIPs := os.Getenv(podIPsEnv)
	if podIPs == "" {
		return ""
	}

	ipv6 := util.GetIPv6Address(podIPs)
	if ipv6 == "" {
		return ""
	}

	// 在实际中，ipv6不能是回环地址
	if v := net.ParseIP(ipv6); v == nil || v.IsLoopback() {
		return ""
	}
	return ipv6
}

// GetListenAddr for dualstack
func GetListenAddr(addr string, port int) string {
	if ip := net.ParseIP(addr); ip == nil {
		return ""
	}

	if util.IsIPv6(addr) {
		// local link ipv6 需要带上 interface， 格式如::%eth0
		ipv6Interface := os.Getenv(ipv6Interface)
		if ipv6Interface != "" {
			addr = addr + "%" + ipv6Interface
		}
	}

	return net.JoinHostPort(addr, strconv.Itoa(port))
}
