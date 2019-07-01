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

//BcsService service definition
type BcsService struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata,omitempty"`
	Spec       ServiceSpec `json:"spec"`
}

//ServiceSpec service info specifics
type ServiceSpec struct {
	Selector  map[string]string `json:"selector"`
	Type      string            `json:"type,omitempty"` //k8s only
	ClusterIP []string          `json:"clusterIP,omitemtpy"`
	Ports     []ServicePort     `json:"ports"`
}

//ServicePort port info for Service
type ServicePort struct {
	Name       string `json:"name"`
	DomainName string `json:"domainName,omitempty"` //mesos only
	Path       string `json:"path"`
	Protocol   string `json:"protocol"`
	Port       int    `json:"servicePort"`
	TargetPort int    `json:"targetPort,omitempty"` //k8s only
	NodePort   int    `json:"nodePort,omitempty"`   //k8s only
}
