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

package account

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// VerifyAction action for verify cloud account
type VerifyAction struct {
	ctx   context.Context
	model store.ClusterManagerModel

	cloud *cmproto.Cloud
	req   *cmproto.VerifyCloudAccountRequest
	resp  *cmproto.VerifyCloudAccountResponse
}

// NewVerifyAction verify cloud account action
func NewVerifyAction(model store.ClusterManagerModel) *VerifyAction {
	return &VerifyAction{
		model: model,
	}
}

func (va *VerifyAction) setResp(code uint32, msg string) {
	va.resp.Code = code
	va.resp.Message = msg
	va.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (va *VerifyAction) validate() error {
	err := va.req.Validate()
	if err != nil {
		return err
	}

	va.cloud, err = actions.GetCloudByCloudID(va.model, va.req.CloudID)
	if err != nil {
		return err
	}

	cloudValidateMgr, err := cloudprovider.GetCloudValidateMgr(va.cloud.CloudProvider)
	if err != nil {
		return err
	}
	err = cloudValidateMgr.CreateCloudAccountValidate(va.req.Account)
	if err != nil {
		return err
	}

	return nil
}

// Handle verify cloud account
func (va *VerifyAction) Handle(
	ctx context.Context, req *cmproto.VerifyCloudAccountRequest, resp *cmproto.VerifyCloudAccountResponse) {
	if req == nil || resp == nil {
		blog.Errorf("verify cloudAccount failed, req or resp is empty")
		return
	}
	va.ctx = ctx
	va.req = req
	va.resp = resp

	if err := va.validate(); err != nil {
		va.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	va.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
