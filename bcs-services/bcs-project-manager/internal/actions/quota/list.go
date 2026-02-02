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

	common "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/quota"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/convert"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
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
		if errQ != nil {
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
	pqs, errC := clustermanager.GetResourceUsage(ctx, req.ProjectID, req.Provider, quota.Host.String())

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

	quotas, _, err := la.model.ListProjectQuotas(la.ctx, cond, &page.Pagination{All: true})
	if err != nil {
		return pquota, errorx.NewDBErr(err.Error())
	}

	for _, q := range quotas {
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
	count   uint32
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

	err := la.validate()
	if err != nil {
		return err
	}

	var PQ []*proto.ProjectQuota
	pQuotas, count, err := la.getProjectQuotas(ctx, req, PQ)
	if err != nil {
		return err
	}
	la.count = uint32(count)

	pQuotas, err = la.getQuotasUsage(ctx, req, pQuotas)
	if err != nil {
		return err
	}

	// 设置响应数据
	resp.Data = &proto.ListProjectQuotasData{
		Total:   la.count,
		Results: pQuotas,
	}

	return nil
}

// validate validate
func (la *ListQuotaV2Action) validate() error {
	if la.req.GetProjectIDOrCode() == "" {
		return errorx.NewParamErr("project id or code is required")
	}

	if la.req.GetQuotaType() == "" {
		return errorx.NewParamErr("quota type is required")
	}

	if la.req.Page <= 0 {
		la.req.Page = 1
	}

	if la.req.Limit <= 0 {
		la.req.Limit = page.DefaultPageLimit
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

// getProjectQuotas 从数据库获取项目配额信息
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
	if req.GetProjectIDOrCode() != "" {
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

	quotas, total, err := la.model.ListProjectQuotas(ctx, cond, la.getPaginationOpt())
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

// getPaginationOpt 获取分页参数
func (la *ListQuotaV2Action) getPaginationOpt() *page.Pagination {
	pagination := &page.Pagination{All: false}
	pagination.Sort = map[string]int{"createtime": -1}

	offset := int64(la.req.Page - 1)
	pagination.Offset = offset
	pagination.Limit = int64(la.req.Limit)

	return pagination
}

// getQuotasUsage 获取配额使用情况
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

	// 处理 Host 类型的 Usage，quota-gray 与 V1 保持一致
	if _, ok := la.project.Labels["quota-gray"]; ok {
		if len(quotaHostList) > 0 {
			var hostQuotas []*proto.ProjectQuota
			for _, item := range quotaHostList {
				hostQuotas = append(hostQuotas, item.quota)
			}
			usagesHostQuota, err := doHostUsage(ctx, hostQuotas,
				la.project, req.GetProvider())
			if err != nil {
				return nil, err
			}
			for i, item := range quotaHostList {
				pQuotas[item.index] = usagesHostQuota[i]
			}
		}
	} else if req.GetQuotaType() == quota.Host.String() {
		usagesHostQuota, err := doHostUsage(ctx, make([]*proto.ProjectQuota, 0),
			la.project, req.GetProvider())
		if err != nil {
			return nil, err
		}
		newPQuotas := make([]*proto.ProjectQuota, 0)
		newPQuotas = append(newPQuotas, usagesHostQuota...)
		pQuotas = newPQuotas
		la.count = uint32(len(pQuotas))
	}

	// 处理 SelfHost 类型的 Usage
	if len(quotaSelfHostList) > 0 {
		var selfHostQuotas []*proto.ProjectQuota
		for _, item := range quotaSelfHostList {
			selfHostQuotas = append(selfHostQuotas, item.quota)
		}
		usagesSelfHostQuota, err := doSelfHostUsage(ctx, selfHostQuotas, la.project, req.GetProvider())
		if err != nil {
			return nil, err
		}
		for i, item := range quotaSelfHostList {
			pQuotas[item.index] = usagesSelfHostQuota[i]
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

// doHostUsage 获取 host 类型使用量
func doHostUsage(ctx context.Context, pquota []*proto.ProjectQuota,
	p *project.Project, provider string) ([]*proto.ProjectQuota, error) {
	if p == nil || provider == "" {
		return pquota, nil
	}
	// 获取指定项目和提供商的资源使用情况
	pqs, errC := clustermanager.GetResourceUsage(ctx, p.ProjectID, provider, quota.Host.String())

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
				ProjectID:  p.ProjectID,
				Status:     string(quota.Running),
			})
		}
	}
	return pquota, nil
}

// doSelfHostUsage 获取 self_host 类型使用量
func doSelfHostUsage(ctx context.Context, pquota []*proto.ProjectQuota,
	p *project.Project, provider string) ([]*proto.ProjectQuota, error) {
	if p == nil || provider == "" {
		return pquota, nil
	}
	// 获取指定项目和提供商的资源使用情况
	pqs, errC := clustermanager.GetResourceUsage(ctx, p.ProjectID, provider, quota.SelfHost.String())

	if errC != nil {
		return pquota, errC
	}

	// 遍历每个项目配额，构建响应数据
	for _, pqpq := range pquota {
		for _, pq := range pqs {
			if pqpq.QuotaId == pq.GetQuotaID() && pqpq.Status == string(quota.Running) {
				var NG []*proto.NodeGroup
				// 获取每个节点组的信息
				for _, gpid := range pq.TotalGroupIds {
					ng, errG := clustermanager.GetNodeGroup(ctx, gpid)
					if errG != nil {
						return pquota, errG
					}
					NG = append(NG, &proto.NodeGroup{
						NodeGroupId: ng.NodeGroupID,
						ClusterId:   ng.ClusterID,
					})
				}
				pqpq.Quota.ZoneResources.QuotaUsed = pq.Used
				pqpq.NodeGroups = NG
			}
		}

	}
	return pquota, nil
}

// GetQuotaStatisticsAction action for get statistics
type GetQuotaStatisticsAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.GetProjectQuotasStatisticsRequest
	resp  *proto.GetProjectQuotasStatisticsResponse

	project *project.Project
}

// NewGetQuotaStatisticsAction new get statistics action
func NewGetQuotaStatisticsAction(model store.ProjectModel) *GetQuotaStatisticsAction {
	return &GetQuotaStatisticsAction{
		model: model,
	}
}

// Do get project statistics info
func (ga *GetQuotaStatisticsAction) Do(ctx context.Context, req *proto.GetProjectQuotasStatisticsRequest,
	resp *proto.GetProjectQuotasStatisticsResponse) error {
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	err := ga.validate()
	if err != nil {
		return err
	}

	pQuotas, _, err := ga.getProjectQuotas(ctx, req)
	if err != nil {
		return err
	}

	usageProjectQuotas, err := ga.getQuotasUsage(ctx, req, pQuotas)
	if err != nil {
		return err
	}

	statisticsData := ga.getQuotasStatistics(usageProjectQuotas)

	ga.resp.Data = statisticsData

	return nil
}

// validate validate
func (ga *GetQuotaStatisticsAction) validate() error {
	if ga.req.GetProjectIDOrCode() == "" {
		return errorx.NewParamErr("project id or code is required")
	}

	proj, err := ga.model.GetProject(context.TODO(), ga.req.GetProjectIDOrCode())
	if err != nil {
		logging.Error("get project from db failed, err: %s", err.Error())
		return errorx.NewDBErr(fmt.Sprintf("get project from db failed,"+
			" req:[%s], err:[%s]", ga.req.String(), err.Error()))
	}
	ga.project = proj

	return nil
}

// getProjectQuotas 从数据库获取项目配额
func (ga *GetQuotaStatisticsAction) getProjectQuotas(ctx context.Context,
	req *proto.GetProjectQuotasStatisticsRequest) ([]*proto.ProjectQuota, int64, error) {
	var conds, sharedConds []*operator.Condition

	// 构建基础查询条件
	conds = ga.buildBaseConditions(req, conds)

	finalCond := operator.NewBranchCondition(operator.And, conds...)

	// 如果包含共享配额，增加查询共享配额条件
	if ga.req.GetIsContainShared() {
		sharedConds = ga.buildSharedConditions(req, sharedConds)

		if len(sharedConds) > 0 {
			sharedQuotaCond := operator.NewBranchCondition(operator.And, sharedConds...)
			finalCond = operator.NewBranchCondition(operator.Or, finalCond, sharedQuotaCond)
		}
	}

	quotas, total, err := ga.model.ListProjectQuotas(ctx, finalCond, &page.Pagination{All: true})
	if err != nil {
		return nil, 0, errorx.NewDBErr(fmt.Sprintf("GetQuotaStatistics ListProjectQuotas failed,"+
			" req:[%s], err:[%s]", req.String(), err.Error()))
	}

	var pQuotas []*proto.ProjectQuota

	for _, q := range quotas {
		tmp := q
		pQuotas = append(pQuotas, quota.TransStore2ProtoQuota(&tmp))
	}

	return pQuotas, total, nil
}

// buildBaseConditions 构建基础查询条件
func (ga *GetQuotaStatisticsAction) buildBaseConditions(req *proto.GetProjectQuotasStatisticsRequest,
	conds []*operator.Condition) []*operator.Condition {

	if req.GetProjectIDOrCode() != "" {
		projectCond := operator.NewBranchCondition(operator.Or,
			operator.NewLeafCondition(operator.Eq, operator.M{"projectId": req.GetProjectIDOrCode()}),
			operator.NewLeafCondition(operator.Eq, operator.M{"projectCode": req.GetProjectIDOrCode()}),
		)
		conds = append(conds, projectCond)
	}

	conds = ga.buildCommonConditions(req, conds)

	return conds
}

// buildCommonConditions 构建共同查询条件
func (ga *GetQuotaStatisticsAction) buildCommonConditions(req *proto.GetProjectQuotasStatisticsRequest,
	conds []*operator.Condition) []*operator.Condition {
	if req.GetQuotaType() != "" {
		conds = append(conds, operator.NewLeafCondition(operator.Eq, operator.M{
			"quotaType": req.GetQuotaType(),
		}))
	} else {
		quotaTypes := []string{
			quota.Host.String(),
			quota.SelfHost.String(),
			quota.Federation.String(),
		}
		conds = append(conds, operator.NewLeafCondition(operator.In, operator.M{
			"quotaType": quotaTypes,
		}))
	}

	conds = append(conds, operator.NewLeafCondition(operator.Eq, operator.M{
		"provider": common.ProviderInternal,
	}))
	conds = append(conds, operator.NewLeafCondition(operator.Ne, operator.M{
		"status": quota.Deleted,
	}))

	return conds
}

// buildSharedConditions 构建共享配额查询条件
func (ga *GetQuotaStatisticsAction) buildSharedConditions(req *proto.GetProjectQuotasStatisticsRequest,
	conds []*operator.Condition) []*operator.Condition {

	if req.GetProjectIDOrCode() != "" {
		projectCond := operator.NewBranchCondition(operator.Or,
			operator.NewLeafCondition(operator.Eq, operator.M{"quotaSharedProjectList.projectId": req.GetProjectIDOrCode()}),
			operator.NewLeafCondition(operator.Eq, operator.M{"quotaSharedProjectList.projectCode": req.GetProjectIDOrCode()}),
		)
		conds = append(conds, projectCond)
	}

	conds = append(conds, operator.NewLeafCondition(operator.Eq, operator.M{
		"quotaSharedEnabled": true,
	}))

	conds = ga.buildCommonConditions(req, conds)

	return conds
}

// getQuotasUsage 获取配额使用情况
func (ga *GetQuotaStatisticsAction) getQuotasUsage(ctx context.Context, req *proto.GetProjectQuotasStatisticsRequest,
	pQuotas []*proto.ProjectQuota) ([]*proto.ProjectQuota, error) {
	if len(pQuotas) == 0 {
		return pQuotas, nil
	}

	var (
		quotaUsageList []*proto.ProjectQuota

		quotaHostList       []*proto.ProjectQuota
		quotaSelfHostList   []*proto.ProjectQuota
		quotaFederationList []*proto.ProjectQuota
	)

	// 按 QuotaType 分类，同时记录原始索引
	for _, pQuota := range pQuotas {
		switch pQuota.QuotaType {
		case quota.Host.String():
			quotaHostList = append(quotaHostList, pQuota)
		case quota.SelfHost.String():
			quotaSelfHostList = append(quotaSelfHostList, pQuota)
		case quota.Federation.String():
			quotaFederationList = append(quotaFederationList, pQuota)
		}
	}

	// 处理 Host 类型的 Usage, quota-gray 保持原数据查询信息结果
	if req.GetQuotaType() == quota.Host.String() {
		if _, ok := ga.project.Labels["quota-gray"]; ok {
			if len(quotaHostList) > 0 {
				usagesHostQuota, err := doHostUsage(ctx, quotaHostList,
					ga.project, common.ProviderInternal)
				if err != nil {
					return nil, err
				}
				quotaUsageList = append(quotaUsageList, usagesHostQuota...)
			}
		} else {
			usagesHostQuota, err := doHostUsage(ctx, make([]*proto.ProjectQuota, 0),
				ga.project, common.ProviderInternal)
			if err != nil {
				return nil, err
			}
			quotaUsageList = append(quotaUsageList, usagesHostQuota...)
		}
	}

	// 处理 SelfHost 类型的 Usage
	if len(quotaSelfHostList) > 0 {
		usagesSelfHostQuota, err := doSelfHostUsage(ctx, quotaSelfHostList,
			ga.project, common.ProviderInternal)
		if err != nil {
			return nil, err
		}
		quotaUsageList = append(quotaUsageList, usagesSelfHostQuota...)
	}

	// 处理 Federation 类型的 Usage
	if len(quotaFederationList) > 0 {
		for _, item := range quotaFederationList {
			// Federation 类型需要转换为 store 类型处理
			getQuotaUsageForProto(item)
			quotaUsageList = append(quotaUsageList, item)
		}
	}

	return quotaUsageList, nil
}

func (ga *GetQuotaStatisticsAction) getQuotasStatistics(
	pQuotas []*proto.ProjectQuota) *proto.ProjectQuotasStatisticsData {
	if len(pQuotas) == 0 {
		return &proto.ProjectQuotasStatisticsData{}
	}

	var (
		statisticsData = &proto.ProjectQuotasStatisticsData{
			Cpu: &proto.QuotaResourceData{},
			Mem: &proto.QuotaResourceData{},
			Gpu: &proto.QuotaResourceData{},
		}
	)

	for _, pQuota := range pQuotas {
		switch pQuota.QuotaType {
		case quota.Host.String(), quota.SelfHost.String():
			// Host和SelfHost类型处理相同的逻辑
			zoneResource := pQuota.GetQuota().GetZoneResources()
			if zoneResource != nil {
				ga.handleHostStatisticsData(statisticsData, zoneResource)
			}
		case quota.Federation.String():
			// Federation类型处理，只计算 Running 状态
			if pQuota.Status == quota.Running.String() {
				ga.handleFederationStatisticsData(statisticsData,
					pQuota.GetQuota().GetCpu(),
					pQuota.GetQuota().GetMem(),
					pQuota.GetQuota().GetGpu())
			}
		}
	}

	return statisticsData
}

// handleHostStatisticsData 处理 host/self_host 类型的统计数据
func (ga *GetQuotaStatisticsAction) handleHostStatisticsData(
	statisticsData *proto.ProjectQuotasStatisticsData,
	zoneResource *proto.InstanceTypeConfig) *proto.ProjectQuotasStatisticsData {
	if zoneResource != nil {
		statisticsData.Cpu.TotalNum += zoneResource.Cpu * zoneResource.QuotaNum
		statisticsData.Mem.TotalNum += zoneResource.Mem * zoneResource.QuotaNum
		statisticsData.Gpu.TotalNum += zoneResource.Gpu * zoneResource.QuotaNum

		statisticsData.Cpu.UsedNum += zoneResource.Cpu * zoneResource.QuotaUsed
		statisticsData.Mem.UsedNum += zoneResource.Mem * zoneResource.QuotaUsed
		statisticsData.Gpu.UsedNum += zoneResource.Gpu * zoneResource.QuotaUsed

		statisticsData.Cpu.AvailableNum = statisticsData.Cpu.TotalNum - statisticsData.Cpu.UsedNum
		statisticsData.Mem.AvailableNum = statisticsData.Mem.TotalNum - statisticsData.Mem.UsedNum
		statisticsData.Gpu.AvailableNum = statisticsData.Gpu.TotalNum - statisticsData.Gpu.UsedNum

		if statisticsData.Cpu.TotalNum != 0 {
			statisticsData.Cpu.UseRate = convert.RoundToTwoDecimals(
				convert.RoundDivisionToTwoDecimals(statisticsData.Cpu.UsedNum, statisticsData.Cpu.TotalNum) * 100)
		}
		if statisticsData.Mem.TotalNum != 0 {
			statisticsData.Mem.UseRate = convert.RoundToTwoDecimals(
				convert.RoundDivisionToTwoDecimals(statisticsData.Mem.UsedNum, statisticsData.Mem.TotalNum) * 100)
		}
		if statisticsData.Gpu.TotalNum != 0 {
			statisticsData.Gpu.UseRate = convert.RoundToTwoDecimals(
				convert.RoundDivisionToTwoDecimals(statisticsData.Gpu.UsedNum, statisticsData.Gpu.TotalNum) * 100)
		}
	}

	return statisticsData
}

// handleFederationStatisticsData 处理 federation 类型的统计数据
func (ga *GetQuotaStatisticsAction) handleFederationStatisticsData(
	statisticsData *proto.ProjectQuotasStatisticsData,
	cpu, mem, gpu *proto.DeviceInfo) *proto.ProjectQuotasStatisticsData {
	if cpu != nil {
		statisticsData.Cpu.TotalNum += stringx.StringToUint32(cpu.GetDeviceQuota())
		statisticsData.Cpu.UsedNum += stringx.StringToUint32(cpu.GetDeviceQuotaUsed())
		statisticsData.Cpu.AvailableNum = statisticsData.Cpu.TotalNum - statisticsData.Cpu.UsedNum
		if statisticsData.Cpu.TotalNum != 0 {
			statisticsData.Cpu.UseRate = convert.RoundToTwoDecimals(
				convert.RoundDivisionToTwoDecimals(statisticsData.Cpu.UsedNum, statisticsData.Cpu.TotalNum) * 100)
		}
	}
	if mem != nil {
		statisticsData.Mem.TotalNum += stringx.StringToUint32(mem.GetDeviceQuota())
		statisticsData.Mem.UsedNum += stringx.StringToUint32(mem.GetDeviceQuotaUsed())
		statisticsData.Mem.AvailableNum = statisticsData.Mem.TotalNum - statisticsData.Mem.UsedNum
		if statisticsData.Mem.TotalNum != 0 {
			statisticsData.Mem.UseRate = convert.RoundToTwoDecimals(
				convert.RoundDivisionToTwoDecimals(statisticsData.Mem.UsedNum, statisticsData.Mem.TotalNum) * 100)
		}
	}
	if gpu != nil {
		statisticsData.Gpu.TotalNum += stringx.StringToUint32(gpu.GetDeviceQuota())
		statisticsData.Gpu.UsedNum += stringx.StringToUint32(gpu.GetDeviceQuotaUsed())
		statisticsData.Gpu.AvailableNum = statisticsData.Gpu.TotalNum - statisticsData.Gpu.UsedNum
		if statisticsData.Gpu.TotalNum != 0 {
			statisticsData.Gpu.UseRate = convert.RoundToTwoDecimals(
				convert.RoundDivisionToTwoDecimals(statisticsData.Gpu.UsedNum, statisticsData.Gpu.TotalNum) * 100)
		}
	}
	return statisticsData
}
