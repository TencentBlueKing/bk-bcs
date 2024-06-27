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

package client

// CommonResp common response
type CommonResp struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Result    bool   `json:"result"`
	RequestID string `json:"request_id"`
}

// AllocateReq allocate request
type AllocateReq struct {
	PodName      string `json:"podName"`
	PodNamespace string `json:"podNamespace"`
	IPAddr       string `json:"ipAddr"`
	ContainerID  string `json:"containerID"`
	Host         string `json:"host"`
	HostGateway  string `json:"hostGateway"`
}

// AllocateRespData data of allocate response
type AllocateRespData struct {
	PodName      string `json:"podName"`
	PodNamespace string `json:"podNamespace"`
	IPAddr       string `json:"ipAddr"`
	ContainerID  string `json:"containerID"`
	Host         string `json:"host"`
	Mask         int    `json:"mask"`
	MacAddr      string `json:"macAddr,omitempty"`
	Gateway      string `json:"gateway"`
}

// AllocateResp allocate response
type AllocateResp struct {
	CommonResp
	Data *AllocateRespData `json:"data"`
}

// ReleaseReq release request
type ReleaseReq struct {
	PodName      string `json:"podName"`
	PodNamespace string `json:"podNamespace"`
	ContainerID  string `json:"containerID"`
	Host         string `json:"host"`
}

// ReleaseRespData data of release response
type ReleaseRespData struct {
	PodName      string `json:"podName"`
	PodNamespace string `json:"podNamespace"`
	ContainerID  string `json:"containerID"`
	Host         string `json:"host"`
}

// ReleaseResp release response
type ReleaseResp struct {
	CommonResp
	Data *ReleaseRespData `json:"data"`
}
