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

// IngressSubset subset info
type IngressSubset struct {
	LabelSelector map[string]string `json:"labelSelector"`
	Weight        int               `json:"weight"`
}

// ServiceRoute service info
type ServiceRoute struct {
	ServiceName      string           `json:"serviceName"`
	ServiceNamespace string           `json:"serviceNamespace"`
	ServicePort      int              `json:"servicePort"`
	IsDirectConnect  bool             `json:"isDirectConnect"`
	ServiceWeight    int              `json:"serviceWeight"`
	Subsets          []*IngressSubset `json:"subsets"`
}

// Layer7Route 7 layer route config
type Layer7Route struct {
	Domain   string          `json:"domain"`
	Path     string          `json:"path"`
	Services []*ServiceRoute `json:"services"`
}

// ListenerHealthCheck health check setting for listener
type ListenerHealthCheck struct {
	Enabled         bool   `json:"enabled"`
	Timeout         int    `json:"timeout"`
	IntervalTime    int    `json:"intervalTime"`
	HealthNum       int    `json:"healthNum"`
	UnHealthNum     int    `json:"unHealthNum"`
	HTTPCode        int    `json:"httpCode"`
	HTTPCheckPath   string `json:"httpCheckPath"`
	HTTPCheckMethod string `json:"httpCheckMethod"`
}

// IngressListenerAttribute attribute for listener
type IngressListenerAttribute struct {
	SessionTime int                  `json:"sessionTime"`
	LbPolicy    string               `json:"lbPolicy"`
	HealthCheck *ListenerHealthCheck `json:"healthCheck"`
}

// IngressListenerCertificate certificate configs for listener
type IngressListenerCertificate struct {
	Mode                string `json:"mode,omitempty"`
	CertID              string `json:"certId,omitempty"`
	CertCaID            string `json:"certCaId,omitempty"`
	CertServerName      string `json:"certServerName,omitempty"`
	CertServerKey       string `json:"certServerKey,omitempty"`
	CertServerContent   string `json:"certServerContent,omitempty"`
	CertClientCaName    string `json:"certClientCaName,omitempty"`
	CertClientCaContent string `json:"certCilentCaContent,omitempty"`
}

// IngressRule rule of ingress
type IngressRule struct {
	Port              int                         `json:"port"`
	Protocol          string                      `json:"protocol"`
	ListenerAttribute *IngressListenerAttribute   `json:"listenerAttribute"`
	Certificate       *IngressListenerCertificate `json:"certificates"`
	Services          []*ServiceRoute             `json:"layer4Services"`
	Routes            []*Layer7Route              `json:"layer7Services"`
}

// IngressPortMapping mapping of ingress
type IngressPortMapping struct {
	WorkloadKind       string                      `json:"workloadKind"`
	WorkloadName       string                      `json:"workloadName"`
	WorkloadNamespace  string                      `json:"workloadNamespace"`
	StartPort          int                         `json:"startPort"`
	StartIndex         int                         `json:"startIndex"`
	EndIndex           int                         `json:"endIndex"`
	SegmentLength      int                         `json:"segmentLength"`
	Protocol           string                      `json:"protocol"`
	Domain             string                      `json:"domain"`
	Path               string                      `json:"path"`
	IsBackendPortFixed bool                        `json:"isBackendPortFixed"`
	IgnoreHostPort     bool                        `json:"ignoreHostPort"`
	ListenerAttribute  *IngressListenerAttribute   `json:"listenerAttribute"`
	Certificate        *IngressListenerCertificate `json:"certificates"`
}

// IngressSpec defines the desired state of Ingress
type IngressSpec struct {
	Rules        []*IngressRule        `json:"rules"`
	PortMappings []*IngressPortMapping `json:"portMappings"`
}

// IngressLoadBalancer loadbalancer for ingress
type IngressLoadBalancer struct {
	IPs []string `json:"ips"`
}

// IngressStatus defines the observed state of Ingress
type IngressStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Loadbalancers []*IngressLoadBalancer `json:"loadbalancers"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// Ingress is the Schema for the ingresses API
type Ingress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IngressSpec   `json:"spec,omitempty"`
	Status IngressStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// IngressList contains a list of Ingress
type IngressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ingress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ingress{}, &IngressList{})
}
