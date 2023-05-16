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

package types

import "sort"

// HTTPBackend hold Backend info for load balance
type HTTPBackend struct {
	Path         string
	UpstreamName string
	BackendList
}

// HTTPBackendList to ho
type HTTPBackendList []HTTPBackend

// Len 用于排序
func (hbl HTTPBackendList) Len() int {
	return len(hbl)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (hbl HTTPBackendList) Less(i, j int) bool {
	return hbl[i].Path < hbl[j].Path
}

// Swap swaps the elements with indexes i and j.
func (hbl HTTPBackendList) Swap(i, j int) {
	hbl[i], hbl[j] = hbl[j], hbl[i]
}

// NewHTTPServiceInfo construct http service info
func NewHTTPServiceInfo(s ServiceInfo, host string) HTTPServiceInfo {
	return HTTPServiceInfo{
		ServiceInfo:   s,
		BCSVHost:      host,
		CookieSession: true,
	}
}

// HTTPServiceInfo http service info
type HTTPServiceInfo struct {
	ServiceInfo
	Backends HTTPBackendList
	BCSVHost string // virtual host name, only use for http/hhtps
	ACL      string // ACL match rules, reserved
	// If SessionAffinity is set and without CookieStickySession, requests are routed to
	// a backend based on client ip. If both SessionAffinity and CookieStickSession are
	// set, a SERVERID cookie is inserted by the loadbalancer and used to route subsequent
	// requests. If neither is set, requests are routed based on the algorithm.
	// CookieStickySession use a cookie to enable sticky sessions.
	// The name of the cookie is SERVERID
	// This only can be used in http services
	CookieSession bool
	// Path          string //location nginx to transport specified uri
	SSLFlag bool
}

// AddBackend add backend to list
func (hsi *HTTPServiceInfo) AddBackend(b HTTPBackend) {
	hsi.Backends = append(hsi.Backends, b)
}

// SortBackends sort backend list
func (hsi *HTTPServiceInfo) SortBackends() {
	sort.Sort(hsi.Backends)
}

// HTTPServiceInfoList to hold http service info list
type HTTPServiceInfoList []HTTPServiceInfo

// SortBackends sort https backends by path,
// no need to sort HTTPBackend.BackendList because sort before assgin
func (hil *HTTPServiceInfoList) SortBackends() {
	for i, item := range *hil {
		sort.Sort(item.Backends)
		(*hil)[i] = item
	}
}

// AddItem for add HTTPServiceInfoItem to HTTPServiceInfoList
func (hil *HTTPServiceInfoList) AddItem(h HTTPServiceInfo) {
	for i, item := range *hil {
		if h.BCSVHost == item.BCSVHost && h.ServicePort == item.ServicePort {
			item.Backends = append(item.Backends, h.Backends...)
			(*hil)[i] = item
			return
		}
	}

	// not match data already exist
	*hil = append(*hil, h)
}

// Len is the number of elements in the collection.
func (hil HTTPServiceInfoList) Len() int {
	return len(hil)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (hil HTTPServiceInfoList) Less(i, j int) bool {
	return hil[i].Name < hil[j].Name
}

// Swap swaps the elements with indexes i and j.
func (hil HTTPServiceInfoList) Swap(i, j int) {
	hil[i], hil[j] = hil[j], hil[i]
}
