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

package cluster

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// GetClusterUpgradeInfoAction action for get cluster upgrade info
type GetClusterUpgradeInfoAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.GetClusterUpgradeInfoReq
	resp  *cmproto.GetClusterUpgradeInfoResp
}

// NewGetClusterUpgradeInfoAction get clusters cluster upgrade info action
func NewGetClusterUpgradeInfoAction(model store.ClusterManagerModel) *GetClusterUpgradeInfoAction {
	return &GetClusterUpgradeInfoAction{
		model: model,
	}
}

func (ga *GetClusterUpgradeInfoAction) validate() error {
	return ga.req.Validate()
}

func (ga *GetClusterUpgradeInfoAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle get cluster upgrade info request
func (ga *GetClusterUpgradeInfoAction) Handle(ctx context.Context, req *cmproto.GetClusterUpgradeInfoReq,
	resp *cmproto.GetClusterUpgradeInfoResp) {
	if req == nil || resp == nil {
		blog.Errorf("get cluster upgrade info failed, req or resp is empty")
		return
	}
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	// check request parameter validate
	err := ga.validate()
	if err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	err = ga.getClusterUpgradeInfo()
	if err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

func (ga *GetClusterUpgradeInfoAction) getClusterUpgradeInfo() error {
	cloud, err := ga.model.GetCloud(ga.ctx, ga.req.ProviderID)
	if err != nil {
		return fmt.Errorf("get cloud failed, provider[%s], err:%s", ga.req.ProviderID, err.Error())
	}

	if cloud.ClusterManagement.AdvDays == 0 {
		cloud.ClusterManagement.AdvDays = 30

	}

	ga.resp.Data = &cmproto.ClusterUpgradeInfo{
		Version: cloud.ClusterManagement.UpgradeVersion,
		AdvDays: cloud.ClusterManagement.AdvDays,
	}

	return nil
}
