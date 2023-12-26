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
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

const (
	// EmptyCreator 无创建者
	EmptyCreator = "--"

	// LabelCreator 创建者标签
	LabelCreator = "io.tencent.paas.creator"

	// OpEQ 等于
	OpEQ = "="
	// OpIn 包含
	OpIn = "In"
	// OpNotIn 不包含
	OpNotIn = "NotIn"
	// OpExists 存在
	OpExists = "Exists"
	// OpDoesNotExist 不存在
	OpDoesNotExist = "DoesNotExist"
)

// Query represents a query for multicluster resources.
type Query interface {
	Fetch(ctx context.Context, groupVersion, kind string) (map[string]interface{}, error)
}

// QueryFilter 查询条件
type QueryFilter struct {
	Creator       []string // -- 代表无创建者
	Name          string
	LabelSelector []*clusterRes.LabelSelector
	Limit         int
	Offset        int
}

// LabelSelectorString 转换为标签选择器字符串
// 操作符，=, In, NotIn, Exists, DoesNotExist，如果是 Exists/DoesNotExist，如果是, values为空，如果是=，values只有一个值，
// 如果是in/notin，values有多个值
func (f *QueryFilter) LabelSelectorString() string {
	var ls []string
	for _, v := range f.LabelSelector {
		if len(v.Values) == 0 && v.Op != OpExists && v.Op != OpDoesNotExist {
			continue
		}
		switch v.Op {
		case OpEQ:
			ls = append(ls, fmt.Sprintf("%s=%s", v.Key, v.Values[0]))
		case OpIn:
			values := strings.Join(v.Values, ",")
			ls = append(ls, fmt.Sprintf("%s in (%s)", v.Key, values))
		case OpNotIn:
			values := strings.Join(v.Values, ",")
			ls = append(ls, fmt.Sprintf("%s notin (%s)", v.Key, values))
		case OpExists:
			ls = append(ls, v.Key)
		case OpDoesNotExist:
			ls = append(ls, fmt.Sprintf("!%s", v.Key))
		}
	}
	return strings.Join(ls, ",")
}

// ToConditions 转换为查询条件
func (f *QueryFilter) ToConditions() []*operator.Condition {
	conditions := []*operator.Condition{}

	// creator 过滤条件
	var emptyCreator bool
	for _, v := range f.Creator {
		if v == EmptyCreator {
			emptyCreator = true
		}
	}
	if len(f.Creator) > 0 {
		if emptyCreator {
			// 使用全角符号代替 '.',区分字段分隔，无创建者的资源，creator字段为 null
			conditions = append(conditions, operator.NewLeafCondition(
				operator.Eq, map[string]interface{}{
					"data.metadata.annotations." + mapx.ConvertPath(LabelCreator): nil}))
		} else {
			conditions = append(conditions, operator.NewLeafCondition(
				operator.In, map[string]interface{}{
					"data.metadata.annotations." + mapx.ConvertPath(LabelCreator): f.Creator}))
		}
	}

	// name 过滤条件
	if f.Name != "" {
		conditions = append(conditions, operator.NewLeafCondition(
			operator.Con, map[string]string{"data.metadata.name": f.Name}))
	}

	// labelSelector 过滤条件
	for _, v := range f.LabelSelector {
		if len(v.Values) == 0 && v.Op != OpExists && v.Op != OpDoesNotExist {
			continue
		}
		switch v.Op {
		case OpEQ:
			conditions = append(conditions, operator.NewLeafCondition(
				operator.Eq, map[string]interface{}{
					fmt.Sprintf("data.metadata.labels.%s", mapx.ConvertPath(v.Key)): v.Values[0]}))
		case OpIn:
			conditions = append(conditions, operator.NewLeafCondition(
				operator.In, map[string]interface{}{
					fmt.Sprintf("data.metadata.labels.%s", mapx.ConvertPath(v.Key)): v.Values}))
		case OpNotIn:
			conditions = append(conditions, operator.NewLeafCondition(
				operator.Nin, map[string]interface{}{
					fmt.Sprintf("data.metadata.labels.%s", mapx.ConvertPath(v.Key)): v.Values}))
		case OpExists:
			conditions = append(conditions, operator.NewLeafCondition(
				operator.Ext, map[string]interface{}{
					fmt.Sprintf("data.metadata.labels.%s", mapx.ConvertPath(v.Key)): ""}))
		case OpDoesNotExist:
			conditions = append(conditions, operator.NewLeafCondition(
				operator.Eq, map[string]interface{}{
					fmt.Sprintf("data.metadata.labels.%s", mapx.ConvertPath(v.Key)): nil}))
		}
	}
	return conditions
}

// CreatorFilter 创建者过滤器
func (f *QueryFilter) CreatorFilter(resources []*storage.Resource) []*storage.Resource {
	result := []*storage.Resource{}
	if len(f.Creator) == 0 {
		return resources
	}

	var emptyCreator bool
	for _, v := range f.Creator {
		if v == EmptyCreator {
			emptyCreator = true
		}
	}

	if emptyCreator {
		for _, v := range resources {
			if mapx.GetStr(v.Data, []string{"metadata", "annotations", mapx.ConvertPath(LabelCreator)}) == "" {
				result = append(result, v)
			}
		}
	} else {
		for _, v := range resources {
			if slice.StringInSlice(
				mapx.GetStr(v.Data, []string{"metadata", "annotations", mapx.ConvertPath(LabelCreator)}), f.Creator) {
				result = append(result, v)
			}
		}
	}

	return result
}

// NameFilter 名称过滤器
func (f *QueryFilter) NameFilter(resources []*storage.Resource) []*storage.Resource {
	result := []*storage.Resource{}
	if f.Name == "" {
		return resources
	}
	for _, v := range resources {
		if strings.Contains(mapx.GetStr(v.Data, "metadata.name"), f.Name) {
			result = append(result, v)
		}
	}
	return result
}

// SortResourceByName 按照名称排序
type SortResourceByName []*storage.Resource

func (s SortResourceByName) Len() int {
	return len(s)
}

func (s SortResourceByName) Less(i, j int) bool {
	return mapx.GetStr(s[i].Data, "metadata.name") < mapx.GetStr(s[j].Data, "metadata.name")
}

func (s SortResourceByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Page 分页
func (f *QueryFilter) Page(resources []*storage.Resource) []*storage.Resource {
	sort.Sort(SortResourceByName(resources))
	if f.Offset >= len(resources) {
		return []*storage.Resource{}
	}
	end := f.Offset + f.Limit
	if end > len(resources) {
		end = len(resources)
	}
	return resources[f.Offset:end]
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
	// TODO check access
	if err := checkMultiClusterAccess(ctx, kind, q.ClusterdNamespaces); err != nil {
		return nil, err
	}
	clusteredNamespaces := []storage.ClusteredNamespaces{}
	for _, v := range q.ClusterdNamespaces {
		clusteredNamespaces = append(clusteredNamespaces, storage.ClusteredNamespaces{
			ClusterID:  v.GetClusterID(),
			Namespaces: v.GetNamespaces(),
		})
	}
	resource, total, err := storage.ListMultiClusterResources(ctx, storage.ListMultiClusterResourcesReq{
		Kind:                kind,
		Limit:               q.QueryFilter.Limit,
		Offset:              q.QueryFilter.Offset,
		ClusteredNamespaces: clusteredNamespaces,
		Conditions:          q.QueryFilter.ToConditions(),
	})
	if err != nil {
		return nil, err
	}
	resp := buildList(resource)
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
	if err := checkMultiClusterAccess(ctx, kind, q.ClusterdNamespaces); err != nil {
		return nil, err
	}
	resources, err := listResource(ctx, q.ClusterdNamespaces, groupVersion, kind, metav1.ListOptions{
		LabelSelector: q.QueryFilter.LabelSelectorString()})
	if err != nil {
		return nil, err
	}
	resources = q.QueryFilter.CreatorFilter(resources)
	resources = q.QueryFilter.NameFilter(resources)
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
			ret, innerErr := cli.NewResClient(clusterConf, k8sRes).ListWithoutPerm(ctx, ns, opts)
			if innerErr != nil {
				return innerErr
			}
			if len(ret.Items) == 0 {
				return nil
			}
			mux.Lock()
			defer mux.Unlock()
			for _, item := range ret.Items {
				result = append(result, &storage.Resource{ClusterID: clusterID, Data: item.UnstructuredContent()})
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
	// 遍历列表中的每个资源，生成 manifestExt
	for _, item := range resources {
		uid, _ := mapx.GetItems(item.Data, "metadata.uid")
		ext := formatFunc(item.Data)
		ext["clusterID"] = item.ClusterID
		manifestExt[uid.(string)] = ext
		manifestItems = append(manifestItems, item.Data)
	}
	manifest["items"] = manifestItems
	// 处理pod资源manifest返回数据过多问题
	newManifest := formatter.FormatPodManifestRes(kind, manifest)
	return map[string]interface{}{"manifest": newManifest, "manifestExt": manifestExt}
}

// checkMultiClusterAccess 检查多集群共享集群中的资源访问权限
func checkMultiClusterAccess(ctx context.Context, kind string, clusters []*clusterRes.ClusterNamespaces) error {
	checkShare := false
	for _, v := range clusters {
		clusterInfo, err := cluster.GetClusterInfo(ctx, v.ClusterID)
		if err != nil {
			return err
		}
		if clusterInfo.IsShared {
			checkShare = true
			break
		}
	}
	if !checkShare {
		return nil
	}
	// SC 允许用户查看，PV 返回空，不报错
	if slice.StringInSlice(kind, cluster.SharedClusterBypassNativeKinds) {
		return nil
	}
	// 不允许的资源类型，直接抛出错误
	if !slice.StringInSlice(kind, cluster.SharedClusterEnabledNativeKinds) &&
		!slice.StringInSlice(kind, config.G.SharedCluster.EnabledCObjKinds) {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "该请求资源类型 %s 在共享集群中不可用"), kind)
	}
	return nil
}
