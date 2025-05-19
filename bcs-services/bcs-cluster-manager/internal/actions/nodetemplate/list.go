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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/template"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// ListAction action for list nodeTemplate
type ListAction struct {
	ctx   context.Context
	model store.ClusterManagerModel

	req              *cmproto.ListNodeTemplateRequest
	resp             *cmproto.ListNodeTemplateResponse
	nodeTemplateList []*cmproto.NodeTemplate
}

// NewListAction create list action for cluster nodeTemplate list
func NewListAction(model store.ClusterManagerModel) *ListAction {
	return &ListAction{
		model: model,
	}
}

func (la *ListAction) listNodeTemplates() error {
	condM := make(operator.M)
	//! we don't setting bson tag in proto file
	//! all fields are in lowcase
	if len(la.req.ProjectID) != 0 {
		condM["projectid"] = la.req.ProjectID
	}
	if len(la.req.NodeTemplateID) != 0 {
		condM["nodetemplateid"] = la.req.NodeTemplateID
	}

	cond := operator.NewLeafCondition(operator.Eq, condM)
	templates, err := la.model.ListNodeTemplate(la.ctx, cond, &storeopt.ListOption{
		Sort: map[string]int{
			"updatetime": -1,
		},
	})
	if err != nil {
		return err
	}

	for i := range templates {
		templates[i].UpdateTime = utils.TransTimeFormat(templates[i].UpdateTime)
		templates[i].CreateTime = utils.TransTimeFormat(templates[i].CreateTime)
		err = decodeTemplateScript(templates[i])
		if err != nil {
			return err
		}
		err = decodeAction(templates[i])
		if err != nil {
			return err
		}

		la.nodeTemplateList = append(la.nodeTemplateList, templates[i])
	}
	return nil
}

func (la *ListAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.nodeTemplateList
}

// Handle handle list cluster nodeTemplate list
func (la *ListAction) Handle(
	ctx context.Context, req *cmproto.ListNodeTemplateRequest, resp *cmproto.ListNodeTemplateResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list nodeTemplate failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listNodeTemplates(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// GetAction action for getting project nodeTemplate
type GetAction struct {
	ctx context.Context

	model store.ClusterManagerModel
	req   *cmproto.GetNodeTemplateRequest
	resp  *cmproto.GetNodeTemplateResponse
}

// NewGetAction create get action for online project template
func NewGetAction(model store.ClusterManagerModel) *GetAction {
	return &GetAction{
		model: model,
	}
}

func (ga *GetAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handle get cluster credential
func (ga *GetAction) Handle(
	ctx context.Context, req *cmproto.GetNodeTemplateRequest, resp *cmproto.GetNodeTemplateResponse) {
	if req == nil || resp == nil {
		blog.Errorf("get nodeTemplate failed, req or resp is empty")
		return
	}
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	if err := req.Validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	template, err := ga.model.GetNodeTemplate(ctx, req.ProjectID, req.NodeTemplateID)
	if err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err = decodeTemplateScript(template)
	if err != nil {
		ga.setResp(common.BcsErrClusterManagerDecodeBase64ScriptErr, err.Error())
		return
	}

	err = decodeAction(template)
	if err != nil {
		ga.setResp(common.BcsErrClusterManagerDecodeActionErr, err.Error())
		return
	}

	resp.Data = template
	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

func decodeTemplateScript(template *cmproto.NodeTemplate) error {
	if len(template.GetPreStartUserScript()) > 0 {
		preScript, err := base64.StdEncoding.DecodeString(template.GetPreStartUserScript())
		if err != nil {
			return err
		}
		template.PreStartUserScript = string(preScript)
	}

	if len(template.GetUserScript()) > 0 {
		afterScript, err := base64.StdEncoding.DecodeString(template.GetUserScript())
		if err != nil {
			return err
		}
		template.UserScript = string(afterScript)
	}

	return nil
}

func decodeAction(template *cmproto.NodeTemplate) error {
	if template.ScaleOutExtraAddons != nil {
		err := decodeTemplateAction(template.ScaleOutExtraAddons)
		if err != nil {
			return err
		}
	}

	if template.ScaleInExtraAddons != nil {
		err := decodeTemplateAction(template.ScaleInExtraAddons)
		if err != nil {
			return err
		}
	}

	return nil
}

func decodeTemplateAction(action *cmproto.Action) error {
	for _, pre := range action.PreActions {
		if plugin, ok := action.Plugins[pre]; ok { // nolint
			actionTransParasToFront(plugin.Params)
		} else {
			return fmt.Errorf("PreActions plugins not contain[%s]", pre)
		}
	}

	for _, post := range action.PostActions {
		if plugin, ok := action.Plugins[post]; ok { // nolint
			actionTransParasToFront(plugin.Params)
		} else {
			return fmt.Errorf("PostActions plugins not contain[%s]", post)
		}
	}

	return nil
}

func actionTransParasToFront(paras map[string]string) {
	newParas := make(map[string]string, 0)
	for k := range paras {
		newParas[k] = paras[k]
	}
	// trans to sops CM template paras
	for k, v := range newParas {
		if isSopsCMTemplateVars(v) {
			paras[k] = template.TransToReferMethod[v]
		}
	}
}
