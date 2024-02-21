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

package multicluster

import (
	"context"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam"
	clusterAuth "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/cluster"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// Query represents a query for multicluster resources.
type Query interface {
	Fetch(ctx context.Context, groupVersion, kind string) (map[string]interface{}, error)
}

// StorageQuery represents a query for multicluster resources.
type StorageQuery struct {
	ClusterdNamespaces []*clusterRes.ClusterNamespaces
	QueryFilter        QueryFilter
}

// NewStorageQuery creates a new query for multicluster resources.
func NewStorageQuery(ns []*clusterRes.ClusterNamespaces, filter QueryFilter) Query {
	return &StorageQuery{
		ClusterdNamespaces: ns,
		QueryFilter:        filter,
	}
}

// Fetch fetches multicluster resources.
func (q *StorageQuery) Fetch(ctx context.Context, groupVersion, kind string) (map[string]interface{}, error) {
	var err error
	if q.ClusterdNamespaces, err = checkMultiClusterAccess(ctx, kind, q.ClusterdNamespaces); err != nil {
		return nil, err
	}
	log.Info(ctx, "fetch multi cluster resources, kind: %s, clusterdNamespaces: %v", kind, q.ClusterdNamespaces)
	clusteredNamespaces := []storage.ClusteredNamespaces{}
	for _, v := range q.ClusterdNamespaces {
		clusteredNamespaces = append(clusteredNamespaces, storage.ClusteredNamespaces{
			ClusterID:  v.GetClusterID(),
			Namespaces: v.GetNamespaces(),
		})
	}
	resources, err := storage.ListAllMultiClusterResources(ctx, storage.ListMultiClusterResourcesReq{
		Kind:                kind,
		ClusteredNamespaces: clusteredNamespaces,
		Conditions:          q.QueryFilter.ToConditions(),
	})
	if err != nil {
		return nil, err
	}
	resources = ApplyFilter(resources, q.QueryFilter.StatusFilter, q.QueryFilter.IPFilter)
	total := len(resources)
	resources = q.QueryFilter.Page(resources)
	resp := buildList(resources)
	resp["total"] = total
	return resp, nil
}

// APIServerQuery represents a query for multicluster resources.
type APIServerQuery struct {
	ClusterdNamespaces []*clusterRes.ClusterNamespaces
	QueryFilter        QueryFilter
	Limit              int
	Offset             int
}

// NewAPIServerQuery creates a new query for multicluster resources.
func NewAPIServerQuery(ns []*clusterRes.ClusterNamespaces, filter QueryFilter) Query {
	return &APIServerQuery{
		ClusterdNamespaces: ns,
		QueryFilter:        filter,
	}
}

// Fetch fetches multicluster resources.
func (q *APIServerQuery) Fetch(ctx context.Context, groupVersion, kind string) (map[string]interface{}, error) {
	var err error
	if q.ClusterdNamespaces, err = checkMultiClusterAccess(ctx, kind, q.ClusterdNamespaces); err != nil {
		return nil, err
	}
	log.Info(ctx, "fetch multi cluster resources, kind: %s, clusterdNamespaces: %v", kind, q.ClusterdNamespaces)
	resources, err := listResource(ctx, q.ClusterdNamespaces, groupVersion, kind, metav1.ListOptions{
		LabelSelector: q.QueryFilter.LabelSelectorString()})
	if err != nil {
		return nil, err
	}
	resources = ApplyFilter(resources, q.QueryFilter.CreatorFilter, q.QueryFilter.NameFilter,
		q.QueryFilter.StatusFilter, q.QueryFilter.IPFilter)
	total := len(resources)
	resources = q.QueryFilter.Page(resources)
	resp := buildList(resources)
	resp["total"] = total
	return resp, nil
}

// listResource 列出多集群资源
func listResource(ctx context.Context, clusterdNamespaces []*clusterRes.ClusterNamespaces, groupVersion, kind string,
	opts metav1.ListOptions) ([]*storage.Resource, error) {
	errGroups := errgroup.Group{}
	errGroups.SetLimit(10)
	result := []*storage.Resource{}
	mux := sync.Mutex{}
	for _, v := range clusterdNamespaces {
		ns := v
		errGroups.Go(func() error {
			resources, err := listNamespaceResources(ctx, ns.ClusterID, ns.Namespaces, groupVersion, kind, opts)
			if err != nil {
				return err
			}
			mux.Lock()
			defer mux.Unlock()
			result = append(result, resources...)
			return nil
		})
	}
	if err := errGroups.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

// listNamespaceResources 列出某个集群下某些命名空间的资源
func listNamespaceResources(ctx context.Context, clusterID string, namespaces []string, groupVersion, kind string,
	opts metav1.ListOptions) ([]*storage.Resource, error) {
	clusterConf := res.NewClusterConf(clusterID)
	k8sRes, err := res.GetGroupVersionResource(ctx, clusterConf, kind, groupVersion)
	if err != nil {
		log.Error(ctx, "get group version resource error, %v", err)
		// 多集群查询场景，如果 crd 不存在，直接返回空
		if strings.Contains(err.Error(), "not found in cluster") {
			return nil, nil
		}
		if strings.Contains(err.Error(), "the server could not find the requested resource") {
			return nil, nil
		}
		return nil, err
	}
	if len(namespaces) == 0 {
		namespaces = append(namespaces, "")
	}
	errGroups := errgroup.Group{}
	errGroups.SetLimit(20)
	result := []*storage.Resource{}
	mux := sync.Mutex{}
	// 根据命名空间列表，并发查询资源
	for _, v := range namespaces {
		ns := v
		errGroups.Go(func() error {
			ret, innerErr := cli.NewResClient(clusterConf, k8sRes).ListAllWithoutPerm(ctx, ns, opts)
			if innerErr != nil {
				return innerErr
			}
			if len(ret) == 0 {
				return nil
			}
			mux.Lock()
			defer mux.Unlock()
			for _, item := range ret {
				result = append(result, &storage.Resource{ClusterID: clusterID, ResourceType: kind,
					Data: item.UnstructuredContent()})
			}
			return nil
		})
	}
	if err = errGroups.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

// BuildList build list response data
func buildList(resources []*storage.Resource) map[string]interface{} {
	result := map[string]interface{}{}
	if len(resources) == 0 {
		return result
	}
	manifestExt := map[string]interface{}{}
	manifest := map[string]interface{}{}
	manifestItems := []interface{}{}
	// 获取 apiVersion
	apiVersion := mapx.GetStr(resources[0].Data, "apiVersion")
	kind := resources[0].ResourceType
	formatFunc := formatter.GetFormatFunc(kind, apiVersion)
	pruneFunc := formatter.GetPruneFunc(kind)
	// 遍历列表中的每个资源，生成 manifestExt
	for _, item := range resources {
		uid, _ := mapx.GetItems(item.Data, "metadata.uid")
		ext := formatFunc(item.Data)
		ext["clusterID"] = item.ClusterID
		manifestExt[uid.(string)] = ext
		manifestItems = append(manifestItems, pruneFunc(item.Data))
	}
	manifest["items"] = manifestItems
	return map[string]interface{}{"manifest": manifest, "manifestExt": manifestExt}
}

// checkMultiClusterAccess 检查多集群共享集群中的资源访问权限
// NOCC:CCN_threshold(设计如此)
// nolint
func checkMultiClusterAccess(ctx context.Context, kind string, clusters []*clusterRes.ClusterNamespaces) (
	[]*clusterRes.ClusterNamespaces, error) {
	newClusters := []*clusterRes.ClusterNamespaces{}
	projInfo, err := project.FromContext(ctx)
	if err != nil {
		return nil, errorx.New(errcode.General, i18n.GetMsg(ctx, "由 Context 获取项目信息失败"))
	}

	// 共享集群过滤
	for _, v := range clusters {
		clusterInfo, err := cluster.GetClusterInfo(ctx, v.ClusterID)
		if err != nil {
			return nil, err
		}
		// 集群不存在或者不是运行状态，则忽略
		if clusterInfo.Status != cluster.ClusterStatusRunning {
			continue
		}
		if !clusterInfo.IsShared {
			newClusters = append(newClusters, v)
			continue
		}

		// 共享集群，如果没有命名空间，则直接返回
		var nss []string
		for _, ns := range v.Namespaces {
			if ns == "" {
				continue
			}
			nss = append(nss, ns)
		}
		if len(nss) == 0 {
			clusterNs, err := project.GetProjectNamespace(ctx, projInfo.Code, v.ClusterID)
			if err != nil {
				log.Error(ctx, "get project %s cluster %s ns failed, %v", projInfo.Code, v.ClusterID, err)
				continue
			}
			if len(clusterNs) == 0 {
				continue
			}
			for _, nsItem := range clusterNs {
				if !nsItem.IsActive() {
					continue
				}
				nss = append(nss, nsItem.Name)
			}
		}

		// SC 允许用户查看
		if slice.StringInSlice(kind, cluster.SharedClusterBypassNativeKinds) {
			newClusters = append(newClusters, &clusterRes.ClusterNamespaces{ClusterID: v.ClusterID, Namespaces: nss})
			continue
		}
		// 共享集群不允许访问的资源类型
		if !slice.StringInSlice(kind, cluster.SharedClusterEnabledNativeKinds) &&
			!slice.StringInSlice(kind, config.G.SharedCluster.EnabledCObjKinds) {
			continue
		}
		// 其他可访问的资源类型
		newClusters = append(newClusters, &clusterRes.ClusterNamespaces{ClusterID: v.ClusterID, Namespaces: nss})
	}

	// iam 权限过滤，只允许访问有权限的集群和命名空间
	errGroups := errgroup.Group{}
	errGroups.SetLimit(10)
	result := []*clusterRes.ClusterNamespaces{}
	mux := sync.Mutex{}
	for _, v := range newClusters {
		cls := v
		errGroups.Go(func() error {
			permCtx := clusterAuth.NewPermCtx(
				ctx.Value(ctxkey.UsernameKey).(string), projInfo.ID, cls.ClusterID,
			)
			if allow, err := iam.NewClusterPerm(projInfo.ID).CanView(permCtx); err != nil {
				return nil
			} else if !allow {
				return nil
			}
			mux.Lock()
			defer mux.Unlock()
			result = append(result, cls)
			return nil
		})
	}
	if err := errGroups.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

// getScopedByKind 根据资源类型获取作用域
func getScopedByKind(kind string) apiextensions.ResourceScope {
	if slice.StringInSlice(kind, []string{"PersistentVolume", "StorageClass", "CustomResourceDefinition"}) {
		return apiextensions.ClusterScoped
	}
	return apiextensions.NamespaceScoped
}

// filterClusteredNamespace 过滤集群命名空间，如果是集群域资源，则不能带命名空间
func filterClusteredNamespace(clusterNs []*clusterRes.ClusterNamespaces,
	scoped string) []*clusterRes.ClusterNamespaces {
	if scoped == string(apiextensions.NamespaceScoped) {
		return clusterNs
	}
	newClusterNs := []*clusterRes.ClusterNamespaces{}
	for _, v := range clusterNs {
		newClusterNs = append(newClusterNs, &clusterRes.ClusterNamespaces{ClusterID: v.ClusterID})
	}
	return newClusterNs
}
