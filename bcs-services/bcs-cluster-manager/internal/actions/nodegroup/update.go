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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// UpdateAction update action for online cluster credential
type UpdateAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	group   *cmproto.NodeGroup
	cluster *cmproto.Cluster
	cloud   *cmproto.Cloud
	req     *cmproto.UpdateNodeGroupRequest
	resp    *cmproto.UpdateNodeGroupResponse
}

// NewUpdateAction create update action for online cluster credential
func NewUpdateAction(model store.ClusterManagerModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

func (ua *UpdateAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ua *UpdateAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}
	if ua.req.NodeGroupID == "" {
		return fmt.Errorf("nodeGroupID is empty")
	}
	if ua.req.ClusterID == "" {
		return fmt.Errorf("clusterID is empty")
	}
	return nil
}

// trans request args to node group
func (ua *UpdateAction) modifyNodeGroupField() {
	timeStr := time.Now().Format(time.RFC3339)
	//update field if required
	group := ua.group
	group.UpdateTime = timeStr
	group.Updater = ua.req.Updater
	if len(ua.req.Name) != 0 {
		group.Name = ua.req.Name
	}
	if ua.req.AutoScaling != nil {
		if ua.req.AutoScaling.MaxSize != 0 {
			group.AutoScaling.MaxSize = ua.req.AutoScaling.MaxSize
		}
		group.AutoScaling.MinSize = ua.req.AutoScaling.MinSize
		if ua.req.AutoScaling.DefaultCooldown != 0 {
			group.AutoScaling.DefaultCooldown = ua.req.AutoScaling.DefaultCooldown
		}
		if ua.req.AutoScaling.SubnetIDs != nil {
			group.AutoScaling.SubnetIDs = ua.req.AutoScaling.SubnetIDs
		}
		if ua.req.AutoScaling.RetryPolicy != "" {
			group.AutoScaling.RetryPolicy = ua.req.AutoScaling.RetryPolicy
		}
		if ua.req.AutoScaling.MultiZoneSubnetPolicy != "" {
			group.AutoScaling.MultiZoneSubnetPolicy = ua.req.AutoScaling.MultiZoneSubnetPolicy
		}
		if ua.req.AutoScaling.ScalingMode != "" {
			group.AutoScaling.ScalingMode = ua.req.AutoScaling.ScalingMode
		}
	}
	if ua.req.LaunchTemplate != nil {
		if ua.req.LaunchTemplate.InstanceType != "" {
			group.LaunchTemplate.InstanceType = ua.req.LaunchTemplate.InstanceType
		}
		if ua.req.LaunchTemplate.InstanceChargeType != "" {
			group.LaunchTemplate.InstanceChargeType = ua.req.LaunchTemplate.InstanceChargeType
		}
		if ua.req.LaunchTemplate.InternetAccess != nil {
			group.LaunchTemplate.InternetAccess = ua.req.LaunchTemplate.InternetAccess
		}
		if ua.req.LaunchTemplate.InitLoginPassword != "" {
			group.LaunchTemplate.InitLoginPassword = ua.req.LaunchTemplate.InitLoginPassword
		}
		if ua.req.LaunchTemplate.SecurityGroupIDs != nil {
			group.LaunchTemplate.SecurityGroupIDs = ua.req.LaunchTemplate.SecurityGroupIDs
		}
		if ua.req.LaunchTemplate.UserData != "" {
			group.LaunchTemplate.UserData = ua.req.LaunchTemplate.UserData
		}
		if ua.req.LaunchTemplate.ImageInfo != nil {
			group.LaunchTemplate.ImageInfo = ua.req.LaunchTemplate.ImageInfo
		}
		group.LaunchTemplate.IsMonitorService = ua.req.LaunchTemplate.IsMonitorService
		group.LaunchTemplate.IsSecurityService = ua.req.LaunchTemplate.IsSecurityService
	}
	if ua.req.NodeTemplate != nil {
		if ua.req.NodeTemplate.NodeTemplateID != "" {
			group.NodeTemplate.NodeTemplateID = ua.req.NodeTemplate.NodeTemplateID
		}
		if ua.req.NodeTemplate.Name != "" {
			group.NodeTemplate.Name = ua.req.NodeTemplate.Name
		}
		if ua.req.NodeTemplate.ProjectID != "" {
			group.NodeTemplate.ProjectID = ua.req.NodeTemplate.ProjectID
		}
		if ua.req.NodeTemplate.DockerGraphPath != "" {
			group.NodeTemplate.DockerGraphPath = ua.req.NodeTemplate.DockerGraphPath
		}
		if ua.req.NodeTemplate.MountTarget != "" {
			group.NodeTemplate.MountTarget = ua.req.NodeTemplate.MountTarget
		}
		if ua.req.NodeTemplate.UserScript != "" {
			group.NodeTemplate.UserScript = ua.req.NodeTemplate.UserScript
		}
		if ua.req.NodeTemplate.DataDisks != nil {
			group.NodeTemplate.DataDisks = ua.req.NodeTemplate.DataDisks
		}
		if ua.req.NodeTemplate.BcsScaleOutAddons != nil {
			group.NodeTemplate.BcsScaleOutAddons = ua.req.NodeTemplate.BcsScaleOutAddons
		}
		if ua.req.NodeTemplate.BcsScaleInAddons != nil {
			group.NodeTemplate.BcsScaleInAddons = ua.req.NodeTemplate.BcsScaleInAddons
		}
		if ua.req.NodeTemplate.ScaleOutExtraAddons != nil {
			group.NodeTemplate.ScaleOutExtraAddons = ua.req.NodeTemplate.ScaleOutExtraAddons
		}
		if ua.req.NodeTemplate.ScaleInExtraAddons != nil {
			group.NodeTemplate.ScaleInExtraAddons = ua.req.NodeTemplate.ScaleInExtraAddons
		}
		if ua.req.NodeTemplate.ModuleID != "" {
			group.NodeTemplate.ModuleID = ua.req.NodeTemplate.ModuleID
		}
		if ua.req.NodeTemplate.ExtraArgs != nil {
			group.NodeTemplate.ExtraArgs = ua.req.NodeTemplate.ExtraArgs
		}
		group.NodeTemplate.UnSchedulable = ua.req.NodeTemplate.UnSchedulable
	}
	group.Labels = ua.req.Labels
	group.Taints = ua.req.Taints
	group.Tags = ua.req.Tags
	if len(ua.req.NodeOS) != 0 {
		group.NodeOS = ua.req.NodeOS
	}
	ua.group = group
}

func (ua *UpdateAction) getRelativeResource() error {
	group, err := ua.model.GetNodeGroup(ua.ctx, ua.req.NodeGroupID)
	if err != nil {
		blog.Errorf("get NodeGroup %s failed, %s", ua.req.NodeGroupID, err.Error())
		return err
	}
	ua.group = group

	cluster, err := ua.model.GetCluster(ua.ctx, ua.req.ClusterID)
	if err != nil {
		blog.Errorf("can not get relative Cluster %s when update NodeGroup", ua.req.ClusterID)
		return fmt.Errorf("get relative cluster %s info err, %s", ua.req.ClusterID, err.Error())
	}
	ua.cluster = cluster

	cloud, err := actions.GetCloudByCloudID(ua.model, ua.group.Provider)
	if err != nil {
		blog.Errorf("can not get relative Cloud %s when update NodeGroup for Cluster %s, %s",
			ua.group.Provider, ua.req.ClusterID, err.Error(),
		)
		return err
	}
	ua.cloud = cloud

	return nil
}

func (ua *UpdateAction) updateCloudNodeGroup() error {
	//get credential for cloudprovider operation
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ua.cloud,
		AccountID: ua.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get Credential for Cloud %s/%s when update NodeGroup %s in Cluster %s failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, ua.group.NodeGroupID, ua.group.ClusterID, err.Error(),
		)
		return err
	}
	//create nodegroup with cloudprovider
	mgr, err := cloudprovider.GetNodeGroupMgr(ua.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get NodeGroup Manager cloudprovider %s/%s for update nodegroup %s in cluster %s failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, ua.group.NodeGroupID, ua.group.ClusterID, err.Error(),
		)
		return err
	}
	cmOption.Region = ua.group.Region
	err = mgr.UpdateNodeGroup(ua.group, cmOption)
	if err != nil {
		blog.Errorf("update nodegroup %s in cluster %s with cloudprovider %s failed, %s",
			ua.group.NodeGroupID, ua.group.ClusterID, ua.cloud.CloudProvider, err.Error(),
		)
		return err
	}
	ua.resp.Data = ua.group
	return nil
}

func (ua *UpdateAction) setNodeGroupUpdating() error {
	ua.group.Status = common.StatusUpdating
	if err := ua.model.UpdateNodeGroup(ua.ctx, ua.group); err != nil {
		blog.Infof("update nodegroup %s failed, %v", ua.group.NodeGroupID, err)
		return err
	}
	blog.Infof("update nodegroup %s successfully", ua.group.NodeGroupID)
	return nil
}

func (ua *UpdateAction) saveDB() error {
	ua.group.Status = common.StatusRunning
	if err := ua.model.UpdateNodeGroup(ua.ctx, ua.group); err != nil {
		blog.Errorf("nodegroup %s update in cloudprovider %s success, but update failed in local storage, %s. detail: %+v",
			ua.group.NodeGroupID, ua.group.Provider, err.Error(), utils.ToJSONString(ua.group),
		)
		return err
	}
	blog.Infof("update nodegroup %s successfully", ua.group.NodeGroupID)
	return nil
}

func (ua *UpdateAction) checkStatus() error {
	// if nodegroup is creating/deleting/deleted, return error
	if ua.group.Status == common.StatusCreating || ua.group.Status == common.StatusDeleting || ua.group.Status == common.StatusDeleted {
		err := fmt.Errorf("nodegroup %s status is not running, can not disable auto scale", ua.group.NodeGroupID)
		return err
	}
	return nil
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

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ua.getRelativeResource(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if err := ua.checkStatus(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ua.setNodeGroupUpdating(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ua.modifyNodeGroupField()
	if err := ua.updateCloudNodeGroup(); err != nil {
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	if err := ua.saveDB(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if err := ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   ua.req.NodeGroupID,
		TaskID:       "",
		Message:      fmt.Sprintf("集群%s节点池%s更新配置信息", ua.req.ClusterID, ua.req.NodeGroupID),
		OpUser:       req.Updater,
		CreateTime:   time.Now().String(),
	}); err != nil {
		blog.Errorf("UpdateNodeGroup[%s] CreateOperationLog failed: %v", ua.req.NodeGroupID, err)
	}

	ua.resp.Data = ua.group
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
	cluster   *cmproto.Cluster
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
	// get cluster info
	cluster, err := ua.model.GetCluster(ua.ctx, ua.group.ClusterID)
	if err != nil {
		blog.Errorf("get Cluster %s for NodeGroup %s to move node %s failed, %s",
			ua.group.ClusterID, ua.group.NodeGroupID, ua.req.Nodes, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	ua.cluster = cluster
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
	if err := ua.moveCloudNodeGroupNodes(); err != nil {
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	//try to update Node
	for _, node := range ua.moveNodes {
		node.NodeGroupID = ua.group.NodeGroupID
		if err := ua.model.UpdateNode(ctx, node); err != nil {
			blog.Errorf("update NodeGroup %s with Nodes %s move in failed, %s",
				ua.group.NodeGroupID, node.InnerIP, err.Error(),
			)
			ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
			return
		}
		blog.Infof("Nodes %s remove in NodeGroup %s record successfully", node.InnerIP, ua.group.NodeGroupID)
	}

	err := ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   req.NodeGroupID,
		TaskID:       "",
		Message:      fmt.Sprintf("集群%s移入节点至节点池%s", ua.cluster.ClusterID, req.NodeGroupID),
		OpUser:       ua.group.Updater,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("MoveNodesToGroup[%s] CreateOperationLog failed: %v", req.NodeGroupID, err)
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

func (ua *MoveNodeAction) moveCloudNodeGroupNodes() error {
	//try to move node in cloudprovider
	cloud, cluster, err := actions.GetCloudAndCluster(ua.model, ua.group.Provider, ua.group.ClusterID)
	if err != nil {
		blog.Errorf("get cloud %s and project %s when move nodes %v to NodeGroup %s failed, %s",
			ua.group.Provider, ua.group.ProjectID, ua.req.Nodes, ua.group.NodeGroupID, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
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
		return err
	}
	mgr, err := cloudprovider.GetNodeGroupMgr(cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s NodeGroupMgr when move nodes %v to NodeGroup %s failed, %s",
			ua.group.Provider, ua.req.Nodes, ua.group.NodeGroupID, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return err
	}
	cmOption.Region = ua.group.Region
	task, err := mgr.MoveNodesToGroup(ua.moveNodes, ua.group, &cloudprovider.MoveNodesOption{
		CommonOption: *cmOption,
		Cluster:      ua.cluster,
	})
	if err != nil {
		blog.Errorf("move Node %v to NodeGroup %s with cloudprovider %s failed, %s",
			ua.req.Nodes, ua.group.NodeGroupID, cloud.CloudProvider, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return err
	}
	// create task and dispatch task
	ua.resp.Data = task
	if err := ua.model.CreateTask(ua.ctx, task); err != nil {
		blog.Errorf("save move nodes to node group task for cluster %s failed, %s",
			ua.group.ClusterID, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	if err := taskserver.GetTaskServer().Dispatch(task); err != nil {
		blog.Errorf("dispatch move nodes to node group task for cluster %s failed, %s",
			ua.group.ClusterID, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerTaskErr, err.Error())
		return err
	}
	blog.Infof(
		"Nodes %v move to NodeGroup %s in cloudprovider %s/%s successfully",
		ua.req.Nodes, ua.group.NodeGroupID, cloud.CloudID, cloud.CloudProvider,
	)
	return nil
}

// UpdateDesiredNodeAction update action for desired nodes
type UpdateDesiredNodeAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.UpdateGroupDesiredNodeRequest
	resp  *cmproto.UpdateGroupDesiredNodeResponse

	group   *cmproto.NodeGroup
	cluster *cmproto.Cluster
	cloud   *cmproto.Cloud
	task    *cmproto.Task
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
	// validate nodegroup exist
	group, err := ua.model.GetNodeGroup(ua.ctx, ua.req.NodeGroupID)
	if err != nil {
		blog.Errorf("get NodeGroup %s failed when updateDesiredNode to %d, %s",
			ua.req.NodeGroupID, ua.req.DesiredNode, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	ua.group = group

	// validate req.DesiredNode by NodeGroup.DesiredSize
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

	// build scale nodes task and dispatch to run
	task, err := mgr.BuildUpdateDesiredNodesTask(scaling, ua.group, &cloudprovider.UpdateDesiredNodeOption{
		Cloud:     ua.cloud,
		Cluster:   ua.cluster,
		NodeGroup: ua.group,
		Operator:  ua.req.Operator,
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

	ua.task = task
	ua.resp.Data = task
	blog.Infof("scaling %d node, %v desired node task for NodeGroup successfully for %s", scaling, ua.req.DesiredNode, ua.group.NodeGroupID)
	return nil
}

func (ua *UpdateDesiredNodeAction) getRelativeData() error {
	cloud, cluster, err := actions.GetCloudAndCluster(ua.model, ua.group.Provider, ua.group.ClusterID)
	if err != nil {
		blog.Errorf("get cloud %s and project %s when updateDesiredNode %d in NodeGroup %s failed, %s",
			ua.group.Provider, ua.group.ProjectID, ua.req.DesiredNode, ua.group.NodeGroupID, err.Error(),
		)
		return err
	}
	ua.cluster = cluster
	ua.cloud = cloud

	return nil
}

// returnCurrentScaleNodesNum count
func (ua *UpdateDesiredNodeAction) returnCurrentScaleNodesNum() (uint32, error) {
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ua.cloud,
		AccountID: ua.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential from cloud %s when updateDesiredNode %d in NodeGroup %s failed, %s",
			ua.group.Provider, ua.req.DesiredNode, ua.group.NodeGroupID, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return 0, err
	}
	mgr, err := cloudprovider.GetNodeGroupMgr(ua.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s NodeGroupMgr when updateDesiredNode %d in NodeGroup %s failed, %s",
			ua.group.Provider, ua.req.DesiredNode, ua.group.NodeGroupID, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return 0, err
	}
	cmOption.Region = ua.group.Region
	// pay more attention, in order to compatible with aws/tencentcloud/blueking
	// implementation, no common UpdateDesiredNodes task flow definition, just
	// try to encapsulate in cloudprovider implementation
	scaleResp, err := mgr.UpdateDesiredNodes(ua.req.DesiredNode, ua.group, &cloudprovider.UpdateDesiredNodeOption{
		CommonOption: *cmOption,
		Cluster:      ua.cluster,
		Cloud:        ua.cloud,
	})
	if err != nil {
		blog.Errorf("updateDesiredNode to %d for NodeGroup %s with cloudprovider %s failed, %s",
			ua.req.DesiredNode, ua.group.NodeGroupID, ua.cloud.CloudProvider, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return 0, err
	}

	return scaleResp.ScalingUp, nil
}

func (ua *UpdateDesiredNodeAction) updateNodeGroupDesiredSize(desiredNode uint32) error {
	ua.group.AutoScaling.DesiredSize = desiredNode

	// update DesiredSize in local storage
	if err := ua.model.UpdateNodeGroup(ua.ctx, ua.group); err != nil {
		blog.Errorf("updateDesiredNode %d to NodeGroup %s in local storage failed, %s",
			ua.req.DesiredNode, ua.req.NodeGroupID, err.Error(),
		)
		return err
	}

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
	if err := ua.getRelativeData(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// update DesiredNode with cloud provider
	scaleResp, err := ua.returnCurrentScaleNodesNum()
	if err != nil {
		blog.Errorf("udpateDesiredNode to %d for NodeGroup %s with cloudprovider %s failed, %s",
			req.DesiredNode, ua.group.NodeGroupID, ua.cloud.CloudProvider, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	// update nodeGroup size
	err = ua.updateNodeGroupDesiredSize(req.DesiredNode)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// handler cloud update desired node
	if err := ua.handleTask(scaleResp); err != nil {
		ua.setResp(common.BcsErrClusterManagerTaskErr, err.Error())
		return
	}

	// record operation log
	err = ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   req.NodeGroupID,
		TaskID:       ua.task.TaskID,
		Message:      fmt.Sprintf("集群%s扩容节点池%s节点数至%v", ua.cluster.ClusterID, req.NodeGroupID, req.DesiredNode),
		OpUser:       ua.group.Updater,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("UpdateDesiredNode[%s] CreateOperationLog failed: %v", req.NodeGroupID, err)
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
