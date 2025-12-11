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

// ListCloudVpcsPageAction action for list cloud vpc page
type ListCloudVpcsPageAction struct {
	ctx context.Context

	cloud   *cmproto.Cloud
	account *cmproto.CloudAccount

	model store.ClusterManagerModel
	req   *cmproto.ListCloudVpcsPageRequest
	resp  *cmproto.ListCloudVpcsPageResponse

	vpcs []*cmproto.CloudVpcs
}

// NewListCloudVpcsPageAction create list action for cloud vpcs page
func NewListCloudVpcsPageAction(model store.ClusterManagerModel) *ListCloudVpcsPageAction {
	return &ListCloudVpcsPageAction{
		model: model,
	}
}

func (la *ListCloudVpcsPageAction) validate() error {
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
		req := &cmproto.ListCloudVpcsRequest{
			CloudID:           la.req.CloudID,
			AccountID:         la.req.AccountID,
			ResourceGroupName: la.req.ResourceGroupName,
			Region:            la.req.Region,
			VpcID:             la.req.VpcID,
		}
		err = validate.ListCloudVpcsValidate(req, la.account.Account)
		if err != nil {
			return err
		}
	}

	return nil
}

func (la *ListCloudVpcsPageAction) getRelativeData() error {
	cloud, err := actions.GetCloudByCloudID(la.model, la.req.CloudID)
	if err != nil {
		return err
	}
	la.cloud = cloud

	if la.req.AccountID != "" {
		account, err := la.model.GetCloudAccount(la.ctx, la.req.CloudID, la.req.AccountID, false)
		if err != nil {
			return err
		}

		la.account = account
	}

	return nil
}

func (la *ListCloudVpcsPageAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.vpcs
}

// ListCloudVpcs list cloud vpcs
func (la *ListCloudVpcsPageAction) ListCloudVpcs() error {
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
	vpcPage := cloudprovider.VpcPage{
		CloudId: la.req.CloudID,
		VpcIds:  []string{},
		VpcName: []string{},
		Offset:  la.req.Offset,
		Limit:   la.req.Limit,
	}
	if la.req.VpcID != "" {
		vpcPage.VpcIds = append(vpcPage.VpcIds, la.req.VpcID)
	}
	if la.req.VpcName != "" {
		vpcPage.VpcName = append(vpcPage.VpcName, la.req.VpcName)
	}
	if la.req.Limit == 0 {
		vpcPage.Limit = 10
	}

	// get vpc list
	total, vpcs, err := vpcMgr.ListVpcsByPage(&cloudprovider.ListNetworksOption{
		CommonOption:      *cmOption,
		ResourceGroupName: la.req.ResourceGroupName,
		VpcPage:           vpcPage,
	})
	if err != nil {
		return err
	}
	la.vpcs = vpcs
	la.resp.Total = uint32(total)

	return nil
}

// Handle list cloud vpcs
func (la *ListCloudVpcsPageAction) Handle(
	ctx context.Context, req *cmproto.ListCloudVpcsPageRequest, resp *cmproto.ListCloudVpcsPageResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list cloud vpcs page failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.ListCloudVpcs(); err != nil {
		la.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
