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
	Nodes   map[string]int `json:"nodes"`
	Retries int            `json:"retries,omitempty"`
	//Checks  HealthCheck `json:"checks,omitempty"`
	Key          string            `json:"key,omitempty"`
	Timeout      *Timeout          `json:"timeout,omitempty"`
	HashOn       string            `json:"hash_on,omitempty"`
	PassHost     string            `json:"pass_host,omitempty"`
	UpstreamHost string            `json:"upstream_host,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	CreateTime   int               `json:"create_time,omitempty"`
	UpdateTime   int               `json:"update_time,omitempty"`
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
	ServiceProtocol string                 `json:"service_protocol,omitempty"`
	Plugins         map[string]interface{} `json:"plugins,omitempty"`
	Upstream        *Upstream              `json:"upstream,omitempty"`
	UpstreamID      string                 `json:"upstream_id,omitempty"`
	Service         *Service               `json:"service,omitempty"`
	ServiceID       string                 `json:"service_id,omitempty"`
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
