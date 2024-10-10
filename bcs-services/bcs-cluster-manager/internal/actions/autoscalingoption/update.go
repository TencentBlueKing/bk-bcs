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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
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

	// default cluster and autoscaler use same cloudProvider, but allow user to use different cloudProvider
	// inter: use different cloudProvider to provider autoscaler
	provider := cluster.Provider
	if ua.req.Provider != "" {
		provider = ua.req.Provider
	}
	cloud, err := actions.GetCloudByCloudID(ua.model, provider)
	if err != nil {
		blog.Errorf("can not get relative Cloud %s when update autoScaling for %s, %s",
			provider, ua.req.ClusterID, err.Error(),
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

	if err = ua.saveAutoScalingOption(ua.asOption); err != nil {
		blog.Errorf("update asOption %s failed, err %s", ua.req.ClusterID, err.Error())
		return err
	}

	// only update autoScalingOption info
	if ua.req.OnlyUpdateInfo {
		return nil
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
		ClusterID:    ua.req.ClusterID,
		ProjectID:    ua.cluster.ProjectID,
		ResourceName: ua.cluster.ClusterName,
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
	if ua.req.OnlyUpdateInfo {
		option.Status = common.StatusAutoScalingOptionNormal
	}
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

	if ua.req.SkipNodesWithLocalStorage != nil {
		option.SkipNodesWithLocalStorage = ua.req.SkipNodesWithLocalStorage.GetValue()
	}
	if ua.req.SkipNodesWithSystemPods != nil {
		option.SkipNodesWithSystemPods = ua.req.SkipNodesWithSystemPods.GetValue()
	}
	if ua.req.IgnoreDaemonSetsUtilization != nil {
		option.IgnoreDaemonSetsUtilization = ua.req.IgnoreDaemonSetsUtilization.GetValue()
	}

	option.OkTotalUnreadyCount = ua.req.OkTotalUnreadyCount
	option.MaxTotalUnreadyPercentage = ua.req.MaxTotalUnreadyPercentage
	option.ScaleDownUnreadyTime = ua.req.ScaleDownUnreadyTime

	if ua.req.BufferResourceRatio != nil {
		option.BufferResourceRatio = ua.req.BufferResourceRatio.GetValue()
	}

	option.MaxGracefulTerminationSec = ua.req.MaxGracefulTerminationSec
	option.ScanInterval = ua.req.ScanInterval
	option.MaxNodeProvisionTime = ua.req.MaxNodeProvisionTime

	option.ScaleUpFromZero = func() bool {
		if ua.req.ScaleUpFromZero != nil {
			return ua.req.ScaleUpFromZero.GetValue()
		}
		return true
	}()

	option.ScaleDownDelayAfterAdd = ua.req.ScaleDownDelayAfterAdd
	option.ScaleDownDelayAfterDelete = ua.req.ScaleDownDelayAfterDelete
	option.ScaleDownDelayAfterFailure = func() uint32 {
		if ua.req.ScaleDownDelayAfterFailure != nil {
			return ua.req.ScaleDownDelayAfterFailure.GetValue()
		}
		return 180
	}()
	option.ScaleDownGpuUtilizationThreshold = ua.req.ScaleDownGpuUtilizationThreshold

	option.BufferResourceCpuRatio = ua.req.BufferResourceCpuRatio
	option.BufferResourceMemRatio = ua.req.BufferResourceMemRatio

	if ua.req.Module != nil {
		option.Module = ua.req.Module
	}
	if ua.req.Webhook != nil {
		option.Webhook = ua.req.Webhook
	}

	if ua.req.ExpendablePodsPriorityCutoff != nil {
		option.ExpendablePodsPriorityCutoff = ua.req.ExpendablePodsPriorityCutoff.GetValue()
	}
	if ua.req.NewPodScaleUpDelay != nil {
		option.NewPodScaleUpDelay = ua.req.NewPodScaleUpDelay.GetValue()
	}

	ua.asOption = option
	ua.resp.Data = option
	return ua.model.UpdateAutoScalingOption(ua.ctx, option)
}

func (ua *UpdateAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ua *UpdateAction) validate() error {
	err := ua.req.Validate()
	if err != nil {
		return err
	}

	return nil
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

	if err := ua.validate(); err != nil {
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
}

// SyncAction update autoscalingOption for cluster
type SyncAction struct {
	ctx      context.Context
	model    store.ClusterManagerModel
	cluster  *cmproto.Cluster
	asOption *cmproto.ClusterAutoScalingOption

	req  *cmproto.SyncAutoScalingOptionRequest
	resp *cmproto.SyncAutoScalingOptionResponse
}

// NewSyncAction create sync action for cluster autoscaling option
func NewSyncAction(model store.ClusterManagerModel) *SyncAction {
	return &SyncAction{
		model: model,
	}
}

func (ua *SyncAction) getRelativeResource() error {
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

	return nil
}

func (ua *SyncAction) syncClusterAutoScalingOption(option *cmproto.ClusterAutoScalingOption) error {
	timeStr := time.Now().Format(time.RFC3339)
	// update field if required
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

	option.SkipNodesWithLocalStorage = ua.req.SkipNodesWithLocalStorage
	option.SkipNodesWithSystemPods = ua.req.SkipNodesWithSystemPods
	option.IgnoreDaemonSetsUtilization = ua.req.IgnoreDaemonSetsUtilization

	option.OkTotalUnreadyCount = ua.req.OkTotalUnreadyCount
	option.MaxTotalUnreadyPercentage = ua.req.MaxTotalUnreadyPercentage
	option.ScaleDownUnreadyTime = ua.req.ScaleDownUnreadyTime

	// option.UnregisteredNodeRemovalTime
	option.BufferResourceRatio = ua.req.BufferResourceRatio
	option.MaxGracefulTerminationSec = ua.req.MaxGracefulTerminationSec
	option.ScanInterval = ua.req.ScanInterval
	option.MaxNodeProvisionTime = ua.req.MaxNodeProvisionTime

	option.ScaleUpFromZero = func() bool {
		if ua.req.ScaleUpFromZero != nil {
			return ua.req.ScaleUpFromZero.GetValue()
		}
		return true
	}()
	option.ScaleDownDelayAfterAdd = ua.req.ScaleDownDelayAfterAdd
	option.ScaleDownDelayAfterDelete = ua.req.ScaleDownDelayAfterDelete
	option.ScaleDownDelayAfterFailure = func() uint32 {
		if ua.req.ScaleDownDelayAfterFailure != nil {
			return ua.req.ScaleDownDelayAfterFailure.GetValue()
		}
		return 180
	}()

	option.ScaleDownGpuUtilizationThreshold = ua.req.ScaleDownGpuUtilizationThreshold

	option.BufferResourceCpuRatio = ua.req.BufferResourceCpuRatio
	option.BufferResourceMemRatio = ua.req.BufferResourceMemRatio

	if ua.req.Webhook != nil && ua.req.Webhook.Mode != "" && ua.req.Webhook.Server != "" {
		option.Webhook = ua.req.Webhook
	}

	if ua.req.ExpendablePodsPriorityCutoff != nil {
		option.ExpendablePodsPriorityCutoff = ua.req.ExpendablePodsPriorityCutoff.GetValue()
	}
	if ua.req.NewPodScaleUpDelay != nil {
		option.NewPodScaleUpDelay = ua.req.NewPodScaleUpDelay.GetValue()
	}

	ua.asOption = option
	ua.resp.Data = option
	return ua.model.UpdateAutoScalingOption(ua.ctx, option)
}

func (ua *SyncAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ua *SyncAction) validate() error {
	err := ua.req.Validate()
	if err != nil {
		return err
	}

	return nil
}

// Handle handle sync cluster autoscaling option
func (ua *SyncAction) Handle(
	ctx context.Context, req *cmproto.SyncAutoScalingOptionRequest, resp *cmproto.SyncAutoScalingOptionResponse) {

	if req == nil || resp == nil {
		blog.Errorf("sync ClusterAutoScalingOption failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// getRelativeResource get autoScalingOption / cluster
	if err := ua.getRelativeResource(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if err := ua.syncClusterAutoScalingOption(ua.asOption); err != nil {
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	err := ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.AutoScalingOption.String(),
		ResourceID:   ua.req.ClusterID,
		TaskID:       "",
		Message:      fmt.Sprintf("CA集群[%s]实际参数同步至管控端", ua.req.ClusterID),
		OpUser:       auth.GetUserFromCtx(ctx),
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    ua.req.ClusterID,
		ProjectID:    ua.cluster.ProjectID,
		ResourceName: ua.cluster.ClusterName,
	})
	if err != nil {
		blog.Errorf("SyncAutoScalingOption[%s] CreateOperationLog failed: %v", ua.req.ClusterID, err)
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// UpdateAsOptionDpAction update autoscalingOption provider pool
type UpdateAsOptionDpAction struct {
	ctx      context.Context
	model    store.ClusterManagerModel
	cluster  *cmproto.Cluster
	asOption *cmproto.ClusterAutoScalingOption

	req  *cmproto.UpdateAsOptionDeviceProviderRequest
	resp *cmproto.UpdateAsOptionDeviceProviderResponse
}

// NewUpdateAsOptionDpAction create update action for cluster autoscaling option device pool
func NewUpdateAsOptionDpAction(model store.ClusterManagerModel) *UpdateAsOptionDpAction {
	return &UpdateAsOptionDpAction{
		model: model,
	}
}

func (ua *UpdateAsOptionDpAction) getRelativeResource() error {
	// get relative cluster for information injection
	asOption, err := ua.model.GetAutoScalingOption(ua.ctx, ua.req.ClusterID)
	if err != nil {
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

	return nil
}

func (ua *UpdateAsOptionDpAction) updateClusterAsOptionDeviceProvider(option *cmproto.ClusterAutoScalingOption) error {
	timeStr := time.Now().Format(time.RFC3339)
	option.UpdateTime = timeStr
	option.DevicePoolProvider = ua.req.GetProvider()
	ua.asOption = option

	return ua.model.UpdateAutoScalingOption(ua.ctx, option)
}

func (ua *UpdateAsOptionDpAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ua *UpdateAsOptionDpAction) validate() error {
	err := ua.req.Validate()
	if err != nil {
		return err
	}

	return nil
}

// Handle handle cluster autoscaling option decide pool provider
func (ua *UpdateAsOptionDpAction) Handle(ctx context.Context,
	req *cmproto.UpdateAsOptionDeviceProviderRequest, resp *cmproto.UpdateAsOptionDeviceProviderResponse) {

	if req == nil || resp == nil {
		blog.Errorf("update ClusterAutoScalingOption DevicePool provider failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// getRelativeResource get autoScalingOption / cluster
	if err := ua.getRelativeResource(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if err := ua.updateClusterAsOptionDeviceProvider(ua.asOption); err != nil {
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	err := ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.AutoScalingOption.String(),
		ResourceID:   ua.req.ClusterID,
		TaskID:       "",
		Message:      fmt.Sprintf("更新集群[%s]扩缩容配置使用的资源池", ua.req.ClusterID),
		OpUser:       auth.GetUserFromCtx(ctx),
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    ua.req.ClusterID,
		ProjectID:    ua.cluster.ProjectID,
		ResourceName: ua.cluster.ClusterName,
	})
	if err != nil {
		blog.Errorf("UpdateAsOptionDeviceProvider[%s] CreateOperationLog failed: %v", ua.req.ClusterID, err)
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
