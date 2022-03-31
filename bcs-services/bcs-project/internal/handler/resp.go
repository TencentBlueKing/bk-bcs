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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project/proto/bcsproject"
)

// setResp 设置单项目数据的返回
func setResp(resp *proto.ProjectResponse, code uint32, msg string, data interface{}) {
	resp.Code = code
	resp.Message = msg
	// 处理数据
	if val, ok := data.(*pm.Project); ok {
		var project proto.Project
		util.CopyStruct(&project, val)
		resp.Data = &project
	} else {
		resp.Data = nil
	}
}

// setListResp 设置多个项目数据的返回
func setListResp(resp *proto.ListProjectsResponse, code uint32, msg string, data *map[string]interface{}) {
	resp.Code = code
	resp.Message = msg
	if val, ok := (*data)["results"].([]*pm.Project); ok {
		projectData := proto.ListProjectData{Total: (*data)["total"].(uint32)}
		var projects []*proto.Project
		// 组装返回数据
		for i := range val {
			var dstProject proto.Project
			util.CopyStruct(&dstProject, val[i])
			projects = append(projects, &dstProject)
		}
		projectData.Results = projects
		// 赋值到response
		resp.Data = &projectData
	} else {
		resp.Data = nil
	}
}
