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
 *
 */

package account

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// UpdateAction update action for cloud account
type UpdateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.UpdateCloudAccountRequest
	resp  *cmproto.UpdateCloudAccountResponse
}

// NewUpdateAction create update action for cloud account
func NewUpdateAction(model store.ClusterManagerModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

func (ua *UpdateAction) updateCloudAccount(destCloudAccount *cmproto.CloudAccount) error {
	timeStr := time.Now().Format(time.RFC3339)
	destCloudAccount.UpdateTime = timeStr
	destCloudAccount.Updater = ua.req.Updater

	if len(ua.req.AccountName) != 0 {
		destCloudAccount.AccountName = ua.req.AccountName
	}
	if len(ua.req.Desc) != 0 {
		destCloudAccount.Desc = ua.req.Desc
	}
	if ua.req.Enable != nil {
		destCloudAccount.Enable = ua.req.Enable.GetValue()
	}
	if len(ua.req.ProjectID) != 0 {
		destCloudAccount.ProjectID = ua.req.ProjectID
	}

	return ua.model.UpdateCloudAccount(ua.ctx, destCloudAccount)
}

func (ua *UpdateAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handle update cloud account
func (ua *UpdateAction) Handle(
	ctx context.Context, req *cmproto.UpdateCloudAccountRequest, resp *cmproto.UpdateCloudAccountResponse) {

	if req == nil || resp == nil {
		blog.Errorf("update cloudAccount failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := req.Validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	destCloudAccount, err := ua.model.GetCloudAccount(ua.ctx, req.CloudID, req.AccountID)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("find cloudAccount %s failed when pre-update checking, err %s", req.AccountID, err.Error())
		return
	}
	if err := ua.updateCloudAccount(destCloudAccount); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
