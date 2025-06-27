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

package nodetemplate

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	// "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// UpdateAction update action for node template
type UpdateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.UpdateNodeTemplateRequest
	resp  *cmproto.UpdateNodeTemplateResponse

	project *project.Project
}

// NewUpdateAction create update action for nodeTemplate
func NewUpdateAction(model store.ClusterManagerModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

// NOCC:CCN_threshold(工具误报:),golint/fnsize(设计如此:)
func (ua *UpdateAction) updateNodeTemplate(destNodeTemplate *cmproto.NodeTemplate) error {
	timeStr := time.Now().Format(time.RFC3339)
	destNodeTemplate.UpdateTime = timeStr
	destNodeTemplate.Updater = ua.req.Updater

	if len(ua.req.Name) != 0 {
		destNodeTemplate.Name = ua.req.Name
	}
	if len(ua.req.Desc) != 0 {
		destNodeTemplate.Desc = ua.req.Desc
	}
	if len(ua.req.Labels) != 0 {
		destNodeTemplate.Labels = ua.req.Labels
	}
	if len(ua.req.Taints) != 0 {
		destNodeTemplate.Taints = ua.req.Taints
	}
	if len(ua.req.DockerGraphPath) != 0 {
		destNodeTemplate.DockerGraphPath = ua.req.DockerGraphPath
	}
	if len(ua.req.MountTarget) != 0 {
		destNodeTemplate.MountTarget = ua.req.MountTarget
	}
	if len(ua.req.UserScript) > 0 {
		afterScript := base64.StdEncoding.EncodeToString([]byte(ua.req.UserScript))
		destNodeTemplate.UserScript = afterScript
	} else {
		destNodeTemplate.UserScript = ua.req.UserScript
	}
	if ua.req.UnSchedulable != nil {
		destNodeTemplate.UnSchedulable = ua.req.UnSchedulable.GetValue()
	}
	if len(ua.req.DataDisks) > 0 {
		destNodeTemplate.DataDisks = ua.req.DataDisks
	}
	if len(ua.req.ExtraArgs) > 0 {
		destNodeTemplate.ExtraArgs = ua.req.ExtraArgs
	}
	if len(ua.req.PreStartUserScript) > 0 {
		preScript := base64.StdEncoding.EncodeToString([]byte(ua.req.PreStartUserScript))
		destNodeTemplate.PreStartUserScript = preScript
	} else {
		destNodeTemplate.PreStartUserScript = ua.req.PreStartUserScript
	}
	if ua.req.ScaleOutExtraAddons != nil {
		err := checkExtraActionAddons(ua.req.ScaleOutExtraAddons)
		if err != nil {
			return err
		}
		destNodeTemplate.ScaleOutExtraAddons = ua.req.ScaleOutExtraAddons
	}
	if ua.req.ScaleInExtraAddons != nil {
		err := checkExtraActionAddons(ua.req.ScaleInExtraAddons)
		if err != nil {
			return err
		}
		destNodeTemplate.ScaleInExtraAddons = ua.req.ScaleInExtraAddons
	}
	if len(ua.req.NodeOS) > 0 {
		destNodeTemplate.NodeOS = ua.req.NodeOS
	}
	if ua.req.Runtime != nil {
		destNodeTemplate.Runtime = ua.req.Runtime
	}
	if ua.req.Module != nil {
		destNodeTemplate.Module = ua.req.Module
	}
	if len(ua.req.Desc) > 0 {
		destNodeTemplate.Desc = ua.req.Desc
	}
	if ua.req.ScaleInPreScript != nil {
		destNodeTemplate.ScaleInPreScript = ua.req.ScaleInPreScript.String()
	}
	if ua.req.ScaleInPostScript != nil {
		destNodeTemplate.ScaleInPostScript = ua.req.ScaleInPostScript.String()
	}
	if ua.req.Annotations != nil {
		destNodeTemplate.Annotations = ua.req.Annotations.Values
	}
	if ua.req.GetImageInfo() != nil {
		destNodeTemplate.Image = ua.req.GetImageInfo()
	}
	if ua.req.GetGpuArgs() != nil {
		destNodeTemplate.GpuArgs = ua.req.GetGpuArgs()
	}
	if ua.req.GetExtraInfo() != nil && ua.req.GetExtraInfo().GetValues() != nil {
		destNodeTemplate.ExtraInfo = ua.req.GetExtraInfo().GetValues()
	}

	return ua.model.UpdateNodeTemplate(ua.ctx, destNodeTemplate)
}

func (ua *UpdateAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ua *UpdateAction) checkNodeTemplateAction() error {
	if ua.req.ScaleOutExtraAddons != nil {
		return checkExtraActionAddons(ua.req.ScaleOutExtraAddons)
	}
	if ua.req.ScaleInExtraAddons != nil {
		return checkExtraActionAddons(ua.req.ScaleInExtraAddons)
	}

	return nil
}

func (ua *UpdateAction) validate() error {
	err := ua.req.Validate()
	if err != nil {
		return err
	}

	err = ua.checkNodeTemplateAction()
	if err != nil {
		return err
	}

	proInfo, errLocal := project.GetProjectManagerClient().GetProjectInfo(ua.ctx, ua.req.ProjectID, true)
	if errLocal == nil {
		ua.project = proInfo
	}
	return nil
}

// Handle handle update nodeTemplate
func (ua *UpdateAction) Handle(
	ctx context.Context, req *cmproto.UpdateNodeTemplateRequest, resp *cmproto.UpdateNodeTemplateResponse) {

	if req == nil || resp == nil {
		blog.Errorf("update nodeTemplate failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	destNodeTemplate, err := ua.model.GetNodeTemplate(ua.ctx, req.ProjectID, req.NodeTemplateID)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("find nodeTemplate %s failed when pre-update checking, err %s", req.NodeTemplateID, err.Error())
		return
	}
	if err = ua.updateNodeTemplate(destNodeTemplate); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err = ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.Project.String(),
		ResourceID:   req.ProjectID,
		TaskID:       "",
		Message:      fmt.Sprintf("项目[%s]更新节点模版信息[%s]", ua.project.GetName(), req.NodeTemplateID),
		OpUser:       auth.GetUserFromCtx(ctx),
		CreateTime:   time.Now().Format(time.RFC3339),
		ProjectID:    req.ProjectID,
		ResourceName: ua.project.GetName(),
	})
	if err != nil {
		blog.Errorf("UpdateNodeTemplate[%s] CreateOperationLog failed: %v", req.ProjectID, err)
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
