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

package netservice

import "strings"

const (
	// IPStatus_RESERVED the reserved status of ip
	IPStatus_RESERVED = "reserved"
	// IPStatus_ACTIVE the active status of ip
	IPStatus_ACTIVE = "active"
	// IPStatus_AVAILABLE the available status of ip
	IPStatus_AVAILABLE = "available"
)

// NetPool pool info
type NetPool struct {
	Net       string   `json:"net"`                 // network string
	Mask      int      `json:"mask"`                // network mask
	Gateway   string   `json:"gateway"`             // network gateway
	Cluster   string   `json:"cluster,omitempty"`   // cluster that ip pool belongs to
	Hosts     []string `json:"hosts,omitempty"`     // host using network, can be none
	Available []string `json:"available,omitempty"` // available ip address
	Reserved  []string `json:"reserved,omitempty"`  // reserved ip address
	Active    []string `json:"active,omitempty"`    // active ip address
	Created   string   `json:"created,omitempty"`   // node create time
	Update    string   `json:"update,omitempty"`    // pool update time
}

// GetKey get key for pool
func (pool *NetPool) GetKey() string {
	return pool.Cluster + "/" + pool.Net
}

// IsValid check pool data is valid
func (pool *NetPool) IsValid() bool {
	if strings.TrimSpace(pool.Net) == "" {
		return false
	}
	if strings.TrimSpace(pool.Gateway) == "" {
		return false
	}
	if pool.Mask == 0 {
		return false
	}
	if strings.TrimSpace(pool.Cluster) == "" {
		return false
	}
	return true
}

// IPInst ip address instance in a pool
type IPInst struct {
	IPAddr     string `json:"ipaddr"`
	MacAddr    string `json:"macaddr,omitempty"`
	Pool       string `json:"pool"`
	Mask       int    `json:"mask"`
	Gateway    string `json:"gateway"`
	LastStatus string `json:"laststatus,omitempty"`
	Status     string `json:"status,omitempty"`
	Update     string `json:"update,omitempty"`
	Container  string `json:"container,omitempty"`
	Host       string `json:"host,omitempty"`
	Cluster    string `json:"cluster,omitempty"`
	App        string `json:"app,omitempty"`
}

// GetKey get key for ip instance
func (inst *IPInst) GetKey() string {
	return inst.IPAddr
}

// HostInfo for host node
type HostInfo struct {
	IPAddr     string             `json:"ipaddr"`
	MacAddr    string             `json:"macaddr,omitempty"`
	Gateway    string             `json:"gateway,omitempty"`
	Pool       string             `json:"pool"`
	Cluster    string             `json:"cluster,omitempty"`
	Created    string             `json:"created,omitempty"`
	Update     string             `json:"update,omitempty"`
	Containers map[string]*IPInst `json:"containers,omitempty"`
}

// GetKey get key for ip instance
func (host *HostInfo) GetKey() string {
	return host.IPAddr
}

// IsValid check host data is valid
func (host *HostInfo) IsValid() bool {
	if strings.TrimSpace(host.IPAddr) == "" {
		return false
	}
	if strings.TrimSpace(host.Cluster) == "" {
		return false
	}
	if strings.TrimSpace(host.Pool) == "" {
		return false
	}
	return true
}

// IPInfo ip information for ip resource request
type IPInfo struct {
	IPAddr  string `json:"ipaddr"`
	MacAddr string `json:"macaddr,omitempty"`
	Pool    string `json:"pool"`
	Mask    int    `json:"mask"`
	Gateway string `json:"gateway"`
}

// IPLease lease ip address
type IPLease struct {
	Host      string `json:"host"`
	Container string `json:"container"`
	IPAddr    string `json:"ipaddr"`
	App       string `json:"app,omitempty"`
}

// IPRelease release ip address from host
type IPRelease struct {
	Host      string `json:"host"`
	Container string `json:"container"`
	App       string `json:"app,omitempty"`
}

// NetStatic static info for net pool
type NetStatic struct {
	PoolNum     int `json:"poolnum"`
	ActiveIP    int `json:"activeip"`
	AvailableIP int `json:"availableip"`
	ReservedIP  int `json:"reservedip"`
}

// SSLInfo path for ca, privKey, pubKey
type SSLInfo struct {
	Key    string `json:"key"`
	PubKey string `json:"pubkey"`
	CACert string `json:"cacert,omitempty"`
	Passwd string `json:"passwd,omitempty"`
}

// ######################################################################################
//
// IP Pool management HTTP Request/Response definition
//
// ######################################################################################

// NetType type for bcs-netservice http request
type NetType int

const (
	// RequestType_POOL pool type
	RequestType_POOL NetType = 1
	// RequestType_HOST host type
	RequestType_HOST NetType = 2
	// RequestType_LEASE lease type
	RequestType_LEASE NetType = 3
	// RequestType_RELEASE release type
	RequestType_RELEASE NetType = 4

	// ResponseType_POOL poll type
	ResponseType_POOL NetType = 5
	// ResponseType_HOST host type
	ResponseType_HOST NetType = 6
	// ResponseType_LEASE lease type
	ResponseType_LEASE NetType = 7
	// ResponseType_RELEASE release type
	ResponseType_RELEASE NetType = 8
	// ResponseType_PSTATIC pstatic type
	ResponseType_PSTATIC NetType = 9
	// ResponseType_VIRTUALIP virtualip type
	ResponseType_VIRTUALIP NetType = 10
)

// NetRequest bcs-netservice http json request
type NetRequest struct {
	Type    NetType    `json:"type"`
	Pool    *NetPool   `json:"pool,omitempty"`
	Host    *HostInfo  `json:"host,omitempty"`
	Lease   *IPLease   `json:"lean,omitempty"`
	Release *IPRelease `json:"release,omitempty"`
	IPs     []string   `json:"ips,omitempty"`
}

// NetResponse bcs-netservice http json request
type NetResponse struct {
	Type    NetType     `json:"type"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Pool    []*NetPool  `json:"pool,omitempty"`
	Host    []*HostInfo `json:"host,omitempty"`
	Lease   *IPLease    `json:"lean,omitempty"`
	Release *IPRelease  `json:"release,omitempty"`
	Info    []*IPInfo   `json:"ipinfo,omitempty"`
	Inst    []*IPInst   `json:"ipinst,omitempty"`
	PStatic *NetStatic  `json:"poolStatic,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// IsSucc check NetResponse is success when request
func (nr *NetResponse) IsSucc() bool {
	return nr.Code == 0
}

// ResourceRequest for host available ip qeury
type ResourceRequest struct {
	Cluster string   `json:"cluster"`
	Hosts   []string `json:"hosts"`
}

// IsValid check data is valid
func (req *ResourceRequest) IsValid() bool {
	if strings.TrimSpace(req.Cluster) == "" {
		return false
	}
	if len(req.Hosts) == 0 {
		return false
	}
	return true
}

// ResourceResponse response for IPInfoResponse
type ResourceResponse struct {
	Code         int            `json:"code"`
	Message      string         `json:"message"`
	Cluster      string         `json:"cluster,omitempty"`
	HostResource map[string]int `json:"resource,omitempty"`
}

// ///////////////////////////////////////////////////////////////
// /////////////// IP transfer request/response  ////////////////
// /////////////////////////////////////////////////////////////

const (
	// ALL_IP_FAILED all ip failed
	ALL_IP_FAILED int = 1
	// SOME_IP_FAILED some ip failed
	SOME_IP_FAILED int = 2
)

// TranIPAttrInput request info
type TranIPAttrInput struct {
	Net        string   `json:"net"`     // network string
	Cluster    string   `json:"cluster"` // cluster that ip pool belongs to
	IPList     []string `json:"iplist"`  // tran ip address
	SrcStatus  string   `json:"src"`     // src status
	DestStatus string   `json:"dest"`    // src status
}

// IsValid if trans ip attribute input is valid
func (tr *TranIPAttrInput) IsValid() bool {
	if tr.SrcStatus != IPStatus_RESERVED && tr.SrcStatus != IPStatus_AVAILABLE {
		return false
	}

	if tr.DestStatus != IPStatus_RESERVED && tr.DestStatus != IPStatus_AVAILABLE {
		return false
	}

	if strings.TrimSpace(tr.Net) == "" {
		return false
	}

	if strings.TrimSpace(tr.Cluster) == "" {
		return false
	}

	if len(tr.IPList) == 0 {
		return false
	}

	return true
}

// TranIPAttrOutput response info
type TranIPAttrOutput struct {
	Result
}
