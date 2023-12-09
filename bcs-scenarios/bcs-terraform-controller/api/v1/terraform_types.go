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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	// TerraformFinalizer 标志
	TerraformFinalizer = "finalizer.terraform.bkbcs.tencent.com"

	// TerraformManualAnnotation 手动
	TerraformManualAnnotation = "terraformextesions.sync.bkbcs.tencent.com"

	// ManualSyncPolicy 手动策略
	// 同步策略: manual / auto-sync
	ManualSyncPolicy = "manual"

	// AutoSyncPolicy 自动策略
	// 同步策略: manual / auto-sync
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
	// Repo storage repo url.仓库url
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +required
	Repo string `json:"repo,omitempty"`

	// User storage user.用户名; 若是公开仓库, 不需要用户名和密码
	// +optional
	User string `json:"user,omitempty"`

	// Pass storage password.密码; 若是公开仓库, 不需要用户名和密码
	// +optional
	Pass string `json:"pass,omitempty"`

	// Path terraform execute path.执行路径
	// +optional
	Path string `json:"path,omitempty"`

	// TargetRevision git commit or branch.
	// +optional
	TargetRevision string `json:"targetRevision,omitempty"`
}

// BackendConfigsReference specify where to store backend config. 用于terraform初始化时，定义一些配置文件
type BackendConfigsReference struct {
	// Kind of the values referent, valid values are ('Secret', 'ConfigMap').
	// +kubebuilder:validation:Enum=Secret;ConfigMap
	// +required
	Kind string `json:"kind"`

	// Name of the configs referent. Should reside in the same namespace as the
	// referring resource.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +required
	Name string `json:"name"`

	// Keys is the data key where a specific value can be found at. Defaults to all keys.
	// +optional
	Keys []string `json:"keys,omitempty"`

	// Optional marks this BackendConfigsReference as optional. When set, a not found error
	// for the values reference is ignored, but any Key or
	// transient error will still result in a reconciliation failure.
	// +optional
	Optional bool `json:"optional,omitempty"`
}

// PlanStatus status of plan
type PlanStatus struct {
	// +optional
	LastApplied string `json:"lastApplied,omitempty"`

	// +optional
	Pending string `json:"pending,omitempty"`

	// +optional
	IsDestroyPlan bool `json:"isDestroyPlan,omitempty"`

	// +optional
	IsDriftDetectionPlan bool `json:"isDriftDetectionPlan,omitempty"`
}

// TerraformSpec defines the desired state of Terraform. Terraform对象声明清单
type TerraformSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ApprovePlan specifies name of a plan wanted to approve. 一个计划的审批, 如果该字段为auto, 则自动执行.
	// If its value is "auto", the controller will automatically approve every plan.
	// +optional
	ApprovePlan string `json:"approvePlan,omitempty"`

	// SyncPolicy Synchronization strategy, divided into 'manual' and automatic synchronization
	// 同步策略: manual / auto-sync
	// +kubebuilder:default:=manual
	// +optional
	SyncPolicy string `json:"syncPolicy,omitempty"`

	// Destroy produces a destroy plan. Applying the plan will destroy all resources.
	// +optional
	Destroy bool `json:"destroy,omitempty"`

	// +optional
	BackendConfigsFrom []BackendConfigsReference `json:"backendConfigsFrom,omitempty"`

	// Create destroy plan and apply it to destroy terraform resources
	// upon deletion of this object. Defaults to false.
	// +kubebuilder:default:=false
	// +optional
	DestroyResourcesOnDeletion bool `json:"destroyResourcesOnDeletion,omitempty"`

	// Targets specify the resource, module or collection of resources to target. 按模块执行
	// +optional
	Targets []string `json:"targets,omitempty"`

	// +optional
	Repository GitRepository `json:"repository,omitempty"`

	// Project bk project
	Project string `json:"project,omitempty"`
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

// TerraformStatus defines the observed state of Terraform
type TerraformStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ObservedGeneration is the last reconciled generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// The last successfully applied revision.
	// The revision format for Git sources is <branch|tag>/<commit-sha>.
	// +optional
	LastAppliedRevision string `json:"lastAppliedRevision,omitempty"`

	// LastAttemptedRevision is the revision of the last reconciliation attempt.
	// +optional
	LastAttemptedRevision string `json:"lastAttemptedRevision,omitempty"`

	// LastPlannedRevision is the revision used by the last planning process.
	// The result could be either no plan change or a new plan generated.
	// +optional
	LastPlannedRevision string `json:"lastPlannedRevision,omitempty"`

	// LastPlanAt is the time when the last terraform plan was performed
	// +optional
	LastPlanAt *metav1.Time `json:"lastPlanAt,omitempty"`

	// LastAppliedAt is the time when the last drift was detected and
	// terraform apply was performed as a result
	// +optional
	LastAppliedAt *metav1.Time `json:"lastAppliedAt,omitempty"`

	// LastPlanError this is an error in terraform execution plan.terraform执行plan报错信息字段
	// +optional
	LastPlanError string `json:"lastPlanError,omitempty"`

	// LastApplyError this is an error in terraform execution apply.terraform执行apply报错信息字段
	// +optional
	LastApplyError string `json:"lastApplyError,omitempty"`

	// SyncStatus terraform sync status.同步状态(OutOfSync/Synced)
	// +optional
	SyncStatus string `json:"syncStatus,omitempty"`

	// +optional
	Plan PlanStatus `json:"plan,omitempty"`

	// OperationStatus operation Terraform detail
	OperationStatus OperationStatus `json:"operationStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Terraform is the Schema for the terraforms API
type Terraform struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TerraformSpec   `json:"spec,omitempty"`
	Status TerraformStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TerraformList contains a list of Terraform
type TerraformList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of secret objects.
	Items []Terraform `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Terraform{}, &TerraformList{})
}
