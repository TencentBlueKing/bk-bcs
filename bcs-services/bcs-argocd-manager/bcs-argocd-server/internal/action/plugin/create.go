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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-server/internal/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/apis/tkex/v1alpha1"
	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/clientset/versioned/typed/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/sdk/plugin"
	"k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewCreateArgocdPluginAction return a new CreateArgocdPluginAction instance
func NewCreateArgocdPluginAction(tkexIf tkexv1alpha1.TkexV1alpha1Interface) *CreateArgocdPluginAction {
	return &CreateArgocdPluginAction{tkexIf: tkexIf}
}

// CreateArgocdPluginAction provides the action to create argocd plugin
type CreateArgocdPluginAction struct {
	ctx context.Context

	tkexIf tkexv1alpha1.TkexV1alpha1Interface

	req  *plugin.CreateArgocdPluginRequest
	resp *plugin.CreateArgocdPluginResponse
}

// Handle the create process
func (action *CreateArgocdPluginAction) Handle(ctx context.Context,
	req *plugin.CreateArgocdPluginRequest, resp *plugin.CreateArgocdPluginResponse) error {
	if req == nil || resp == nil {
		blog.Errorf("action/plugin/create: create plugin failed, req or resp is empty")
		return common.ErrArgocdServerReqOrRespEmpty.GenError()
	}

	action.ctx = ctx
	action.req = req
	action.resp = resp

	p := req.GetPlugin()
	if err := action.hasValidProject(p); err != nil {
		blog.Errorf("check plugin project failed, %v", err)
		action.setResp(common.ErrActionFailed, err.Error(), nil)
		return nil
	}

	if p.Labels == nil {
		p.Labels = make(map[string]string)
	}
	p.Labels[common.ArgocdProjectLabel] = p.Spec.Project
	p.Labels[common.ArgocdNickNameLabel] = p.Spec.NickName

	created, err := action.generateIDAndCreate(p)
	if err != nil {
		blog.Errorf("generate ID and create plugin failed, %v", err)
		action.setResp(common.ErrActionFailed, err.Error(), nil)
		return nil
	}

	blog.Infof("success to create plugin %s", created.Name)
	action.setResp(common.ErrArgocdServerSuccess, "", created)
	return nil
}

func (action *CreateArgocdPluginAction) generateIDAndCreate(p *v1alpha1.ArgocdPlugin) (
	*v1alpha1.ArgocdPlugin, error) {

	for i := 0; i < 3; i++ {
		p.Name = utils.RandomString(common.PluginNamePrefix, 5)
		createdPlugin, err := action.tkexIf.ArgocdPlugins(common.ArgocdManagerNamespace).
			Create(action.ctx, p, metav1.CreateOptions{})
		if err != nil {
			if errors.IsAlreadyExists(err) {
				continue
			}

			return nil, err
		}

		return createdPlugin, nil
	}

	return nil, fmt.Errorf("generate argocd plugin name and create too much time")
}

func (action *CreateArgocdPluginAction) hasValidProject(p *v1alpha1.ArgocdPlugin) error {
	if p.Spec.Project == "" {
		return fmt.Errorf("spec.project is empty")
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

func (action *CreateArgocdPluginAction) setResp(err common.ArgocdServerError, message string, p *v1alpha1.ArgocdPlugin) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	action.resp.Code = &code
	action.resp.Message = &msg
	action.resp.Plugin = p
}
