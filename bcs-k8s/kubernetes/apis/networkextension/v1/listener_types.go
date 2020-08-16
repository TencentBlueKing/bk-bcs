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
	// LabelValueForIngressName label value for ingress name
	LabelValueForIngressName = "ingress-name"
	// LabelValueForIngressNamespace label value for ingress namespace
	LabelValueForIngressNamespace = "ingress-namespace"
	// LabelKeyForLoadbalanceID label key for loadbalance id
	LabelKeyForLoadbalanceID = "ingress.bkbcs.tencent.com/lbid"
)

// ListenerBackend info for backend
type ListenerBackend struct {
	IP     string `json:"IP"`
	Port   int    `json:"port"`
	Weight int    `json:"weight"`
}

// ListenerTargetGroup backend set for listener
type ListenerTargetGroup struct {
	TargetGroupID       string             `json:"targetGroupID,omitempty"`
	TargetGroupName     string             `json:"targetGroupName,omitempty"`
	TargetGroupProtocol string             `json:"protocol,omitempty"`
	Backends            []*ListenerBackend `json:"backends,omitempty"`
}

// ListenerRule route rule for listener
type ListenerRule struct {
	RuleID      string               `json:"ruleID,omitempty"`
	Domain      string               `json:"domain,omitempty"`
	Path        string               `json:"path,omitempty"`
	TargetGroup *ListenerTargetGroup `json:"targetGroup,omitempty"`
}

// ListenerSpec defines the desired state of Listener
type ListenerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	LoadbalancerID    string                      `json:"loadbalancerID"`
	Port              int                         `json:"port"`
	EndPort           int                         `json:"endPort,omitempty"`
	Protocol          string                      `json:"protocol"`
	ListenerAttribute *IngressListenerAttribute   `json:"listenerAttribute,omitempty"`
	Certificate       *IngressListenerCertificate `json:"certificate,omitempty"`
	TargetGroup       *ListenerTargetGroup        `json:"targetGroup,omitempty"`
	Rules             []*ListenerRule             `json:"rules,omitempty"`
}

// ListenerBackendHealthStatus backend health status of listener
type ListenerBackendHealthStatus struct {
	IP                 string `json:"ip"`
	Port               int    `json:"port"`
	HealthStatus       bool   `json:"healthStatus"`
	TargetID           string `json:"targetID"`
	HealthStatusDetail string `json:"healthStatusDetail"`
}

// ListenerRuleHealthStatus rule health status of listener
type ListenerRuleHealthStatus struct {
	Domain   string                         `json:"domain"`
	URL      string                         `json:"path"`
	Backends []*ListenerBackendHealthStatus `json:"backends,omitempty"`
}

// ListenerHealthStatus health status of listener
type ListenerHealthStatus struct {
	RulesHealth []*ListenerRuleHealthStatus `json:"rules,omitempty"`
}

// ListenerStatus defines the observed state of Listener
type ListenerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ListenerID   string                `json:"listenerID,omitempty"`
	HealthStatus *ListenerHealthStatus `json:"healthStatus,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// Listener is the Schema for the listeners API
type Listener struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ListenerSpec   `json:"spec,omitempty"`
	Status ListenerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// ListenerList contains a list of Listener
type ListenerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Listener `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Listener{}, &ListenerList{})
}
