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

package namespace

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

// DeleteAction action for delete namespace
type DeleteAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.DeleteNamespaceReq
	resp  *cmproto.DeleteNamespaceResp
}

// NewDeleteAction delete namespace
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

func (da *DeleteAction) deleteNamespace() error {
	return da.model.DeleteNamespace(da.ctx, da.req.Name, da.req.FederationClusterID)
}

func (da *DeleteAction) setResp(code uint64, msg string) {
	da.resp.Seq = da.req.Seq
	da.resp.ErrCode = code
	da.resp.ErrMsg = msg
}

// Handle handle delete namespace reqeust
func (da *DeleteAction) Handle(ctx context.Context,
	req *cmproto.DeleteNamespaceReq, resp *cmproto.DeleteNamespaceResp) {
	if req == nil || resp == nil {
		blog.Errorf("delete namespace failed, req or resp is empty")
		return
	}
	da.ctx = ctx
	da.req = req
	da.resp = resp

	if err := da.validate(); err != nil {
		da.setResp(types.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := da.deleteNamespace(); err != nil {
		da.setResp(types.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	da.setResp(types.BcsErrClusterManagerSuccess, types.BcsErrClusterManagerSuccessStr)
	return
}
