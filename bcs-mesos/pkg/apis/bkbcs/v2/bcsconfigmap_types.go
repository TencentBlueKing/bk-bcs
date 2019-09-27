package v2

import (
	"bk-bcs/bcs-common/common/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BcsConfigMapSpec defines the desired state of BcsConfigMap
type BcsConfigMapSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	types.BcsConfigMap
}

// BcsConfigMapStatus defines the observed state of BcsConfigMap
type BcsConfigMapStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BcsConfigMap is the Schema for the bcsconfigmaps API
// +k8s:openapi-gen=true
type BcsConfigMap struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BcsConfigMapSpec   `json:"spec,omitempty"`
	Status BcsConfigMapStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BcsConfigMapList contains a list of BcsConfigMap
type BcsConfigMapList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BcsConfigMap `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BcsConfigMap{}, &BcsConfigMapList{})
}
