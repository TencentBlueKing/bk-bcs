/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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
	//ClusterId
	ClusterId string `json:"clusterId,omitempty"`
	//MeshType, default ISTIO
	Type MeshType `json:"type,omitempty"`
}

//mesh type: istio、tbuspp
type MeshType string
const (
	MeshIstio MeshType = "ISTIO"
	MeshTbuspp MeshType = "TBUSPP"
)

// MeshClusterStatus defines the observed state of MeshCluster
type MeshClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Individual status of each component controlled by the operator. The map key is the name of the component.
	ComponentStatus      map[string]*InstallStatus_VersionStatus `json:"componentStatus,omitempty"`
}

// VersionStatus is the status and version of a component.
type InstallStatus_VersionStatus struct {
	Name                 string               `json:"name,omitempty"`
	Namespace            string               `json:"namespace,omitempty"`
	Status               InstallStatus_Status `json:"status,omitempty"`
	Message              string               `json:"message,omitempty"`
}

// Status describes the current state of a component.
type InstallStatus_Status string

const (
	// Component is not present.
	InstallStatus_NONE InstallStatus_Status = "NONE"
	// Component is deploying now,
	InstallStatus_DEPLOY InstallStatus_Status = "DEPLOY"
	// Component is starting now,
	InstallStatus_STARTING InstallStatus_Status = "STARTING"
	// Component is running.
	InstallStatus_RUNNING InstallStatus_Status = "RUNNING"
	// Component is failed.
	InstallStatus_FAILED InstallStatus_Status = "FAILED"
	// Component is in updating.
	InstallStatus_UPDATE InstallStatus_Status = "UPDATE"
)

// +kubebuilder:object:root=true

// MeshCluster is the Schema for the MeshClusters API
type MeshCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MeshClusterSpec   `json:"spec,omitempty"`
	Status MeshClusterStatus `json:"status,omitempty"`
}

func (m *MeshCluster) GetUuid()string{
	return fmt.Sprintf("%s.%s",)
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
