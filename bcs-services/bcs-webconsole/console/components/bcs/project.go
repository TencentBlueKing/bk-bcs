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
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/storage"
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

// localProjectCache :
var localProjectCache = storage.NewSlotCache[*Project]()

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

// GetProject 通过 project_id/code 获取项目信息
func GetProject(ctx context.Context, bcsConf *config.BCSConf, projectIDOrCode string) (*Project, error) {
	cacheKey := fmt.Sprintf("bcs.GetProject:%s", projectIDOrCode)
	if cacheResult, ok := localProjectCache.Slot.Get(cacheKey); ok {
		return cacheResult, nil
	}

	url := fmt.Sprintf("%s/bcsapi/v4/bcsproject/v1/projects/%s", bcsConf.InnerHost, projectIDOrCode)
	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetAuthToken(bcsConf.Token).
		Get(url)

	if err != nil {
		return nil, err
	}

	project := new(Project)
	if err := components.UnmarshalBKResult(resp, project); err != nil {
		return nil, err
	}

	localProjectCache.Slot.Set(cacheKey, project, localProjectCache.DefaultExpiration)

	return project, nil
}
