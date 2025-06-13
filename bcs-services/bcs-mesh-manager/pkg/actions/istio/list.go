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

package istio

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/utils"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// ListIstioAction action for listing istio
type ListIstioAction struct {
	model store.MeshManagerModel
	req   *meshmanager.ListIstioRequest
	resp  *meshmanager.ListIstioResponse
}

// NewListIstioAction create list istio action
func NewListIstioAction(model store.MeshManagerModel) *ListIstioAction {
	return &ListIstioAction{
		model: model,
	}
}

// Handle processes the mesh list request
func (l *ListIstioAction) Handle(
	ctx context.Context,
	req *meshmanager.ListIstioRequest,
	resp *meshmanager.ListIstioResponse,
) error {
	l.req = req
	l.resp = resp

	if err := l.req.Validate(); err != nil {
		blog.Errorf("list mesh failed, invalid request, %s, param: %v", err.Error(), l.req)
		l.setResp(common.ParamErrorCode, err.Error(), nil)
		return nil
	}

	result, err := l.list(ctx)
	if err != nil {
		blog.Errorf("list mesh failed, %s, projectID: %s", err.Error(), l.req.ProjectID)
		l.setResp(common.DBErrorCode, err.Error(), nil)
		return nil
	}

	// 设置成功响应
	l.setResp(common.SuccessCode, "", result)
	blog.Infof("list mesh successfully, projectID: %s", l.req.ProjectID)
	return nil
}

// setResp sets the response with code, message and data
func (l *ListIstioAction) setResp(code uint32, message string, data *meshmanager.ListIstioData) {
	l.resp.Code = code
	l.resp.Message = message
	l.resp.Data = data
}

// list implements the business logic for listing meshes
func (l *ListIstioAction) list(ctx context.Context) (*meshmanager.ListIstioData, error) {
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

	total, meshIstios, err := l.model.List(ctx, cond, opt)
	if err != nil {
		return nil, err
	}
	items := make([]*meshmanager.IstioListItem, 0, len(meshIstios))
	blog.Infof("list mesh istio: total: %d, cond: %v", total, cond)
	for _, mesh := range meshIstios {
		item := mesh.Transfer2Proto()
		if item == nil {
			blog.Warnf("list mesh: failed to convert mesh to proto, meshID: %s", mesh.MeshID)
			continue
		}
		items = append(items, item)
	}

	return &meshmanager.ListIstioData{
		Total: int32(total),
		Items: items,
	}, nil
}
