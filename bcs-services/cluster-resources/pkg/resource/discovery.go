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

package resource

import (
	"fmt"
	"strings"
	"sync"
	"time"

	openapiv2 "github.com/googleapis/gnostic/openapiv2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cache"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cache/redis"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
)

// RedisCacheClient 基于 Redis 缓存的，单个集群资源信息 Client
type RedisCacheClient struct {
	delegate discovery.DiscoveryInterface

	// redis 缓存
	rdsCache *redis.Cache

	// 集群 ID
	clusterID string

	// mutex 锁保护 cacheValid 字段
	mutex sync.Mutex

	// cacheValid 为 false 则缓存无效
	cacheValid bool
}

// RESTClient ...
func (d *RedisCacheClient) RESTClient() rest.Interface {
	return d.delegate.RESTClient()
}

// ServerResources 获取集群中所有资源 Groups 与 Versions
// Deprecated: use ServerGroupsAndResources instead
func (d *RedisCacheClient) ServerResources() ([]*metav1.APIResourceList, error) {
	_, rs, err := d.ServerGroupsAndResources()
	return rs, err
}

// ServerGroupsAndResources 获取集群中所有资源 Groups 与 Versions
func (d *RedisCacheClient) ServerGroupsAndResources() ([]*metav1.APIGroup, []*metav1.APIResourceList, error) {
	return discovery.ServerGroupsAndResources(d)
}

// ServerPreferredResources 获取集群资源 preferred 版本
func (d *RedisCacheClient) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	return discovery.ServerPreferredResources(d)
}

// ServerPreferredNamespacedResources 获取集群命名空间维度资源 preferred 版本
func (d *RedisCacheClient) ServerPreferredNamespacedResources() ([]*metav1.APIResourceList, error) {
	return discovery.ServerPreferredNamespacedResources(d)
}

// ServerVersion 获取集群 Server 版本（git version）
func (d *RedisCacheClient) ServerVersion() (*version.Info, error) {
	return d.delegate.ServerVersion()
}

// OpenAPISchema 获取集群支持的 Swagger API Schema
func (d *RedisCacheClient) OpenAPISchema() (*openapiv2.Document, error) {
	return d.delegate.OpenAPISchema()
}

// Invalidate 使缓存失效
func (d *RedisCacheClient) Invalidate() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.cacheValid = false
}

// Fresh 检查缓存状态
func (d *RedisCacheClient) Fresh() bool {
	return d.cacheValid
}

// ServerGroups 获取集群中的 Group，包含 versions, preferred 信息（支持 redis 缓存）
func (d *RedisCacheClient) ServerGroups() (*metav1.APIGroupList, error) {
	if cachedBytes, err := d.readCache(""); err == nil {
		cachedGroups := &metav1.APIGroupList{}
		if err = runtime.DecodeInto(scheme.Codecs.UniversalDecoder(), cachedBytes, cachedGroups); err == nil {
			log.Info("cluster: %s, get cache (ServerGroups) from redis", d.clusterID)
			return cachedGroups, nil
		}
	}

	liveGroups, err := d.delegate.ServerGroups()
	if err != nil {
		log.Warn("cluster: %s, skipped caching discovery info due to %v", d.clusterID, err)
		return liveGroups, err
	}
	if liveGroups == nil || len(liveGroups.Groups) == 0 {
		log.Warn("cluster: %s, skipped caching discovery info, no groups found", d.clusterID)
		return liveGroups, err
	}
	if err = d.writeCache("", liveGroups); err != nil {
		// TODO Redis 缓存写失败应该有通知机制
		log.Warn("cluster: %s, failed to write cache due to %v", d.clusterID, err)
	}
	return liveGroups, nil
}

// ServerResourcesForGroupVersion 获取指定 Group 与 Version 拥有的资源（支持 redis 缓存）
func (d *RedisCacheClient) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	if cachedBytes, err := d.readCache(groupVersion); err == nil {
		cachedResources := &metav1.APIResourceList{}
		if err = runtime.DecodeInto(scheme.Codecs.UniversalDecoder(), cachedBytes, cachedResources); err == nil {
			log.Info("cluster: %s, get cache (ServerResources, groupVersion: %s) from redis", d.clusterID, groupVersion)
			return cachedResources, nil
		}
	}

	liveResources, err := d.delegate.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		log.Warn("cluster: %s, skipped caching discovery info due to %v", d.clusterID, err)
		return liveResources, err
	}
	if liveResources == nil || len(liveResources.APIResources) == 0 {
		log.Warn("cluster: %s, skipped caching discovery info, no resources found", d.clusterID)
		return liveResources, err
	}
	if err = d.writeCache(groupVersion, liveResources); err != nil {
		// TODO Redis 缓存写失败应该有通知机制
		log.Warn("cluster: %s, failed to write cache due to %v", d.clusterID, err)
	}
	return liveResources, nil
}

// GetGroupVersionResource 根据配置，名称等信息，获取指定资源对应的 GroupVersionResource
// 若指定 GroupVersion，则在对应的 Group 中寻找资源信息，否则获取 preferred version
// 包含刷新缓存逻辑，若首次从缓存中找不到对应资源，会刷新缓存再次查询，若还是找不到，则返回错误
func GetGroupVersionResource(
	conf *rest.Config, clusterID, kind, groupVersion string,
) (schema.GroupVersionResource, error) {
	cli, err := newRedisCacheClient4Conf(conf, clusterID)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	// 按指定 groupVersion 查询（含刷新缓存重试）
	if len(groupVersion) != 0 {
		res, err := cli.getResWithGroupVersion(kind, groupVersion)
		if err != nil {
			cli.Invalidate()
			return cli.getResWithGroupVersion(kind, groupVersion)
		}
		return res, nil
	}
	// 查询 preferred version（含刷新缓存重试）
	res, err := cli.getPreferredResource(kind)
	if err != nil {
		cli.Invalidate()
		return cli.getPreferredResource(kind)
	}
	return res, nil
}

// 根据 kind 过滤出对应的资源信息
func filterResByKind(kind string, allRes []*metav1.APIResourceList) (schema.GroupVersionResource, error) {
	for _, apiResList := range allRes {
		for _, res := range apiResList.APIResources {
			if res.Kind == kind {
				// 可能存在如 v1 这种只有 version，group 为空的情况
				group, version := "", apiResList.GroupVersion
				if strings.Contains(apiResList.GroupVersion, "/") {
					splitRet := strings.Split(apiResList.GroupVersion, "/")
					group, version = splitRet[0], splitRet[1]
				}
				return schema.GroupVersionResource{Group: group, Version: version, Resource: res.Name}, nil
			}
		}
	}
	return schema.GroupVersionResource{}, fmt.Errorf("not result for %s", kind)
}

// 根据指定的 Group, Version 获取对应资源信息
func (d *RedisCacheClient) getResWithGroupVersion(kind, groupVersion string) (schema.GroupVersionResource, error) {
	all, err := d.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	return filterResByKind(kind, []*metav1.APIResourceList{all})
}

// 获取指定资源当前集群 Preferred 版本
func (d *RedisCacheClient) getPreferredResource(kind string) (schema.GroupVersionResource, error) {
	all, err := d.ServerPreferredResources()
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	// 逐个检查出第一个同名资源，作为 Preferred 结果返回
	return filterResByKind(kind, all)
}

func genCacheKey(clusterID, groupVersion string) cache.StringKey {
	// 不指定 groupVersion 说明是整个集群的 group 资源
	if len(groupVersion) == 0 {
		return cache.NewStringKey(fmt.Sprintf("%s:all:servergroups", clusterID))
	}
	// 否则则为指定 group version 拥有的资源
	return cache.NewStringKey(fmt.Sprintf("%s:%s:serverresources", clusterID, groupVersion))
}

// 读缓存逻辑
func (d *RedisCacheClient) readCache(groupVersion string) ([]byte, error) {
	if !d.Fresh() {
		return nil, fmt.Errorf("cache invalidated")
	}

	key := genCacheKey(d.clusterID, groupVersion)
	if !d.rdsCache.Exists(key) {
		return nil, fmt.Errorf("key %s cache not exists", key.Key())
	}

	var ret []byte
	err := d.rdsCache.Get(key, &ret)
	return ret, err
}

// 写缓存逻辑
func (d *RedisCacheClient) writeCache(groupVersion string, obj runtime.Object) error {
	key := genCacheKey(d.clusterID, groupVersion)

	bytes, err := runtime.Encode(scheme.Codecs.LegacyCodec(), obj)
	if err != nil {
		return err
	}

	err = d.rdsCache.Set(key, bytes, 0)
	if err != nil {
		return err
	}
	log.Info("write %s to redis done", key)

	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.cacheValid = true
	return nil
}

func newRedisCacheClient(
	delegate discovery.DiscoveryInterface,
	clusterID string,
	rdsCache *redis.Cache,
) *RedisCacheClient {
	return &RedisCacheClient{
		delegate:   delegate,
		clusterID:  clusterID,
		rdsCache:   rdsCache,
		cacheValid: true,
	}
}

// 根据 Conf 创建 RedisCacheClient
func newRedisCacheClient4Conf(conf *rest.Config, clusterID string) (*RedisCacheClient, error) {
	delegate, err := discovery.NewDiscoveryClientForConfig(conf)
	if err != nil {
		return nil, err
	}
	rdsCache := redis.NewCache(ResCacheKeyPrefix, ResCacheTTL*time.Second)
	return newRedisCacheClient(delegate, clusterID, rdsCache), nil
}
