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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/storage"
)

// Project 项目信息
type Project struct {
	Name      string `json:"project_name"`
	ProjectId string `json:"project_id"`
	Code      string `json:"english_name"`
	CcBizID   uint   `json:"cc_app_id"`
	Creator   string `json:"creator"`
	Kind      uint   `json:"kind"`
}

func (p *Project) String() string {
	var displayCode string
	if p.Code == "" {
		displayCode = "-"
	} else {
		displayCode = p.Code
	}
	return fmt.Sprintf("project<%s, %s|%s|%d>", p.Name, displayCode, p.ProjectId, p.CcBizID)
}

// GetProject 通过project_id获取项目信息
func GetProject(ctx context.Context, projectIDOrCode string) (*Project, error) {
	cacheKey := fmt.Sprintf("bcs.GetProject:%s.%s", config.G.BCSCC.Stage, projectIDOrCode)
	if cacheResult, ok := storage.LocalCache.Slot.Get(cacheKey); ok {
		return cacheResult.(*Project), nil
	}

	url := fmt.Sprintf("%s/%s/projects/%s/", config.G.BCSCC.Host, config.G.BCSCC.Stage, projectIDOrCode)
	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetQueryParam("app_code", config.G.Base.AppCode).
		SetQueryParam("app_secret", config.G.Base.AppSecret).
		Get(url)

	if err != nil {
		return nil, err
	}

	project := new(Project)
	if err := components.UnmarshalBKResult(resp, project); err != nil {
		return nil, err
	}

	storage.LocalCache.Slot.Set(cacheKey, project, storage.LocalCache.DefaultExpiration)

	return project, nil
}
