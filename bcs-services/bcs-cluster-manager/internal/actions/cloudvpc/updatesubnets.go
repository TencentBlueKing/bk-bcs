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

// UpdateSubnetsAction action for update subnets
type UpdateSubnetsAction struct {
	ctx     context.Context
	cloud   *cmproto.Cloud
	account *cmproto.CloudAccount

	model store.ClusterManagerModel
	req   *cmproto.UpdateCloudSubnetsRequest
	resp  *cmproto.UpdateCloudSubnetsResponse
}

// NewUpdateSubnetsAction create update action for subnets
func NewUpdateSubnetsAction(model store.ClusterManagerModel) *UpdateSubnetsAction {
	return &UpdateSubnetsAction{
		model: model,
	}
}

func (la *UpdateSubnetsAction) validate() error {
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
		err = validate.UpdateCloudSubnetsValidate(la.req, la.account.Account)
		if err != nil {
			return err
		}
	}

	return nil
}

func (la *UpdateSubnetsAction) getRelativeData() error {
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

func (la *UpdateSubnetsAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// UpdateCloudSubnets update cloud subnets
func (la *UpdateSubnetsAction) UpdateCloudSubnets() error {
	// create vpc client with cloudProvider
	vpcMgr, err := cloudprovider.GetVPCMgr(la.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s VPCManager for update subnets failed, %s", la.cloud.CloudProvider, err.Error())
		return err
	}

	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     la.cloud,
		AccountID: la.req.AccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s update subnets failed, %s",
			la.cloud.CloudID, la.cloud.CloudProvider, err.Error())
		return err
	}
	cmOption.Region = la.req.Region

	// update subnet
	return vpcMgr.UpdateSubnets(&cloudprovider.NetworksSubnetOption{
		CommonOption: *cmOption,
		Subnets: cloudprovider.Subnets{
			SubnetId:   la.req.SubnetID,
			SubnetName: la.req.SubnetName,
		},
	})

}

// Handle handle list vpc subnets
func (la *UpdateSubnetsAction) Handle(
	ctx context.Context, req *cmproto.UpdateCloudSubnetsRequest, resp *cmproto.UpdateCloudSubnetsResponse) {
	if req == nil || resp == nil {
		blog.Errorf("update subnets failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.UpdateCloudSubnets(); err != nil {
		la.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
