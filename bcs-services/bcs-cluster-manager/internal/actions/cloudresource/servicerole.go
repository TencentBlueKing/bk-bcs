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

package cloudresource

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// GetServiceRolesAction action for get node role
type GetServiceRolesAction struct {
	ctx   context.Context
	model store.ClusterManagerModel

	cloud           *cmproto.Cloud
	account         *cmproto.CloudAccount
	req             *cmproto.GetServiceRolesRequest
	resp            *cmproto.GetServiceRolesResponse
	serviceRoleList []*cmproto.ServiceRoleInfo
}

// NewGetServiceRolesAction create list action for node role
func NewGetServiceRolesAction(model store.ClusterManagerModel) *GetServiceRolesAction {
	return &GetServiceRolesAction{
		model: model,
	}
}

func (ga *GetServiceRolesAction) listServiceRoles() error {
	nodeMgr, err := cloudprovider.GetNodeMgr(ga.cloud.CloudProvider)
	if err != nil {
		return err
	}

	serviceRoleList, err := nodeMgr.GetServiceRoles(&cloudprovider.CommonOption{
		Account: func() *cmproto.Account {
			if ga.account != nil {
				return ga.account.Account
			}
			return nil
		}(),
	}, ga.req.RoleType)
	if err != nil {
		return err
	}

	ga.serviceRoleList = serviceRoleList
	return nil
}

func (ga *GetServiceRolesAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	ga.resp.Data = ga.serviceRoleList
}

func (ga *GetServiceRolesAction) validate() error {
	err := ga.req.Validate()
	if err != nil {
		return err
	}

	err = ga.getRelativeData()
	if err != nil {
		return err
	}

	return nil
}

func (ga *GetServiceRolesAction) getRelativeData() error {
	cloud, err := actions.GetCloudByCloudID(ga.model, ga.req.CloudID)
	if err != nil {
		return err
	}
	ga.cloud = cloud

	if ga.req.GetAccountID() != "" {
		account, errLocal := ga.model.GetCloudAccount(ga.ctx, ga.req.CloudID, ga.req.AccountID, false)
		if errLocal != nil {
			return errLocal
		}

		ga.account = account
	}

	return nil
}

// Handle handle list node role
func (ga *GetServiceRolesAction) Handle(
	ctx context.Context, req *cmproto.GetServiceRolesRequest, resp *cmproto.GetServiceRolesResponse) {
	if req == nil || resp == nil {
		blog.Errorf("get node role failed, req or resp is empty")
		return
	}
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	if err := ga.validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ga.listServiceRoles(); err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
