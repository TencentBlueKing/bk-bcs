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
	"net"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-apiserver-proxy/pkg/ipvs"
)

// EndPoint wrap IP&Port
type EndPoint struct {
	IP   string
	Port uint32
}

// String trans endpoint to ip:port
func (ep EndPoint) String() string {
	port := strconv.Itoa(int(ep.Port))
	return ep.IP + ":" + port
}

// SplitServer split server to ip, port
func SplitServer(server string) (string, uint32) {
	s := strings.Split(server, ":")
	if len(s) != 2 {
		blog.Warn("SplitServer error: len(s) is not two.")
		return "", 0
	}
	blog.V(5).Infof("SplitServer debug: IP: %s, Port: %s", s[0], s[1])

	p, err := strconv.Atoi(s[1])
	if err != nil {
		blog.Warn("SplitServer error: ", err)
		return "", 0
	}
	return s[0], uint32(p)
}

// BuildVirtualServer build vip to ipvs.VirtualServer
func BuildVirtualServer(vip string) *ipvs.VirtualServer {
	ip, port := SplitServer(vip)
	virServer := &ipvs.VirtualServer{
		Address:   net.ParseIP(ip),
		Protocol:  "TCP",
		Port:      port,
		Scheduler: "rr",
		Flags:     0,
		Timeout:   0,
	}
	return virServer
}

// BuildRealServer build real to ipvs.RealServer
func BuildRealServer(real string) *ipvs.RealServer {
	ip, port := SplitServer(real)
	realServer := &ipvs.RealServer{
		Address: net.ParseIP(ip),
		Port:    port,
		Weight:  1,
	}
	return realServer
}
