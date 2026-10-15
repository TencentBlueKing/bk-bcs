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

// Package nodegroup xxx
package nodegroup

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/taskserver"
)

// EnableNodeGroupAutoScaleAction set nodegroup autoscaling enable
type EnableNodeGroupAutoScaleAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	cluster *cmproto.Cluster
	group   *cmproto.NodeGroup
	cloud   *cmproto.Cloud
	req     *cmproto.EnableNodeGroupAutoScaleRequest
	resp    *cmproto.EnableNodeGroupAutoScaleResponse
}

// NewEnableNodeGroupAutoScaleAction create update action for update
func NewEnableNodeGroupAutoScaleAction(model store.ClusterManagerModel) *EnableNodeGroupAutoScaleAction {
	return &EnableNodeGroupAutoScaleAction{
		model: model,
	}
}

func (ua *EnableNodeGroupAutoScaleAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ua *EnableNodeGroupAutoScaleAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}
	return nil
}

func (ua *EnableNodeGroupAutoScaleAction) getRelativeResource() error {
	// get relative cluster for information injection
	group, err := ua.model.GetNodeGroup(ua.ctx, ua.req.NodeGroupID)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("find nodegroup %s failed when enable nodegroup auto scale, err %s", ua.req.NodeGroupID, err.Error())
		return err
	}
	ua.group = group

	cluster, err := ua.model.GetCluster(ua.ctx, group.ClusterID)
	if err != nil {
		blog.Errorf("can not get relative Cluster %s when enable nodegroup auto scale", group.ClusterID)
		return fmt.Errorf("get relative cluster %s info err, %s", group.ClusterID, err.Error())
	}
	ua.cluster = cluster

	cloud, err := actions.GetCloudByCloudID(ua.model, group.Provider)
	if err != nil {
		blog.Errorf("can not get relative Cloud %s when enable nodegroup auto scale for nodegroup %s, %s",
			group.Provider, ua.req.NodeGroupID, err.Error(),
		)
		return err
	}
	ua.cloud = cloud

	return nil
}

func (ua *EnableNodeGroupAutoScaleAction) handleNodeGroupVirtualNodes() error {
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"clusterid":   ua.cluster.ClusterID,
		"nodegroupid": ua.group.NodeGroupID,
	})
	nodes, err := ua.model.ListNode(ua.ctx, cond, &storeopt.ListOption{})
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		blog.Errorf("get Cluster %s Nodes failed, %s", ua.cluster.ClusterID, err.Error())
		return err
	}

	if len(nodes) == 0 {
		return nil
	}

	var (
		virtualNodeNum = 0
	)
	for i := range nodes {
		if nodes[i].Status == common.StatusResourceApplyFailed && strings.HasPrefix(nodes[i].NodeID, "bcs") {
			virtualNodeNum++
			err = ua.model.DeleteClusterNode(ua.ctx, ua.cluster.ClusterID, nodes[i].NodeID)
			if err != nil {
				blog.Errorf("EnableNodeGroupAutoScaleAction[%s:%s] DeleteClusterNode[%s] failed: %v",
					ua.cluster.ClusterID, ua.req.NodeGroupID, nodes[i].GetNodeID(), err)
			}
		}
	}

	if virtualNodeNum > 0 {
		err = actions.UpdateNodeGroupDesiredSize(ua.model, ua.req.NodeGroupID, virtualNodeNum, true)
		if err != nil {
			blog.Errorf("EnableNodeGroupAutoScaleAction[%s:%s] UpdateNodeGroupDesiredSize[%v] failed: %v",
				ua.cluster.ClusterID, ua.req.NodeGroupID, virtualNodeNum, err)
		}
	}

	return nil
}

func (ua *EnableNodeGroupAutoScaleAction) enableNodeGroupAutoScale() error {
	// check auto scale state, if enable, return success
	if ua.group.EnableAutoscale {
		blog.Infof("nodegroup %s is already enable auto scale", ua.group.Name)
		return nil
	}

	// handle virtual nodes to keep data consistency when notEnableCA nodegroup enable CA
	err := ua.handleNodeGroupVirtualNodes()
	if err != nil {
		blog.Errorf("EnableNodeGroupAutoScaleAction[%s:%s] handleNodeGroupVirtualNodes failed: %v",
			ua.cluster.ClusterID, ua.req.NodeGroupID, err)
	}

	mgr, err := cloudprovider.GetNodeGroupMgr(ua.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get NodeGroup Manager cloudprovider %s/%s for enable nodegroup auto scale in Cluster %s failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, ua.cluster.ClusterID, err.Error(),
		)
		return err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ua.cloud,
		AccountID: ua.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get Credential for Cloud %s/%s when enable nodegroup auto scale in cluster %s failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, ua.cluster.ClusterID, err.Error(),
		)
		return err
	}
	cmOption.Region = ua.group.Region

	// set nodegroup to updating
	ua.group.Status = common.StatusNodeGroupUpdating
	ua.group.EnableAutoscale = true
	if err = ua.model.UpdateNodeGroup(ua.ctx, ua.group); err != nil {
		blog.Errorf("update nodegroup %s status to updating failed, err %s", ua.group.NodeGroupID, err.Error())
		return err
	}

	// cloud provider nodeGroup
	task, err := mgr.SwitchNodeGroupAutoScaling(ua.group, true, &cloudprovider.SwitchNodeGroupAutoScalingOption{
		CommonOption: *cmOption,
		Cluster:      ua.cluster,
		Cloud:        ua.cloud,
	})
	if err != nil {
		blog.Errorf("enable nodegroup auto scale in cloudprovider %s/%s for group %s failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, ua.group.NodeGroupID, err.Error())
		return err
	}

	// create task and dispatch task
	taskID := ""
	if task != nil {
		taskID = task.TaskID
		// create task and dispatch task
		if err = ua.model.CreateTask(ua.ctx, task); err != nil {
			blog.Errorf("save enable nodegroup auto scale task for nodegroup %s failed, %s",
				ua.group.NodeGroupID, err.Error(),
			)
			return err
		}
		if err = taskserver.GetTaskServer().Dispatch(task); err != nil {
			blog.Errorf("dispatch enable nodegroup auto scale task for nodegroup %s failed, %s",
				ua.group.NodeGroupID, err.Error(),
			)
			return err
		}
	}

	err = ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   ua.group.NodeGroupID,
		TaskID:       taskID,
		Message:      fmt.Sprintf("%s 开启节点池", ua.group.NodeGroupID),
		OpUser:       ua.group.Updater,
		CreateTime:   time.Now().UTC().Format(time.RFC3339),
		ClusterID:    ua.cluster.ClusterID,
		ProjectID:    ua.cluster.ProjectID,
		ResourceName: ua.group.GetName(),
	})
	if err != nil {
		blog.Errorf("EnableNodeGroupAutoScale[%s] CreateOperationLog failed: %v", ua.group.NodeGroupID, err)
	}
	return nil
}

// Handle handle set nodegroup autoscaling enable
func (ua *EnableNodeGroupAutoScaleAction) Handle(
	ctx context.Context, req *cmproto.EnableNodeGroupAutoScaleRequest, resp *cmproto.EnableNodeGroupAutoScaleResponse) {

	if req == nil || resp == nil {
		blog.Errorf("EnableNodeGroupAutoScale failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// getRelativeResource get cluster / cloud provider
	if err := ua.getRelativeResource(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// if nodegroup is updating, return error
	if ua.group.Status == common.StatusNodeGroupUpdating {
		blog.Errorf("nodegroup %s status is updating, can not enable auto scale", ua.group.Name)
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, errors.New("nodegroup is updating").Error())
		return
	}

	if err := ua.enableNodeGroupAutoScale(); err != nil {
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// DisableNodeGroupAutoScaleAction set nodegroup autoscaling disable
type DisableNodeGroupAutoScaleAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	cluster *cmproto.Cluster
	group   *cmproto.NodeGroup
	cloud   *cmproto.Cloud
	req     *cmproto.DisableNodeGroupAutoScaleRequest
	resp    *cmproto.DisableNodeGroupAutoScaleResponse
}

// NewDisableNodeGroupAutoScaleAction create update action for update
func NewDisableNodeGroupAutoScaleAction(model store.ClusterManagerModel) *DisableNodeGroupAutoScaleAction {
	return &DisableNodeGroupAutoScaleAction{
		model: model,
	}
}

func (ua *DisableNodeGroupAutoScaleAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ua *DisableNodeGroupAutoScaleAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}
	return nil
}

func (ua *DisableNodeGroupAutoScaleAction) getRelativeResource() error {
	// get relative cluster for information injection
	group, err := ua.model.GetNodeGroup(ua.ctx, ua.req.NodeGroupID)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("find nodegroup %s failed when disable nodegroup auto scale, err %s", ua.req.NodeGroupID, err.Error())
		return err
	}
	ua.group = group

	cluster, err := ua.model.GetCluster(ua.ctx, group.ClusterID)
	if err != nil {
		blog.Errorf("can not get relative Cluster %s when enable nodegroup auto scale", group.ClusterID)
		return fmt.Errorf("get relative cluster %s info err, %s", group.ClusterID, err.Error())
	}
	ua.cluster = cluster

	cloud, err := actions.GetCloudByCloudID(ua.model, group.Provider)
	if err != nil {
		blog.Errorf("can not get relative Cloud %s when enable nodegroup auto scale for nodegroup %s, %s",
			group.Provider, ua.req.NodeGroupID, err.Error(),
		)
		return err
	}
	ua.cloud = cloud

	return nil
}

func (ua *DisableNodeGroupAutoScaleAction) disableNodeGroupAutoScale() error {
	// check auto scale state, if enable, return success
	if !ua.group.EnableAutoscale {
		blog.Infof("nodegroup %s is already disable auto scale", ua.group.Name)
		return nil
	}
	mgr, err := cloudprovider.GetNodeGroupMgr(ua.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get NodeGroup Manager cloudprovider %s/%s for enable nodegroup auto scale in Cluster %s failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, ua.cluster.ClusterID, err.Error(),
		)
		return err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ua.cloud,
		AccountID: ua.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get Credential for Cloud %s/%s when enable nodegroup auto scale in cluster %s failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, ua.cluster.ClusterID, err.Error(),
		)
		return err
	}
	cmOption.Region = ua.group.Region

	// set nodegroup to updating
	ua.group.Status = common.StatusNodeGroupUpdating
	ua.group.EnableAutoscale = false
	if err = ua.model.UpdateNodeGroup(ua.ctx, ua.group); err != nil {
		blog.Errorf("update nodegroup %s status to updating failed, err %s", ua.group.NodeGroupID, err.Error())
		return err
	}

	// cloud provider nodeGroup
	task, err := mgr.SwitchNodeGroupAutoScaling(ua.group, false,
		&cloudprovider.SwitchNodeGroupAutoScalingOption{CommonOption: *cmOption})
	if err != nil {
		blog.Errorf("disable nodegroup auto scale in cloudprovider %s/%s for group %s failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, ua.group.NodeGroupID, err.Error())
		return err
	}

	// create task and dispatch task
	taskID := ""

	if task != nil {
		taskID = task.TaskID
		// create task and dispatch task
		if err = ua.model.CreateTask(ua.ctx, task); err != nil {
			blog.Errorf("save disable nodegroup auto scale task for nodegroup %s failed, %s",
				ua.group.NodeGroupID, err.Error(),
			)
			return err
		}
		if err = taskserver.GetTaskServer().Dispatch(task); err != nil {
			blog.Errorf("dispatch disable nodegroup auto scale task for nodegroup %s failed, %s",
				ua.group.NodeGroupID, err.Error(),
			)
			return err
		}
	}

	err = ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   ua.group.NodeGroupID,
		TaskID:       taskID,
		Message:      fmt.Sprintf("%s 关闭节点池", ua.group.NodeGroupID),
		OpUser:       ua.group.Updater,
		CreateTime:   time.Now().UTC().Format(time.RFC3339),
		ClusterID:    ua.cluster.ClusterID,
		ProjectID:    ua.cluster.ProjectID,
	})
	if err != nil {
		blog.Errorf("DisableNodeGroupAutoScale[%s] CreateOperationLog failed: %v", ua.group.NodeGroupID, err)
	}
	return nil
}

// Handle handle set nodegroup autoscaling disable
func (ua *DisableNodeGroupAutoScaleAction) Handle(
	ctx context.Context, req *cmproto.DisableNodeGroupAutoScaleRequest, resp *cmproto.DisableNodeGroupAutoScaleResponse) {

	if req == nil || resp == nil {
		blog.Errorf("DisableNodeGroupAutoScale failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// getRelativeResource get cluster / cloud provider
	if err := ua.getRelativeResource(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// if nodegroup is updating, return error  ua.group.Status != common.StatusRunning
	if ua.group.Status == common.StatusNodeGroupUpdating {
		blog.Errorf("nodegroup %s status is updating, can not disable auto scale", ua.group.Name)
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, errors.New("nodegroup is updating").Error())
		return
	}

	if err := ua.disableNodeGroupAutoScale(); err != nil {
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
