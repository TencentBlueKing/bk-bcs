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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ServiceMonitorSpec defines the desired state of ServiceMonitor
type ServiceMonitorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Endpoints []Endpoint    `json:"endpoints,omitempty"`
	Selector  LabelSelector `json:"selector,omitempty"`
}

// LabelSelector selector for service
type LabelSelector struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

// Endpoint collecte enpoint path information
type Endpoint struct {
	Port     string              `json:"port,omitempty"`
	Path     string              `json:"path,omitempty"`
	Interval string              `json:"interval,omitempty"`
	Params   map[string][]string `json:"params,omitempty"`
}

// GetUuid key generator
func (s *ServiceMonitor) GetUuid() string {
	return fmt.Sprintf("%s.%s", s.Namespace, s.Name)
}

// GetSelector get k8s selector implementation
func (s *ServiceMonitor) GetSelector() (labels.Requirements, error) {
	rms := labels.Requirements{}
	for k, v := range s.Spec.Selector.MatchLabels {
		r, err := labels.NewRequirement(k, selection.Equals, []string{v})
		if err != nil {
			return nil, err
		}
		rms = append(rms, *r)
	}
	return rms, nil
}

// Match check selector match
func (s *ServiceMonitor) Match(labels map[string]string) bool {
	for k, v := range s.Spec.Selector.MatchLabels {
		val, ok := labels[k]
		if !ok || val != v {
			return false
		}
	}

	return true
}

// ServiceMonitorStatus defines the observed state of ServiceMonitor
type ServiceMonitorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// ServiceMonitor is the Schema for the servicemonitors API
type ServiceMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceMonitorSpec   `json:"spec,omitempty"`
	Status ServiceMonitorStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// ServiceMonitorList contains a list of ServiceMonitor
type ServiceMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceMonitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceMonitor{}, &ServiceMonitorList{})
}
