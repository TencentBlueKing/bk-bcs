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

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/types"
)

// CheckKind check object if expected
func CheckKind(kind types.BcsDataType, by []byte) error {
	var meta *types.TypeMeta

	err := json.Unmarshal(by, &meta)
	if err != nil {
		return fmt.Errorf("Unmarshal TypeMeta failed: %s", err.Error())
	}

	if meta.Kind != kind {
		return fmt.Errorf("Kind %s is invalid", meta.Kind)
	}

	return nil
}

// IsIPv4 判断是否为IPv4地址
func IsIPv4(ip string) bool {
	return strings.Contains(ip, ".")
}

// IsIPv6 判断是否为IPv6地址
func IsIPv6(ip string) bool {
	return strings.Contains(ip, ":")
}

// VerifyIPv6 检验IP是否为合法IPv6
func VerifyIPv6(ip string) bool {
	return net.ParseIP(ip) != nil && IsIPv6(ip)
}

// GetIPv6Address 获取合法IPv6地址
func GetIPv6Address(ip string) string {
	ips := strings.Split(ip, ",")
	for _, v := range ips {
		if VerifyIPv6(v) { // 校验合法IP+合法v6地址
			return v
		}
	}
	return ""
}

// InitIPv6Address 根据 字段自身值 或 环境变量值 来初始化 字段
// 如：在IPv6集群时，默认k8s会给组件分配ipv6地址，可以在deployment中，通过status.podIPs获取ipv6地址，最后写入到容器的环境变量中。
// 假设，定义如下，该环境变量为"localIpv6"：
//
// ....
//
// - name: localIpv6
//   valueFrom
//     fieldRef
//       fieldPath: status.podIPs
// ....
//
// 则，在容器中，通过该环境变量“localIpv6”，获取到的值是"localIpv6=10.9.8.113,fd00:3:2:55",可以注意到这个值是IPv4地址在前面，
// IPv6地址在后面。因此，要获取IPv6地址的话，还需要进行截取操作，本方法就是用于该功能的。如果，当前IPv6Address已是合法地址，
// 则不会进行任何操作。否则，会进行如下操作：
//
// 1.检查当前字段 IPv6Address 是否为合法IPv6，若是合法IPv6，则结束执行；否则，执行下一步。
// 2.依次遍历当前字段 IPv6Address、“localIpv6”环境变量，检查是否存在"IPv4,IPv6"地址表示法，并检查IPv6地址合法性，
// 若，存在并合法，则把新的IPv6地址 赋值给 IPv6Address字段，并结束执行 ；否则，执行下一步
// 3.设置 IPv6Address 字段为默认值 "::1"
func InitIPv6Address(ip string) string {
	if VerifyIPv6(ip) {
		return ip
	}
	for _, ips := range []string{ip, os.Getenv(types.LOCALIPV6)} {
		if ipv6 := GetIPv6Address(ips); ipv6 != "" {
			return ipv6
		}
	}
	return net.IPv6loopback.String()
}

// errorIsBindAlreadyInUse 单栈ipv6环境，会出现该错误
func errorIsBindAlreadyInUse(err error) bool {
	return strings.Contains(err.Error(), "bind: address already in use")
}

// errorIsBindingCannotAssignAddress 单栈ipv4环境，会出现该错误
func errorIsBindingCannotAssignAddress(err error) bool {
	return strings.Contains(err.Error(), "bind: cannot assign requested address")
}

// CheckBindError 检查listen绑定错误
func CheckBindError(err error) bool {
	checkFunSets := []func(error) bool{
		errorIsBindAlreadyInUse,
		errorIsBindingCannotAssignAddress,
	}
	for _, fun := range checkFunSets {
		if fun(err) {
			return true
		}
	}
	return false
}
