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

package gw

import (
	"bk-bcs/bcs-common/common/blog"
	"strconv"
)

// LocationForwardStrategy forward strategy for location
type LocationForwardStrategy struct {
	Type string `json:"type,omitempty"`
	//1.DEFAULT: detail默认为”wrr”,根据rs的权重转发请求数, detai:[“wrr”, “ip_hash”, “least_conn”]
	//2.REDIRECT: 无RS，将请求重定向到detail指定的url,
	//3.RETRUNFAIL:无RS,直接返回错误码,detail为错误码
	Detail string `json:"detail,omitempty"`
}

// Diff judge if anotherLfs LocationForwardStrategy is the same with lfs LocationForwardStrategy
func (lfs *LocationForwardStrategy) Diff(anotherLfs *LocationForwardStrategy) bool {
	if lfs == nil && anotherLfs == nil {
		return false
	} else if lfs == nil || anotherLfs == nil {
		return true
	}
	if lfs.Type != anotherLfs.Type ||
		lfs.Detail != anotherLfs.Detail {
		return true
	}
	return false
}

// LocationSessionPersistence sesseion for location
type LocationSessionPersistence struct {
	Type           string `json:"type"`
	CookieTimeMode int    `json:"cookietime_mode"`
	Timeout        int    `json:"timeout"`
	CookieKey      string `json:"cookie_key"`
}

// Diff judge if anotherLsp LocationSessionPersistence is the same with lsp LocationSessionPersistence
func (lsp *LocationSessionPersistence) Diff(anotherLsp *LocationSessionPersistence) bool {
	if lsp == nil && anotherLsp == nil {
		return false
	} else if lsp == nil || anotherLsp == nil {
		return true
	}
	if lsp.Type != anotherLsp.Type ||
		lsp.CookieTimeMode != anotherLsp.CookieTimeMode ||
		lsp.Timeout != anotherLsp.Timeout ||
		lsp.CookieKey != anotherLsp.CookieKey {
		return true
	}
	return false
}

// LocationHealthCheck health check for location
type LocationHealthCheck struct {
	OP            string `json:"op"`       // set or del
	Protocol      string `json:"protocol"` // HTTP
	AliveNum      int    `json:"alive_num"`
	KickNum       int    `json:"kick_num"`
	ProbeInterval int    `json:"probe_interval"`
	AliveCode     int    `json:"alive_code"`
	ProbeURL      string `json:"probe_url"`
	Method        string `json:"method"`
	ServerName    string `json:"server_name"`
}

// Diff judge anotherLhc LocationHealthCheck is the same with lhc LocationHealthCheck
func (lhc *LocationHealthCheck) Diff(anotherLhc *LocationHealthCheck) bool {
	if lhc == nil && anotherLhc == nil {
		return false
	} else if lhc == nil || anotherLhc == nil {
		return true
	}
	if lhc.OP != anotherLhc.OP ||
		lhc.Protocol != anotherLhc.Protocol ||
		lhc.AliveNum != anotherLhc.AliveNum ||
		lhc.KickNum != anotherLhc.KickNum ||
		lhc.ProbeInterval != anotherLhc.ProbeInterval ||
		lhc.AliveCode != anotherLhc.AliveCode ||
		lhc.ProbeURL != anotherLhc.ProbeURL ||
		lhc.Method != anotherLhc.Method ||
		lhc.ServerName != anotherLhc.ServerName {
		return true
	}
	return false
}

// RealServer real server for location
type RealServer struct {
	IP     string `json:"rs_ip"`
	Port   int    `json:"rs_port"`
	VpcID  int    `json:"vpcid,omitempty"`
	HostIP string `json:"host_ip,omitempty"`
	Weight int    `json:"rs_weight"`
}

// Key get key string of RealServer
func (rs *RealServer) Key() string {
	return rs.IP + ":" + strconv.Itoa(rs.Port)
}

// Diff judge anotherRs RealServer is the same with rs RealServer
func (rs *RealServer) Diff(anotherRs *RealServer) bool {
	if rs == nil && anotherRs == nil {
		return false
	} else if rs == nil || anotherRs == nil {
		return true
	}
	if rs.IP != anotherRs.IP ||
		rs.Port != anotherRs.Port ||
		rs.VpcID != anotherRs.VpcID ||
		rs.HostIP != anotherRs.HostIP ||
		rs.Weight != anotherRs.Weight {
		return true
	}
	return false
}

// LocationRewrite rewrite config for location
type LocationRewrite struct {
	OP   string `json:"op,omitempty"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

// Diff judge anotherLr LocationRewrite is the same with lr LocationRewrite
func (lr *LocationRewrite) Diff(anotherLr *LocationRewrite) bool {
	if lr == nil && anotherLr == nil {
		return false
	} else if lr == nil || anotherLr == nil {
		return true
	}
	if lr.OP != anotherLr.OP ||
		lr.Type != anotherLr.Type ||
		lr.URL != anotherLr.URL {
		return true
	}
	return false
}

// Location location in request
type Location struct {
	LocationID             string                      `json:"location_id,omitempty"`
	URL                    string                      `json:"url"`
	LocationCustomizedConf string                      `json:"location_customized_conf"`
	LocLimitRate           int                         `json:"loc_limit_rate"`
	LocLimitStatusCode     int                         `json:"loc_limit_status_code"`
	ForwardStrategy        *LocationForwardStrategy    `json:"forward_strategy,omitempty"`
	SessionPersistence     *LocationSessionPersistence `json:"session_persistence,omitempty"`
	HealthCheck            *LocationHealthCheck        `json:"health_check,omitempty"`
	RSList                 []*RealServer               `json:"rs_list,omitempty"`
	Rewrite                *LocationRewrite            `json:"rewrite,omitempty"`
}

// Key get key string of location
func (l *Location) Key() string {
	return l.URL
}

// Diff judge anotherL Location is the same with l Location
func (l *Location) Diff(anotherL *Location) bool {
	if l == nil && anotherL == nil {
		return false
	} else if l == nil || anotherL == nil {
		return true
	}
	if l.LocationID != anotherL.LocationID ||
		l.URL != anotherL.URL ||
		l.LocationCustomizedConf != anotherL.LocationCustomizedConf ||
		l.LocLimitRate != anotherL.LocLimitRate ||
		l.LocLimitStatusCode != anotherL.LocLimitStatusCode ||
		l.ForwardStrategy.Diff(anotherL.ForwardStrategy) ||
		l.SessionPersistence.Diff(anotherL.SessionPersistence) ||
		l.HealthCheck.Diff(anotherL.HealthCheck) ||
		l.Rewrite.Diff(anotherL.Rewrite) {
		return true
	}

	if len(l.RSList) != len(anotherL.RSList) {
		return true
	}

	rsMap := make(map[string]*RealServer)
	for _, rs := range l.RSList {
		rsMap[rs.Key()] = rs
	}
	for _, rs := range anotherL.RSList {
		existedRs, ok := rsMap[rs.Key()]
		if !ok || existedRs.Diff(rs) {
			return true
		}
	}
	return false
}

// GetNewLocationWithExtraRS get new location with extra real server of newL Location campared with l Location
func (l *Location) GetNewLocationWithExtraRS(newL *Location) *Location {
	rsMap := make(map[string]*RealServer)
	for _, rs := range l.RSList {
		rsMap[rs.IP+strconv.Itoa(rs.Port)] = rs
	}
	var retRSList []*RealServer
	for _, rs := range newL.RSList {
		if _, ok := rsMap[rs.IP+strconv.Itoa(rs.Port)]; !ok {
			retRSList = append(retRSList, rs)
		}
	}
	if len(retRSList) == 0 {
		blog.Infof("no rs need to add")
		return nil
	}
	return &Location{
		LocationID:             l.LocationID,
		URL:                    l.URL,
		LocationCustomizedConf: l.LocationCustomizedConf,
		LocLimitRate:           l.LocLimitRate,
		LocLimitStatusCode:     l.LocLimitStatusCode,
		ForwardStrategy:        l.ForwardStrategy,
		SessionPersistence:     l.SessionPersistence,
		HealthCheck:            l.HealthCheck,
		RSList:                 retRSList,
		Rewrite:                l.Rewrite,
	}
}

// NewServiceWithoutLocationList new Service obj with empty location list
func NewServiceWithoutLocationList(svc *Service) *Service {
	return &Service{
		BizID:                   svc.BizID,
		VIPList:                 svc.VIPList,
		Domain:                  svc.Domain,
		VPort:                   svc.VPort,
		VpcID:                   svc.VpcID,
		Type:                    svc.Type,
		SSLEnable:               svc.SSLEnable,
		SSLVerifyClientEnable:   svc.SSLVerifyClientEnable,
		CertID:                  svc.CertID,
		DefaultServer:           svc.DefaultServer,
		ServerCustomizedConf:    svc.ServerCustomizedConf,
		VIPProtoLimitRate:       svc.VIPProtoLimitRate,
		VIPProtoLimitStatusCode: svc.VIPProtoLimitStatusCode,
		VSLimitRate:             svc.VSLimitRate,
		VSLimitStatusCode:       svc.VSLimitStatusCode,
	}
}

// ExclusiveSet exclusive set parameter
type ExclusiveSet struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Service service list
type Service struct {
	BizID                   string      `json:"biz_id"`
	VIPList                 []string    `json:"vip_list"`
	Domain                  string      `json:"domain"`
	VPort                   int         `json:"vport"`
	VpcID                   int         `json:"vpcid,omitempty"`                    // optional, just for private network loadbalance
	Type                    string      `json:"type,omitempty"`                     // optional, HTTP default, HTTP or HTTPS
	SSLEnable               bool        `json:"ssl_enable,omitempty"`               // optional, must be true for HTTPS
	SSLVerifyClientEnable   bool        `json:"ssl_verify_client_enable,omitempty"` // optional
	CertID                  string      `json:"cert_id,omitempty"`
	DefaultServer           bool        `json:"default_server"`
	ServerCustomizedConf    string      `json:"server_customized_conf,omitempty"`
	VIPProtoLimitRate       int         `json:"vip_proto_limit_rate,omitempty"`        // speed limit value
	VIPProtoLimitStatusCode int         `json:"vip_proto_limit_status_code,omitempty"` // speed limit status code
	VSLimitRate             int         `json:"vs_limit_rate,omitempty"`
	VSLimitStatusCode       int         `json:"vs_limit_status_code,omitempty"`
	LocationList            []*Location `json:"location_list,omitempty"`
}

// Key get service key of domain and vport
func (s *Service) Key() string {
	return s.Domain + "-" + strconv.Itoa(s.VPort)
}

// Diff judge another Service is the same with s Service
func (s *Service) Diff(another *Service) bool {
	if s == nil && another == nil {
		return false
	} else if s == nil || another == nil {
		return true
	}

	if s.BizID != another.BizID ||
		s.Domain != another.Domain ||
		s.VPort != another.VPort ||
		s.VpcID != another.VpcID ||
		s.Type != another.Type ||
		s.SSLEnable != another.SSLEnable ||
		s.SSLVerifyClientEnable != another.SSLVerifyClientEnable ||
		s.CertID != another.CertID ||
		s.DefaultServer != another.DefaultServer ||
		s.ServerCustomizedConf != another.ServerCustomizedConf ||
		s.VIPProtoLimitRate != another.VIPProtoLimitRate ||
		s.VIPProtoLimitStatusCode != another.VIPProtoLimitStatusCode ||
		s.VSLimitRate != another.VSLimitRate ||
		s.VSLimitStatusCode != another.VSLimitStatusCode ||
		len(s.LocationList) != len(another.LocationList) ||
		len(s.VIPList) != len(another.VIPList) {
		return true
	}

	vipMap := make(map[string]string)
	for _, ip := range s.VIPList {
		vipMap[ip] = ip
	}
	for _, ip := range s.VIPList {
		_, ok := vipMap[ip]
		if !ok {
			return true
		}
	}

	sMap := make(map[string]*Location)
	for _, l := range s.LocationList {
		sMap[l.Key()] = l
	}
	for _, l := range another.LocationList {
		existedL, ok := sMap[l.Key()]
		if !ok || existedL.Diff(l) {
			return true
		}
	}
	return false
}

// Request request
type Request struct {
	Cluster string `json:"cluster"`
}

// Response response
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// UpdateRequest request for sync gw services
type UpdateRequest struct {
	Request
	ServiceList []*Service `json:"service_list"`
}

// UpdateResponse response for sync gw services
type UpdateResponse struct {
	Response
}

// DeleteRequest request for delete gw services from concentrator cache
type DeleteRequest struct {
	Request
	ServiceList []*Service `json:"service_list"`
}

// DeleteResponse response for delete gw services
type DeleteResponse struct {
	Response
}
