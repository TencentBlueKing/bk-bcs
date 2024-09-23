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

package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cache"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cache/redis"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

const (
	// ResCacheTTL 资源信息默认过期时间 14 天
	ResCacheTTL = 14 * 24 * 60 * 60

	// ResCacheKeyPrefix 集群资源信息 Redis 缓存键前缀
	ResCacheKeyPrefix = "osrcp"

	// CacheLockTTL 自动重置 ServerGroup 缓存时间间隔 10 分钟
	CacheLockTTL = 10 * 60
)

// RedisCacheClient 基于 Redis 缓存的，单个集群资源信息 Client
type RedisCacheClient struct {
	ctx      context.Context
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

// GroupKindVersionResource api-resouces model
type GroupKindVersionResource []map[string]interface{}

// RESTClient xxx
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
	// 在我们的使用场景中，若某个 Group（如 v1beta1.metrics.k8s.io）异常，
	// 不应当影响在其他 Group 中寻找 Preferred 的资源，因此这里只记录错误日志并忽略
	ret, err := discovery.ServerPreferredResources(d)
	if err != nil {
		log.Warn(d.ctx, "fetch some group's version resources in cluster %s failed: %v", d.clusterID, err)
	}
	return ret, nil
}

// ServerPreferredNamespacedResources 获取集群命名空间维度资源 preferred 版本
// NOTE 由于该方法暂未在项目中被使用，因此不做忽略异常 Group 处理，若启用可参考 ServerPreferredResources 进行改造
func (d *RedisCacheClient) ServerPreferredNamespacedResources() ([]*metav1.APIResourceList, error) {
	return discovery.ServerPreferredNamespacedResources(d)
}

// ServerVersion 获取集群 Server 版本（git version）
func (d *RedisCacheClient) ServerVersion() (*version.Info, error) {
	return d.delegate.ServerVersion()
}

// OpenAPISchema 获取集群支持的 Swagger API Schema
func (d *RedisCacheClient) OpenAPISchema() (*openapi_v2.Document, error) {
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

// ClearCache 清理缓存内容 慎用！
func (d *RedisCacheClient) ClearCache() error {
	log.Warn(d.ctx, "invalidate cluster %s discovery cache", d.clusterID)

	if err := d.checkCacheLock(); err != nil {
		return err
	}

	var ret []byte
	allGroupCacheKey := genCacheKey(d.clusterID, "")
	if err := d.rdsCache.Get(d.ctx, allGroupCacheKey, &ret); err != nil {
		// 如果集群没有对应缓存，是取不到数据的，也不需要清理，因此忽略异常
		log.Warn(d.ctx, "failed to get all group cache: %v", err)
		return nil
	}
	var allGroup map[string]interface{}
	if err := json.Unmarshal(ret, &allGroup); err != nil {
		return err
	}
	// 逐个遍历 AllGroup 中缓存的 GroupVersion 并删除
	for _, group := range mapx.GetList(allGroup, "groups") {
		for _, ver := range mapx.GetList(group.(map[string]interface{}), "versions") {
			groupVersion := mapx.GetStr(ver.(map[string]interface{}), "groupVersion")
			cacheKey := genCacheKey(d.clusterID, groupVersion)
			if err := d.rdsCache.Delete(d.ctx, cacheKey); err != nil {
				log.Warn(d.ctx, "delete cache key %s failed: %v, continue", cacheKey, err)
			}
		}
	}
	// 最后再删除 AllGroup 的缓存
	if err := d.rdsCache.Delete(d.ctx, allGroupCacheKey); err != nil {
		return err
	}

	return d.setCacheLock()
}

// ServerGroups 获取集群中的 Group，包含 versions, preferred 信息（支持 redis 缓存）
func (d *RedisCacheClient) ServerGroups() (*metav1.APIGroupList, error) {
	if cachedBytes, err := d.readCache(""); err == nil {
		cachedGroups := &metav1.APIGroupList{}
		if err = runtime.DecodeInto(scheme.Codecs.UniversalDecoder(), cachedBytes, cachedGroups); err == nil {
			return cachedGroups, nil
		}
	}

	liveGroups, err := d.delegate.ServerGroups()
	if err != nil {
		log.Warn(d.ctx, "cluster: %s, skip caching discovery info due to %v", d.clusterID, err)
		return liveGroups, err
	}
	if liveGroups == nil || len(liveGroups.Groups) == 0 {
		log.Warn(d.ctx, "cluster: %s, skip caching discovery info, no groups found", d.clusterID)
		return liveGroups, err
	}
	if err = d.writeCache("", liveGroups); err != nil {
		// Redis 缓存写失败应该有通知机制
		log.Warn(d.ctx, "cluster: %s, failed to write cache due to %v", d.clusterID, err)
	}
	return liveGroups, nil
}

// ServerResourcesForGroupVersion 获取指定 Group 与 Version 拥有的资源（支持 redis 缓存）
func (d *RedisCacheClient) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	if cachedBytes, err := d.readCache(groupVersion); err == nil {
		cachedResources := &metav1.APIResourceList{}
		if err = runtime.DecodeInto(scheme.Codecs.UniversalDecoder(), cachedBytes, cachedResources); err == nil {
			return cachedResources, nil
		}
	}

	liveResources, err := d.delegate.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		log.Warn(d.ctx, "cluster: %s, skip caching %s discovery info due to %v", d.clusterID, groupVersion, err)
		return liveResources, err
	}
	if liveResources == nil || len(liveResources.APIResources) == 0 {
		log.Warn(d.ctx, "cluster: %s, skip caching %s discovery info, no res found", d.clusterID, groupVersion)
		return liveResources, err
	}
	if err = d.writeCache(groupVersion, liveResources); err != nil {
		// Redis 缓存写失败应该有通知机制
		log.Warn(d.ctx, "cluster: %s, failed to write cache due to %v", d.clusterID, err)
	}
	return liveResources, nil
}

// getResWithGroupVersion 根据指定的 Group, Version 获取对应资源信息
func (d *RedisCacheClient) getResWithGroupVersion(kind, groupVersion string) (schema.GroupVersionResource, error) {
	all, err := d.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	return filterResByKind(kind, d.clusterID, groupVersion, []*metav1.APIResourceList{all})
}

// getPreferredResource 获取指定资源当前集群 Preferred 版本
func (d *RedisCacheClient) getPreferredResource(kind string) (schema.GroupVersionResource, error) {
	all, err := d.ServerPreferredResources()
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	// 逐个检查出第一个同名资源，作为 Preferred 结果返回
	return filterResByKind(kind, d.clusterID, "", all)
}

// getPreferredApiResources 获取指定api资源当前集群 Preferred 版本列表
func (d *RedisCacheClient) getPreferredApiResources(kind, crdName string) (map[string]GroupKindVersionResource, error) {
	all, err := d.ServerPreferredResources()
	if err != nil {
		return nil, err
	}

	// 逐个检查出第一个同名资源，作为 Preferred 结果返回
	return filterApiResByKind(kind, crdName, all), nil
}

// readCache 读缓存逻辑
func (d *RedisCacheClient) readCache(groupVersion string) ([]byte, error) {
	if !d.Fresh() {
		return nil, errorx.New(errcode.General, "cache invalidated")
	}

	key := genCacheKey(d.clusterID, groupVersion)
	if !d.rdsCache.Exists(d.ctx, key) {
		return nil, errorx.New(errcode.General, "key %s cache not exists", key.Key())
	}

	var ret []byte
	err := d.rdsCache.Get(d.ctx, key, &ret)
	return ret, err
}

// writeCache 写缓存逻辑
func (d *RedisCacheClient) writeCache(groupVersion string, obj runtime.Object) error {
	key := genCacheKey(d.clusterID, groupVersion)

	bytes, err := runtime.Encode(scheme.Codecs.LegacyCodec(), obj)
	if err != nil {
		return err
	}

	err = d.rdsCache.Set(d.ctx, key, bytes, 0)
	if err != nil {
		return err
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.cacheValid = true
	return nil
}

// checkCacheLock 检查缓存锁
func (d *RedisCacheClient) checkCacheLock() error {
	lockCacheKey := genLockKey(d.clusterID)
	if d.rdsCache.Exists(d.ctx, lockCacheKey) {
		log.Warn(d.ctx, "the interval is too short for reset cluster %s cache, please try again later", d.clusterID)
		return errorx.New(errcode.General, i18n.GetMsg(d.ctx, "清理集群资源缓存时间间隔过短，请稍后再试"))
	}
	return nil
}

// setCacheLock 设置缓存锁
func (d *RedisCacheClient) setCacheLock() error {
	lockCacheKey := genLockKey(d.clusterID)
	return d.rdsCache.Set(d.ctx, lockCacheKey, "locked", CacheLockTTL*time.Second)
}

// GetServerVersion 获取集群版本信息
func GetServerVersion(ctx context.Context, clusterID string) (*version.Info, error) {
	cli, err := NewRedisCacheClient4Conf(ctx, NewClusterConf(clusterID))
	if err != nil {
		return nil, err
	}
	return cli.ServerVersion()
}

// GetGroupVersionResource 根据配置，名称等信息，获取指定资源对应的 GroupVersionResource
// 若指定 GroupVersion，则在对应的 Group 中寻找资源信息，否则获取 preferred version
// 包含刷新缓存逻辑，若首次从缓存中找不到对应资源，会刷新缓存再次查询，若还是找不到，则返回错误
func GetGroupVersionResource(
	ctx context.Context, conf *ClusterConf, kind, groupVersion string,
) (schema.GroupVersionResource, error) {
	cli, err := NewRedisCacheClient4Conf(ctx, conf)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	// 按指定 groupVersion 查询（含刷新缓存重试）
	if len(groupVersion) != 0 {
		var res schema.GroupVersionResource
		res, err = cli.getResWithGroupVersion(kind, groupVersion)
		if err != nil {
			cli.Invalidate()
			return cli.getResWithGroupVersion(kind, groupVersion)
		}
		return res, nil
	}
	// 查询 preferred version（含刷新缓存重试）
	var res schema.GroupVersionResource
	res, err = cli.getPreferredResource(kind)
	if err != nil {
		cli.Invalidate()
		return cli.getPreferredResource(kind)
	}
	return res, nil
}

// GetApiResources 根据配置，名称等信息，获取指定资源对应的 GetApiResources
// 直接获取 preferred api version
// 包含刷新缓存逻辑，若首次从缓存中找不到对应资源，会刷新缓存再次查询，若还是找不到，则返回错误
func GetApiResources(
	ctx context.Context, conf *ClusterConf, kind, crdName string) (map[string]GroupKindVersionResource, error) {
	cli, err := NewRedisCacheClient4Conf(ctx, conf)
	if err != nil {
		return nil, err
	}
	// 查询 preferred version（含刷新缓存重试）
	var res map[string]GroupKindVersionResource
	res, err = cli.getPreferredApiResources(kind, crdName)
	if err != nil || len(res) == 0 {
		_ = cli.ClearCache()
		return cli.getPreferredApiResources(kind, crdName)
	}
	return res, nil
}

// GetResPreferredVersion 获取某类资源在集群中的 Preferred 版本
func GetResPreferredVersion(ctx context.Context, clusterID, kind string) (string, error) {
	resInfo, err := GetGroupVersionResource(ctx, NewClusterConf(clusterID), kind, "")
	if err != nil {
		return "", errorx.New(errcode.General, i18n.GetMsg(ctx, "获取资源 APIVersion 信息失败：%v"), err)
	}
	if resInfo.Group != "" {
		return resInfo.Group + "/" + resInfo.Version, nil
	}
	return resInfo.Version, nil
}

// NewRedisCacheClient4Conf 根据 Conf 创建 RedisCacheClient
func NewRedisCacheClient4Conf(ctx context.Context, conf *ClusterConf) (*RedisCacheClient, error) {
	delegate, err := discovery.NewDiscoveryClientForConfig(conf.Rest)
	if err != nil {
		return nil, err
	}
	rdsCache := redis.NewCache(ResCacheKeyPrefix, ResCacheTTL*time.Second)
	return newRedisCacheClient(ctx, delegate, conf.ClusterID, rdsCache), nil
}

// filterApiResByKind 获取对应的api资源信息
func filterApiResByKind(kind, crdName string, allRes []*metav1.APIResourceList) map[string]GroupKindVersionResource {
	resources := make(map[string]GroupKindVersionResource, 0)
	for _, apiResList := range allRes {
		for _, res := range apiResList.APIResources {
			// 可能存在如 v1 这种，只有 version，group 为空的情况
			group, ver := "", apiResList.GroupVersion
			if strings.Contains(apiResList.GroupVersion, "/") {
				group, ver = stringx.Partition(apiResList.GroupVersion, "/")
			}
			if (kind != "" && res.Kind == kind) || (crdName != "" && res.Name == crdName) {
				resources[apiResList.GroupVersion] = append(resources[apiResList.GroupVersion],
					map[string]interface{}{
						"group":      group,
						"kind":       res.Kind,
						"version":    ver,
						"resource":   res.Name,
						"namespaced": res.Namespaced,
					})
				return resources
			}
		}
	}

	return resources
}

// filterResByKind 根据 kind 过滤出对应的资源信息
func filterResByKind(
	kind, clusterID, groupVersion string, allRes []*metav1.APIResourceList,
) (schema.GroupVersionResource, error) {
	for _, apiResList := range allRes {
		for _, res := range apiResList.APIResources {
			if res.Kind == kind {
				// 可能存在如 v1 这种只有 version，group 为空的情况
				group, ver := "", apiResList.GroupVersion
				if strings.Contains(apiResList.GroupVersion, "/") {
					group, ver = stringx.Partition(apiResList.GroupVersion, "/")
				}
				return schema.GroupVersionResource{Group: group, Version: ver, Resource: res.Name}, nil
			}
		}
	}
	errMsg := fmt.Sprintf("kind %s not found in cluster %s", kind, clusterID)
	if groupVersion != "" {
		errMsg += ", groupVersion: " + groupVersion
	}
	return schema.GroupVersionResource{}, errorx.New(errcode.General, errMsg)
}

func genCacheKey(clusterID, groupVersion string) cache.StringKey {
	// 不指定 groupVersion 说明是整个集群的 group 资源
	if groupVersion == "" {
		return cache.NewStringKey(fmt.Sprintf("%s:all:servergroups", clusterID))
	}
	// 否则则为指定 group version 拥有的资源
	return cache.NewStringKey(fmt.Sprintf("%s:%s:serverresources", clusterID, groupVersion))
}

// 生成缓存重置锁 Redis 键
func genLockKey(clusterID string) cache.StringKey {
	return cache.NewStringKey(fmt.Sprintf("%s:cache-lock", clusterID))
}

func newRedisCacheClient(
	ctx context.Context,
	delegate discovery.DiscoveryInterface,
	clusterID string,
	rdsCache *redis.Cache,
) *RedisCacheClient {
	return &RedisCacheClient{
		ctx:        ctx,
		delegate:   delegate,
		clusterID:  clusterID,
		rdsCache:   rdsCache,
		cacheValid: true,
	}
}
