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
	grpcUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/grpc"
)

var clusterMgrCli *CMClient

var initOnce sync.Once

const (
	// ClusterInfoCacheKeyPrefix 集群信息缓存键前缀
	ClusterInfoCacheKeyPrefix = "cluster_manager_cluster_info"
	// CacheExpireTime 缓存过期时间，单位：min
	CacheExpireTime = 20
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
		if clusterMgrCli, err = NewCMClient(microRtr, cliTLSConf); err != nil {
			clusterMgrCli = nil
			panic(err)
		}
	})
}

// NewCMClient ...
func NewCMClient(microRtr registry.Registry, cliTLSConf *tls.Config) (*CMClient, error) {
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

// 获取集群信息（支持缓存）
func (c *CMClient) fetchClusterInfoWithCache(ctx context.Context, clusterID string) (map[string]interface{}, error) {
	cacheKey := genClusterInfoCacheKey(clusterID)
	if info, ok := c.cache.Get(cacheKey); info != nil && ok {
		return info.(map[string]interface{}), nil
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

// 获取集群信息
func (c *CMClient) fetchClusterInfo(ctx context.Context, clusterID string) (map[string]interface{}, error) {
	cli, err := c.genAvailableCli(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := cli.GetCluster(grpcUtil.SetMD4CTX(ctx), &bcsapicm.GetClusterReq{ClusterID: clusterID})
	if err != nil {
		return nil, errorx.New(errcode.ComponentErr, "call for cluster %s info failed: %v", clusterID, err)
	}
	if !resp.Result {
		return nil, errorx.New(errcode.ComponentErr, "cluster: %s, errMsg: %s", clusterID, resp.Message)
	}

	clusterInfo := map[string]interface{}{
		"id":     resp.Data.ClusterID,
		"name":   resp.Data.ClusterName,
		"type":   ClusterTypeSingle,
		"projID": resp.Data.ProjectID,
	}
	if resp.Data.IsShared {
		clusterInfo["type"] = ClusterTypeShared
	}
	log.Info(ctx, "fetch cluster info: %v", clusterInfo)
	return clusterInfo, nil
}

// 获取可用的 ClusterManager 服务
func (c *CMClient) genAvailableCli(ctx context.Context) (bcsapicm.ClusterManagerClient, error) {
	node, err := c.discovery.GetRandServiceInst(ctx)
	if err != nil {
		return nil, err
	}
	log.Info(ctx, "get remote cluster manager instance [%s] from etcd registry successfully", node.Address)

	conn, err := grpcUtil.NewGrpcConn(node.Address, c.CliTLSConfig)
	if conn == nil || err != nil {
		log.Error(ctx, "create cluster manager grpc client with %s failed: %v", node.Address, err)
	}

	cli := bcsapicm.NewClusterManagerClient(conn)
	if cli == nil {
		return nil, errorx.New(errcode.ComponentErr, "no available cluster manager client")
	}
	return cli, nil
}
