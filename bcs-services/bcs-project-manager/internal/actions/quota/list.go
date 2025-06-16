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

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/quota"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
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

func (la *ListQuotaAction) doHost(ctx context.Context, req *proto.ListProjectQuotasRequest,
	pquota []*proto.ProjectQuota) ([]*proto.ProjectQuota, error) {
	p, err := la.model.GetProject(la.ctx, req.ProjectID)
	if err != nil {
		return pquota, errorx.NewDBErr(err.Error())
	}

	if p == nil {
		return pquota, errorx.NewReadableErr(errorx.ParamErr, "project not found")
	}

	if _, ok := p.Labels["quota-gray"]; ok {
		var conds []*operator.Condition

		conds = append(conds, operator.NewLeafCondition(operator.Eq, operator.M{
			"quotaType": quota.Host,
		}))

		if req.ProjectID != "" {
			conds = append(conds, operator.NewLeafCondition(operator.Eq, operator.M{
				"projectId": la.req.GetProjectID(),
			}))
		}

		cond := operator.NewBranchCondition(operator.And, conds...)

		quotas, _, errQ := la.model.ListProjectQuotas(la.ctx, cond, &page.Pagination{All: true})
		if err != nil {
			return pquota, errorx.NewDBErr(errQ.Error())
		}

		for _, q := range quotas {
			tmp := q
			pquota = append(pquota, quota.TransStore2ProtoQuota(&tmp))
		}
	}

	pq, errU := la.doHostUsage(ctx, req, pquota, p)
	if errU != nil {
		return pquota, errU
	}
	pquota = pq

	return pquota, nil
}

func (la *ListQuotaAction) doHostUsage(ctx context.Context, req *proto.ListProjectQuotasRequest,
	pquota []*proto.ProjectQuota, p *project.Project) ([]*proto.ProjectQuota, error) {
	// 获取指定项目和提供商的资源使用情况
	pqs, errC := clustermanager.GetResourceUsage(ctx, req.ProjectID, req.Provider)

	if errC != nil {
		return pquota, errC
	}

	// 遍历每个项目配额，构建响应数据
	for _, pq := range pqs {
		var NG []*proto.NodeGroup
		// 获取每个节点组的详细信息
		for _, gpid := range pq.TotalGroupIds {
			ng, errG := clustermanager.GetNodeGroup(ctx, gpid)
			if errG != nil {
				return pquota, errG
			}
			NG = append(NG, &proto.NodeGroup{
				NodeGroupId: ng.NodeGroupID,
				ClusterId:   ng.ClusterID,
				QuotaNum:    ng.AutoScaling.MaxSize,
				QuotaUsed:   ng.AutoScaling.DesiredSize,
			})
		}

		cpu, mem := GetCpuMemFromInstanceType(pq.InstanceType)

		if _, ok := p.Labels["quota-gray"]; ok {
			for _, pqpq := range pquota {
				if pqpq.Quota.ZoneResources.ZoneName == pq.Zone &&
					pqpq.Quota.ZoneResources.InstanceType == pq.InstanceType &&
					pqpq.Status == string(quota.Running) {
					pqpq.Quota.ZoneResources.Cpu = cpu
					pqpq.Quota.ZoneResources.Mem = mem
					pqpq.Quota.ZoneResources.QuotaUsed = pq.Used
					pqpq.NodeGroups = NG
				}
			}
		} else {
			// 构建项目配额信息
			pquota = append(pquota, &proto.ProjectQuota{
				Quota: &proto.QuotaResource{
					ZoneResources: &proto.InstanceTypeConfig{
						Region:       pq.Region,
						InstanceType: pq.InstanceType,
						Cpu:          cpu,
						Mem:          mem,
						ZoneName:     pq.Zone,
						QuotaNum:     pq.Total,
						QuotaUsed:    pq.Used,
					},
				},
				NodeGroups: NG,
				QuotaType:  string(quota.Host),
				ProjectID:  req.ProjectID,
				Status:     string(quota.Running),
			})
		}
	}
	return pquota, nil
}

func (la *ListQuotaAction) doFed(req *proto.ListProjectQuotasRequest,
	pquota []*proto.ProjectQuota) ([]*proto.ProjectQuota, error) {
	var conds []*operator.Condition

	conds = append(conds, operator.NewLeafCondition(operator.Eq, operator.M{
		"quotaType": quota.Federation,
	}))

	if req.ProjectID != "" {
		conds = append(conds, operator.NewLeafCondition(operator.Eq, operator.M{
			"projectId": la.req.GetProjectID(),
		}))
	}

	cond := operator.NewBranchCondition(operator.And, conds...)

	Quotas, _, err := la.model.ListProjectQuotas(la.ctx, cond, &page.Pagination{All: true})
	if err != nil {
		return pquota, errorx.NewDBErr(err.Error())
	}

	for _, q := range Quotas {
		tmp := q
		getQuotaUsage(&tmp)
		pquota = append(pquota, quota.TransStore2ProtoQuota(&tmp))
	}
	return pquota, nil
}

// Do list projectquotas
// Do 方法处理获取项目配额列表的请求
func (la *ListQuotaAction) Do(ctx context.Context,
	req *proto.ListProjectQuotasRequest, resp *proto.ListProjectQuotasResponse) error {
	la.ctx = ctx
	la.req = req
	la.resp = resp

	var PQ []*proto.ProjectQuota

	if req.ProjectID != "" && req.GetQuotaType() != string(quota.Federation) {
		pq, err := la.doHost(ctx, req, PQ)
		if err != nil {
			return err
		}
		PQ = pq
	}

	if req.GetQuotaType() != string(quota.Host) {
		pq, err := la.doFed(req, PQ)
		if err != nil {
			return err
		}
		PQ = pq
	}

	// 设置响应数据
	resp.Data = &proto.ListProjectQuotasData{
		Total:   uint32(len(PQ)),
		Results: PQ,
	}

	return nil
}
