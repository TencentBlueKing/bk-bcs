/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plugin

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-server/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/apis/tkex/v1alpha1"
	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/clientset/versioned/typed/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/sdk/plugin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewGetArgocdPluginAction return a new GetArgocdPluginAction instance
func NewGetArgocdPluginAction(tkexIf tkexv1alpha1.TkexV1alpha1Interface) *GetArgocdPluginAction {
	return &GetArgocdPluginAction{tkexIf: tkexIf}
}

// GetArgocdPluginAction provides the action to get argocd plugin
type GetArgocdPluginAction struct {
	ctx context.Context

	tkexIf tkexv1alpha1.TkexV1alpha1Interface

	req  *plugin.GetArgocdPluginRequest
	resp *plugin.GetArgocdPluginResponse
}

// Handle the get process
func (action *GetArgocdPluginAction) Handle(ctx context.Context,
	req *plugin.GetArgocdPluginRequest, resp *plugin.GetArgocdPluginResponse) error {

	if req == nil || resp == nil {
		blog.Errorf("action/plugin/get: get plugin failed, req or resp is empty")
		return common.ErrArgocdServerReqOrRespEmpty.GenError()
	}
	action.ctx = ctx
	action.req = req
	action.resp = resp

	name := req.GetName()
	p, err := action.tkexIf.ArgocdPlugins(common.ArgocdManagerNamespace).
		Get(action.ctx, name, metav1.GetOptions{})
	if err != nil {
		blog.Errorf("get plugin %s failed, %v", name, err)
		action.setResp(common.ErrActionFailed, err.Error(), nil)
		return nil
	}

	if err := action.hasValidProject(p); err != nil {
		blog.Errorf("check plugin project failed, %v", err)
		action.setResp(common.ErrActionFailed, err.Error(), nil)
		return nil
	}

	blog.Infof("success to get plugin %s", name)
	action.setResp(common.ErrArgocdServerSuccess, "", p)
	return nil
}

func (action *GetArgocdPluginAction) hasValidProject(p *v1alpha1.ArgocdPlugin) error {
	if p == nil {
		return fmt.Errorf("plugin is empty")
	}

	project, err := action.tkexIf.ArgocdProjects(common.ArgocdManagerNamespace).
		Get(action.ctx, p.Spec.Project, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if project == nil {
		return fmt.Errorf("query and get empty project with name %s", p.Spec.Project)
	}

	// TODO: check if current operator has permission to deal with this project
	return nil
}

func (action *GetArgocdPluginAction) setResp(err common.ArgocdServerError, message string, p *v1alpha1.ArgocdPlugin) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	action.resp.Code = &code
	action.resp.Message = &msg
	action.resp.Plugin = p
}
