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

package handler

import (
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/convert"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/copier"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project/proto/bcsproject"
)

// setResp 设置返回的数据和权限信息
func setResp(resp *proto.ProjectResponse, data *pm.Project) {
	if data != nil {
		var project proto.Project
		copier.CopyStruct(&project, data)
		resp.Data = &project
	} else {
		resp.Data = nil
	}
}

// setListResp 设置列表数据的返回
func setListResp(resp interface{}, data *map[string]interface{}) {
	// 返回
	if listProjectResp, ok := resp.(*proto.ListProjectsResponse); ok {
		listProjectResp.Data = getProjectData(data)
		return
	}
	if listAuthProjectResp, ok := resp.(*proto.ListAuthorizedProjResp); ok {
		listAuthProjectResp.Data = getProjectData(data)
		return
	}
}

// setListPermsResp 添加权限信息
func setListPermsResp(resp interface{}, data *map[string]interface{}, perm map[string]map[string]bool) {
	setListResp(resp, data)
	// NOTE: 当根据条件查询项目信息时，带上项目对应的权限
	if listProjectResp, ok := resp.(*proto.ListProjectsResponse); ok {
		listProjectResp.WebAnnotations = &proto.Perms{Perms: convert.MapBool2pbStruct(perm)}
	}
}

func getProjectData(d *map[string]interface{}) *proto.ListProjectData {
	if d == nil {
		return &proto.ListProjectData{Total: 0, Results: []*proto.Project{}}
	}
	if val, ok := (*d)["results"].([]*pm.Project); ok {
		projectData := proto.ListProjectData{Total: (*d)["total"].(uint32)}
		var projects []*proto.Project
		// 组装返回数据
		for i := range val {
			var dstProject proto.Project
			copier.CopyStruct(&dstProject, val[i])
			projects = append(projects, &dstProject)
		}
		projectData.Results = projects
		return &projectData
	}
	return &proto.ListProjectData{Total: 0, Results: []*proto.Project{}}
}
