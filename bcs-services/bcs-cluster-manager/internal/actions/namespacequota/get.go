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

package namespacequota

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// GetAction action for get namespace quota
type GetAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.GetNamespaceQuotaReq
	resp  *cmproto.GetNamespaceQuotaResp
	quota *cmproto.ResourceQuota
}

// NewGetAction create action for get namespace quota
func NewGetAction(model store.ClusterManagerModel) *GetAction {
	return &GetAction{
		model: model,
	}
}

func (ga *GetAction) validate() error {
	if err := ga.req.Validate(); err != nil {
		return err
	}
	return nil
}

func (ga *GetAction) getQuota() error {
	quota, err := ga.model.GetQuota(ga.ctx, ga.req.Namespace, ga.req.FederationClusterID, ga.req.ClusterID)
	if err != nil {
		return err
	}
	ga.quota = quota
	return nil
}

func (ga *GetAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	ga.resp.Data = ga.quota
}

// Handle handle get namespace quota request
func (ga *GetAction) Handle(ctx context.Context,
	req *cmproto.GetNamespaceQuotaReq, resp *cmproto.GetNamespaceQuotaResp) {
	if req == nil || resp == nil {
		blog.Errorf("get namespace quota failed, req or resp is empty")
		return
	}
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	if err := ga.validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ga.getQuota(); err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
