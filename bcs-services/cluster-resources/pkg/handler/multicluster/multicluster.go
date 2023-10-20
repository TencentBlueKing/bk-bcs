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
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// Handler Multicluster handler
type Handler struct{}

// New Multicluster handler
func New() *Handler {
	return &Handler{}
}

// FetchMultiClusterResource Fetch multi cluster resource
func (h *Handler) FetchMultiClusterResource(ctx context.Context, req *clusterRes.FetchMultiClusterResourceReq,
	resp *clusterRes.CommonResp) (err error) {
	filter := QueryFilter{
		Creator:       req.GetCreator(),
		Name:          req.GetName(),
		LabelSelector: req.GetLabelSelector(),
		Limit:         int(req.GetLimit()),
		Offset:        int(req.GetOffset()),
	}
	// from api server
	var query = NewAPIServerQuery(req.GetClusterNamespaces(), filter)
	// 多集群且 bcs-storage 支持的资源则从 bcs-storage 查询
	if len(req.GetClusterNamespaces()) > 1 &&
		slice.StringInSlice(req.GetKind(), config.G.MultiCluster.EnabledQueryFromStorageKinds) {
		// from storage
		query = NewStorageQuery(req.GetClusterNamespaces(), filter)
	}

	data, err := query.Fetch(ctx, "", req.GetKind())
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
	filter := QueryFilter{
		Creator:       req.GetCreator(),
		Name:          req.GetName(),
		LabelSelector: req.GetLabelSelector(),
		Limit:         int(req.GetLimit()),
		Offset:        int(req.GetOffset()),
	}
	var query = NewAPIServerQuery(req.GetClusterNamespaces(), filter)

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

	data, err := query.Fetch(ctx, groupVersion, kind)
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

// MultiClusterResourceCount get mul
func (h *Handler) MultiClusterResourceCount(ctx context.Context,
	req *clusterRes.MultiClusterResourceCountReq,
	resp *clusterRes.CommonResp) (err error) {
	filter := QueryFilter{
		Creator:       req.GetCreator(),
		Name:          req.GetName(),
		LabelSelector: req.GetLabelSelector(),
		Limit:         1,
	}
	var query = NewStorageQuery(req.GetClusterNamespaces(), filter)

	errgroup := errgroup.Group{}
	mux := sync.Mutex{}
	data := map[string]interface{}{}
	for _, kind := range config.G.MultiCluster.EnabledQueryFromStorageKinds {
		kind := kind
		errgroup.Go(func() error {
			result, innerErr := query.Fetch(ctx, "", kind)
			mux.Lock()
			defer mux.Unlock()
			if err != nil {
				log.Error(ctx, "query storage failed, %v", innerErr)
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
