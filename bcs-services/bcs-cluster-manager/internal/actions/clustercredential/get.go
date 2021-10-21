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
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

// GetAction action for getting cluster credential
type GetAction struct {
	ctx context.Context

	model       store.ClusterManagerModel
	req         *cmproto.GetClusterCredentialReq
	resp        *cmproto.GetClusterCredentialResp
	clusterCred *cmproto.ClusterCredential
}

// NewGetAction create get action for online cluster credential
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

func (ga *GetAction) getCredential() error {
	cred, isExisted, err := ga.model.GetClusterCredential(ga.ctx, ga.req.ServerKey)
	if err != nil {
		return err
	}
	if !isExisted {
		return fmt.Errorf("credential with serverkey %s not found", ga.req.ServerKey)
	}
	ga.clusterCred = &cmproto.ClusterCredential{
		ServerKey:     cred.ServerKey,
		ClusterID:     cred.ClusterID,
		ClientModule:  cred.ClientModule,
		ServerAddress: cred.ServerAddress,
		CaCertData:    cred.CaCertData,
		UserToken:     cred.UserToken,
		ClusterDomain: cred.ClusterDomain,
		ConnectMode:   cred.ConnectMode,
		CreateTime:    cred.CreateTime.String(),
		UpdateTime:    cred.UpdateTime.String(),
	}
	return nil
}

func (ga *GetAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == types.BcsErrClusterManagerSuccess)
	ga.resp.Data = ga.clusterCred
}

// Handle handle get cluster credential
func (ga *GetAction) Handle(
	ctx context.Context, req *cmproto.GetClusterCredentialReq, resp *cmproto.GetClusterCredentialResp) {
	if req == nil || resp == nil {
		blog.Errorf("get cluster credential failed, req or resp is empty")
		return
	}
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	if err := ga.validate(); err != nil {
		ga.setResp(types.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ga.getCredential(); err != nil {
		ga.setResp(types.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ga.setResp(types.BcsErrClusterManagerSuccess, types.BcsErrClusterManagerSuccessStr)
	return
}
