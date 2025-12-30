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
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/discovery"
	"github.com/patrickmn/go-cache"
	"go-micro.dev/v4/registry"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runmode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/project/bcsproject"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
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

	// ProjectManagerServiceName project manager service name
	ProjectManagerServiceName = "project.bkbcs.tencent.com"
)

// NewClient create project service client
func NewClient(tlsConfig *tls.Config, microRgt registry.Registry) error {
	if !discovery.UseServiceDiscovery() {
		dis := discovery.NewModuleDiscovery(ProjectManagerServiceName, microRgt)
		err := dis.Start()
		if err != nil {
			return err
		}
		bcsproject.SetClientConfig(tlsConfig, dis)
	} else {
		bcsproject.SetClientConfig(tlsConfig, nil)
	}
	return nil
}

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

	projInfo, err := fetchProjInfo(ctx, projectID)
	if err != nil {
		return nil, err
	}

	c.cache.Set(cacheKey, projInfo, cache.DefaultExpiration)
	return projInfo, nil
}

// fetchProjInfo 获取项目信息
func fetchProjInfo(ctx context.Context, projectID string) (*Project, error) {
	cli, close, err := bcsproject.GetClient(conf.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	p, err := cli.Project.GetProject(ctx,
		&bcsproject.GetProjectRequest{ProjectIDOrCode: projectID})
	if err != nil {
		return nil, fmt.Errorf("GetProject error: %s", err)
	}
	if p.Code != 0 || p.Data == nil {
		return nil, fmt.Errorf("GetProject error, code: %d, message: %s, requestID: %s",
			p.Code, p.GetMessage(), p.GetRequestID())
	}

	return &Project{
		ID:         p.Data.ProjectID,
		Code:       p.Data.ProjectCode,
		BusinessID: p.Data.BusinessID,
		TenantID:   p.Data.TenantID,
	}, nil
}

// fetchProjInfoWithCache 获取项目信息（支持缓存）
func (c *ProjClient) fetchSharedClusterProjNsWitchCache(ctx context.Context, projectID, clusterID string) (
	[]Namespace, error) {
	cacheKey := genProjNsCacheKey(projectID, clusterID)
	if info, ok := c.cache.Get(cacheKey); info != nil && ok {
		return info.([]Namespace), nil
	}
	log.Info(ctx, "project %s cluster %s ns not in cache, start call project manager", projectID, clusterID)

	ns, err := fetchSharedClusterProjNs(ctx, projectID, clusterID)
	if err != nil {
		return nil, err
	}

	c.cache.Set(cacheKey, ns, time.Minute)
	return ns, nil
}

// fetchShardClusterProjNs 获取共享集群项目下的命名空间
func fetchSharedClusterProjNs(ctx context.Context, projectID, clusterID string) ([]Namespace, error) {
	cli, close, err := bcsproject.GetClient(conf.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	p, err := cli.Namespace.ListNamespaces(ctx, &bcsproject.ListNamespacesRequest{
		ProjectCode: projectID,
		ClusterID:   clusterID,
	})
	if err != nil {
		return nil, fmt.Errorf("ListNamespaces error: %s", err)
	}
	if p.Code != 0 {
		return nil, fmt.Errorf("ListNamespaces error, code: %d, message: %s, requestID: %s",
			p.Code, p.GetMessage(), p.GetRequestID())
	}
	ns := []Namespace{}
	for _, v := range p.Data {
		ns = append(ns, Namespace{
			Name:   v.Name,
			Status: v.Status,
		})
	}
	return ns, nil
}

// getVariable get
func getVariable(ctx context.Context, projectCode, clusterID, namespace string) ([]VariableValue, error) {
	client, close, err := bcsproject.GetClient(conf.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	resp, err := client.Variable.RenderVariables(ctx,
		&bcsproject.RenderVariablesRequest{ProjectCode: projectCode, ClusterID: clusterID, Namespace: namespace})
	if err != nil {
		return nil, fmt.Errorf("ListNamespaceVariables error: %s", err)
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("ListNamespaceVariables error, code: %d, message: %s, requestID: %s",
			resp.Code, resp.GetMessage(), resp.GetRequestID())
	}

	vv := []VariableValue{}
	for _, v := range resp.GetData() {
		vv = append(vv, VariableValue{
			Id:          v.Id,
			Key:         v.Key,
			Name:        v.Name,
			ClusterID:   v.ClusterID,
			ClusterName: v.ClusterName,
			Namespace:   v.Namespace,
			Value:       v.Value,
			Scope:       v.Scope,
		})
	}
	return vv, nil
}
