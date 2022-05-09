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

package admin

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

const (
	//BalanceTypeRoundrobin roundrobin strategy
	BalanceTypeRoundrobin = "roundrobin"
	//BalanceTypeChash chash strategy
	BalanceTypeChash = "chash"

	//ProtocolHTTP protocol http
	ProtocolHTTP = "http"
	//ProtocolGrpc protocol grpc
	ProtocolGrpc = "grpc"

	//IndexKeyRemoteAddr key define for remote_addr
	IndexKeyRemoteAddr = "remote_addr"
	//IndexKeyServerAddr key define for server_addr
	IndexKeyServerAddr = "server_addr"

	//ApisixAdmin system apisix
	ApisixAdmin = "apisix_admin"
)

// Timeout definition for proxy
type Timeout struct {
	Connect int `json:"connect"`
	Send    int `json:"send"`
	Read    int `json:"read"`
}

//Upstream apisix upstream object definition
type Upstream struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	//roundrobin or chash
	Type string `json:"type"`
	//backend info, format like ip:port
	Nodes            json.RawMessage `json:"nodes"`
	MapStructedNodes *map[string]int `json:"-"`
	// UpstreamNodes structure is needed because apisix loadbalancer use priority field of nodes
	UpstreamNodes *[]UpstreamNode `json:"-"`
	Retries       int             `json:"retries,omitempty"`
	//Checks  HealthCheck `json:"checks,omitempty"`
	Key          string            `json:"key,omitempty"`
	Timeout      *Timeout          `json:"timeout,omitempty"`
	HashOn       string            `json:"hash_on,omitempty"`
	PassHost     string            `json:"pass_host,omitempty"`
	UpstreamHost string            `json:"upstream_host,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	Scheme       string            `json:"scheme,omitempty"`
	CreateTime   int               `json:"create_time,omitempty"`
	UpdateTime   int               `json:"update_time,omitempty"`
}

//UpstreamNode apisix upstream's node object definition
type UpstreamNode struct {
	Host     string `json:"host"`
	Port     *int   `json:"port,omitempty"`
	Weight   int    `json:"weight"`
	Priority int    `json:"priority"`
}

// Service apisix service object definition
type Service struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name,omitempty"`
	Upstream   *Upstream              `json:"upstream,omitempty"`
	UpstreamID string                 `json:"upstream_id,omitempty"`
	Plugins    map[string]interface{} `json:"plugins,omitempty"`
	Labels     map[string]string      `json:"labels,omitempty"`
	Websocket  bool                   `json:"enable_websocket,omitempty"`
	CreateTime int                    `json:"create_time,omitempty"`
	UpdateTime int                    `json:"update_time,omitempty"`
}

// Route apisix service object definition
type Route struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	// match rules
	URI         string            `json:"uri,omitempty"`
	URIs        []string          `json:"uris,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Host        string            `json:"host,omitempty"`
	Hosts       []string          `json:"hosts,omitempty"`
	RemoteAddr  string            `json:"remote_addr,omitempty"`
	RemoteAddrs []string          `json:"remote_addrs,omitempty"`
	Method      []string          `json:"methods,omitempty"`
	Priority    int               `json:"priority,omitempty"`
	Vars        [][]string        `json:"vars,omitempty"`
	//information
	// ServiceProtocol string                 `json:"service_protocol,omitempty"`
	Plugins    map[string]interface{} `json:"plugins,omitempty"`
	Upstream   *Upstream              `json:"upstream,omitempty"`
	UpstreamID string                 `json:"upstream_id,omitempty"`
	Service    *Service               `json:"service,omitempty"`
	ServiceID  string                 `json:"service_id,omitempty"`
	//Status: 1 up, 0 down
	Status     int  `json:"status,omitempty"`
	Websocket  bool `json:"enable_websocket,omitempty"`
	CreateTime int  `json:"create_time,omitempty"`
	UpdateTime int  `json:"update_time,omitempty"`
}

//ProxyRewrite plugin definition
type ProxyRewrite struct {
	Scheme   string            `json:"scheme,omitempty"`
	URI      string            `json:"uri,omitempty"`
	RegexURI []string          `json:"regex_uri,omitempty"`
	Host     string            `json:"host,omitempty"`
	Header   map[string]string `json:"headers,omitempty"`
}

// LimitRequest plugin definition
type LimitRequest struct {
	Rate  uint   `json:"rate"`
	Burst uint   `json:"burst"`
	Key   string `json:"key"`
	//default 503
	RejectCode uint `json:"reject_code,omitempty"`
}

// RequestID plugin definition
type RequestID struct {
	HeaderName      string `json:"header_name,omitempty"`
	IncludeResponse bool   `json:"include_in_response,omitempty"`
}

func (r *Route) DeepEqual(in *Route) bool {
	createTime := r.CreateTime
	updateTime := r.UpdateTime
	status := r.Status
	r.CreateTime = in.CreateTime
	r.UpdateTime = in.UpdateTime
	r.Status = in.Status
	ret := reflect.DeepEqual(r, in)
	r.CreateTime = createTime
	r.UpdateTime = updateTime
	r.Status = status
	return ret
}

func (u *Upstream) NodesEuqal(in *Upstream) bool {
	if u.UpstreamNodes == nil || in.UpstreamNodes == nil {
		return u.UpstreamNodes == in.UpstreamNodes
	}
	nodeMap := UpstreamNodes2UpstreamNodesMap(u.UpstreamNodes)
	nodeMapIn := UpstreamNodes2UpstreamNodesMap(in.UpstreamNodes)
	for k, v := range *nodeMap {
		if vin, ok := (*nodeMapIn)[k]; ok && reflect.DeepEqual(vin, v) {
			delete(*nodeMapIn, k)
		} else {
			return false
		}
	}
	if len(*nodeMapIn) != 0 {
		return false
	}
	return true
}

func (s *Service) DeepEqual(in *Service) bool {
	createTime := s.CreateTime
	updateTime := s.UpdateTime
	s.CreateTime = in.CreateTime
	s.UpdateTime = in.UpdateTime
	ret := reflect.DeepEqual(s, in)
	s.CreateTime = createTime
	s.UpdateTime = updateTime
	return ret
}

// NodesMap2UpstreamNodes convert to apisix upstream information
func NodesMap2UpstreamNodes(nodes *map[string]int) *[]UpstreamNode {
	retNodes := make([]UpstreamNode, 0)
	for host, weight := range *nodes {
		hostport := strings.Split(host, ":")
		node := UpstreamNode{
			Host:   hostport[0],
			Weight: weight,
		}

		if len(hostport) == 2 {
			port, err := strconv.Atoi(hostport[1])
			if err != nil {
				blog.Errorf("Convert nodes port from string to int failed, host is: %s, port value is: %s", hostport[0], hostport[1])
				continue
			}
			node.Port = &port
		}
		retNodes = append(retNodes, node)
	}
	return &retNodes
}

// UpstreamNodes2NodesMap convert to apisix upstream information
func UpstreamNodes2NodesMap(nodes *[]UpstreamNode) *map[string]int {
	retNodes := make(map[string]int)
	for _, upstreamNode := range *nodes {
		if upstreamNode.Port != nil {
			retNodes[fmt.Sprintf("%s:%d", upstreamNode.Host, *upstreamNode.Port)] = upstreamNode.Weight
		} else {
			retNodes[upstreamNode.Host] = upstreamNode.Weight
		}
	}
	return &retNodes
}

// UpstreamNodes2UpstreamNodesMap convert to apisix upstream information
func UpstreamNodes2UpstreamNodesMap(nodes *[]UpstreamNode) *map[string]UpstreamNode {
	retNodes := make(map[string]UpstreamNode)
	for _, upstreamNode := range *nodes {
		if upstreamNode.Port != nil {
			retNodes[fmt.Sprintf("%s:%d", upstreamNode.Host, *upstreamNode.Port)] = upstreamNode
		} else {
			retNodes[upstreamNode.Host] = upstreamNode
		}
	}
	return &retNodes
}
