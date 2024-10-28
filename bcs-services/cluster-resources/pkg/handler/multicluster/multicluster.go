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

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/web"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/featureflag"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/entity"
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

// FetchMultiClusterApiResources Fetch multi cluster api resources
func (h *Handler) FetchMultiClusterApiResources(ctx context.Context, req *clusterRes.FetchMultiClusterApiResourcesReq,
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
		LabelSelector: req.GetLabelSelector(),
	}
	// 是否仅获取crd资源
	kind := ""
	if req.GetOnlyCrd() {
		kind = constants.CRD
	}
	clusterNS := filterClusteredNamespace(req.GetClusterNamespaces(), string(getScopedByKind(kind)))

	// from api server
	var query = NewAPIServerQuery(clusterNS, filter, viewQueryToQueryFilter(view.Filter))

	var data map[string]interface{}
	data, err = query.FetchApiResources(ctx, kind)
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

// FetchMultiClusterCustomResource Fetch multi cluster custom resource
func (h *Handler) FetchMultiClusterCustomResource(ctx context.Context,
	req *clusterRes.FetchMultiClusterCustomResourceReq,
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
