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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project/proto/bcsproject"
)

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

func (la *ListAction) ProjectList(ctx context.Context, req *proto.ListProjectsRequest, resp *proto.ListProjectsResponse) {
	if req == nil || resp == nil {
		return
	}
	la.ctx = ctx
	la.req = req

	if err := req.Validate(); err != nil {
		setListResp(resp, common.BcsProjectParamErr, common.BcsProjectParamErrMsg, err.Error(), nil)
		return
	}
	projects, total, err := la.listProjects()
	if err != nil {
		setListResp(resp, common.BcsProjectDbErr, common.BcsProjectDbErrMsg, err.Error(), nil)
		return
	}
	data := proto.ListProjectData{
		Total:   uint32(total),
		Results: projects,
	}
	setListResp(resp, common.BcsProjectSuccess, "", common.BcsProjectSuccessMsg, &data)
	return
}

func (la *ListAction) listProjects() ([]*proto.Project, int64, error) {
	condM := make(operator.M)

	var cond *operator.Condition
	// 通过项目名称进行模糊查询
	if la.req.SearchName != "" {
		condM["name"] = la.req.SearchName
		cond = operator.NewLeafCondition(operator.Con, condM)
	} else {
		if la.req.ProjectIDs != "" {
			condM["projectID"] = util.SplitString(la.req.ProjectIDs)
		}
		if la.req.Names != "" {
			condM["name"] = util.SplitString(la.req.Names)
		}
		if la.req.EnglishNames != "" {
			condM["englishname"] = util.SplitString(la.req.EnglishNames)
		}
		if la.req.Kind != "" {
			condM["kind"] = []string{la.req.Kind}
		}
		cond = operator.NewLeafCondition(operator.In, condM)
	}

	// 查询项目信息
	projects, total, err := la.model.ListProjects(la.ctx, cond, &common.Pagination{
		Limit: la.req.Limit, Offset: la.req.Offset, All: la.req.All,
	})
	if err != nil {
		return nil, total, err
	}
	projectList := []*proto.Project{}
	for i := range projects {
		projectList = append(projectList, &projects[i])
	}
	return projectList, total, nil
}
