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

const (
	// DefaultConfigType default log config type
	DefaultConfigType = "default"
	// BcsSystemConfigType bcs system log config type
	BcsSystemConfigType = "bcs-system"
	// CustomConfigType custom log config type
	CustomConfigType = "custom"
	// HostConfigType is config for host path collection
	HostConfigType = "host"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BcsLogConfigSpec defines the desired state of BcsLogConfig
type BcsLogConfigSpec struct {
	ConfigType                     string            `json:"configType"`
	AppId                          string            `json:"appId"`
	ClusterId                      string            `json:"clusterId"`
	Stdout                         bool              `json:"stdout"`
	StdDataId                      string            `json:"stdDataId"`
	NonStdDataId                   string            `json:"nonStdDataId"`
	LogPaths                       []string          `json:"logPaths"`
	HostPaths                      []string          `json:"hostPaths"`
	LogTags                        map[string]string `json:"logTags"`
	Multiline                      *MultilineConf    `json:"multiline,omitempty"`
	WorkloadType                   string            `json:"workloadType"`
	WorkloadName                   string            `json:"workloadName"`
	WorkloadNamespace              string            `json:"workloadNamespace"`
	ContainerConfs                 []ContainerConf   `json:"containerConfs"`
	PodLabels                      bool              `json:"podLabels"`
	PodNamePattern                 string            `json:"podNamePattern"`
	Selector                       PodSelector       `json:"selector"`
	PackageCollection              bool              `json:"packageCollection"`
	ExitedContainerLogCloseTimeout int               `json:"exitedContainerLogCloseTimeout"`
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
	Multiline     *MultilineConf    `json:"multiline,omitempty"`
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

type MultilineConf struct {
	Type         string `json:"type,omitempty" yaml:"type,omitempty"`
	Pattern      string `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Negate       bool   `json:"negate,omitempty" yaml:"negate,omitempty"`
	Match        string `json:"match,omitempty" yaml:"match,omitempty"`
	FlushPattern string `json:"flushPattern,omitempty" yaml:"flush_pattern,omitempty"`
	MaxLines     *int   `json:"maxLines,omitempty" yaml:"max_lines,omitempty"`
	Timeout      string `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	CountLines   string `json:"countLines,omitempty" yaml:"count_lines,omitempty"`
	SkipNewline  *bool  `json:"skiplNewline,omitempty" yaml:"skip_newline,omitempty"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// BcsLogConfig is the Schema for the bcslogconfigs API
type BcsLogConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec BcsLogConfigSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// BcsLogConfigList contains a list of BcsLogConfig
type BcsLogConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BcsLogConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BcsLogConfig{}, &BcsLogConfigList{})
}
