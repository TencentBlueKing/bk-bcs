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

package nodegroup

import (
	"context"
	"fmt"
	"sort"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	v1 "k8s.io/api/core/v1"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	autils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// ListClusterGroupAction action for list online cluster credential
type ListClusterGroupAction struct {
	ctx       context.Context
	model     store.ClusterManagerModel
	req       *cmproto.ListClusterNodeGroupRequest
	resp      *cmproto.ListClusterNodeGroupResponse
	groupList []*cmproto.NodeGroup
}

// NewListClusterGroupAction create list action for cluster nodeGroup
func NewListClusterGroupAction(model store.ClusterManagerModel) *ListClusterGroupAction {
	return &ListClusterGroupAction{
		model: model,
	}
}

func (la *ListClusterGroupAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.groupList
}

// Handle list cluster credential
func (la *ListClusterGroupAction) Handle(
	ctx context.Context, req *cmproto.ListClusterNodeGroupRequest, resp *cmproto.ListClusterNodeGroupResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list cluster NodeGroup failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	var err error
	if err = req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	var (
		groupList              []*cmproto.NodeGroup
		enableAutoscalerGroup  []*cmproto.NodeGroup
		disableAutoscalerGroup []*cmproto.NodeGroup
		enableCA               bool
	)

	asOption, _ := la.model.GetAutoScalingOption(la.ctx, la.req.ClusterID)
	if asOption != nil && asOption.EnableAutoscale {
		enableCA = true
	}

	groupList, err = listNodeGroupByConds(la.model, filterNodeGroupOption{
		ClusterID: la.req.GetClusterID(),
	})
	if err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if !la.req.GetEnableFilter() {
		sort.Sort(utils.NodeGroupSlice(groupList))
		la.groupList = groupList

		la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
		return
	}

	for i := range groupList {
		if enableCA && groupList[i].GetEnableAutoscale() {
			enableAutoscalerGroup = append(enableAutoscalerGroup, groupList[i])
			continue
		}

		groupList[i].EnableAutoscale = false
		disableAutoscalerGroup = append(disableAutoscalerGroup, groupList[i])
	}
	sort.Sort(utils.NodeGroupSlice(enableAutoscalerGroup))
	sort.Sort(utils.NodeGroupSlice(disableAutoscalerGroup))

	la.groupList = append(la.groupList, disableAutoscalerGroup...)
	la.groupList = append(la.groupList, enableAutoscalerGroup...)

	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

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

	var err error
	if err = req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	la.groupList, err = listNodeGroupByConds(la.model, filterNodeGroupOption{
		Name:      la.req.GetName(),
		ClusterID: la.req.GetClusterID(),
		ProjectID: la.req.GetProjectID(),
		Region:    la.req.GetRegion(),
	})
	if err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// ListNodesAction action for list online cluster credential
type ListNodesAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.GetNodeGroupRequest
	resp  *cmproto.ListNodesInGroupResponse
}

// NewListNodesAction create list action for cluster credential
func NewListNodesAction(model store.ClusterManagerModel) *ListNodesAction {
	return &ListNodesAction{
		model: model,
	}
}

func (la *ListNodesAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (la *ListNodesAction) listNodesInGroup() error {
	group, err := la.model.GetNodeGroup(la.ctx, la.req.NodeGroupID)
	if err != nil {
		blog.Errorf("get NodeGroup %s failed when list its all nodes, %s", la.req.NodeGroupID, err.Error())
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}

	// check if nodes are already in cluster
	condM := make(operator.M)
	condM["nodegroupid"] = group.NodeGroupID
	cond := operator.NewLeafCondition(operator.Eq, condM)
	nodes, err := la.model.ListNode(la.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		blog.Errorf("get NodeGroup %s all Nodes failed, %s", group.NodeGroupID, err.Error())
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
	}

	for i := range nodes {
		nodes[i].Passwd = ""
		la.resp.Data = append(la.resp.Data, nodes[i])
	}

	return nil
}

// Handle handle list cluster credential
func (la *ListNodesAction) Handle(
	ctx context.Context, req *cmproto.GetNodeGroupRequest, resp *cmproto.ListNodesInGroupResponse) {
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

	if err := la.listNodesInGroup(); err != nil {
		la.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// ListNodesV2Action action for list online cluster nodes
type ListNodesV2Action struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	group   *cmproto.NodeGroup
	cluster *cmproto.Cluster
	cloud   *cmproto.Cloud
	k8sOp   *clusterops.K8SOperator
	req     *cmproto.ListNodesInGroupV2Request
	resp    *cmproto.ListNodesInGroupV2Response
}

// NewListNodesV2Action create list action for cluster credential
func NewListNodesV2Action(model store.ClusterManagerModel, k8sOp *clusterops.K8SOperator) *ListNodesV2Action {
	return &ListNodesV2Action{
		model: model,
		k8sOp: k8sOp,
	}
}

func (la *ListNodesV2Action) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (la *ListNodesV2Action) getRelativeResource() error {
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

func (la *ListNodesV2Action) listNodesInGroup() error {
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

func (la *ListNodesV2Action) appendNodeInfo() {
	connect := autils.CheckClusterConnection(la.k8sOp, la.cluster.ClusterID)
	if !connect {
		return
	}

	k8sNodes, err := la.k8sOp.ListClusterNodes(la.ctx, la.group.ClusterID)
	if err != nil {
		blog.Warnf("ListClusterNodes %s failed, %s", la.group.ClusterID, err.Error())
	}
	nodeMap := make(map[string]*v1.Node)
	for _, v := range k8sNodes {
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
			}
			la.resp.Data[i].Status = actions.TransNodeStatus(la.resp.Data[i].Status, node)
		}
	}
}

// Handle list cluster credential
func (la *ListNodesV2Action) Handle(
	ctx context.Context, req *cmproto.ListNodesInGroupV2Request, resp *cmproto.ListNodesInGroupV2Response) {
	if req == nil || resp == nil {
		blog.Errorf("ListNodesV2Action failed, req or resp is empty")
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
}
