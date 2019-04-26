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

//ServiceLoadBalance loadbalance between multiple service with weight
type ServiceLoadBalance struct {
	ServiceName string `json:"name"`
	Protocol    string `json:"protocol"`
	Domain      string `json:"domain,omitempty"`
	ExportPort  int    `json:"exportPort,omitempty"`
	Weight      uint   `json:"weight"`
}

//BcsLoadBalance loadbalance for bcs-api
type BcsLoadBalance struct {
	TypeMeta `json:",inline"`
	//AppMeta     `json:",inline"`
	ObjectMeta  `json:"metadata"`
	Protocol    string               `json:"protocol"`
	Port        int                  `json:"port"`
	DomainName  string               `json:"domainName,omitempty"`
	ClusterIP   []string             `json:"clusterIP,omitempty"`
	LoadBalance []ServiceLoadBalance `json:"loadBalance"`
}
