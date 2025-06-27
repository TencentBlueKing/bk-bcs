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

// Package multicluster 多集群接口实现
package multicluster

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	respUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resp"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/trans"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/web"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/featureflag"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// Handler Multicluster handler
type Handler struct {
	model store.ClusterResourcesModel
}

// New Multicluster handler
func New(model store.ClusterResourcesModel) *Handler {
	return &Handler{model: model}
}

// FetchMultiClusterResource Fetch multi cluster resource
func (h *Handler) FetchMultiClusterResource(ctx context.Context, req *clusterRes.FetchMultiClusterResourceReq,
	resp *clusterRes.CommonResp) (err error) {
	// 获取视图信息
	view := &entity.View{}
	if req.GetViewID() != "" {
		view, err = h.model.GetView(ctx, req.GetViewID())
		if err != nil {
			return err
		}
	}
	filter := QueryFilter{
		Creator:       req.GetCreator(),
		Name:          req.GetName(),
		CreateSource:  req.GetCreateSource(),
		LabelSelector: req.GetLabelSelector(),
		IP:            req.GetIp(),
		Status:        req.GetStatus(),
		SortBy:        SortBy(req.GetSortBy()),
		Order:         Order(req.GetOrder()),
		Limit:         int(req.GetLimit()),
		Offset:        int(req.GetOffset()),
	}
	clusterNS := filterClusteredNamespace(req.GetClusterNamespaces(), string(getScopedByKind(req.GetKind())))

	// from api server
	var query = NewAPIServerQuery(clusterNS, filter, viewQueryToQueryFilter(view.Filter))
	// 多集群且 bcs-storage 支持的资源则从 bcs-storage 查询
	if slice.StringInSlice(req.GetKind(), config.G.MultiCluster.EnabledQueryFromStorageKinds) && (len(clusterNS) > 1 ||
		inBlackList(req.GetProjectCode(), clusterNS)) {
		// from storage
		query = NewStorageQuery(clusterNS, filter, viewQueryToQueryFilter(view.Filter))
	}

	var data map[string]interface{}
	data, err = query.Fetch(ctx, "", req.GetKind())
	if err != nil {
		return err
	}
	resp.Data, err = pbstruct.Map2pbStruct(data)
	if err != nil {
		return err
	}
	resp.WebAnnotations, err = web.NewAnnos(
		web.NewFeatureFlag(featureflag.FormCreate, true),
	).ToPbStruct()
	return err
}

// FetchMultiClusterCustomResources Fetch multi cluster resource
func (h *Handler) FetchMultiClusterCustomResources(ctx context.Context,
	req *clusterRes.FetchMultiClusterCustomResourcesReq, resp *clusterRes.CommonResp) (err error) {
	// 获取视图信息
	view := &entity.View{}
	if req.GetViewID() != "" {
		view, err = h.model.GetView(ctx, req.GetViewID())
		if err != nil {
			return err
		}
	}
	filter := QueryFilter{
		Creator:       req.GetCreator(),
		Name:          req.GetName(),
		CreateSource:  req.GetCreateSource(),
		LabelSelector: req.GetLabelSelector(),
		SortBy:        SortBy(req.GetSortBy()),
		Order:         Order(req.GetOrder()),
		Limit:         int(req.GetLimit()),
		Offset:        int(req.GetOffset()),
	}

	nsScope := apiextensions.ClusterScoped
	if req.Namespaced {
		nsScope = apiextensions.NamespaceScoped
	}
	clusterNS := filterClusteredNamespace(req.GetClusterNamespaces(), string(nsScope))

	// from api server
	var query = NewAPIServerQuery(clusterNS, filter, viewQueryToQueryFilter(view.Filter))

	// 获取 crd 信息，返回web anno
	var clusterIDs []string
	for _, v := range req.GetClusterNamespaces() {
		clusterIDs = append(clusterIDs, v.ClusterID)
	}
	crdName := req.Resource + "." + req.Group
	crdInfo, err := cli.GetClustersCRDInfoDirect(ctx, clusterIDs, crdName)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		crdInfo = map[string]interface{}{
			"kind": "",
		}
	}

	var data map[string]interface{}
	data, err = query.FetchPreferred(ctx, &schema.GroupVersionResource{
		Group:    req.GetGroup(),
		Version:  req.GetVersion(),
		Resource: req.GetResource(),
	})
	if err != nil {
		return err
	}
	resp.Data, err = pbstruct.Map2pbStruct(data)
	if err != nil {
		return err
	}
	resp.WebAnnotations, err = web.GenListCObjWebAnnos(ctx, data, crdInfo, "")
	return err
}

// FetchMultiClusterApiResources Fetch multi cluster api resources
func (h *Handler) FetchMultiClusterApiResources(ctx context.Context,
	req *clusterRes.FetchMultiClusterApiResourcesReq, resp *clusterRes.CommonResp) (err error) {

	var data map[string]interface{}
	data, err = FetchApiResourcesPreferred(ctx, req.OnlyCrd, req.ResourceName, req.ClusterIDs)
	if err != nil {
		return err
	}

	resp.Data, err = pbstruct.Map2pbStruct(data)
	if err != nil {
		return err
	}

	resp.WebAnnotations, err = web.NewAnnos(
		web.NewFeatureFlag(featureflag.FormCreate, true),
	).ToPbStruct()
	return err
}

// FetchMultiClusterCustomObject Fetch multi cluster custom object
func (h *Handler) FetchMultiClusterCustomObject(ctx context.Context,
	req *clusterRes.FetchMultiClusterCustomObjectReq,
	resp *clusterRes.CommonResp) (err error) {
	// 获取视图信息
	view := &entity.View{}
	if req.GetViewID() != "" {
		view, err = h.model.GetView(ctx, req.GetViewID())
		if err != nil {
			return err
		}
	}
	filter := QueryFilter{
		Creator:       req.GetCreator(),
		Name:          req.GetName(),
		CreateSource:  req.GetCreateSource(),
		LabelSelector: req.GetLabelSelector(),
		IP:            req.GetIp(),
		Status:        req.GetStatus(),
		SortBy:        SortBy(req.GetSortBy()),
		Order:         Order(req.GetOrder()),
		Limit:         int(req.GetLimit()),
		Offset:        int(req.GetOffset()),
	}

	var groupVersion string
	var kind string
	// 则获取 crd 信息
	var clusterIDs []string
	for _, v := range req.GetClusterNamespaces() {
		clusterIDs = append(clusterIDs, v.ClusterID)
	}
	crdInfo, err := cli.GetClustersCRDInfo(ctx, clusterIDs, req.GetCrd())
	if err != nil {
		return nil
	}
	kind, groupVersion = crdInfo["kind"].(string), crdInfo["apiVersion"].(string)
	clusterNS := filterClusteredNamespace(req.GetClusterNamespaces(), crdInfo["scope"].(string))

	var query = NewAPIServerQuery(clusterNS, filter, viewQueryToQueryFilter(view.Filter))
	data, err := query.Fetch(ctx, groupVersion, kind)
	if err != nil {
		return err
	}
	resp.Data, err = pbstruct.Map2pbStruct(data)
	if err != nil {
		return err
	}
	resp.WebAnnotations, err = web.GenListCObjWebAnnos(ctx, data, crdInfo, "")
	return err
}

// MultiClusterResourceCount get mul
func (h *Handler) MultiClusterResourceCount(ctx context.Context, req *clusterRes.MultiClusterResourceCountReq,
	resp *clusterRes.CommonResp) error {
	filter := QueryFilter{
		Creator:       req.GetCreator(),
		Name:          req.GetName(),
		LabelSelector: req.GetLabelSelector(),
		Limit:         1,
		CreateSource:  req.GetCreateSource(),
	}
	var err error

	// 获取视图信息
	view := &entity.View{}
	if req.GetViewID() != "" {
		view, err = h.model.GetView(ctx, req.GetViewID())
		if err != nil {
			return err
		}
	}

	errgroup := errgroup.Group{}
	mux := sync.Mutex{}
	data := map[string]interface{}{}
	for _, kind := range config.G.MultiCluster.EnabledCountKinds {
		kind := kind
		errgroup.Go(func() error {
			clusterNS := filterClusteredNamespace(req.GetClusterNamespaces(), string(getScopedByKind(kind)))
			var query = NewAPIServerQuery(clusterNS, filter, viewQueryToQueryFilter(view.Filter))
			// 多集群且 bcs-storage 支持的资源则从 bcs-storage 查询
			if slice.StringInSlice(kind, config.G.MultiCluster.EnabledQueryFromStorageKinds) && (len(clusterNS) > 1 ||
				inBlackList(req.GetProjectCode(), clusterNS)) {
				query = NewStorageQuery(clusterNS, filter, viewQueryToQueryFilter(view.Filter))
			}
			result, innerErr := query.Fetch(ctx, "", kind)
			mux.Lock()
			defer mux.Unlock()
			if innerErr != nil {
				log.Error(ctx, "query resource %s failed, %v", kind, innerErr)
				data[kind] = 0
				return nil
			}
			data[kind] = result["total"]
			return nil
		})
	}
	_ = errgroup.Wait()
	resp.Data, err = pbstruct.Map2pbStruct(data)
	if err != nil {
		return err
	}
	resp.WebAnnotations, err = web.NewAnnos(
		web.NewFeatureFlag(featureflag.FormCreate, true),
	).ToPbStruct()
	return err
}

// GetApiResourcesObject get api resources object
func (h *Handler) GetApiResourcesObject(ctx context.Context,
	req *clusterRes.GetApiResourcesObjectReq, resp *clusterRes.CommonResp) (err error) {

	resInfo, err := cli.GetResObjectInfo(ctx, req.ClusterID, req.Namespace, req.ResName, req.Kind, "")
	if err != nil {
		return err
	}

	manifest := resInfo.UnstructuredContent()
	respDataBuilder, err := respUtil.NewRespDataBuilder(ctx, respUtil.DataBuilderParams{
		Manifest: manifest, Kind: req.Kind, Format: req.Format,
	})
	if err != nil {
		return err
	}
	result, err := respDataBuilder.Build()
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(result); err != nil {
		return err
	}

	return nil
}

// CreateApiResourcesObject create api resources object
func (h *Handler) CreateApiResourcesObject(ctx context.Context,
	req *clusterRes.CreateApiResourcesObjectReq, resp *clusterRes.CommonResp) (err error) {

	rawData := req.RawData.AsMap()
	kind := mapx.GetStr(rawData, "kind")
	apiVersion := mapx.GetStr(rawData, "apiVersion")

	transformer, err := trans.New(ctx, req.RawData.AsMap(), req.ClusterID, kind, constants.CreateAction, req.Format)
	if err != nil {
		return err
	}
	manifest, err := transformer.ToManifest()
	if err != nil {
		return err
	}
	if err = checkAccess(ctx, "", req.ClusterID, kind, manifest); err != nil {
		return err
	}
	resInfo, err := cli.CreateResObjectInfo(ctx, req.ClusterID, kind, apiVersion, req.Namespaced, manifest)
	if err != nil {
		return err
	}

	resp.Data = pbstruct.Unstructured2pbStruct(resInfo)
	return nil
}

// UpdateApiResourcesObject update api resources object
func (h *Handler) UpdateApiResourcesObject(ctx context.Context,
	req *clusterRes.UpdateApiResourcesObjectReq, resp *clusterRes.CommonResp) (err error) {

	rawData := req.RawData.AsMap()
	kind := mapx.GetStr(rawData, "kind")
	apiVersion := mapx.GetStr(rawData, "apiVersion")

	transformer, err := trans.New(ctx, req.RawData.AsMap(), req.ClusterID, kind, constants.CreateAction, req.Format)
	if err != nil {
		return err
	}
	manifest, err := transformer.ToManifest()
	if err != nil {
		return err
	}

	if err = checkAccess(ctx, "", req.ClusterID, kind, manifest); err != nil {
		return err
	}

	resInfo, err := cli.UpdateResObjectInfo(ctx, req.ClusterID, kind, apiVersion, manifest)
	if err != nil {
		return err
	}

	respDataBuilder, err := respUtil.NewRespDataBuilder(ctx, respUtil.DataBuilderParams{
		Manifest: resInfo.UnstructuredContent(), Kind: kind, Format: req.Format,
	})
	if err != nil {
		return err
	}
	result, err := respDataBuilder.Build()
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(result); err != nil {
		return err
	}

	return err
}

// DeleteApiResourcesObject delete api resources object
func (h *Handler) DeleteApiResourcesObject(ctx context.Context,
	req *clusterRes.DeleteApiResourcesObjectReq, resp *clusterRes.CommonResp) (err error) {

	if err = checkAccess(ctx, req.Namespace, req.ClusterID, req.Kind, nil); err != nil {
		return err
	}

	return cli.DeleteResObjectInfo(ctx, req.ClusterID, req.Namespace, req.ResName, req.Kind, "")
}

// checkAccess 访问权限检查（如共享集群禁用等）
func checkAccess(ctx context.Context, namespace, clusterID, kind string, manifest map[string]interface{}) error {
	clusterInfo, err := cluster.GetClusterInfo(ctx, clusterID)
	if err != nil {
		return err
	}
	// 独立集群中，不需要做类似校验
	if clusterInfo.Type == cluster.ClusterTypeSingle {
		return nil
	}
	// SC 允许用户查看，PV 返回空，不报错
	if slice.StringInSlice(kind, cluster.SharedClusterBypassClusterScopedKinds) {
		return nil
	}
	// 不允许的资源类型，直接抛出错误
	if !slice.StringInSlice(kind, cluster.SharedClusterEnabledNativeKinds) &&
		!slice.StringInSlice(kind, config.G.SharedCluster.EnabledCObjKinds) {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "该请求资源类型 %s 在共享集群中不可用"), kind)
	}
	// 对命名空间进行检查，确保是属于项目的，命名空间以 manifest 中的为准
	if manifest != nil {
		namespace = mapx.GetStr(manifest, "metadata.namespace")
	}
	if err = cli.CheckIsProjNSinSharedCluster(ctx, clusterID, namespace); err != nil {
		return err
	}
	return nil
}
