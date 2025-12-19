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
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/discovery"
	"github.com/patrickmn/go-cache"
	"go-micro.dev/v4/registry"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runmode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
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

	// ClusterManagerServiceName cluster manager service name
	ClusterManagerServiceName = "clustermanager.bkbcs.tencent.com"
)

// NewClient create cluster manager service client
func NewClient(tlsConfig *tls.Config, microRgt registry.Registry) error {
	if !discovery.UseServiceDiscovery() {
		dis := discovery.NewModuleDiscovery(ClusterManagerServiceName, microRgt)
		err := dis.Start()
		if err != nil {
			return err
		}
		clustermanager.SetClientConfig(tlsConfig, dis)
	} else {
		clustermanager.SetClientConfig(tlsConfig, nil)
	}
	return nil
}

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

	clusterInfo, err := fetchClusterInfo(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	c.cache.Set(cacheKey, clusterInfo, cache.DefaultExpiration)
	return clusterInfo, nil
}

// fetchClusterInfo get cluster from cluster manager
func fetchClusterInfo(ctx context.Context, clusterID string) (*Cluster, error) {
	cli, close, err := clustermanager.GetClient(conf.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	p, err := cli.GetCluster(ctx, &clustermanager.GetClusterReq{ClusterID: clusterID})
	if err != nil {
		return nil, fmt.Errorf("GetCluster error: %s", err)
	}
	if p.Code != 0 || p.Data == nil {
		return nil, fmt.Errorf("GetCluster error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	cluster := &Cluster{
		ID:       p.Data.ClusterID,
		Name:     p.Data.ClusterName,
		ProjID:   p.Data.ProjectID,
		Status:   p.Data.Status,
		IsShared: p.Data.IsShared,
		Type:     "",
	}
	// 设置集群类型
	if cluster.IsShared {
		cluster.Type = ClusterTypeShared
	} else {
		cluster.Type = ClusterTypeSingle
	}

	return cluster, nil
}
