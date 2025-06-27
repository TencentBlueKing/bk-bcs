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

package thirdparty

import (
	"context"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tenant"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/nodeman"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// ListBKCloudAction action for list bk cloud
type ListBKCloudAction struct {
	ctx  context.Context
	req  *cmproto.ListBKCloudRequest
	resp *cmproto.CommonListResp
}

// NewListBKCloudAction create action
func NewListBKCloudAction() *ListBKCloudAction {
	return &ListBKCloudAction{}
}

func (la *ListBKCloudAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}

	return nil
}

func (la *ListBKCloudAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (la *ListBKCloudAction) listBKCloud() error {
	user := auth.GetAuthAndTenantInfoFromCtx(la.ctx)
	ctx := tenant.WithTenantIdFromContext(la.ctx, user.ResourceTenantId)

	cli := nodeman.GetNodeManClient()
	clouds, err := cli.CloudList(ctx)
	if err != nil {
		blog.Errorf("list bk cloud failed, err %s", err.Error())
		return err
	}

	result, err := utils.MarshalInterfaceToListValue(clouds)
	if err != nil {
		blog.Errorf("marshal clouds err, %s", err.Error())
		la.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
		return err
	}
	la.resp.Data = result
	return nil
}

// Handle handles list bk cloud
func (la *ListBKCloudAction) Handle(ctx context.Context, req *cmproto.ListBKCloudRequest,
	resp *cmproto.CommonListResp) {
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.listBKCloud(); err != nil {
		la.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
		return
	}

	blog.Infof("list bk cloud successfully")
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
