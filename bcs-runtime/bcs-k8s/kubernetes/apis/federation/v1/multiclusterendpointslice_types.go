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
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
const (
	LabelKeyEpsRelatedServiceNamespace = "federation.bkbcs.tencent.com/multi-cluster-service-namespace"
	LabelKeyEpsRelatedServiceName      = "federation.bkbcs.tencent.com/multi-cluster-service-name"
	LabelKeyEpsRelatedClusterName      = "federation.bkbcs.tencent.com/cluster-name"
)

// AddressType represents the type of address referred to by an endpoint.
// +enum
type AddressType string

// EndpointPort represents a Port used by an EndpointSlice
// +structType=atomic
type EndpointPort struct {
	// name represents the name of this port. All ports in an EndpointSlice must have a unique name.
	// If the EndpointSlice is derived from a Kubernetes service, this corresponds to the Service.ports[].name.
	// Name must either be an empty string or pass DNS_LABEL validation:
	// * must be no more than 63 characters long.
	// * must consist of lower case alphanumeric characters or '-'.
	// * must start and end with an alphanumeric character.
	// Default is empty string.
	Name *string `json:"name,omitempty" protobuf:"bytes,1,name=name"`

	// protocol represents the IP protocol for this port.
	// Must be UDP, TCP, or SCTP.
	// Default is TCP.
	Protocol *k8scorev1.Protocol `json:"protocol,omitempty" protobuf:"bytes,2,name=protocol"`

	// port represents the port number of the endpoint.
	// If this is not specified, ports are not restricted and must be
	// interpreted in the context of the specific consumer.
	Port *int32 `json:"port,omitempty" protobuf:"bytes,3,opt,name=port"`

	// The application protocol for this port.
	// This is used as a hint for implementations to offer richer behavior for protocols that they understand.
	// This field follows standard Kubernetes label syntax.
	// Valid values are either:
	//
	// * Un-prefixed protocol names - reserved for IANA standard service names (as per
	// RFC-6335 and https://www.iana.org/assignments/service-names).
	//
	// * Kubernetes-defined prefixed names:
	//   * 'kubernetes.io/h2c' - HTTP/2 prior knowledge over cleartext as described in https://www.rfc-editor.org/rfc/rfc9113.html#name-starting-http-2-with-prior-
	//   * 'kubernetes.io/ws'  - WebSocket over cleartext as described in https://www.rfc-editor.org/rfc/rfc6455
	//   * 'kubernetes.io/wss' - WebSocket over TLS as described in https://www.rfc-editor.org/rfc/rfc6455
	//
	// * Other protocols should use implementation-defined prefixed names such as
	// mycompany.com/my-custom-protocol.
	// +optional
	AppProtocol *string `json:"appProtocol,omitempty" protobuf:"bytes,4,name=appProtocol"`

	// hostPort represents the hostport of the pod using bcs randhosport plugin
	// +optional
	HostPort *int32 `json:"hostPort,omitempty" protobuf:"bytes,5,opt,name=hostPort"`
}

// EndpointConditions represents the current condition of an endpoint.
type EndpointConditions struct {
	// ready indicates that this endpoint is prepared to receive traffic,
	// according to whatever system is managing the endpoint. A nil value
	// indicates an unknown state. In most cases consumers should interpret this
	// unknown state as ready. For compatibility reasons, ready should never be
	// "true" for terminating endpoints, except when the normal readiness
	// behavior is being explicitly overridden, for example when the associated
	// Service has set the publishNotReadyAddresses flag.
	// +optional
	Ready *bool `json:"ready,omitempty" protobuf:"bytes,1,name=ready"`

	// serving is identical to ready except that it is set regardless of the
	// terminating state of endpoints. This condition should be set to true for
	// a ready endpoint that is terminating. If nil, consumers should defer to
	// the ready condition.
	// +optional
	Serving *bool `json:"serving,omitempty" protobuf:"bytes,2,name=serving"`

	// terminating indicates that this endpoint is terminating. A nil value
	// indicates an unknown state. Consumers should interpret this unknown state
	// to mean that the endpoint is not terminating.
	// +optional
	Terminating *bool `json:"terminating,omitempty" protobuf:"bytes,3,name=terminating"`
}

// MultiClusterEndpointSliceSpec defines the desired state of MultiClusterEndpointSlice
type MultiClusterEndpointSliceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	AddressType AddressType `json:"addressType"`

	Endpoints []MultiClusterEndpointSliceEd `json:"endpoints"`
}

// MultiClusterEndpointSliceEd defines the endpoint of MultiClusterEndpointSlice
type MultiClusterEndpointSliceEd struct {
	// addresses of this endpoint. The contents of this field are interpreted
	// according to the corresponding EndpointSlice addressType field. Consumers
	// must handle different types of addresses in the context of their own
	// capabilities. This must contain at least one address but no more than
	// 100. These are all assumed to be fungible and clients may choose to only
	// use the first element. Refer to: https://issue.k8s.io/106267
	// +listType=set
	Addresses []string `json:"addresses" protobuf:"bytes,1,rep,name=addresses"`

	// conditions contains information about the current status of the endpoint.
	Conditions EndpointConditions `json:"conditions,omitempty" protobuf:"bytes,2,opt,name=conditions"`

	// hostname of this endpoint. This field may be used by consumers of
	// endpoints to distinguish endpoints from each other (e.g. in DNS names).
	// Multiple endpoints which use the same hostname should be considered
	// fungible (e.g. multiple A values in DNS). Must be lowercase and pass DNS
	// Label (RFC 1123) validation.
	// +optional
	Hostname *string `json:"hostname,omitempty" protobuf:"bytes,3,opt,name=hostname"`

	// targetRef is a reference to a Kubernetes object that represents this
	// endpoint.
	// +optional
	TargetRef *k8scorev1.ObjectReference `json:"targetRef,omitempty" protobuf:"bytes,4,opt,name=targetRef"`

	// deprecatedTopology contains topology information part of the v1beta1
	// API. This field is deprecated, and will be removed when the v1beta1
	// API is removed (no sooner than kubernetes v1.24).  While this field can
	// hold values, it is not writable through the v1 API, and any attempts to
	// write to it will be silently ignored. Topology information can be found
	// in the zone and nodeName fields instead.
	// +optional
	DeprecatedTopology map[string]string `json:"deprecatedTopology,omitempty" protobuf:"bytes,5,opt,name=deprecatedTopology"`

	// nodeName represents the name of the Node hosting this endpoint. This can
	// be used to determine endpoints local to a Node.
	// +optional
	NodeName *string `json:"nodeName,omitempty" protobuf:"bytes,6,opt,name=nodeName"`

	// NodeAddresses represents the ip addresses of the Node hosting this endpoints.
	// +optional
	NodeAddresses []string `json:"nodeAddresses,omitempty" protobuf:"bytes,7,opt,name=nodeAddresses"`

	Ports []EndpointPort `json:"ports,omitempty" protobuf:"bytes,8,opt,name=ports"`
}

// MultiClusterEndpointSliceStatus defines the observed state of MultiClusterEndpointSlice
type MultiClusterEndpointSliceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// MultiClusterEndpointSlice is the Schema for the multiclusterendpointslice API
type MultiClusterEndpointSlice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MultiClusterEndpointSliceSpec   `json:"spec,omitempty"`
	Status MultiClusterEndpointSliceStatus `json:"status,omitempty"`
}

// GetRelatedServiceNameSpace return related multi-cluster service namespace from label
func (in *MultiClusterEndpointSlice) GetRelatedServiceNameSpace() string {
	if in.Labels == nil {
		return ""
	}

	namespace, ok := in.Labels[LabelKeyEpsRelatedServiceNamespace]
	if !ok {
		return ""
	}

	return namespace
}

// GetRelatedServiceName return related multi-cluster service name from label
func (in *MultiClusterEndpointSlice) GetRelatedServiceName() string {
	if in.Labels == nil {
		return in.Name
	}

	name, ok := in.Labels[LabelKeyEpsRelatedServiceName]
	if !ok {
		return ""
	}

	return name
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// MultiClusterEndpointSliceList contains a list of MultiClusterEndpointSlice
type MultiClusterEndpointSliceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MultiClusterEndpointSlice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MultiClusterEndpointSlice{}, &MultiClusterEndpointSliceList{})
}
