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

// BcsDbPrivConfigSpec defines the desired state of BcsDbPrivConfig
type BcsDbPrivConfigSpec struct {
	PodSelector map[string]string `json:"podSelector"`
	AppName     string            `json:"appName"`
	TargetDb    string            `json:"targetDb"`
	DbType      string            `json:"dbType"`
	CallUser    string            `json:"callUser"`
	DbName      string            `json:"dbName"`
	Operator    string            `json:"operator"`
	UseCDP      bool              `json:"useCDP"`
}

// +kubebuilder:object:root=true
// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BcsDbPrivConfig is the Schema for the bcsdbprivconfigs API
type BcsDbPrivConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec BcsDbPrivConfigSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// BcsDbPrivConfigList contains a list of BcsDbPrivConfig
type BcsDbPrivConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BcsDbPrivConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BcsDbPrivConfig{}, &BcsDbPrivConfigList{})
}
