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

// UpdateAction update action for online cluster credential
type UpdateAction struct {
	ctx      context.Context
	model    store.ClusterManagerModel
	asOption *cmproto.ClusterAutoScalingOption
	cluster  *cmproto.Cluster
	cloud    *cmproto.Cloud
	req      *cmproto.UpdateAutoScalingOptionRequest
	resp     *cmproto.UpdateAutoScalingOptionResponse
}

// NewUpdateAction create update action for online cluster credential
func NewUpdateAction(model store.ClusterManagerModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

func (ua *UpdateAction) getRelativeResource() error {
	// get relative cluster for information injection
	asOption, err := ua.model.GetAutoScalingOption(ua.ctx, ua.req.ClusterID)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("find asOption %s failed when update autoScaling, err %s", ua.req.ClusterID, err.Error())
		return err
	}
	ua.asOption = asOption

	cluster, err := ua.model.GetCluster(ua.ctx, ua.req.ClusterID)
	if err != nil {
		blog.Errorf("can not get relative Cluster %s when update autoScaling", ua.req.ClusterID)
		return fmt.Errorf("get relative cluster %s info err, %s", ua.req.ClusterID, err.Error())
	}
	ua.cluster = cluster

	cloud, err := actions.GetCloudByCloudID(ua.model, cluster.Provider)
	if err != nil {
		blog.Errorf("can not get relative Cloud %s when update autoScaling for %s, %s",
			cluster.Provider, ua.req.ClusterID, err.Error(),
		)
		return err
	}
	ua.cloud = cloud

	return nil
}

func (ua *UpdateAction) updateAutoScaling() error {
	mgr, err := cloudprovider.GetNodeGroupMgr(ua.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get NodeGroup Manager cloudprovider %s/%s for update autoScaling in Cluster %s failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, ua.cluster.ClusterID, err.Error(),
		)
		return err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ua.cloud,
		AccountID: ua.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get Credential for Cloud %s/%s when update autoScaling in cluster %s failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, ua.cluster.ClusterID, err.Error(),
		)
		return err
	}
	cmOption.Region = ua.cluster.Region

	if err := ua.saveAutoScalingOption(ua.asOption); err != nil {
		blog.Errorf("update asOption %s failed, err %s", ua.req.ClusterID, err.Error())
		return err
	}

	// cloud provider
	task, err := mgr.UpdateAutoScalingOption(ua.asOption, &cloudprovider.UpdateScalingOption{CommonOption: *cmOption})
	if err != nil {
		blog.Errorf("update autoScaling in cloudprovider %s/%s for cluster %s failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, ua.req.ClusterID, err.Error())
		return err
	}

	taskID := ""
	if task != nil {
		taskID = task.TaskID
		// create task and dispatch task
		if err = ua.model.CreateTask(ua.ctx, task); err != nil {
			blog.Errorf("save update autoScaling task for cluster %s failed, %s",
				ua.req.ClusterID, err.Error(),
			)
			return err
		}
		if err = taskserver.GetTaskServer().Dispatch(task); err != nil {
			blog.Errorf("dispatch update autoScaling task for cluster %s failed, %s",
				ua.req.ClusterID, err.Error(),
			)
			return err
		}
	}

	err = ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.AutoScalingOption.String(),
		ResourceID:   ua.req.ClusterID,
		TaskID:       taskID,
		Message:      fmt.Sprintf("编辑集群[%s]扩缩容配置", ua.req.ClusterID),
		OpUser:       ua.req.Updater,
		CreateTime:   time.Now().Format(time.RFC3339),
	})
	if err != nil {
		blog.Errorf("UpdateAutoScalingOption[%s] CreateOperationLog failed: %v", ua.req.ClusterID, err)
	}
	return nil
}

func (ua *UpdateAction) saveAutoScalingOption(option *cmproto.ClusterAutoScalingOption) error {
	timeStr := time.Now().Format(time.RFC3339)
	// update field if required
	option.Status = common.StatusAutoScalingOptionUpdating
	option.UpdateTime = timeStr
	option.Updater = ua.req.Updater
	option.IsScaleDownEnable = ua.req.IsScaleDownEnable
	if len(option.Expander) != 0 {
		option.Expander = ua.req.Expander
	}
	option.MaxEmptyBulkDelete = ua.req.MaxEmptyBulkDelete
	option.ScaleDownDelay = ua.req.ScaleDownDelay
	option.ScaleDownUnneededTime = ua.req.ScaleDownUnneededTime
	option.ScaleDownUtilizationThreahold = ua.req.ScaleDownUtilizationThreahold
	option.ScaleDownGpuUtilizationThreshold = ua.req.ScaleDownGpuUtilizationThreshold
	option.OkTotalUnreadyCount = ua.req.OkTotalUnreadyCount
	option.MaxTotalUnreadyPercentage = ua.req.MaxTotalUnreadyPercentage
	option.ScaleDownUnreadyTime = ua.req.ScaleDownUnreadyTime
	option.BufferResourceRatio = ua.req.BufferResourceRatio
	option.MaxGracefulTerminationSec = ua.req.MaxGracefulTerminationSec
	option.ScanInterval = ua.req.ScanInterval
	option.MaxNodeProvisionTime = ua.req.MaxNodeProvisionTime
	option.ScaleUpFromZero = ua.req.ScaleUpFromZero
	option.ScaleDownDelayAfterAdd = ua.req.ScaleDownDelayAfterAdd
	option.ScaleDownDelayAfterDelete = ua.req.ScaleDownDelayAfterDelete
	option.ScaleDownDelayAfterFailure = ua.req.ScaleDownDelayAfterFailure

	ua.asOption = option
	ua.resp.Data = option
	return ua.model.UpdateAutoScalingOption(ua.ctx, option)
}

func (ua *UpdateAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handle update cluster credential
func (ua *UpdateAction) Handle(
	ctx context.Context, req *cmproto.UpdateAutoScalingOptionRequest, resp *cmproto.UpdateAutoScalingOptionResponse) {

	if req == nil || resp == nil {
		blog.Errorf("update ClusterAutoScalingOption failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := req.Validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// getRelativeResource get autoScalingOption / cloud provider
	if err := ua.getRelativeResource(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if err := ua.updateAutoScaling(); err != nil {
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
