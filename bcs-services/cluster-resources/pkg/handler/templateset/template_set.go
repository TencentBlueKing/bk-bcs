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

// Package templateset 模板集类接口实现
package templateset

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/envmanage"
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

// GetEnvManage xxx
func (h *Handler) GetEnvManage(
	ctx context.Context, req *clusterRes.GetEnvManageReq, resp *clusterRes.CommonResp) error {
	action := envmanage.NewEnvManageAction(h.model)
	m, err := action.Get(ctx, req.Id, req.ProjectCode)
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(m); err != nil {
		return err
	}
	return nil
}

// ListEnvManages xxx
func (h *Handler) ListEnvManages(
	ctx context.Context, req *clusterRes.ListEnvManagesReq, resp *clusterRes.CommonListResp) error {
	action := envmanage.NewEnvManageAction(h.model)
	m, err := action.List(ctx)
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.MapSlice2ListValue(m); err != nil {
		return err
	}
	return nil
}

// CreateEnvManage xxx
func (h *Handler) CreateEnvManage(
	ctx context.Context, req *clusterRes.CreateEnvManageReq, resp *clusterRes.CommonResp) error {
	action := envmanage.NewEnvManageAction(h.model)
	id, err := action.Create(ctx, req)
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(map[string]interface{}{"id": id}); err != nil {
		return err
	}
	return nil
}

// UpdateEnvManage xxx
func (h *Handler) UpdateEnvManage(
	ctx context.Context, req *clusterRes.UpdateEnvManageReq, resp *clusterRes.CommonResp) error {
	action := envmanage.NewEnvManageAction(h.model)
	err := action.Update(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

// RenameEnvManage xxx
func (h *Handler) RenameEnvManage(
	ctx context.Context, req *clusterRes.RenameEnvManageReq, resp *clusterRes.CommonResp) error {
	action := envmanage.NewEnvManageAction(h.model)
	err := action.Rename(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

// DeleteEnvManage xxx
func (h *Handler) DeleteEnvManage(
	ctx context.Context, req *clusterRes.DeleteEnvManageReq, resp *clusterRes.CommonResp) error {
	action := envmanage.NewEnvManageAction(h.model)
	err := action.Delete(ctx, req.GetId())
	if err != nil {
		return err
	}
	return nil
}
