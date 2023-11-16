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
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/server/types"
)

// WriteErrorResponse writes a standard error response
func WriteErrorResponse(rw http.ResponseWriter, statusCode int, err *types.ErrorResponse) {
	payload, _ := json.Marshal(err)

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)
	rw.Write(payload)
}

// RefineIPv6Addr 修复ipv6链接问题(正常应该在上报地方处理, 目前只有websocket异常, 暂时这里修复)
func RefineIPv6Addr(serverAddress string) string {
	u, err := url.Parse(serverAddress)
	if err != nil {
		return serverAddress
	}

	// 正确的ipv6地址
	if strings.LastIndex(u.Host, "]") > 0 {
		return serverAddress
	}

	// ipv6 addr
	parts := strings.Split(u.Host, ":")
	if len(parts) > 2 {
		host := strings.Join(parts[:len(parts)-1], ":")
		port := parts[len(parts)-1]
		u.Host = net.JoinHostPort(host, port)
	}

	return u.String()
}
