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

	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/clientset/versioned/typed/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/sdk/project"
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
	return nil
}
