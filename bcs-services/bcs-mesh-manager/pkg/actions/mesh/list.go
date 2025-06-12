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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/utils"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// ListMeshAction action for listing mesh
type ListMeshAction struct {
	model store.MeshManagerModel
	req   *meshmanager.ListMeshRequest
	resp  *meshmanager.ListMeshResponse
}

// NewListMeshAction create list mesh action
func NewListMeshAction(model store.MeshManagerModel) *ListMeshAction {
	return &ListMeshAction{
		model: model,
	}
}

// Handle processes the mesh list request
func (l *ListMeshAction) Handle(
	ctx context.Context,
	req *meshmanager.ListMeshRequest,
	resp *meshmanager.ListMeshResponse,
) error {
	l.req = req
	l.resp = resp

	if err := l.req.Validate(); err != nil {
		blog.Errorf("list mesh failed, invalid request, %s, param: %v", err.Error(), l.req)
		l.setResp(common.ParamErr, err.Error(), nil)
		return nil
	}

	result, err := l.list(ctx)
	if err != nil {
		blog.Errorf("list mesh failed, %s, projectID: %s", err.Error(), l.req.ProjectID)
		l.setResp(common.DBErr, err.Error(), nil)
		return nil
	}

	// 设置成功响应
	l.setResp(common.Success, common.SuccessMsg, result)
	blog.Infof("list mesh successfully, projectID: %s", l.req.ProjectID)
	return nil
}

// setResp sets the response with code, message and data
func (l *ListMeshAction) setResp(code uint32, message string, data *meshmanager.ListMeshData) {
	l.resp.Code = code
	l.resp.Message = message
	l.resp.Data = data
}

// list implements the business logic for listing meshes
func (l *ListMeshAction) list(ctx context.Context) (*meshmanager.ListMeshData, error) {
	// 构建查询条件
	conditions := make([]*operator.Condition, 0)
	if l.req.ProjectID != "" {
		conditions = append(conditions, operator.NewLeafCondition(operator.Eq, operator.M{
			"projectID": l.req.ProjectID,
		}))
	}
	if l.req.Status != "" {
		conditions = append(conditions, operator.NewLeafCondition(operator.Eq, operator.M{
			"status": l.req.Status,
		}))
	}

	var cond *operator.Condition
	if len(conditions) > 0 {
		cond = operator.NewBranchCondition(operator.And, conditions...)
	} else {
		cond = operator.NewBranchCondition(operator.And)
	}

	// 构建分页选项
	page := int64(1)
	if l.req.Page > 0 {
		page = int64(l.req.Page)
	}
	pageSize := int64(10)
	if l.req.PageSize > 0 {
		pageSize = int64(l.req.PageSize)
	}
	opt := &utils.ListOption{
		Sort: map[string]int{
			"createTime": -1,
		},
		Page: page,
		Size: pageSize,
	}

	total, meshes, err := l.model.ListMesh(ctx, cond, opt)
	if err != nil {
		return nil, err
	}
	items := make([]*meshmanager.MeshListItem, 0, len(meshes))
	for _, mesh := range meshes {
		item := mesh.Transfer2Proto()
		if item == nil {
			blog.Warnf("list mesh: failed to convert mesh to proto, meshID: %s", mesh.MeshID)
			continue
		}
		items = append(items, item)
	}

	return &meshmanager.ListMeshData{
		Total: int32(total),
		Items: items,
	}, nil
}
