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

package mesh

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// InstallMeshAction action for installing mesh
type InstallMeshAction struct {
	model store.MeshManagerModel
	req   *meshmanager.InstallIstioRequest
	resp  *meshmanager.InstallIstioResponse
}

// NewInstallMeshAction create install mesh action
func NewInstallMeshAction(model store.MeshManagerModel) *InstallMeshAction {
	return &InstallMeshAction{
		model: model,
	}
}

// Handle handles the install mesh request
func (i *InstallMeshAction) Handle(
	ctx context.Context,
	req *meshmanager.InstallIstioRequest,
	resp *meshmanager.InstallIstioResponse,
) error {
	// 打印 req 检查
	blog.Infof("install mesh request: %+v", req)

	// 打印关键字段
	if req != nil {
		blog.Infof("install mesh request details - projectID: %s, projectCode: %s, meshName: %s, version: %s",
			req.GetProjectID(),
			req.GetProjectCode(),
			req.GetMeshName(),
			req.GetVersion())
	}

	// 使用测试数据替换传入的请求
	i.req = req
	i.resp = resp

	if err := i.install(ctx); err != nil {
		i.setResp(common.DBErr, err.Error())
		return nil
	}

	i.setResp(common.Success, common.SuccessMsg)
	return nil
}

// setResp sets the response with code and message
func (i *InstallMeshAction) setResp(code uint32, message string) {
	i.resp.Code = code
	i.resp.Message = message
}

// install implements the business logic for installing mesh
func (i *InstallMeshAction) install(ctx context.Context) error {
	// 创建 Mesh 实体并转换
	mesh := &entity.Mesh{}
	mesh.TransferFromProto(i.req)

	// Create mesh in database
	return i.model.CreateMesh(ctx, mesh)
}
