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

package instance

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/sdk/instance"
)

// NewCreateArgocdInstanceAction return a new CreateArgocdInstanceAction instance
func NewCreateArgocdInstanceAction() *CreateArgocdInstanceAction {
	return &CreateArgocdInstanceAction{}
}

// CreateArgocdInstanceAction provides the action to create argocd instance
type CreateArgocdInstanceAction struct {
	ctx context.Context

	req  *instance.CreateArgocdInstanceRequest
	resp *instance.CreateArgocdInstanceResponse
}

// Handle the create process
func (action *CreateArgocdInstanceAction) Handle(ctx context.Context,
	req *instance.CreateArgocdInstanceRequest, resp *instance.CreateArgocdInstanceResponse) error {
	return nil
}
