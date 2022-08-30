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

package nodegroup

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	v1 "k8s.io/api/core/v1"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

// ListAction action for list online cluster credential
type ListAction struct {
	ctx       context.Context
	model     store.ClusterManagerModel
	req       *cmproto.ListNodeGroupRequest
	resp      *cmproto.ListNodeGroupResponse
	groupList []*cmproto.NodeGroup
}

// NewListAction create list action for cluster credential
func NewListAction(model store.ClusterManagerModel) *ListAction {
	return &ListAction{
		model: model,
	}
}

func (la *ListAction) listNodeGroup() error {
	condM := make(operator.M)

	if len(la.req.Name) != 0 {
		condM["name"] = la.req.Name
	}
	if len(la.req.ClusterID) != 0 {
		condM["clusterid"] = la.req.ClusterID
	}
	if len(la.req.ProjectID) != 0 {
		condM["projectid"] = la.req.ProjectID
	}
	if len(la.req.Region) != 0 {
		condM["region"] = la.req.Region
	}

	cond := operator.NewLeafCondition(operator.Eq, condM)
	condStatus := operator.NewLeafCondition(operator.Ne, operator.M{"status": common.StatusDeleted})
	branchCond := operator.NewBranchCondition(operator.And, cond, condStatus)
	groups, err := la.model.ListNodeGroup(la.ctx, branchCond, &storeopt.ListOption{})
	if err != nil {
		return err
	}
	for i := range groups {
		la.groupList = append(la.groupList, removeSensitiveInfo(&groups[i]))
	}
	return nil
}

func (la *ListAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.groupList
}

// Handle handle list cluster credential
func (la *ListAction) Handle(
	ctx context.Context, req *cmproto.ListNodeGroupRequest, resp *cmproto.ListNodeGroupResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list NodeGroup failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listNodeGroup(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

// ListNodesAction action for list online cluster credential
type ListNodesAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	group   *cmproto.NodeGroup
	cluster *cmproto.Cluster
	cloud   *cmproto.Cloud
	k8sOp   *clusterops.K8SOperator
	req     *cmproto.ListNodesInGroupRequest
	resp    *cmproto.ListNodesInGroupResponse
}

// NewListNodesAction create list action for cluster credential
func NewListNodesAction(model store.ClusterManagerModel, k8sOp *clusterops.K8SOperator) *ListNodesAction {
	return &ListNodesAction{
		model: model,
		k8sOp: k8sOp,
	}
}

func (la *ListNodesAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (la *ListNodesAction) getRelativeResource() error {
	group, err := la.model.GetNodeGroup(la.ctx, la.req.NodeGroupID)
	if err != nil {
		blog.Errorf("get NodeGroup %s failed when list its all nodes, %s", la.req.NodeGroupID, err.Error())
		return err
	}
	la.group = group

	// get relative cluster for information injection
	cluster, err := la.model.GetCluster(la.ctx, la.group.ClusterID)
	if err != nil {
		blog.Errorf("can not get relative Cluster %s when list nodes in group", la.group.ClusterID)
		return fmt.Errorf("get relative cluster %s info err, %s", la.group.ClusterID, err.Error())
	}
	la.cluster = cluster

	cloud, err := actions.GetCloudByCloudID(la.model, la.group.Provider)
	if err != nil {
		blog.Errorf("can not get relative Cloud %s when list nodes in group for Cluster %s, %s",
			la.group.Provider, la.group.ClusterID, err.Error(),
		)
		return err
	}
	la.cloud = cloud

	return nil
}

func (la *ListNodesAction) listNodesInGroup() error {
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"nodegroupid": la.req.NodeGroupID,
	})
	nodes, err := la.model.ListNode(la.ctx, cond, &storeopt.ListOption{All: true})
	if err != nil {
		blog.Errorf("list group nodes in nodegroup %s for Cluster %s failed, %s",
			la.req.NodeGroupID, la.cluster.ClusterID, err.Error(),
		)
		return err
	}
	for _, v := range nodes {
		la.resp.Data = append(la.resp.Data, &cmproto.NodeGroupNode{
			NodeID:       v.NodeID,
			InnerIP:      v.InnerIP,
			InstanceType: v.InstanceType,
			CPU:          v.CPU,
			Mem:          v.Mem,
			GPU:          v.GPU,
			Status:       v.Status,
			ZoneID:       v.ZoneID,
			NodeGroupID:  v.NodeGroupID,
			ClusterID:    v.ClusterID,
			VPC:          v.VPC,
			Region:       v.Region,
			Zone:         v.Zone,
		})
	}

	// get node schedulable status
	if la.req.Output == "wide" {
		la.appendNodeInfo()
	}
	return nil
}

func (la *ListNodesAction) appendNodeInfo() {
	k8sNodes, err := la.k8sOp.ListClusterNodes(la.ctx, la.group.ClusterID)
	if err != nil {
		blog.Warnf("ListClusterNodes %s failed, %s", la.group.ClusterID, err.Error())
	}
	nodeMap := make(map[string]v1.Node)
	for _, v := range k8sNodes.Items {
		for _, addr := range v.Status.Addresses {
			if addr.Type == v1.NodeInternalIP {
				nodeMap[addr.Address] = v
			}
		}
	}
	for i := range la.resp.Data {
		if node, ok := nodeMap[la.resp.Data[i].InnerIP]; ok {
			la.resp.Data[i].UnSchedulable = 0
			if node.Spec.Unschedulable {
				// append unschedulable status
				la.resp.Data[i].UnSchedulable = 1
				// append REMOVABLE status
				if la.resp.Data[i].Status == common.StatusRunning {
					la.resp.Data[i].Status = common.StatusNodeRemovable
				}
			}
		}
	}
	return
}

// Handle handle list cluster credential
func (la *ListNodesAction) Handle(
	ctx context.Context, req *cmproto.ListNodesInGroupRequest, resp *cmproto.ListNodesInGroupResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list NodeGroup failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.getRelativeResource(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if err := la.listNodesInGroup(); err != nil {
		la.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
