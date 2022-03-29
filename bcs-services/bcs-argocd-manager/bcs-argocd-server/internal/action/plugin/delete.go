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

package plugin

import (
	"context"

	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/clientset/versioned/typed/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/sdk/plugin"
)

// NewDeleteArgocdPluginAction return a new DeleteArgocdPluginAction instance
func NewDeleteArgocdPluginAction(tkexIf tkexv1alpha1.TkexV1alpha1Interface) *DeleteArgocdPluginAction {
	return &DeleteArgocdPluginAction{tkexIf: tkexIf}
}

// DeleteArgocdPluginAction provides the action to delete argocd plugin
type DeleteArgocdPluginAction struct {
	ctx context.Context

	tkexIf tkexv1alpha1.TkexV1alpha1Interface

	req  *plugin.DeleteArgocdPluginRequest
	resp *plugin.DeleteArgocdPluginResponse
}

// Handle the delete process
func (action *DeleteArgocdPluginAction) Handle(ctx context.Context,
	req *plugin.DeleteArgocdPluginRequest, resp *plugin.DeleteArgocdPluginResponse) error {
	return nil
}
