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
 */

package core

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	contextinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/context"
)

const (
	// WebMode gets responses via web
	WebMode = "Web"
	// ConfigMapMode gets responses via configmap
	ConfigMapMode = "ConfigMap"
)

// Webhook is the interface of webhook
type Webhook interface {
	DoWebhook(context *contextinternal.Context, clusterStateRegistry *clusterstate.ClusterStateRegistry,
		sd *ScaleDown, nodes []*apiv1.Node, pods []*apiv1.Pod) errors.AutoscalerError
	GetResponses(context *contextinternal.Context, clusterStateRegistry *clusterstate.ClusterStateRegistry,
		nodeNameToNodeInfo map[string]*schedulernodeinfo.NodeInfo, nodes []*apiv1.Node,
		sd *ScaleDown) (ScaleUpOptions, ScaleDownCandidates, error)
	ExecuteScale(context *contextinternal.Context, clusterStateRegistry *clusterstate.ClusterStateRegistry,
		sd *ScaleDown, nodes []*apiv1.Node, options ScaleUpOptions, candidates ScaleDownCandidates,
		nodeNameToNodeInfo map[string]*schedulernodeinfo.NodeInfo) error
}

// ClusterAutoscalerReview is passed to the webhook with a populated Request value,
// and then returned with a populated Response.
type ClusterAutoscalerReview struct {
	// Request is the request to webhook server
	Request *AutoscalerRequest `json:"request"`
	// Response is the response of webhook server
	Response *AutoscalerResponse `json:"response"`
}

// AutoscalerRequest defines the request to webhook server
type AutoscalerRequest struct {
	// UID is used for tracing the request and response.
	UID types.UID `json:"uid"`
	// NodeGroups contain information of node groups. Key is node group ID.
	NodeGroups map[string]*NodeGroup `json:"nodeGroups"`
}

// NodeGroup is the information of node group
type NodeGroup struct {
	// NodeGroupID is the ID of the node group
	NodeGroupID string `json:"nodeGroupID"`
	// MaxSize is the upper limit of the node group
	MaxSize int `json:"maxSize"`
	// MinSize is the lower limit of the node group
	MinSize int `json:"minSize"`
	// DesiredSize is the current size of the node group. It is possible that the
	// number is different from the number of nodes registered in Kuberentes.
	DesiredSize int `json:"desiredSize"`
	// UpcomingSize is the number that indicates how many nodes have not registered in
	// Kubernetes or have not been ready to be used.
	UpcomingSize int `json:"upcomingSize"`
	// DeletingSize is the number of nodes  in the node group that
	// are in the process of deletion
	DeletingSize int `json:"deletingSize"`
	// NodeTemplate is the template information of node in the node group
	NodeTemplate Template `json:"nodeTemplate"`
	// NodeIPs are the IP of nodes which belongs to the node group
	NodeIPs []string `json:"nodeIPs"`
	// Priority is the priority of node group
	Priority int `json:"priority"`
}

// Template is the information of node
type Template struct {
	// CPU is the CPU resource of node. The unit is core.
	CPU int64 `json:"cpu"`
	// Mem is the memory resource of node. The unit is Gi.
	Mem int64 `json:"mem"`
	// GPU is the GPU resource of node.
	GPU int64 `json:"gpu"`
	// Labels is the Labels of node.
	Labels map[string]string `json:"labels"`
	// Taint is the taints of node.
	Taints []apiv1.Taint `json:"taints"`
}

// AutoscalerResponse defines the response of webhook server
type AutoscalerResponse struct {
	// UID is used for tracing the request and response.
	// It should be same as it in the request.
	UID types.UID `json:"uid"`
	// ScaleUps are the policy of scale up.
	ScaleUps []*ScaleUpPolicy `json:"scaleUps"`
	// ScaleDowns are the policy of scale down.
	ScaleDowns []*ScaleDownPolicy `json:"scaleDowns"`
}

// String returns the string of response
func (res AutoscalerResponse) String() string {
	str := "\nScaleUps: \n"
	for _, up := range res.ScaleUps {
		str += fmt.Sprintf("nodegroup: %v, desiredsize: %v \n", up.NodeGroupID, up.DesiredSize)
	}
	str += "ScaleDowns: \n"
	for _, down := range res.ScaleDowns {
		str += fmt.Sprintf("nodegroup: %v, nodeNum: %v, ips: %v \n", down.NodeGroupID, down.NodeNum, down.NodeIPs)
	}
	return str
}

// ScaleUpPolicy defines the details of scaling up a node group
type ScaleUpPolicy struct {
	// NodeGroupID is the ID of node group
	NodeGroupID string `json:"nodeGroupID"`
	// DesiredSize is the desired size of node group
	DesiredSize int `json:"desiredSize"`
}

// ScaleDownPolicy defines the details of scaling down a node group
type ScaleDownPolicy struct {
	// NodeGroupID is the ID of node group
	NodeGroupID string `json:"nodeGroupID"`
	// Type decides the way to scale down nodes. Available values: [NodeNum, NodeIPs]
	Type ScaleDownType `json:"type"`
	// NodeIPs are the ip of nodes that should be scale down
	NodeIPs []string `json:"nodeIPs"`
	// NodeNum is the number of nodes that should be retained
	NodeNum int `json:"nodeNum"`
}

// ScaleDownType is the type of scale down
type ScaleDownType string

const (
	// NodeNumScaleDownType scales down nodes with specific number.
	NodeNumScaleDownType ScaleDownType = "NodeNum"
	// NodeIPsScaleDownType scales down nodes with specific ips.
	NodeIPsScaleDownType ScaleDownType = "NodeIPs"
)

// ScaleUpOptions are the scale up option of webhook
type ScaleUpOptions map[string]int

// ScaleDownCandidates are the scale down candidate of webhook
type ScaleDownCandidates []string
