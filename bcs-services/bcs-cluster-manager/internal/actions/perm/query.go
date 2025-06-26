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

// Package perm xxx
package perm

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	structpb "github.com/golang/protobuf/ptypes/struct"
	spb "google.golang.org/protobuf/types/known/structpb"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
)

// CheckUserPermAction action for user perm
type CheckUserPermAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	iam   iam.PermClient

	req  *cmproto.QueryPermByActionIDRequest
	resp *cmproto.QueryPermByActionIDResponse
}

// NewCheckUserPermAction check user perm
func NewCheckUserPermAction(model store.ClusterManagerModel, iam iam.PermClient) *CheckUserPermAction {
	return &CheckUserPermAction{
		model: model,
		iam:   iam,
	}
}

func (check *CheckUserPermAction) setResp(code uint32, allow bool, url string) {
	check.resp.Code = code
	if check.resp.Data == nil {
		check.resp.Data = &cmproto.Perms{}
	}

	check.resp.Data.Perms = check.transToPerms(allow, url)
}

func (check *CheckUserPermAction) validate() error {
	err := check.req.Validate()
	if err != nil {
		return err
	}

	if check.req.PermCtx == nil || check.req.PermCtx.Operator == "" {
		return fmt.Errorf("CheckUserPermAction failed: invalid body")
	}

	return nil
}

func (check *CheckUserPermAction) transToPerms(allow bool, url string) map[string]*structpb.Value {
	perms := make(map[string]*structpb.Value, 0)
	if url == "" {
		url = options.GetGlobalCMOptions().IAM.ApplyPermAddress
	}
	perms[check.req.ActionID] = spb.NewBoolValue(allow)
	perms[applyURL] = spb.NewStringValue(url)

	return perms
}

func (check *CheckUserPermAction) checkUserPermByActionID() error {
	user := auth.GetAuthAndTenantInfoFromCtx(check.ctx)
	allow, applyUrl, err := CheckUserPermByActionID(check.ctx, check.iam, UserActionPerm{
		User: authutils.UserInfo{
			BkUserName: check.req.PermCtx.GetOperator(),
			TenantId:   user.ResourceTenantId,
		},
		ActionID: check.req.GetActionID(),
		RelatedPermResource: RelatedPermResource{
			ProjectID:      check.req.PermCtx.GetProjectId(),
			ClusterID:      check.req.PermCtx.GetClusterId(),
			Namespace:      check.req.PermCtx.GetName(),
			TemplateID:     check.req.PermCtx.GetTemplateId(),
			CloudAccountID: check.req.PermCtx.GetAccountId(),
		},
	})
	if err != nil {
		blog.Errorf("CheckUserPermByActionID failed: %v", err)
		return err
	}

	check.setResp(common.BcsErrClusterManagerSuccess, allow, applyUrl)
	return nil
}

// Handle handle list cluster account list
func (check *CheckUserPermAction) Handle(
	ctx context.Context, req *cmproto.QueryPermByActionIDRequest, resp *cmproto.QueryPermByActionIDResponse) {
	if req == nil || resp == nil {
		blog.Errorf("CheckUserPermAction queryPermByActionID failed, req or resp is empty")
		return
	}
	check.ctx = ctx
	check.req = req
	check.resp = resp

	if err := check.validate(); err != nil {
		check.setResp(common.BcsErrClusterManagerInvalidParameter, false, "")
		return
	}
	if err := check.checkUserPermByActionID(); err != nil {
		check.setResp(common.BcsErrClusterManagerGetPermErr, false, "")
		return
	}
}
