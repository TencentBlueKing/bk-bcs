/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package project

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/perm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project/proto/bcsproject"
)

// ListAction ...
type ListAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.ListProjectsRequest
}

// NewListAction new list project action
func NewListAction(model store.ProjectModel) *ListAction {
	return &ListAction{
		model: model,
	}
}

// Do ...
func (la *ListAction) Do(ctx context.Context, req *proto.ListProjectsRequest) (*map[string]interface{}, error) {
	la.ctx = ctx
	la.req = req

	projects, total, err := la.listProjects()
	if err != nil {
		return nil, errorx.NewDBErr(err)
	}
	data := map[string]interface{}{
		"total":   uint32(total),
		"results": projects,
	}
	return &data, nil
}

func (la *ListAction) listProjects() ([]*pm.Project, int64, error) {
	condM := make(operator.M)

	var cond *operator.Condition
	// 通过项目名称进行模糊查询
	if la.req.SearchName != "" {
		condM["name"] = primitive.Regex{Pattern: la.req.SearchName, Options: "i"}
		cond = operator.NewLeafCondition(operator.Con, condM)
	} else {
		if la.req.ProjectIDs != "" {
			condM["projectID"] = stringx.SplitString(la.req.ProjectIDs)
		}
		if la.req.Names != "" {
			condM["name"] = stringx.SplitString(la.req.Names)
		}
		if la.req.ProjectCode != "" {
			condM["projectcode"] = stringx.SplitString(la.req.ProjectCode)
		}
		if la.req.Kind != "" {
			condM["kind"] = []string{la.req.Kind}
		}
		cond = operator.NewLeafCondition(operator.In, condM)
	}

	// 查询项目信息
	projects, total, err := la.model.ListProjects(la.ctx, cond, &page.Pagination{
		Limit: la.req.Limit, Offset: la.req.Offset, All: la.req.All,
	})
	if err != nil {
		return nil, total, err
	}
	projectList := []*pm.Project{}
	for i := range projects {
		projectList = append(projectList, &projects[i])
	}
	return projectList, total, nil
}

// ListAuthorizedProject ...
type ListAuthorizedProject struct {
	model store.ProjectModel
}

// NewListAuthorizedProj new list authorized project action
func NewListAuthorizedProj(model store.ProjectModel) *ListAuthorizedProject {
	return &ListAuthorizedProject{
		model: model,
	}
}

// Do ...
func (lap *ListAuthorizedProject) Do(ctx context.Context, req *proto.ListAuthorizedProjReq) (*map[string]interface{}, error) {
	username := auth.GetUserFromCtx(ctx)
	ids, err := perm.ListAuthorizedProjectIDs(username)
	// 没有权限的项目时，返回为空，并记录信息
	if ids == nil || err != nil {
		logging.Error("%s no permission project", username)
		return nil, nil
	}
	// 通过 project id 获取项目详情
	projects, total, err := lap.model.ListProjectByIDs(ctx, ids, &page.Pagination{All: true})
	if err != nil {
		return nil, err
	}

	projectList := []*pm.Project{}
	for i := range projects {
		projectList = append(projectList, &projects[i])
	}
	data := map[string]interface{}{
		"total":   uint32(total),
		"results": projectList,
	}

	return &data, nil
}
