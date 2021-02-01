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

// UpdateAction action for update namespace
type UpdateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.UpdateNamespaceReq
	resp  *cmproto.UpdateNamespaceResp
}

// NewUpdateAction create action for udpate
func NewUpdateAction(model store.ClusterManagerModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

func (ua *UpdateAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}
	return nil
}

func (ua *UpdateAction) updateNamespace() error {
	newNs := &types.Namespace{
		Name:                ua.req.Name,
		FederationClusterID: ua.req.FederationClusterID,
		ProjectID:           ua.req.ProjectID,
		BusinessID:          ua.req.BusinessID,
		Labels:              ua.req.Labels,
	}
	return ua.model.UpdateNamespace(ua.ctx, newNs)
}

func (ua *UpdateAction) setResp(code uint64, msg string) {
	ua.resp.Seq = ua.req.Seq
	ua.resp.ErrCode = code
	ua.resp.ErrMsg = msg
}

// Handle update namespace request
func (ua *UpdateAction) Handle(ctx context.Context,
	req *cmproto.UpdateNamespaceReq, resp *cmproto.UpdateNamespaceResp) {
	if req == nil || resp == nil {
		blog.Errorf("update cluster failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(types.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ua.updateNamespace(); err != nil {
		ua.setResp(types.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ua.setResp(types.BcsErrClusterManagerSuccess, types.BcsErrClusterManagerSuccessStr)
	return
}
