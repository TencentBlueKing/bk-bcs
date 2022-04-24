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

package cloud

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// GetAction action for getting cluster credential
type GetAction struct {
	ctx context.Context

	model store.ClusterManagerModel
	req   *cmproto.GetCloudRequest
	resp  *cmproto.GetCloudResponse
}

// NewGetAction create get action for online cluster credential
func NewGetAction(model store.ClusterManagerModel) *GetAction {
	return &GetAction{
		model: model,
	}
}

func (ga *GetAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handle get cluster credential
func (ga *GetAction) Handle(
	ctx context.Context, req *cmproto.GetCloudRequest, resp *cmproto.GetCloudResponse) {
	if req == nil || resp == nil {
		blog.Errorf("get cloud failed, req or resp is empty")
		return
	}
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	if err := req.Validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	cloud, err := ga.model.GetCloud(ctx, req.CloudID)
	if err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	if cloud.CloudCredential != nil && !req.ShowCredential {
		cloud.CloudCredential.Key = ""
		cloud.CloudCredential.Secret = ""
	}

	resp.Data = cloud
	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
