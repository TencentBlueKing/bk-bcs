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
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/avast/retry-go"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/lock"
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
	// check nodes exist
	if len(nodes) != 0 && !da.req.IsForce {
		blog.Warnf(
			"NodeGroup %s for Cluster %s still has Nodes, no deletion execute. try isForce",
			group.NodeGroupID, group.ClusterID,
		)
		da.setResp(common.BcsErrClusterManagerInvalidParameter, "nodegroup still has nodes")
		return fmt.Errorf("nodegroup still has nodes")
	}
	da.nodes = append(da.nodes, nodes...)
	return nil
}

// Handle handle delete cluster nodeGroup
func (da *DeleteAction) Handle( // nolint
	ctx context.Context, req *cmproto.DeleteNodeGroupRequest, resp *cmproto.DeleteNodeGroupResponse) {
	if req == nil || resp == nil {
		blog.Errorf("delete cloud failed, req or resp is empty")
		return
	}
	da.ctx = ctx
	da.req = req
	da.resp = resp
	da.resp.Data = &cmproto.DeleteNodeGroupResponseData{}

	// get nodeGroup and all nodes under group
	// IsForce = false, not allow delete directly
	if err := da.validate(); err != nil {
		// validate already setting response
		return
	}

	// set nodeGroup status to deleting
	da.group.Status = common.StatusDeleteNodeGroupDeleting
	if err := da.model.UpdateNodeGroup(da.ctx, da.group); err != nil {
		blog.Errorf("update NodeGroup %s status to deleting failed, %s", da.group.NodeGroupID, err.Error())
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
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
	asOption, _ := actions.GetAsOptionByClusterID(da.model, da.group.ClusterID)

	// get dependency resource for cloudprovider operation
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
		CommonOption:           *cmOption,
		IsForce:                req.IsForce,
		ReserveNodesInCluster:  req.ReserveNodesInCluster,
		ReservedNodeInstance:   req.KeepNodesInstance,
		CleanInstanceInCluster: !req.KeepNodesInstance,
		Operator:               req.Operator,
		AsOption:               asOption,
		Cloud:                  cloud,
		Cluster:                cluster,
		OnlyData:               req.OnlyDeleteInfo,
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
	} else {
		// sync to delete nodeGroup
		if err = da.model.DeleteNodeGroup(da.ctx, da.req.NodeGroupID); err != nil {
			blog.Errorf("DeleteNodeGroup[%s] failed: %v", da.req.NodeGroupID, err)
			da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
			return
		}
		blog.Infof("DeleteNodeGroup[%s] successful", da.req.NodeGroupID)
	}

	resp.Data.NodeGroup = removeSensitiveInfo(da.group)
	resp.Data.Task = task

	err = da.model.CreateOperationLog(da.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   da.group.NodeGroupID,
		TaskID: func() string {
			if task == nil {
				return ""
			}
			return task.TaskID
		}(),
		Message:      fmt.Sprintf("集群%s删除节点池%s", da.group.ClusterID, da.group.NodeGroupID),
		OpUser:       req.Operator,
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    da.group.ClusterID,
		ProjectID:    da.group.ProjectID,
		ResourceName: da.group.GetName(),
	})
	if err != nil {
		blog.Errorf("DeleteNodeGroup[%s] CreateOperationLog failed: %v", da.group.NodeGroupID, err)
	}
	da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
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
	// get nodegroup for validation
	destGroup, err := da.model.GetNodeGroup(da.ctx, da.req.NodeGroupID)
	if err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("Get NodeGroup %s in pre-RemoveNode checking failed, err %s", da.req.NodeGroupID, err.Error())
		return err
	}
	// check cluster info consistency
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
	// check cluster existence
	cluster, err := da.model.GetCluster(da.ctx, destGroup.ClusterID)
	if err != nil {
		blog.Errorf("get Cluster %s for NodeGroup %s to remove Node failed, %s",
			destGroup.ClusterID, destGroup.NodeGroupID, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	da.cluster = cluster
	// get specified node for remove validation
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
	// Nodes validation for remove
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
	// get dependency resource and release it
	cloud, cluster, err := actions.GetCloudAndCluster(da.model, da.group.Provider, da.group.ClusterID)
	if err != nil {
		blog.Errorf("get Cloud %s for NodeGroup %s to remove Node failed, %s",
			da.group.Provider, da.group.NodeGroupID, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	// get dependency resource for cloudprovider operation
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cluster %s:%s to remove Node failed, %s",
			da.group.ClusterID, da.group.NodeGroupID, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	// get cloudprovider and then start to remove
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

	// try to update Node
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
		CreateTime:   time.Now().Format(time.RFC3339),
		ResourceName: da.group.GetName(),
	})
	if err != nil {
		blog.Errorf("RemoveNodesFromNodeGroup[%s] CreateOperationLog failed: %v", da.group.NodeGroupID, err)
	}

	da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// CleanNodesAction action for delete nodes
type CleanNodesAction struct {
	ctx context.Context

	model store.ClusterManagerModel
	req   *cmproto.CleanNodesInGroupRequest
	resp  *cmproto.CleanNodesInGroupResponse

	locker     lock.DistributedLock
	cloud      *cmproto.Cloud
	cluster    *cmproto.Cluster
	group      *cmproto.NodeGroup
	asOption   *cmproto.ClusterAutoScalingOption
	task       *cmproto.Task
	cleanNodes []*cmproto.Node
}

// NewCleanNodesAction create delete action for online cluster credential
func NewCleanNodesAction(model store.ClusterManagerModel, locker lock.DistributedLock) *CleanNodesAction {
	return &CleanNodesAction{
		model:  model,
		locker: locker,
	}
}

func (da *CleanNodesAction) setResp(code uint32, msg string) {
	da.resp.Code = code
	da.resp.Message = msg
	da.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (da *CleanNodesAction) checkNodeGroupClusterID(groupID string) error { // nolint
	// try to get original data for return
	group, err := da.model.GetNodeGroup(da.ctx, da.req.NodeGroupID)
	if err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("Get NodeGroup %s in pre-CleanNode checking failed, err %s", da.req.NodeGroupID, err.Error())
		return err
	}
	da.group = group

	if group.ClusterID != da.req.ClusterID {
		blog.Errorf("request ClusterID %s is not same with NodeGroup.ClusterID %s when CleanNode",
			da.req.ClusterID, group.ClusterID,
		)
		err := fmt.Errorf("request ClusterID is not same with NodeGroup.ClusterID %s", group.ClusterID)
		da.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
		return err
	}

	return nil
}

func (da *CleanNodesAction) allowNodeIfDelete(ip string) bool {
	node, err := da.model.GetNodeByIP(da.ctx, ip)
	if err != nil {
		blog.Errorf("CleanNodesAction allowNodeIfDelete[%s] failed: %v", ip, err)
		return false
	}

	blog.Infof("CleanNodesAction allowNodeIfDelete[%s] status: %v", ip, node.Status)

	if node.Status == common.StatusDeleting || node.Status == common.StatusInitialization {
		blog.Infof("CleanNodesAction allowNodeIfDelete[%s:%s]: %v", ip, node.Status, false)
		return false
	}

	return true
}

func (da *CleanNodesAction) checkCleanNodesExistInCluster() error {
	// get specified node for clean validation
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

	groupNodes := make(map[string]*cmproto.Node)
	for i := range nodes {
		groupNodes[nodes[i].InnerIP] = nodes[i]
	}

	// Nodes validation for clean
	// deleteNodes not exist in db or node DELETING/INITIALIZATION status
	for _, ip := range da.req.Nodes {
		// check different cluster clean nodes
		allow := da.allowNodeIfDelete(ip)
		if !allow {
			continue
		}

		// check same cluster/nodeGroup exist node
		node, ok := groupNodes[ip]
		if !ok || node.Status == common.StatusDeleting || node.Status == common.StatusInitialization {
			blog.Errorf("clean Node %s is Not under NodeGroup %s control or status "+
				"DELETING/INITIALIZATION", ip, da.group.NodeGroupID)
			continue
		}

		da.cleanNodes = append(da.cleanNodes, node)
	}

	if len(da.cleanNodes) == 0 {
		da.setResp(common.BcsErrClusterManagerCACleanNodesEmptyErr, "cleanNodes empty")
		return fmt.Errorf("NodeGroup[%s] cleanNodes empty", da.group.NodeGroupID)
	}

	return nil
}

func (da *CleanNodesAction) validate() error {
	if err := da.req.Validate(); err != nil {
		da.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return err
	}

	// check group clusterID
	err := da.checkNodeGroupClusterID(da.req.NodeGroupID)
	if err != nil {
		return err
	}

	// check clean nodes
	err = da.checkCleanNodesExistInCluster()
	if err != nil {
		return err
	}

	return nil
}

func (da *CleanNodesAction) getRelativeData() error {
	// get dependency resource
	cloud, cluster, err := actions.GetCloudAndCluster(da.model, da.group.Provider, da.group.ClusterID)
	if err != nil {
		blog.Errorf("get Cloud %s Project %s for NodeGroup %s to clean Node failed, %s",
			da.group.Provider, da.group.ProjectID, da.group.NodeGroupID, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	da.cluster = cluster
	da.cloud = cloud

	da.asOption, _ = actions.GetAsOptionByClusterID(da.model, da.group.ClusterID)

	return nil
}

func (da *CleanNodesAction) updateCleanNodesStatus() error {
	// try to update Cluster & NodeGroup
	for _, node := range da.cleanNodes {
		node.Status = common.StatusDeleting

		var (
			err error
		)
		// maybe update node timeout thus solve problem by retry
		err = retry.Do(func() error {
			if errLocal := da.model.UpdateNode(da.ctx, node); errLocal != nil {
				return errLocal
			}

			return nil
		}, retry.Attempts(3), retry.DelayType(retry.FixedDelay), retry.Delay(time.Millisecond*500))
		if err != nil {
			blog.Errorf("update NodeGroup %s with Nodes %v status change to DELETING failed, %s",
				da.group.ClusterID, da.req.Nodes, err.Error(),
			)
			da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
			return err
		}
		blog.Infof("Nodes %v of NodeGroup %s change to DELETING successfully", node.InnerIP, da.group.NodeGroupID)
	}

	return nil
}

func (da *CleanNodesAction) handleTask() error {
	// ready to create background task
	nodeGroupMgr, err := cloudprovider.GetNodeGroupMgr(da.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s/%s for NodeGroup %s to clean Nodes %v failed, %s",
			da.cloud.CloudID, da.cloud.CloudProvider, da.group.NodeGroupID, da.req.Nodes, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return err
	}

	// build clean task and dispatch to run
	task, err := nodeGroupMgr.CleanNodesInGroup(da.cleanNodes, da.group, &cloudprovider.CleanNodesOption{
		Cloud:    da.cloud,
		Cluster:  da.cluster,
		AsOption: da.asOption,
		Operator: da.req.Operator,
	})
	if err != nil {
		blog.Errorf("build clean Node %v task from NodeGroup %s with cloudprovider %s failed, %s",
			da.req.Nodes, da.group.NodeGroupID, da.cloud.CloudProvider, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return err
	}

	if err = da.model.CreateTask(da.ctx, task); err != nil {
		blog.Errorf("save clean Node %v task from NodeGroup %s failed, %s",
			da.req.Nodes, da.group.NodeGroupID, err.Error())
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	if err = taskserver.GetTaskServer().Dispatch(task); err != nil {
		blog.Errorf("dispatch clean Node %v task from NodeGroup %s failed, %s",
			da.req.Nodes, da.group.NodeGroupID, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return err
	}
	blog.Infof("NodeGroup %s clean nodes %v task created with cloudprovider %s/%s successfully",
		da.group.NodeGroupID, da.req.Nodes, da.cloud.CloudID, da.cloud.CloudProvider)

	da.task = task
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

	const (
		cleanNodeGroupNodesLockKey = "/bcs-services/bcs-cluster-manager/CleanNodesAction"
	)
	da.locker.Lock(cleanNodeGroupNodesLockKey, []lock.LockOption{lock.LockTTL(time.Second * 5)}...) // nolint
	defer da.locker.Unlock(cleanNodeGroupNodesLockKey)                                              // nolint

	if err := da.validate(); err != nil {
		// validate already sets response information
		return
	}
	if err := da.getRelativeData(); err != nil {
		return
	}

	// update clean node status
	if err := da.updateCleanNodesStatus(); err != nil {
		return
	}

	// dispatch task
	if err := da.handleTask(); err != nil {
		return
	}

	err := da.model.CreateOperationLog(da.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   da.group.NodeGroupID,
		TaskID:       da.task.TaskID,
		Message:      fmt.Sprintf("集群%s节点池%s删除节点", da.group.ClusterID, da.group.NodeGroupID),
		OpUser:       req.Operator,
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    da.cluster.ClusterID,
		ProjectID:    da.cluster.ProjectID,
		ResourceName: da.group.GetName(),
	})
	if err != nil {
		blog.Errorf("CleanNodesFromNodeGroup[%s] CreateOperationLog failed: %v", da.group.NodeGroupID, err)
	}

	resp.Data = da.task
	da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
