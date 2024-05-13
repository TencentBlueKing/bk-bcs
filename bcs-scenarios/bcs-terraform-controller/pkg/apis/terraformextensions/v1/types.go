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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	// TerraformFinalizer 标志
	TerraformFinalizer = "finalizer.terraform.bkbcs.tencent.com"

	// TerraformOperationSync defines sync the operation of terraform cr
	TerraformOperationSync = "terraformextensions.sync.bkbcs.tencent.com"
	// TerraformOperationClean defines the clean operation of terraform cr
	TerraformOperationClean = "terraformextensions.clean.bkbcs.tencent.com"

	// ManualSyncPolicy 手动策略
	ManualSyncPolicy = "manual"

	// AutoSyncPolicy 自动策略
	AutoSyncPolicy = "auto-sync"

	// OutOfSyncStatus 不同步
	OutOfSyncStatus = "OutOfSync"

	// SyncedStatus 已经同步
	SyncedStatus = "Synced"

	// PhaseSucceeded 成功
	PhaseSucceeded = "Succeeded"

	// PhaseError 失败(报错)
	PhaseError = "Error"
)

// GitRepository is used to define the git warehouse address of bcs argo cd. bcs argo cd git仓库地址
type GitRepository struct {
	// Repo storage repo url.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +required
	Repo string `json:"repo,omitempty"`

	// Path terraform execute path.
	// +optional
	Path string `json:"path,omitempty"`

	// TargetRevision git commit or branch.
	// +optional
	TargetRevision string `json:"targetRevision,omitempty"`
}

// TerraformSpec defines the desired state of Terraform. Terraform对象声明清单
type TerraformSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Project gitops project
	Project string `json:"project,omitempty"`

	// SyncPolicy Synchronization strategy, divided into 'manual' and automatic synchronization
	// +kubebuilder:default:=manual
	// +kubebuilder:validation:Enum=manual;auto-sync
	// +optional
	SyncPolicy string `json:"syncPolicy,omitempty"`

	// Create destroy plan and apply it to destroy terraform resources
	// upon deletion of this object. Defaults to false.
	// +kubebuilder:default:=false
	// +optional
	DestroyResourcesOnDeletion bool `json:"destroyResourcesOnDeletion,omitempty"`

	// +optional
	Repository GitRepository `json:"repository,omitempty"`
}

// OperationStatus operation Terraform detail
type OperationStatus struct {
	// FinishAt operation Terraform finish time
	// +optional
	FinishAt *metav1.Time `json:"finishAt,omitempty"`

	// Message operation Terraform error message
	// +optional
	Message string `json:"message,omitempty"`

	// Phase operation Terraform status
	// +optional
	Phase string `json:"phase,omitempty"`
}

// ApplyHistory defines the history of apply
type ApplyHistory struct {
	ID         int          `json:"id,omitempty"`
	StartedAt  *metav1.Time `json:"startedAt,omitempty"`
	FinishedAt *metav1.Time `json:"finishedAt,omitempty"`
	Revision   string       `json:"revision,omitempty"`
}

// TerraformStatus defines the observed state of Terraform
type TerraformStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The last successfully applied revision.
	// The revision format for Git sources is <branch|tag>/<commit-sha>.
	// +optional
	LastAppliedRevision string `json:"lastAppliedRevision,omitempty"`

	// LastPlannedRevision is the revision used by the last planning process.
	// The result could be either no plan change or a new plan generated.
	// +optional
	LastPlannedRevision string `json:"lastPlannedRevision,omitempty"`

	// LastPlannedAt is the time when the last terraform plan was performed
	// +optional
	LastPlannedAt *metav1.Time `json:"lastPlannedAt,omitempty"`

	// LastAppliedAt is the time when the last drift was detected and
	// terraform apply was performed as a result
	// +optional
	LastAppliedAt *metav1.Time `json:"lastAppliedAt,omitempty"`

	// LastPlanError this is an error in terraform execution plan.
	// +optional
	LastPlanError string `json:"lastPlanError,omitempty"`

	// LastApplyError this is an error in terraform execution apply.
	// +optional
	LastApplyError string `json:"lastApplyError,omitempty"`

	// SyncStatus terraform sync statu
	// +optional
	SyncStatus string `json:"syncStatus,omitempty"`

	// OperationStatus operation Terraform detail
	OperationStatus OperationStatus `json:"operationStatus,omitempty"`

	// History of apply, only set the current history, not set all history
	History ApplyHistory `json:"history,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Terraform is the Schema for the terraforms API
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="REPO",type=string,JSONPath=".spec.repository.repo"
// +kubebuilder:printcolumn:name="POLICY",type=string,JSONPath=".spec.syncPolicy"
// +kubebuilder:printcolumn:name="SYNC",type=string,JSONPath=".status.syncStatus"
// +kubebuilder:printcolumn:name="OPERATION",type=string,JSONPath=".status.operationStatus.phase"
// +kubebuilder:printcolumn:name="DESTROY",type=string,JSONPath=".spec.destroyResourcesOnDeletion"
// +kubebuilder:printcolumn:name="LAST APPLY",type=date,JSONPath=".status.lastAppliedAt"
// +kubebuilder:subresource:status
type Terraform struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TerraformSpec   `json:"spec,omitempty"`
	Status TerraformStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TerraformList contains a list of Terraform
type TerraformList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of secret objects.
	Items []Terraform `json:"items"`
}
