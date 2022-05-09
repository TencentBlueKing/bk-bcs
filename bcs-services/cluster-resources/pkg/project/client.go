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
	"crypto/tls"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"go-micro.dev/v4/registry"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runmode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
	bcsapiProj "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/bcsapi/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/discovery"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	grpcUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/grpc"
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

// 获取项目信息缓存键
func genProjInfoCacheKey(projectID string) string {
	return ProjectInfoCacheKeyPrefix + "-" + projectID
}

// ProjClient ...
type ProjClient struct {
	ServiceName  string
	EtcdRtr      registry.Registry
	CliTLSConfig *tls.Config
	discovery    *discovery.ServiceDiscovery
	cache        *cache.Cache
	ctx          context.Context
	cancel       context.CancelFunc
}

// InitProjClient 初始化项目管理客户端
func InitProjClient(microRtr registry.Registry, cliTLSConf *tls.Config) {
	if projMgrCli != nil || runtime.RunMode == runmode.Dev {
		return
	}
	initOnce.Do(func() {
		var err error
		if projMgrCli, err = NewProjClient(microRtr, cliTLSConf); err != nil {
			projMgrCli = nil
			panic(err)
		}
	})
}

// NewProjClient ...
func NewProjClient(microRtr registry.Registry, cliTLSConf *tls.Config) (*ProjClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cli := ProjClient{
		ServiceName:  conf.ProjectMgrServiceName,
		EtcdRtr:      microRtr,
		CliTLSConfig: cliTLSConf,
		discovery:    discovery.NewServiceDiscovery(conf.ProjectMgrServiceName, microRtr),
		cache:        cache.New(time.Minute*CacheExpireTime, time.Minute*PurgeExpiredCacheTime),
		ctx:          ctx,
		cancel:       cancel,
	}
	if err := cli.discovery.Start(); err != nil {
		return nil, err
	}
	return &cli, nil
}

// 获取项目信息（支持缓存）
func (c *ProjClient) fetchProjInfoWithCache(ctx context.Context, projectID string) (map[string]interface{}, error) {
	cacheKey := genProjInfoCacheKey(projectID)
	if info, ok := c.cache.Get(cacheKey); info != nil && ok {
		return info.(map[string]interface{}), nil
	}
	log.Info(ctx, "project %s info not in cache, start call project manager", projectID)

	projInfo, err := c.fetchProjInfo(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if err = c.cache.Add(cacheKey, projInfo, cache.DefaultExpiration); err != nil {
		log.Warn(ctx, "set project info to cache failed: %v", err)
	}
	return projInfo, nil
}

// 获取项目信息
func (c *ProjClient) fetchProjInfo(ctx context.Context, projectID string) (map[string]interface{}, error) {
	cli, err := c.genAvailableCli(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := cli.GetProject(grpcUtil.SetMD4CTX(ctx), &bcsapiProj.GetProjectRequest{ProjectIDOrCode: projectID})
	if err != nil {
		return nil, errorx.New(errcode.ComponentErr, "call for project %s info failed: %v", projectID, err)
	}
	if resp.Code != 0 {
		return nil, errorx.New(errcode.ComponentErr, "project: %s, errMsg: %s", projectID, resp.Message)
	}
	projInfo := map[string]interface{}{
		"id":   projectID,
		"name": resp.Data.Name,
		"code": resp.Data.ProjectCode,
	}
	log.Info(ctx, "fetch project info: %v", projInfo)
	return projInfo, nil
}

// 获取可用的 ProjManager 服务
func (c *ProjClient) genAvailableCli(ctx context.Context) (bcsapiProj.BCSProjectClient, error) {
	node, err := c.discovery.GetRandServiceInst(ctx)
	if err != nil {
		return nil, err
	}
	log.Info(ctx, "get remote project manager instance [%s] from etcd registry successfully", node.Address)

	conn, err := grpcUtil.NewGrpcConn(node.Address, c.CliTLSConfig)
	if conn == nil || err != nil {
		log.Error(ctx, "create project manager grpc client with %s failed: %v", node.Address, err)
	}

	cli := bcsapiProj.NewBCSProjectClient(conn)
	if cli == nil {
		return nil, errorx.New(errcode.ComponentErr, "no available project manager client")
	}
	return cli, nil
}
