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

const (
	defaultRegion = "ap-nanjing"
)

// cloud region list

// GetCloudRegionsAction action for get cloud regions
type GetCloudRegionsAction struct {
	ctx   context.Context
	model store.ClusterManagerModel

	cloud      *cmproto.Cloud
	account    *cmproto.CloudAccount
	req        *cmproto.GetCloudRegionsRequest
	resp       *cmproto.GetCloudRegionsResponse
	regionList []*cmproto.RegionInfo
}

// NewGetCloudRegionsAction create list action for cloud regions
func NewGetCloudRegionsAction(model store.ClusterManagerModel) *GetCloudRegionsAction {
	return &GetCloudRegionsAction{
		model: model,
	}
}

func (ga *GetCloudRegionsAction) listCloudRegions() error {
	nodeMgr, err := cloudprovider.GetNodeMgr(ga.cloud.CloudProvider)
	if err != nil {
		return err
	}

	regionList, err := nodeMgr.GetCloudRegions(&cloudprovider.CommonOption{
		Key:    ga.account.Account.SecretID,
		Secret: ga.account.Account.SecretKey,
		// Region trick data, cloud need underlying dependence
		Region: defaultRegion,
		CommonConf: cloudprovider.CloudConf{
			CloudInternalEnable: ga.cloud.ConfInfo.CloudInternalEnable,
			CloudDomain:         ga.cloud.ConfInfo.CloudDomain,
			MachineDomain:       ga.cloud.ConfInfo.MachineDomain,
		},
	})
	if err != nil {
		return err
	}

	ga.regionList = regionList
	return nil
}

func (ga *GetCloudRegionsAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	ga.resp.Data = ga.regionList
}

func (ga *GetCloudRegionsAction) validate() error {
	err := ga.req.Validate()
	if err != nil {
		return err
	}

	err = ga.getRelativeData()
	if err != nil {
		return err
	}

	validate, err := cloudprovider.GetCloudValidateMgr(ga.cloud.CloudProvider)
	if err != nil {
		return err
	}
	err = validate.ImportCloudAccountValidate(&cmproto.Account{
		SecretID:  ga.account.Account.SecretID,
		SecretKey: ga.account.Account.SecretKey,
	})
	if err != nil {
		return err
	}

	return nil
}

func (ga *GetCloudRegionsAction) getRelativeData() error {
	cloud, err := actions.GetCloudByCloudID(ga.model, ga.req.CloudID)
	if err != nil {
		return err
	}
	account, err := ga.model.GetCloudAccount(ga.ctx, ga.req.CloudID, ga.req.AccountID)
	if err != nil {
		return err
	}

	ga.account = account
	ga.cloud = cloud
	return nil
}

// Handle handle list cloud regions
func (ga *GetCloudRegionsAction) Handle(
	ctx context.Context, req *cmproto.GetCloudRegionsRequest, resp *cmproto.GetCloudRegionsResponse) {
	if req == nil || resp == nil {
		blog.Errorf("get cloud region list failed, req or resp is empty")
		return
	}
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	if err := ga.validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ga.listCloudRegions(); err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

// cloud region zone list

// GetCloudRegionZonesAction action for get cloud region zones
type GetCloudRegionZonesAction struct {
	ctx   context.Context
	model store.ClusterManagerModel

	cloud    *cmproto.Cloud
	account  *cmproto.CloudAccount
	req      *cmproto.GetCloudRegionZonesRequest
	resp     *cmproto.GetCloudRegionZonesResponse
	zoneList []*cmproto.ZoneInfo
}

// NewGetCloudRegionZonesAction create list action for cloud region zones
func NewGetCloudRegionZonesAction(model store.ClusterManagerModel) *GetCloudRegionZonesAction {
	return &GetCloudRegionZonesAction{
		model: model,
	}
}

func (ga *GetCloudRegionZonesAction) listCloudRegionZones() error {
	nodeMgr, err := cloudprovider.GetNodeMgr(ga.cloud.CloudProvider)
	if err != nil {
		return err
	}

	zoneList, err := nodeMgr.GetZoneList(&cloudprovider.CommonOption{
		Key:    ga.account.Account.SecretID,
		Secret: ga.account.Account.SecretKey,
		Region: ga.req.Region,
		CommonConf: cloudprovider.CloudConf{
			CloudInternalEnable: ga.cloud.ConfInfo.CloudInternalEnable,
			CloudDomain:         ga.cloud.ConfInfo.CloudDomain,
			MachineDomain:       ga.cloud.ConfInfo.MachineDomain,
		},
	})
	if err != nil {
		return err
	}

	ga.zoneList = zoneList
	return nil
}

func (ga *GetCloudRegionZonesAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	ga.resp.Data = ga.zoneList
}

func (ga *GetCloudRegionZonesAction) validate() error {
	err := ga.req.Validate()
	if err != nil {
		return err
	}

	err = ga.getRelativeData()
	if err != nil {
		return err
	}

	validate, err := cloudprovider.GetCloudValidateMgr(ga.cloud.CloudProvider)
	if err != nil {
		return err
	}
	err = validate.ImportCloudAccountValidate(&cmproto.Account{
		SecretID:  ga.account.Account.SecretID,
		SecretKey: ga.account.Account.SecretKey,
	})
	if err != nil {
		return err
	}

	return nil
}

func (ga *GetCloudRegionZonesAction) getRelativeData() error {
	cloud, err := actions.GetCloudByCloudID(ga.model, ga.req.CloudID)
	if err != nil {
		return err
	}
	account, err := ga.model.GetCloudAccount(ga.ctx, ga.req.CloudID, ga.req.AccountID)
	if err != nil {
		return err
	}

	ga.account = account
	ga.cloud = cloud
	return nil
}

// Handle handle list cloud regions
func (ga *GetCloudRegionZonesAction) Handle(
	ctx context.Context, req *cmproto.GetCloudRegionZonesRequest, resp *cmproto.GetCloudRegionZonesResponse) {
	if req == nil || resp == nil {
		blog.Errorf("get cloud region zone list failed, req or resp is empty")
		return
	}
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	if err := ga.validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ga.listCloudRegionZones(); err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
