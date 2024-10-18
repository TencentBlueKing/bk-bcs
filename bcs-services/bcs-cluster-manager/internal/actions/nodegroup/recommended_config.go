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

package nodegroup

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// RecommendNodeGroupConfAction action for get cluster
type RecommendNodeGroupConfAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	req     *cmproto.RecommendNodeGroupConfReq
	resp    *cmproto.RecommendNodeGroupConfResp
	cloud   *cmproto.Cloud
	configs []*cmproto.RecommendNodeGroupConf
}

// NewRecommendNodeGroupConfAction create get action
func NewRecommendNodeGroupConfAction(model store.ClusterManagerModel) *RecommendNodeGroupConfAction {
	return &RecommendNodeGroupConfAction{
		model: model,
	}
}

func (ra *RecommendNodeGroupConfAction) validate() error { // nolint
	return ra.req.Validate()
}

func (ra *RecommendNodeGroupConfAction) setResp(code uint32, msg string) {
	ra.resp.Code = code
	ra.resp.Message = msg
	ra.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	ra.resp.Data = ra.configs
}

// Handle get cluster request
func (ra *RecommendNodeGroupConfAction) Handle(ctx context.Context, req *cmproto.RecommendNodeGroupConfReq,
	resp *cmproto.RecommendNodeGroupConfResp) {
	if req == nil || resp == nil {
		blog.Errorf("get recommendedNodeGroupConf failed, req or resp is empty")
		return
	}
	ra.ctx = ctx
	ra.req = req
	ra.resp = resp
	var err error

	if err = req.Validate(); err != nil {
		ra.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if ra.cloud, err = actions.GetCloudByCloudID(ra.model, req.CloudID); err != nil {
		ra.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ra.cloud,
		AccountID: ra.req.AccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloud provider %s/%s failed, %s",
			ra.cloud.CloudID, ra.cloud.CloudProvider, err.Error())
		ra.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	cmOption.Region = ra.req.Region

	mgr, err := cloudprovider.GetNodeGroupMgr(ra.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get nodegroup manager for cloud provider %s failed, %s", ra.cloud.CloudProvider, err.Error())
		ra.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	configs, err := mgr.RecommendNodeGroupConf(cmOption)
	if err != nil {
		ra.setResp(common.BcsErrClusterManagerCheckKubeConnErr, err.Error())
		return
	}
	ra.configs = configs
	ra.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
