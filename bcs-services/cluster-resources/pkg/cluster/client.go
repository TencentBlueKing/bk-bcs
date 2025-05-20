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

package cluster

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
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/contextx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/httpclient"
)

var clusterMgrCli *CMClient

var initOnce sync.Once

const (
	// ClusterInfoCacheKeyPrefix 集群信息缓存键前缀
	ClusterInfoCacheKeyPrefix = "cluster_manager_cluster_info"
	// CacheExpireTime 缓存过期时间，单位：min
	CacheExpireTime = 5
	// PurgeExpiredCacheTime 清理过期缓存时间，单位：min
	PurgeExpiredCacheTime = 60
)

// genClusterInfoCacheKey 获取集群信息缓存键
func genClusterInfoCacheKey(clusterID string) string {
	return ClusterInfoCacheKeyPrefix + "-" + clusterID
}

// CMClient ClusterManagerClient
type CMClient struct {
	cache *cache.Cache
}

// InitCMClient 初始化集群管理客户端
func InitCMClient() {
	if clusterMgrCli != nil || runtime.RunMode == runmode.Dev {
		return
	}
	initOnce.Do(func() {
		var err error
		if clusterMgrCli, err = NewCMClient(); err != nil {
			clusterMgrCli = nil
			panic(err)
		}
	})
}

// NewCMClient xxx
func NewCMClient() (*CMClient, error) {
	cli := CMClient{cache: cache.New(time.Minute*CacheExpireTime, time.Minute*PurgeExpiredCacheTime)}

	return &cli, nil
}

// fetchClusterInfoWithCache 获取集群信息（支持缓存）
func (c *CMClient) fetchClusterInfoWithCache(ctx context.Context, clusterID string) (*Cluster, error) {
	cacheKey := genClusterInfoCacheKey(clusterID)
	if info, ok := c.cache.Get(cacheKey); info != nil && ok {
		return info.(*Cluster), nil
	}
	log.Info(ctx, "cluster %s info not in cache, start call cluster manager", clusterID)

	clusterInfo, err := c.fetchClusterInfo(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	if err = c.cache.Add(cacheKey, clusterInfo, cache.DefaultExpiration); err != nil {
		log.Warn(ctx, "set cluster info to cache failed: %v", err)
	}
	return clusterInfo, nil
}

// fetchClusterInfo 获取集群信息
func (c *CMClient) fetchClusterInfo(ctx context.Context, clusterID string) (*Cluster, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cluster/%s", config.G.BCSAPIGW.Host, clusterID)

	resp, err := httpclient.GetClient().R().
		SetContext(ctx).
		SetHeaders(contextx.GetLaneIDByCtx(ctx)).
		SetAuthToken(config.G.BCSAPIGW.AuthToken).
		Get(url)

	if err != nil {
		return nil, err
	}

	var cluster *Cluster
	if err := httpclient.UnmarshalBKResult(resp, &cluster); err != nil {
		return nil, err
	}

	// 设置集群类型
	if cluster.IsShared {
		cluster.Type = ClusterTypeShared
	} else {
		cluster.Type = ClusterTypeSingle
	}

	return cluster, nil
}
