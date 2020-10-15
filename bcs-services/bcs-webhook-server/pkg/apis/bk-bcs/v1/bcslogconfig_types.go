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

const (
	// DefaultConfigType default log config type
	DefaultConfigType = "default"
	// BcsSystemConfigType bcs system log config type
	BcsSystemConfigType = "bcs-system"
	// CustomConfigType custom log config type
	CustomConfigType = "custom"
)

// BcsLogConfig defines BcsLogConfig CRD format
type BcsLogConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              BcsLogConfigSpec `json:"spec"`
}

// BcsLogConfigSpec defines the content of BcsLogConfig CRD
type BcsLogConfigSpec struct {
	ConfigType        string            `json:"configType"`
	AppId             string            `json:"appId"`
	ClusterId         string            `json:"clusterId"`
	Stdout            bool              `json:"stdout"`
	StdDataId         string            `json:"stdDataId"`
	NonStdDataId      string            `json:"nonStdDataId"`
	LogPaths          []string          `json:"logPaths"`
	HostPaths         []string          `json:"hostPaths"`
	LogTags           map[string]string `json:"logTags"`
	WorkloadType      string            `json:"workloadType"`
	WorkloadName      string            `json:"workloadName"`
	WorkloadNamespace string            `json:"workloadNamespace"`
	ContainerConfs    []ContainerConf   `json:"containerConfs"`
	PodLabels         bool              `json:"podLabels"`
	Selector          PodSelector       `json:"selector"`
}

// ContainerConf defines log config for containers
type ContainerConf struct {
	ContainerName string            `json:"containerName"`
	Stdout        bool              `json:"stdout"`
	StdDataId     string            `json:"stdDataId"`
	NonStdDataId  string            `json:"nonStdDataId"`
	HostPaths     []string          `json:"hostPaths"`
	LogPaths      []string          `json:"logPaths"`
	LogTags       map[string]string `json:"logTags"`
}

// PodSelector defines selector format for BcsLogConfig CRD
type PodSelector struct {
	MatchLabels      map[string]string    `json:"matchLabels"`
	MatchExpressions []SelectorExpression `json:"matchExpressions"`
}

// SelectorExpression is universal expression for selector
type SelectorExpression struct {
	Key      string   `json:"key"`
	Operator string   `json:"operator"`
	Values   []string `json:"values"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BcsLogConfigList is the list expression of BcsLogConfig CRD
type BcsLogConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []BcsLogConfig `json:"items"`
}
