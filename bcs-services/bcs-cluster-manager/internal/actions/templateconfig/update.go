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

// Package templateconfig xxx
package templateconfig

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	// "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	iauth "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// UpdateAction action for update templateConfig
type UpdateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.UpdateTemplateConfigRequest
	resp  *cmproto.UpdateTemplateConfigResponse

	project *project.Project
}

// NewUpdateAction update templateConfig action
func NewUpdateAction(model store.ClusterManagerModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

// updateTemplateConfig update templateConfig
func (ca *UpdateAction) updateTemplateConfig() error {
	timeStr := time.Now().Format(time.RFC3339)

	templateConfigContent, err := json.Marshal(ca.req.CloudTemplateConfig)
	if err != nil {
		return fmt.Errorf("marshal templateConfig[%s] failed, err: %s",
			ca.req.CloudTemplateConfig, err.Error())
	}

	username := iauth.GetUserFromCtx(ca.ctx)

	templateConfig := &cmproto.TemplateConfig{
		TemplateConfigID: ca.req.TemplateConfigID,
		BusinessID:       ca.req.BusinessID,
		ProjectID:        ca.req.ProjectID,
		ClusterID:        ca.req.ClusterID,
		Provider:         ca.req.Provider,
		ConfigType:       ca.req.ConfigType,
		ConfigContent:    string(templateConfigContent),
		Updater:          username,
		UpdateTime:       timeStr,
	}

	return ca.model.UpdateTemplateConfig(ca.ctx, templateConfig)
}

// setResp set resp info
func (ca *UpdateAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle update cloud templateConfig request
func (ca *UpdateAction) Handle(ctx context.Context,
	req *cmproto.UpdateTemplateConfigRequest, resp *cmproto.UpdateTemplateConfigResponse) {
	if req == nil || resp == nil {
		blog.Errorf("update templateConfig failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	if err := ca.validate(); err != nil {
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ca.updateTemplateConfig(); err != nil {
		if errors.Is(err, drivers.ErrTableRecordDuplicateKey) {
			ca.setResp(common.BcsErrClusterManagerDatabaseRecordDuplicateKey, err.Error())
			return
		}
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err := ca.model.CreateOperationLog(ca.ctx, &cmproto.OperationLog{
		ResourceType: common.Project.String(),
		ResourceID:   req.ProjectID,
		TaskID:       "",
		Message:      fmt.Sprintf("业务[%s]在[%s]修改[%s]类型模板配置", req.BusinessID, req.Provider, req.ConfigType),
		OpUser:       auth.GetUserFromCtx(ctx),
		CreateTime:   time.Now().Format(time.RFC3339),
		ResourceName: ca.project.Name,
		ProjectID:    ca.project.ProjectID,
		ClusterID:    ca.req.ClusterID,
	})
	if err != nil {
		blog.Errorf("UpdateTemplateConfig[%s] CreateOperationLog failed: %v", req.ProjectID, err)
	}

	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// validate validate params
func (ca *UpdateAction) validate() error {
	err := ca.req.Validate()
	if err != nil {
		return err
	}

	if !checkTemplateConfigType(ca.req.ConfigType) {
		return fmt.Errorf("configType[%s] is invalid", ca.req.ConfigType)
	}

	proInfo, errLocal := project.GetProjectManagerClient().GetProjectInfo(ca.ctx, ca.req.ProjectID, true)
	if errLocal != nil {
		return errLocal
	}
	ca.project = proInfo

	err = ca.checkTemplateConfig()
	if err != nil {
		return err
	}

	return nil
}

func (ca *UpdateAction) checkTemplateConfig() error {
	configInfo, err := ca.model.GetTemplateConfigByID(ca.ctx, ca.req.TemplateConfigID)
	if err != nil {
		return err
	}

	if configInfo == nil {
		return fmt.Errorf("templateConfig templateConfigID[%s]"+
			"in provider[%s] configType[%s] not exists",
			ca.req.TemplateConfigID, ca.req.Provider, ca.req.ConfigType)
	}

	return nil
}
