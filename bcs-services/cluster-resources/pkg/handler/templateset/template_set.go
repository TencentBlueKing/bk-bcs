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
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/template"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/templatespace"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/templatespacecollect"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/templateversion"
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

// GetTemplateSpace 获取模板文件文件夹详情
func (h *Handler) GetTemplateSpace(
	ctx context.Context, in *clusterRes.GetTemplateSpaceReq, out *clusterRes.CommonResp) error {
	action := templatespace.NewTemplateSpaceAction(h.model)
	data, err := action.Get(ctx, in.GetId())
	if err != nil {
		return err
	}
	if out.Data, err = pbstruct.Map2pbStruct(data); err != nil {
		return err
	}
	return nil
}

// ListTemplateSpace 获取模板文件文件夹列表
func (h *Handler) ListTemplateSpace(
	ctx context.Context, in *clusterRes.ListTemplateSpaceReq, out *clusterRes.CommonListResp) error {
	action := templatespace.NewTemplateSpaceAction(h.model)
	data, err := action.List(ctx, in.Name)
	if err != nil {
		return err
	}
	if out.Data, err = pbstruct.MapSlice2ListValue(data); err != nil {
		return err
	}
	return nil
}

// CreateTemplateSpace 创建模板文件文件夹
func (h *Handler) CreateTemplateSpace(
	ctx context.Context, in *clusterRes.CreateTemplateSpaceReq, out *clusterRes.CommonResp) error {
	action := templatespace.NewTemplateSpaceAction(h.model)
	id, err := action.Create(ctx, in)
	if err != nil {
		return err
	}
	if out.Data, err = pbstruct.Map2pbStruct(map[string]interface{}{"id": id}); err != nil {
		return err
	}
	return nil
}

// UpdateTemplateSpace 更新模板文件文件夹
func (h *Handler) UpdateTemplateSpace(
	ctx context.Context, in *clusterRes.UpdateTemplateSpaceReq, out *clusterRes.CommonResp) error {
	action := templatespace.NewTemplateSpaceAction(h.model)
	err := action.Update(ctx, in)
	if err != nil {
		return err
	}
	return nil
}

// DeleteTemplateSpace 删除模板文件文件夹
func (h *Handler) DeleteTemplateSpace(
	ctx context.Context, in *clusterRes.DeleteTemplateSpaceReq, out *clusterRes.CommonResp) error {
	action := templatespace.NewTemplateSpaceAction(h.model)
	err := action.Delete(ctx, in.GetId())
	if err != nil {
		return err
	}
	return nil
}

// CopyTemplateSpace 复制模板文件文件夹
func (h *Handler) CopyTemplateSpace(
	ctx context.Context, in *clusterRes.CopyTemplateSpaceReq, out *clusterRes.CommonResp) error {
	action := templatespace.NewTemplateSpaceAction(h.model)
	id, err := action.Copy(ctx, in.GetId())
	if err != nil {
		return err
	}
	if out.Data, err = pbstruct.Map2pbStruct(map[string]interface{}{"id": id}); err != nil {
		return err
	}
	return nil
}

// CreateTemplateSpaceCollect 创建模板文件文件夹收藏
func (h *Handler) CreateTemplateSpaceCollect(
	ctx context.Context, in *clusterRes.CreateTemplateSpaceCollectReq, out *clusterRes.CommonResp) error {
	action := templatespacecollect.NewTemplateSpaceCollectAction(h.model)
	id, err := action.Create(ctx, in)
	if err != nil {
		return err
	}
	if out.Data, err = pbstruct.Map2pbStruct(map[string]interface{}{"id": id}); err != nil {
		return err
	}
	return nil
}

// DeleteTemplateSpaceCollect 删除模板文件文件夹收藏
func (h *Handler) DeleteTemplateSpaceCollect(
	ctx context.Context, in *clusterRes.DeleteTemplateSpaceCollectReq, out *clusterRes.CommonResp) error {
	action := templatespacecollect.NewTemplateSpaceCollectAction(h.model)
	err := action.Delete(ctx, in.GetTemplateSpaceID())
	if err != nil {
		return err
	}
	return nil
}

// GetTemplateMetadata 获取模板文件元数据详情
func (h *Handler) GetTemplateMetadata(
	ctx context.Context, in *clusterRes.GetTemplateMetadataReq, out *clusterRes.CommonResp) error {
	action := template.NewTemplateAction(h.model)
	data, err := action.Get(ctx, in.GetId())
	if err != nil {
		return err
	}
	if out.Data, err = pbstruct.Map2pbStruct(data); err != nil {
		return err
	}
	return nil
}

// ListTemplateMetadata 获取模板文件元数据列表
func (h *Handler) ListTemplateMetadata(
	ctx context.Context, in *clusterRes.ListTemplateMetadataReq, out *clusterRes.CommonListResp) error {
	action := template.NewTemplateAction(h.model)
	data, err := action.List(ctx, in.GetTemplateSpaceID())
	if err != nil {
		return err
	}
	if out.Data, err = pbstruct.MapSlice2ListValue(data); err != nil {
		return err
	}
	return nil
}

// CreateTemplateMetadata 创建模板文件元数据
func (h *Handler) CreateTemplateMetadata(
	ctx context.Context, in *clusterRes.CreateTemplateMetadataReq, out *clusterRes.CommonResp) error {
	action := template.NewTemplateAction(h.model)
	id, versionID, err := action.Create(ctx, in)
	if err != nil {
		return err
	}
	if out.Data, err = pbstruct.Map2pbStruct(map[string]interface{}{"id": id, "versionID": versionID}); err != nil {
		return err
	}
	return nil
}

// UpdateTemplateMetadata 更新模板文件元数据
func (h *Handler) UpdateTemplateMetadata(
	ctx context.Context, in *clusterRes.UpdateTemplateMetadataReq, out *clusterRes.CommonResp) error {
	action := template.NewTemplateAction(h.model)
	err := action.Update(ctx, in)
	if err != nil {
		return err
	}
	return nil
}

// DeleteTemplateMetadata 删除模板文件元数据
func (h *Handler) DeleteTemplateMetadata(
	ctx context.Context, in *clusterRes.DeleteTemplateMetadataReq, out *clusterRes.CommonResp) error {
	action := template.NewTemplateAction(h.model)
	err := action.Delete(ctx, in.GetId())
	if err != nil {
		return err
	}
	return nil
}

// GetTemplateVersion 获取模板文件版本详情
func (h *Handler) GetTemplateVersion(
	ctx context.Context, in *clusterRes.GetTemplateVersionReq, out *clusterRes.CommonResp) error {
	action := templateversion.NewTemplateVersionAction(h.model)
	data, err := action.Get(ctx, in.GetId())
	if err != nil {
		return err
	}
	if out.Data, err = pbstruct.Map2pbStruct(data); err != nil {
		return err
	}
	return nil
}

// GetTemplateContent 获取模板文件详情
func (h *Handler) GetTemplateContent(
	ctx context.Context, in *clusterRes.GetTemplateContentReq, out *clusterRes.CommonResp) error {
	action := templateversion.NewTemplateVersionAction(h.model)
	data, err := action.GetContent(ctx, in.GetTemplateSpace(), in.GetTemplateName(), in.GetVersion())
	if err != nil {
		return err
	}
	if out.Data, err = pbstruct.Map2pbStruct(data); err != nil {
		return err
	}
	return nil
}

// ListTemplateVersion 获取模板文件版本列表
func (h *Handler) ListTemplateVersion(
	ctx context.Context, in *clusterRes.ListTemplateVersionReq, out *clusterRes.CommonListResp) error {
	action := templateversion.NewTemplateVersionAction(h.model)
	data, err := action.List(ctx, in.GetTemplateID())
	if err != nil {
		return err
	}
	if out.Data, err = pbstruct.MapSlice2ListValue(data); err != nil {
		return err
	}
	return nil
}

// CreateTemplateVersion 创建模板文件版本
func (h *Handler) CreateTemplateVersion(
	ctx context.Context, in *clusterRes.CreateTemplateVersionReq, out *clusterRes.CommonResp) error {
	action := templateversion.NewTemplateVersionAction(h.model)
	id, err := action.Create(ctx, in)
	if err != nil {
		return err
	}
	if out.Data, err = pbstruct.Map2pbStruct(map[string]interface{}{"id": id}); err != nil {
		return err
	}
	return nil
}

// DeleteTemplateVersion 删除模板文件版本
func (h *Handler) DeleteTemplateVersion(
	ctx context.Context, in *clusterRes.DeleteTemplateVersionReq, out *clusterRes.CommonResp) error {
	action := templateversion.NewTemplateVersionAction(h.model)
	err := action.Delete(ctx, in.GetId())
	if err != nil {
		return err
	}
	return nil
}

// CreateTemplateSet xxx
func (h *Handler) CreateTemplateSet(
	ctx context.Context, req *clusterRes.CreateTemplateSetReq, resp *clusterRes.CommonResp) error {
	action := template.NewTemplateAction(h.model)
	err := action.CreateTemplateSet(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

// ListTemplateFileVariables xxx
func (h *Handler) ListTemplateFileVariables(
	ctx context.Context, req *clusterRes.ListTemplateFileVariablesReq, resp *clusterRes.CommonResp) error {
	action := template.NewTemplateAction(h.model)
	m, err := action.ListTemplateFileVariables(ctx, req)
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(m); err != nil {
		return err
	}
	return nil
}

// PreviewTemplateFile xxx
func (h *Handler) PreviewTemplateFile(
	ctx context.Context, req *clusterRes.DeployTemplateFileReq, resp *clusterRes.CommonResp) error {
	action := template.NewTemplateAction(h.model)
	m, err := action.PreviewTemplateFile(ctx, req)
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(m); err != nil {
		return err
	}
	return nil
}

// DeployTemplateFile xxx
func (h *Handler) DeployTemplateFile(
	ctx context.Context, req *clusterRes.DeployTemplateFileReq, resp *clusterRes.CommonResp) error {
	action := template.NewTemplateAction(h.model)
	m, err := action.DeployTemplateFile(ctx, req)
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(m); err != nil {
		return err
	}
	return nil
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
