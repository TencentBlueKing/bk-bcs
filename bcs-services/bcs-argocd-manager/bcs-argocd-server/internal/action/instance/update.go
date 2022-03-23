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

// NewUpdateArgocdInstanceAction return a new UpdateArgocdInstanceAction instance
func NewUpdateArgocdInstanceAction() *UpdateArgocdInstanceAction {
	return &UpdateArgocdInstanceAction{}
}

// UpdateArgocdInstanceAction provides the action to update argocd instance
type UpdateArgocdInstanceAction struct {
	ctx context.Context

	req  *instance.UpdateArgocdInstanceRequest
	resp *instance.UpdateArgocdInstanceResponse
}

// Handle the update process
func (action *UpdateArgocdInstanceAction) Handle(ctx context.Context,
	req *instance.UpdateArgocdInstanceRequest, resp *instance.UpdateArgocdInstanceResponse) error {
	return nil
}
