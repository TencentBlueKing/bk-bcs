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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/helmmanager"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"gopkg.in/yaml.v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/helm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	storeutils "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/utils"
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
	blog.Infof("list mesh successfully, projectCode: %s, total: %d, len(meshIstios): %d, cond: %v, opt: %v",
		l.req.ProjectCode, total, len(meshIstios), cond, opt)
	if err != nil {
		return nil, err
	}
	items := make([]*meshmanager.IstioListItem, 0, len(meshIstios))
	if len(meshIstios) == 0 {
		return &meshmanager.ListIstioData{
			Total: int32(total),
			Items: items,
		}, nil
	}

	for _, mesh := range meshIstios {
		clusterID := ""
		if mesh == nil {
			continue
		}
		if len(mesh.PrimaryClusters) == 0 {
			continue
		}
		clusterID = mesh.PrimaryClusters[0]
		// 如果没有找到有效的clusterID，直接返回空结果
		if clusterID == "" {
			blog.Warnf("no valid clusterID found in mesh list, returning empty result")
			return &meshmanager.ListIstioData{
				Total: int32(total),
				Items: items,
			}, nil
		}

		namespace := common.IstioNamespace
		istiodName := common.IstioInstallIstiodName
		// 获取 release 详情
		release, err := helm.GetReleaseDetail(
			ctx,
			&helmmanager.GetReleaseDetailV1Req{
				ProjectCode: &l.req.ProjectCode,
				ClusterID:   &clusterID,
				Namespace:   &namespace,
				Name:        &istiodName,
			},
		)

		if err != nil {
			blog.Errorf("get release detail failed, clusterID: %s, err: %s", clusterID, err.Error())
			return nil, fmt.Errorf("get release detail failed, clusterID: %s, err: %s", clusterID, err.Error())
		}

		if release == nil || release.Data == nil {
			blog.Warnf("release is nil, clusterID: %s", clusterID)
			continue
		}

		// 检查 release values 是否为空
		if len(release.Data.Values) == 0 {
			blog.Errorf("release values is empty, clusterID: %s", clusterID)
			return nil, fmt.Errorf("release values is empty, clusterID: %s", clusterID)
		}
		values := release.Data.Values[0]

		// 解析 values 为 IstiodInstallValues 结构
		istiodValues := &common.IstiodInstallValues{}
		if err = yaml.Unmarshal([]byte(values), istiodValues); err != nil {
			blog.Errorf("unmarshal istiod values failed, clusterID: %s, err: %s", clusterID, err.Error())
			return nil, fmt.Errorf("unmarshal istiod values failed, clusterID: %s, err: %s", clusterID, err.Error())
		}

		item, err := utils.ConvertValuesToListItem(mesh, istiodValues)
		if err != nil {
			blog.Errorf("build istio list item failed, meshID: %s, err: %s", mesh.MeshID, err.Error())
			return nil, err
		}
		items = append(items, item)
	}
	return &meshmanager.ListIstioData{
		Total: int32(total),
		Items: items,
	}, nil
}

// buildQueryConditions 构建查询条件
func (l *ListIstioAction) buildQueryConditions() *operator.Condition {
	conditions := make([]*operator.Condition, 0)

	if l.req.ProjectCode != "" {
		conditions = append(conditions, operator.NewLeafCondition(operator.Eq, operator.M{
			"projectCode": l.req.ProjectCode,
		}))
	}
	if l.req.MeshID != "" {
		conditions = append(conditions, operator.NewLeafCondition(operator.Eq, operator.M{
			"meshID": l.req.MeshID,
		}))
	}
	if l.req.Name != "" {
		conditions = append(conditions, operator.NewLeafCondition(operator.Con, operator.M{
			"meshName": l.req.Name,
		}))
	}
	if l.req.Status != "" {
		conditions = append(conditions, operator.NewLeafCondition(operator.Eq, operator.M{
			"status": l.req.Status,
		}))
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
			"createTime": -1,
		},
		Page: page,
		Size: pageSize,
	}
}
