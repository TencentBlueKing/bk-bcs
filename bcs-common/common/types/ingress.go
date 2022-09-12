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

package types

// BcsIngress xxx
type BcsIngress struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata"`
	Spec       IngressSpec `json:"spec"`
}

// IngressSpec describes the Ingress the user wishes to exist.
type IngressSpec struct {
	LBGroup   string `json:"lbGroup"`
	ClusterID string `json:"clusterid"`

	// A list of host rules used to configure the Ingress. If unspecified, or
	// no rule matches, all traffic is sent to the default backend.
	Rules []IngressRule `json:"rules"`

	// TLS []IngressTLS
}

// IngressRule represents the rules mapping the paths under a specified host to
// the related backend services. Incoming requests are first evaluated for a host
// match, then routed to the backend associated with the matching IngressRuleValue.
type IngressRule struct {
	Kind        IngressRuleKind  `json:"kind"`
	Balance     BalanceType      `json:"balance"`
	MaxConn     int              `json:"maxConn"`
	HTTPIngress *HTTPIngressRule `json:"httpIngress"`
	TCPIngress  *TCPIngressRule  `json:"tcpIngress"`
}

// HTTPIngressRule Value is a list of http selectors pointing to backends.
// In the example: http://<host>/<path>?<searchpart> -> backend where
// where parts of the url correspond to RFC 3986, this resource will be used
// to match against everything after the last '/' and before the first '?'
// or '#'.
type HTTPIngressRule struct {
	// Host is the fully qualified domain name of a network host, as defined by RFC 3986.
	// Incoming requests are matched against the host before the IngressRuleValue.
	// If the host is unspecified, the Ingress routes all traffic based on the
	// specified IngressRuleValue.
	Host  string            `json:"host"`
	Paths []HTTPIngressPath `json:"paths"`
}

// BalanceType xxx
type BalanceType string

// IngressRuleKind xxx
type IngressRuleKind string

const (
	// HttpIngressKind xxx
	HttpIngressKind IngressRuleKind = "HTTP"
	// TCPIngressKind xxx
	TCPIngressKind IngressRuleKind = "TCP"

	// RoundrobinBalanceType xxx
	RoundrobinBalanceType BalanceType = "roundrobin"
	// SourceBalanceType xxx
	SourceBalanceType BalanceType = "source"
)

// HTTPIngressPath associates a path regex with a backend. Incoming urls matching
// the path are forwarded to the backend.
type HTTPIngressPath struct {
	// part of a URL as defined by RFC 3986. Paths must begin with
	// a '/'. If unspecified, the path defaults to a catch all sending
	// traffic to the backend.
	Path string `json:"path"`

	// Backend defines the referenced service endpoint to which the traffic
	// will be forwarded to.
	Backend []IngressBackend `json:"backend"`
}

// IngressBackend describes all endpoints for a given service and port.
type IngressBackend struct {
	ServiceName string `json:"serviceName"`
	ServicePort int32  `json:"servicePort"`
	Weight      int32  `json:"weight"`
}

// TCPIngressRule xxx
type TCPIngressRule struct {
	ListenPort int32            `json:"listenPort"`
	Backend    []IngressBackend `json:"backend"`
}
