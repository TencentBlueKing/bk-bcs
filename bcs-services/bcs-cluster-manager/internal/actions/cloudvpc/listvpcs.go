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

// ListVPCV2Action action for list vpcs
type ListVPCV2Action struct {
	ctx     context.Context
	cloud   *cmproto.Cloud
	account *cmproto.CloudAccount
	model   store.ClusterManagerModel
	req     *cmproto.ListCloudVPCV2Request
	resp    *cmproto.ListCloudVPCV2Response
	vpcs    []*cmproto.CloudVPC
}

// NewListVPCV2Action create list action for vpcs
func NewListVPCV2Action(model store.ClusterManagerModel) *ListVPCV2Action {
	return &ListVPCV2Action{
		model: model,
	}
}

func (la *ListVPCV2Action) validate() error {
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

	err = validate.ListCloudVPCV2Validate(la.req, la.account.Account)
	if err != nil {
		return err
	}

	return nil
}

func (la *ListVPCV2Action) getRelativeData() error {
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

func (la *ListVPCV2Action) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.vpcs
}

// ListCloudVPCs list cloud vpcs
func (la *ListVPCV2Action) ListCloudVPCs() error {
	// create vpc client with cloudProvider
	vpcMgr, err := cloudprovider.GetVPCMgr(la.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s VPCManager for list vpcs failed, %s", la.cloud.CloudProvider, err.Error())
		return err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     la.cloud,
		AccountID: la.req.AccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s list vpcs failed, %s",
			la.cloud.CloudID, la.cloud.CloudProvider, err.Error())
		return err
	}
	cmOption.Region = la.req.Region

	// get vpc list
	vpcs, err := vpcMgr.ListVPCs(la.req.VpcID, cmOption)
	if err != nil {
		return err
	}
	la.vpcs = vpcs

	return nil
}

// Handle handle list vpc
func (la *ListVPCV2Action) Handle(
	ctx context.Context, req *cmproto.ListCloudVPCV2Request, resp *cmproto.ListCloudVPCV2Response) {
	if req == nil || resp == nil {
		blog.Errorf("list vpcs failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.ListCloudVPCs(); err != nil {
		la.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
