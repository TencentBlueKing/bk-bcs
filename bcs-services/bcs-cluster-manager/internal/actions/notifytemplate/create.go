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

// Package notifytemplate xxx
package notifytemplate

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	autils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// CreateAction action for create notifyTemplate
type CreateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.CreateNotifyTemplateRequest
	resp  *cmproto.CreateNotifyTemplateResponse
}

// NewCreateAction create notifyTemplate action
func NewCreateAction(model store.ClusterManagerModel) *CreateAction {
	return &CreateAction{
		model: model,
	}
}

func (ca *CreateAction) createNotifyTemplate() error {
	timeStr := time.Now().Format(time.RFC3339)
	templateID := autils.GenerateTemplateID(autils.NotifyTemplate)

	notifyTemplate := &cmproto.NotifyTemplate{
		NotifyTemplateID:  templateID,
		Name:              ca.req.Name,
		ProjectID:         ca.req.ProjectID,
		NotifyType:        ca.req.NotifyType,
		Desc:              ca.req.Desc,
		Enable:            ca.req.Enable.GetValue(),
		Config:            ca.req.Config,
		CreateCluster:     ca.req.CreateCluster,
		DeleteCluster:     ca.req.DeleteCluster,
		CreateNodeGroup:   ca.req.CreateNodeGroup,
		DeleteNodeGroup:   ca.req.DeleteNodeGroup,
		UpdateNodeGroup:   ca.req.UpdateNodeGroup,
		GroupScaleOutNode: ca.req.GroupScaleOutNode,
		GroupScaleInNode:  ca.req.GroupScaleInNode,
		Receivers:         ca.req.Receivers,
		Creator:           ca.req.Creator,
		Updater:           ca.req.Creator,
		CreateTime:        timeStr,
		UpdateTime:        timeStr,
	}

	return ca.model.CreateNotifyTemplate(ca.ctx, notifyTemplate)
}

func (ca *CreateAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle create cloud notifyTemplate request
func (ca *CreateAction) Handle(ctx context.Context,
	req *cmproto.CreateNotifyTemplateRequest, resp *cmproto.CreateNotifyTemplateResponse) {
	if req == nil || resp == nil {
		blog.Errorf("create notifyTemplate failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	if err := ca.validate(); err != nil {
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ca.createNotifyTemplate(); err != nil {
		if errors.Is(err, drivers.ErrTableRecordDuplicateKey) {
			ca.setResp(common.BcsErrClusterManagerDatabaseRecordDuplicateKey, err.Error())
			return
		}
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

func (ca *CreateAction) validate() error {
	err := ca.req.Validate()
	if err != nil {
		return err
	}
	err = ca.checkNotifyTemplateName()
	if err != nil {
		return err
	}

	return nil
}

func (ca *CreateAction) checkNotifyTemplateName() error {
	templates, err := getAllNotifyTemplates(ca.ctx, ca.model, ca.req.ProjectID)
	if err != nil {
		return err
	}

	for i := range templates {
		if ca.req.Name == templates[i].Name {
			return fmt.Errorf("project[%s] NotifyTemplate[%s] duplicate", ca.req.ProjectID, ca.req.Name)
		}
	}

	return nil
}
