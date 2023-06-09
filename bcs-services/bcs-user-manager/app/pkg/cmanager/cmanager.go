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

// Package cmanager xxx
package cmanager

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/patrickmn/go-cache"
	"go-micro.dev/v4/registry"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
)

const (
	// ModuleClusterManager default discovery clustermanager module
	ModuleClusterManager = "clustermanager.bkbcs.tencent.com"
)

var (
	// errServerNotInit server not inited
	errServerNotInit = errors.New("ClusterManagerClient not inited")
)

const (
	// CacheClusterProjectKeyPrefix key prefix for cluster project id
	CacheClusterProjectKeyPrefix = "cached_cluster_manager"
	// CacheClusterBusinessKeyPrefix key prefix for cluster business id
	CacheClusterBusinessKeyPrefix = "cached_cluster_manager_cluster_business"
	// CacheClusterSharedKeyPrefix key prefix for cluster shared
	CacheClusterSharedKeyPrefix = "cached_cluster_manager_shared"
)

// Options for init clusterManager
type Options struct {
	Module          string
	EtcdRegistry    registry.Registry
	ClientTLSConfig *tls.Config
}

func (o *Options) validate() bool {
	if o == nil {
		return false
	}

	if o.Module == "" {
		o.Module = ModuleClusterManager
	}

	return true
}

// NewClusterManagerClient init cluster manager and start discovery module(clustermanager)
func NewClusterManagerClient(opts *Options) *ClusterManagerClient {
	ok := opts.validate()
	if !ok {
		return nil
	}

	cmClient := &ClusterManagerClient{
		opts: opts,
		// Create a cache with a default expiration time of 5 minutes, and which
		// purges expired items every 1 hour
		cache: cache.New(time.Minute*5, time.Minute*60),
	}

	cmClient.ctx, cmClient.cancel = context.WithCancel(context.Background())
	cmClient.discovery = discovery.NewServiceDiscovery(opts.Module, opts.EtcdRegistry)
	err := cmClient.discovery.Start()
	if err != nil {
		blog.Errorf("start discovery client failed: %v", err)
		return nil
	}

	return cmClient
}

// ClusterManagerClient client for clustermanager
type ClusterManagerClient struct {
	opts      *Options
	discovery discovery.Discovery
	cache     *cache.Cache
	ctx       context.Context
	cancel    context.CancelFunc
}

// GetProjectIDByClusterID get projectID by clusterID
func (cm *ClusterManagerClient) GetProjectIDByClusterID(clusterID string) (string, error) {
	if cm == nil {
		return "", errServerNotInit
	}

	cacheName := func(id string) string {
		return fmt.Sprintf("%s_%v", CacheClusterProjectKeyPrefix, id)
	}
	val, ok := cm.cache.Get(cacheName(clusterID))
	if ok && val != nil {
		if projectID, ok1 := val.(string); ok1 {
			return projectID, nil
		}
	}
	blog.V(3).Infof("GetProjectIDByClusterID miss clusterID cache")

	// calling clustermanager through http
	resp := models.GetClusterResp{}
	bcsApi := config.GetGlobalConfig().BcsAPI
	url := fmt.Sprintf(constant.ClusterUrl, clusterID)
	err := bcsApi.HttpRequest(http.MethodGet, url, nil, &resp)
	if err != nil {
		blog.Errorf("GetProjectIDByClusterID failed: %v", err.Error())
		return "", err
	}

	if !resp.Result {
		blog.Errorf("GetProjectIDByClusterID failed: %v", resp.Message)
		return "", err
	}
	projectID := resp.Data.ProjectID

	err = cm.cache.Add(cacheName(clusterID), projectID, cache.DefaultExpiration)
	if err != nil {
		blog.Errorf("GetProjectIDByClusterID set cache by cacheName[%s] failed: %v", cacheName(clusterID), err)
	}

	return projectID, nil
}

// GetBusinessIDByClusterID get businessID by clusterID
func (cm *ClusterManagerClient) GetBusinessIDByClusterID(clusterID string) (string, error) {
	if cm == nil {
		return "", errServerNotInit
	}

	cacheName := func(id string) string {
		return fmt.Sprintf("%s_%v", CacheClusterBusinessKeyPrefix, id)
	}
	val, ok := cm.cache.Get(cacheName(clusterID))
	if ok && val != nil {
		if businessID, ok1 := val.(string); ok1 {
			return businessID, nil
		}
	}
	blog.V(3).Infof("GetBusinessIDByClusterID miss clusterID cache")

	resp := &models.GetClusterResp{}
	bcsApi := config.GetGlobalConfig().BcsAPI
	url := fmt.Sprintf(constant.ClusterUrl, clusterID)
	err := bcsApi.HttpRequest(http.MethodGet, url, nil, resp)
	if err != nil {
		blog.Errorf("GetProjectIDByClusterID failed: %v", err.Error())
		return "", err
	}

	if !resp.Result {
		blog.Errorf("GetBusinessIDByClusterID failed: %v", resp.Message)
		return "", err
	}
	businessID := resp.Data.BusinessID

	err = cm.cache.Add(cacheName(clusterID), businessID, cache.DefaultExpiration)
	if err != nil {
		blog.Errorf("GetBusinessIDByClusterID set cache by cacheName[%s] failed: %v", cacheName(clusterID), err)
	}

	return businessID, nil
}

// Stop stop clusterManagerClient
func (cm *ClusterManagerClient) Stop() {
	if cm == nil {
		return
	}

	cm.discovery.Stop()
	cm.cancel()
}
