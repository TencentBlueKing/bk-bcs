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
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
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

		conds = append(conds, operator.NewLeafCondition(operator.Ne, operator.M{
			"status": quota.Deleted,
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

	conds = append(conds, operator.NewLeafCondition(operator.Ne, operator.M{
		"status": quota.Deleted,
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

// ListQuotaV2Action 定义了列出项目配额的操作结构体
type ListQuotaV2Action struct {
	ctx   context.Context                    // 用于传递请求上下文信息
	model store.ProjectModel                 // 项目模型，用于数据库操作
	req   *proto.ListProjectQuotasV2Request  // 列出项目配额的请求参数
	resp  *proto.ListProjectQuotasV2Response // 列出项目配额的响应结果

	project *project.Project
}

// NewListQuotaV2Action new list projectquotas action
func NewListQuotaV2Action(model store.ProjectModel) *ListQuotaV2Action {
	return &ListQuotaV2Action{
		model: model,
		resp:  &proto.ListProjectQuotasV2Response{},
	}
}

// Do list projectquotas
// Do 方法处理获取项目配额列表的请求
func (la *ListQuotaV2Action) Do(ctx context.Context,
	req *proto.ListProjectQuotasV2Request, resp *proto.ListProjectQuotasV2Response) error {
	la.ctx = ctx
	la.req = req
	la.resp = resp

	var PQ []*proto.ProjectQuota

	err := la.validate()
	if err != nil {
		return err
	}

	pQuotas, count, err := la.getProjectQuotas(ctx, req, PQ)
	if err != nil {
		return err
	}

	pQuotas, err = la.getQuotasUsage(ctx, req, pQuotas)
	if err != nil {
		return err
	}

	// 设置响应数据
	resp.Data = &proto.ListProjectQuotasData{
		Total:   uint32(count),
		Results: pQuotas,
	}

	return nil
}

func (la *ListQuotaV2Action) validate() error {
	if la.req.GetProjectIDOrCode() == "" {
		return errorx.NewParamErr("project id or code is required")
	}

	if la.req.GetQuotaType() == "" {
		return errorx.NewParamErr("quota type is required")
	}

	proj, err := la.model.GetProject(context.TODO(), la.req.GetProjectIDOrCode())
	if err != nil {
		logging.Error("get project from db failed, err: %s", err.Error())
		return errorx.NewDBErr(fmt.Sprintf("get project from db failed,"+
			" req:[%s], err:[%s]", la.req.String(), err.Error()))
	}
	la.project = proj

	return nil
}

func (la *ListQuotaV2Action) getProjectQuotas(ctx context.Context, req *proto.ListProjectQuotasV2Request,
	pQuotas []*proto.ProjectQuota) ([]*proto.ProjectQuota, int64, error) {
	var conds []*operator.Condition
	if req.GetQuotaId() != "" {
		conds = append(conds, operator.NewLeafCondition(operator.Eq, operator.M{
			"quotaId": req.GetQuotaId(),
		}))
	}
	if req.GetQuotaName() != "" {
		conds = append(conds, operator.NewLeafCondition(operator.Eq, operator.M{
			"quotaName": req.GetQuotaName(),
		}))
	}
	if req.GetProjectIDOrCode() != "-" {
		if req.GetProjectIDOrCode() == la.project.ProjectID {
			conds = append(conds, operator.NewLeafCondition(operator.Eq, operator.M{
				"projectId": la.project.ProjectID,
			}))
		} else if req.GetProjectIDOrCode() == la.project.ProjectCode {
			conds = append(conds, operator.NewLeafCondition(operator.Eq, operator.M{
				"projectCode": la.project.ProjectCode,
			}))
		}
	}
	if req.GetBusinessID() != "" {
		conds = append(conds, operator.NewLeafCondition(operator.Eq, operator.M{
			"businessId": req.GetBusinessID(),
		}))
	}
	if req.GetProvider() != "" {
		conds = append(conds, operator.NewLeafCondition(operator.Eq, operator.M{
			"provider": req.GetProvider(),
		}))
	}
	if req.GetQuotaType() != "" {
		conds = append(conds, operator.NewLeafCondition(operator.Eq, operator.M{
			"quotaType": req.GetQuotaType(),
		}))
	}

	conds = append(conds, operator.NewLeafCondition(operator.Ne, operator.M{
		"status": quota.Deleted,
	}))

	cond := operator.NewBranchCondition(operator.And, conds...)

	quotas, total, err := la.model.ListProjectQuotas(ctx, cond, la.paginationOpt())
	if err != nil {
		return nil, 0, errorx.NewDBErr(fmt.Sprintf("ListProjectQuotas failed,"+
			" req:[%s], err:[%s]", req.String(), err.Error()))
	}

	for _, q := range quotas {
		tmp := q
		pQuotas = append(pQuotas, quota.TransStore2ProtoQuota(&tmp))
	}

	return pQuotas, total, nil
}

const (
	listProjectQuotasPageLimit = 100
)

func (la *ListQuotaV2Action) paginationOpt() *page.Pagination {
	pagination := &page.Pagination{All: true}
	if la.req.Page == 0 && la.req.Limit == 0 {
		return pagination
	}

	pagination.All = false
	pagination.Sort = map[string]int{"createtime": -1}

	if la.req.Page > 0 {
		offset := int64(la.req.Page) - 1
		pagination.Offset = offset
	} else {
		pagination.Offset = 0
	}

	if la.req.Limit > 0 {
		pagination.Limit = int64(la.req.Limit)
	} else {
		pagination.Limit = listProjectQuotasPageLimit
	}

	return pagination
}

func (la *ListQuotaV2Action) getQuotasUsage(ctx context.Context, req *proto.ListProjectQuotasV2Request,
	pQuotas []*proto.ProjectQuota) ([]*proto.ProjectQuota, error) {
	if len(pQuotas) == 0 {
		return pQuotas, nil
	}

	type quotaWithIndex struct {
		index int
		quota *proto.ProjectQuota
	}

	var (
		quotaHostList       []*quotaWithIndex
		quotaSelfHostList   []*quotaWithIndex
		quotaFederationList []*quotaWithIndex
	)

	// 按 QuotaType 分类，同时记录原始索引
	for i, pQuota := range pQuotas {
		item := &quotaWithIndex{
			index: i,
			quota: pQuota,
		}
		switch pQuota.QuotaType {
		case quota.Host.String():
			quotaHostList = append(quotaHostList, item)
		case quota.SelfHost.String():
			quotaSelfHostList = append(quotaSelfHostList, item)
		case quota.Federation.String():
			quotaFederationList = append(quotaFederationList, item)
		}
	}

	// 处理 Host 类型的 Usage
	if len(quotaHostList) > 0 {
		var hostQuotas []*proto.ProjectQuota
		for _, item := range quotaHostList {
			hostQuotas = append(hostQuotas, item.quota)
		}
		usagesHostQuota, err := la.doHostUsage(ctx, hostQuotas,
			la.project, req.GetProvider())
		if err != nil {
			return nil, err
		}
		for i, item := range quotaHostList {
			pQuotas[item.index] = usagesHostQuota[i]
		}
	}

	// 处理 SelfHost 类型的 Usage
	if len(quotaSelfHostList) > 0 {
		var selfHostQuotas []*proto.ProjectQuota
		for _, item := range quotaSelfHostList {
			selfHostQuotas = append(selfHostQuotas, item.quota)
		}
		// 目前 SelfHost 类型不支持 Usage
		// usagesSelfHostQuota, err := la.doHostUsage(ctx, selfHostQuotas, la.project, req.GetProvider())
		// if err != nil {
		//	 return nil, err
		// }
		for i, item := range quotaSelfHostList {
			pQuotas[item.index] = selfHostQuotas[i]
		}
	}

	// 处理 Federation 类型的 Usage
	if len(quotaFederationList) > 0 {
		for _, item := range quotaFederationList {
			// Federation 类型需要转换为 store 类型处理
			getQuotaUsageForProto(item.quota)
			pQuotas[item.index] = item.quota
		}
	}

	return pQuotas, nil
}

func (la *ListQuotaV2Action) doHostUsage(ctx context.Context, pquota []*proto.ProjectQuota,
	p *project.Project, provider string) ([]*proto.ProjectQuota, error) {
	if p == nil || provider == "" {
		return pquota, nil
	}
	// 获取指定项目和提供商的资源使用情况
	pqs, errC := clustermanager.GetResourceUsage(ctx, p.ProjectID, provider)

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
		}
	}
	return pquota, nil
}
