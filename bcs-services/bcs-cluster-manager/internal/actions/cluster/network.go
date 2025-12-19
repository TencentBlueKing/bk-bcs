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

package cluster

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

// SwitchClusterUnderlayNetworkAction for cluster switch network
type SwitchClusterUnderlayNetworkAction struct {
	ctx   context.Context
	model store.ClusterManagerModel

	req  *cmproto.SwitchClusterUnderlayNetworkReq
	resp *cmproto.SwitchClusterUnderlayNetworkResp

	cluster *cmproto.Cluster
	cloud   *cmproto.Cloud
	task    *cmproto.Task
}

// NewSwitchClsNetworkAction create on/off underlay network action
func NewSwitchClsNetworkAction(model store.ClusterManagerModel) *SwitchClusterUnderlayNetworkAction {
	return &SwitchClusterUnderlayNetworkAction{
		model: model,
	}
}

func (sa *SwitchClusterUnderlayNetworkAction) setResp(code uint32, msg string) {
	sa.resp.Code = code
	sa.resp.Message = msg
	sa.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	sa.resp.Task = sa.task
	sa.resp.Cluster = sa.cluster
}

func (sa *SwitchClusterUnderlayNetworkAction) validate() error {
	err := sa.req.Validate()
	if err != nil {
		return err
	}

	if !sa.req.GetDisable() {
		if sa.req.Subnet == nil || (len(sa.req.Subnet.GetNew()) == 0 && len(sa.req.Subnet.GetExisted().GetIds()) == 0) {
			return fmt.Errorf("SwitchClusterUnderlayNetworkAction subnetInfo empty")
		}
	}

	return nil
}

func (sa *SwitchClusterUnderlayNetworkAction) getRelativeData() error {
	cloud, cluster, err := actions.GetCloudAndCluster(sa.model, "", sa.req.GetClusterID())
	if err != nil {
		return err
	}

	sa.cluster = cluster
	sa.cloud = cloud

	return nil
}

func (sa *SwitchClusterUnderlayNetworkAction) checkClusterNetworkStatus() (bool, error) {
	clsMgr, err := cloudprovider.GetClusterMgr(sa.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("cloudProvider %s get cluster manager failed", sa.cloud.CloudProvider)
		return false, err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     sa.cloud,
		AccountID: sa.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("", "get credential failed, err %s", err.Error())
		return false, err
	}
	cmOption.Region = sa.cluster.Region

	return clsMgr.CheckClusterNetworkStatus(sa.cluster.SystemID, &cloudprovider.CheckClusterNetworkStatusOption{
		CommonOption: *cmOption,
		Cluster:      sa.cluster,
		Disable:      sa.req.Disable,

		SubnetSource:        sa.req.Subnet,
		IsStaticIPMode:      sa.req.IsStaticIpMode,
		ClaimExpiredSeconds: sa.req.GetClaimExpiredSeconds(),
	})
}

// Handle switch cluster network action
func (sa *SwitchClusterUnderlayNetworkAction) Handle(ctx context.Context,
	req *cmproto.SwitchClusterUnderlayNetworkReq, resp *cmproto.SwitchClusterUnderlayNetworkResp) {
	if req == nil || resp == nil {
		blog.Errorf("SwitchClusterUnderlayNetworkReq failed, req or resp is empty")
		return
	}
	sa.ctx = ctx
	sa.req = req
	sa.resp = resp

	if err := sa.validate(); err != nil {
		sa.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := sa.getRelativeData(); err != nil {
		sa.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	createTask, err := sa.checkClusterNetworkStatus()
	if err != nil {
		sa.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	if createTask {
		if err = sa.switchClusterNetworkTask(); err != nil {
			sa.setResp(common.BcsErrClusterManagerClsMgrCloudErr, err.Error())
			return
		}
		// 同步网络状态
		sa.cluster.NetworkSettings.Status = common.StatusInitialization
		sa.cluster.NetworkSettings.EnableVPCCni = !sa.req.Disable
	}

	// update cluster
	if err = sa.model.UpdateCluster(sa.ctx, sa.cluster); err != nil {
		blog.Errorf("SwitchClusterUnderlayNetworkAction update cluster %s failed: %+v",
			sa.cluster.ClusterID, err)
		sa.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// build operationLog
	err = sa.model.CreateOperationLog(ctx, &cmproto.OperationLog{
		ResourceType: common.Cluster.String(),
		ResourceID:   sa.cluster.ClusterID,
		TaskID: func() string {
			if sa.task != nil {
				return sa.task.TaskID
			}
			return ""
		}(),
		Message: func() string {
			if sa.req.GetDisable() {
				return fmt.Sprintf("集群[%s]关闭underlay网络模式", sa.cluster.ClusterID)
			}
			return fmt.Sprintf("集群[%s]开启underlay网络模式", sa.cluster.ClusterID)
		}(),
		OpUser:       sa.req.GetOperator(),
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    sa.cluster.ClusterID,
		ProjectID:    sa.cluster.ProjectID,
		ResourceName: sa.cluster.GetClusterName(),
	})
	if err != nil {
		blog.Errorf("switchClusterUnderlayNetwork[%s] CreateOperationLog failed: %v", sa.req.ClusterID, err)
	}

	sa.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

func (sa *SwitchClusterUnderlayNetworkAction) switchClusterNetworkTask() error {
	provider, err := cloudprovider.GetClusterMgr(sa.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cluster %s relative cloud provider %s failed, %s",
			sa.req.ClusterID, sa.cloud.CloudProvider, err.Error())
		sa.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return err
	}

	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     sa.cloud,
		AccountID: sa.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s switchClusterNetwork failed, %s",
			sa.cloud.CloudID, sa.cloud.CloudProvider, err.Error())
		sa.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return err
	}
	cmOption.Region = sa.cluster.Region

	// switch cluster network task by task manager
	task, err := provider.SwitchClusterNetwork(sa.cluster, sa.req.Subnet, &cloudprovider.SwitchClusterNetworkOption{
		CommonOption:        *cmOption,
		Operator:            sa.req.GetOperator(),
		Cloud:               sa.cloud,
		Disable:             sa.req.Disable,
		IsStaticIPMode:      sa.req.IsStaticIpMode,
		ClaimExpiredSeconds: sa.req.ClaimExpiredSeconds,
	})
	if err != nil {
		blog.Errorf("switch cluster network %s by Cloud %s with provider %s failed, %s",
			sa.cluster.ClusterID, sa.cloud.CloudID, sa.cloud.CloudProvider, err.Error())
		sa.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return err
	}

	// create task and dispatch task
	if err = sa.model.CreateTask(sa.ctx, task); err != nil {
		blog.Errorf("switch cluster network task for cluster %s failed, %s",
			sa.cluster.ClusterName, err.Error(),
		)
		sa.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	if err = taskserver.GetTaskServer().Dispatch(task); err != nil {
		blog.Errorf("dispatch create cluster task for cluster %s failed, %s",
			sa.cluster.ClusterName, err.Error(),
		)
		sa.setResp(common.BcsErrClusterManagerTaskErr, err.Error())
		return err
	}

	sa.task = task
	return nil
}
