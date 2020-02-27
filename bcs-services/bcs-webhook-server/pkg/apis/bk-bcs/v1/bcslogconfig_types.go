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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BcsLogConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              BcsLogConfigSpec `json:"spec"`
}

type BcsLogConfigSpec struct {
	ConfigType     string            `json:"configType"`
	AppId          string            `json:"appId"`
	ClusterId      string            `json:"clusterId"`
	Stdout         bool              `json:"stdout"`
	StdDataId      string            `json:"stdDataId"`
	NonStdDataId   string            `json:"nonStdDataId"`
	LogPaths       []string          `json:"logPaths"`
	LogTags        map[string]string `json:"logTags"`
	WorkloadType   string            `json:"workloadType"`
	WorkloadName   string            `json:"workloadName"`
	ContainerConfs []ContainerConf   `json:"containerConfs"`
}

type ContainerConf struct {
	ContainerName string            `json:"containerName"`
	Stdout        bool              `json:"stdout"`
	StdDataId     string            `json:"stdDataId"`
	NonStdDataId  string            `json:"nonStdDataId"`
	LogPaths      []string          `json:"logPaths"`
	LogTags       map[string]string `json:"logTags"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BcsLogConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []BcsLogConfig `json:"items"`
}
