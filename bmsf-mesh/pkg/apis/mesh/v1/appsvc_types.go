/*
Copyright (C) 2019 The BlueKing Authors. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AppSvcSpec defines the desired state of AppSvc
type AppSvcSpec struct {
	Selector map[string]string `json:"selector"`
	// service type, ClusterIP, Intergration or Empty
	// +kubebuilder:validation:Enum=ClusterIP,Intergration,None
	Type string `json:"type,omitempty"`
	// service version
	Version string `json:"version,omitempty"`
	// frontend represents service ip address, use for proxy or intergate
	Frontend []string `json:"frontend,omitempty"`
	// domain alias
	Alias string `json:"alias,omitempty"`
	// use for wan export
	WANIP        []string      `json:"wanip,omitempty"`
	ServicePorts []ServicePort `json:"ports"` //BcsService.Ports
}

//ServicePort port definition for application
type ServicePort struct {
	// name for service port
	// +kubebuilder:validation:MaxLength=100
	// +kubebuilder:validation:MinLength=2
	Name string `json:"name"`
	// protocol for service port
	// +kubebuilder:validation:Enum=tcp,udp,http
	Protocol string `json:"protocol"`
	// domain value for http proxy
	// +kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:MinLength=3
	Domain string `json:"domain,omitempty"`
	// http url path
	Path string `json:"path,omitempty"`
	// service port for all AppNode, ServicePort.Name == AppNode.Ports[i].Name
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	ServicePort int `json:"serviceport"`
	// proxy port for this Service Port if exist
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=0
	ProxyPort int `json:"proxyport,omitempty"`
}

// AppSvcStatus defines the observed state of AppSvc
type AppSvcStatus struct {
	// last updated timestamp
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppSvc is the Schema for the appsvcs API
// +k8s:openapi-gen=true
type AppSvc struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppSvcSpec   `json:"spec,omitempty"`
	Status AppSvcStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppSvcList contains a list of AppSvc
type AppSvcList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppSvc `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppSvc{}, &AppSvcList{})
}
