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
	"encoding/json"
	"fmt"
	"strings"
	"time"

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

type ClbTls struct {
	Mode                string `json:"mode,omitempty"`
	CertID              string `json:"certId,omitempty"`
	CertCaID            string `json:"certCaId,omitempty"`
	CertServerName      string `json:"certServerName,omitempty"`
	CertServerKey       string `json:"certServerKey,omitempty"`
	CertServerContent   string `json:"certServerContent,omitempty"`
	CertClientCaName    string `json:"certClientCaName,omitempty"`
	CertClientCaContent string `json:"certCilentCaContent,omitempty"`
}

func (tls *ClbTls) Validate() error {
	if tls.Mode == ClbListenerTLSModeUniDirectional {
		if len(tls.CertID) == 0 {
			if len(tls.CertServerKey) == 0 || len(tls.CertServerContent) == 0 || len(tls.CertServerName) == 0 {
				return fmt.Errorf("need (certId) or (certServerName, certServerKey, certServerContent)")
			}
		}
		return nil
	} else if tls.Mode == ClbListenerTLSModeMutual {
		if len(tls.CertID) == 0 {
			if len(tls.CertServerKey) == 0 || len(tls.CertServerContent) == 0 || len(tls.CertServerName) == 0 {
				return fmt.Errorf("in mutual mode, need (certId) or (certServerName, certServerKey, certServerContent)")
			}
		}
		if len(tls.CertCaID) == 0 {
			if len(tls.CertClientCaName) == 0 || len(tls.CertClientCaContent) == 0 {
				return fmt.Errorf("in mutual mode, need (certCaId) or (certClientCaName, certCilentCaContent)")
			}
		}
		return nil
	}
	return fmt.Errorf("tls mode invalid: must be [%s, %s]", ClbListenerTLSModeUniDirectional, ClbListenerTLSModeMutual)
}

type ClbBackendWeight struct {
	LabelSeletor map[string]string `json:"labelSelector"`
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:validation:Minimum=0
	Weight int `json:"weight"`
}

type ClbLoadBalance struct {
	Strategy       string             `json:"strategy"`
	BackendWeights []ClbBackendWeight `json:"backendWeights,omitempty"`
}

type ClbHealthCheck struct {
	Enabled bool `json:"enabled,omitempty"`
	// +kubebuilder:validation:Maximum=60
	// +kubebuilder:validation:Minimum=2
	Timeout int `json:"timeout,omitempty"`
	// +kubebuilder:validation:Maximum=300
	// +kubebuilder:validation:Minimum=5
	IntervalTime int `json:"intervalTime,omitempty"`
	// +kubebuilder:validation:Maximum=10
	// +kubebuilder:validation:Minimum=2
	HealthNum int `json:"healthNum,omitempty"`
	// +kubebuilder:validation:Maximum=10
	// +kubebuilder:validation:Minimum=2
	UnHealthNum int `json:"unHealthNum,omitempty"`
	// +kubebuilder:validation:Maximum=31
	// +kubebuilder:validation:Minimum=1
	HTTPCode int `json:"httpCode,omitempty"`
	// +kubebuilder:validation:MaxLength=80
	// +kubebuilder:validation:MinLength=1
	HTTPCheckPath string `json:"httpCheckPath,omitempty"`
	// HTTPCheckDomain string `json:"httpCheckDomain,omitempty"`
	// HTTPCheckMethod string `json:"httpCheckMethod,omitempty"`
}

func (hc *ClbHealthCheck) Validate4LayerConfig() error {
	if hc.Enabled {
		if hc.Timeout < 2 || hc.Timeout > 60 {
			return fmt.Errorf("health check timeout must be 2~60")
		}
		if hc.IntervalTime < 5 || hc.IntervalTime > 300 {
			return fmt.Errorf("health check interval time must be 5~300")
		}
		if hc.Timeout >= hc.IntervalTime {
			return fmt.Errorf("health check timeout must be small than interval time")
		}
		if hc.HealthNum < 2 || hc.HealthNum > 10 {
			return fmt.Errorf("health check health num must be 2~10")
		}
		if hc.UnHealthNum < 2 || hc.UnHealthNum > 10 {
			return fmt.Errorf("health check unhealth num must be 2~10")
		}
	}
	return nil
}

func (hc *ClbHealthCheck) Validate7LayerConfig() error {
	err := hc.Validate4LayerConfig()
	if err != nil {
		return err
	}
	if hc.Enabled {
		if hc.HTTPCode < 1 || hc.HTTPCode > 31 {
			return fmt.Errorf("health check http code must be 1~31")
		}
		if !strings.HasPrefix(hc.HTTPCheckPath, "/") {
			return fmt.Errorf("health check path must begin with /, but get %s", hc.HTTPCheckPath)
		}
	}
	return nil
}

type ClbHttpRule struct {
	// +kubebuilder:validation:MaxLength=80
	// +kubebuilder:validation:MinLength=1
	Host string `json:"host"`
	// +kubebuilder:validation:MaxLength=80
	// +kubebuilder:validation:MinLength=1
	Path    string  `json:"path"`
	TLS     *ClbTls `json:"tls,omitempty"`
	ClbRule `json:",inline"`
}

func (httpRule *ClbHttpRule) ValidateHTTP() error {
	err := httpRule.ClbRule.Validate()
	if err != nil {
		return err
	}
	if len(httpRule.Host) == 0 {
		return fmt.Errorf("host cannot be empty")
	}
	if len(httpRule.Path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}
	return nil
}

func (httpRule *ClbHttpRule) ValidateHTTPS() error {
	err := httpRule.Validate()
	if err != nil {
		return err
	}
	if httpRule.TLS == nil {
		return fmt.Errorf("https listener's tls config cannot be empty")
	}
	return httpRule.TLS.Validate()
}

// ToString convert ClbHttpRule to String
func (clbHttpRule *ClbHttpRule) ToString() string {
	str, err := json.Marshal(clbHttpRule)
	if err != nil {
		return ""
	}
	return string(str)
}

type ClbRule struct {
	ServiceName string `json:"serviceName"`
	Namespace   string `json:"namespace"`
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	ClbPort int `json:"clbPort"`
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	ServicePort int             `json:"servicePort"`
	LbPolicy    *ClbLoadBalance `json:"lbPolicy,omitempty"`
	HealthCheck *ClbHealthCheck `json:"healthCheck,omitempty"`
	// +kubebuilder:validation:Maximum=3600
	// +kubebuilder:validation:Minimum=30
	SessionTime int `json:"sessionTime,omitempty"`
}

func (clbRule *ClbRule) Validate() error {
	if len(clbRule.ServiceName) == 0 {
		return fmt.Errorf("serviceName cannot be empty")
	}
	if len(clbRule.Namespace) == 0 {
		return fmt.Errorf("namespace cannot be empty")
	}
	if clbRule.ClbPort < 1 || clbRule.ClbPort > 65535 {
		return fmt.Errorf("clbPort must be 1 ~ 65535")
	}
	if clbRule.ServicePort < 1 {
		return fmt.Errorf("servicePort cannot be small than 1")
	}
	if clbRule.LbPolicy != nil {
		if clbRule.LbPolicy.Strategy != ClbLBPolicyWRR &&
			clbRule.LbPolicy.Strategy != ClbLBPolicyLeastConn &&
			clbRule.LbPolicy.Strategy != ClbLBPolicyIPHash {
			return fmt.Errorf("invalid lb policy strategy %s", clbRule.LbPolicy.Strategy)
		}
	} else {
		clbRule.LbPolicy = &ClbLoadBalance{
			Strategy: ClbLBPolicyWRR,
		}
	}

	return nil
}

// ToString convert ClbRule to String
func (clbRule *ClbRule) ToString() string {
	str, err := json.Marshal(clbRule)
	if err != nil {
		return ""
	}
	return string(str)
}

// ClbStatefulSetHttPRule http rule for stateful set
type ClbStatefulSetHttpRule struct {
	StartPort     int `json:"startPort"`
	StartIndex    int `json:"startIndex,omitempty"`
	EndIndex      int `json:"endIndex,omitempty"`
	SegmentLength int `json:"segmentLength,omitempty"`
	ClbHttpRule   `json:",inline"`
}

// ClbStatefulSetRule rule for stateful Set
type ClbStatefulSetRule struct {
	StartPort     int `json:"startPort"`
	StartIndex    int `json:"startIndex,omitempty"`
	EndIndex      int `json:"endIndex,omitempty"`
	SegmentLength int `json:"segmentLength,omitempty"`
	ClbRule       `json:",inline"`
}

// ClbStatefulSet ingress for Stateful Set
type ClbStatefulSet struct {
	UDP   []*ClbStatefulSetRule     `json:"udp,omitempty"`
	TCP   []*ClbStatefulSetRule     `json:"tcp,omitempty"`
	HTTP  []*ClbStatefulSetHttpRule `json:"http,omitempty"`
	HTTPS []*ClbStatefulSetHttpRule `json:"https,omitempty"`
}

// ClbIngressSpec defines the desired state of ClbIngress
type ClbIngressSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	HTTP        []*ClbHttpRule  `json:"http,omitempty"`
	HTTPS       []*ClbHttpRule  `json:"https,omitempty"`
	TCP         []*ClbRule      `json:"tcp,omitempty"`
	UDP         []*ClbRule      `json:"udp,omitempty"`
	StatefulSet *ClbStatefulSet `json:"statefulset,omitempty"`
}

// ClbIngressStatus defines the observed state of ClbIngress
type ClbIngressStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status         string      `json:"status"`
	Message        string      `json:"message"`
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
}

const (
	// ClbIngressStatusNormal normal status for clb ingress
	ClbIngressStatusNormal = "Normal"
	// ClbIngressStatusAbnormal abnormal status for clb ingress
	ClbIngressStatusAbnormal = "Abnormal"
	// ClbIngressMessagePortConflict message for por conflict
	ClbIngressMessagePortConflict = "Port Conflict"
	// ClbIngressMessage
)

// SetStatusMessage set clb ingress status message
func (c *ClbIngress) SetStatusMessage(status, message string) {
	c.Status.Status = status
	c.Status.Message = message
	c.Status.LastUpdateTime = metav1.NewTime(time.Now())
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// ClbIngress is the Schema for the clbingresses API
type ClbIngress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClbIngressSpec   `json:"spec,omitempty"`
	Status ClbIngressStatus `json:"status,omitempty"`
}

// ToString convert ClbIngress to String
func (c *ClbIngress) ToString() string {
	str, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(str)
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// ClbIngressList contains a list of ClbIngress
type ClbIngressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClbIngress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClbIngress{}, &ClbIngressList{})
}
