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
	"fmt"
	"reflect"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	// network type
	ClbNetworkTypePublic  = "public"
	ClbNetworkTypePrivate = "private"
	// lb policy
	ClbLBPolicyWRR       = "wrr"
	ClbLBPolicyLeastConn = "least_conn"
	ClbLBPolicyIPHash    = "ip_hash"
	// protocol
	ClbListenerProtocolHTTP  = "http"
	ClbListenerProtocolHTTPS = "https"
	ClbListenerProtocolTCP   = "tcp"
	ClbListenerProtocolUDP   = "udp"
	// tls
	ClbListenerTLSModeUniDirectional = "unidirectional"
	ClbListenerTLSModeMutual         = "mutual"
)

//Rule only use for http/https
type Rule struct {
	ID     string `json:"id,omitempty"`
	Domain string `json:"domain"`
	URL    string `json:"url"`
	//Balance     string       `json:"balance,omitempty"`
	TargetGroup *TargetGroup `json:"targetGroup"`
}

// NewRule new a rule
func NewRule(domain string, url string) *Rule {
	return &Rule{
		ID:          "",
		Domain:      domain,
		URL:         url,
		TargetGroup: NewTargetGroup("", "", "", 0),
	}
}

//IsEqual check if a Rule is equal to the other
func (r *Rule) IsEqual(other *Rule) bool {
	if r.Domain != other.Domain ||
		r.URL != other.URL {
		return false
	}
	return r.TargetGroup.IsEqual(other.TargetGroup)
}

//IsConflict check if a rule is conflict with the other
func (r *Rule) IsConflict(ruleList RuleList) bool {
	for _, other := range ruleList {
		if r.Domain == other.Domain && r.URL == other.URL {
			return true
		}
	}
	return false
}

//RuleList list of Rule
type RuleList []*Rule

// Len is the number of elements in the collection.
func (rl RuleList) Len() int {
	return len(rl)
}

// TargetGroupHealthCheck target group health check setting
type TargetGroupHealthCheck struct {
	Enabled       int    `json:"enabled"`
	Timeout       int    `json:"timeOut,omitempty"`
	IntervalTime  int    `json:"intervalTime,omitempty"`
	HealthNum     int    `json:"healthNum,omitempty"`
	UnHealthNum   int    `json:"unHealthNum,omitempty"`
	HTTPCode      int    `json:"httpCode,omitempty"`
	HTTPCheckPath string `json:"httpCheckPath,omitempty"`
}

// NewTargetGroupHealthCheck new target group health check
func NewTargetGroupHealthCheck() *TargetGroupHealthCheck {
	return &TargetGroupHealthCheck{
		Enabled:       1,
		IntervalTime:  5,
		Timeout:       2,
		HealthNum:     3,
		UnHealthNum:   3,
		HTTPCode:      31,
		HTTPCheckPath: "/",
	}
}

// TargetGroup backend info for all
type TargetGroup struct {
	ID       string `json:"id,omitempty"`       //elb arn
	Name     string `json:"name"`               //elb or clb name
	Protocol string `json:"protocol,omitempty"` //elb protocol info
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	Port          int `json:"port,omitempty"` //elb port
	SessionExpire int `json:"sessionExpire,omitempty"`
	//HealthCheckPath string      `json:"healthCheckPath,omitempty"` //need health check path for http and https
	HealthCheck *TargetGroupHealthCheck `json:"healthCheck,omitempty"`
	LBPolicy    string                  `json:"lbPolicy,omitempty"`
	Backends    BackendList             `json:"backends,omitempty"` //CVM instance backend
}

func NewTargetGroup(id, name, protocol string, port int) *TargetGroup {
	return &TargetGroup{
		ID:            id,
		Name:          name,
		Protocol:      protocol,
		Port:          port,
		SessionExpire: 0,
		HealthCheck:   NewTargetGroupHealthCheck(),
		LBPolicy:      ClbLBPolicyWRR,
		Backends:      make([]*Backend, 0),
	}
}

// IsAttrEqual if TargetGroup equal to another except for backends
func (tg *TargetGroup) IsAttrEqual(other *TargetGroup) bool {
	if tg == nil || other == nil {
		if tg == nil && other == nil {
			return true
		}
		return false
	}
	if tg.Name != other.Name ||
		tg.Protocol != other.Protocol ||
		tg.Port != other.Port ||
		tg.SessionExpire != other.SessionExpire ||
		tg.LBPolicy != other.LBPolicy ||
		!reflect.DeepEqual(tg.HealthCheck, other.HealthCheck) {
		return false
	}
	return true
}

// IsEqual if TargetGroup equal to another
func (tg *TargetGroup) IsEqual(other *TargetGroup) bool {

	if !tg.IsAttrEqual(other) {
		return false
	}

	adds, dels := tg.GetDiffBackend(other)
	updates := tg.GetUpdateBackend(other)
	if len(adds) == 0 && len(dels) == 0 && len(updates) == 0 {
		return true
	}
	return false
}

// GetDiffBackend check
func (tg *TargetGroup) GetDiffBackend(cur *TargetGroup) (dels BackendList, adds BackendList) {
	tmpMapAdd := make(map[string]*Backend)
	for _, tgBack := range tg.Backends {
		tmpMapAdd[tgBack.IP+strconv.Itoa(int(tgBack.Port))+strconv.Itoa(tgBack.Weight)] = tgBack
	}
	for _, curBack := range cur.Backends {
		if _, ok := tmpMapAdd[curBack.IP+strconv.Itoa(int(curBack.Port))+strconv.Itoa(curBack.Weight)]; !ok {
			adds = append(adds, curBack)
		}
	}

	tmpMapDel := make(map[string]*Backend)
	for _, curBack := range cur.Backends {
		tmpMapDel[curBack.IP+strconv.Itoa(int(curBack.Port))+strconv.Itoa(curBack.Weight)] = curBack
	}
	for _, tgBack := range tg.Backends {
		if _, ok := tmpMapDel[tgBack.IP+strconv.Itoa(int(tgBack.Port))+strconv.Itoa(tgBack.Weight)]; !ok {
			dels = append(dels, tgBack)
		}
	}
	return
}

// GetUpdateBackend get update backend from tg compared to cur
//** [NOT USED] to reduce api calling times, we do not use this function temporarily, to update backend, we delete it and create a new one **
func (tg *TargetGroup) GetUpdateBackend(cur *TargetGroup) (updates BackendList) {
	tmpMapAdd := make(map[string]*Backend)
	for _, tgBack := range tg.Backends {
		tmpMapAdd[tgBack.IP+strconv.Itoa(int(tgBack.Port))] = tgBack
	}
	for _, curBack := range cur.Backends {
		backendOld, ok := tmpMapAdd[curBack.IP+strconv.Itoa(int(curBack.Port))]
		if ok && !backendOld.IsEqual(curBack) {
			updates = append(updates, curBack)
		}
	}
	return
}

// AddBackends add backends to target group
func (tg *TargetGroup) AddBackends(adds BackendList) {
	tmp := make(map[string]*Backend)
	for _, bk := range tg.Backends {
		tmp[bk.IP+strconv.Itoa(int(bk.Port))] = bk
	}
	for _, add := range adds {
		if _, ok := tmp[add.IP+strconv.Itoa(int(add.Port))]; !ok {
			tg.Backends = append(tg.Backends, add)
		}
	}
}

// RemoveBackend clean all backend info
func (tg *TargetGroup) RemoveBackend(dels BackendList) {
	tmp := make(map[string]*Backend)
	for _, del := range dels {
		tmp[del.IP+strconv.Itoa(int(del.Port))] = del
	}
	var retList BackendList
	for _, bk := range tg.Backends {
		if _, ok := tmp[bk.IP+strconv.Itoa(int(bk.Port))]; !ok {
			retList = append(retList, bk)
		}
	}
	tg.Backends = retList
}

// Backend info for one service node
type Backend struct {
	IP string `json:"ip"`
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	Port   int `json:"port"`
	Weight int `json:"weight"`
}

func NewBackend(ip string, port int) *Backend {
	return &Backend{
		IP:     ip,
		Port:   port,
		Weight: 10,
	}
}

func (b *Backend) SetWeight(weight int) {
	b.Weight = weight
}

// IsEqual check if a backend is equal to the other
func (b *Backend) IsEqual(other *Backend) bool {
	if b == nil || other == nil {
		if b == nil && other == nil {
			return true
		}
		return false
	}
	if b.IP != other.IP ||
		b.Port != other.Port ||
		b.Weight != other.Weight {
		return false
	}
	return true
}

// BackendList sort for backend list
type BackendList []*Backend

// Len is the number of elements in the collection.
func (bl BackendList) Len() int {
	return len(bl)
}

//CloudLoadBalancer information for cloud Loadbalancer in aws/qcloud
type CloudLoadBalancer struct {
	ID          string   `json:"id"` //clb id or elb arn
	NetworkType string   `json:"net"`
	Name        string   `json:"name"`
	PublicIPs   []string `json:"publicIPs"`
	VIPS        []string `json:"vips"`
}

type CloudListenerTls struct {
	Mode                string `json:"mode,omitempty"`
	CertID              string `json:"certId,omitempty"`
	CertCaID            string `json:"certCaId,omitempty"`
	CertServerName      string `json:"certServerName,omitempty"`
	CertServerKey       string `json:"certServerKey,omitempty"`
	CertServerContent   string `json:"certServerContent,omitempty"`
	CertClientCaName    string `json:"certClientCaName,omitempty"`
	CertClientCaContent string `json:"certCilentCaContent,omitempty"`
}

// CloudListenerSpec defines the desired state of CloudListener
type CloudListenerSpec struct {
	ListenerID     string            `json:"listenerId"`    //clb listenerId/elb arn
	LoadBalancerID string            `json:"loadbalanceId"` //loadbalancer reference id
	Protocol       string            `json:"protocol"`      //service name
	TLS            *CloudListenerTls `json:"tls,omitempty"` //clb tls config
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	ListenPort int `json:"listenPort"` //export ports info
	//SSLCertID   string       `json:"sslCertId,omitempty"`   //SSL certificate Id for https
	TargetGroup *TargetGroup `json:"targetGroup,omitempty"` //only for tcp & udp
	Rules       RuleList     `json:"rules,omitempty"`       //only for http/https
}

// CloudListenerBackendHealthStatus backend health status of listener
type CloudListenerBackendHealthStatus struct {
	IP                 string `json:"ip"`
	Port               int    `json:"port"`
	HealthStatus       bool   `json:"healthStatus"`
	TargetID           string `json:"targetId"`
	HealthStatusDetail string `json:"healthStatusDetail"`
}

// CloudListenerRuleHealthStatus rule health status of listener
type CloudListenerRuleHealthStatus struct {
	Domain   string                              `json:"domain"`
	URL      string                              `json:"url"`
	Backends []*CloudListenerBackendHealthStatus `json:"backends,omitempty"`
}

// CloudListenerHealthStatus health status of listener
type CloudListenerHealthStatus struct {
	RulesHealth []*CloudListenerRuleHealthStatus `json:"rules,omitempty"`
}

// CloudListenerStatus defines the observed state of CloudListener
type CloudListenerStatus struct {
	// last updated timestamp
	LastUpdateTime metav1.Time                `json:"lastUpdateTime,omitempty"`
	HealthStatus   *CloudListenerHealthStatus `json:"healthStatus,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CloudListener is the Schema for the cloudlisteners API
// +k8s:openapi-gen=true
type CloudListener struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudListenerSpec   `json:"spec,omitempty"`
	Status CloudListenerStatus `json:"status,omitempty"`
}

// Key get key of cloudlistener
func (cl *CloudListener) Key() string {
	return cl.GetNamespace() + "/" + cl.GetName()
}

//ToString format object to string
func (cl *CloudListener) ToString() string {
	data, err := json.Marshal(cl)
	if err != nil {
		return ""
	}
	return string(data)
}

//IsEqual check if Listener is equal to the other
func (cl *CloudListener) IsEqual(other *CloudListener) bool {
	//check listener info
	//do not compare listener id, new listener struct has no id
	if cl.GetName() != other.GetName() ||
		cl.GetNamespace() != other.GetNamespace() ||
		cl.Spec.LoadBalancerID != other.Spec.LoadBalancerID ||
		cl.Spec.ListenPort != other.Spec.ListenPort ||
		cl.Spec.Protocol != other.Spec.Protocol {
		return false
	}
	//4 layer protocol do not use rules field
	if cl.Spec.Protocol == ClbListenerProtocolTCP || cl.Spec.Protocol == ClbListenerProtocolUDP {
		return cl.Spec.TargetGroup.IsEqual(other.Spec.TargetGroup)
	}
	if cl.Spec.Protocol == ClbListenerProtocolHTTPS &&
		!reflect.DeepEqual(cl.Spec.TLS, other.Spec.TLS) {
		return false
	}
	//7 layer listener check
	dels, adds := cl.GetDiffRules(other)
	updates, _ := cl.GetUpdateRules(other)
	if len(dels) == 0 && len(adds) == 0 && len(updates) == 0 {
		return true
	}
	return false
}

//GetDiffRules get deleted rules and added rules for cur listener compared to cl listener
func (cl *CloudListener) GetDiffRules(cur *CloudListener) (dels RuleList, adds RuleList) {
	tmpMapAdd := make(map[string]*Rule)
	for _, rule := range cl.Spec.Rules {
		tmpMapAdd[rule.Domain+rule.URL] = rule
	}
	for _, rule := range cur.Spec.Rules {
		if _, ok := tmpMapAdd[rule.Domain+rule.URL]; !ok {
			adds = append(adds, rule)
		}
	}

	tmpMapDel := make(map[string]*Rule)
	for _, rule := range cur.Spec.Rules {
		tmpMapDel[rule.Domain+rule.URL] = rule
	}
	for _, rule := range cl.Spec.Rules {
		if _, ok := tmpMapDel[rule.Domain+rule.URL]; !ok {
			dels = append(dels, rule)
		}
	}
	return
}

//GetUpdateRules get updated rules for cur listener compared to cl listener
func (cl *CloudListener) GetUpdateRules(cur *CloudListener) (olds RuleList, updates RuleList) {
	tmpMap := make(map[string]*Rule)
	for _, rule := range cl.Spec.Rules {
		tmpMap[rule.Domain+rule.URL] = rule
	}
	for _, rule := range cur.Spec.Rules {
		ruleOld, ok := tmpMap[rule.Domain+rule.URL]
		if ok {
			if !ruleOld.IsEqual(rule) {
				olds = append(olds, ruleOld)
				updates = append(updates, rule)
			}
		}
	}
	return
}

//GetRuleByID rule by rule id
func (cl *CloudListener) GetRuleByID(id string) (*Rule, error) {
	for _, rule := range cl.Spec.Rules {
		if rule.ID == id {
			return rule, nil
		}
	}
	return nil, fmt.Errorf("no rule in listener %s", cl.Spec.ListenerID)
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CloudListenerList contains a list of CloudListener
type CloudListenerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudListener `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudListener{}, &CloudListenerList{})
}
