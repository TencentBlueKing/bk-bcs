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

package cluster

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	"github.com/fatih/structs"
	"github.com/patrickmn/go-cache"
	"go-micro.dev/v4/registry"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runmode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
	bcsapicm "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/bcsapi/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/discovery"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
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

// 获取集群信息缓存键
func genClusterInfoCacheKey(clusterID string) string {
	return ClusterInfoCacheKeyPrefix + "-" + clusterID
}

// CMClient ClusterManagerClient
type CMClient struct {
	ServiceName  string
	EtcdRtr      registry.Registry
	CliTLSConfig *tls.Config
	discovery    *discovery.ServiceDiscovery
	cache        *cache.Cache
	ctx          context.Context
	cancel       context.CancelFunc
}

// InitCMClient 初始化集群管理客户端
func InitCMClient(microRtr registry.Registry, cliTLSConf *tls.Config) {
	if clusterMgrCli != nil || runtime.RunMode == runmode.Dev {
		return
	}
	initOnce.Do(func() {
		var err error
		if clusterMgrCli, err = newCMClient(microRtr, cliTLSConf); err != nil {
			clusterMgrCli = nil
			panic(err)
		}
	})
}

func newCMClient(microRtr registry.Registry, cliTLSConf *tls.Config) (*CMClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cli := CMClient{
		ServiceName:  conf.ClusterMgrServiceName,
		EtcdRtr:      microRtr,
		CliTLSConfig: cliTLSConf,
		discovery:    discovery.NewServiceDiscovery(conf.ClusterMgrServiceName, microRtr),
		cache:        cache.New(time.Minute*CacheExpireTime, time.Minute*PurgeExpiredCacheTime),
		ctx:          ctx,
		cancel:       cancel,
	}
	if err := cli.discovery.Start(); err != nil {
		return nil, err
	}
	return &cli, nil
}

// 获取集群信息
func (c *CMClient) fetchClusterInfo(ctx context.Context, clusterID string) (map[string]interface{}, error) {
	cacheKey := genClusterInfoCacheKey(clusterID)
	info, ok := c.cache.Get(cacheKey)
	if info != nil && ok {
		return info.(map[string]interface{}), nil
	}

	log.Info(ctx, "cluster %s info not in cache, start call cluster manager", clusterID)

	cli, err := c.genAvailableCli(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := cli.GetCluster(ctx, &bcsapicm.GetClusterReq{ClusterID: clusterID})
	if err != nil || !resp.Result {
		return nil, errorx.New(errcode.ComponentErr, "获取集群 %s 信息失败", clusterID)
	}
	log.Info(ctx, "get cluster %s info: %v", clusterID, structs.Map(resp.Data))

	clusterInfo := map[string]interface{}{
		"id":     resp.Data.ClusterID,
		"name":   resp.Data.ClusterName,
		"type":   ClusterTypeSingle,
		"projID": resp.Data.ProjectID,
	}
	if resp.Data.IsShared {
		clusterInfo["type"] = ClusterTypeShared
	}

	err = c.cache.Add(cacheKey, clusterInfo, cache.DefaultExpiration)
	if err != nil {
		log.Warn(ctx, "set cluster info to cache failed: %v", err)
	}
	return clusterInfo, nil
}

// 获取可用的 ClusterManager 服务
func (c *CMClient) genAvailableCli(ctx context.Context) (bcsapicm.ClusterManagerClient, error) {
	node, err := c.discovery.GetRandServiceInst(ctx)
	if err != nil {
		return nil, err
	}
	log.Info(ctx, "get remote cluster manager instance [%s] from etcd registry successfully", node.Address)

	cli := NewClusterManager(&Config{
		Hosts:     []string{node.Address},
		TLSConfig: c.CliTLSConfig,
	})

	if cli == nil {
		return nil, errorx.New(errcode.ComponentErr, "no available cluster manager client")
	}
	return cli, nil
}
