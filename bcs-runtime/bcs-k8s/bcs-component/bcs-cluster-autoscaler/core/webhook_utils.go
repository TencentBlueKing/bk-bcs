/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	simulator "k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/kubelet/types"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	contextinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/context"
)

// GenerateAutoscalerRequest generates requests based on current states of node groups
func GenerateAutoscalerRequest(nodeGroups []cloudprovider.NodeGroup,
	upcomingNodes map[string]int) (*AutoscalerRequest, error) {
	localNgs := make(map[string]*NodeGroup)
	for _, ng := range nodeGroups {
		localNg, err := generateNodeGroup(ng, upcomingNodes)
		if err != nil {
			return nil, err
		}
		localNgs[localNg.NodeGroupID] = localNg
	}
	req := &AutoscalerRequest{
		UID:        apitypes.UID(uuid.New().String()),
		NodeGroups: localNgs,
	}
	return req, nil
}

func generateNodeGroup(nodeGroup cloudprovider.NodeGroup,
	upcomingNodes map[string]int) (*NodeGroup, error) {
	targetSize, err := nodeGroup.TargetSize()
	if err != nil {
		return nil, fmt.Errorf("failed to get target size of nodegroup %v: %v", nodeGroup.Id(), err)
	}
	nodes, err := nodeGroup.Nodes()
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes of nodegroup %v: %v", nodeGroup.Id(), err)
	}
	template, err := nodeGroup.TemplateNodeInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get node template of nodegroup %v: %v", nodeGroup.Id(), err)
	}
	ips := make([]string, 0)
	for _, n := range nodes {
		ips = append(ips, n.Id)
	}

	return &NodeGroup{
		NodeGroupID:  nodeGroup.Id(),
		MaxSize:      nodeGroup.MaxSize(),
		MinSize:      nodeGroup.MinSize(),
		DesiredSize:  targetSize,
		UpcomingSize: upcomingNodes[nodeGroup.Id()],
		NodeTemplate: Template{
			CPU:    template.AllocatableResource().MilliCPU / 1000,
			Mem:    template.AllocatableResource().Memory,
			GPU:    template.AllocatableResource().ScalarResources[gpu.ResourceNvidiaGPU],
			Labels: template.Node().Labels,
			Taints: template.Node().Spec.Taints,
		},
		NodeIPs: ips,
	}, nil
}

// HandleResponse abstracts options of scale up and candidates of scale down from response
func HandleResponse(review ClusterAutoscalerReview, nodes []*corev1.Node,
	nodeNameToNodeInfo map[string]*schedulernodeinfo.NodeInfo,
	sd *ScaleDown) (ScaleUpOptions, ScaleDownCandidates, error) {
	var options ScaleUpOptions
	var candidates ScaleDownCandidates
	var err error

	if review.Response != nil && review.Response.ScaleUps != nil {
		options, err = handleScaleUpResponse(review.Request, review.Response.ScaleUps)
		if err != nil {
			return nil, nil, err
		}
	}

	if review.Response != nil && review.Response.ScaleDowns != nil {
		candidates, err = handleScaleDownResponse(review.Request, review.Response.ScaleDowns,
			nodes, nodeNameToNodeInfo, sd)
		if err != nil {
			return nil, nil, err
		}
	}

	return options, candidates, nil
}

func handleScaleUpResponse(req *AutoscalerRequest, policies []*ScaleUpPolicy) (ScaleUpOptions, error) {
	options := make(ScaleUpOptions, 0)
	if len(policies) <= 0 {
		return options, nil
	}
	for _, policy := range policies {
		originNodeGroup, ok := req.NodeGroups[policy.NodeGroupID]
		if !ok {
			return nil, fmt.Errorf("Cannot find node group info in requests for %s", policy.NodeGroupID)
		}
		switch {
		case policy.DesiredSize < 0:
			return nil, fmt.Errorf("Desired size %d cannot be negative for node group %s",
				policy.DesiredSize, policy.NodeGroupID)
		case policy.DesiredSize > originNodeGroup.MaxSize:
			return nil, fmt.Errorf("Desired size %d should less than node group %s 's max size %d",
				policy.DesiredSize, policy.NodeGroupID, originNodeGroup.MaxSize)
		case policy.DesiredSize < originNodeGroup.MinSize:
			return nil, fmt.Errorf("Desired size %d should greater than node group %s 's min size %d",
				policy.DesiredSize, policy.NodeGroupID, originNodeGroup.MinSize)
		case policy.DesiredSize < originNodeGroup.DesiredSize:
			return nil, fmt.Errorf("Desired size %d should greater than node group %s 's desired size %d when scale up",
				policy.DesiredSize, policy.NodeGroupID, originNodeGroup.DesiredSize)
		case policy.DesiredSize == originNodeGroup.DesiredSize:
			continue
		default:
			options[policy.NodeGroupID] = policy.DesiredSize
		}
	}
	return options, nil
}

func handleScaleDownResponse(req *AutoscalerRequest, policies []*ScaleDownPolicy, nodes []*corev1.Node,
	nodeNameToNodeInfo map[string]*schedulernodeinfo.NodeInfo, sd *ScaleDown) (ScaleDownCandidates, error) {
	candidates := make(ScaleDownCandidates, 0)
	if len(policies) <= 0 {
		return candidates, nil
	}
	for _, policy := range policies {
		originNodeGroup, ok := req.NodeGroups[policy.NodeGroupID]
		if !ok {
			return nil, fmt.Errorf("Cannot find node group info in requests for %s", policy.NodeGroupID)
		}
		switch policy.Type {
		case NodeNumScaleDownType:
			if policy.NodeNum == originNodeGroup.DesiredSize {
				continue
			}
			if policy.NodeNum > originNodeGroup.DesiredSize {
				return nil, fmt.Errorf("In scale down policy of nodegroup %v, node num %d should not greater than desired num %d",
					policy.NodeGroupID, policy.NodeNum, originNodeGroup.DesiredSize)
			}
			// 节点缩容时有短暂时间获取不到 InternalIP，但此时 desiredSize 还没变小，因此以 NodeIPs 长度为准 double check
			if policy.NodeNum == len(originNodeGroup.NodeIPs) {
				continue
			}
			if policy.NodeNum > len(originNodeGroup.NodeIPs) {
				return nil, fmt.Errorf("In scale down policy of nodegroup %v, node num %d should not greater than len(NodeIPs) %v",
					policy.NodeGroupID, policy.NodeNum, len(originNodeGroup.NodeIPs))
			}
			ips, err := sortNodesWithCostAndUtilization(nodes, originNodeGroup.NodeIPs, nodeNameToNodeInfo, sd)
			if err != nil {
				return nil, fmt.Errorf("Sort nodes with cost and utilization failed: %v", err)
			}
			// 缩容中的节点可能出现不在 ips，但在 DeletionsInProgress 的情况，所以可能这一逻辑周期会少缩，但下一周期会继续处理缩容，此处保守处理
			scaleDownNum := len(ips) - policy.NodeNum - sd.nodeDeletionTracker.GetDeletionsInProgress(policy.NodeGroupID)
			if scaleDownNum <= 0 {
				continue
			}
			if scaleDownNum > len(ips) {
				return nil, fmt.Errorf("Get candidates for nodegroup %v failed, scaleDownNum %v should not"+
					" greater than len(ips) %v", policy.NodeGroupID, scaleDownNum, len(ips))
			}
			candidates = append(candidates, ips[:scaleDownNum]...)
		case NodeIPsScaleDownType:
			ips := intersect(originNodeGroup.NodeIPs, policy.NodeIPs)
			if originNodeGroup.DesiredSize-len(ips) < originNodeGroup.MinSize {
				return nil, fmt.Errorf("Cannot scale down node group %v to %d after scaling down %d nodes, the min size is %d",
					originNodeGroup.NodeGroupID, originNodeGroup.DesiredSize-len(ips), len(ips), originNodeGroup.MinSize)
			}
			candidates = append(candidates, ips...)
		default:
			klog.Infof("Scale down type \"%v\" is not supprted", policy.Type)
			continue
		}

	}
	return candidates, nil

}

func intersect(slice1, slice2 []string) []string {
	m := make(map[string]int)
	n := make([]string, 0)
	for _, v := range slice1 {
		m[v]++
	}

	for _, v := range slice2 {
		times := m[v]
		if times == 1 {
			n = append(n, v)
		}
	}
	return n
}

func sortNodesWithCostAndUtilization(nodes []*corev1.Node, candidates []string,
	nodeNameToNodeInfo map[string]*schedulernodeinfo.NodeInfo, sd *ScaleDown) ([]string, error) {
	nodeToUtilInfo := make(map[string]simulator.UtilizationInfo)
	nodeToCost := make(map[string]float64)
	for i := range nodes {
		node := nodes[i]
		ip, found, err := checkCandidates(node, candidates)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		nodeInfo, found := nodeNameToNodeInfo[node.Name]
		if !found {
			return nil, fmt.Errorf("Node info for %s not found", node.Name)
		}
		utilInfo, err := simulator.CalculateUtilization(node, nodeInfo, sd.context.IgnoreDaemonSetsUtilization,
			sd.context.IgnoreMirrorPodsUtilization, sd.context.CloudProvider.GPULabel())
		if err != nil {
			return nil, fmt.Errorf("Failed to calculate utilization for %s: %v", node.Name, err)
		}
		nodeToUtilInfo[ip] = utilInfo
		cost := getCostFromNode(node)
		nodeToCost[ip] = cost
	}
	sort.Slice(candidates, func(i, j int) bool {
		if nodeToCost[candidates[i]] != nodeToCost[candidates[j]] {
			return nodeToCost[candidates[i]] < nodeToCost[candidates[j]]
		}
		return nodeToUtilInfo[candidates[i]].Utilization < nodeToUtilInfo[candidates[j]].Utilization
	})
	return candidates, nil
}

// ExecuteScaleUp execute scale up with scale up options
func ExecuteScaleUp(context *contextinternal.Context, clusterStateRegistry *clusterstate.ClusterStateRegistry,
	options ScaleUpOptions) error {
	nodegroups := context.CloudProvider.NodeGroups()
	for _, ng := range nodegroups {
		desired, ok := options[ng.Id()]
		if !ok {
			continue
		}
		target, err := ng.TargetSize()
		if err != nil {
			return fmt.Errorf("Cannot get target size of nodegroup %v, err: %v", ng.Id(), err)
		}
		info := nodegroupset.ScaleUpInfo{
			Group:       ng,
			CurrentSize: target,
			NewSize:     desired,
			MaxSize:     ng.MaxSize(),
		}
		err = executeScaleUp(context.AutoscalingContext, clusterStateRegistry, info, "", time.Now())
		if err != nil {
			return fmt.Errorf("Failed to scale up nodegroup %v to %v: %v", ng.Id(), desired, err)
		}
		klog.Infof("Successfully scale up , setting nodegroup %v size to %v", ng.Id(), desired)
	}
	clusterStateRegistry.Recalculate()

	return nil
}

// ExecuteScaleDown execute scale down with scale down candidates
func ExecuteScaleDown(context *contextinternal.Context, sd *ScaleDown,
	nodes []*corev1.Node, candidates ScaleDownCandidates,
	nodeNameToNodeInfo map[string]*schedulernodeinfo.NodeInfo) error {

	scaleDownNodes := make([]string, 0)
	defer func() {
		if len(scaleDownNodes) > 0 {
			sd.context.LogRecorder.Eventf(corev1.EventTypeNormal, "ScaleDown", "Scale-down: removing %d nodes"+
				" based on webhook response: %v", len(scaleDownNodes), scaleDownNodes)
		}
	}()

	for i := range nodes {
		node := nodes[i]
		// whether is under deleting
		if hasToBeDeletedTaint(node.Spec.Taints) {
			klog.V(4).Infof("node %s is under deleting...", node.Name)
		}
		_, found, err := checkCandidates(node, candidates)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		// get corresponding node group
		ng, err := context.CloudProvider.NodeGroupForNode(node)
		if err != nil {
			return fmt.Errorf("Failed to find node group info for %v", node.Name)
		}
		if ng == nil || reflect.ValueOf(ng).IsNil() {
			klog.V(4).Infof("Skipping %s - no node group config", node.Name)
			continue
		}
		// double check
		size, err := ng.TargetSize()
		if err != nil {
			return fmt.Errorf("Failed to get target size of node group %v", ng.Id())
		}
		deletionsInProgress := sd.nodeDeletionTracker.GetDeletionsInProgress(ng.Id())
		if size-deletionsInProgress <= ng.MinSize() {
			klog.V(1).Infof("Skipping %s - node group min size reached", node.Name)
			continue
		}

		podsToRemove := simpleGetPodsToMove(nodeNameToNodeInfo[node.Name])
		klog.V(0).Infof("Scale-down: removing node %s based on webhook response", node.Name)
		scaleDownNodes = append(scaleDownNodes, node.Name)

		// Starting deletion.
		go func() {
			// Finishing the delete process once this goroutine is over.
			var result status.NodeDeleteResult
			defer func() { sd.nodeDeletionTracker.AddNodeDeleteResult(node.Name, result) }()

			result = sd.deleteNode(node, podsToRemove, ng)
			if result.ResultType != status.NodeDeleteOk {
				klog.Errorf("Failed to delete %s: %v", node.Name, result.Err)
				return
			}
			metrics.RegisterScaleDown(1, gpu.GetGpuTypeForMetrics(sd.context.CloudProvider.GPULabel(),
				sd.context.CloudProvider.GetAvailableGPUTypes(), node, ng),
				metrics.NodeScaleDownReason("webhook"))
		}()

	}
	return nil

}

func checkCandidates(node *corev1.Node, candidates ScaleDownCandidates) (string, bool, error) {
	// get internal IP
	if len(node.Status.Addresses) == 0 {
		return "", false, fmt.Errorf("Cannot get Address for node %v", node.Name)
	}
	ip := ""
	for _, ad := range node.Status.Addresses {
		if ad.Type == corev1.NodeInternalIP {
			ip = ad.Address
			break
		}
	}
	if ip == "" {
		return "", false, fmt.Errorf("Cannot get Internal IP for node %v", node.Name)
	}
	// check candidates
	found := false
	for _, candidate := range candidates {
		if candidate == ip {
			found = true
			break
		}
	}
	return ip, found, nil
}

func simpleGetPodsToMove(nodeInfo *schedulernodeinfo.NodeInfo) []*corev1.Pod {
	pods := []*corev1.Pod{}
	for _, pod := range nodeInfo.Pods() {
		if _, found := pod.ObjectMeta.Annotations[types.ConfigMirrorAnnotationKey]; found {
			continue
		}
		if pod.DeletionTimestamp != nil {
			continue
		}
		if controllerRef := metav1.GetControllerOf(pod); controllerRef.Kind == "DaemonSet" {
			continue
		}
		pods = append(pods, pod)
	}
	return pods
}

func hasToBeDeletedTaint(taints []corev1.Taint) bool {
	if len(taints) == 0 {
		return false
	}
	for _, taint := range taints {
		if taint.Key == deletetaint.ToBeDeletedTaint && taint.Effect == corev1.TaintEffectNoSchedule {
			return true
		}
	}
	return false
}
