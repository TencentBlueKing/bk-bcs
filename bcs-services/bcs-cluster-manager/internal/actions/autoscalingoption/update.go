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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// UpdateAction update action for online cluster credential
type UpdateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.UpdateAutoScalingOptionRequest
	resp  *cmproto.UpdateAutoScalingOptionResponse
}

// NewUpdateAction create update action for online cluster credential
func NewUpdateAction(model store.ClusterManagerModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

func (ua *UpdateAction) updateAutoScalingOption(option *cmproto.ClusterAutoScalingOption) error {
	timeStr := time.Now().Format(time.RFC3339)
	// update field if required
	option.UpdateTime = timeStr
	option.Updater = ua.req.Updater
	// update clusterAutoScalingOption, should not update cluster autoscaling options
	// if need to update config, you must reinstall autoscaler

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

	//get old ClusterAutoScalingOption information, update fields if required
	destOption, err := ua.model.GetAutoScalingOption(ua.ctx, req.ClusterID)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("find ClusterAutoScalingOption %s failed when pre-update checking, err %s", req.ClusterID, err.Error())
		return
	}
	if err := ua.updateAutoScalingOption(destOption); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ua.resp.Data = destOption
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
