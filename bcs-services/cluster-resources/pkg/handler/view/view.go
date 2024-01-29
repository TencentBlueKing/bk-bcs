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

// Package view 视图类接口实现
package view

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/view"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// Handler xxx
type Handler struct {
	model store.ClusterResourcesModel
}

// New xxx
func New(model store.ClusterResourcesModel) *Handler {
	return &Handler{model: model}
}

// ListViewConfigs xxx
func (h *Handler) ListViewConfigs(
	ctx context.Context, req *clusterRes.ListViewConfigsReq, resp *clusterRes.CommonListResp,
) error {
	action := view.NewViewAction(h.model)
	m, err := action.List(ctx)
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.MapSlice2ListValue(m); err != nil {
		return err
	}
	return nil
}

// GetViewConfig xxx
func (h *Handler) GetViewConfig(
	ctx context.Context, req *clusterRes.GetViewConfigReq, resp *clusterRes.CommonResp,
) error {
	action := view.NewViewAction(h.model)
	m, err := action.Get(ctx, req.Id, req.ProjectCode)
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(m); err != nil {
		return err
	}
	return nil
}

// CreateViewConfig xxx
func (h *Handler) CreateViewConfig(
	ctx context.Context, req *clusterRes.CreateViewConfigReq, resp *clusterRes.CommonResp,
) error {
	action := view.NewViewAction(h.model)
	id, err := action.Create(ctx, req)
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(map[string]interface{}{"id": id}); err != nil {
		return err
	}
	return nil
}

// UpdateViewConfig xxx
func (h *Handler) UpdateViewConfig(
	ctx context.Context, req *clusterRes.UpdateViewConfigReq, resp *clusterRes.CommonResp,
) error {
	action := view.NewViewAction(h.model)
	err := action.Update(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

// RenameViewConfig xxx
func (h *Handler) RenameViewConfig(
	ctx context.Context, req *clusterRes.RenameViewConfigReq, resp *clusterRes.CommonResp,
) error {
	action := view.NewViewAction(h.model)
	err := action.Rename(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

// DeleteViewConfig xxx
func (h *Handler) DeleteViewConfig(
	ctx context.Context, req *clusterRes.DeleteViewConfigReq, resp *clusterRes.CommonResp,
) error {
	action := view.NewViewAction(h.model)
	err := action.Delete(ctx, req.GetId())
	if err != nil {
		return err
	}
	return nil
}

// ResourceNameSuggest xxx
func (h *Handler) ResourceNameSuggest(
	ctx context.Context, req *clusterRes.ViewSuggestReq, resp *clusterRes.CommonResp,
) error {
	action := view.NewViewAction(h.model)
	m, err := action.ResourceNameSuggest(ctx, req.GetClusterNamespaces())
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(map[string]interface{}{"values": m}); err != nil {
		return err
	}
	return nil
}

// LabelSuggest xxx
func (h *Handler) LabelSuggest(
	ctx context.Context, req *clusterRes.ViewSuggestReq, resp *clusterRes.CommonResp,
) error {
	action := view.NewViewAction(h.model)
	m, err := action.LabelSuggest(ctx, req.GetClusterNamespaces())
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(map[string]interface{}{"values": m}); err != nil {
		return err
	}
	return nil
}

// ValuesSuggest xxx
func (h *Handler) ValuesSuggest(
	ctx context.Context, req *clusterRes.ViewSuggestReq, resp *clusterRes.CommonResp,
) error {
	action := view.NewViewAction(h.model)
	m, err := action.ValuesSuggest(ctx, req.GetClusterNamespaces(), req.Label)
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(map[string]interface{}{"values": m}); err != nil {
		return err
	}
	return nil
}
