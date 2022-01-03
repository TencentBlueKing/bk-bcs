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

package resources

import (
	"errors"
	"sync"
	"time"

	openapiv2 "github.com/googleapis/gnostic/openapiv2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

// RedisCacheClient 基于 Redis 缓存的，单个集群资源信息 Client
type RedisCacheClient struct {
	delegate discovery.DiscoveryInterface

	// 集群 ID
	clusterID string

	// redis 服务地址
	redisURL string

	// ttl 缓存过期事件，默认 7 天
	ttl time.Duration

	// mutex 锁保护以下字段信息
	mutex sync.Mutex

	// invalidated 为 true 说明缓存不可用，可通过调用 `Invalidate()` 强制失效
	invalidated bool

	// fresh 为 true 说明缓存可用
	fresh bool
}

var (
	// ErrCacheNotFound
	ErrCacheNotFound = errors.New("cache not found")
)

// RESTClient
func (d *RedisCacheClient) RESTClient() rest.Interface {
	return d.delegate.RESTClient()
}

// ServerGroups returns the supported groups, with information like supported versions and the preferred version.
func (d *RedisCacheClient) ServerGroups() (*metav1.APIGroupList, error) {
	return d.delegate.ServerGroups()
}

// ServerResourcesForGroupVersion returns the supported resources for a group and version.
func (d *RedisCacheClient) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	return d.delegate.ServerResourcesForGroupVersion(groupVersion)
}

// ServerResources returns the supported resources for all groups and versions.
// The returned resource list might be non-nil with partial results even in the case of non-nil error.
// Deprecated: use ServerGroupsAndResources instead.
func (d *RedisCacheClient) ServerResources() ([]*metav1.APIResourceList, error) {
	return d.delegate.ServerResources()
}

// ServerGroupsAndResources returns the supported groups and resources for all groups and versions.
func (d *RedisCacheClient) ServerGroupsAndResources() ([]*metav1.APIGroup, []*metav1.APIResourceList, error) {
	return d.delegate.ServerGroupsAndResources()
}

// ServerPreferredResources returns the supported resources with the version preferred by the server.
func (d *RedisCacheClient) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	return d.delegate.ServerPreferredResources()
}

// ServerPreferredNamespacedResources returns the supported namespaced resources with the version preferred by the server.
func (d *RedisCacheClient) ServerPreferredNamespacedResources() ([]*metav1.APIResourceList, error) {
	return d.delegate.ServerPreferredNamespacedResources()
}

// ServerVersion 获取并返回集群 Server 版本（git version）
func (d *RedisCacheClient) ServerVersion() (*version.Info, error) {
	return d.delegate.ServerVersion()
}

// OpenAPISchema 获取并返回集群支持的 Swagger API Schema
func (d *RedisCacheClient) OpenAPISchema() (*openapiv2.Document, error) {
	return d.delegate.OpenAPISchema()
}

// Invalidate 使缓存失效
func (d *RedisCacheClient) Invalidate() {
}

// Fresh 检查缓存状态
func (d *RedisCacheClient) Fresh() bool {
	return d.fresh
}

// NewRedisCacheClient 创建 Redis Cached Discovery Client
func NewRedisCacheClient(
	delegate discovery.DiscoveryInterface,
	clusterID string,
	redisURL string,
) discovery.CachedDiscoveryInterface {
	return &RedisCacheClient{
		delegate: delegate, clusterID: clusterID, redisURL: redisURL,
	}
}
