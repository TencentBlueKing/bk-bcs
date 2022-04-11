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
func setResp(resp *proto.ProjectResponse, data *pm.Project, perm map[string]interface{}) {
	if data != nil {
		var project proto.Project
		copier.CopyStruct(&project, data)
		resp.Data = &project
	} else {
		resp.Data = nil
	}

	resp.WebAnnotations = &proto.Perms{Perms: convert.Map2pbStruct(perm)}
}

func setListResp(resp *proto.ListProjectsResponse, data *map[string]interface{}, perm map[string]map[string]bool) {
	if data == nil {
		resp.Data = &proto.ListProjectData{Total: 0, Results: []*proto.Project{}}
		return
	}
	if val, ok := (*data)["results"].([]*pm.Project); ok {
		projectData := proto.ListProjectData{Total: (*data)["total"].(uint32)}
		var projects []*proto.Project
		// 组装返回数据
		for i := range val {
			var dstProject proto.Project
			copier.CopyStruct(&dstProject, val[i])
			projects = append(projects, &dstProject)
		}
		projectData.Results = projects
		// 赋值到response
		resp.Data = &projectData
	} else {
		resp.Data = &proto.ListProjectData{Total: 0, Results: []*proto.Project{}}
	}
	resp.WebAnnotations = &proto.Perms{Perms: convert.MapBool2pbStruct(perm)}
}

func setListAuthProjResp(resp *proto.ListAuthorizedProjResp, data *map[string]interface{}) {
	if data == nil {
		resp.Data = &proto.ListProjectData{Total: 0, Results: []*proto.Project{}}
		return
	}
	if val, ok := (*data)["results"].([]pm.Project); ok {
		projectData := proto.ListProjectData{Total: (*data)["total"].(uint32)}
		var projects []*proto.Project
		// 组装返回数据
		for i := range val {
			var dstProject proto.Project
			copier.CopyStruct(&dstProject, val[i])
			projects = append(projects, &dstProject)
		}
		projectData.Results = projects
		// 赋值到response
		resp.Data = &projectData
	} else {
		resp.Data = &proto.ListProjectData{Total: 0, Results: []*proto.Project{}}
	}
}
