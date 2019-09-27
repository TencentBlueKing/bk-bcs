package v2

import (
	"bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AgentSchedInfoSpec defines the desired state of AgentSchedInfo
type AgentSchedInfoSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	types.AgentSchedInfo
}

// AgentSchedInfoStatus defines the observed state of AgentSchedInfo
type AgentSchedInfoStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AgentSchedInfo is the Schema for the agentschedinfos API
// +k8s:openapi-gen=true
type AgentSchedInfo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentSchedInfoSpec   `json:"spec,omitempty"`
	Status AgentSchedInfoStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AgentSchedInfoList contains a list of AgentSchedInfo
type AgentSchedInfoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AgentSchedInfo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AgentSchedInfo{}, &AgentSchedInfoList{})
}
