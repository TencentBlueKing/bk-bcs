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
	return nil
}
