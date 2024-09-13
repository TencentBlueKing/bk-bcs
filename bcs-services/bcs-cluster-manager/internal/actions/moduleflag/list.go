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

package moduleflag

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

// ListAction action for list cloud moduleFlagList
type ListAction struct {
	ctx   context.Context
	model store.ClusterManagerModel

	req                *cmproto.ListCloudModuleFlagRequest
	resp               *cmproto.ListCloudModuleFlagResponse
	moduleFlagListList []*cmproto.CloudModuleFlag
}

// NewListAction create list action for cluster moduleFlag
func NewListAction(model store.ClusterManagerModel) *ListAction {
	return &ListAction{
		model: model,
	}
}

func (la *ListAction) listCloudModuleFlag() error {
	conds := make([]*operator.Condition, 0)

	condM := make(operator.M)
	//! we don't setting bson tag in proto file
	//! all fields are in lowcase
	if len(la.req.CloudID) != 0 {
		condM["cloudid"] = la.req.CloudID
	}
	if len(la.req.Version) != 0 {
		condM["version"] = la.req.Version
	}
	if len(la.req.ModuleID) != 0 {
		condM["moduleid"] = la.req.ModuleID
	}
	condEqual := operator.NewLeafCondition(operator.Eq, condM)
	conds = append(conds, condEqual)

	if len(la.req.FlagNameList) > 0 {
		condIn := operator.NewLeafCondition(operator.In, operator.M{"flagname": la.req.FlagNameList})
		conds = append(conds, condIn)
	}
	cond := operator.NewBranchCondition(operator.And, conds...)

	cloudModuleFlags, err := la.model.ListCloudModuleFlag(la.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return err
	}

	// get all cloud default module flags
	flags, err := getCommonModuleFlags(la.model, "default", la.req.ModuleID)
	if err != nil {
		blog.Errorf("listCloudModuleFlag getCommonModuleFlags failed: %v", err)
	}
	cloudModuleFlags = append(cloudModuleFlags, flags...)

	for i := range cloudModuleFlags {
		cloudModuleFlags[i].FlagDesc = i18n.T(la.ctx, cloudModuleFlags[i].FlagDesc)
		la.moduleFlagListList = append(la.moduleFlagListList, &cloudModuleFlags[i])
	}

	return nil
}

func getCommonModuleFlags(model store.ClusterManagerModel, cloudID, moduleID string) (
	[]cmproto.CloudModuleFlag, error) {
	condM := make(operator.M)
	condM["cloudid"] = cloudID
	condM["moduleid"] = moduleID

	cond := operator.NewLeafCondition(operator.Eq, condM)

	return model.ListCloudModuleFlag(context.Background(), cond, &storeopt.ListOption{})
}

func (la *ListAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Data = la.moduleFlagListList
}

// Handle handle list cluster module flag list
func (la *ListAction) Handle(
	ctx context.Context, req *cmproto.ListCloudModuleFlagRequest, resp *cmproto.ListCloudModuleFlagResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list cloudAccount failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listCloudModuleFlag(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
