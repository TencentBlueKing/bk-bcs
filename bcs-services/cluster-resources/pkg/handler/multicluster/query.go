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
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"golang.org/x/sync/errgroup"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm"
	clusterAuth "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/cluster"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// Query represents a query for multicluster resources.
type Query interface {
	Fetch(ctx context.Context, groupVersion, kind string) (map[string]interface{}, error)
	FetchPreferred(
		ctx context.Context, gvr *schema.GroupVersionResource) (map[string]interface{}, error)
}

// StorageQuery represents a query for multicluster resources.
type StorageQuery struct {
	ClusterdNamespaces []*clusterRes.ClusterNamespaces
	QueryFilter        QueryFilter
	ViewFilter         QueryFilter
}

// NewStorageQuery creates a new query for multicluster resources.
func NewStorageQuery(ns []*clusterRes.ClusterNamespaces, queryFilter, viewFilter QueryFilter) Query {
	return &StorageQuery{
		ClusterdNamespaces: ns,
		QueryFilter:        queryFilter,
		ViewFilter:         viewFilter,
	}
}

// Fetch fetches multicluster resources.
func (q *StorageQuery) Fetch(ctx context.Context, groupVersion, kind string) (map[string]interface{}, error) {
	var (
		err      error
		applyURL string
	)
	if q.ClusterdNamespaces, applyURL, err = checkMultiClusterAccess(ctx, kind, q.ClusterdNamespaces); err != nil {
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
	// status查询需要全量获取
	var allRecord bool
	if len(q.QueryFilter.Status) != 0 {
		allRecord = true
	}
	resources, total, err := storage.ListAllMultiClusterResources(ctx, allRecord, storage.ListMultiClusterResourcesReq{
		Kind:                kind,
		ClusteredNamespaces: clusteredNamespaces,
		Conditions:          q.toConditions(),
		Sort:                q.QueryFilter.SortByStorage(),
		Offset:              q.QueryFilter.Offset,
		Limit:               q.QueryFilter.Limit,
	})
	if err != nil {
		return nil, err
	}
	if allRecord {
		resources = ApplyFilter(resources, q.QueryFilter.StatusFilter)
		total = len(resources)
		resources = q.QueryFilter.PageWithoutSort(resources)
	}
	resp := buildList(ctx, resources)
	resp["total"] = total
	resp["perms"] = map[string]interface{}{"applyURL": applyURL}
	return resp, nil
}

// toConditions storage query
func (q *StorageQuery) toConditions() []*operator.Condition {
	conditions := []*operator.Condition{}
	// creator 过滤条件
	conditions = append(conditions, q.creatorToCondition()...)

	// name 过滤条件
	conditions = append(conditions, q.nameToCondition()...)

	// label 过滤条件
	conditions = append(conditions, q.labelToCondition()...)

	// createSource 过滤条件
	conditions = append(conditions, q.createSourceToCondition()...)

	// ip 过滤条件
	conditions = append(conditions, q.ipToCondition()...)

	return conditions
}

// FetchPreferred fetches multicluster resources.
func (q *StorageQuery) FetchPreferred(
	ctx context.Context, gvr *schema.GroupVersionResource) (map[string]interface{}, error) {
	// 目前仅仅只是添加此方法，暂时不做实现
	return map[string]interface{}{}, nil
}

// APIServerQuery represents a query for multicluster resources.
type APIServerQuery struct {
	ClusterdNamespaces []*clusterRes.ClusterNamespaces
	QueryFilter        QueryFilter
	ViewFilter         QueryFilter
}

// NewAPIServerQuery creates a new query for multicluster resources.
func NewAPIServerQuery(ns []*clusterRes.ClusterNamespaces, queryFilter, viewFilter QueryFilter) Query {
	return &APIServerQuery{
		ClusterdNamespaces: ns,
		QueryFilter:        queryFilter,
		ViewFilter:         viewFilter,
	}
}

// Fetch fetches multicluster resources.
func (q *APIServerQuery) Fetch(ctx context.Context, groupVersion, kind string) (map[string]interface{}, error) {
	var (
		err      error
		applyURL string
	)
	if q.ClusterdNamespaces, applyURL, err = checkMultiClusterAccess(ctx, kind, q.ClusterdNamespaces); err != nil {
		return nil, err
	}
	log.Info(ctx, "fetch multi cluster resources, kind: %s, clusterdNamespaces: %v", kind, q.ClusterdNamespaces)
	resources, err := listResource(ctx, q.ClusterdNamespaces, groupVersion, kind, metav1.ListOptions{
		LabelSelector: q.ViewFilter.LabelSelectorString()})
	if err != nil {
		return nil, err
	}
	resources = ApplyFilter(resources, q.ViewFilter.CreatorFilter, q.ViewFilter.NameFilter,
		q.ViewFilter.CreateSourceFilter)
	// 第二次过滤
	resources = ApplyFilter(resources, q.QueryFilter.CreatorFilter, q.QueryFilter.NameFilter,
		q.QueryFilter.StatusFilter, q.QueryFilter.LabelSelectorFilter, q.QueryFilter.IPFilter,
		q.QueryFilter.CreateSourceFilter)
	total := len(resources)
	resources = q.QueryFilter.Page(resources)
	resp := buildList(ctx, resources)
	resp["total"] = total
	resp["perms"] = map[string]interface{}{"applyURL": applyURL}
	return resp, nil
}

// FetchPreferred fetches multicluster resources.
func (q *APIServerQuery) FetchPreferred(
	ctx context.Context, gvr *schema.GroupVersionResource) (map[string]interface{}, error) {
	var (
		err      error
		applyURL string
	)
	// kind 设置为空跳过某些特殊检查检查
	if q.ClusterdNamespaces, applyURL, err = checkMultiClusterAccess(ctx, "", q.ClusterdNamespaces); err != nil {
		return nil, err
	}
	log.Info(ctx, "fetch multi cluster resources, gvr: %s, clusterdNamespaces: %v", gvr, q.ClusterdNamespaces)
	resources, err := listResourcePreferred(ctx, q.ClusterdNamespaces, gvr, metav1.ListOptions{
		LabelSelector: q.ViewFilter.LabelSelectorString()})
	if err != nil {
		return nil, err
	}
	resources = ApplyFilter(resources, q.ViewFilter.CreatorFilter, q.ViewFilter.NameFilter,
		q.ViewFilter.CreateSourceFilter)
	// 第二次过滤
	resources = ApplyFilter(resources, q.QueryFilter.CreatorFilter, q.QueryFilter.NameFilter,
		q.QueryFilter.StatusFilter, q.QueryFilter.LabelSelectorFilter, q.QueryFilter.IPFilter,
		q.QueryFilter.CreateSourceFilter)
	total := len(resources)
	resources = q.QueryFilter.Page(resources)
	resp := buildList(ctx, resources)
	resp["total"] = total
	resp["perms"] = map[string]interface{}{"applyURL": applyURL}
	return resp, nil
}

// FetchApiResources fetches api resources.
func (q *APIServerQuery) FetchApiResources(ctx context.Context, kind string) (map[string]interface{}, error) {
	var (
		err      error
		applyURL string
	)
	if q.ClusterdNamespaces, applyURL, err = checkMultiClusterAccess(ctx, kind, q.ClusterdNamespaces); err != nil {
		return nil, err
	}
	log.Info(ctx, "fetch multi cluster resources, kind: %s, clusterdNamespaces: %v", kind, q.ClusterdNamespaces)
	resources, err := listApiResource(ctx, q.ClusterdNamespaces, kind, metav1.ListOptions{
		LabelSelector: q.ViewFilter.LabelSelectorString()})
	if err != nil {
		return nil, err
	}
	resp := make(map[string]interface{}, 0)
	resp["resources"] = resources
	resp["perms"] = map[string]interface{}{"applyURL": applyURL}
	return resp, nil
}

// FetchApiResourcesPreferred fetches api resources.
func FetchApiResourcesPreferred(
	ctx context.Context, onlyCrd bool, resName string, clusters []string) (map[string]interface{}, error) {
	var (
		err           error
		applyURL      string
		accessCluster []string
		kind          string
	)
	// 仅获取crd资源
	if onlyCrd {
		kind = constants.CRD
	}

	if accessCluster, applyURL, err = checkMultiOnlyClusterAccess(ctx, kind, clusters); err != nil {
		return nil, err
	}

	resources, err := listClusterApiResource(ctx, onlyCrd, resName, accessCluster)
	if err != nil {
		return nil, err
	}
	resp := make(map[string]interface{}, 0)
	resp["resources"] = resources
	resp["perms"] = map[string]interface{}{"applyURL": applyURL}
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

// listResourcePreferred 列出多集群资源
func listResourcePreferred(ctx context.Context, clusterdNamespaces []*clusterRes.ClusterNamespaces,
	gvr *schema.GroupVersionResource, opts metav1.ListOptions) ([]*storage.Resource, error) {
	errGroups := errgroup.Group{}
	errGroups.SetLimit(10)
	result := []*storage.Resource{}
	mux := sync.Mutex{}
	for _, v := range clusterdNamespaces {
		ns := v
		errGroups.Go(func() error {
			resources, err := listNamespaceResourcesPreferred(ctx, ns.Namespaces, ns.ClusterID, gvr, opts)
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

// listApiResource 列出多集群资源
func listApiResource(ctx context.Context, clusterdNamespaces []*clusterRes.ClusterNamespaces, kind string,
	opts metav1.ListOptions) ([]interface{}, error) {
	errGroups := errgroup.Group{}
	errGroups.SetLimit(10)
	results := make([]interface{}, 0)
	mux := sync.Mutex{}
	for _, v := range clusterdNamespaces {
		ns := v
		errGroups.Go(func() error {
			result := make(map[string]interface{}, 0)
			resources, err := listNamespaceApiResources(ctx, ns.ClusterID, ns.Namespaces, kind, opts)
			if err != nil {
				return err
			}
			// 需要转一下, res.GroupKindVersionResourc -> []map[string] interface
			// pbstruct.Map2pbStruct无法识别res.GroupKindVersionResource类型
			for key, value := range resources {
				result[key] = []map[string]interface{}(value)
			}
			mux.Lock()
			defer mux.Unlock()
			results = append(results, result)
			return nil
		})
	}

	if err := errGroups.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

// listClusterApiResource 列出多集群资源
func listClusterApiResource(
	ctx context.Context, onlyCrd bool, resName string, clusters []string) (map[string]interface{}, error) {
	errGroups := errgroup.Group{}
	errGroups.SetLimit(10)
	resources := make([]map[string]interface{}, 0)
	mux := sync.Mutex{}
	for _, v := range clusters {
		clusterID := v
		errGroups.Go(func() error {
			var (
				res []map[string]interface{}
				err error
			)
			// 仅获取crd资源
			if onlyCrd {
				res, err = getClusterCrdResources(ctx, clusterID, resName)
				if err != nil {
					return err
				}
			} else {
				// 获取所有api-resources
				res, err = getClusterApiResources(ctx, clusterID, resName)
				if err != nil {
					return err
				}
			}
			if len(res) == 0 {
				return nil
			}

			mux.Lock()
			defer mux.Unlock()
			resources = append(resources, res...)
			return nil
		})
	}

	if err := errGroups.Wait(); err != nil {
		return nil, err
	}

	result := make(map[string]interface{}, 0)
	uniq := filterRes()
	// 返回格式key: group/version, value: []map[string] interface
	for _, value := range resources {
		// group+version+resource+kind确定一条唯一的资源,
		// 存在不同集群group+version+resource一致，Kind不一致的情况
		uniqKey := filepath.Join(value["group"].(string), value["version"].(string),
			value["resource"].(string), value["kind"].(string))
		if _, ok := uniq[uniqKey]; ok {
			continue
		} else {
			uniq[uniqKey] = struct{}{}
			key := filepath.Join(value["group"].(string), value["version"].(string))
			if r, ok := result[key].([]map[string]interface{}); ok {
				result[key] = append(r, value)
			} else {
				result[key] = []map[string]interface{}{
					value,
				}
			}
		}

	}
	return result, nil
}

// 过滤掉一些不支持get权限操作的资源以及workload资源
func filterRes() map[string]struct{} {
	uniq := make(map[string]struct{}, 0)
	// 如果是需要所有的Api resource，则不进行过滤
	if config.G.MultiCluster.AllApiResources {
		return uniq
	}
	uniq = map[string]struct{}{
		"apps/v1/deployments/Deployment":   {},
		"apps/v1/statefulsets/StatefulSet": {},
		"apps/v1/daemonsets/DaemonSet":     {},
		"batch/v1/jobs/Job":                {},
		"batch/v1/cronjobs/CronJob":        {},
		"v1/pods/Pod":                      {},

		"networking.k8s.io/v1/ingresses/Ingress": {},
		"v1/services/Service":                    {},
		"v1/endpoints/Endpoints":                 {},

		"bk.tencent.com/v1alpha1/bscpconfigs/BscpConfig": {},
		"v1/configmaps/ConfigMap":                        {},
		"v1/secrets/Secret":                              {},

		"v1/persistentvolumes/PersistentVolume":           {},
		"v1/persistentvolumeclaims/PersistentVolumeClaim": {},
		"storage.k8s.io/v1/storageclasses/StorageClass":   {},

		"v1/serviceaccounts/ServiceAccount": {},

		"autoscaling/v1/horizontalpodautoscalers/HorizontalPodAutoscaler": {},

		"apiextensions.k8s.io/v1/customresourcedefinitions/CustomResourceDefinition": {},

		"tkex.tencent.com/v1alpha1/gamedeployments/GameDeployment":   {},
		"tkex.tencent.com/v1alpha1/gamestatefulsets/GameStatefulset": {},
		"tkex.tencent.com/v1alpha1/hooktemplates/HookTemplates":      {},

		"v1/componentstatuses/ComponentStatus":                                       {},
		"authorization.k8s.io/v1/localsubjectaccessreviews/LocalSubjectAccessReview": {},
		"authorization.k8s.io/v1/subjectaccessreviews/SubjectAccessReview":           {},
		"authorization.k8s.io/v1/selfsubjectaccessreviews/SelfSubjectAccessReview":   {},
		"authorization.k8s.io/v1/selfsubjectrulesreviews/SelfSubjectRulesReview":     {},
		"authentication.k8s.io/v1/tokenreviews/TokenReview":                          {},
	}

	return uniq
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
	// 如果命名空间为空，则查询所有命名空间，如果命名空间数量大于 5，则查全部命名空间并最后筛选命名空间，这样能减少并发请求
	filterNamespace := namespaces
	if len(namespaces) == 0 {
		filterNamespace = []string{""}
	}
	errGroups := errgroup.Group{}
	errGroups.SetLimit(5)
	result := []*storage.Resource{}
	mux := sync.Mutex{}
	// 根据命名空间列表，并发查询资源
	for _, v := range filterNamespace {
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
				// 过滤命名空间，如果命名空间为空，则表示查询全部命名空间或者集群域资源，这种直接通过。其他情况则筛选命名空间
				if len(namespaces) == 0 || slice.StringInSlice(item.GetNamespace(), namespaces) {
					result = append(result, &storage.Resource{ClusterID: clusterID, ResourceType: kind,
						Data: item.UnstructuredContent()})
				}
			}
			return nil
		})
	}
	if err = errGroups.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

// listNamespaceResourcesPreferred 列出某个集群下某些命名空间的资源
func listNamespaceResourcesPreferred(ctx context.Context, namespaces []string,
	clusterID string, gvr *schema.GroupVersionResource, opts metav1.ListOptions) ([]*storage.Resource, error) {

	clusterConf := res.NewClusterConf(clusterID)

	// 如果命名空间为空，则查询所有命名空间，如果命名空间数量大于 5，则查全部命名空间并最后筛选命名空间，这样能减少并发请求
	filterNamespace := namespaces
	if len(namespaces) == 0 {
		filterNamespace = []string{""}
	}
	errGroups := errgroup.Group{}
	errGroups.SetLimit(5)
	result := []*storage.Resource{}
	mux := sync.Mutex{}
	// 根据命名空间列表，并发查询资源
	for _, v := range filterNamespace {
		ns := v
		errGroups.Go(func() error {
			ret, innerErr := cli.NewResClient(clusterConf, *gvr).ListAllWithoutPermPreferred(ctx, ns, opts)
			if innerErr != nil {
				return innerErr
			}
			if len(ret) == 0 {
				return nil
			}
			mux.Lock()
			defer mux.Unlock()
			for _, item := range ret {
				// 过滤命名空间，如果命名空间为空，则表示查询全部命名空间或者集群域资源，这种直接通过。其他情况则筛选命名空间
				if len(namespaces) == 0 || slice.StringInSlice(item.GetNamespace(), namespaces) {
					result = append(result, &storage.Resource{ClusterID: clusterID, Data: item.UnstructuredContent()})
				}
			}
			return nil
		})
	}
	if err := errGroups.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

// listNamespaceApiResources 列出某个集群下某些命名空间的api资源
func listNamespaceApiResources(ctx context.Context, clusterID string, namespaces []string, kind string,
	opts metav1.ListOptions) (map[string]res.GroupKindVersionResource, error) {
	clusterConf := res.NewClusterConf(clusterID)
	k8sResources, err := res.GetApiResources(ctx, clusterConf, kind, "")
	if err != nil {
		log.Error(ctx, "get api resource error, %v", err)
		// 多集群查询场景，如果 crd 不存在，直接返回空
		if strings.Contains(err.Error(), "the server could not find the requested resource") {
			return nil, nil
		}
		return nil, err
	}

	// 默认列出所有api-resources
	if kind == "" {
		return k8sResources, nil
	}

	// 仅列出crd资源的情况下，只有一条内容
	if len(k8sResources) != 1 {
		return map[string]res.GroupKindVersionResource{}, nil
	}

	k8sRes := schema.GroupVersionResource{}
	for _, value := range k8sResources {
		k8sRes.Group = mapx.GetStr(value[0], "group")
		k8sRes.Version = mapx.GetStr(value[0], "version")
		k8sRes.Resource = mapx.GetStr(value[0], "resource")
	}

	// 如果命名空间为空，则查询所有命名空间，如果命名空间数量大于 5，则查全部命名空间并最后筛选命名空间，这样能减少并发请求
	filterNamespace := namespaces
	if len(namespaces) == 0 {
		filterNamespace = []string{""}
	}
	errGroups := errgroup.Group{}
	errGroups.SetLimit(5)
	result := make(map[string]res.GroupKindVersionResource, 0)
	mux := sync.Mutex{}
	// 根据命名空间列表，并发查询资源
	for _, v := range filterNamespace {
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
				// 过滤命名空间，如果命名空间为空，则表示查询全部命名空间或者集群域资源，这种直接通过。其他情况则筛选命名空间
				if len(namespaces) == 0 || slice.StringInSlice(item.GetNamespace(), namespaces) {
					// 获取特殊字段返回
					groupVersion, crdResources := getCrdResources(item.Object)
					result[groupVersion] = append(result[groupVersion], crdResources)
				}
			}
			return nil
		})
	}
	if err = errGroups.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

// getClusterApiResources 获取某个集群下的所有api资源
func getClusterApiResources(
	ctx context.Context, clusterID string, resName string) ([]map[string]interface{}, error) {
	clusterConf := res.NewClusterConf(clusterID)
	return res.GetApiResourcesByName(ctx, clusterConf, resName)
}

// getClusterCrdResources 获取某个集群下的所有crd资源
func getClusterCrdResources(
	ctx context.Context, clusterID string, resName string) ([]map[string]interface{}, error) {
	crdList, err := cli.ListCrdResources(ctx, clusterID, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return parseCrdToMap(resName, crdList), nil
}

func parseCrdToMap(resName string, crdList *apiextensions.CustomResourceDefinitionList) []map[string]interface{} {
	resp := make([]map[string]interface{}, 0)
	for _, v := range crdList.Items {
		var version string
		for _, vv := range v.Spec.Versions {
			if vv.Storage {
				// 当前使用版本
				version = vv.Name
			}
		}
		name := strings.SplitN(v.Name, ".", 2)
		// 通过资源名称筛选
		if resName == "" {
			if len(name) < 1 {
				continue
			}
			resp = append(resp, map[string]interface{}{
				"group":      v.Spec.Group,
				"kind":       v.Spec.Names.Kind,
				"version":    version,
				"resource":   name[0],
				"namespaced": v.Spec.Scope == apiextensions.NamespaceScoped,
			})
			continue
		}

		if resName != "" && resName == name[0] {
			resp = append(resp, map[string]interface{}{
				"group":      v.Spec.Group,
				"kind":       v.Spec.Names.Kind,
				"version":    version,
				"resource":   name[0],
				"namespaced": v.Spec.Scope == apiextensions.NamespaceScoped,
			})
			return resp
		}

	}
	return resp
}

// 获取crd 中的相关资源
func getCrdResources(object map[string]interface{}) (string, map[string]interface{}) {
	resouces := mapx.GetStr(object, "metadata.name")
	if resouces != "" {
		s := strings.Split(resouces, ".")
		resouces = s[0]
	}
	group := mapx.GetStr(object, "spec.group")
	kind := mapx.GetStr(object, "spec.names.kind")
	var version string
	versions := mapx.GetList(object, "status.storedVersions")
	if len(versions) > 0 {
		if _, ok := versions[0].(string); ok {
			version = versions[0].(string)
		}
	}
	namespaced := mapx.GetStr(object, "spec.scope") == "Namespaced"
	groupVersion := path.Join(group, version)
	return groupVersion, map[string]interface{}{
		"group":      group,
		"kind":       kind,
		"version":    version,
		"resource":   resouces,
		"namespaced": namespaced,
	}
}

// BuildList build list response data
func buildList(ctx context.Context, resources []*storage.Resource) map[string]interface{} {
	result := map[string]interface{}{}
	if len(resources) == 0 {
		return result
	}
	manifestExt := map[string]interface{}{}
	manifest := map[string]interface{}{}
	manifestItems := []interface{}{}
	// 获取 apiVersion
	apiVersion := mapx.GetStr(resources[0].Data, "apiVersion")
	kind := mapx.GetStr(resources[0].Data, "kind")
	formatFunc := formatter.GetFormatFunc(kind, apiVersion)
	pruneFunc := formatter.GetPruneFunc(kind)
	// 遍历列表中的每个资源，生成 manifestExt
	for _, item := range resources {
		uid, _ := mapx.GetItems(item.Data, "metadata.uid")
		ext := formatFunc(item.Data)
		// 共享集群不展示集群域资源
		clusterInfo, err := cluster.GetClusterInfo(ctx, item.ClusterID)
		if err != nil {
			continue
		}
		if ext["scope"] == "Cluster" && clusterInfo.IsShared {
			continue
		}
		ext["clusterID"] = item.ClusterID
		// 过滤掉不支持uid的资源
		if uidStr, ok := uid.(string); ok {
			manifestExt[uidStr] = ext
			manifestItems = append(manifestItems, pruneFunc(item.Data))
		}
	}
	manifest["items"] = manifestItems
	return map[string]interface{}{"manifest": manifest, "manifestExt": manifestExt}
}

// checkMultiClusterAccess 检查多集群共享集群中的资源访问权限
// NOCC:CCN_threshold(设计如此)
// nolint
func checkMultiClusterAccess(ctx context.Context, kind string, clusters []*clusterRes.ClusterNamespaces) (
	[]*clusterRes.ClusterNamespaces, string, error) {
	newClusters := []*clusterRes.ClusterNamespaces{}
	projInfo, err := project.FromContext(ctx)
	if err != nil {
		return nil, "", errorx.New(errcode.General, i18n.GetMsg(ctx, "由 Context 获取项目信息失败"))
	}

	// 共享集群过滤
	for _, v := range clusters {
		clusterInfo, err := cluster.GetClusterInfo(ctx, v.ClusterID)
		if err != nil {
			return nil, "", err
		}
		// 集群不存在或者不是运行状态，则忽略
		if clusterInfo.Status != cluster.ClusterStatusRunning {
			continue
		}
		if !clusterInfo.IsShared {
			newClusters = append(newClusters, v)
			continue
		}

		// kind为空情况下跳过下面的检查，允许查询各种资源
		if kind == "" {
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
		// 命名空间为空，则查询集群下用户所有命名空间
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
		if slice.StringInSlice(kind, cluster.SharedClusterBypassClusterScopedKinds) {
			newClusters = append(newClusters, &clusterRes.ClusterNamespaces{ClusterID: v.ClusterID})
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
	username := ctx.Value(ctxkey.UsernameKey).(string)
	clusterIDs := make([]string, 0)
	for _, v := range newClusters {
		cls := v
		clusterIDs = append(clusterIDs, cls.ClusterID)
		errGroups.Go(func() error {
			permCtx := clusterAuth.NewPermCtx(username, projInfo.ID, cls.ClusterID)
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
		return nil, "", err
	}

	// get apply url
	permCtx := clusterAuth.NewPermCtx(username, projInfo.ID, strings.Join(clusterIDs, ","))
	applyURL := ""
	if _, err := iam.NewClusterPerm(projInfo.ID).CanView(permCtx); err != nil {
		if perr, ok := err.(*perm.IAMPermError); ok {
			applyURL, _ = perr.ApplyURL()
		}
	}
	return result, applyURL, nil
}

// checkMultiOnlyClusterAccess 检查多集群共享集群中的资源访问权限
func checkMultiOnlyClusterAccess(ctx context.Context, kind string, clusters []string) ([]string, string, error) {
	newClusters := []string{}
	projInfo, err := project.FromContext(ctx)
	if err != nil {
		return nil, "", errorx.New(errcode.General, i18n.GetMsg(ctx, "由 Context 获取项目信息失败"))
	}

	// 共享集群过滤
	for _, v := range clusters {
		clusterInfo, err := cluster.GetClusterInfo(ctx, v)
		if err != nil {
			return nil, "", err
		}
		// 集群不存在或者不是运行状态，则忽略
		if clusterInfo.Status != cluster.ClusterStatusRunning {
			continue
		}
		if !clusterInfo.IsShared {
			newClusters = append(newClusters, v)
			continue
		}

		// kind为空情况下查询所有资源
		if kind == "" {
			newClusters = append(newClusters, v)
			continue
		}

		// SC 允许用户查看
		if slice.StringInSlice(kind, cluster.SharedClusterBypassClusterScopedKinds) {
			newClusters = append(newClusters, v)
			continue
		}
		// 共享集群不允许访问的资源类型
		if !slice.StringInSlice(kind, cluster.SharedClusterEnabledNativeKinds) &&
			!slice.StringInSlice(kind, config.G.SharedCluster.EnabledCObjKinds) {
			continue
		}
		// 其他可访问的资源类型
		newClusters = append(newClusters, v)
	}

	// iam 权限过滤，只允许访问有权限的集群
	errGroups := errgroup.Group{}
	errGroups.SetLimit(10)
	result := []string{}
	mux := sync.Mutex{}
	username := ctx.Value(ctxkey.UsernameKey).(string)
	for _, v := range newClusters {
		errGroups.Go(func() error {
			permCtx := clusterAuth.NewPermCtx(username, projInfo.ID, v)
			if allow, err := iam.NewClusterPerm(projInfo.ID).CanView(permCtx); err != nil {
				return nil
			} else if !allow {
				return nil
			}
			mux.Lock()
			defer mux.Unlock()
			result = append(result, v)
			return nil
		})
	}
	if err := errGroups.Wait(); err != nil {
		return nil, "", err
	}

	// get apply url
	permCtx := clusterAuth.NewPermCtx(username, projInfo.ID, strings.Join(newClusters, ","))
	applyURL := ""
	if _, err := iam.NewClusterPerm(projInfo.ID).CanView(permCtx); err != nil {
		if perr, ok := err.(*perm.IAMPermError); ok {
			applyURL, _ = perr.ApplyURL()
		}
	}
	return result, applyURL, nil
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

// viewQueryToQueryFilter transform view filter to query filter
func viewQueryToQueryFilter(filter *entity.ViewFilter) QueryFilter {
	if filter == nil {
		return QueryFilter{}
	}
	ls := make([]*clusterRes.LabelSelector, 0, len(filter.LabelSelector))
	for _, v := range filter.LabelSelector {
		ls = append(ls, &clusterRes.LabelSelector{
			Key:    v.Key,
			Op:     v.Op,
			Values: v.Values,
		})
	}
	var createSource *clusterRes.CreateSource
	if filter.CreateSource != nil {
		createSource = &clusterRes.CreateSource{
			Source: filter.CreateSource.Source,
		}
		if filter.CreateSource.Template != nil {
			createSource.Template = &clusterRes.Template{
				TemplateName:    filter.CreateSource.Template.TemplateName,
				TemplateVersion: filter.CreateSource.Template.TemplateVersion,
			}
		}

		if filter.CreateSource.Chart != nil {
			createSource.Chart = &clusterRes.Chart{
				ChartName: filter.CreateSource.Chart.ChartName,
			}
		}
	}
	return QueryFilter{
		Creator:       filter.Creator,
		Name:          filter.Name,
		LabelSelector: ls,
		CreateSource:  createSource,
	}
}

// 检查集群是否在黑名单中，如果在黑名单中，则禁止从 api server 查询
// 先匹配项目，再匹配集群
func inBlackList(projectCode string, clusterIDs []*clusterRes.ClusterNamespaces) bool {
	var reg string
	for _, v := range config.G.MultiCluster.BlacklistForAPIServerQuery {
		// 项目配置只允许一个，多个以最后一个为准
		if v.ProjectCode == projectCode {
			reg = v.ClusterIDReg
		}
	}
	if reg == "" {
		return false
	}
	regexp, err := regexp.Compile(reg)
	if err != nil {
		return false
	}
	for _, v := range clusterIDs {
		if regexp.MatchString(v.ClusterID) {
			return true
		}
	}
	return false
}
