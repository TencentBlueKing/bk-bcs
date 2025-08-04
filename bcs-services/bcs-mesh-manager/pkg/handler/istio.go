/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package handler

import (
	"context"

	istioaction "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/actions/istio"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// ListIstio implements meshmanager.MeshManagerHandler
func (m *MeshManager) ListIstio(
	ctx context.Context,
	req *meshmanager.ListIstioRequest,
	resp *meshmanager.ListIstioResponse,
) error {
	action := istioaction.NewListIstioAction(m.opt.IstioConfig, m.model)
	return action.Handle(ctx, req, resp)
}

// InstallIstio implements meshmanager.MeshManagerHandler
func (m *MeshManager) InstallIstio(
	ctx context.Context,
	req *meshmanager.IstioRequest,
	resp *meshmanager.InstallIstioResponse,
) error {
	action := istioaction.NewInstallIstioAction(m.opt.IstioConfig, m.model)
	return action.Handle(ctx, req, resp)
}

// UpdateIstio implements meshmanager.MeshManagerHandler
func (m *MeshManager) UpdateIstio(
	ctx context.Context,
	req *meshmanager.IstioRequest,
	resp *meshmanager.UpdateIstioResponse,
) error {
	action := istioaction.NewUpdateIstioAction(m.opt.IstioConfig, m.model)
	return action.Handle(ctx, req, resp)
}

// DeleteIstio implements meshmanager.MeshManagerHandler
func (m *MeshManager) DeleteIstio(
	ctx context.Context,
	req *meshmanager.DeleteIstioRequest,
	resp *meshmanager.DeleteIstioResponse,
) error {
	action := istioaction.NewDeleteIstioAction(m.model)
	return action.Handle(ctx, req, resp)
}

// ListIstioConfig implements meshmanager.MeshManagerHandler
func (m *MeshManager) ListIstioConfig(
	ctx context.Context,
	req *meshmanager.ListIstioConfigRequest,
	resp *meshmanager.ListIstioConfigResponse,
) error {
	action := istioaction.NewListIstioConfigAction(m.opt.IstioConfig)
	return action.Handle(ctx, req, resp)
}

// GetIstioDetail implements meshmanager.MeshManagerHandler
func (m *MeshManager) GetIstioDetail(
	ctx context.Context,
	req *meshmanager.GetIstioDetailRequest,
	resp *meshmanager.GetIstioDetailResponse,
) error {
	action := istioaction.NewGetIstioDetailAction(m.opt.IstioConfig, m.model)
	return action.Handle(ctx, req, resp)
}

// GetClusterInfo implements meshmanager.MeshManagerHandler
func (m *MeshManager) GetClusterInfo(
	ctx context.Context,
	req *meshmanager.GetClusterInfoRequest,
	resp *meshmanager.GetClusterInfoResponse,
) error {
	action := istioaction.NewGetClusterInfoAction(m.model)
	return action.Handle(ctx, req, resp)
}
