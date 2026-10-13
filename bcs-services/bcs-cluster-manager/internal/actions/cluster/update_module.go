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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// UpdateClusterModuleAction action for update cluster module
type UpdateClusterModuleAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	req     *cmproto.UpdateClusterModuleRequest
	resp    *cmproto.UpdateClusterModuleResponse
	cluster *cmproto.Cluster
}

// NewUpdateClusterModuleAction create update module action
func NewUpdateClusterModuleAction(model store.ClusterManagerModel) *UpdateClusterModuleAction {
	return &UpdateClusterModuleAction{
		model: model,
	}
}

// validate check
func (ua *UpdateClusterModuleAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}
	return nil
}

// getCluster cluster/cloud
func (ua *UpdateClusterModuleAction) getCluster() error {
	cluster, err := ua.model.GetCluster(ua.ctx, ua.req.ClusterID)
	if err != nil {
		return err
	}
	ua.cluster = cluster

	return nil
}

// updateClusterModuleInfo update cluster node module
func (ua *UpdateClusterModuleAction) updateClusterModuleInfo() error {
	if ua.cluster.GetClusterBasicSettings().GetModule() == nil {
		ua.cluster.GetClusterBasicSettings().Module = &cmproto.ClusterModule{}
	}

	if ua.req.GetModule().GetMasterModuleID() != "" {
		updateClusterModule(ua.cluster, master,
			ua.req.GetModule().GetMasterModuleID(), ua.req.GetModule().GetMasterModuleName())
		err := transClusterNodes(ua.model, ua.cluster, master, ua.req.GetModule().GetMasterModuleID())
		if err != nil {
			blog.Errorf("updateClusterModuleInfo[%s] category[%s] transClusterNodes failed: %v",
				ua.req.GetClusterID(), master, err)
		}
	}
	if ua.req.GetModule().GetWorkerModuleID() != "" {
		updateClusterModule(ua.cluster, worker,
			ua.req.GetModule().GetWorkerModuleID(), ua.req.GetModule().GetWorkerModuleName())
		err := transClusterNodes(ua.model, ua.cluster, worker, ua.req.GetModule().GetWorkerModuleID())
		if err != nil {
			blog.Errorf("updateClusterModuleInfo[%s] category[%s] transClusterNodes failed: %v",
				ua.req.GetClusterID(), worker, err)
		}
	}
	ua.cluster.UpdateTime = time.Now().UTC().Format(time.RFC3339)
	err := ua.model.UpdateCluster(ua.ctx, ua.cluster)
	if err != nil {
		return err
	}

	// update autoscaling option
	option, err := ua.model.GetAutoScalingOption(ua.ctx, ua.req.GetClusterID())
	if err == nil {
		updateAutoScalingModule(ua.cluster, option,
			ua.req.GetModule().GetWorkerModuleID(), ua.req.GetModule().GetWorkerModuleName())
		_ = ua.model.UpdateAutoScalingOption(ua.ctx, option)
	}

	// record operation log
	err = ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.Cluster.String(),
		ResourceID:   ua.req.GetClusterID(),
		TaskID:       "",
		Message:      fmt.Sprintf("集群%s更新节点模块[%s]", ua.req.ClusterID, ua.req.GetModule().GetWorkerModuleID()),
		OpUser:       auth.GetUserFromCtx(ua.ctx),
		CreateTime:   time.Now().UTC().Format(time.RFC3339),
		ClusterID:    ua.req.ClusterID,
		ProjectID:    ua.cluster.ProjectID,
		ResourceName: ua.cluster.ClusterName,
	})
	if err != nil {
		blog.Errorf("updateClusterModuleInfo[%s] CreateOperationLog failed: %v", ua.cluster.ClusterID, err)
	}

	return nil
}

func (ua *UpdateClusterModuleAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = code == common.BcsErrClusterManagerSuccess
}

// Handle handles update cluster module action
func (ua *UpdateClusterModuleAction) Handle(ctx context.Context,
	req *cmproto.UpdateClusterModuleRequest, resp *cmproto.UpdateClusterModuleResponse) {
	if req == nil || resp == nil {
		blog.Errorf("update cluster module failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ua.getCluster(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if err := ua.updateClusterModuleInfo(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
