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

package autoscalingoption

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/taskserver"
)

// UpdateAutoScalingStatusAction update action for auto scaling status
type UpdateAutoScalingStatusAction struct {
	ctx      context.Context
	model    store.ClusterManagerModel
	asOption *cmproto.ClusterAutoScalingOption
	cluster  *cmproto.Cluster
	cloud    *cmproto.Cloud
	req      *cmproto.UpdateAutoScalingStatusRequest
	resp     *cmproto.UpdateAutoScalingStatusResponse
}

// NewUpdateAutoScalingStatusAction create update action for auto scaling status
func NewUpdateAutoScalingStatusAction(model store.ClusterManagerModel) *UpdateAutoScalingStatusAction {
	return &UpdateAutoScalingStatusAction{
		model: model,
	}
}

func (ua *UpdateAutoScalingStatusAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ua *UpdateAutoScalingStatusAction) getRelativeResource() error {
	// get relative cluster for information injection
	asOption, err := ua.model.GetAutoScalingOption(ua.ctx, ua.req.ClusterID)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("find asOption %s failed when update autoScaling status, err %s", ua.req.ClusterID, err.Error())
		return err
	}
	ua.asOption = asOption

	cluster, err := ua.model.GetCluster(ua.ctx, ua.req.ClusterID)
	if err != nil {
		blog.Errorf("can not get relative Cluster %s when update autoScaling status", ua.req.ClusterID)
		return fmt.Errorf("get relative cluster %s info err, %s", ua.req.ClusterID, err.Error())
	}
	ua.cluster = cluster

	// get cloud provider
	provider := cluster.Provider
	if ua.req.Provider != "" {
		provider = ua.req.Provider
	}
	cloud, err := actions.GetCloudByCloudID(ua.model, provider)
	if err != nil {
		blog.Errorf("can not get relative Cloud %s when update autoScaling status for %s, %s",
			provider, ua.req.ClusterID, err.Error())
		return err
	}
	ua.cloud = cloud

	return nil
}

func (ua *UpdateAutoScalingStatusAction) updateAutoScalingStatus() error {
	if ua.req.Enable == ua.asOption.EnableAutoscale {
		blog.Infof("skip update autoScaling status in cluster %s, cause of same status", ua.req.ClusterID)
		return nil
	}
	mgr, err := cloudprovider.GetNodeGroupMgr(ua.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get NodeGroup Manager cloudprovider %s/%s for update autoScaling status in Cluster %s failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, ua.cluster.ClusterID, err.Error(),
		)
		return err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ua.cloud,
		AccountID: ua.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get Credential for Cloud %s/%s when update autoScaling status in cluster %s failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, ua.cluster.ClusterID, err.Error(),
		)
		return err
	}
	cmOption.Region = ua.cluster.Region

	// set autoScalingOption to updating
	ua.asOption.Status = common.StatusAutoScalingOptionUpdating
	ua.asOption.EnableAutoscale = ua.req.Enable
	ua.asOption.UpdateTime = time.Now().Format(time.RFC3339)
	ua.asOption.Updater = ua.req.Updater
	if err = ua.model.UpdateAutoScalingOption(ua.ctx, ua.asOption); err != nil {
		blog.Errorf("update asOption %s status to updating status failed, err %s", ua.req.ClusterID, err.Error())
		return err
	}

	// cloud provider
	task, err := mgr.SwitchAutoScalingOptionStatus(ua.asOption, ua.req.Enable, cmOption)
	if err != nil {
		blog.Errorf("update autoScaling status in cloudprovider %s/%s for cluster %s failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, ua.req.ClusterID, err.Error())
		return err
	}

	taskID := ""
	if task != nil {
		taskID = task.TaskID
		// create task and dispatch task
		if err = ua.model.CreateTask(ua.ctx, task); err != nil {
			blog.Errorf("save update autoScaling status task for cluster %s failed, %s",
				ua.req.ClusterID, err.Error(),
			)
			return err
		}
		if err = taskserver.GetTaskServer().Dispatch(task); err != nil {
			blog.Errorf("dispatch update autoScaling status task for cluster %s failed, %s",
				ua.req.ClusterID, err.Error(),
			)
			return err
		}
	}

	err = ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.AutoScalingOption.String(),
		ResourceID:   ua.req.ClusterID,
		TaskID:       taskID,
		Message:      fmt.Sprintf("修改集群[%s]扩缩容开启状态为 %v", ua.req.ClusterID, ua.req.Enable),
		OpUser:       ua.req.Updater,
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    ua.req.ClusterID,
		ProjectID:    ua.cluster.ProjectID,
		ResourceName: ua.cluster.ClusterName,
	})
	if err != nil {
		blog.Errorf("UpdateAutoScalingStatus[%s] CreateOperationLog failed: %v", ua.req.ClusterID, err)
	}
	return nil
}

func (ua *UpdateAutoScalingStatusAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	return nil
}

// Handle handle update auto scaling status
func (ua *UpdateAutoScalingStatusAction) Handle(
	ctx context.Context, req *cmproto.UpdateAutoScalingStatusRequest, resp *cmproto.UpdateAutoScalingStatusResponse) {

	if req == nil || resp == nil {
		blog.Errorf("UpdateAutoScalingStatus failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	// getRelativeResource get autoScalingOption / cloud provider
	if err := ua.getRelativeResource(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ua.updateAutoScalingStatus(); err != nil {
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
