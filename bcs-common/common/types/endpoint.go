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

//BcsEndpoint endpoint definition
type BcsEndpoint struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata,omitempty"`
	Endpoints  []Endpoint `json:"eps"`
}

//Endpoint endpoint info for ip address
type Endpoint struct {
	NetworkMode string          `json:"networkMode"`
	NodeIP      string          `json:"nodeIP"`
	ContainerIP string          `json:"containerIP"`
	Target      TargetRef       `json:"targetRef,omitempty"`
	Ports       []ContainerPort `json:"ports,omitempty"`
}

//TargetRef referrence for endpoint
type TargetRef struct {
	Kind      string `json:"kind"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}
