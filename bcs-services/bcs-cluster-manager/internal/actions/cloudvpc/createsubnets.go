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

// CreateSubnetsAction action for create subnets
type CreateSubnetsAction struct {
	ctx     context.Context
	cloud   *cmproto.Cloud
	account *cmproto.CloudAccount

	model  store.ClusterManagerModel
	req    *cmproto.CreateCloudSubnetsRequest
	resp   *cmproto.CreateCloudSubnetsResponse
	subnet *cmproto.Subnet
}

// NewCreateSubnetsAction create action for subnets
func NewCreateSubnetsAction(model store.ClusterManagerModel) *CreateSubnetsAction {
	return &CreateSubnetsAction{
		model: model,
	}
}

func (la *CreateSubnetsAction) validate() error {
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
		err = validate.CreateCloudSubnetsValidate(la.req, la.account.Account)
		if err != nil {
			return err
		}
	}

	return nil
}

func (la *CreateSubnetsAction) getRelativeData() error {
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

func (la *CreateSubnetsAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.subnet
}

// CreateCloudSubnets create cloud subnets
func (la *CreateSubnetsAction) CreateCloudSubnets() (*cmproto.Subnet, error) {
	// create vpc client with cloudProvider
	vpcMgr, err := cloudprovider.GetVPCMgr(la.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s VPCManager for create subnets failed, %s", la.cloud.CloudProvider, err.Error())
		return nil, err
	}

	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     la.cloud,
		AccountID: la.req.AccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s create subnets failed, %s",
			la.cloud.CloudID, la.cloud.CloudProvider, err.Error())
		return nil, err
	}
	cmOption.Region = la.req.Region

	// create subnet
	return vpcMgr.CreateSubnets(&cloudprovider.NetworksSubnetOption{
		CommonOption: *cmOption,
		Subnets: cloudprovider.Subnets{
			VpcId:      la.req.VpcID,
			CidrBlock:  la.req.CidrBlock,
			Zone:       la.req.Zone,
			SubnetName: la.req.SubnetName,
		},
	})

}

// Handle handle list vpc subnets
func (la *CreateSubnetsAction) Handle(
	ctx context.Context, req *cmproto.CreateCloudSubnetsRequest, resp *cmproto.CreateCloudSubnetsResponse) {
	if req == nil || resp == nil {
		blog.Errorf("create subnets failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	subnet, err := la.CreateCloudSubnets()
	if err != nil {
		la.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	la.subnet = subnet
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
