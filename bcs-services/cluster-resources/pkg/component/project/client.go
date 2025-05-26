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

package project

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runmode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/httpclient"
)

var projMgrCli *ProjClient

var initOnce sync.Once

const (
	// ProjectInfoCacheKeyPrefix 项目信息缓存键前缀
	ProjectInfoCacheKeyPrefix = "bcs_project_proj_info"
	// CacheExpireTime 缓存过期时间，单位：min
	CacheExpireTime = 20
	// PurgeExpiredCacheTime 清理过期缓存时间，单位：min
	PurgeExpiredCacheTime = 60
)

// genProjInfoCacheKey 获取项目信息缓存键
func genProjInfoCacheKey(projectID string) string {
	return ProjectInfoCacheKeyPrefix + "-" + projectID
}

// genProjNsCacheKey 获取项目命名空间缓存键
func genProjNsCacheKey(projectID, clusterID string) string {
	return "bcs_project_ns" + "-" + projectID + "-" + clusterID
}

// ProjClient xxx
type ProjClient struct {
	cache *cache.Cache
}

// InitProjClient 初始化项目管理客户端
func InitProjClient() {
	if projMgrCli != nil || runtime.RunMode == runmode.Dev {
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
	cli := ProjClient{cache: cache.New(time.Minute*CacheExpireTime, time.Minute*PurgeExpiredCacheTime)}
	return &cli, nil
}

// fetchProjInfoWithCache 获取项目信息（支持缓存）
func (c *ProjClient) fetchProjInfoWithCache(ctx context.Context, projectID string) (*Project, error) {
	cacheKey := genProjInfoCacheKey(projectID)
	if info, ok := c.cache.Get(cacheKey); info != nil && ok {
		return info.(*Project), nil
	}
	log.Info(ctx, "project %s info not in cache, start call project manager", projectID)

	projInfo, err := c.fetchProjInfo(ctx, projectID)
	if err != nil {
		return nil, err
	}

	c.cache.Set(cacheKey, projInfo, cache.DefaultExpiration)
	return projInfo, nil
}

// fetchProjInfo 获取项目信息
func (c *ProjClient) fetchProjInfo(ctx context.Context, projectID string) (*Project, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/bcsproject/v1/projects/%s", config.G.BCSAPIGW.Host, projectID)

	resp, err := httpclient.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Project-Username", ""). // bcs_project 要求有这个header
		SetAuthToken(config.G.BCSAPIGW.AuthToken).
		Get(url)

	if err != nil {
		return nil, err
	}

	project := new(Project)
	if err := httpclient.UnmarshalBKResult(resp, project); err != nil {
		return nil, err
	}

	return project, nil
}

// fetchProjInfoWithCache 获取项目信息（支持缓存）
func (c *ProjClient) fetchSharedClusterProjNsWitchCache(ctx context.Context, projectID, clusterID string) (
	[]Namespace, error) {
	cacheKey := genProjNsCacheKey(projectID, clusterID)
	if info, ok := c.cache.Get(cacheKey); info != nil && ok {
		return info.([]Namespace), nil
	}
	log.Info(ctx, "project %s cluster %s ns not in cache, start call project manager", projectID, clusterID)

	ns, err := c.fetchSharedClusterProjNs(ctx, projectID, clusterID)
	if err != nil {
		return nil, err
	}

	c.cache.Set(cacheKey, ns, time.Minute)
	return ns, nil
}

// fetchShardClusterProjNs 获取共享集群项目下的命名空间
func (c *ProjClient) fetchSharedClusterProjNs(ctx context.Context, projectID, clusterID string) ([]Namespace, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/bcsproject/v1/projects/%s/clusters/%s/namespaces", config.G.BCSAPIGW.Host,
		projectID, clusterID)

	resp, err := httpclient.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Project-Username", ""). // bcs_project 要求有这个header
		SetAuthToken(config.G.BCSAPIGW.AuthToken).
		Get(url)

	if err != nil {
		return nil, err
	}

	ns := make([]Namespace, 0)
	if err := httpclient.UnmarshalBKResult(resp, &ns); err != nil {
		return nil, err
	}

	return ns, nil
}

// getVariable get
func (c *ProjClient) getVariable(ctx context.Context, projectCode, clusterID, namespace string) ([]VariableValue,
	error) {
	url := fmt.Sprintf("%s/bcsapi/v4/bcsproject/v1/projects/%s/clusters/%s/namespaces/%s/variables/render",
		config.G.BCSAPIGW.Host, projectCode, clusterID, namespace)

	resp, err := httpclient.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Project-Username", "").
		SetAuthToken(config.G.BCSAPIGW.AuthToken).
		Get(url)

	if err != nil {
		return nil, err
	}

	data := make([]VariableValue, 0)
	if err := httpclient.UnmarshalBKResult(resp, &data); err != nil {
		return nil, err
	}
	return data, nil
}
