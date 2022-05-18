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
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// DeleteAction action for delete cloudAccount
type DeleteAction struct {
	ctx context.Context

	model store.ClusterManagerModel
	req   *cmproto.DeleteCloudAccountRequest
	resp  *cmproto.DeleteCloudAccountResponse
}

// NewDeleteAction create delete action for cloudAccount
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

func (da *DeleteAction) validate() error {
	err := da.req.Validate()
	if err != nil {
		return err
	}
	clusters, err := getRelativeClustersByAccountID(da.ctx, da.model, CloudAccount{
		CloudID:   da.req.CloudID,
		AccountID: da.req.AccountID,
	})
	if err != nil {
		return err
	}
	if len(clusters) > 0 {
		return fmt.Errorf("cloudAccount[%s] can't be deleted, exist clusters using", da.req.AccountID)
	}

	return nil
}

// Handle handle delete cloud account
func (da *DeleteAction) Handle(
	ctx context.Context, req *cmproto.DeleteCloudAccountRequest, resp *cmproto.DeleteCloudAccountResponse) {
	if req == nil || resp == nil {
		blog.Errorf("delete cloudAccount failed, req or resp is empty")
		return
	}
	da.ctx = ctx
	da.req = req
	da.resp = resp

	if err := da.validate(); err != nil {
		da.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// try to get original cloud account for return
	_, err := da.model.GetCloudAccount(da.ctx, da.req.CloudID, da.req.AccountID)
	if err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("Get CloudAccount %s:%s in pre-delete checking failed, err %s", da.req.CloudID, da.req.AccountID, err.Error())
		return
	}

	if err = da.model.DeleteCloudAccount(da.ctx, da.req.CloudID, da.req.AccountID); err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
