package v2

import (
	"bk-bcs/bcs-common/common/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BcsCommandInfoSpec defines the desired state of BcsCommandInfo
type BcsCommandInfoSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	types.BcsCommandInfo
}

// BcsCommandInfoStatus defines the observed state of BcsCommandInfo
type BcsCommandInfoStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BcsCommandInfo is the Schema for the bcscommandinfos API
// +k8s:openapi-gen=true
type BcsCommandInfo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BcsCommandInfoSpec   `json:"spec,omitempty"`
	Status BcsCommandInfoStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BcsCommandInfoList contains a list of BcsCommandInfo
type BcsCommandInfoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BcsCommandInfo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BcsCommandInfo{}, &BcsCommandInfoList{})
}
