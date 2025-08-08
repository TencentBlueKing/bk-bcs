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

// Package utils xxx
package utils

import (
	"net"
	"os"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/util"
)

const (
	podIPsEnv     = "POD_IPs"        // 双栈监听环境变量
	ipv6Interface = "IPV6_INTERFACE" // ipv6本地网关地址

	// QueryFallbackTime metrics 查询回退时间
	QueryFallbackTime = time.Minute
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

// GetListenAddr xxx
func GetListenAddr(addr, port string) string {
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

	return net.JoinHostPort(addr, port)
}

// StringInSlice 判断字符串是否存在 Slice 中
func StringInSlice(str string, list []string) bool {
	for _, item := range list {
		if item == str {
			return true
		}
	}
	return false
}

// StringJoinWithRegex 数组转化为字符串，并添加正则
func StringJoinWithRegex(list []string, sep, reg string) string {
	ss := make([]string, 0)
	for i := range list {
		if len(list[i]) == 0 {
			continue
		}
		ss = append(ss, list[i]+reg)
	}
	return strings.Join(ss, sep)
}

// StringJoinIPWithRegex 数组转化为字符串，并添加正则
func StringJoinIPWithRegex(list []string, sep, reg string) string {
	ss := make([]string, 0)
	for i := range list {
		if len(list[i]) == 0 {
			continue
		}
		ss = append(ss, net.JoinHostPort(list[i], "")+reg)
	}
	return strings.Join(ss, sep)
}

// GetNowQueryTime 获取当前时间
func GetNowQueryTime() time.Time {
	// 往前退一分钟, 避免数据上报慢
	return time.Now().Add(-QueryFallbackTime)
}
