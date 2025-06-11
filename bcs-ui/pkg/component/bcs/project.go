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
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/contextx"
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
	bcsConf := config.G.BCS
	url := fmt.Sprintf("%s/bcsapi/v4/bcsproject/v1/projects/%s", bcsConf.Host, projectIDOrCode)
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Project-Username", ""). // bcs_project 要求有这个header
		SetHeaders(contextx.GetLaneIDByCtx(ctx)).
		SetAuthToken(bcsConf.Token).
		Get(url)

	if err != nil {
		return nil, err
	}

	project := new(Project)
	if err := component.UnmarshalBKResult(resp, project); err != nil {
		return nil, err
	}
	return project, nil
}
