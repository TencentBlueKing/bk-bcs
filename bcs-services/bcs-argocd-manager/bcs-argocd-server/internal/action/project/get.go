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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/apis/tkex/v1alpha1"
	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/clientset/versioned/typed/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/sdk/project"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewGetArgocdProjectAction return a new GetArgocdProjectAction instance
func NewGetArgocdProjectAction(tkexIf tkexv1alpha1.TkexV1alpha1Interface) *GetArgocdProjectAction {
	return &GetArgocdProjectAction{tkexIf: tkexIf}
}

// GetArgocdProjectAction provides the action to get argocd project
type GetArgocdProjectAction struct {
	ctx context.Context

	tkexIf tkexv1alpha1.TkexV1alpha1Interface

	req  *project.GetArgocdProjectRequest
	resp *project.GetArgocdProjectResponse
}

// Handle the get process
func (action *GetArgocdProjectAction) Handle(ctx context.Context,
	req *project.GetArgocdProjectRequest, resp *project.GetArgocdProjectResponse) error {
	if req == nil || resp == nil {
		blog.Errorf("action/project/get: get project failed, req or resp is empty")
		return common.ErrArgocdServerReqOrRespEmpty.GenError()
	}
	action.ctx = ctx
	action.req = req
	action.resp = resp

	name := req.GetName()
	p, err := action.tkexIf.ArgocdProjects(common.ArgocdManagerNamespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		blog.Errorf("get project %s failed, err: %s", name, err.Error())
		action.setResp(common.ErrActionFailed, "get argocd project failed", nil)
		return nil
	}
	// TODO: check if the operator has permission in project
	blog.Infof("get argocd project %s success", name)
	action.setResp(common.ErrArgocdServerSuccess, "", p)
	return nil
}

func (action *GetArgocdProjectAction) setResp(err common.ArgocdServerError, message string, project *v1alpha1.ArgocdProject) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	action.resp.Code = &code
	action.resp.Message = &msg
	action.resp.Project = project
}
