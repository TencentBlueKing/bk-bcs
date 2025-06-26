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

package project

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bkmonitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/tenant"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// ListForIAMAction xxx
type ListForIAMAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.ListProjectsForIAMReq
}

// NewListForIAMActionAction new list projects for iam action
func NewListForIAMActionAction(model store.ProjectModel) *ListForIAMAction {
	return &ListForIAMAction{
		model: model,
	}
}

// Do xxx
func (la *ListForIAMAction) Do(ctx context.Context, req *proto.ListProjectsForIAMReq) (
	[]*proto.ListProjectsForIAMResp_Project, error) {
	la.ctx = ctx
	la.req = req

	projects, _, err := la.listProjects()
	if err != nil {
		return nil, errorx.NewDBErr(err.Error())
	}

	spaces, err := bkmonitor.ListSpaces(ctx)
	if err != nil {
		return nil, err
	}
	spaceMap := make(map[string]*bkmonitor.Space)
	for _, space := range spaces {
		spaceMap[space.SpaceCode] = space
	}

	filtered := []*proto.ListProjectsForIAMResp_Project{}
	for _, project := range projects {
		if space, ok := spaceMap[project.ProjectID]; ok {
			filtered = append(filtered, &proto.ListProjectsForIAMResp_Project{
				Name:        project.Name,
				ProjectID:   project.ProjectID,
				ProjectCode: project.ProjectCode,
				BusinessID:  project.BusinessID,
				Managers:    project.Managers,
				// 监控BCS项目空间ID注册到权限中心的资源ID需要取负数
				BkmSpaceBizID: int32(-space.ID),
				BkmSpaceName:  space.DisplayName,
			})
		}
	}
	lfiConf := config.GlobalConf.ListForIAM
	if !lfiConf.All {
		filtered = la.filterByBizs(filtered, lfiConf.Bizs)
	}
	return filtered, nil
}

func (la *ListForIAMAction) listProjects() ([]*pm.Project, int64, error) {

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"kind": "k8s",
	})
	if tenant.IsMultiTenantEnabled() {
		tenantCond := operator.NewLeafCondition(operator.Eq, operator.M{"tenantId": tenant.GetTenantIdFromContext(la.ctx)})
		cond = operator.NewBranchCondition(operator.And, cond, tenantCond)
	}

	// 查询所有开启了容器服务的项目
	projects, total, err := la.model.ListProjects(la.ctx, cond, &page.Pagination{All: true})
	if err != nil {
		return nil, total, err
	}
	projectList := []*pm.Project{}
	for i := range projects {
		projectList = append(projectList, &projects[i])
	}
	return projectList, total, nil
}

func (la *ListForIAMAction) filterByBizs(
	projects []*proto.ListProjectsForIAMResp_Project, bizs []string) []*proto.ListProjectsForIAMResp_Project {
	bizMap := make(map[string]bool)
	for _, biz := range bizs {
		bizMap[biz] = true
	}
	filtered := []*proto.ListProjectsForIAMResp_Project{}
	for _, project := range projects {
		if _, ok := bizMap[project.BusinessID]; ok {
			filtered = append(filtered, project)
		}
	}
	return filtered
}
