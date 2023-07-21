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

package project

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"

	log "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	config "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/httpclient"
)

var projMgrCli *ProjClient

var initOnce sync.Once

const (
	// cache key 项目信息缓存键前缀
	cacheProjectKeyPrefix = "project_%s"
	// defaultExpiration
	defaultExpiration = time.Hour
)

// genProjInfoCacheKey 获取项目信息缓存键
func genProjInfoCacheKey(projectID string) string {
	return cacheProjectKeyPrefix + "-" + projectID
}

// ProjClient xxx
type ProjClient struct {
	cache *cache.Cache
}

// InitProjClient 初始化项目管理客户端
func InitProjClient() {
	if projMgrCli != nil {
		return
	}
	initOnce.Do(func() {
		var err error
		if projMgrCli, err = NewProjClient(); err != nil {
			projMgrCli = nil
			panic(err)
		}
	})
}

// NewProjClient xxx
func NewProjClient() (*ProjClient, error) {
	cli := ProjClient{cache: cache.New(defaultExpiration, cache.NoExpiration)}
	return &cli, nil
}

// GetProjectByCode get project from project code
func (c *ProjClient) fetchProjInfoWithCache(ctx context.Context, projectIDOrCode string) (*Project, error) {
	cacheKey := genProjInfoCacheKey(projectIDOrCode)
	if info, ok := c.cache.Get(cacheKey); info != nil && ok {
		return info.(*Project), nil
	}
	log.Info("project %s info not in cache, start call project manager", projectIDOrCode)

	projInfo, err := c.fetchProjInfo(ctx, projectIDOrCode)
	if err != nil {
		return nil, err
	}

	if err = c.cache.Add(cacheKey, projInfo, defaultExpiration); err != nil {
		log.Warn("project %s info not in cache, start call project manager", projectIDOrCode)
	}
	return projInfo, nil
}

// GetVariable get project from project code
func GetVariable(ctx context.Context, projectCode, clusterID, namespace string) ([]*VariableValue, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/bcsproject/v1/projects/%s/clusters/%s/namespaces/%s/variables/render",
		config.GlobalOptions.BCSAPIGW.Host, projectCode, clusterID, namespace)

	resp, err := httpclient.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Project-Username", ""). // bcs_project 要求有这个header
		SetAuthToken(config.GlobalOptions.BCSAPIGW.AuthToken).
		Get(url)

	if err != nil {
		return nil, nil
	}

	variableValue := make([]*VariableValue, 0)
	if err := httpclient.UnmarshalBKResult(resp, variableValue); err != nil {
		return nil, err
	}

	return variableValue, nil
}

// fetchProjInfo 获取项目信息
func (c *ProjClient) fetchProjInfo(ctx context.Context, projectIDOrCode string) (*Project, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/bcsproject/v1/projects/%s", config.GlobalOptions.BCSAPIGW.Host, projectIDOrCode)

	resp, err := httpclient.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Project-Username", ""). // bcs_project 要求有这个header
		SetAuthToken(config.GlobalOptions.BCSAPIGW.AuthToken).
		Get(url)

	if err != nil {
		return nil, nil
	}
	project := new(Project)
	if err := httpclient.UnmarshalBKResult(resp, project); err != nil {
		return nil, err
	}
	return project, nil
}
