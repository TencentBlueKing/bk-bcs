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
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// ListAction xxx
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

// Do xxx
func (la *ListAction) Do(ctx context.Context, req *proto.ListProjectsRequest) (*map[string]interface{}, error) {
	la.ctx = ctx
	la.req = req

	projects, total, err := la.listProjects()
	if err != nil {
		return nil, errorx.NewDBErr(err.Error())
	}
	data := map[string]interface{}{
		"total":   uint32(total),
		"results": projects,
	}
	return &data, nil
}

func (la *ListAction) listProjects() ([]*pm.Project, int64, error) {

	var cond *operator.Condition
	// 通过项目名称进行模糊查询
	if la.req.SearchName != "" {
		condName := operator.NewLeafCondition(operator.Con,
			operator.M{"name": primitive.Regex{Pattern: la.req.SearchName, Options: "i"}})
		condProjectCode := operator.NewLeafCondition(operator.Con,
			operator.M{"projectCode": primitive.Regex{Pattern: la.req.SearchName, Options: "i"}})
		cond = operator.NewBranchCondition(operator.Or, condName, condProjectCode)
	} else {
		condM := make(operator.M)
		if la.req.ProjectIDs != "" {
			condM["projectID"] = stringx.SplitString(la.req.ProjectIDs)
		}
		if la.req.Names != "" {
			condM["name"] = stringx.SplitString(la.req.Names)
		}
		if la.req.ProjectCode != "" {
			condM["projectCode"] = stringx.SplitString(la.req.ProjectCode)
		}
		if la.req.Kind != "" {
			condM["kind"] = stringx.SplitString(la.req.Kind)
		}
		if la.req.BusinessID != "" {
			condM["businessID"] = stringx.SplitString(la.req.GetBusinessID())
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

// ListAuthorizedProject xxx
type ListAuthorizedProject struct {
	model store.ProjectModel
}

// NewListAuthorizedProj new list authorized project action
func NewListAuthorizedProj(model store.ProjectModel) *ListAuthorizedProject {
	return &ListAuthorizedProject{
		model: model,
	}
}

// Do xxx
func (lap *ListAuthorizedProject) Do(ctx context.Context,
	req *proto.ListAuthorizedProjReq) (*map[string]interface{}, error) {
	var projects []pm.Project
	var total int64
	authUser, err := middleware.GetUserFromContext(ctx)
	if err == nil && authUser.Username != "" {
		// username 为空时，该接口请求没有意义
		ids, any, err := auth.ListAuthorizedProjectIDs(authUser.Username)
		if err != nil {
			logging.Error("get user project permissions failed, err: %s", err.Error())
			return nil, nil
		}
		if req.All {
			if config.GlobalConf.RestrictAuthorizedProjects {
				// all 为 true 且限权显示用户授权的项目列表时，仅查看用户有权限的项目，并支持模糊查询和分页
				if any {
					projects, total, err = lap.model.SearchProjects(ctx, ids, nil, req.SearchKey, req.Kind,
						&page.Pagination{Offset: req.Offset, Limit: req.Limit})
				} else {
					projects, total, err = lap.model.SearchProjects(ctx, nil, ids, req.SearchKey, req.Kind,
						&page.Pagination{Offset: req.Offset, Limit: req.Limit})
				}
			} else {
				// all 为 true 且不限权时，返回所有项目并排序和分页，支持模糊查询
				projects, total, err = lap.model.SearchProjects(ctx, ids, nil, req.SearchKey, req.Kind,
					&page.Pagination{Offset: req.Offset, Limit: req.Limit})
			}
		} else {
			// all 为 false 且用户没有全部项目查看权限时，返回用户有权限的项目，模糊查询都无效
			var cond *operator.Condition
			condKind := make(operator.M)
			if req.Kind != "" {
				condKind["kind"] = req.Kind
			}
			if any {
				cond = operator.NewBranchCondition(operator.And, operator.NewLeafCondition(operator.Eq, condKind))
			} else {
				condID := make(operator.M)
				condID["projectID"] = ids
				cond = operator.NewBranchCondition(operator.And,
					operator.NewLeafCondition(operator.In, condID), operator.NewLeafCondition(operator.Eq, condKind))
			}
			pagination := &page.Pagination{All: false}
			if req.Limit == 0 && req.Offset == 0 {
				pagination.All = true
			} else {
				pagination.Offset = req.Offset
				pagination.Limit = req.Limit
			}
			projects, total, err = lap.model.ListProjects(ctx, cond, pagination)
		}
		if err != nil {
			return nil, err
		}
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
