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

package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// HeaderKey string
const (
	// RequestIDHeaderKey xxx
	RequestIDHeaderKey = "X-Request-Id"
)

// ParseOpenTelemetryEndpoint 解析OpenTelemetry endpoint
// nolint:lll
// 返回值：host, port, path
// bkm-collector.bkmonitor-operator.svc.cluster.local:443/v1/traces (bkm-collector.bkmonitor-operator.svc.cluster.local, 443, /v1/traces)
// bkm-collector.bkmonitor-operator.svc.cluster.local:4318 (bkm-collector.bkmonitor-operator.svc.cluster.local, 4318, "")
func ParseOpenTelemetryEndpoint(endpoint string) (string, int32, string, error) {
	if endpoint == "" {
		return "", 0, "", fmt.Errorf("endpoint cannot be empty")
	}

	// 移除协议前缀
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	// 分离主机:端口和路径部分
	var hostPort, path string
	if idx := strings.Index(endpoint, "/"); idx >= 0 {
		hostPort = endpoint[:idx]
		path = endpoint[idx:]
	} else {
		hostPort = endpoint
		path = ""
	}

	// 解析主机和端口，需要特殊处理IPv6地址
	var host, portStr string
	if strings.HasPrefix(hostPort, "[") {
		// IPv6 地址格式: [::1]:8080
		closeBracketIdx := strings.Index(hostPort, "]:")
		if closeBracketIdx == -1 {
			return "", 0, "", fmt.Errorf("invalid IPv6 endpoint format, expected [host]:port[/path], got: %s", endpoint)
		}
		host = hostPort[:closeBracketIdx+1] // 包含方括号
		portStr = hostPort[closeBracketIdx+2:]
	} else {
		// IPv4 地址或域名格式: host:port
		parts := strings.Split(hostPort, ":")
		if len(parts) != 2 {
			return "", 0, "", fmt.Errorf("invalid endpoint format, expected host:port[/path], got: %s", endpoint)
		}
		host = parts[0]
		portStr = parts[1]
	}

	if host == "" {
		return "", 0, "", fmt.Errorf("host cannot be empty in endpoint: %s", endpoint)
	}

	// 解析端口，使用 strconv.Atoi 来确保端口是纯数字
	portInt, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, "", fmt.Errorf("invalid port in endpoint %s: %v", endpoint, err)
	}

	if portInt <= 0 || portInt > 65535 {
		return "", 0, "", fmt.Errorf("port must be between 1 and 65535, got: %d", portInt)
	}

	return host, int32(portInt), path, nil
}
