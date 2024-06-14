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
	// LabelKeyForUptimeCheckListener label key for if it is listener that enable uptime check
	LabelKeyForUptimeCheckListener = "uptime_check.bkbcs.tencent.com"
	// LabelKeyForSourceNamespace label key for namespace where original resources located
	LabelKeyForSourceNamespace = "source_namespace.bkbcs.tencent.com"
	// LabelValueTrue label value for true
	LabelValueTrue = "true"
	// LabelValueFalse label value for false
	LabelValueFalse = "false"
	// LabelValueForPortPoolItemName label value for port pool and item name
	LabelValueForPortPoolItemName = "portpool-item-name"
	// LabelKeyForOwnerKind mark which kind of resource generate this listener, e.g. portpool / ingress
	LabelKeyForOwnerKind = "owner-kind"
	// LabelKeyForOwnerName mark which resource generate this listener. Value is name of portpool or ingress.
	LabelKeyForOwnerName = "owner-name"
	// LabelKetForTargetGroupType label key for target group type
	LabelKetForTargetGroupType = "listener.bkbcs.tencent.com/target_group_type"
	// LabelValueForTargetGroupNormal normal target group
	LabelValueForTargetGroupNormal = "normal"
	// LabelValueForTargetGroupEmpty empty target group
	LabelValueForTargetGroupEmpty = "empty"
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
	RuleID            string                      `json:"ruleID,omitempty"`
	Domain            string                      `json:"domain,omitempty"`
	Path              string                      `json:"path,omitempty"`
	Certificate       *IngressListenerCertificate `json:"certificate,omitempty"`
	ListenerAttribute *IngressListenerAttribute   `json:"listenerAttribute,omitempty"`
	TargetGroup       *ListenerTargetGroup        `json:"targetGroup,omitempty"`
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

// UptimeCheckTaskStatus uptime check task status
type UptimeCheckTaskStatus struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

// ListenerStatus defines the observed state of Listener
type ListenerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ListenerID        string                 `json:"listenerID,omitempty"`
	Status            string                 `json:"status,omitempty"`
	HealthStatus      *ListenerHealthStatus  `json:"healthStatus,omitempty"`
	UptimeCheckStatus *UptimeCheckTaskStatus `json:"uptimeCheckStatus,omitempty"`
	Msg               string                 `json:"msg,omitempty"`
	PortPool          string                 `json:"portpool,omitempty"`
	Ingress           string                 `json:"ingress,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="protocol",type=string,JSONPath=`.spec.protocol`
// +kubebuilder:printcolumn:name="port",type=integer,JSONPath=`.spec.port`
// +kubebuilder:printcolumn:name="endPort",type=integer,JSONPath=`.spec.endPort`
// +kubebuilder:printcolumn:name="loadbalancerID",type=string,JSONPath=`.spec.loadbalancerID`
// +kubebuilder:printcolumn:name="ingress",type=string,JSONPath=`.status.ingress`
// +kubebuilder:printcolumn:name="portpool",type=string,JSONPath=`.status.portpool`

// Listener is the Schema for the listeners API
type Listener struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ListenerSpec   `json:"spec,omitempty"`
	Status ListenerStatus `json:"status,omitempty"`
}

// IsUptimeCheckEnable return true if uptime check enabled
func (l *Listener) IsUptimeCheckEnable() bool {
	if l.Spec.ListenerAttribute == nil || l.Spec.ListenerAttribute.UptimeCheck == nil || l.Spec.ListenerAttribute.
		UptimeCheck.Enabled == false {
		return false
	}
	return true
}

// GetUptimeCheckTaskStatus get uptime check task status
func (l *Listener) GetUptimeCheckTaskStatus() *UptimeCheckTaskStatus {
	if l.Status.UptimeCheckStatus == nil {
		return &UptimeCheckTaskStatus{}
	}

	return l.Status.UptimeCheckStatus
}

// GetListenerSourceNamespace 返回listener对应实例所在的命名空间，如果没有指定则使用listener的命名空间
func (l *Listener) GetListenerSourceNamespace() string {
	if l.Labels != nil {
		ns, exist := l.Labels[LabelKeyForSourceNamespace]
		if exist {
			return ns
		}
	}
	return l.Namespace
}

// IsEmptyTargetGroup return true if listener's target group is empty
func (l *Listener) IsEmptyTargetGroup() bool {
	if l.Spec.TargetGroup != nil && len(l.Spec.TargetGroup.Backends) != 0 {
		return false
	}
	for _, rule := range l.Spec.Rules {
		if rule.TargetGroup != nil && len(rule.TargetGroup.Backends) != 0 {
			return false
		}
	}
	return true
}

func (l *Listener) GetRegion() string {
	if l.GetLabels() == nil {
		return ""
	}
	region, ok := l.GetLabels()[LabelKeyForLoadbalanceRegion]
	if !ok {
		return ""
	}
	return region
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
