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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/taskserver"
)

// UpdateAction update action for online cluster credential
type UpdateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.UpdateNodeGroupRequest
	resp  *cmproto.UpdateNodeGroupResponse
}

// NewUpdateAction create update action for online cluster credential
func NewUpdateAction(model store.ClusterManagerModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

func (ua *UpdateAction) updateCloudNodeGroup(group *cmproto.NodeGroup) error {
	cloud, cluster, err := actions.GetCloudAndCluster(ua.model, group.Provider, group.ClusterID)
	if err != nil {
		blog.Errorf("get Cloud %s and Project %s failed when update NodeGroup %s, %s",
			group.Provider, group.ProjectID, group.NodeGroupID, err.Error(),
		)
		return err
	}
	//get credential for cloudprovider operation
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get Credential for Cloud %s/%s when update NodeGroup %s in Cluster %s failed, %s",
			cloud.CloudID, cloud.CloudProvider, group.NodeGroupID, group.ClusterID, err.Error(),
		)
		return err
	}
	//create nodegroup with cloudprovider
	mgr, err := cloudprovider.GetNodeGroupMgr(cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get NodeGroup Manager cloudprovider %s/%s for update nodegroup %s in cluster %s failed, %s",
			cloud.CloudID, cloud.CloudProvider, group.NodeGroupID, group.ClusterID, err.Error(),
		)
		return err
	}
	cmOption.Region = group.Region
	if err := mgr.UpdateNodeGroup(group, cmOption); err != nil {
		blog.Errorf("update nodegroup %s in cluster %s with cloudprovider %s failed, %s",
			group.NodeGroupID, group.ClusterID, cloud.CloudProvider, err.Error(),
		)
		return err
	}
	return nil
}

func (ua *UpdateAction) updateNodeGroupField(group *cmproto.NodeGroup) error {
	timeStr := time.Now().Format(time.RFC3339)
	//update field if required
	group.UpdateTime = timeStr
	group.Updater = ua.req.Updater
	if len(ua.req.Name) != 0 {
		group.Name = ua.req.Name
	}
	if len(ua.req.ClusterID) != 0 {
		group.ClusterID = ua.req.ClusterID
	}
	if len(ua.req.Region) != 0 {
		group.Region = ua.req.Region
	}
	if ua.req.EnableAutoscale != nil && ua.req.EnableAutoscale.GetValue() != group.EnableAutoscale {
		group.EnableAutoscale = ua.req.EnableAutoscale.GetValue()
	}
	if ua.req.AutoScaling != nil {
		group.AutoScaling = ua.req.AutoScaling
	}
	if ua.req.LaunchTemplate != nil {
		group.LaunchTemplate = ua.req.LaunchTemplate
	}
	if ua.req.Labels != nil {
		group.Labels = ua.req.Labels
	}
	if ua.req.Taints != nil {
		group.Taints = ua.req.Taints
	}
	if len(ua.req.NodeOS) != 0 {
		group.NodeOS = ua.req.NodeOS
	}
	if len(ua.req.Provider) != 0 {
		group.Provider = ua.req.Provider
	}
	if len(ua.req.ConsumerID) != 0 {
		group.ConsumerID = ua.req.ConsumerID
	}

	return nil
}

func (ua *UpdateAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handle update cluster credential
func (ua *UpdateAction) Handle(
	ctx context.Context, req *cmproto.UpdateNodeGroupRequest, resp *cmproto.UpdateNodeGroupResponse) {

	if req == nil || resp == nil {
		blog.Errorf("update cloud failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := req.Validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	//get old project information, update fields if required
	destGroup, err := ua.model.GetNodeGroup(ua.ctx, req.NodeGroupID)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("find nodegroup %s failed when pre-update checking, err %s", req.NodeGroupID, err.Error())
		return
	}
	if err := ua.updateNodeGroupField(destGroup); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ua.updateCloudNodeGroup(destGroup); err != nil {
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	if err = ua.model.UpdateNodeGroup(ctx, destGroup); err != nil {
		blog.Errorf("nodegroup %s update in cloudprovider %s success, but update failed in local storage, %s. detail: %+v",
			destGroup.NodeGroupID, destGroup.Provider, err.Error(), destGroup,
		)
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	blog.Infof("update nodegroup %s successfully", destGroup.NodeGroupID)

	err = ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   destGroup.NodeGroupID,
		TaskID:       "",
		Message:      fmt.Sprintf("集群%s节点池%s更新配置信息", destGroup.ClusterID, destGroup.NodeGroupID),
		OpUser:       req.Updater,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("UpdateNodeGroup[%s] CreateOperationLog failed: %v", destGroup.NodeGroupID, err)
	}

	ua.resp.Data = destGroup
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

// MoveNodeAction move nodes to nodegroup
type MoveNodeAction struct {
	ctx       context.Context
	model     store.ClusterManagerModel
	req       *cmproto.MoveNodesToGroupRequest
	resp      *cmproto.MoveNodesToGroupResponse
	group     *cmproto.NodeGroup
	moveNodes []*cmproto.Node
}

// NewMoveNodeAction create update action for move nodes to nodegroup
func NewMoveNodeAction(model store.ClusterManagerModel) *MoveNodeAction {
	return &MoveNodeAction{
		model: model,
	}
}

func (ua *MoveNodeAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ua *MoveNodeAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return err
	}
	//get nodegroup for validation
	destGroup, err := ua.model.GetNodeGroup(ua.ctx, ua.req.NodeGroupID)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("Get NodeGroup %s in pre-MoveNode checking failed, err %s", ua.req.NodeGroupID, err.Error())
		return err
	}
	//check cluster info consistency
	if destGroup.ClusterID != ua.req.ClusterID {
		blog.Errorf(
			"request ClusterID %s is not same with NodeGroup.ClusterID %s when MoveNode",
			ua.req.ClusterID, destGroup.ClusterID,
		)
		ua.setResp(
			common.BcsErrClusterManagerInvalidParameter,
			fmt.Sprintf("request ClusterID is not same with NodeGroup.ClusterID %s", destGroup.ClusterID),
		)
		return err
	}
	ua.group = destGroup
	//get specified node for move validation
	condM := make(operator.M)
	condM["clusterid"] = ua.group.ClusterID
	cond := operator.NewLeafCondition(operator.Eq, condM)
	nodes, err := ua.model.ListNode(ua.ctx, cond, &options.ListOption{})
	if err != nil {
		blog.Errorf("get Cluster %s all Nodes failed when MoveNode, %s", ua.group.ClusterID, err.Error())
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	allNodes := make(map[string]*cmproto.Node)
	for i := range nodes {
		allNodes[nodes[i].InnerIP] = nodes[i]
	}
	for _, ip := range ua.req.Nodes {
		node, ok := allNodes[ip]
		if !ok {
			blog.Errorf("move node %s is not under Cluster %s when MoveNodeToNodeGroup %s",
				ip, ua.group.ClusterID, ua.group.NodeGroupID,
			)
			err := fmt.Errorf("move node %s is not under cluster %s", ip, ua.group.ClusterID)
			ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
			return err
		}
		ua.moveNodes = append(ua.moveNodes, node)
	}
	return nil
}

// Handle handle update cluster credential
func (ua *MoveNodeAction) Handle(
	ctx context.Context, req *cmproto.MoveNodesToGroupRequest, resp *cmproto.MoveNodesToGroupResponse) {

	if req == nil || resp == nil {
		blog.Errorf("move nodes to NodeGroup failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		//valiate already setting response message
		return
	}
	//try to move node in cloudprovider
	cloud, cluster, err := actions.GetCloudAndCluster(ua.model, ua.group.Provider, ua.group.ClusterID)
	if err != nil {
		blog.Errorf("get cloud %s and project %s when move nodes %v to NodeGroup %s failed, %s",
			ua.group.Provider, ua.group.ProjectID, ua.req.Nodes, ua.group.NodeGroupID, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential from cloud %s and project %s when move nodes %v to NodeGroup %s failed, %s",
			ua.group.Provider, ua.group.ProjectID, ua.req.Nodes, ua.group.NodeGroupID, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	mgr, err := cloudprovider.GetNodeGroupMgr(cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s NodeGroupMgr when move nodes %v to NodeGroup %s failed, %s",
			ua.group.Provider, ua.req.Nodes, ua.group.NodeGroupID, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	cmOption.Region = ua.group.Region
	if err = mgr.MoveNodesToGroup(ua.moveNodes, ua.group, &cloudprovider.MoveNodesOption{
		CommonOption: *cmOption,
		Cluster:      cluster,
	}); err != nil {
		blog.Errorf("move Node %v to NodeGroup %s with cloudprovider %s failed, %s",
			req.Nodes, ua.group.NodeGroupID, cloud.CloudProvider, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	blog.Infof(
		"Nodes %v move to NodeGroup %s in cloudprovider %s/%s successfully",
		req.Nodes, ua.group.NodeGroupID, cloud.CloudID, cloud.CloudProvider,
	)
	//try to update Node
	for _, node := range ua.moveNodes {
		node.NodeGroupID = ua.group.NodeGroupID
		if err = ua.model.UpdateNode(ctx, node); err != nil {
			blog.Errorf("update NodeGroup %s with Nodes %s move in failed, %s",
				ua.group.NodeGroupID, node.InnerIP, err.Error(),
			)
			ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
			return
		}
		blog.Infof("Nodes %s remove in NodeGroup %s record successfully", node.InnerIP, ua.group.NodeGroupID)
	}

	err = ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   req.NodeGroupID,
		TaskID:       "",
		Message:      fmt.Sprintf("集群%s移入节点至节点池%s", cluster.ClusterID, req.NodeGroupID),
		OpUser:       ua.group.Updater,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("MoveNodesToGroup[%s] CreateOperationLog failed: %v", req.NodeGroupID, err)
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

// UpdateDesiredNodeAction update action for desired nodes
type UpdateDesiredNodeAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	req     *cmproto.UpdateGroupDesiredNodeRequest
	resp    *cmproto.UpdateGroupDesiredNodeResponse
	group   *cmproto.NodeGroup
	cluster *cmproto.Cluster
	cloud   *cmproto.Cloud
}

// NewUpdateDesiredNodeAction create update action for online cluster credential
func NewUpdateDesiredNodeAction(model store.ClusterManagerModel) *UpdateDesiredNodeAction {
	return &UpdateDesiredNodeAction{
		model: model,
	}
}

func (ua *UpdateDesiredNodeAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ua *UpdateDesiredNodeAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return err
	}
	//validate nodegroup existence
	group, err := ua.model.GetNodeGroup(ua.ctx, ua.req.NodeGroupID)
	if err != nil {
		blog.Errorf("get NodeGroup %s failed when updateDesiredNode to %d, %s",
			ua.req.NodeGroupID, ua.req.DesiredNode, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	ua.group = group
	//valiate req.DesiredNode by NodeGroup.DesiredSize
	if ua.req.DesiredNode < group.AutoScaling.MinSize || ua.req.DesiredNode > group.AutoScaling.MaxSize {
		blog.Errorf("NodeGroup %s update DesiredNode %d is invalid, must in [%d, %d]",
			group.NodeGroupID, ua.req.DesiredNode, group.AutoScaling.MinSize, group.AutoScaling.MaxSize)
		retErr := fmt.Errorf("desiredNode is invalid, must in [%d, %d]",
			group.AutoScaling.MinSize, group.AutoScaling.MaxSize)
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, retErr.Error())
		return retErr
	}
	return nil
}

func (ua *UpdateDesiredNodeAction) handleTask(scaling uint32) error {
	mgr, err := cloudprovider.GetTaskManager(ua.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when updateDesiredNode %d in NodeGroup %s failed, %s",
			ua.group.Provider, ua.req.DesiredNode, ua.group.NodeGroupID, err.Error(),
		)
		return err
	}

	// build clean task and dispatch to run
	task, err := mgr.BuildScalingNodesTask(scaling, ua.group, &cloudprovider.TaskOptions{
		Cloud:    ua.cloud,
		Cluster:  ua.cluster,
		Operator: ua.req.Operator,
	})
	if err != nil {
		blog.Errorf("build scaling task for NodeGroup %s with cloudprovider %s failed, %s",
			ua.group.NodeGroupID, ua.cloud.CloudProvider, err.Error(),
		)
		return err
	}
	if err := ua.model.CreateTask(ua.ctx, task); err != nil {
		blog.Errorf("save scaling task for NodeGroup %s failed, %s",
			ua.group.NodeGroupID, err.Error(),
		)
		return err
	}
	if err := taskserver.GetTaskServer().Dispatch(task); err != nil {
		blog.Errorf("dispatch scaling task for NodeGroup %s failed, %s",
			ua.group.NodeGroupID, err.Error(),
		)
		return err
	}
	ua.resp.Data = task
	blog.Infof("scaling %d node task for NodeGroup successfully for %s", scaling, ua.group.NodeGroupID)
	return nil
}

// Handle handle update cluster credential
func (ua *UpdateDesiredNodeAction) Handle(
	ctx context.Context, req *cmproto.UpdateGroupDesiredNodeRequest, resp *cmproto.UpdateGroupDesiredNodeResponse) {

	if req == nil || resp == nil {
		blog.Errorf("update NodeGroup failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		// validation already setting response
		return
	}
	// update DesiredNode with cloud provider
	cloud, cluster, err := actions.GetCloudAndCluster(ua.model, ua.group.Provider, ua.group.ClusterID)
	if err != nil {
		blog.Errorf("get cloud %s and project %s when updateDesiredNode %d in NodeGroup %s failed, %s",
			ua.group.Provider, ua.group.ProjectID, ua.req.DesiredNode, ua.group.NodeGroupID, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ua.cloud = cloud
	ua.cluster = cluster

	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential from cloud %s and project %s when updateDesiredNode %d in NodeGroup %s failed, %s",
			ua.group.Provider, ua.group.ProjectID, ua.req.DesiredNode, ua.group.NodeGroupID, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	mgr, err := cloudprovider.GetNodeGroupMgr(cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s NodeGroupMgr when updateDesiredNode %d in NodeGroup %s failed, %s",
			ua.group.Provider, ua.req.DesiredNode, ua.group.NodeGroupID, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	cmOption.Region = ua.group.Region
	// pay more attention, in order to compatible with aws/tencentcloud/blueking
	// implementation, no common UpdateDesiredNodes task flow definition, just
	// try to encapsulate in cloudprovider implementation
	scaleResp, err := mgr.UpdateDesiredNodes(req.DesiredNode, ua.group, &cloudprovider.UpdateDesiredNodeOption{
		CommonOption: *cmOption,
		Cluster:      cluster,
		Cloud:        cloud,
	})
	if err != nil {
		blog.Errorf("udpateDesiredNode to %d for NodeGroup %s with cloudprovider %s failed, %s",
			req.DesiredNode, ua.group.NodeGroupID, cloud.CloudProvider, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	ua.group.AutoScaling.DesiredSize = req.DesiredNode
	//update DesiredSize in local storage
	if err = ua.model.UpdateNodeGroup(ctx, ua.group); err != nil {
		blog.Errorf("updateDesiredNode %d to NodeGroup %s in local storage failed, %s",
			req.DesiredNode, req.NodeGroupID, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	if err := ua.handleTask(scaleResp.ScalingUp); err != nil {
		ua.setResp(common.BcsErrClusterManagerTaskErr, err.Error())
		return
	}

	err = ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   req.NodeGroupID,
		TaskID:       "",
		Message:      fmt.Sprintf("集群%s移入节点至节点池%s", cluster.ClusterID, req.NodeGroupID),
		OpUser:       ua.group.Updater,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("MoveNodesToGroup[%s] CreateOperationLog failed: %v", req.NodeGroupID, err)
	}

	blog.Infof("updateDesiredNode %d to NodeGroup %s successfully", req.DesiredNode, req.NodeGroupID)
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

// UpdateDesiredSizeAction update nodegroup autoscaling desiredSize
type UpdateDesiredSizeAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.UpdateGroupDesiredSizeRequest
	resp  *cmproto.UpdateGroupDesiredSizeResponse
}

// NewUpdateDesiredSizeAction create update action for update
func NewUpdateDesiredSizeAction(model store.ClusterManagerModel) *UpdateDesiredSizeAction {
	return &UpdateDesiredSizeAction{
		model: model,
	}
}

func (ua *UpdateDesiredSizeAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handle update cluster credential
func (ua *UpdateDesiredSizeAction) Handle(
	ctx context.Context, req *cmproto.UpdateGroupDesiredSizeRequest, resp *cmproto.UpdateGroupDesiredSizeResponse) {

	if req == nil || resp == nil {
		blog.Errorf("update nodegroup desiredSize failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := req.Validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	// get old project information, update fields if required
	destGroup, err := ua.model.GetNodeGroup(ua.ctx, req.NodeGroupID)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("find nodegroup %s failed when pre-update checking, err %s", req.NodeGroupID, err.Error())
		return
	}
	destGroup.AutoScaling.DesiredSize = req.DesiredSize

	if err := ua.model.UpdateNodeGroup(ctx, destGroup); err != nil {
		blog.Errorf("nodegroup %s update desiredSize failed in local storage, %s",
			destGroup.NodeGroupID, err.Error())
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err = ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   req.NodeGroupID,
		TaskID:       "",
		Message:      fmt.Sprintf("更新集群%s节点池%s期望扩容节点数至%d", destGroup.ClusterID, req.NodeGroupID, req.DesiredSize),
		OpUser:       req.Operator,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("UpdateGroupDesiredSize[%s] CreateOperationLog failed: %v", req.NodeGroupID, err)
	}

	blog.Infof("update nodegroup desiredSize %s successfully", destGroup.NodeGroupID)
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
