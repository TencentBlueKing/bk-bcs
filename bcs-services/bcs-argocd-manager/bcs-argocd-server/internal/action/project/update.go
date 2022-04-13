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

// NewUpdateArgocdProjectAction return a new UpdateArgocdProjectAction instance
func NewUpdateArgocdProjectAction(tkexIf tkexv1alpha1.TkexV1alpha1Interface) *UpdateArgocdProjectAction {
	return &UpdateArgocdProjectAction{tkexIf: tkexIf}
}

// UpdateArgocdProjectAction provides the action to update argocd project
type UpdateArgocdProjectAction struct {
	ctx context.Context

	tkexIf tkexv1alpha1.TkexV1alpha1Interface

	req  *project.UpdateArgocdProjectRequest
	resp *project.UpdateArgocdProjectResponse
}

// Handle the update process
func (action *UpdateArgocdProjectAction) Handle(ctx context.Context,
	req *project.UpdateArgocdProjectRequest, resp *project.UpdateArgocdProjectResponse) error {
	if req == nil || resp == nil {
		blog.Errorf("action/project/update: update project failed, req or resp is empty")
		return common.ErrArgocdServerReqOrRespEmpty.GenError()
	}
	action.ctx = ctx
	action.req = req
	action.resp = resp

	p := req.GetProject()
	// TODO: check if the operator has permission in project
	updated, err := action.tkexIf.ArgocdProjects(common.ArgocdManagerNamespace).Update(ctx, p, metav1.UpdateOptions{})
	if err != nil {
		blog.Errorf("update argocd project failed, err: %s", err.Error())
		action.setResp(common.ErrActionFailed, "update argocd project failed", p)
		return nil
	}
	blog.Infof("update argocd project %s success", updated.Name)
	action.setResp(common.ErrArgocdServerSuccess, "", updated)
	return nil
}

func (action *UpdateArgocdProjectAction) setResp(err common.ArgocdServerError, message string, project *v1alpha1.ArgocdProject) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	action.resp.Code = &code
	action.resp.Message = &msg
	action.resp.Project = project
}
