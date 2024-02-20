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

// ListSecurityGroupsAction action for list security groups
type ListSecurityGroupsAction struct {
	ctx            context.Context
	cloud          *cmproto.Cloud
	account        *cmproto.CloudAccount
	model          store.ClusterManagerModel
	req            *cmproto.ListCloudSecurityGroupsRequest
	resp           *cmproto.ListCloudSecurityGroupsResponse
	securityGroups []*cmproto.SecurityGroup
}

// NewListSecurityGroupsAction create list action for security groups
func NewListSecurityGroupsAction(model store.ClusterManagerModel) *ListSecurityGroupsAction {
	return &ListSecurityGroupsAction{
		model: model,
	}
}

func (la *ListSecurityGroupsAction) validate() error {
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

	err = validate.ListSecurityGroupsValidate(la.req, func() *cmproto.Account {
		if la.account == nil || la.account.Account == nil {
			return nil
		}

		return la.account.Account
	}())
	if err != nil {
		return err
	}

	return nil
}

func (la *ListSecurityGroupsAction) getRelativeData() error {
	cloud, err := actions.GetCloudByCloudID(la.model, la.req.CloudID)
	if err != nil {
		return err
	}
	la.cloud = cloud

	if len(la.req.AccountID) > 0 {
		account, err := la.model.GetCloudAccount(la.ctx, la.req.CloudID, la.req.AccountID, false)
		if err != nil {
			return err
		}

		la.account = account
	}

	return nil
}

func (la *ListSecurityGroupsAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.securityGroups
}

func (la *ListSecurityGroupsAction) listCloudSecurityGroups() error {
	// create vpc client with cloudProvider
	vpcMgr, err := cloudprovider.GetVPCMgr(la.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s VPCManager for list SecurityGroups failed, %s", la.cloud.CloudProvider, err.Error())
		return err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     la.cloud,
		AccountID: la.req.AccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s list SecurityGroups failed, %s",
			la.cloud.CloudID, la.cloud.CloudProvider, err.Error())
		return err
	}
	cmOption.Region = la.req.Region

	// get sgs list
	sgs, err := vpcMgr.ListSecurityGroups(&cloudprovider.ListNetworksOption{
		CommonOption:      *cmOption,
		ResourceGroupName: la.req.ResourceGroupName,
	})
	if err != nil {
		return err
	}
	la.securityGroups = sgs

	return nil
}

// Handle handle list security groups
func (la *ListSecurityGroupsAction) Handle(
	ctx context.Context, req *cmproto.ListCloudSecurityGroupsRequest, resp *cmproto.ListCloudSecurityGroupsResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list security groups failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.listCloudSecurityGroups(); err != nil {
		la.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
