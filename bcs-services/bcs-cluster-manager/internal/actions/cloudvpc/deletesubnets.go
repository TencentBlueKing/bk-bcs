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

package cloudvpc

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// DeleteSubnetsAction action for delete subnets
type DeleteSubnetsAction struct {
	ctx     context.Context
	cloud   *cmproto.Cloud
	account *cmproto.CloudAccount

	model store.ClusterManagerModel
	req   *cmproto.DeleteCloudSubnetsRequest
	resp  *cmproto.DeleteCloudSubnetsResponse
}

// NewDeleteSubnetsAction create action for subnets
func NewDeleteSubnetsAction(model store.ClusterManagerModel) *DeleteSubnetsAction {
	return &DeleteSubnetsAction{
		model: model,
	}
}

func (la *DeleteSubnetsAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}

	// get cloud/account info
	err := la.getRelativeData()
	if err != nil {
		return err
	}

	validate, err := cloudprovider.GetCloudValidateMgr(la.cloud.CloudProvider)
	if err != nil {
		return err
	}

	if la.account != nil {
		err = validate.DeleteCloudSubnetsValidate(la.req, la.account.Account)
		if err != nil {
			return err
		}
	}

	return nil
}

func (la *DeleteSubnetsAction) getRelativeData() error {
	cloud, err := actions.GetCloudByCloudID(la.model, la.req.CloudID)
	if err != nil {
		return err
	}

	if la.req.GetAccountID() != "" {
		account, errLocal := la.model.GetCloudAccount(la.ctx, la.req.CloudID, la.req.AccountID, false)
		if errLocal != nil {
			return errLocal
		}

		la.account = account
	}

	la.cloud = cloud
	return nil
}

func (la *DeleteSubnetsAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// DeleteCloudSubnets delete cloud subnets
func (la *DeleteSubnetsAction) DeleteCloudSubnets() error {
	// create vpc client with cloudProvider
	vpcMgr, err := cloudprovider.GetVPCMgr(la.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s VPCManager for delete subnets failed, %s", la.cloud.CloudProvider, err.Error())
		return err
	}

	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     la.cloud,
		AccountID: la.req.AccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s delete subnets failed, %s",
			la.cloud.CloudID, la.cloud.CloudProvider, err.Error())
		return err
	}
	cmOption.Region = la.req.Region

	// create subnet
	return vpcMgr.DeleteSubnets(&cloudprovider.NetworksSubnetOption{
		CommonOption: *cmOption,
		Subnets: cloudprovider.Subnets{
			SubnetId: la.req.SubnetID,
		},
	})

}

// Handle handle delete vpc subnets
func (la *DeleteSubnetsAction) Handle(
	ctx context.Context, req *cmproto.DeleteCloudSubnetsRequest, resp *cmproto.DeleteCloudSubnetsResponse) {
	if req == nil || resp == nil {
		blog.Errorf("delete subnets failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.DeleteCloudSubnets(); err != nil {
		la.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
