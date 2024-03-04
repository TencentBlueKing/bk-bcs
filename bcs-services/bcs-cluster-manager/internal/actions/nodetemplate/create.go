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

// Package nodetemplate xxx
package nodetemplate

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	autils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/template"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// CreateAction action for create nodeTemplate
type CreateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	cloud *cmproto.Cloud // nolint
	req   *cmproto.CreateNodeTemplateRequest
	resp  *cmproto.CreateNodeTemplateResponse
}

// NewCreateAction create nodeTemplate action
func NewCreateAction(model store.ClusterManagerModel) *CreateAction {
	return &CreateAction{
		model: model,
	}
}

func (ca *CreateAction) createNodeTemplate() error {
	timeStr := time.Now().Format(time.RFC3339)
	templateID := autils.GenerateNodeTemplateID()

	// 扩容节点脚本 trans user to base64 encoding
	afterScript := ca.req.UserScript
	if len(afterScript) > 0 {
		afterScript = base64.StdEncoding.EncodeToString([]byte(ca.req.UserScript))
	}
	preScript := ca.req.PreStartUserScript
	if len(preScript) > 0 {
		preScript = base64.StdEncoding.EncodeToString([]byte(ca.req.PreStartUserScript))
	}
	// 缩容节点脚本
	scaleInPreScript := ""
	if ca.req.ScaleInPreScript != nil {
		scaleInPreScript = base64.StdEncoding.EncodeToString([]byte(ca.req.ScaleInPreScript.String()))
	}
	scaleInPostScript := ""
	if ca.req.ScaleInPreScript != nil {
		scaleInPostScript = base64.StdEncoding.EncodeToString([]byte(ca.req.ScaleInPostScript.String()))
	}

	nodeTemplate := &cmproto.NodeTemplate{
		NodeTemplateID:  templateID,
		Name:            ca.req.Name,
		ProjectID:       ca.req.ProjectID,
		Labels:          ca.req.Labels,
		Taints:          ca.req.Taints,
		DockerGraphPath: ca.req.DockerGraphPath,
		MountTarget:     ca.req.MountTarget,
		// base64格式编码后置脚本
		UserScript:    afterScript,
		UnSchedulable: ca.req.UnSchedulable,
		DataDisks:     ca.req.DataDisks,
		ExtraArgs:     ca.req.ExtraArgs,
		// base64格式编码前置脚本
		PreStartUserScript:  preScript,
		ScaleOutExtraAddons: ca.req.ScaleOutExtraAddons,
		ScaleInExtraAddons:  ca.req.ScaleInExtraAddons,
		// 若为空默认使用集群镜像
		NodeOS:            ca.req.NodeOS,
		Runtime:           ca.req.Runtime,
		Module:            ca.req.Module,
		Creator:           ca.req.Creator,
		Updater:           ca.req.Creator,
		CreateTime:        timeStr,
		UpdateTime:        timeStr,
		Desc:              ca.req.Desc,
		ScaleInPostScript: scaleInPostScript,
		ScaleInPreScript:  scaleInPreScript,
		Annotations: func() map[string]string {
			if ca.req.Annotations == nil || len(ca.req.Annotations.Values) == 0 {
				return nil
			}
			return ca.req.Annotations.Values
		}(),
	}

	return ca.model.CreateNodeTemplate(ca.ctx, nodeTemplate)
}

func (ca *CreateAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle create cloud nodeTemplate request
func (ca *CreateAction) Handle(ctx context.Context,
	req *cmproto.CreateNodeTemplateRequest, resp *cmproto.CreateNodeTemplateResponse) {
	if req == nil || resp == nil {
		blog.Errorf("create nodeTemplate failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	if err := ca.validate(); err != nil {
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ca.createNodeTemplate(); err != nil {
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
	err = ca.checkNodeTemplateName()
	if err != nil {
		return err
	}
	err = ca.checkNodeTemplateAction()
	if err != nil {
		return err
	}

	return nil
}

func (ca *CreateAction) checkNodeTemplateName() error {
	templates, err := getAllNodeTemplates(ca.ctx, ca.model, ca.req.ProjectID)
	if err != nil {
		return err
	}

	for i := range templates {
		if ca.req.Name == templates[i].Name {
			return fmt.Errorf("project[%s] NodeTemplate[%s] duplicate", ca.req.ProjectID, ca.req.Name)
		}
	}

	return nil
}

func (ca *CreateAction) checkNodeTemplateAction() error {
	if ca.req.ScaleOutExtraAddons != nil {
		return checkExtraActionAddons(ca.req.ScaleOutExtraAddons)
	}
	if ca.req.ScaleInExtraAddons != nil {
		return checkExtraActionAddons(ca.req.ScaleInExtraAddons)
	}

	return nil
}

func checkExtraActionAddons(action *cmproto.Action) error {
	for _, pre := range action.PreActions {
		if plugin, ok := action.Plugins[pre]; ok { // nolint
			err := actionMustExistParas(pre, plugin.Params)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("PreActions plugins not contain[%s]", pre)
		}
	}

	for _, post := range action.PostActions {
		if plugin, ok := action.Plugins[post]; ok { // nolint
			err := actionMustExistParas(post, plugin.Params)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("PostActions plugins not contain[%s]", post)
		}
	}

	return nil
}

func actionMustExistParas(actionName string, paras map[string]string) error {
	templateParas := []string{cloudprovider.BkSopsBizIDKey.String(),
		cloudprovider.BkSopsTemplateIDKey.String(), cloudprovider.BkSopsTemplateUserKey.String()}

	for _, templatePara := range templateParas {
		ok, ele := utils.StringContainInMap(templatePara, paras)
		if !ok || ele == "" {
			return fmt.Errorf("action[%s] not contain templatePara[%s]", actionName, templatePara)
		}
	}

	newParas := make(map[string]string, 0)
	for k := range paras {
		newParas[k] = paras[k]
	}
	// trans to sops CM template paras
	for k, v := range newParas {
		if templateVar, ok := template.InnerTemplateVars[strings.TrimSpace(v)]; ok {
			paras[k] = templateVar.TransMethod
		}
	}

	return nil
}
