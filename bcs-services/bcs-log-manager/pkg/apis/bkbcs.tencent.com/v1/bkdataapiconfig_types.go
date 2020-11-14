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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/esb/apigateway/bkdata"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BKDataApiConfigSpec defines the desired state of BKDataApiConfig
type BKDataApiConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ApiName                 string                              `json:"api_name"`
	AccessDeployPlanConfig  bkdata.CustomAccessDeployPlanConfig `json:"access_deploy_plan_config,omitempty"`
	DataCleanStrategyConfig bkdata.DataCleanStrategy            `json:"data_clean_strategy_config,omitempty"`
	Response                BKDataApiResponse                   `json:"response,omitempty"`
}

type BKDataApiResponse struct {
	Errors  string `json:"errors"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Data    string `json:"data"`
	Result  bool   `json:"result"`
}

// // BKDataApiConfigStatus defines the observed state of BKDataApiConfig
// type BKDataApiConfigStatus struct {
// 	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
// 	// Important: Run "make" to regenerate code after modifying this file
// }

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// BKDataApiConfig is the Schema for the bkdataapiconfigs API
type BKDataApiConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec BKDataApiConfigSpec `json:"spec,omitempty"`
	// Status BKDataApiConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// BKDataApiConfigList contains a list of BKDataApiConfig
type BKDataApiConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BKDataApiConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BKDataApiConfig{}, &BKDataApiConfigList{})
}
