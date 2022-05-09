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

const (
	// ProtocolHTTP name of http protocol
	ProtocolHTTP = "http"
	// ProtocolHTTPS name of https protocol
	ProtocolHTTPS = "https"
	// ProtocolTCP name of tcp protocol
	ProtocolTCP = "tcp"
	// ProtocolUDP name of udp protocol
	ProtocolUDP = "udp"

	// AnnotationKeyForLoadbalanceIDs annotation key for cloud lb ids
	AnnotationKeyForLoadbalanceIDs = "networkextension.bkbcs.tencent.com/lbids"

	// AnnotationKeyForLoadbalanceNames annotation key for cloud lb names
	AnnotationKeyForLoadbalanceNames = "networkextension.bkbcs.tencent.com/lbnames"

	// DefaultWeight default weight value
	DefaultWeight = 10

	// WorkloadKindStatefulset kind name of workload statefulset
	WorkloadKindStatefulset = "statefulset"
	// WorkloadKindGameStatefulset kind name of workload game statefulset
	WorkloadKindGameStatefulset = "gamestatefulset"

	// Pod CLB weight annotation key
	AnnotationKeyForLoadbalanceWeight      = "networkextension.bkbcs.tencent.com/clb-weight"
	AnnotationKeyForLoadbalanceWeightReady = "networkextension.bkbcs.tencent.com/clb-weight-ready"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// IngressWeight ingress weight struct
type IngressWeight struct {
	Value int `json:"value"`
}

// IngressSubset subset info
type IngressSubset struct {
	LabelSelector map[string]string `json:"labelSelector"`
	Weight        *IngressWeight    `json:"weight,omitempty"`
}

// GetWeight get weight of ingress subset
func (is *IngressSubset) GetWeight() int {
	if is.Weight == nil {
		return DefaultWeight
	}
	return is.Weight.Value
}

// ServiceRoute service info
type ServiceRoute struct {
	ServiceName      string `json:"serviceName"`
	ServiceNamespace string `json:"serviceNamespace"`
	ServicePort      int    `json:"servicePort"`
	// If specified, will use the hostport as backend's port
	// +optional
	HostPort        bool            `json:"hostPort,omitempty"`
	IsDirectConnect bool            `json:"isDirectConnect,omitempty"`
	Weight          *IngressWeight  `json:"weight,omitempty"`
	Subsets         []IngressSubset `json:"subsets,omitempty"`
}

// GetWeight get weight of service route
func (sr *ServiceRoute) GetWeight() int {
	if sr.Weight == nil {
		return DefaultWeight
	}
	return sr.Weight.Value
}

// Layer7Route 7 layer route config
type Layer7Route struct {
	Domain            string                    `json:"domain"`
	Path              string                    `json:"path,omitempty"`
	ListenerAttribute *IngressListenerAttribute `json:"listenerAttribute,omitempty"`
	Services          []ServiceRoute            `json:"services,omitempty"`
}

// ListenerHealthCheck health check setting for listener
type ListenerHealthCheck struct {
	Enabled             bool   `json:"enabled,omitempty"`
	Timeout             int    `json:"timeout,omitempty"`
	IntervalTime        int    `json:"intervalTime,omitempty"`
	HealthNum           int    `json:"healthNum,omitempty"`
	UnHealthNum         int    `json:"unHealthNum,omitempty"`
	HTTPCode            int    `json:"httpCode,omitempty"`
	HealthCheckPort     int    `json:"healthCheckPort,omitempty"`
	HealthCheckProtocol string `json:"healthCheckProtocol,omitempty"`
	// HTTPCodeValues specifies a set of HTTP response status codes of health check.
	// You can specify multiple values (for example, "200,202") or a range of values
	// (for example, "200-299"). https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2@v1.17.0/types#Matcher
	// +optional
	HTTPCodeValues  string `json:"httpCodeValues,omitempty"`
	HTTPCheckPath   string `json:"httpCheckPath,omitempty"`
	HTTPCheckMethod string `json:"httpCheckMethod,omitempty"`
}

// IngressListenerAttribute attribute for listener
type IngressListenerAttribute struct {
	SessionTime int    `json:"sessionTime,omitempty"`
	LbPolicy    string `json:"lbPolicy,omitempty"`
	// BackendInsecure specifies whether to enable insecure access to the backend.
	BackendInsecure bool `json:"backendInsecure,omitempty"`
	// aws targetGroup attributes, https://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_ModifyTargetGroupAttributes.html
	AWSAttributes []AWSAttribute       `json:"awsAttributes,omitempty"`
	HealthCheck   *ListenerHealthCheck `json:"healthCheck,omitempty"`
}

// AWSAttribute define aws target group attribute
type AWSAttribute struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// IngressListenerCertificate certificate configs for listener
type IngressListenerCertificate struct {
	Mode     string `json:"mode,omitempty"`
	CertID   string `json:"certID,omitempty"`
	CertCaID string `json:"certCaID,omitempty"`
}

// IngressRule rule of ingress
type IngressRule struct {
	Port              int                         `json:"port"`
	Protocol          string                      `json:"protocol"`
	ListenerAttribute *IngressListenerAttribute   `json:"listenerAttribute,omitempty"`
	Certificate       *IngressListenerCertificate `json:"certificate,omitempty"`
	Services          []ServiceRoute              `json:"services,omitempty"`
	Routes            []Layer7Route               `json:"layer7Routes,omitempty"`
}

// IngressPortMappingLayer7Route 7 layer route config for port mapping
type IngressPortMappingLayer7Route struct {
	Domain            string                    `json:"domain"`
	Path              string                    `json:"path,omitempty"`
	ListenerAttribute *IngressListenerAttribute `json:"listenerAttribute,omitempty"`
}

// IngressPortMapping mapping of ingress
type IngressPortMapping struct {
	WorkloadKind      string                          `json:"workloadKind"`
	WorkloadName      string                          `json:"workloadName"`
	WorkloadNamespace string                          `json:"workloadNamespace"`
	StartPort         int                             `json:"startPort"`
	RsStartPort       int                             `json:"rsStartPort,omitempty"`
	StartIndex        int                             `json:"startIndex"`
	EndIndex          int                             `json:"endIndex"`
	SegmentLength     int                             `json:"segmentLength,omitempty"`
	Protocol          string                          `json:"protocol"`
	IsRsPortFixed     bool                            `json:"isRsPortFixed,omitempty"`
	IgnoreSegment     bool                            `json:"ignoreSegment,omitempty"`
	HostPort          bool                            `json:"hostPort,omitempty"`
	ListenerAttribute *IngressListenerAttribute       `json:"listenerAttribute,omitempty"`
	Certificate       *IngressListenerCertificate     `json:"certificate,omitempty"`
	Routes            []IngressPortMappingLayer7Route `json:"routes,omitempty"`
}

// IngressSpec defines the desired state of Ingress
type IngressSpec struct {
	Rules        []IngressRule        `json:"rules,omitempty"`
	PortMappings []IngressPortMapping `json:"portMappings,omitempty"`
}

// IngressLoadBalancer loadbalancer for ingress
type IngressLoadBalancer struct {
	LoadbalancerName string   `json:"loadbalancerName,omitempty"`
	LoadbalancerID   string   `json:"loadbalancerID,omitempty"`
	Region           string   `json:"region,omitempty"`
	Type             string   `json:"type,omitempty"`
	IPs              []string `json:"ips,omitempty"`
	DNSName          string   `json:"dnsName,omitempty"`
	Scheme           string   `json:"scheme,omitempty"`
	AWSLBType        string   `json:"awsLBType,omitempty"`
}

// IngressStatus defines the observed state of Ingress
type IngressStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Loadbalancers []IngressLoadBalancer `json:"loadbalancers,omitempty"`
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
