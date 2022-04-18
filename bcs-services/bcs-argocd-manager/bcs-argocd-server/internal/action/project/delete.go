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
	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/clientset/versioned/typed/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/sdk/project"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewDeleteArgocdProjectAction return a new DeleteArgocdProjectAction instance
func NewDeleteArgocdProjectAction(tkexIf tkexv1alpha1.TkexV1alpha1Interface) *DeleteArgocdProjectAction {
	return &DeleteArgocdProjectAction{tkexIf: tkexIf}
}

// DeleteArgocdProjectAction provides the action to delete argocd project
type DeleteArgocdProjectAction struct {
	ctx context.Context

	tkexIf tkexv1alpha1.TkexV1alpha1Interface

	req  *project.DeleteArgocdProjectRequest
	resp *project.DeleteArgocdProjectResponse
}

// Handle the delete process
func (action *DeleteArgocdProjectAction) Handle(ctx context.Context,
	req *project.DeleteArgocdProjectRequest, resp *project.DeleteArgocdProjectResponse) error {
	if req == nil || resp == nil {
		blog.Errorf("action/project/delete: delete project failed, req or resp is empty")
		return common.ErrArgocdServerReqOrRespEmpty.GenError()
	}
	action.ctx = ctx
	action.req = req
	action.resp = resp

	name := req.GetName()
	// TODO: check if the operator has permission in project
	err := action.tkexIf.ArgocdProjects(common.ArgocdManagerNamespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		blog.Errorf("delete argocd project failed, err: %s", err.Error())
		action.setResp(common.ErrActionFailed, "delete argocd project failed")
		return nil
	}
	blog.Infof("delete argocd project %s success", name)
	action.setResp(common.ErrArgocdServerSuccess, "")
	return nil
}

func (action *DeleteArgocdProjectAction) setResp(err common.ArgocdServerError, message string) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	action.resp.Code = &code
	action.resp.Message = &msg
}
