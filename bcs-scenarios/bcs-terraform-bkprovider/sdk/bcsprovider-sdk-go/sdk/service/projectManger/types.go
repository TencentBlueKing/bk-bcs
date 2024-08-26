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

// Package projectManger project-service
package projectManger

import "github.com/golang/protobuf/jsonpb"

const (
	// createProjectApi post
	createProjectApi = "/bcsproject/v1/projects"

	// deleteProjectApi delete (projectID)
	deleteProjectApi = "/bcsproject/v1/projects/%s"

	// updateProjectApi put (projectID)
	updateProjectApi = "/bcsproject/v1/projects/%s"

	// getProjectApi get (projectIDOrCode)
	getProjectApi = "/bcsproject/v1/projects/%s"

	// listProjectsApi get
	listProjectsApi = "/bcsproject/v1/projects"

	// projectCodeRegexp 正则语句
	projectCodeRegexp = `^[a-z][a-z0-9-]*$`
)

// CreateProjectRequest body
//
// ProjectCode 项目编码(英文缩写), 全局唯一, 长度不能超过64字符
// ***不能为空***
//
// Name 项目中文名称, 长度不能超过64字符
// ***不能为空***
//
// BusinessID 项目绑定的蓝鲸CMDB中业务ID信息
// ***不能为空***
type CreateProjectRequest struct {
	// RequestID 请求ID
	RequestID string

	// ProjectCode 项目编码(英文缩写), 全局唯一, 长度不能超过64字符
	// ***不能为空***
	ProjectCode string `json:"projectCode"`

	// Name 项目中文名称, 长度不能超过64字符
	// ***不能为空***
	Name string `json:"name,omitempty"`

	// BusinessID 项目绑定的蓝鲸CMDB中业务ID信息
	// ***不能为空***
	BusinessID string `json:"businessID,omitempty"`

	// Description 项目描述, 尽量限制在100字符
	Description string `json:"description,omitempty"`
}

// DeleteProjectRequest  body
//
// ProjectID 项目ID
// ***不能为空***
type DeleteProjectRequest struct {
	// RequestID 请求ID
	RequestID string

	// ProjectID 项目ID
	// ***不能为空***
	ProjectID string `json:"projectID"`
}

// UpdateProjectRequest update body.
//
// ProjectID 项目ID
// ***不能为空***
//
// Name 项目中文名称, 长度不能超过64字符
// ***不能为空***
type UpdateProjectRequest struct {
	// RequestID 请求ID
	RequestID string

	// ProjectID 项目ID
	// ***不能为空***
	ProjectID string `json:"projectID"`

	// Name 项目中文名称, 长度不能超过64字符
	// ***不能为空***
	Name string `json:"name,omitempty"`

	// Description 项目描述, 尽量限制在100字符
	Description string `json:"description,omitempty"`

	// BusinessID 项目绑定的蓝鲸CMDB中业务ID信息
	BusinessID string `json:"businessID,omitempty"`

	// Managers 项目管理员，如果存在多个管理请以英文分号分隔，如："huiwen;porter"
	Managers string `json:"managers,omitempty"`

	// Creator 创建人
	Creator string `json:"creator,omitempty"`
}

// GetProjectRequest  body
//
// ProjectID 项目ID
// ***不能为空***
type GetProjectRequest struct {
	// RequestID 请求ID
	RequestID string

	// ProjectID 项目ID
	// ***不能为空***
	ProjectID string `json:"projectID"`
}

// ListProjectsRequest  body
type ListProjectsRequest struct {
	// RequestID 请求ID
	RequestID string
}

var (
	// pbMarshaller 创建一个jsonpb.Marshaler
	pbMarshaller = new(jsonpb.Marshaler)
)
