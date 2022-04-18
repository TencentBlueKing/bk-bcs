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

package project

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-server/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-server/internal/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/apis/tkex/v1alpha1"
	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/clientset/versioned/typed/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/sdk/project"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewCreateArgocdProjectAction return a new CreateArgocdProjectAction instance
func NewCreateArgocdProjectAction(tkexIf tkexv1alpha1.TkexV1alpha1Interface) *CreateArgocdProjectAction {
	return &CreateArgocdProjectAction{tkexIf: tkexIf}
}

// CreateArgocdProjectAction provides the action to create argocd project
type CreateArgocdProjectAction struct {
	ctx context.Context

	tkexIf tkexv1alpha1.TkexV1alpha1Interface

	req  *project.CreateArgocdProjectRequest
	resp *project.CreateArgocdProjectResponse
}

// Handle the create process
func (action *CreateArgocdProjectAction) Handle(ctx context.Context,
	req *project.CreateArgocdProjectRequest, resp *project.CreateArgocdProjectResponse) error {
	if req == nil || resp == nil {
		blog.Errorf("action/project/create: create project failed, req or resp is empty")
		return common.ErrArgocdServerReqOrRespEmpty.GenError()
	}
	action.ctx = ctx
	action.req = req
	action.resp = resp

	var err error
	p := req.GetProject()
	// TODO: check if the operator has permission in project
	// TODO: consider using bcs project id
	if p.GetName() == "" {
		p.Name = utils.RandomString(common.ProjectNamePrefix, 5)
	}
	created := &v1alpha1.ArgocdProject{}
	for {
		created, err = action.tkexIf.ArgocdProjects(common.ArgocdManagerNamespace).Create(ctx, p, metav1.CreateOptions{})
		if errors.IsAlreadyExists(err) {
			blog.Errorf("create argocd project failed, project %s already exists, retrying...", p.Name)
			continue
		}
		if err != nil {
			blog.Errorf("create argocd project failed, err: %s", err.Error())
			action.setResp(common.ErrActionFailed, "create argocd project failed", p)
			return nil
		}
		break
	}
	blog.Infof("create argocd project %s success", created.Name)
	action.setResp(common.ErrArgocdServerSuccess, "", created)
	return nil
}

func (action *CreateArgocdProjectAction) setResp(err common.ArgocdServerError, message string, project *v1alpha1.ArgocdProject) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	action.resp.Code = &code
	action.resp.Message = &msg
	action.resp.Project = project
}
