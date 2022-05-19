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

// cloud cluster list

// ListCloudClusterAction action for get cloud region clusters
type ListCloudClusterAction struct {
	ctx   context.Context
	model store.ClusterManagerModel

	cloud       *cmproto.Cloud
	account     *cmproto.CloudAccount
	req         *cmproto.ListCloudRegionClusterRequest
	resp        *cmproto.ListCloudRegionClusterResponse
	clusterList []*cmproto.CloudClusterInfo
}

// NewGetCloudRegionsAction create list action for cloud regions
func NewListCloudClusterAction(model store.ClusterManagerModel) *ListCloudClusterAction {
	return &ListCloudClusterAction{
		model: model,
	}
}

func (la *ListCloudClusterAction) listCloudRegions() error {
	clsMgr, err := cloudprovider.GetClusterMgr(la.cloud.CloudProvider)
	if err != nil {
		return err
	}

	clusterList, err := clsMgr.ListCluster(&cloudprovider.ListClusterOption{
		CommonOption: cloudprovider.CommonOption{
			Key:    la.account.Account.SecretID,
			Secret: la.account.Account.SecretKey,
			Region: la.req.Region,
			CommonConf: cloudprovider.CloudConf{
				CloudInternalEnable: la.cloud.ConfInfo.CloudInternalEnable,
				CloudDomain:         la.cloud.ConfInfo.CloudDomain,
				MachineDomain:       la.cloud.ConfInfo.MachineDomain,
			},
		},
	})
	if err != nil {
		return err
	}

	la.clusterList = clusterList
	return nil
}

func (la *ListCloudClusterAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.clusterList
}

func (la *ListCloudClusterAction) validate() error {
	err := la.req.Validate()
	if err != nil {
		return err
	}

	err = la.getRelativeData()
	if err != nil {
		return err
	}

	validate, err := cloudprovider.GetCloudValidateMgr(la.cloud.CloudProvider)
	if err != nil {
		return err
	}
	err = validate.ImportCloudAccountValidate(&cmproto.Account{
		SecretID:  la.account.Account.SecretID,
		SecretKey: la.account.Account.SecretKey,
	})
	if err != nil {
		return err
	}

	return nil
}

func (la *ListCloudClusterAction) getRelativeData() error {
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

// Handle handle list cloud cluster list
func (la *ListCloudClusterAction) Handle(
	ctx context.Context, req *cmproto.ListCloudRegionClusterRequest, resp *cmproto.ListCloudRegionClusterResponse) {
	if req == nil || resp == nil {
		blog.Errorf("get cloud region cluster list failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listCloudRegions(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
