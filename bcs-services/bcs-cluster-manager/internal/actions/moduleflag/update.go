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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// UpdateAction update action for cloud account
type UpdateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel

	req  *cmproto.UpdateCloudModuleFlagRequest
	resp *cmproto.UpdateCloudModuleFlagResponse
}

// NewUpdateAction create update action for cloud account
func NewUpdateAction(model store.ClusterManagerModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

func (ua *UpdateAction) updateModuleFlagList() error { // nolint
	for i := range ua.req.FlagList {
		// try to get original cloud module flag for return
		flag, err := ua.model.GetCloudModuleFlag(ua.ctx, ua.req.CloudID, ua.req.Version, ua.req.ModuleID,
			ua.req.FlagList[i].FlagName)
		if err != nil {
			blog.Errorf("Get CloudModuleFlag %s:%s in pre-update failed, err %s", ua.req.CloudID,
				ua.req.ModuleID, err.Error())
			continue
		}

		err = ua.updateCloudModuleFlag(flag, ua.req.FlagList[i])
		if err != nil {
			blog.Errorf("updateCloudModuleFlag %s:%s failed, err %s", ua.req.CloudID, ua.req.ModuleID,
				err.Error())
			continue
		}
	}

	return nil
}

func (ua *UpdateAction) updateCloudModuleFlag(
	destModuleFlag *cmproto.CloudModuleFlag, flagInfo *cmproto.FlagInfo) error {
	timeStr := time.Now().Format(time.RFC3339)
	destModuleFlag.UpdateTime = timeStr
	destModuleFlag.Updater = ua.req.Operator

	if len(flagInfo.DefaultValue) != 0 {
		destModuleFlag.DefaultValue = flagInfo.DefaultValue
	}
	if len(flagInfo.FlagDesc) != 0 {
		destModuleFlag.FlagDesc = flagInfo.FlagDesc
	}
	if flagInfo.Enable != nil {
		destModuleFlag.Enable = flagInfo.Enable.GetValue()
	}
	if len(flagInfo.FlagType) != 0 {
		destModuleFlag.FlagType = flagInfo.FlagType
	}
	if len(flagInfo.FlagValueList) > 0 {
		destModuleFlag.FlagValueList = flagInfo.FlagValueList
	}

	if destModuleFlag.Regex == nil {
		destModuleFlag.Regex = &cmproto.ValueRegex{}
	}
	if flagInfo.Regex != nil && len(flagInfo.Regex.Validator) > 0 {
		destModuleFlag.Regex.Validator = flagInfo.Regex.Validator
	}
	if flagInfo.Regex != nil && len(flagInfo.Regex.Message) > 0 {
		destModuleFlag.Regex.Message = flagInfo.Regex.Message
	}

	if destModuleFlag.Range == nil {
		destModuleFlag.Range = &cmproto.NumberRange{}
	}
	if flagInfo.Range != nil {
		if flagInfo.Range.Min > 0 {
			destModuleFlag.Range.Min = flagInfo.Range.Min
		}
		if flagInfo.Range.Max > 0 {
			destModuleFlag.Range.Max = flagInfo.Range.Max
		}
	}
	if flagInfo.NetworkType != "" {
		destModuleFlag.NetworkType = flagInfo.NetworkType
	}

	return ua.model.UpdateCloudModuleFlag(ua.ctx, destModuleFlag)
}

func (ua *UpdateAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
}

// Handle handle update cloud module flag
func (ua *UpdateAction) Handle(
	ctx context.Context, req *cmproto.UpdateCloudModuleFlagRequest, resp *cmproto.UpdateCloudModuleFlagResponse) {

	if req == nil || resp == nil {
		blog.Errorf("update cloudModuleFlag failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := req.Validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ua.updateModuleFlagList(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
