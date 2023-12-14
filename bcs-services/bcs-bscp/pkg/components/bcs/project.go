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

// Package bcs provides bcs api client.
package bcs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/components"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/config"
)

// Project 项目信息
type Project struct {
	Name          string `json:"name"`
	ProjectId     string `json:"projectID"`
	Code          string `json:"projectCode"`
	CcBizID       string `json:"businessID"`
	Creator       string `json:"creator"`
	Kind          string `json:"kind"`
	RawCreateTime string `json:"createTime"`
}

// String :
func (p *Project) String() string {
	var displayCode string
	if p.Code == "" {
		displayCode = "-"
	} else {
		displayCode = p.Code
	}
	return fmt.Sprintf("project<%s, %s|%s|%s>", p.Name, displayCode, p.ProjectId, p.CcBizID)
}

// CreateTime xxx
func (p *Project) CreateTime() (time.Time, error) {
	return time.ParseInLocation("2006-01-02T15:04:05Z", p.RawCreateTime, config.G.Base.Location)
}

// ListAuthorizedProjects 通过 用户 获取项目信息
func ListAuthorizedProjects(ctx context.Context, username string) ([]*Project, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/bcsproject/v1/authorized_projects", "")
	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bcs-Username", username).
		SetAuthToken("").
		Get(url)

	if err != nil {
		return nil, err
	}

	projects := []*Project{}
	if err := components.UnmarshalBKResult(resp, projects); err != nil {
		return nil, err
	}

	return projects, nil
}

// ListProjects 按项目 Code 查询
func ListProjects(ctx context.Context, projectCodeList []string) ([]*Project, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/bcsproject/v1/projects", "")
	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bcs-Username", "").
		SetQueryParam("projectCode", strings.Join(projectCodeList, ",")).
		SetAuthToken("").
		Get(url)

	if err != nil {
		return nil, err
	}

	projects := []*Project{}
	if err := components.UnmarshalBKResult(resp, projects); err != nil {
		return nil, err
	}

	return projects, nil
}
