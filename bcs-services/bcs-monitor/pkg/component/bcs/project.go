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
 *
 */

package bcs

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage"
)

// Project 项目信息
type Project struct {
	Name      string `json:"name"`
	ProjectId string `json:"projectID"`
	Code      string `json:"projectCode"`
	CcBizID   string `json:"businessID"`
	Creator   string `json:"creator"`
	Kind      string `json:"kind"`
}

// String
func (p *Project) String() string {
	var displayCode string
	if p.Code == "" {
		displayCode = "-"
	} else {
		displayCode = p.Code
	}
	return fmt.Sprintf("project<%s, %s|%s|%s>", p.Name, displayCode, p.ProjectId, p.CcBizID)
}

// GetProject 通过 project_id/code 获取项目信息
func GetProject(ctx context.Context, bcsConf *config.BCSConf, projectIDOrCode string) (*Project, error) {
	cacheKey := fmt.Sprintf("bcs.GetProject:%s", projectIDOrCode)
	if cacheResult, ok := storage.LocalCache.Slot.Get(cacheKey); ok {
		return cacheResult.(*Project), nil
	}

	url := fmt.Sprintf("%s/bcsapi/v4/bcsproject/v1/projects/%s", bcsConf.Host, projectIDOrCode)
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Project-Username", "bcs-monitor").
		SetAuthToken(bcsConf.BCSProjectToken).
		Get(url)

	if err != nil {
		return nil, err
	}

	project := new(Project)
	if err := component.UnmarshalBKResult(resp, project); err != nil {
		return nil, err
	}

	storage.LocalCache.Slot.Set(cacheKey, project, storage.LocalCache.DefaultExpiration)

	return project, nil
}
