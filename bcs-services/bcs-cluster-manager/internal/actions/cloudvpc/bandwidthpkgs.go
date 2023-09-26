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

// GetCloudBandwidthPackagesAction action for list cloud bwps
type GetCloudBandwidthPackagesAction struct {
	ctx   context.Context
	model store.ClusterManagerModel

	cloud   *cmproto.Cloud
	account *cmproto.CloudAccount

	req  *cmproto.GetCloudBandwidthPackagesRequest
	resp *cmproto.GetCloudBandwidthPackagesResponse
	bwps []*cmproto.BandwidthPackageInfo
}

// NewGGetCloudBandwidthPackagesAction create list action for cloud bwps
func NewGGetCloudBandwidthPackagesAction(model store.ClusterManagerModel) *GetCloudBandwidthPackagesAction {
	return &GetCloudBandwidthPackagesAction{
		model: model,
	}
}

func (ga *GetCloudBandwidthPackagesAction) getCloudBwpsResource() error {
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ga.cloud,
		AccountID: ga.req.AccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s list CloudBwpsResource failed, %s",
			ga.cloud.CloudID, ga.cloud.CloudProvider, err.Error())
		return err
	}
	cmOption.Region = ga.req.Region

	// create vpc client with cloudProvider
	vpcMgr, err := cloudprovider.GetVPCMgr(ga.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s VPCManager for list CloudBwpsResource failed, %s",
			ga.cloud.CloudProvider, err.Error())
		return err
	}

	ga.bwps, err = vpcMgr.ListBandwidthPacks(cmOption)
	if err != nil {
		return err
	}

	return nil
}

func (ga *GetCloudBandwidthPackagesAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	ga.resp.Data = ga.bwps
}

func (ga *GetCloudBandwidthPackagesAction) validate() error {
	err := ga.req.Validate()
	if err != nil {
		return err
	}

	err = ga.getRelativeData()
	if err != nil {
		return err
	}

	if len(ga.req.AccountID) > 0 {
		validate, errGet := cloudprovider.GetCloudValidateMgr(ga.cloud.CloudProvider)
		if errGet != nil {
			return errGet
		}
		err = validate.ImportCloudAccountValidate(ga.account.Account)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ga *GetCloudBandwidthPackagesAction) getRelativeData() error {
	cloud, err := actions.GetCloudByCloudID(ga.model, ga.req.CloudID)
	if err != nil {
		return err
	}

	if len(ga.req.AccountID) > 0 {
		account, errGet := ga.model.GetCloudAccount(ga.ctx, ga.req.CloudID, ga.req.AccountID, false)
		if errGet != nil {
			return errGet
		}
		ga.account = account
	}

	ga.cloud = cloud
	return nil
}

// Handle list cloud account types
func (ga *GetCloudBandwidthPackagesAction) Handle(
	ctx context.Context, req *cmproto.GetCloudBandwidthPackagesRequest, resp *cmproto.GetCloudBandwidthPackagesResponse) {
	if req == nil || resp == nil {
		blog.Errorf("get cloud bwps resource failed, req or resp is empty")
		return
	}
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	if err := ga.validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ga.getCloudBwpsResource(); err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
