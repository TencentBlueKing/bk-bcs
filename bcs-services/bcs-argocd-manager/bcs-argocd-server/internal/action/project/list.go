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

// NewListArgocdProjectsAction return a new ListArgocdProjectsAction instance
func NewListArgocdProjectsAction(tkexIf tkexv1alpha1.TkexV1alpha1Interface) *ListArgocdProjectsAction {
	return &ListArgocdProjectsAction{tkexIf: tkexIf}
}

// ListArgocdProjectsAction provides the action to list argocd project
type ListArgocdProjectsAction struct {
	ctx context.Context

	tkexIf tkexv1alpha1.TkexV1alpha1Interface

	req  *project.ListArgocdProjectsRequest
	resp *project.ListArgocdProjectsResponse
}

// Handle the list process
func (action *ListArgocdProjectsAction) Handle(ctx context.Context,
	req *project.ListArgocdProjectsRequest, resp *project.ListArgocdProjectsResponse) error {
	if req == nil || resp == nil {
		blog.Errorf("action/project/list: list projects failed, req or resp is empty")
		return common.ErrArgocdServerReqOrRespEmpty.GenError()
	}
	action.ctx = ctx
	action.req = req
	action.resp = resp

	listOptions := metav1.ListOptions{}
	list, err := action.tkexIf.ArgocdProjects(common.ArgocdManagerNamespace).List(ctx, listOptions)
	if err != nil {
		blog.Errorf("list projects failed, err: %s", err.Error())
		action.setResp(common.ErrActionFailed, "list projects failed", nil)
		return nil
	}
	blog.Info("list projects success")
	action.setResp(common.ErrArgocdServerSuccess, "", list)
	return nil
}

func (action *ListArgocdProjectsAction) setResp(err common.ArgocdServerError, message string, projects *v1alpha1.ArgocdProjectList) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	action.resp.Code = &code
	action.resp.Message = &msg
	action.resp.Projects = projects
}
