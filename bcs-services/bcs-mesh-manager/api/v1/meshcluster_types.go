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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MeshClusterSpec defines the desired state of MeshCluster
type MeshClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//version, istio version
	Version string `json:"version,omitempty"`
	//ClusterID
	ClusterID string `json:"clusterId,omitempty"`
	//MeshType, default ISTIO
	MeshType      MeshType `json:"type,omitempty"`
	Configuration []string `json:"configuration,omitempty"`
}

//MeshType mesh type: istio、tbuspp
type MeshType string

const (
	// MeshIstio istio type, now this is default
	MeshIstio MeshType = "ISTIO"
	// MeshTbuspp tbus type, feature support in future
	MeshTbuspp MeshType = "TBUSPP"
)

// MeshClusterStatus defines the observed state of MeshCluster
type MeshClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Individual status of each component controlled by the operator. The map key is the name of the component.
	ComponentStatus map[string]*ComponentState `json:"componentStatus,omitempty"`
}

// ComponentState VersionStatus is the status and version of a component.
type ComponentState struct {
	Name       string        `json:"name,omitempty"`
	Namespace  string        `json:"namespace,omitempty"`
	Status     InstallStatus `json:"status,omitempty"`
	Message    string        `json:"message,omitempty"`
	UpdateTime int64         `json:"updateTime,omitempty"`
}

// InstallStatus Status describes the current state of a component.
type InstallStatus string

const (
	// InstallStatusNONE Component is not present.
	InstallStatusNONE InstallStatus = "NONE"
	// InstallStatusDEPLOY Component is deploying now,
	InstallStatusDEPLOY InstallStatus = "DEPLOY"
	// InstallStatusSTARTING Component is starting now,
	InstallStatusSTARTING InstallStatus = "STARTING"
	// InstallStatusRUNNING Component is running.
	InstallStatusRUNNING InstallStatus = "RUNNING"
	// InstallStatusFAILED Component is failed.
	InstallStatusFAILED InstallStatus = "FAILED"
	// InstallStatusUPDATE Component is in updating.
	InstallStatusUPDATE InstallStatus = "UPDATE"
)

// +kubebuilder:object:root=true

// MeshCluster is the Schema for the MeshClusters API
type MeshCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MeshClusterSpec   `json:"spec,omitempty"`
	Status MeshClusterStatus `json:"status,omitempty"`
}

// GetUUID GetUuid
func (m *MeshCluster) GetUUID() string {
	return fmt.Sprintf("%s.%s", m.Namespace, m.Name)
}

// +kubebuilder:object:root=true

// MeshClusterList contains a list of MeshCluster
type MeshClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MeshCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MeshCluster{}, &MeshClusterList{})
}
