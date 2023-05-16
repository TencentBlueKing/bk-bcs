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

// ListNodeTypeAction list action for node type
type ListNodeTypeAction struct {
	ctx          context.Context
	cloud        *cmproto.Cloud
	account      *cmproto.CloudAccount
	model        store.ClusterManagerModel
	req          *cmproto.ListCloudInstanceTypeRequest
	resp         *cmproto.ListCloudInstanceTypeResponse
	nodeTypeList []*cmproto.InstanceType
}

// NewListNodeTypeAction create list action for node type
func NewListNodeTypeAction(model store.ClusterManagerModel) *ListNodeTypeAction {
	return &ListNodeTypeAction{
		model: model,
	}
}

func (la *ListNodeTypeAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}
	err := la.getRelativeData()
	if err != nil {
		return err
	}

	validate, err := cloudprovider.GetCloudValidateMgr(la.cloud.CloudProvider)
	if err != nil {
		return err
	}

	err = validate.ListInstanceTypeValidate(la.req, la.account.Account)
	if err != nil {
		return err
	}

	return nil
}

func (la *ListNodeTypeAction) getRelativeData() error {
	cloud, err := actions.GetCloudByCloudID(la.model, la.req.CloudID)
	if err != nil {
		return err
	}
	account, err := la.model.GetCloudAccount(la.ctx, la.req.CloudID, la.req.AccountID)
	if err != nil {
		return err
	}

	la.account = account
	la.cloud = cloud
	return nil
}

func (la *ListNodeTypeAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.nodeTypeList
}

func (la *ListNodeTypeAction) listCloudInstancetypes() error {
	// create vpc client with cloudProvider
	nodeMgr, err := cloudprovider.GetNodeMgr(la.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s VPCManager for list subnets failed, %s", la.cloud.CloudProvider, err.Error())
		return err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     la.cloud,
		AccountID: la.req.AccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s list subnets failed, %s",
			la.cloud.CloudID, la.cloud.CloudProvider, err.Error())
		return err
	}
	cmOption.Region = la.req.Region

	// get instance types list
	insTypes, err := nodeMgr.ListNodeInstanceType(la.req.Zone, la.req.NodeFamily,
		la.req.Cpu, la.req.Memory, cmOption)
	if err != nil {
		return err
	}
	la.nodeTypeList = insTypes

	return nil
}

// Handle handle list node type request
func (la *ListNodeTypeAction) Handle(ctx context.Context,
	req *cmproto.ListCloudInstanceTypeRequest, resp *cmproto.ListCloudInstanceTypeResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list node type failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.listCloudInstancetypes(); err != nil {
		la.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
