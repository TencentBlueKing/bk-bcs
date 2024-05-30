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

package httpserver

// HttpResult defines the http result
type HttpResult struct {
	NodeNum []*NodeNumObj `json:"nodeNum"`

	CPUPackingRate []*PackingRateObj `json:"CPUPackingRate"`
	CPUCapacity    []*CapacityObj    `json:"CPUCapacity"`

	MEMPackingRate []*PackingRateObj `json:"MEMPackingRate"`
	MEMCapacity    []*CapacityObj    `json:"MEMCapacity"`

	OptimizedNode     []NodeInfo `json:"OptimizedNode"`
	OptimizePrice     []PriceObj `json:"OptimizePrice"`
	TargetPackingRate float64    `json:"TargetPackingRate"`
}

// PriceObj defines the price obj
type PriceObj struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// NodeNumObj defines the node num obj
type NodeNumObj struct {
	Kind string `json:"kind"`
	Num  int    `json:"num"`
}

// PackingRateObj defines the rate obj
type PackingRateObj struct {
	Kind string  `json:"kind"`
	Rate float64 `json:"rate"`
}

// CapacityObj defines the capacity obj
type CapacityObj struct {
	Kind     string  `json:"kind"`
	Capacity float64 `json:"capacity"`
}

// NodeInfo defines node info
type NodeInfo struct {
	Name           string `json:"节点名称"`
	PodNum         string `json:"POD 数量"`
	CPUPackingRate string `json:"装箱率 (CPU)"`
	MEMPackingRate string `json:"装箱率 (MEM)"`
	CPUCapacity    string `json:"本身容量 (CPU)"`
	MEMCapacity    string `json:"本身容量 (MEM)"`
}

// NodeInfoList defines list of NodeINFO
type NodeInfoList []NodeInfo

// Len defines the len of NodeInfoList
func (n NodeInfoList) Len() int {
	return len(n)
}

// Less defines the less of NodeInfoList
func (n NodeInfoList) Less(i, j int) bool {
	return (n[i].CPUPackingRate + n[i].MEMPackingRate) < (n[j].CPUPackingRate + n[j].MEMPackingRate)
}

// Swap defines the swap of NodeInfoList
func (n NodeInfoList) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}
