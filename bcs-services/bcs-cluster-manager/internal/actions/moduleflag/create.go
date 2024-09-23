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

// Package moduleflag xxx
package moduleflag

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// CreateAction action for create cloud moduleFlag
type CreateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel

	moduleFlagList []*cmproto.CloudModuleFlag
	req            *cmproto.CreateCloudModuleFlagRequest
	resp           *cmproto.CreateCloudModuleFlagResponse
}

// NewCreateAction create cloudModuleFlag action
func NewCreateAction(model store.ClusterManagerModel) *CreateAction {
	return &CreateAction{
		model: model,
	}
}

func (ca *CreateAction) importCloudModuleFlagList() error {
	moduleFlagList := make([]*cmproto.CloudModuleFlag, 0)

	timeStr := time.Now().Format(time.RFC3339)
	for i := range ca.req.FlagList {
		moduleFlagList = append(moduleFlagList, &cmproto.CloudModuleFlag{
			CloudID:       ca.req.CloudID,
			Version:       ca.req.Version,
			ModuleID:      ca.req.ModuleID,
			FlagName:      ca.req.FlagList[i].FlagName,
			FlagDesc:      ca.req.FlagList[i].FlagDesc,
			DefaultValue:  ca.req.FlagList[i].DefaultValue,
			Enable:        ca.req.FlagList[i].Enable.GetValue(),
			Creator:       ca.req.Operator,
			Updater:       ca.req.Operator,
			CreatTime:     timeStr,
			UpdateTime:    timeStr,
			FlagType:      ca.req.FlagList[i].FlagType,
			FlagValueList: ca.req.FlagList[i].FlagValueList,
			Regex: func() *cmproto.ValueRegex {
				if ca.req.FlagList[i].GetRegex() != nil {
					return &cmproto.ValueRegex{
						Validator: ca.req.FlagList[i].GetRegex().Validator,
						Message:   ca.req.FlagList[i].GetRegex().Message,
					}
				}
				return nil
			}(),
			Range: func() *cmproto.NumberRange {
				if ca.req.FlagList[i].GetRange() != nil {
					return &cmproto.NumberRange{
						Min: ca.req.FlagList[i].GetRange().Min,
						Max: ca.req.FlagList[i].GetRange().Max,
					}
				}
				return nil
			}(),
			NetworkType: ca.req.FlagList[i].GetNetworkType(),
		})
	}
	ca.moduleFlagList = moduleFlagList

	for i := range moduleFlagList {
		err := ca.model.CreateCloudModuleFlag(ca.ctx, moduleFlagList[i])
		if err != nil {
			if errors.Is(err, drivers.ErrTableRecordDuplicateKey) {
				continue
			}
			return err
		}
		blog.Infof("cloud[%s] version[%s] module[%s] flag[%s] import successful",
			moduleFlagList[i].CloudID, moduleFlagList[i].Version, moduleFlagList[i].ModuleID, moduleFlagList[i].FlagName)
	}

	return nil
}

func (ca *CreateAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
}

// Handle create cloud moduleFlag request
func (ca *CreateAction) Handle(ctx context.Context,
	req *cmproto.CreateCloudModuleFlagRequest, resp *cmproto.CreateCloudModuleFlagResponse) {
	if req == nil || resp == nil {
		blog.Errorf("create cloudModuleFlag failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	if err := ca.validate(); err != nil {
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ca.importCloudModuleFlagList(); err != nil {
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err := ca.model.CreateOperationLog(ca.ctx, &cmproto.OperationLog{
		ResourceType: common.Cloud.String(),
		ResourceID:   ca.req.CloudID,
		TaskID:       "",
		Message:      fmt.Sprintf("录入云组件参数[%s]", ca.req.CloudID),
		OpUser:       auth.GetUserFromCtx(ca.ctx),
		CreateTime:   time.Now().Format(time.RFC3339),
	})
	if err != nil {
		blog.Errorf("CreateCloudModuleFlag[%s] CreateOperationLog failed: %v", ca.req.CloudID, err)
	}

	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

func (ca *CreateAction) validate() error {
	err := ca.req.Validate()
	if err != nil {
		return err
	}

	return nil
}
