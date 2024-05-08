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

// GetNodeRolesAction action for get node role
type GetNodeRolesAction struct {
	ctx   context.Context
	model store.ClusterManagerModel

	cloud        *cmproto.Cloud
	account      *cmproto.CloudAccount
	req          *cmproto.GetNodeRolesRequest
	resp         *cmproto.GetNodeRolesResponse
	nodeRoleList []*cmproto.NodeRoleInfo
}

// NewGetNodeRolesAction create list action for node role
func NewGetNodeRolesAction(model store.ClusterManagerModel) *GetNodeRolesAction {
	return &GetNodeRolesAction{
		model: model,
	}
}

func (ga *GetNodeRolesAction) listNodeRoles() error {
	nodeMgr, err := cloudprovider.GetNodeMgr(ga.cloud.CloudProvider)
	if err != nil {
		return err
	}

	nodeRoleList, err := nodeMgr.GetNodeRoles(&cloudprovider.CommonOption{
		Account: func() *cmproto.Account {
			if ga.account != nil {
				return ga.account.Account
			}
			return nil
		}(),
	})
	if err != nil {
		return err
	}

	ga.nodeRoleList = nodeRoleList
	return nil
}

func (ga *GetNodeRolesAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	ga.resp.Data = ga.nodeRoleList
}

func (ga *GetNodeRolesAction) validate() error {
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

func (ga *GetNodeRolesAction) getRelativeData() error {
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
func (ga *GetNodeRolesAction) Handle(
	ctx context.Context, req *cmproto.GetNodeRolesRequest, resp *cmproto.GetNodeRolesResponse) {
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
	if err := ga.listNodeRoles(); err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
