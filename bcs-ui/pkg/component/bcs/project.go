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

package bcs

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/component/bcs/project"
)

// Project 项目信息
type Project struct {
	Name        string `json:"name"`
	ProjectID   string `json:"projectID"`
	ProjectCode string `json:"projectCode"`
	BusinessID  string `json:"businessID"`
	Creator     string `json:"creator"`
	Kind        string `json:"kind"`
}

// GetProjectResponse 项目信息响应
type GetProjectResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    *Project `json:"data"`
}

// GetProject 通过 project_id/code 获取项目信息
func GetProject(ctx context.Context, projectIDOrCode string) (*Project, error) {
	p, err := project.GetProjectByCode(ctx, projectIDOrCode)
	if err != nil {
		return nil, err
	}
	return &Project{
		Name:        p.Name,
		ProjectID:   p.ProjectID,
		ProjectCode: p.ProjectCode,
		BusinessID:  p.BusinessID,
		Creator:     p.Creator,
		Kind:        p.Kind,
	}, nil
}
