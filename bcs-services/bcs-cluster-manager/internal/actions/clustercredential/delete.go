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

package clustercredential

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// DeleteAction action for delete clustercredentail
type DeleteAction struct {
	ctx context.Context

	model store.ClusterManagerModel
	req   *cmproto.DeleteClusterCredentialReq
	resp  *cmproto.DeleteClusterCredentialResp
}

// NewDeleteAction create delete action for online cluster credential
func NewDeleteAction(model store.ClusterManagerModel) *DeleteAction {
	return &DeleteAction{
		model: model,
	}
}

func (da *DeleteAction) validate() error {
	if err := da.req.Validate(); err != nil {
		return err
	}
	return nil
}

func (da *DeleteAction) deleteCredential() error {
	return da.model.DeleteClusterCredential(da.ctx, da.req.ServerKey)
}

func (da *DeleteAction) setResp(code uint32, msg string) {
	da.resp.Code = code
	da.resp.Message = msg
	da.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handle delete cluster credential
func (da *DeleteAction) Handle(
	ctx context.Context, req *cmproto.DeleteClusterCredentialReq, resp *cmproto.DeleteClusterCredentialResp) {
	if req == nil || resp == nil {
		blog.Errorf("delete cluster credential failed, req or resp is empty")
		return
	}
	da.ctx = ctx
	da.req = req
	da.resp = resp

	if err := da.validate(); err != nil {
		da.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := da.deleteCredential(); err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
	}
	da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
