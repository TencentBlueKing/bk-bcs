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

package tools

import (
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/util"
)

const (
	podIPEnv      = "POD_IP"         // 单栈监听环境变量
	podIPsEnv     = "POD_IPs"        // 双栈监听环境变量
	ipv6Interface = "IPV6_INTERFACE" // ipv6本地网关地址
)

// GetIPsFromEnv get podIP and podIPs from env
func GetIPsFromEnv() (string, []string) {
	podIP := os.Getenv(podIPEnv)
	if podIP == "" {
		podIP = "127.0.0.1"
	}
	if os.Getenv(podIPsEnv) == "" {
		return podIP, []string{}
	}
	podIPs := strings.Split(os.Getenv(podIPsEnv), ",")
	return podIP, podIPs
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

// GetListenAddrs for dualstack
func GetListenAddrs(addrs []string, port int) []string {
	result := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		result = append(result, GetListenAddr(addr, port))
	}
	return result
}
