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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// DeleteAction action for delete clustercredentail
type DeleteAction struct {
	ctx context.Context

	model store.ClusterManagerModel
	req   *cmproto.DeleteAutoScalingOptionRequest
	resp  *cmproto.DeleteAutoScalingOptionResponse

	cluster *cmproto.Cluster
}

// NewDeleteAction create delete action for online cluster credential
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

// nolint result error is always nil
func (da *DeleteAction) getRelativeResource() error {
	cluster, err := da.model.GetCluster(da.ctx, da.req.ClusterID)
	if err == nil {
		da.cluster = cluster
	}

	return nil
}

// Handle handle delete cluster credential
func (da *DeleteAction) Handle(
	ctx context.Context, req *cmproto.DeleteAutoScalingOptionRequest, resp *cmproto.DeleteAutoScalingOptionResponse) {
	if req == nil || resp == nil {
		blog.Errorf("delete ClusterAutoScalingOption failed, req or resp is empty")
		return
	}
	da.ctx = ctx
	da.req = req
	da.resp = resp

	if err := req.Validate(); err != nil {
		da.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := da.getRelativeResource(); err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// try to get original data for return
	deleteOption, err := da.model.GetAutoScalingOption(da.ctx, da.req.ClusterID)
	if err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	da.resp.Data = deleteOption
	if err = da.model.DeleteAutoScalingOption(da.ctx, da.req.ClusterID); err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// create operationLog
	err = da.model.CreateOperationLog(da.ctx, &cmproto.OperationLog{
		ResourceType: common.AutoScalingOption.String(),
		ResourceID:   da.req.ClusterID,
		TaskID:       "",
		Message:      fmt.Sprintf("删除集群[%s]扩缩容配置", da.req.ClusterID),
		OpUser:       auth.GetUserFromCtx(ctx),
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    da.req.ClusterID,
		ProjectID:    da.cluster.ProjectID,
		ResourceName: da.cluster.ClusterName,
	})
	if err != nil {
		blog.Errorf("DeleteAutoScalingOption[%s] CreateOperationLog failed: %v", da.req.ClusterID, err)
	}

	da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
