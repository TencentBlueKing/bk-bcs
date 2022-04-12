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
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	// LabelValueForIngressName label value for ingress name
	LabelValueForIngressName = "ingress-name"
	// LabelKeyForLoadbalanceID label key for loadbalance id
	LabelKeyForLoadbalanceID = "ingress.bkbcs.tencent.com/lbid"
	// LabelKeyForLoadbalanceRegion label key for loadbalance region
	LabelKeyForLoadbalanceRegion = "ingress.bkbcs.tencent.com/lbregion"
	// LabelKeyForIsSegmentListener label key for if it is segment listener
	LabelKeyForIsSegmentListener = "segmentlistener.bkbcs.tencent.com"
	// LabelKeyForPortPoolListener label key for if it is listener for port pool
	LabelKeyForPortPoolListener = "portpool.bkbcs.tencent.com"
	// LabelValueTrue label value for true
	LabelValueTrue = "true"
	// LabelValueFalse label value for false
	LabelValueFalse = "false"
	// LabelValueForPortPoolItemName label value for port pool and item name
	LabelValueForPortPoolItemName = "portpool-item-name"
	// ListenerStatusNotSynced shows listener changes are not synced
	ListenerStatusNotSynced = "NotSynced"
	// ListenerStatusSynced shows listener changes are synced
	ListenerStatusSynced = "Synced"
)

// ListenerBackend info for backend
type ListenerBackend struct {
	IP     string `json:"IP"`
	Port   int    `json:"port"`
	Weight int    `json:"weight"`
}

// ListenerBackendList listener backend list
type ListenerBackendList []ListenerBackend

// Len implements sort interface
func (lbl ListenerBackendList) Len() int {
	return len(lbl)
}

// Swap implements sort interface
func (lbl ListenerBackendList) Swap(i, j int) {
	lbl[i], lbl[j] = lbl[j], lbl[i]
}

// Less implements sort interface
func (lbl ListenerBackendList) Less(i, j int) bool {
	if lbl[i].IP < lbl[j].IP {
		return true
	}
	if lbl[i].IP == lbl[j].IP {
		return lbl[i].Port < lbl[j].Port
	}
	return false
}

// ListenerTargetGroup backend set for listener
type ListenerTargetGroup struct {
	TargetGroupID       string            `json:"targetGroupID,omitempty"`
	TargetGroupName     string            `json:"targetGroupName,omitempty"`
	TargetGroupProtocol string            `json:"protocol,omitempty"`
	Backends            []ListenerBackend `json:"backends,omitempty"`
}

// ListenerRule route rule for listener
type ListenerRule struct {
	RuleID            string                    `json:"ruleID,omitempty"`
	Domain            string                    `json:"domain,omitempty"`
	Path              string                    `json:"path,omitempty"`
	ListenerAttribute *IngressListenerAttribute `json:"listenerAttribute,omitempty"`
	TargetGroup       *ListenerTargetGroup      `json:"targetGroup,omitempty"`
}

// ListenerRuleList list of listener rule
type ListenerRuleList []ListenerRule

// Len implements sort interface
func (lrl ListenerRuleList) Len() int {
	return len(lrl)
}

// Less implements sort interface
func (lrl ListenerRuleList) Less(i, j int) bool {
	if lrl[i].Domain < lrl[j].Domain {
		return true
	}
	if lrl[i].Domain == lrl[j].Domain {
		return lrl[i].Path < lrl[j].Path
	}
	return false
}

// Swap implements sort interface
func (lrl ListenerRuleList) Swap(i, j int) {
	lrl[i], lrl[j] = lrl[j], lrl[i]
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
	Rules             []ListenerRule              `json:"rules,omitempty"`
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
	Domain   string                        `json:"domain"`
	URL      string                        `json:"path"`
	Backends []ListenerBackendHealthStatus `json:"backends,omitempty"`
}

// ListenerHealthStatus health status of listener
type ListenerHealthStatus struct {
	RulesHealth []ListenerRuleHealthStatus `json:"rules,omitempty"`
}

// ListenerStatus defines the observed state of Listener
type ListenerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ListenerID   string                `json:"listenerID,omitempty"`
	Status       string                `json:"status,omitempty"`
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

// ToJSONString convert listener to json string
func (l *Listener) ToJSONString() string {
	data, err := json.Marshal(l)
	if err != nil {
		return ""
	}
	return string(data)
}

// ListenerSlice slice for listener for sort
type ListenerSlice []Listener

// Len implements sort interface
func (ls ListenerSlice) Len() int {
	return len(ls)
}

// Less implements sort interface
func (ls ListenerSlice) Less(i, j int) bool {
	return ls[i].Spec.Port < ls[j].Spec.Port
}

// Swap implements sort interface
func (ls ListenerSlice) Swap(i, j int) {
	ls[i], ls[j] = ls[j], ls[i]
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
