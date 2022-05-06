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

package handler

import (
	"context"

	actions "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-server/internal/action/project"
	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/clientset/versioned/typed/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/sdk/project"
)

// ProjectHandler handler that implements the micro handler interface
type ProjectHandler struct {
	tkexIf tkexv1alpha1.TkexV1alpha1Interface
}

// NewProjectHandler return a new ProjectHandler instance
func NewProjectHandler(tkexIf tkexv1alpha1.TkexV1alpha1Interface) *ProjectHandler {
	return &ProjectHandler{tkexIf: tkexIf}
}

// CreateArgocdProject create argocd project
func (handler *ProjectHandler) CreateArgocdProject(ctx context.Context,
	request *project.CreateArgocdProjectRequest, response *project.CreateArgocdProjectResponse) error {
	action := actions.NewCreateArgocdProjectAction(handler.tkexIf)
	return action.Handle(ctx, request, response)
}

// UpdateArgocdProject update argocd project
func (handler *ProjectHandler) UpdateArgocdProject(ctx context.Context,
	request *project.UpdateArgocdProjectRequest, response *project.UpdateArgocdProjectResponse) error {
	action := actions.NewUpdateArgocdProjectAction(handler.tkexIf)
	return action.Handle(ctx, request, response)
}

// DeleteArgocdProject delete argocd project by name
func (handler *ProjectHandler) DeleteArgocdProject(ctx context.Context,
	request *project.DeleteArgocdProjectRequest, response *project.DeleteArgocdProjectResponse) error {
	action := actions.NewDeleteArgocdProjectAction(handler.tkexIf)
	return action.Handle(ctx, request, response)
}

// GetArgocdProject get argocd project by name
func (handler *ProjectHandler) GetArgocdProject(ctx context.Context,
	request *project.GetArgocdProjectRequest, response *project.GetArgocdProjectResponse) error {
	action := actions.NewGetArgocdProjectAction(handler.tkexIf)
	return action.Handle(ctx, request, response)
}

// ListArgocdProjects list argocd projects
func (handler *ProjectHandler) ListArgocdProjects(ctx context.Context,
	request *project.ListArgocdProjectsRequest, response *project.ListArgocdProjectsResponse) error {
	action := actions.NewListArgocdProjectsAction(handler.tkexIf)
	return action.Handle(ctx, request, response)
}
