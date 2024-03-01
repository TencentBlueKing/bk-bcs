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

package types

// NetPool pool info
type NetPool struct {
	PoolKey    string `json:"poolKey"`
	Mask       int    `json:"mask"`
	Gateway    string `json:"gateway"`
	Cluster    string `json:"cluster,omiempty"`
	CreateTime string `json:"createTime,omitempty"`
	UpdateTime string `json:"updateTime,omitempty"`
}

// IPInstance ip instance info
type IPInstance struct {
	IPAddr       string `json:"ipaddr"`
	MacAddr      string `json:"macaddr,omitempty"`
	NetPool      string `json:"netPool"`
	Mask         int    `json:"mask"`
	Gateway      string `json:"gateway"`
	Status       string `json:"status,omitempty"`
	PodName      string `json:"podName,omitempty"`
	PodNamespace string `json:"podNamespace,omitempty"`
	Container    string `json:"container,omitempty"`
	Host         string `json:"host,omitempty"`
	Cluster      string `json:"cluster,omitempty"`
}
