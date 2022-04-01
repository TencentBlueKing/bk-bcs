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
	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/clientset/versioned/typed/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/sdk/plugin"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewDeleteArgocdPluginAction return a new DeleteArgocdPluginAction instance
func NewDeleteArgocdPluginAction(tkexIf tkexv1alpha1.TkexV1alpha1Interface) *DeleteArgocdPluginAction {
	return &DeleteArgocdPluginAction{tkexIf: tkexIf}
}

// DeleteArgocdPluginAction provides the action to delete argocd plugin
type DeleteArgocdPluginAction struct {
	ctx context.Context

	tkexIf tkexv1alpha1.TkexV1alpha1Interface

	req  *plugin.DeleteArgocdPluginRequest
	resp *plugin.DeleteArgocdPluginResponse
}

// Handle the delete process
func (action *DeleteArgocdPluginAction) Handle(ctx context.Context,
	req *plugin.DeleteArgocdPluginRequest, resp *plugin.DeleteArgocdPluginResponse) error {

	if req == nil || resp == nil {
		blog.Errorf("action/plugin/delete: delete plugin failed, req or resp is empty")
		return common.ErrArgocdServerReqOrRespEmpty.GenError()
	}
	action.ctx = ctx
	action.req = req
	action.resp = resp

	name := req.GetName()
	if err := action.hasValidProject(name); err != nil {
		blog.Errorf("check plugin project failed, %v", err)
		action.setResp(common.ErrActionFailed, err.Error())
		return nil
	}

	if err := action.tkexIf.ArgocdPlugins(common.ArgocdManagerNamespace).
		Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		blog.Errorf("delete plugin %s failed, %v", name, err)
		action.setResp(common.ErrActionFailed, err.Error())
		return nil
	}

	blog.Infof("success tot delete plugin %s", name)
	action.setResp(common.ErrArgocdServerSuccess, "")
	return nil
}

func (action *DeleteArgocdPluginAction) hasValidProject(pluginName string) error {
	if pluginName == "" {
		return fmt.Errorf("plugin name is empty")
	}

	p, err := action.tkexIf.ArgocdPlugins(common.ArgocdManagerNamespace).
		Get(action.ctx, pluginName, metav1.GetOptions{})
	if err != nil {
		return err
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

func (action *DeleteArgocdPluginAction) setResp(err common.ArgocdServerError, message string) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	action.resp.Code = &code
	action.resp.Message = &msg
}
