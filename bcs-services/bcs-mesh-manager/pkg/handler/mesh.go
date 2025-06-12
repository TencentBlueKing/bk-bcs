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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/actions/mesh"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// ListMesh implements meshmanager.MeshManagerHandler
func (m *MeshManager) ListMesh(
	ctx context.Context,
	req *meshmanager.ListMeshRequest,
	resp *meshmanager.ListMeshResponse,
) error {
	action := mesh.NewListMeshAction(m.model)
	return action.Handle(ctx, req, resp)
}

// UpdateMesh implements meshmanager.MeshManagerHandler
func (m *MeshManager) UpdateMesh(
	ctx context.Context,
	req *meshmanager.UpdateMeshRequest,
	resp *meshmanager.UpdateMeshResponse,
) error {
	action := mesh.NewUpdateMeshAction(m.model)
	return action.Handle(ctx, req, resp)
}

// DeleteMesh implements meshmanager.MeshManagerHandler
func (m *MeshManager) DeleteMesh(
	ctx context.Context,
	req *meshmanager.DeleteMeshRequest,
	resp *meshmanager.DeleteMeshResponse,
) error {
	action := mesh.NewDeleteMeshAction(m.model)
	return action.Handle(ctx, req, resp)
}

// InstallIstio install istio
func (m *MeshManager) InstallIstio(
	ctx context.Context,
	req *meshmanager.InstallIstioRequest,
	resp *meshmanager.InstallIstioResponse,
) error {
	action := mesh.NewInstallMeshAction(m.model)
	// 判断 req 是否为 nil 日志打印 req
	if req == nil {
		blog.Errorf("install mesh request is nil")
		return nil
	}
	return action.Handle(ctx, req, resp)
}
