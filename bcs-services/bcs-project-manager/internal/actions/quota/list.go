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

// Package quota xxx
package quota

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// ListQuotaAction 定义了列出项目配额的操作结构体
type ListQuotaAction struct {
	ctx   context.Context                  // 用于传递请求上下文信息
	model store.ProjectModel               // 项目模型，用于数据库操作
	req   *proto.ListProjectQuotasRequest  // 列出项目配额的请求参数
	resp  *proto.ListProjectQuotasResponse // 列出项目配额的响应结果
}

// NewListQuotaAction new list projectquotas action
func NewListQuotaAction(model store.ProjectModel) *ListQuotaAction {
	return &ListQuotaAction{
		model: model,
		resp:  &proto.ListProjectQuotasResponse{},
	}
}

// Do list projectquotas
// Do 方法处理获取项目配额列表的请求
func (la *ListQuotaAction) Do(ctx context.Context,
	req *proto.ListProjectQuotasRequest, resp *proto.ListProjectQuotasResponse) error {
	la.ctx = ctx
	la.req = req
	la.resp = resp

	// 获取指定项目和提供商的资源使用情况
	pqs, err := clustermanager.GetResourceUsage(req.ProjectID, req.Provider)

	if err != nil {
		return err
	}

	var PQ []*proto.ProjectQuota

	// 遍历每个项目配额，构建响应数据
	for _, pq := range pqs {
		var NG []*proto.NodeGroup
		// 获取每个节点组的详细信息
		for _, gpid := range pq.TotalGroupIds {
			ng, errG := clustermanager.GetNodeGroup(gpid)
			if errG != nil {
				return errG
			}
			NG = append(NG, &proto.NodeGroup{
				NodeGroupId: ng.NodeGroupID,
				ClusterId:   ng.ClusterID,
				QuotaNum:    ng.AutoScaling.MaxSize,
				QuotaUsed:   ng.AutoScaling.DesiredSize,
			})
		}

		// 构建项目配额信息
		PQ = append(PQ, &proto.ProjectQuota{
			Quota: &proto.QuotaResource{
				ZoneResources: &proto.InstanceTypeConfig{
					Region:       pq.Region,
					InstanceType: pq.InstanceType,
					ZoneName:     pq.Zone,
					QuotaNum:     pq.Total,
					QuotaUsed:    pq.Used,
				},
			},
			NodeGroups: NG,
		})
	}

	// 设置响应数据
	resp.Data = &proto.ListProjectQuotasData{
		Total:   uint32(len(PQ)),
		Results: PQ,
	}

	return nil
}
