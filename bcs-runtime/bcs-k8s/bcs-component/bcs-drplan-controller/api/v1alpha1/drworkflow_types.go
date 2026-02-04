/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DRWorkflowSpec defines the desired state of DRWorkflow
type DRWorkflowSpec struct {
	// Executor is the execution engine configuration (reserved for extension)
	// +optional
	Executor *ExecutorConfig `json:"executor,omitempty"`

	// Parameters are parameter definitions
	// +optional
	Parameters []Parameter `json:"parameters,omitempty"`

	// Actions are the list of actions to execute in order
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Actions []Action `json:"actions"`

	// FailurePolicy defines how to handle action failures
	// +kubebuilder:validation:Enum=FailFast;Continue
	// +kubebuilder:default=FailFast
	// +optional
	FailurePolicy string `json:"failurePolicy,omitempty"`
}

// DRWorkflowStatus defines the observed state of DRWorkflow.
type DRWorkflowStatus struct {
	// Phase is the workflow phase: Ready, Invalid
	// +kubebuilder:validation:Enum=Ready;Invalid
	// +optional
	Phase string `json:"phase,omitempty"`

	// Conditions represent the current state of the DRWorkflow resource
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration reflects the generation observed by the controller
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DRWorkflow is the Schema for the drworkflows API
type DRWorkflow struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of DRWorkflow
	// +required
	Spec DRWorkflowSpec `json:"spec"`

	// status defines the observed state of DRWorkflow
	// +optional
	Status DRWorkflowStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// DRWorkflowList contains a list of DRWorkflow
type DRWorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []DRWorkflow `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DRWorkflow{}, &DRWorkflowList{})
}
