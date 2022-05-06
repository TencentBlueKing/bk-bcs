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

	actions "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-server/internal/action/plugin"
	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/clientset/versioned/typed/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/sdk/plugin"
)

// PluginHandler handler that implements the micro handler interface
type PluginHandler struct {
	tkexIf tkexv1alpha1.TkexV1alpha1Interface
}

// NewPluginHandler return a new PluginHandler plugin
func NewPluginHandler(tkexIf tkexv1alpha1.TkexV1alpha1Interface) *PluginHandler {
	return &PluginHandler{tkexIf: tkexIf}
}

// CreateArgocdPlugin create a plugin
func (handler *PluginHandler) CreateArgocdPlugin(ctx context.Context,
	request *plugin.CreateArgocdPluginRequest, response *plugin.CreateArgocdPluginResponse) error {
	action := actions.NewCreateArgocdPluginAction(handler.tkexIf)
	return action.Handle(ctx, request, response)
}

// UpdateArgocdPlugin update a plugin
func (handler *PluginHandler) UpdateArgocdPlugin(ctx context.Context,
	request *plugin.UpdateArgocdPluginRequest, response *plugin.UpdateArgocdPluginResponse) error {
	action := actions.NewUpdateArgocdPluginAction(handler.tkexIf)
	return action.Handle(ctx, request, response)
}

// DeleteArgocdPlugin delete a plugin by name
func (handler *PluginHandler) DeleteArgocdPlugin(ctx context.Context,
	request *plugin.DeleteArgocdPluginRequest, response *plugin.DeleteArgocdPluginResponse) error {
	action := actions.NewDeleteArgocdPluginAction(handler.tkexIf)
	return action.Handle(ctx, request, response)
}

// GetArgocdPlugin get plugin by name
func (handler *PluginHandler) GetArgocdPlugin(ctx context.Context,
	request *plugin.GetArgocdPluginRequest, response *plugin.GetArgocdPluginResponse) error {
	action := actions.NewGetArgocdPluginAction(handler.tkexIf)
	return action.Handle(ctx, request, response)
}

// ListArgocdPlugins list plugins
func (handler *PluginHandler) ListArgocdPlugins(ctx context.Context,
	request *plugin.ListArgocdPluginsRequest, response *plugin.ListArgocdPluginsResponse) error {
	action := actions.NewListArgocdPluginsAction(handler.tkexIf)
	return action.Handle(ctx, request, response)
}
