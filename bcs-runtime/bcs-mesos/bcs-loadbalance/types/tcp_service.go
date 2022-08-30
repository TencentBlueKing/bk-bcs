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

// NewFourLayerServiceInfo to new a FourLayerServiceInfo
func NewFourLayerServiceInfo(s ServiceInfo, bl BackendList) FourLayerServiceInfo {
	return FourLayerServiceInfo{
		ServiceInfo: s,
		Backends:    bl,
	}
}

// FourLayerServiceInfo to hold tcp and udp service info
type FourLayerServiceInfo struct {
	ServiceInfo
	Backends BackendList // tcp Backend
}

// AddBackend add backend to list
func (tsi *FourLayerServiceInfo) AddBackend(b Backend) {
	tsi.Backends = append(tsi.Backends, b)
}

// SortBackends sort backend list
func (tsi *FourLayerServiceInfo) SortBackends() {
	sort.Sort(tsi.Backends)
}

// FourLayerServiceInfoList define serviceInfo list implementing sorter interface
type FourLayerServiceInfoList []FourLayerServiceInfo

// Len is the number of elements in the collection.
func (til FourLayerServiceInfoList) Len() int {
	return len(til)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (til FourLayerServiceInfoList) Less(i, j int) bool {
	return til[i].Name < til[j].Name
}

// Swap swaps the elements with indexes i and j.
func (til FourLayerServiceInfoList) Swap(i, j int) {
	til[i], til[j] = til[j], til[i]
}
