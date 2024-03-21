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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DescheduleType the type of policy
type DescheduleType string

// DescheduleSpec the spec of DeschedulePolicy
type DescheduleSpec struct {
	// +optional
	// +kubebuilder:validation:Enum=Converge;Balance
	// +kubebuilder:default=Converge
	Type DescheduleType `json:"type,omitempty" protobuf:"bytes,1,opt,name=type"`

	// Converge defines converge strategy, it will converge the resource of
	// all cluster. It drains the nodes which in lowWaterLevel, centralize
	// pods in cluster.
	Converge DescheduleConvergeStrategy `json:"converge,omitempty" protobuf:"bytes,2,opt,name=converge"`

	// Balance defines balance strategy, it will decentralize the pods in cluster.
	// This strategy is to keep the utilization of each node in the cluster as balanced
	// as possible
	// DOTO: 当前版本未实现
	Balance DescheduleBalanceStrategy `json:"balance,omitempty" protobuf:"bytes,3,opt,name=balance"`
}

// DescheduleConvergeStrategy defines converge strategy of DeschedulePolicy
type DescheduleConvergeStrategy struct {
	// Disabled defines whether this strategy is disabled, default true.
	// +kubebuilder:default=false
	Disabled bool `json:"disabled,omitempty" protobuf:"bytes,1,opt,name=disabled"`

	// TimeRange is the crontab format string, it will do strategy
	// with its define.
	TimeRange string `json:"timeRange,omitempty" protobuf:"bytes,2,opt,name=timeRange"`

	// ProfitTarget represents that the profit target, such as cpu/meme/nodes. Users
	// can fill it to define maximum resource of savings.
	ProfitTarget ProfitTarget `json:"profitTarget,omitempty" protobuf:"bytes,3,opt,name=profitTarget"`

	// MinPods defines the minimum number of pods that migrate at once
	MinPods int32 `json:"minPods,omitempty" protobuf:"bytes,4,opt,name=minPods"`
	// MaxPods defines the maximum number of pods that migrate at once
	MaxPods int32 `json:"maxPods,omitempty" protobuf:"bytes,5,opt,name=maxPods"`

	// LowWaterLevel represents the minimum water level limit the nodes needs to reach
	LowWaterLevel float32 `json:"lowWaterLevel,omitempty" protobuf:"bytes,6,opt,name=lowWaterLevel"`
	// HighWaterLevel represents the maximum water level limit the nodes needs to reach
	HighWaterLevel float32 `json:"highWaterLevel,omitempty" protobuf:"bytes,7,opt,name=highWaterLevel"`

	// Sequences defines the sequence of migrate workloads. Parallel between multiple sequences,
	// serial under a single sequence. If not set, according to random processing.
	// +nullable
	// +optional
	// +kubebuilder:validation:Optional
	// Sequences []Sequence `json:"sequences,omitempty" protobuf:"bytes,9,opt,name=sequences,omitempty"`
}

// ProfitTarget the target of profit
type ProfitTarget struct {
	// Node defines the node num that maximum expect
	Node int32 `json:"node,omitempty" protobuf:"bytes,1,opt,name=node"`
	// Cpu defines the cpu num that maximum expects.
	// unit m
	Cpu int32 `json:"cpu,omitempty" protobuf:"bytes,2,opt,name=cpu"`
	// Mem defines the memory num that maximum expects
	// unit M
	Mem int32 `json:"mem,omitempty" protobuf:"bytes,3,opt,name=mem"`
}

// Sequence defines the workload migrate sequence
type Sequence []SequenceItem

// SequenceItem defines the workload item
type SequenceItem struct {
	// Namespace workload namespace, e.g: kube-system
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,1,opt,name=namespace"`
	// Name workload name, e.g: test
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// Kind the kind of workload, such as deployment/statefulset/gamedeployment/gamestatefulset
	Kind string `json:"kind,omitempty" protobuf:"bytes,1,opt,name=kind"`
}

// DescheduleBalanceStrategy balance strategy
type DescheduleBalanceStrategy struct{}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DeschedulePolicy defines the policy that deschedule
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:resource:shortName=dspolcy
// +kubebuilder:printcolumn:name="DISABLED",type=boolean,JSONPath=".spec.converge.disabled"
// +kubebuilder:printcolumn:name="TIME",type=string,JSONPath=".spec.converge.timeRange"
// +kubebuilder:printcolumn:name="HIGH WATER",type=string,JSONPath=".spec.converge.highWaterLevel"
// +kubebuilder:printcolumn:name="LOW WATER",type=string,JSONPath=".spec.converge.lowWaterLevel"
type DeschedulePolicy struct {
	metav1.TypeMeta `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`

	// Standard object metadata.
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,2,opt,name=metadata"`

	// +kubebuilder:validation:Required
	Spec DescheduleSpec `json:"spec" protobuf:"bytes,3,opt,name=spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DeschedulePolicyList defines the list of DeschedulePolicy
type DeschedulePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Items is a list of deschedule policy objects
	Items []DeschedulePolicy `json:"items,omitempty" protobuf:"bytes,2,opt,name=items"`
}
