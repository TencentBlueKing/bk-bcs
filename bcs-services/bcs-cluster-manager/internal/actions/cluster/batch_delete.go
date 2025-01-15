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

// Package cluster xxx
package cluster

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/nodegroup"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/lock"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// BatchDeleteClusterNodesAction action for delete nodes from cluster
type BatchDeleteClusterNodesAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.BatchDeleteClusterNodesRequest
	resp  *cmproto.BatchDeleteClusterNodesResponse

	normalNodes     []nodeData
	externalNodes   map[string][]nodeData
	nodeGroupNodes  map[string][]nodeData
	virtualNodesMap map[string][]nodeData

	cluster      *cmproto.Cluster
	nodes        []*cmproto.Node
	virtualNodes []*cmproto.Node

	locker lock.DistributedLock
}

type nodeData struct {
	clusterID   string
	innerIp     string
	nodeId      string
	nodeGroupID string
	isExternal  bool
}

// NewBatchDeleteClusterNodesAction delete cluster nodes action
func NewBatchDeleteClusterNodesAction(model store.ClusterManagerModel,
	lock lock.DistributedLock) *BatchDeleteClusterNodesAction {
	return &BatchDeleteClusterNodesAction{
		model:  model,
		locker: lock,
	}
}

func (ba *BatchDeleteClusterNodesAction) setResp(code uint32, msg string) {
	ba.resp.Code = code
	ba.resp.Message = msg
	ba.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ba *BatchDeleteClusterNodesAction) validate() error {
	err := ba.req.Validate()
	if err != nil {
		return err
	}

	// chekc node ip list
	if ba.req.NodeIPs == "" && ba.req.VirtualNodeIDs == "" {
		return fmt.Errorf("BatchDeleteClusterNodesAction paras empty")
	}

	// check node ip list
	if len(strings.Split(ba.req.NodeIPs, ",")) == 0 && len(strings.Split(ba.req.VirtualNodeIDs, ",")) == 0 {
		return fmt.Errorf("BatchDeleteClusterNodesAction paras empty")
	}

	if ba.resp.Data == nil {
		ba.resp.Data = make([]*cmproto.BatchNodesStatus, 0)
	}

	return nil
}

func (ba *BatchDeleteClusterNodesAction) getClusterAndNodes() error {
	var err error
	ba.cluster, err = actions.GetClusterInfoByClusterID(ba.model, ba.req.GetClusterID())
	if err != nil {
		return err
	}
	nodeIpList := strings.Split(ba.req.GetNodeIPs(), ",")
	if len(nodeIpList) == 0 {
		return nil
	}

	// get relative nodes by clusterID
	clusterCond := operator.NewLeafCondition(operator.Eq, operator.M{"clusterid": ba.req.GetClusterID()})
	nodeCond := operator.NewLeafCondition(operator.In, operator.M{"innerip": nodeIpList})
	cond := operator.NewBranchCondition(operator.And, clusterCond, nodeCond)
	nodes, err := ba.model.ListNode(ba.ctx, cond, &storeopt.ListOption{})
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		blog.Errorf("get Cluster %s Nodes failed, %s", ba.req.ClusterID, err.Error())
		return err
	}

	var (
		inDbNodes    = make([]string, 0)
		notInDbNodes = make([]string, 0)
	)

	// check node if deleting or initialization
	for i := range nodes {
		inDbNodes = append(inDbNodes, nodes[i].GetInnerIP())
		// filter deleting/initialization status
		if nodes[i].Status == common.StatusDeleting || nodes[i].Status == common.StatusInitialization {
			continue
		}

		ba.nodes = append(ba.nodes, nodes[i])
	}

	// handle notInDbNodes
	for i := range nodeIpList {
		if !utils.StringInSlice(nodeIpList[i], inDbNodes) {
			notInDbNodes = append(notInDbNodes, nodeIpList[i])
		}
	}

	blog.Infof("BatchDeleteClusterNodesAction[%s] getClusterAndNodes inDbNodes[%+v] notInDbNodes[%+v]",
		ba.req.ClusterID, inDbNodes, notInDbNodes)
	// handle notInDbNodes
	for i := range notInDbNodes {
		ba.nodes = append(ba.nodes, &cmproto.Node{
			InnerIP:   notInDbNodes[i],
			ClusterID: ba.req.ClusterID,
		})
	}

	return nil
}

// getClusterVirtualNodes get cluster virtual nodes
func (ba *BatchDeleteClusterNodesAction) getClusterVirtualNodes() error {
	if ba.req.VirtualNodeIDs == "" {
		return nil
	}

	nodeIDList := strings.Split(ba.req.VirtualNodeIDs, ",")

	// get relative nodes by clusterID
	clusterCond := operator.NewLeafCondition(operator.Eq, operator.M{"clusterid": ba.req.GetClusterID()})
	nodeCond := operator.NewLeafCondition(operator.In, operator.M{"nodeid": nodeIDList})
	cond := operator.NewBranchCondition(operator.And, clusterCond, nodeCond)
	nodes, err := ba.model.ListNode(ba.ctx, cond, &storeopt.ListOption{})
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		blog.Errorf("get Cluster %s Nodes failed, %s", ba.req.ClusterID, err.Error())
		return err
	}

	var (
		virtualInDbNodeIDs = make([]string, 0)
	)

	// check node if deleting or initialization or applying
	for i := range nodes {
		virtualInDbNodeIDs = append(virtualInDbNodeIDs, nodes[i].NodeID)
		if nodes[i].Status == common.StatusDeleting ||
			nodes[i].Status == common.StatusResourceApplying || nodes[i].Status == common.StatusInitialization {
			continue
		}

		ba.virtualNodes = append(ba.virtualNodes, nodes[i])
	}
	blog.Infof("BatchDeleteClusterNodesAction[%s] getClusterVirtualNodes inDbNodeIDs[%+v]",
		ba.req.ClusterID, virtualInDbNodeIDs)

	return nil
}

// sortNodesAndHandleTask sort nodes and handle task
func (ba *BatchDeleteClusterNodesAction) sortNeedToDeleteNodes() error { // nolint
	err := ba.getClusterAndNodes()
	if err != nil {
		return err
	}
	err = ba.getClusterVirtualNodes()
	if err != nil {
		return err
	}

	// init map
	var (
		normalNodes    = make([]nodeData, 0)
		externalNodes  = make(map[string][]nodeData, 0)
		nodeGroupNodes = make(map[string][]nodeData, 0)
		virtualNodes   = make(map[string][]nodeData, 0)
	)
	for i := range ba.nodes {
		if ba.nodes[i].GetNodeGroupID() == "" {
			normalNodes = append(normalNodes, nodeData{
				clusterID: ba.nodes[i].ClusterID,
				innerIp:   ba.nodes[i].InnerIP,
			})
			continue
		}

		// get node group
		group, nodeErr := ba.model.GetNodeGroup(context.Background(), ba.nodes[i].GetNodeGroupID())
		if nodeErr != nil {
			blog.Errorf("BatchDeleteClusterNodesAction sortNodesAndHandleTask failed: %v", nodeErr)
			continue
		}

		// 手动上下架第三方节点
		if group.GetNodeGroupType() == common.External.String() && group.GetAutoScaling().GetMaxSize() == 0 {
			nodes := externalNodes[group.GetNodeGroupID()]
			if nodes == nil {
				externalNodes[group.GetNodeGroupID()] = make([]nodeData, 0)
			}
			externalNodes[group.GetNodeGroupID()] = append(externalNodes[group.GetNodeGroupID()], nodeData{
				clusterID:   ba.nodes[i].ClusterID,
				innerIp:     ba.nodes[i].InnerIP,
				nodeGroupID: ba.nodes[i].GetNodeGroupID(),
				isExternal:  true,
			})
			continue
		}

		// init nodeGroupNodes
		if nodeGroupNodes[group.GetNodeGroupID()] == nil {
			nodeGroupNodes[group.GetNodeGroupID()] = make([]nodeData, 0)
		}

		nodeGroupNodes[group.GetNodeGroupID()] = append(nodeGroupNodes[group.GetNodeGroupID()], nodeData{
			clusterID:   ba.nodes[i].ClusterID,
			innerIp:     ba.nodes[i].InnerIP,
			nodeGroupID: ba.nodes[i].GetNodeGroupID(),
			isExternal:  false,
		})
	}

	// handle virtual nodes
	for i := range ba.virtualNodes {
		blog.Infof("BatchDeleteClusterNodesAction virtualNodes[%s:%s:%s]", ba.virtualNodes[i].NodeID,
			ba.virtualNodes[i].ClusterID, ba.virtualNodes[i].NodeGroupID)

		if ba.virtualNodes[i].NodeGroupID == "" {
			_ = ba.model.DeleteClusterNode(ba.ctx, ba.req.ClusterID, ba.virtualNodes[i].NodeID)
			blog.Infof("BatchDeleteClusterNodesAction virtualNodes[%s:%s:%s] nodeGroup empty",
				ba.virtualNodes[i].NodeID, ba.virtualNodes[i].ClusterID, ba.virtualNodes[i].NodeGroupID)
			continue
		}

		// check virtual node status
		if ba.virtualNodes[i].Status == common.StatusInitialization ||
			ba.virtualNodes[i].Status == common.StatusResourceApplying {
			blog.Infof("BatchDeleteClusterNodesAction virtualNodes[%s:%s:%s] status[%s], not allow delete",
				ba.virtualNodes[i].NodeID, ba.virtualNodes[i].ClusterID, ba.virtualNodes[i].NodeGroupID,
				ba.virtualNodes[i].Status)
			continue
		}

		if virtualNodes[ba.virtualNodes[i].GetNodeGroupID()] == nil {
			virtualNodes[ba.virtualNodes[i].GetNodeGroupID()] = make([]nodeData, 0)
		}
		virtualNodes[ba.virtualNodes[i].GetNodeGroupID()] = append(virtualNodes[ba.virtualNodes[i].GetNodeGroupID()],
			nodeData{
				clusterID:   ba.virtualNodes[i].ClusterID,
				innerIp:     ba.virtualNodes[i].InnerIP,
				nodeId:      ba.virtualNodes[i].NodeID,
				nodeGroupID: ba.virtualNodes[i].GetNodeGroupID(),
				isExternal:  false,
			})
	}

	blog.Infof("BatchDeleteClusterNodesAction sortNodesAndHandleTask normal[%+v] external[%+v] "+
		"nodeGroupNode[%+v], virtualNode[%+v]", normalNodes, externalNodes, nodeGroupNodes, virtualNodes)

	ba.normalNodes = normalNodes
	ba.externalNodes = externalNodes
	ba.nodeGroupNodes = nodeGroupNodes
	ba.virtualNodesMap = virtualNodes

	return nil
}

// getNodeDataIPList get node data ip list
func getNodeDataIPList(nodes []nodeData) []string {
	ipList := make([]string, 0)
	for i := range nodes {
		ipList = append(ipList, nodes[i].innerIp)
	}

	return ipList
}

// handleGroupNodes handle group nodes
func (ba *BatchDeleteClusterNodesAction) handleGroupNodes(ipList []string, groupID string) {
	var (
		cleanNodesRequest = &cmproto.CleanNodesInGroupRequest{
			ClusterID:   ba.req.ClusterID,
			Nodes:       ipList,
			NodeGroupID: groupID,
			Operator:    ba.req.Operator,
			Manual:      true,
		}
		cleanNodesResp = &cmproto.CleanNodesInGroupResponse{}
	)
	nodegroup.NewCleanNodesAction(ba.model, ba.locker).Handle(ba.ctx, cleanNodesRequest, cleanNodesResp)

	if cleanNodesResp.Code != 0 || !cleanNodesResp.Result {
		ba.resp.Data = append(ba.resp.Data, &cmproto.BatchNodesStatus{
			NodeIPs:     ipList,
			Success:     false,
			NodeGroupID: groupID,
			Message:     ba.resp.GetMessage(),
			TaskID:      "",
		})

		return
	}

	ba.resp.Data = append(ba.resp.Data, &cmproto.BatchNodesStatus{
		NodeIPs:     ipList,
		Success:     true,
		NodeGroupID: groupID,
		Message:     cleanNodesResp.GetMessage(),
		TaskID:      cleanNodesResp.Data.GetTaskID(),
	})
}

// handleManualNodes handle manual nodes
func (ba *BatchDeleteClusterNodesAction) handleManualNodes(ipList []string, groupID string, external bool) {
	var (
		deleteNodesRequest = &cmproto.DeleteNodesRequest{
			ClusterID:      ba.req.ClusterID,
			Nodes:          strings.Join(ipList, ","),
			Operator:       ba.req.Operator,
			NodeGroupID:    groupID,
			IsExternalNode: external,
			DeleteMode:     ba.req.DeleteMode,
		}

		deleteNodesResp = &cmproto.DeleteNodesResponse{}
	)

	// delete nodes
	NewDeleteNodesAction(ba.model).Handle(ba.ctx, deleteNodesRequest, deleteNodesResp)
	if deleteNodesResp.Code != 0 || !deleteNodesResp.Result {
		ba.resp.Data = append(ba.resp.Data, &cmproto.BatchNodesStatus{
			NodeIPs:     ipList,
			Success:     false,
			NodeGroupID: groupID,
			Message:     ba.resp.GetMessage(),
			TaskID:      "",
		})

		return
	}

	ba.resp.Data = append(ba.resp.Data, &cmproto.BatchNodesStatus{
		NodeIPs:     ipList,
		Success:     true,
		NodeGroupID: groupID,
		Message:     deleteNodesResp.GetMessage(),
		TaskID:      deleteNodesResp.Data.GetTaskID(),
	})
}

// handleVirtualNodes handle virtual nodes
func (ba *BatchDeleteClusterNodesAction) handleVirtualNodes() {
	if len(ba.virtualNodesMap) == 0 {
		return
	}

	for groupID, groupNodes := range ba.virtualNodesMap {
		if len(groupNodes) == 0 {
			continue
		}

		blog.Infof("handleVirtualNodes group[%s] groupNodes[%+v]", groupID, groupNodes)
		err := actions.UpdateNodeGroupDesiredSize(ba.model, groupID, len(groupNodes), true)
		if err != nil {
			blog.Errorf("handleVirtualNodes[%s] failed: %v", groupID, err)
		}

		for i := range groupNodes {
			err = ba.model.DeleteClusterNode(ba.ctx, ba.req.ClusterID, groupNodes[i].nodeId)
			if err != nil {
				blog.Errorf("handleVirtualNodes[%s] DeleteClusterNode[%s] failed: %v", groupID, groupNodes[i].nodeId,
					err)
			}
		}
	}
}

// handleNormalNodes handle normal nodes
func (ba *BatchDeleteClusterNodesAction) handleNormalNodes() {
	if len(ba.normalNodes) == 0 {
		return
	}

	ipList := getNodeDataIPList(ba.normalNodes)
	if len(ipList) == 0 {
		return
	}

	ba.handleManualNodes(ipList, "", false)
}

// handleExternalNodes handle external nodes
func (ba *BatchDeleteClusterNodesAction) handleExternalNodes() {
	if len(ba.externalNodes) == 0 {
		return
	}

	for groupID, externalNodes := range ba.externalNodes {
		if len(externalNodes) == 0 {
			continue
		}
		ba.handleManualNodes(getNodeDataIPList(externalNodes), groupID, true)
	}
}

// handleNodeGroupNodes handle node group nodes
func (ba *BatchDeleteClusterNodesAction) handleNodeGroupNodes() {
	if len(ba.nodeGroupNodes) == 0 {
		return
	}

	for groupID, groupNodes := range ba.nodeGroupNodes {
		if len(groupNodes) == 0 {
			continue
		}
		ba.handleGroupNodes(getNodeDataIPList(groupNodes), groupID)
	}
}

// Handle delete cluster nodes request
func (ba *BatchDeleteClusterNodesAction) Handle(ctx context.Context, req *cmproto.BatchDeleteClusterNodesRequest,
	resp *cmproto.BatchDeleteClusterNodesResponse) {
	if req == nil || resp == nil {
		blog.Errorf("batch delete clusterNodes failed, req or resp is empty")
		return
	}
	ba.ctx = ctx
	ba.req = req
	ba.resp = resp

	// check request parameter validate
	err := ba.validate()
	if err != nil {
		ba.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	err = ba.sortNeedToDeleteNodes()
	if err != nil {
		ba.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ba.handleVirtualNodes()
	ba.handleNormalNodes()
	ba.handleExternalNodes()
	ba.handleNodeGroupNodes()

	ba.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
