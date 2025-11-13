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

package node

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	corev1 "k8s.io/api/core/v1"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// SyncClusterNodesAction action for sync cluster nodes
type SyncClusterNodesAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	k8sOp *clusterops.K8SOperator
	req   *cmproto.SyncClusterNodesRequest
	resp  *cmproto.SyncClusterNodesResponse
}

// NewSyncClusterNodesAction sync cluster nodes action
func NewSyncClusterNodesAction(model store.ClusterManagerModel, k8sOp *clusterops.K8SOperator) *SyncClusterNodesAction {
	return &SyncClusterNodesAction{
		model: model,
		k8sOp: k8sOp,
	}
}

func (l *SyncClusterNodesAction) validate() error {
	if err := l.req.Validate(); err != nil {
		return err
	}

	return nil
}

func (l *SyncClusterNodesAction) setResp(code uint32, msg string) {
	l.resp.Code = code
	l.resp.Message = msg
	l.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handles list nodes request
func (l *SyncClusterNodesAction) Handle(ctx context.Context, req *cmproto.SyncClusterNodesRequest,
	resp *cmproto.SyncClusterNodesResponse) {
	if req == nil || resp == nil {
		blog.Errorf("sync cluster nodes failed, req or resp is empty")
		return
	}
	l.ctx = ctx
	l.req = req
	l.resp = resp

	if err := l.validate(); err != nil {
		l.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := l.syncNodes(); err != nil {
		l.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	l.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// syncNodes sync cluster nodes
func (l *SyncClusterNodesAction) syncNodes() error {
	condM := make(operator.M)
	condM["clusterid"] = l.req.ClusterID
	cond := operator.NewLeafCondition(operator.Eq, condM)
	nodes, err := l.model.ListNode(l.ctx, cond, &options.ListOption{})
	if err != nil {
		blog.Errorf("list nodes in cluster %s failed, %s", l.req.ClusterID, err.Error())
		return err
	}

	k8sNodes, err := l.k8sOp.ListClusterNodes(l.ctx, l.req.ClusterID)
	if err != nil {
		blog.Errorf("ListClusterNodes %s failed, %s", l.req.ClusterID, err.Error())
		return err
	}

	// k8s不存在的删除，存在的创建
	exitNodes := make(map[string]struct{})
	deleteNodesIPs := []string{}
	createNodes := []*types.Node{}

	for _, node := range nodes {
		exitNodes[node.InnerIP] = struct{}{}
	}

	for _, node := range k8sNodes {
		ipv4, ipv6 := getNodeDualAddress(node)
		if _, ok := exitNodes[ipv4]; ok {
			delete(exitNodes, ipv4)
			continue
		}
		createNodes = append(createNodes, transK8sNodeToNode(l.req.ClusterID, ipv4, ipv6, node))
	}
	for ip := range exitNodes {
		deleteNodesIPs = append(deleteNodesIPs, ip)
	}

	if len(deleteNodesIPs) > 0 {
		err = l.model.DeleteClusterNodesByIPs(l.ctx, l.req.ClusterID, deleteNodesIPs)
		if err != nil {
			blog.Errorf("delete nodes in cluster %s failed, %s", l.req.ClusterID, err.Error())
			return err
		}
	}

	if len(createNodes) > 0 {
		err = l.model.CreateNodes(l.ctx, createNodes)
		if err != nil {
			blog.Errorf("create nodes in cluster %s failed, %s", l.req.ClusterID, err.Error())
			return err
		}
	}

	return nil
}

// transK8sNodeToNode trans k8s node to node
func transK8sNodeToNode(clusterID, ipv4, ipv6 string, k8sNode *corev1.Node) *types.Node {
	// 获取zones
	var nodeZone string
	zone, ok := k8sNode.Labels[utils.ZoneKubernetesFlag]
	if ok {
		nodeZone = zone
	}
	zone, ok = k8sNode.Labels[utils.ZoneTopologyFlag]
	if ok && nodeZone == "" {
		nodeZone = zone
	}

	return &types.Node{
		InnerIP:        ipv4,
		Status:         transNodeStatus("", k8sNode),
		NodeName:       k8sNode.Name,
		NodeID:         k8sNode.Spec.ProviderID,
		InstanceType:   k8sNode.Labels[utils.NodeInstanceTypeFlag],
		CPU:            0,
		Mem:            0,
		GPU:            0,
		ZoneID:         nodeZone,
		NodeGroupID:    "",
		ClusterID:      clusterID,
		VPC:            "",
		Region:         "",
		Passwd:         "",
		Zone:           0,
		DeviceID:       "",
		NodeTemplateID: "",
		NodeType:       "",
		InnerIPv6:      ipv6,
		ZoneName:       "",
		TaskID:         "",
		FailedReason:   "",
		ChargeType:     "",
		DataDiskNum:    0,
		IsGpuNode:      false,
	}
}

// 转换节点状态
func transNodeStatus(cmNodeStatus string, k8sNode *corev1.Node) string {
	if cmNodeStatus == common.StatusInitialization || cmNodeStatus == common.StatusAddNodesFailed ||
		cmNodeStatus == common.StatusDeleting || cmNodeStatus == common.StatusRemoveNodesFailed ||
		cmNodeStatus == common.StatusRemoveCANodesFailed {
		return cmNodeStatus
	}
	for _, v := range k8sNode.Status.Conditions {
		if v.Type != corev1.NodeReady {
			continue
		}
		if v.Status == corev1.ConditionTrue {
			if k8sNode.Spec.Unschedulable {
				return common.StatusNodeRemovable
			}
			return common.StatusRunning
		}
		return common.StatusNodeNotReady
	}

	return common.StatusNodeUnknown
}

// getNodeDualAddress get node dual address
func getNodeDualAddress(node *corev1.Node) (string, string) {
	ipv4s, ipv6s := utils.GetNodeIPAddress(node)
	return utils.SliceToString(ipv4s), utils.SliceToString(ipv6s)
}
