/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

//go:generate mockgen -package mock -destination mock/mockcloud.go github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud LoadBalance

package cloud

import (
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/metrics"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	reqCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "bcs_network",
			Subsystem: "ingress_controller",
			Name:      "cloud_request_total",
			Help:      "request total counter",
		},
		[]string{"rpc", "errcode"},
	)

	respTimeSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "bcs_network",
			Subsystem: "ingress_controller",
			Name:      "cloud_response_time",
			Help:      "response time(ms) summary.",
		},
		[]string{"rpc"},
	)
)

const (
	// MetricAPISuccess api call is successful
	MetricAPISuccess = "success"
	// MetricAPIFailed api call is failed
	MetricAPIFailed = "failed"
	// BackendHealthStatusHealthy healthy status for backend health
	BackendHealthStatusHealthy = "Healthy"
	// BackendHealthStatusUnhealthy unhealthy status for backend health
	BackendHealthStatusUnhealthy = "Unhealthy"
	// BackendHealthStatusUnknown unknown status for backend health
	BackendHealthStatusUnknown = "Unknown"
)

var (
	// ErrLoadbalancerNotFound error that loadbalancer not found
	ErrLoadbalancerNotFound = fmt.Errorf("loadbalancer not found")
	// ErrListenerNotFound error that listener not found
	ErrListenerNotFound = fmt.Errorf("listener not found")
)

func init() {
	metrics.Registry.MustRegister(reqCounter, respTimeSummary)
}

// StatRequest report metrics for rpc requests
func StatRequest(rpc string, errcode string, inTime, outTime time.Time) int64 {
	reqCounter.With(prometheus.Labels{
		"rpc":     rpc,
		"errcode": errcode,
	}).Inc()

	cost := toMSTimestamp(outTime) - toMSTimestamp(inTime)
	respTimeSummary.With(prometheus.Labels{"rpc": rpc}).Observe(float64(cost))

	return cost
}

// toMSTimestamp converts time.Time to millisecond timestamp.
func toMSTimestamp(t time.Time) int64 {
	return t.UnixNano() / 1e6
}

// LoadBalanceObject lb object
type LoadBalanceObject struct {
	LbID   string   `json:"lbID"`
	Region string   `json:"region"`
	Name   string   `json:"name"`
	IPs    []string `json:"ips"`
	// LoadBalancerType OPEN or INTERNAL https://cloud.tencent.com/document/api/214/30694#LoadBalancer
	Type string `json:"type,omitempty"`
	// dns for lb
	DNSName string   `json:"dnsName,omitempty"`
	VIPs    []string `json:"vips,omitempty"`
	// LoadBalancerScheme define Internet-facing or Internal. An internet-facing load balancer routes
	// requests from clients to targets over the internet.
	// An internal load balancer routes requests to targets using private IP addresses.
	Scheme string `json:"scheme,omitempty"`
	// AWSLBType define aws lb type, application, network, or gateway
	AWSLBType string `json:"awsLBType,omitempty"`
}

// BackendHealthStatus health status of cloud loadbalancer backend
type BackendHealthStatus struct {
	ListenerID   string
	ListenerPort int
	Namespace    string
	IP           string
	Port         int
	Protocol     string
	Host         string
	Path         string
	Status       string
}

// LoadBalance interface for clb loadbalancer
type LoadBalance interface {
	// DescribeLoadBalancer get loadbalancer object by id or name
	DescribeLoadBalancer(region, lbID, name string) (*LoadBalanceObject, error)

	// DescribeLoadBalancerWithNs get loadbalancer object by id or name with namespace specified
	DescribeLoadBalancerWithNs(ns, region, lbID, name string) (*LoadBalanceObject, error)

	// IsNamespaced if client is namespaced
	IsNamespaced() bool

	// EnsureListener ensure listener to cloud, and get listener info
	EnsureListener(region string, listener *networkextensionv1.Listener) (string, error)

	// DeleteListener delete listener by name
	DeleteListener(region string, listener *networkextensionv1.Listener) error

	// EnsureMultiListeners ensure multiple listeners to cloud
	EnsureMultiListeners(region, lbID string, listeners []*networkextensionv1.Listener) (map[string]string, error)

	// DeleteMultiListeners delete multiple listeners
	DeleteMultiListeners(region, lbID string, listeners []*networkextensionv1.Listener) error

	// EnsureSegmentListener ensure segment listener
	EnsureSegmentListener(region string, listener *networkextensionv1.Listener) (string, error)

	// EnsureMultiSegmentListeners ensure multi segment listener
	EnsureMultiSegmentListeners(
		region, lbID string, listeners []*networkextensionv1.Listener) (map[string]string, error)

	// DeleteSegmentListener delete segment listener
	DeleteSegmentListener(region string, listener *networkextensionv1.Listener) error

	// DescribeBackendStatus describe backend status
	DescribeBackendStatus(region, ns string, lbIDs []string) (map[string][]*BackendHealthStatus, error)
}

// Validater validate parameter for cloud loadbalancer
type Validater interface {
	// IsIngressValid check bcs ingress parameter
	IsIngressValid(ingress *networkextensionv1.Ingress) (isValid bool, msg string)

	// CheckNoConflictsInIngress return true, if there is no conflicts in ingress itself
	CheckNoConflictsInIngress(ingress *networkextensionv1.Ingress) (isConflicts bool, msg string)
}
