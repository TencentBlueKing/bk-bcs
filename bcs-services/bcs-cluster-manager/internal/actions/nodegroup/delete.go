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
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/taskserver"
)

// DeleteAction action for delete cluster credential
type DeleteAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.DeleteNodeGroupRequest
	resp  *cmproto.DeleteNodeGroupResponse
	group *cmproto.NodeGroup
	nodes []*cmproto.Node
}

// NewDeleteAction create delete action for nodeGroup
func NewDeleteAction(model store.ClusterManagerModel) *DeleteAction {
	return &DeleteAction{
		model: model,
	}
}

func (da *DeleteAction) setResp(code uint32, msg string) {
	da.resp.Code = code
	da.resp.Message = msg
	da.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (da *DeleteAction) validate() error {
	if err := da.req.Validate(); err != nil {
		blog.Errorf("delete nodegroup request is invalidate, %s", err.Error())
		da.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return err
	}
	// validate nodegroup existence
	group, err := da.model.GetNodeGroup(da.ctx, da.req.NodeGroupID)
	if err != nil {
		blog.Errorf("get NodeGroup %s failed, %s", da.req.NodeGroupID, err.Error())
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Warnf("No NodeGroup %s existence, skip delete", da.req.NodeGroupID)
			da.setResp(common.BcsErrClusterManagerSuccess, "no data need to delete")
			return err
		}
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	da.group = group
	// get all node under nodegroup
	condM := make(operator.M)
	condM["nodegroupid"] = group.NodeGroupID
	cond := operator.NewLeafCondition(operator.Eq, condM)
	nodes, err := da.model.ListNode(da.ctx, cond, &options.ListOption{})
	if err != nil {
		blog.Errorf("get NodeGroup %s all Nodes failed, %s", group.NodeGroupID, err.Error())
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	// check
	if len(nodes) != 0 && !da.req.IsForce {
		blog.Warnf(
			"NodeGroup %s for Cluster %s still has Nodes, no deletion execute. try isForce",
			group.NodeGroupID, group.ClusterID,
		)
		da.setResp(common.BcsErrClusterManagerInvalidParameter, "nodegroup still has nodes")
		return fmt.Errorf("nodegroup still has nodes")
	}
	for i := range nodes {
		da.nodes = append(da.nodes, nodes[i])
	}
	return nil
}

// Handle handle delete cluster credential
func (da *DeleteAction) Handle(
	ctx context.Context, req *cmproto.DeleteNodeGroupRequest, resp *cmproto.DeleteNodeGroupResponse) {
	if req == nil || resp == nil {
		blog.Errorf("delete cloud failed, req or resp is empty")
		return
	}
	da.ctx = ctx
	da.req = req
	da.resp = resp

	// get nodeGroup and all nodes under group
	// IsForce = false, not allow delete directly
	if err := da.validate(); err != nil {
		// validate already setting response
		return
	}

	// nodeGroup cloud provider
	cloud, cluster, err := actions.GetCloudAndCluster(da.model, da.group.Provider, da.group.ClusterID)
	if err != nil {
		blog.Errorf("get Cloud %s for NodeGroup %s in Cluster %s deletion failed, %s",
			da.group.Provider, da.group.NodeGroupID, da.group.ClusterID, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	//get dependency resource for cloudprovider operation
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential for NodeGroup %s in Cluster %s deletion failed, %s",
			da.group.NodeGroupID, da.group.ClusterID, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	// get cloudprovider and then start to delete
	mgr, err := cloudprovider.GetNodeGroupMgr(cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s/%s for NodeGroup %s in Cluster %s deletion failed, %s",
			cloud.CloudID, cloud.CloudProvider, da.group.NodeGroupID, da.group.ClusterID, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	cmOption.Region = da.group.Region

	// delete nodeGroup task
	task, err := mgr.DeleteNodeGroup(da.group, da.nodes, &cloudprovider.DeleteNodeGroupOption{
		CommonOption:          *cmOption,
		IsForce:               req.IsForce,
		ReserveNodesInCluster: req.ReserveNodesInCluster,
		ReservedNodeInstance:  req.KeepNodesInstance,
		Operator:              req.Operator,
		Cloud:                 cloud,
	})
	if err != nil {
		blog.Errorf("delete NodeGroup %s in Cluster %s with cloudprovider %s failed, %s",
			da.group.NodeGroupID, da.group.ClusterID, cloud.CloudProvider, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	// distribute task
	if task != nil {
		if err = da.model.CreateTask(da.ctx, task); err != nil {
			blog.Errorf("save delete nodeGroup task for NodeGroup %s failed, %s",
				da.group.NodeGroupID, err.Error(),
			)
			da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
			return
		}
		if err = taskserver.GetTaskServer().Dispatch(task); err != nil {
			blog.Errorf("dispatch delete nodeGroup task for NodeGroup %s failed, %s",
				da.group.NodeGroupID, err.Error(),
			)
			da.setResp(common.BcsErrClusterManagerTaskErr, err.Error())
			return
		}
		blog.Infof("delete nodeGroup task successfully for %s", da.group.NodeGroupID)
	}

	if !req.IsForce && len(da.nodes) == 0 {
		// here means no Nodes in NodeGroup, just delete local information
		if err = da.model.DeleteNodeGroup(da.ctx, da.group.NodeGroupID); err != nil {
			blog.Errorf("delete NodeGroup %s local information in Cluster %s failed, %s",
				da.group.NodeGroupID, da.group.ClusterID, err.Error(),
			)
			//! no return here, task already started, try another deletion in final task step
		}
	}
	resp.Data = da.group
	resp.Task = task

	err = da.model.CreateOperationLog(da.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   da.group.NodeGroupID,
		TaskID: func() string {
			if task == nil {
				return ""
			}
			return task.TaskID
		}(),
		Message:    fmt.Sprintf("集群%s删除节点池%s", da.group.ClusterID, da.group.NodeGroupID),
		OpUser:     req.Operator,
		CreateTime: time.Now().String(),
	})
	if err != nil {
		blog.Errorf("DeleteNodeGroup[%s] CreateOperationLog failed: %v", da.group.NodeGroupID, err)
	}
	da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

// RemoveNodeAction action just move node out of NodeGroup management
type RemoveNodeAction struct {
	ctx context.Context

	model       store.ClusterManagerModel
	req         *cmproto.RemoveNodesFromGroupRequest
	resp        *cmproto.RemoveNodesFromGroupResponse
	group       *cmproto.NodeGroup
	cluster     *cmproto.Cluster
	removeNodes []*cmproto.Node
}

// NewRemoveNodeAction create delete action for removeNodes
func NewRemoveNodeAction(model store.ClusterManagerModel) *RemoveNodeAction {
	return &RemoveNodeAction{
		model: model,
	}
}

func (da *RemoveNodeAction) setResp(code uint32, msg string) {
	da.resp.Code = code
	da.resp.Message = msg
	da.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (da *RemoveNodeAction) validate() error {
	if err := da.req.Validate(); err != nil {
		da.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return err
	}
	//get nodegroup for validation
	destGroup, err := da.model.GetNodeGroup(da.ctx, da.req.NodeGroupID)
	if err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("Get NodeGroup %s in pre-RemoveNode checking failed, err %s", da.req.NodeGroupID, err.Error())
		return err
	}
	//check cluster info consistency
	if destGroup.ClusterID != da.req.ClusterID {
		blog.Errorf(
			"request ClusterID %s is not same with NodeGroup.ClusterID %s when RemoveNode",
			da.req.ClusterID, destGroup.ClusterID,
		)
		da.setResp(
			common.BcsErrClusterManagerInvalidParameter,
			fmt.Sprintf("request ClusterID is not same with NodeGroup.ClusterID %s", destGroup.ClusterID),
		)
		return err
	}
	da.group = destGroup
	//check cluster existence
	cluster, err := da.model.GetCluster(da.ctx, destGroup.ClusterID)
	if err != nil {
		blog.Errorf("get Cluster %s for NodeGroup %s to remove Node failed, %s",
			destGroup.ClusterID, destGroup.NodeGroupID, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	da.cluster = cluster
	//get specified node for remove validation
	condM := make(operator.M)
	condM["nodegroupid"] = da.group.NodeGroupID
	condM["clusterid"] = da.group.ClusterID
	cond := operator.NewLeafCondition(operator.Eq, condM)
	nodes, err := da.model.ListNode(da.ctx, cond, &options.ListOption{})
	if err != nil {
		blog.Errorf("get NodeGroup %s all Nodes failed, %s", da.group.NodeGroupID, err.Error())
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	allNodes := make(map[string]*cmproto.Node)
	for i := range nodes {
		allNodes[nodes[i].InnerIP] = nodes[i]
	}
	//Nodes validation for remove
	for _, ip := range da.req.Nodes {
		node, ok := allNodes[ip]
		if !ok {
			blog.Errorf("remove Node %s is Not under NodeGroup %s control", ip, da.group.NodeGroupID)
			err := fmt.Errorf("node %s is not belong to NodeGroup %s", ip, destGroup.NodeGroupID)
			da.setResp(
				common.BcsErrClusterManagerInvalidParameter,
				err.Error(),
			)
			return err
		}
		da.removeNodes = append(da.removeNodes, node)
	}
	return nil
}

// Handle handle delete nodes
func (da *RemoveNodeAction) Handle(
	ctx context.Context, req *cmproto.RemoveNodesFromGroupRequest, resp *cmproto.RemoveNodesFromGroupResponse) {
	if req == nil || resp == nil {
		blog.Errorf("RemoveNodeFromGroup failed, req or resp is empty")
		return
	}
	da.ctx = ctx
	da.req = req
	da.resp = resp

	if err := da.validate(); err != nil {
		// validate already setting error message
		return
	}
	//get dependency resource and release it
	cloud, cluster, err := actions.GetCloudAndCluster(da.model, da.group.Provider, da.group.ClusterID)
	if err != nil {
		blog.Errorf("get Cloud %s for NodeGroup %s to remove Node failed, %s",
			da.group.Provider, da.group.NodeGroupID, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	//get dependency resource for cloudprovider operation
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential for NodeGroup %s cluster %s to remove Node failed, %s",
			da.group.NodeGroupID, da.group.ClusterID, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	//get cloudprovider and then start to remove
	mgr, err := cloudprovider.GetNodeGroupMgr(cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s/%s for NodeGroup %s to remove Node failed, %s",
			cloud.CloudID, cloud.CloudProvider, da.group.NodeGroupID, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	cmOption.Region = da.group.Region
	if err = mgr.RemoveNodesFromGroup(da.removeNodes, da.group, &cloudprovider.RemoveNodesOption{
		CommonOption: *cmOption,
		Cloud:        cloud,
		Cluster:      da.cluster,
	}); err != nil {
		blog.Errorf("remove Node %v from NodeGroup %s with cloudprovider %s failed, %s",
			req.Nodes, da.group.NodeGroupID, cloud.CloudProvider, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	blog.Infof(
		"Nodes %v remove out NodeGroup %s in cloudprovider %s/%s successfully",
		req.Nodes, da.group.NodeGroupID, cloud.CloudID, cloud.CloudProvider,
	)
	//try to update Node
	for _, node := range da.removeNodes {
		node.NodeGroupID = ""
		if err = da.model.UpdateNode(ctx, node); err != nil {
			blog.Errorf("update NodeGroup %s with Nodes %s remove out failed, %s",
				da.group.NodeGroupID, node.InnerIP, err.Error(),
			)
			da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
			return
		}
		blog.Infof("Nodes %s remove out NodeGroup %s record successfully", node.InnerIP, da.group.NodeGroupID)
	}

	err = da.model.CreateOperationLog(da.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   da.group.NodeGroupID,
		TaskID:       "",
		Message:      fmt.Sprintf("集群%s节点池%s移除节点", da.group.ClusterID, da.group.NodeGroupID),
		OpUser:       da.group.Creator,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("RemoveNodesFromNodeGroup[%s] CreateOperationLog failed: %v", da.group.NodeGroupID, err)
	}

	da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

// CleanNodesAction action for delete nodes
type CleanNodesAction struct {
	ctx context.Context

	model      store.ClusterManagerModel
	req        *cmproto.CleanNodesInGroupRequest
	resp       *cmproto.CleanNodesInGroupResponse
	group      *cmproto.NodeGroup
	cleanNodes []*cmproto.Node
}

// NewCleanNodesAction create delete action for online cluster credential
func NewCleanNodesAction(model store.ClusterManagerModel) *CleanNodesAction {
	return &CleanNodesAction{
		model: model,
	}
}

func (da *CleanNodesAction) setResp(code uint32, msg string) {
	da.resp.Code = code
	da.resp.Message = msg
	da.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (da *CleanNodesAction) validate() error {
	if err := da.req.Validate(); err != nil {
		da.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return err
	}
	//try to get original data for return
	group, err := da.model.GetNodeGroup(da.ctx, da.req.NodeGroupID)
	if err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("Get NodeGroup %s in pre-CleanNode checking failed, err %s", da.req.NodeGroupID, err.Error())
		return err
	}
	if group.ClusterID != da.req.ClusterID {
		blog.Errorf("request ClusterID %s is not same with NodeGroup.ClusterID %s when CleanNode",
			da.req.ClusterID, group.ClusterID,
		)
		err := fmt.Errorf("request ClusterID is not same with NodeGroup.ClusterID %s", group.ClusterID)
		da.setResp(
			common.BcsErrClusterManagerCommonErr,
			err.Error(),
		)
		return err
	}
	da.group = group
	//get specified node for clean validation
	condM := make(operator.M)
	condM["nodegroupid"] = da.group.NodeGroupID
	condM["clusterid"] = da.group.ClusterID
	cond := operator.NewLeafCondition(operator.Eq, condM)
	nodes, err := da.model.ListNode(da.ctx, cond, &options.ListOption{})
	if err != nil {
		blog.Errorf("get NodeGroup %s all Nodes failed, %s", da.group.NodeGroupID, err.Error())
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	allNodes := make(map[string]*cmproto.Node)
	for i := range nodes {
		allNodes[nodes[i].InnerIP] = nodes[i]
	}
	// Nodes validation for clean
	for _, ip := range da.req.Nodes {
		node, ok := allNodes[ip]
		if !ok {
			blog.Errorf("clean Node %s is Not under NodeGroup %s control", ip, da.group.NodeGroupID)
			continue
		}
		da.cleanNodes = append(da.cleanNodes, node)
	}
	return nil
}

// Handle handle delete cluster credential
func (da *CleanNodesAction) Handle(
	ctx context.Context, req *cmproto.CleanNodesInGroupRequest, resp *cmproto.CleanNodesInGroupResponse) {
	if req == nil || resp == nil {
		blog.Errorf("CleanNodesAction failed, req or resp is empty")
		return
	}
	da.ctx = ctx
	da.req = req
	da.resp = resp

	if err := da.validate(); err != nil {
		//valiate already sets response information
		return
	}
	//get dependency resource
	cloud, cluster, err := actions.GetCloudAndCluster(da.model, da.group.Provider, da.group.ClusterID)
	if err != nil {
		blog.Errorf("get Cloud %s for NodeGroup %s to clean Node failed, %s",
			da.group.Provider, da.group.NodeGroupID, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	//ready to create background task
	taskMgr, err := cloudprovider.GetTaskManager(cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s/%s for NodeGroup %s to clean Nodes %v failed, %s",
			cloud.CloudID, cloud.CloudProvider, da.group.NodeGroupID, req.Nodes, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	// try to update Cluster & NodeGroup
	for _, node := range da.cleanNodes {
		node.Status = common.StatusDeleting
		// how to ensure consistency with other operation?
		if err = da.model.UpdateNode(ctx, node); err != nil {
			blog.Errorf("update NodeGroup %s with Nodes %v status change to DELETING failed, %s",
				da.group.ClusterID, req.Nodes, err.Error(),
			)
			da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
			return
		}
		blog.Infof("Nodes %v of NodeGroup %s change to DELETING successfully",
			node.InnerIP, da.group.NodeGroupID,
		)
	}
	// build clean task and dispatch to run
	task, err := taskMgr.BuildCleanNodesInGroupTask(da.cleanNodes, da.group, &cloudprovider.TaskOptions{
		Cloud:    cloud,
		Cluster:  cluster,
		Operator: req.Operator,
	})
	if err != nil {
		blog.Errorf("build clean Node %v task from NodeGroup %s with cloudprovider %s failed, %s",
			req.Nodes, da.group.NodeGroupID, cloud.CloudProvider, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	if err = da.model.CreateTask(ctx, task); err != nil {
		blog.Errorf("save clean Node %v task from NodeGroup %s failed, %s",
			req.Nodes, da.group.NodeGroupID, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	if err = taskserver.GetTaskServer().Dispatch(task); err != nil {
		blog.Errorf("dispatch clean Node %v task from NodeGroup %s failed, %s",
			req.Nodes, da.group.NodeGroupID, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	blog.Infof("NodeGroup %s clean nodes %v task created with cloudprovider %s/%s successfully",
		da.group.NodeGroupID, req.Nodes, cloud.CloudID, cloud.CloudProvider,
	)

	err = da.model.CreateOperationLog(da.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   da.group.NodeGroupID,
		TaskID:       task.TaskID,
		Message:      fmt.Sprintf("集群%s节点池%s删除节点", da.group.ClusterID, da.group.NodeGroupID),
		OpUser:       req.Operator,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("CleanNodesFromNodeGroup[%s] CreateOperationLog failed: %v", da.group.NodeGroupID, err)
	}

	resp.Data = task
	da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
