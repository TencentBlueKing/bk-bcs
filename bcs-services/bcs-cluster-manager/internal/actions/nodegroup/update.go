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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/lock"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/taskserver"
	iutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// UpdateAction update action for online cluster credential
type UpdateAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	group   *cmproto.NodeGroup
	cluster *cmproto.Cluster
	cloud   *cmproto.Cloud
	task    *cmproto.Task
	req     *cmproto.UpdateNodeGroupRequest
	resp    *cmproto.UpdateNodeGroupResponse
}

// NewUpdateAction create update action for online cluster credential
func NewUpdateAction(model store.ClusterManagerModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

// setResp resp body
func (ua *UpdateAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = code == common.BcsErrClusterManagerSuccess
}

// validate check
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
	if ua.req.AutoScaling != nil && ua.req.AutoScaling.MinSize > ua.req.AutoScaling.MaxSize {
		return fmt.Errorf("updateAction AutoScaling MinSize(%v) lt MaxSize(%v)",
			ua.req.AutoScaling.MinSize, ua.req.AutoScaling.MaxSize)
	}

	return nil
}

// modifyNodeGroupField trans request args to node group
func (ua *UpdateAction) modifyNodeGroupField() {
	timeStr := time.Now().Format(time.RFC3339)
	// update field if required
	group := ua.group
	group.UpdateTime = timeStr
	group.Updater = ua.req.Updater
	if len(ua.req.Name) != 0 {
		group.Name = ua.req.Name
	}
	if len(ua.req.Region) != 0 {
		group.Region = ua.req.Region
	}
	if ua.req.EnableAutoscale != nil {
		group.EnableAutoscale = ua.req.EnableAutoscale.GetValue()
	}
	if len(ua.req.Labels) > 0 {
		group.Labels = ua.req.Labels
	}
	if len(ua.req.Tags) > 0 {
		group.Tags = ua.req.Tags
	}
	if len(ua.req.ConsumerID) > 0 {
		group.ConsumerID = ua.req.ConsumerID
	}
	if len(ua.req.GetExtraInfo()) > 0 {
		group.ExtraInfo = ua.req.GetExtraInfo()
	}

	// autoscaling
	ua.modifyNodeGroupAutoScaling(group)
	// launch template
	ua.modifyNodeGroupLaunchTemplate(group)
	// nodeTemplate
	ua.modifyNodeGroupNodeTemplate(group)
	if len(ua.req.NodeOS) != 0 {
		group.NodeOS = ua.req.NodeOS
	}
	if ua.req.BkCloudID != nil {
		group.Area = &cmproto.CloudArea{
			BkCloudID: ua.req.BkCloudID.GetValue(),
			BkCloudName: func() string {
				if ua.req.CloudAreaName != nil {
					return ua.req.CloudAreaName.GetValue()
				}
				return ""
			}(),
		}
	}
	group.Labels = ua.req.Labels
	group.Tags = ua.req.Tags

	ua.group = group
}

// modifyNodeGroupAutoScaling autoscaling field
func (ua *UpdateAction) modifyNodeGroupAutoScaling(group *cmproto.NodeGroup) {
	if ua.req.AutoScaling != nil {
		if group.AutoScaling == nil {
			group.AutoScaling = &cmproto.AutoScalingGroup{}
		}
		group.AutoScaling.MinSize = ua.req.AutoScaling.MinSize
		if ua.req.AutoScaling.MaxSize != 0 {
			group.AutoScaling.MaxSize = ua.req.AutoScaling.MaxSize
		}
		if ua.req.AutoScaling.VpcID != "" {
			group.AutoScaling.VpcID = ua.req.AutoScaling.VpcID
		}
		if ua.req.AutoScaling.DefaultCooldown != 0 {
			group.AutoScaling.DefaultCooldown = ua.req.AutoScaling.DefaultCooldown
		}
		if ua.req.AutoScaling.SubnetIDs != nil {
			group.AutoScaling.SubnetIDs = ua.req.AutoScaling.SubnetIDs
		}
		if ua.req.AutoScaling.Zones != nil {
			group.AutoScaling.Zones = ua.req.AutoScaling.Zones
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
		if len(ua.req.AutoScaling.TimeRanges) > 0 {
			group.AutoScaling.TimeRanges = ua.req.AutoScaling.TimeRanges
		}
	}
}

// modifyNodeGroupLaunchTemplate launchTemplate field
func (ua *UpdateAction) modifyNodeGroupLaunchTemplate(group *cmproto.NodeGroup) {
	if ua.req.LaunchTemplate != nil {
		if group.LaunchTemplate == nil {
			group.LaunchTemplate = &cmproto.LaunchConfiguration{}
		}
		if ua.req.LaunchTemplate.CPU > 0 {
			group.LaunchTemplate.CPU = ua.req.LaunchTemplate.CPU
		}
		if ua.req.LaunchTemplate.Mem > 0 {
			group.LaunchTemplate.Mem = ua.req.LaunchTemplate.Mem
		}
		if ua.req.LaunchTemplate.GPU > 0 {
			group.LaunchTemplate.GPU = ua.req.LaunchTemplate.GPU
		}
		if ua.req.LaunchTemplate.InstanceType != "" {
			group.LaunchTemplate.InstanceType = ua.req.LaunchTemplate.InstanceType
		}
		if ua.req.LaunchTemplate.InstanceChargeType != "" {
			group.LaunchTemplate.InstanceChargeType = ua.req.LaunchTemplate.InstanceChargeType
		}
		if ua.req.LaunchTemplate.InternetAccess != nil {
			group.LaunchTemplate.InternetAccess = ua.req.LaunchTemplate.InternetAccess
		}
		if ua.req.LaunchTemplate.InitLoginUsername != "" {
			group.LaunchTemplate.InitLoginUsername = ua.req.LaunchTemplate.InitLoginUsername
		}
		// not allow update passwd or ssh key
		/*
			if ua.req.LaunchTemplate.InitLoginPassword != "" {
				group.LaunchTemplate.InitLoginPassword = ua.req.LaunchTemplate.InitLoginPassword
			}
			if ua.req.LaunchTemplate.KeyPair != nil {
				if group.LaunchTemplate.KeyPair == nil {
					group.LaunchTemplate.KeyPair = &cmproto.KeyInfo{}
				}
				if len(ua.req.LaunchTemplate.KeyPair.GetKeyID()) > 0 {
					group.LaunchTemplate.KeyPair.KeyID = ua.req.LaunchTemplate.KeyPair.GetKeyID()
				}
				if len(ua.req.LaunchTemplate.KeyPair.GetKeySecret()) > 0 {
					group.LaunchTemplate.KeyPair.KeySecret = utils.Base64Encode(ua.req.LaunchTemplate.KeyPair.GetKeySecret())
				}
			}
		*/
		if ua.req.LaunchTemplate.SecurityGroupIDs != nil {
			group.LaunchTemplate.SecurityGroupIDs = ua.req.LaunchTemplate.SecurityGroupIDs
		}
		if ua.req.LaunchTemplate.UserData != "" {
			group.LaunchTemplate.UserData = ua.req.LaunchTemplate.UserData
		}
		if ua.req.LaunchTemplate.ImageInfo != nil {
			group.LaunchTemplate.ImageInfo = ua.req.LaunchTemplate.ImageInfo
		}
		if ua.req.LaunchTemplate.SystemDisk != nil {
			group.LaunchTemplate.SystemDisk = ua.req.LaunchTemplate.SystemDisk
		}
		if ua.req.LaunchTemplate.DataDisks != nil {
			group.LaunchTemplate.DataDisks = ua.req.LaunchTemplate.DataDisks
		}
		group.LaunchTemplate.Selector = ua.req.LaunchTemplate.Selector
		group.LaunchTemplate.IsMonitorService = ua.req.LaunchTemplate.IsMonitorService
		group.LaunchTemplate.IsSecurityService = ua.req.LaunchTemplate.IsSecurityService
	}
}

// modifyNodeGroupNodeTemplate nodeTemplate field
func (ua *UpdateAction) modifyNodeGroupNodeTemplate(group *cmproto.NodeGroup) {
	if ua.req.NodeTemplate != nil {
		if group.NodeTemplate == nil {
			group.NodeTemplate = &cmproto.NodeTemplate{}
		}
		if ua.req.NodeTemplate.NodeTemplateID != "" {
			group.NodeTemplate.NodeTemplateID = ua.req.NodeTemplate.NodeTemplateID
		}
		if ua.req.NodeTemplate.Name != "" {
			group.NodeTemplate.Name = ua.req.NodeTemplate.Name
		}
		if ua.req.NodeTemplate.ProjectID != "" {
			group.NodeTemplate.ProjectID = ua.req.NodeTemplate.ProjectID
		}
		if len(ua.req.NodeTemplate.Labels) > 0 {
			group.NodeTemplate.Labels = ua.req.NodeTemplate.Labels
		}
		if len(ua.req.NodeTemplate.Taints) > 0 {
			group.NodeTemplate.Taints = ua.req.NodeTemplate.Taints
		}
		if ua.req.NodeTemplate.DockerGraphPath != "" {
			group.NodeTemplate.DockerGraphPath = ua.req.NodeTemplate.DockerGraphPath
		}
		if ua.req.NodeTemplate.MountTarget != "" {
			group.NodeTemplate.MountTarget = ua.req.NodeTemplate.MountTarget
		}
		if ua.req.NodeTemplate.DataDisks != nil {
			group.NodeTemplate.DataDisks = ua.req.NodeTemplate.DataDisks
		}
		if ua.req.NodeTemplate.UserScript != "" {
			group.NodeTemplate.UserScript = iutils.Base64Encode(ua.req.NodeTemplate.UserScript)
		} else {
			group.NodeTemplate.UserScript = ua.req.NodeTemplate.UserScript
		}
		// 组件参数设置: key 为组件标识，value 各模块进程启动参数，多个参数之间使用;间隔，例如Kubelet: root-dir=/var/lib/kubelet;"
		if len(ua.req.NodeTemplate.ExtraArgs) > 0 {
			group.NodeTemplate.ExtraArgs = ua.req.NodeTemplate.ExtraArgs
		}
		if ua.req.NodeTemplate.PreStartUserScript != "" {
			group.NodeTemplate.PreStartUserScript = iutils.Base64Encode(ua.req.NodeTemplate.PreStartUserScript)
		} else {
			group.NodeTemplate.PreStartUserScript = ua.req.NodeTemplate.PreStartUserScript
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
		if ua.req.NodeTemplate.Runtime != nil {
			group.NodeTemplate.Runtime = ua.req.NodeTemplate.Runtime
		}
		if ua.req.NodeTemplate.Module != nil {
			group.NodeTemplate.Module = ua.req.NodeTemplate.Module
		}
		if ua.req.NodeTemplate.Creator != "" {
			group.NodeTemplate.Creator = ua.req.NodeTemplate.Creator
		}
		if ua.req.NodeTemplate.Updater != "" {
			group.NodeTemplate.Updater = ua.req.NodeTemplate.Updater
		}
		if ua.req.NodeTemplate.ScaleInPreScript != "" {
			group.NodeTemplate.ScaleInPreScript = iutils.Base64Encode(ua.req.NodeTemplate.ScaleInPreScript)
		} else {
			group.NodeTemplate.ScaleInPreScript = ua.req.NodeTemplate.ScaleInPreScript
		}
		if ua.req.NodeTemplate.ScaleInPostScript != "" {
			group.NodeTemplate.ScaleInPostScript = iutils.Base64Encode(ua.req.NodeTemplate.ScaleInPostScript)
		} else {
			group.NodeTemplate.ScaleInPostScript = ua.req.NodeTemplate.ScaleInPostScript
		}

		// attention: field will be full update
		group.NodeTemplate.UnSchedulable = ua.req.NodeTemplate.UnSchedulable
		group.NodeTemplate.Taints = ua.req.NodeTemplate.Taints
		group.NodeTemplate.Labels = ua.req.NodeTemplate.Labels
		group.NodeTemplate.Annotations = ua.req.NodeTemplate.Annotations
		// avoid update mixed deployment switch
		// group.NodeTemplate.SkipSystemInit = ua.req.NodeTemplate.SkipSystemInit
		group.NodeTemplate.AllowSkipScaleOutWhenFailed = ua.req.NodeTemplate.AllowSkipScaleOutWhenFailed
		group.NodeTemplate.AllowSkipScaleInWhenFailed = ua.req.NodeTemplate.AllowSkipScaleInWhenFailed
	}
}

// getRelativeResource relative resource
func (ua *UpdateAction) getRelativeResource() error {
	group, err := ua.model.GetNodeGroup(ua.ctx, ua.req.NodeGroupID)
	if err != nil {
		blog.Errorf("get NodeGroup %s failed, %s", ua.req.NodeGroupID, err.Error())
		return err
	}
	ua.group = group

	// cluster
	cluster, err := ua.model.GetCluster(ua.ctx, ua.req.ClusterID)
	if err != nil {
		blog.Errorf("can not get relative Cluster %s when update NodeGroup", ua.req.ClusterID)
		return fmt.Errorf("get relative cluster %s info err, %s", ua.req.ClusterID, err.Error())
	}
	ua.cluster = cluster

	// cloud
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

// updateCloudNodeGroup update cloud nodeGroup
func (ua *UpdateAction) updateCloudNodeGroup() error {
	// get credential for cloudprovider operation
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
	// create nodegroup with cloudprovider
	mgr, err := cloudprovider.GetNodeGroupMgr(ua.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get NodeGroup Manager cloudprovider %s/%s for update nodegroup %s in cluster %s failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, ua.group.NodeGroupID, ua.group.ClusterID, err.Error(),
		)
		return err
	}
	cmOption.Region = ua.group.Region

	// update cloud nodeGroup implement or generate task running
	task, err := mgr.UpdateNodeGroup(ua.group, &cloudprovider.UpdateNodeGroupOption{
		CommonOption: *cmOption,
		Cloud:        ua.cloud,
		Cluster:      ua.cluster,
		OnlyData:     ua.req.OnlyUpdateInfo,
	})
	if err != nil {
		blog.Errorf("update nodegroup %s in cluster %s with cloudprovider %s failed, %s",
			ua.group.NodeGroupID, ua.group.ClusterID, ua.cloud.CloudProvider, err.Error(),
		)
		return err
	}

	// create task and dispatch task
	if task != nil {
		ua.task = task
		// create task and dispatch task
		if err = ua.model.CreateTask(ua.ctx, task); err != nil {
			blog.Errorf("update nodegroup task for nodegroup %s failed, %s",
				ua.group.NodeGroupID, err.Error(),
			)
			return err
		}
		if err = taskserver.GetTaskServer().Dispatch(task); err != nil {
			blog.Errorf("dispatch update nodegroup task for nodegroup %s failed, %s",
				ua.group.NodeGroupID, err.Error(),
			)
			return err
		}
		// update group info
		if err = ua.saveNodeGroupStatus(common.StatusNodeGroupUpdating); err != nil {
			return err
		}

		return nil
	}

	if err = ua.saveNodeGroupStatus(common.StatusRunning); err != nil {
		return err
	}

	ua.resp.Data = removeSensitiveInfo(ua.group)
	return nil
}

// saveNodeGroupStatus save group status
func (ua *UpdateAction) saveNodeGroupStatus(status string) error {
	ua.group.Status = status
	if err := ua.model.UpdateNodeGroup(ua.ctx, ua.group); err != nil {
		blog.Errorf("nodegroup %s update in cloudprovider %s success, but update failed in local storage, %s. detail: %+v",
			ua.group.NodeGroupID, ua.group.Provider, err.Error(), iutils.ToJSONString(ua.group),
		)
		return err
	}
	blog.Infof("update nodegroup %s successfully", ua.group.NodeGroupID)
	return nil
}

// checkStatus check status
func (ua *UpdateAction) checkStatus() error {
	// if nodegroup is creating/deleting/deleted, return error
	if ua.group.Status == common.StatusCreateNodeGroupCreating ||
		ua.group.Status == common.StatusDeleteNodeGroupDeleting ||
		ua.group.Status == common.StatusCreateNodeGroupFailed ||
		ua.group.Status == common.StatusDeleteNodeGroupFailed {
		err := fmt.Errorf("nodegroup %s status is not running, can not update nodegroup", ua.group.NodeGroupID)
		return err
	}
	return nil
}

func (ua *UpdateAction) checkAdjustGroupQuota() error {
	if ua.req.GetAutoScaling().GetMaxSize() <= ua.group.GetAutoScaling().GetMaxSize() {
		return nil
	}

	scaleUpNum := ua.req.GetAutoScaling().GetMaxSize() - ua.group.GetAutoScaling().GetMaxSize()
	// check resource pool quota
	err := checkNodeGroupResourceValidate(ua.cloud.GetCloudProvider(), ua.group, common.OperationUpdate, scaleUpNum)
	if err != nil {
		return err
	}

	return nil
}

// Handle handle update cluster nodeGroup
func (ua *UpdateAction) Handle(
	ctx context.Context, req *cmproto.UpdateNodeGroupRequest, resp *cmproto.UpdateNodeGroupResponse) {

	if req == nil || resp == nil {
		blog.Errorf("update cloud nodeGroup failed, req or resp is empty")
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

	// check nodeGroup resize max quota
	if err := ua.checkAdjustGroupQuota(); err != nil {
		ua.setResp(common.BcsErrClusterManagerCheckCloudResourceQuotaErr, err.Error())
		return
	}

	// sync update cloud nodeGroup
	ua.modifyNodeGroupField()
	if err := ua.updateCloudNodeGroup(); err != nil {
		_ = ua.saveNodeGroupStatus(common.StatusNodeGroupUpdateFailed)
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	if err := ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   ua.req.NodeGroupID,
		TaskID: func() string {
			if ua.task == nil {
				return ""
			}
			return ua.task.TaskID
		}(),
		Message:      fmt.Sprintf("集群%s节点池%s更新配置信息", ua.req.ClusterID, ua.req.NodeGroupID),
		OpUser:       req.Updater,
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    ua.cluster.ClusterID,
		ProjectID:    ua.cluster.ProjectID,
		ResourceName: ua.group.GetName(),
	}); err != nil {
		blog.Errorf("UpdateNodeGroup[%s] CreateOperationLog failed: %v", ua.req.NodeGroupID, err)
	}

	ua.resp.Data = ua.group
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
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

// setResp resp body
func (ua *MoveNodeAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// validate check
func (ua *MoveNodeAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return err
	}
	// get nodegroup for validation
	destGroup, err := ua.model.GetNodeGroup(ua.ctx, ua.req.NodeGroupID)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("Get NodeGroup %s in pre-MoveNode checking failed, err %s", ua.req.NodeGroupID, err.Error())
		return err
	}
	// check cluster info consistency
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
	// get specified node for move validation
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
			err = fmt.Errorf("move node %s is not under cluster %s", ip, ua.group.ClusterID)
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
		// valiate already setting response message
		return
	}

	// moveCloudNodeGroupNodes move node to nodeGroup
	if err := ua.moveCloudNodeGroupNodes(); err != nil {
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	// try to update Node
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
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    ua.group.ClusterID,
		ProjectID:    ua.group.ProjectID,
		ResourceName: ua.group.GetName(),
	})
	if err != nil {
		blog.Errorf("MoveNodesToGroup[%s] CreateOperationLog failed: %v", req.NodeGroupID, err)
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

func (ua *MoveNodeAction) moveCloudNodeGroupNodes() error {
	// try to move node in cloudprovider
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

	// build moveNodesToGroup task
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
	if task != nil {
		ua.resp.Data = task
		if err = ua.model.CreateTask(ua.ctx, task); err != nil {
			blog.Errorf("save move nodes to node group task for cluster %s failed, %s",
				ua.group.ClusterID, err.Error(),
			)
			ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
			return err
		}
		if err = taskserver.GetTaskServer().Dispatch(task); err != nil {
			blog.Errorf("dispatch move nodes to node group task for cluster %s failed, %s",
				ua.group.ClusterID, err.Error(),
			)
			ua.setResp(common.BcsErrClusterManagerTaskErr, err.Error())
			return err
		}
	}

	blog.Infof("Nodes %v move to NodeGroup %s in cloudprovider %s/%s successfully",
		ua.req.Nodes, ua.group.NodeGroupID, cloud.CloudID, cloud.CloudProvider)
	return nil
}

// UpdateDesiredNodeAction update action for desired nodes
type UpdateDesiredNodeAction struct {
	ctx    context.Context
	model  store.ClusterManagerModel
	req    *cmproto.UpdateGroupDesiredNodeRequest
	resp   *cmproto.UpdateGroupDesiredNodeResponse
	locker lock.DistributedLock

	group        *cmproto.NodeGroup
	cluster      *cmproto.Cluster
	cloud        *cmproto.Cloud
	asOption     *cmproto.ClusterAutoScalingOption
	commonOption *cloudprovider.CommonOption

	// 兼容clusterManager和nodeManager
	clusterCloud *cmproto.Cloud
	task         *cmproto.Task

	nodeScheduler bool
}

// NewUpdateDesiredNodeAction create update action for online cluster credential
func NewUpdateDesiredNodeAction(model store.ClusterManagerModel, lock lock.DistributedLock) *UpdateDesiredNodeAction {
	return &UpdateDesiredNodeAction{
		model:  model,
		locker: lock,
	}
}

// setResp resp body
func (ua *UpdateDesiredNodeAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// validate check
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

// handleTask task handle
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
		CommonOption: *ua.commonOption,
		Cloud:        ua.cloud,
		Cluster:      ua.cluster,
		NodeGroup:    ua.group,
		AsOption:     ua.asOption,
		Operator:     ua.req.Operator,
		Manual:       ua.req.Manual,
		NodeSchedule: ua.nodeScheduler,
	})
	if err != nil {
		blog.Errorf("build scaling task for NodeGroup %s with cloudprovider %s failed, %s",
			ua.group.NodeGroupID, ua.cloud.CloudProvider, err.Error(),
		)
		return err
	}
	if err = ua.model.CreateTask(ua.ctx, task); err != nil {
		blog.Errorf("save scaling task for NodeGroup %s failed, %s",
			ua.group.NodeGroupID, err.Error(),
		)
		return err
	}
	if err = taskserver.GetTaskServer().Dispatch(task); err != nil {
		blog.Errorf("dispatch scaling task for NodeGroup %s failed, %s",
			ua.group.NodeGroupID, err.Error(),
		)
		return err
	}

	utils.HandleTaskStepData(ua.ctx, task)

	ua.task = task
	ua.resp.Data = task
	blog.Infof("scaling %d node, %v desired node task for NodeGroup successfully for %s", scaling,
		ua.req.DesiredNode, ua.group.NodeGroupID)
	return nil
}

// getRelativeData cloud/cluster/asOption
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

	clusterCloud, err := actions.GetCloudByCloudID(ua.model, ua.cluster.Provider)
	if err != nil {
		blog.Errorf("get clusterCloud %s when updateDesiredNode %d in NodeGroup %s failed, %s",
			ua.cluster.Provider, ua.req.DesiredNode, ua.group.NodeGroupID, err.Error(),
		)
		return err
	}
	ua.clusterCloud = clusterCloud

	ua.asOption, _ = actions.GetAsOptionByClusterID(ua.model, ua.group.ClusterID)

	nodeScheduler, err := utils.CheckClusterNodeNum(ua.model, ua.cluster)
	if err != nil {
		blog.Errorf("CheckClusterNodeNum[%s:%s] failed: %v", ua.cluster.ClusterID, ua.group.NodeGroupID, err)
	}
	ua.nodeScheduler = nodeScheduler

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
	cmOption.Region = ua.group.Region
	ua.commonOption = cmOption

	mgr, err := cloudprovider.GetNodeGroupMgr(ua.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s NodeGroupMgr when updateDesiredNode %d in NodeGroup %s failed, %s",
			ua.group.Provider, ua.req.DesiredNode, ua.group.NodeGroupID, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return 0, err
	}
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

// updateNodeGroupDesiredSize update desired size
func (ua *UpdateDesiredNodeAction) updateNodeGroupDesiredSize(desiredNode uint32) error {
	ua.group.AutoScaling.DesiredSize = desiredNode
	ua.group.Updater = ua.req.Operator

	// update DesiredSize in local storage
	if err := ua.model.UpdateNodeGroup(ua.ctx, ua.group); err != nil {
		blog.Errorf("updateDesiredNode %d to NodeGroup %s in local storage failed, %s",
			ua.req.DesiredNode, ua.req.NodeGroupID, err.Error(),
		)
		return err
	}

	return nil
}

// checkCloudClusterResource check cloud cluster resource
func (ua *UpdateDesiredNodeAction) checkCloudClusterResource(scaleNodesNum uint32) error {
	// check cluster common resource
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ua.clusterCloud,
		AccountID: ua.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential from cloud %s when checkCloudClusterResource %d in NodeGroup %s failed, %s",
			ua.group.Provider, ua.req.DesiredNode, ua.group.NodeGroupID, err.Error(),
		)
		return err
	}
	cmOption.Region = ua.cluster.Region

	// get cloudprovider cluster implementation
	clusterMgr, err := cloudprovider.GetClusterMgr(ua.clusterCloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s ClusterManager for Cluster %s failed, %s",
			ua.clusterCloud.CloudProvider, ua.cluster.ClusterID, err.Error(),
		)
		return err
	}

	// check cloud CIDR && autoScale cluster cidr
	available, err := clusterMgr.CheckClusterCidrAvailable(ua.cluster, &cloudprovider.CheckClusterCIDROption{
		CommonOption:    *cmOption,
		IncomingNodeCnt: uint64(scaleNodesNum),
		ExternalNode: func() bool {
			return ua.group.GetNodeGroupType() == common.External.String()
		}(),
	})
	if !available {
		blog.Infof("checkCloudClusterResource failed: %v", err)
		return err
	}

	return nil
}

// Handle handle update cluster nodeGroup desired node
func (ua *UpdateDesiredNodeAction) Handle(
	ctx context.Context, req *cmproto.UpdateGroupDesiredNodeRequest, resp *cmproto.UpdateGroupDesiredNodeResponse) {

	if req == nil || resp == nil {
		blog.Errorf("update NodeGroup failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	const (
		updateClusterDesiredNodeLockKey = "/bcs-services/bcs-cluster-manager/UpdateDesiredNodeAction"
	)
	ua.locker.Lock(updateClusterDesiredNodeLockKey, []lock.LockOption{lock.LockTTL(time.Second * 5)}...) // nolint
	defer ua.locker.Unlock(updateClusterDesiredNodeLockKey)                                              // nolint

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
		blog.Errorf("updateDesiredNode to %d for NodeGroup %s with cloudprovider %s failed, %s",
			req.DesiredNode, ua.group.NodeGroupID, ua.cloud.CloudProvider, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	// check cluster cloud resource
	err = ua.checkCloudClusterResource(scaleResp)
	if err != nil {
		blog.Errorf("updateDesiredNode to %d for NodeGroup %s checkCloudClusterResource failed, %s",
			req.DesiredNode, ua.group.NodeGroupID, err.Error(),
		)
		ua.setResp(common.BcsErrClusterManagerCheckCloudClusterResourceErr, err.Error())
		return
	}

	// update nodeGroup size
	err = ua.updateNodeGroupDesiredSize(req.DesiredNode)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// handler cloud update desired node
	if err = ua.handleTask(scaleResp); err != nil {
		ua.setResp(common.BcsErrClusterManagerTaskErr, err.Error())
		return
	}

	// inject virtual nodes
	if ua.req.Manual {
		ua.injectVirtualNodeData(scaleResp)
	}

	// record operation log
	err = ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   req.NodeGroupID,
		TaskID:       ua.task.TaskID,
		Message:      fmt.Sprintf("集群%s扩容节点池%s节点数至%v", ua.cluster.ClusterID, req.NodeGroupID, req.DesiredNode),
		OpUser:       ua.group.Updater,
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    ua.cluster.ClusterID,
		ProjectID:    ua.cluster.ProjectID,
		ResourceName: ua.group.GetName(),
	})
	if err != nil {
		blog.Errorf("UpdateDesiredNode[%s] CreateOperationLog failed: %v", req.NodeGroupID, err)
	}

	blog.Infof("updateDesiredNode %d to NodeGroup %s successfully", req.DesiredNode, req.NodeGroupID)
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

func (ua *UpdateDesiredNodeAction) injectVirtualNodeData(nodeNum uint32) {
	var i uint32

	for ; i < nodeNum; i++ {
		nodeId := virtualNodeID()

		err := ua.model.CreateNode(context.Background(), &cmproto.Node{
			NodeID:      nodeId,
			Status:      common.StatusResourceApplying,
			ZoneID:      "",
			NodeGroupID: ua.group.NodeGroupID,
			ClusterID:   ua.group.ClusterID,
			VPC:         ua.cluster.VpcID,
			Region:      ua.cluster.Region,
			TaskID:      ua.task.GetTaskID(),
		})
		if err != nil {
			blog.Errorf("UpdateDesiredNodeAction injectVirtualNodeData[%s] failed: %v", nodeId, err)
		}
		time.Sleep(time.Millisecond * 10)
	}
	blog.Infof("UpdateDesiredNodeAction injectVirtualNodeData success")
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

// setResp resp body
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

	if err = ua.model.UpdateNodeGroup(ctx, destGroup); err != nil {
		blog.Errorf("nodegroup %s update desiredSize[%s] failed in local storage, %s",
			destGroup.NodeGroupID, destGroup, err.Error())
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err = ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   req.NodeGroupID,
		TaskID:       "",
		Message:      fmt.Sprintf("更新集群%s节点池%s期望扩容节点数至%d", destGroup.ClusterID, req.NodeGroupID, req.DesiredSize),
		OpUser:       req.Operator,
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    destGroup.ClusterID,
		ProjectID:    destGroup.ProjectID,
		ResourceName: destGroup.GetName(),
	})
	if err != nil {
		blog.Errorf("UpdateGroupDesiredSize[%s] CreateOperationLog failed: %v", req.NodeGroupID, err)
	}

	blog.Infof("update nodegroup desiredSize %s successfully", destGroup.NodeGroupID)
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// UpdateGroupMinMaxAction update nodegroup autoscaling min/max size
type UpdateGroupMinMaxAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.UpdateGroupMinMaxSizeRequest
	resp  *cmproto.UpdateGroupMinMaxSizeResponse
}

// NewUpdateGroupMinMaxAction create update action for group
func NewUpdateGroupMinMaxAction(model store.ClusterManagerModel) *UpdateGroupMinMaxAction {
	return &UpdateGroupMinMaxAction{
		model: model,
	}
}

func (ua *UpdateGroupMinMaxAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// validate check
func (ua *UpdateGroupMinMaxAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	if ua.req.MinSize > ua.req.MaxSize {
		return fmt.Errorf("UpdateGroupMinMaxAction minSize > maxSize")
	}

	return nil
}

// Handle handle update cluster credential
func (ua *UpdateGroupMinMaxAction) Handle(
	ctx context.Context, req *cmproto.UpdateGroupMinMaxSizeRequest, resp *cmproto.UpdateGroupMinMaxSizeResponse) {

	if req == nil || resp == nil {
		blog.Errorf("update nodegroup min/max failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
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
	destGroup.AutoScaling.MinSize = req.MinSize
	destGroup.AutoScaling.MaxSize = req.MaxSize

	if err = ua.model.UpdateNodeGroup(ctx, destGroup); err != nil {
		blog.Errorf("nodegroup %s update min/maxSize failed in local storage, %s",
			destGroup.NodeGroupID, err.Error())
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err = ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   req.NodeGroupID,
		TaskID:       "",
		Message: fmt.Sprintf("集群[%s]修改NodeGroup[%s]最小最大的扩容限额[min%d][max%d]", destGroup.ClusterID,
			req.NodeGroupID, req.MinSize, req.MaxSize),
		OpUser:       req.Operator,
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    destGroup.ClusterID,
		ProjectID:    destGroup.ProjectID,
		ResourceName: destGroup.GetName(),
	})
	if err != nil {
		blog.Errorf("UpdateGroupMinMaxSize[%s] CreateOperationLog failed: %v", req.NodeGroupID, err)
	}

	blog.Infof("update nodegroup min/maxSize %s successfully", destGroup.NodeGroupID)
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// UpdateGroupAsTimeRangeAction update nodegroup autoscaling time range strategy
type UpdateGroupAsTimeRangeAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.UpdateGroupAsTimeRangeRequest
	resp  *cmproto.UpdateGroupAsTimeRangeResponse

	group *cmproto.NodeGroup
}

// NewUpdateGroupAsTimeRangeAction create update autoscaling time strategy action for group
func NewUpdateGroupAsTimeRangeAction(model store.ClusterManagerModel) *UpdateGroupAsTimeRangeAction {
	return &UpdateGroupAsTimeRangeAction{
		model: model,
	}
}

func (ua *UpdateGroupAsTimeRangeAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// validate check
func (ua *UpdateGroupAsTimeRangeAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	maxSize := ua.group.GetAutoScaling().GetMaxSize()
	minSize := ua.group.GetAutoScaling().GetMinSize()

	if len(ua.req.GetTimeRanges()) == 0 {
		return fmt.Errorf("UpdateGroupAsTimeRangeAction[%s] timeRanges empty", ua.req.NodeGroupID)
	}

	for i := range ua.req.GetTimeRanges() {
		if ua.req.GetTimeRanges()[i].GetName() == "" {
			return fmt.Errorf("UpdateGroupAsTimeRangeAction[%s] timeRanges name empty", ua.req.NodeGroupID)
		}

		if ua.req.GetTimeRanges()[i].GetDesiredNum() < minSize || ua.req.GetTimeRanges()[i].GetDesiredNum() > maxSize {
			return fmt.Errorf("UpdateGroupAsTimeRangeAction[%s] timeRanges[%s] desiredNum in [%v-%v]",
				ua.req.NodeGroupID, ua.req.GetTimeRanges()[i].GetName(), minSize, maxSize)
		}
		// default Asia/Shanghai
		if ua.req.GetTimeRanges()[i].GetZone() == "" {
			ua.req.GetTimeRanges()[i].Zone = iutils.DefaultTimeZone
		}

		if iutils.ValidateCronExpr(ua.req.GetTimeRanges()[i].GetSchedule()) != nil {
			return fmt.Errorf("UpdateGroupAsTimeRangeAction[%s] timeRanges[%s] schedule invalid",
				ua.req.NodeGroupID, ua.req.GetTimeRanges()[i].GetName())
		}
	}

	return nil
}

// Handle handle update cluster credential
func (ua *UpdateGroupAsTimeRangeAction) Handle(
	ctx context.Context, req *cmproto.UpdateGroupAsTimeRangeRequest, resp *cmproto.UpdateGroupAsTimeRangeResponse) {

	if req == nil || resp == nil {
		blog.Errorf("update nodegroup min/max failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	// get old project information, update fields if required
	destGroup, err := ua.model.GetNodeGroup(ua.ctx, req.NodeGroupID)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("find nodegroup %s failed when pre-update checking, err %s", req.NodeGroupID, err.Error())
		return
	}
	ua.group = destGroup

	if err = ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	destGroup.AutoScaling.TimeRanges = ua.req.GetTimeRanges()

	if err = ua.model.UpdateNodeGroup(ctx, destGroup); err != nil {
		blog.Errorf("nodegroup %s update autoscaling timeRanges failed in local storage, %s",
			destGroup.NodeGroupID, err.Error())
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err = ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   req.NodeGroupID,
		TaskID:       "",
		Message:      fmt.Sprintf("集群[%s]更新节点池[%s]定时扩缩容策略", destGroup.ClusterID, req.NodeGroupID),
		OpUser:       auth.GetUserFromCtx(ctx),
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    destGroup.ClusterID,
		ProjectID:    destGroup.ProjectID,
		ResourceName: destGroup.GetName(),
	})
	if err != nil {
		blog.Errorf("UpdateGroupAsTimeRange[%s] CreateOperationLog failed: %v", req.NodeGroupID, err)
	}

	blog.Infof("update nodegroup autoscaling time ranges %s successfully", destGroup.NodeGroupID)
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
