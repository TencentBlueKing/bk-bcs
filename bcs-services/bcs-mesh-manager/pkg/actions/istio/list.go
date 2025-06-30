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
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	storeutils "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/utils"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// ListIstioAction action for list istio
type ListIstioAction struct {
	istioConfig *options.IstioConfig
	model       store.MeshManagerModel
	req         *meshmanager.ListIstioRequest
	resp        *meshmanager.ListIstioResponse
}

// NewListIstioAction create list istio action
func NewListIstioAction(istioConfig *options.IstioConfig, model store.MeshManagerModel) *ListIstioAction {
	return &ListIstioAction{
		istioConfig: istioConfig,
		model:       model,
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
		blog.Errorf("list mesh failed, %s, projectCode: %s", err.Error(), l.req.ProjectCode)
		l.setResp(common.DBErrorCode, err.Error(), nil)
		return nil
	}

	// 设置成功响应
	l.setResp(common.SuccessCode, "", result)
	blog.Infof("list mesh successfully, projectCode: %s", l.req.ProjectCode)
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
	cond := l.buildQueryConditions()

	// 构建分页选项
	opt := l.buildPaginationOptions()

	total, meshIstios, err := l.model.List(ctx, cond, opt)
	if err != nil {
		return nil, err
	}
	items := make([]*meshmanager.IstioListItem, 0, len(meshIstios))
	for _, mesh := range meshIstios {
		// 存在错误数据，直接返回空结果
		if mesh == nil {
			blog.Errorf("data error, mesh is nil")
			return nil, fmt.Errorf("data error, mesh is nil")
		}
		item := mesh.Transfer2ProtoForListItems()
		items = append(items, item)
	}
	return &meshmanager.ListIstioData{
		Total: int32(total),
		Items: items,
	}, nil
}

// buildQueryConditions 构建查询条件
func (l *ListIstioAction) buildQueryConditions() *operator.Condition {
	conditions := make([]*operator.Condition, 0, 6)

	if l.req.ProjectCode != "" {
		conditions = append(conditions, operator.NewLeafCondition(operator.Eq, operator.M{
			entity.FieldKeyProjectCode: l.req.ProjectCode,
		}))
	}
	if l.req.MeshID != "" {
		conditions = append(conditions, operator.NewLeafCondition(operator.Eq, operator.M{
			entity.FieldKeyMeshID: l.req.MeshID,
		}))
	}
	if l.req.Name != "" {
		conditions = append(conditions, operator.NewLeafCondition(operator.Con, operator.M{
			entity.FieldKeyName: l.req.Name,
		}))
	}
	if l.req.Status != "" {
		conditions = append(conditions, operator.NewLeafCondition(operator.Eq, operator.M{
			entity.FieldKeyStatus: l.req.Status,
		}))
	}
	if l.req.Version != "" {
		conditions = append(conditions, operator.NewLeafCondition(operator.Eq, operator.M{
			entity.FieldKeyVersion: l.req.Version,
		}))
	}
	if l.req.ClusterID != "" {
		clusterIDArray := []string{l.req.ClusterID}
		clusterConditions := []*operator.Condition{
			operator.NewLeafCondition(operator.In, operator.M{
				entity.FieldKeyPrimaryClusters: clusterIDArray,
			}),
			operator.NewLeafCondition(operator.In, operator.M{
				entity.FieldKeyRemoteClusters: clusterIDArray,
			}),
		}
		conditions = append(conditions, operator.NewBranchCondition(operator.Or, clusterConditions...))
	}

	if len(conditions) > 0 {
		return operator.NewBranchCondition(operator.And, conditions...)
	}
	return operator.NewBranchCondition(operator.And)
}

// buildPaginationOptions 构建分页选项
func (l *ListIstioAction) buildPaginationOptions() *storeutils.ListOption {
	page := int64(1)
	pageSize := int64(10)

	if l.req.Page > 0 {
		page = int64(l.req.Page)
	}
	if l.req.PageSize > 0 {
		pageSize = int64(l.req.PageSize)
	}

	return &storeutils.ListOption{
		Sort: map[string]int{
			entity.FieldKeyCreateTime: -1,
		},
		Page: page,
		Size: pageSize,
	}
}
