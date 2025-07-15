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
	//"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	autils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	iauth "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// CreateAction action for create templateConfig
type CreateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.CreateTemplateConfigRequest
	resp  *cmproto.CreateTemplateConfigResponse

	project *project.Project
}

// NewCreateAction create templateConfig action
func NewCreateAction(model store.ClusterManagerModel) *CreateAction {
	return &CreateAction{
		model: model,
	}
}

// createTemplateConfig create templateConfig
func (ca *CreateAction) createTemplateConfig() error {
	timeStr := time.Now().Format(time.RFC3339)
	templateID := autils.GenerateTemplateID(autils.TemplateConfig)

	templateConfigContent, err := json.Marshal(ca.req.CloudTemplateConfig)
	if err != nil {
		return fmt.Errorf("marshal templateConfig[%s] failed, err: %s",
			ca.req.CloudTemplateConfig, err.Error())
	}

	username := iauth.GetUserFromCtx(ca.ctx)

	templateConfig := &cmproto.TemplateConfig{
		TemplateConfigID: templateID,
		BusinessID:       ca.req.BusinessID,
		ProjectID:        ca.req.ProjectID,
		ClusterID:        ca.req.ClusterID,
		Provider:         ca.req.Provider,
		ConfigType:       ca.req.ConfigType,
		ConfigContent:    string(templateConfigContent),
		Creator:          username,
		Updater:          username,
		CreateTime:       timeStr,
		UpdateTime:       timeStr,
	}

	return ca.model.CreateTemplateConfig(ca.ctx, templateConfig)
}

// setResp set resp info
func (ca *CreateAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle create cloud templateConfig request
func (ca *CreateAction) Handle(ctx context.Context,
	req *cmproto.CreateTemplateConfigRequest, resp *cmproto.CreateTemplateConfigResponse) {
	if req == nil || resp == nil {
		blog.Errorf("create templateConfig failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	if err := ca.validate(); err != nil {
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ca.createTemplateConfig(); err != nil {
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
		Message:      fmt.Sprintf("业务[%s]在[%s]创建[%s]类型模板配置", req.BusinessID, req.Provider, req.ConfigType),
		OpUser:       auth.GetUserFromCtx(ctx),
		CreateTime:   time.Now().Format(time.RFC3339),
		ResourceName: ca.project.Name,
		ProjectID:    ca.project.ProjectID,
		ClusterID:    ca.req.ClusterID,
	})
	if err != nil {
		blog.Errorf("CreateTemplateConfig[%s] CreateOperationLog failed: %v", req.ProjectID, err)
	}

	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// validate validate params
func (ca *CreateAction) validate() error {
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

func (ca *CreateAction) checkTemplateConfig() error {
	configInfos, err := getTemplateConfigInfos(ca.ctx, ca.model, ca.req.BusinessID, ca.req.ProjectID, ca.req.ClusterID,
		ca.req.Provider, ca.req.ConfigType, nil)
	if err != nil {
		return err
	}

	if len(configInfos) > 0 {
		return fmt.Errorf("templateConfig businessID[%s] projectID[%s] clusterID[%s] "+
			"in provider[%s] configType[%s] already exists",
			ca.req.BusinessID, ca.req.ProjectID, ca.req.ClusterID, ca.req.Provider, ca.req.ConfigType)
	}

	return nil
}
