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

// Package templateversion 模板文件版本
package templateversion

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam"
	projectAuth "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// TemplateVersionAction provides the action to manager template file
// nolint
type TemplateVersionAction struct {
	ctx context.Context

	model store.ClusterResourcesModel
}

// NewTemplateVersionAction return a new TemplateVersionAction instance
func NewTemplateVersionAction(model store.ClusterResourcesModel) *TemplateVersionAction {
	return &TemplateVersionAction{
		model: model,
	}
}

func (t *TemplateVersionAction) checkAccess(ctx context.Context) error {
	if config.G.Auth.Disabled {
		return nil
	}
	project, err := project.FromContext(ctx)
	if err != nil {
		return err
	}
	// 权限控制为项目查看
	permCtx := &projectAuth.PermCtx{
		Username:  ctx.Value(ctxkey.UsernameKey).(string),
		ProjectID: project.ID,
	}
	if allow, err := iam.NewProjectPerm().CanView(permCtx); err != nil {
		return err
	} else if !allow {
		return errorx.New(errcode.NoIAMPerm, i18n.GetMsg(ctx, "无项目查看权限"))
	}
	return nil
}

// Get xxx
func (t *TemplateVersionAction) Get(ctx context.Context, id string) (map[string]interface{}, error) {
	if err := t.checkAccess(ctx); err != nil {
		return nil, err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	templateVersion, err := t.model.GetTemplateVersion(ctx, id)
	if err != nil {
		return nil, err
	}

	// 只能查看当前项目的版本
	if templateVersion.ProjectCode != p.Code {
		return nil, errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
	}

	return templateVersion.ToMap(), nil
}

// List xxx
func (t *TemplateVersionAction) List(
	ctx context.Context, templateName, templateSpace string) ([]map[string]interface{}, error) {
	if err := t.checkAccess(ctx); err != nil {
		return nil, err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 检测条件
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode:   p.Code,
		entity.FieldKeyTemplateName:  templateName,
		entity.FieldKeyTemplateSpace: templateSpace,
	})

	templateVersion, err := t.model.ListTemplateVersion(ctx, cond)
	if err != nil {
		return nil, err
	}

	m := make([]map[string]interface{}, 0)
	for _, value := range templateVersion {
		m = append(m, value.ToMap())
	}
	return m, nil
}

// Create xxx
func (t *TemplateVersionAction) Create(ctx context.Context, req *clusterRes.CreateTemplateVersionReq) (string, error) {
	if err := t.checkAccess(ctx); err != nil {
		return "", err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return "", err
	}

	// 检测是否重复
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode:   p.Code,
		entity.FieldKeyTemplateName:  req.GetTemplateName(),
		entity.FieldKeyTemplateSpace: req.GetTemplateSpace(),
		entity.FieldKeyVersion:       req.GetVersion(),
	})
	templateVersions, err := t.model.ListTemplateVersion(ctx, cond)
	if err != nil {
		return "", err
	}

	if len(templateVersions) > 0 {
		return "", errorx.New(errcode.DuplicationNameErr, i18n.GetMsg(ctx, "版本号重复"))
	}

	templateVersion := &entity.TemplateVersion{
		ProjectCode:   p.Code,
		Description:   req.GetDescription(),
		TemplateName:  req.GetTemplateName(),
		TemplateSpace: req.GetTemplateSpace(),
		Version:       req.GetVersion(),
		Content:       req.GetContent(),
		Creator:       req.GetCreator(),
	}
	id, err := t.model.CreateTemplateVersion(ctx, templateVersion)
	if err != nil {
		return "", err
	}
	return id, nil
}

// Update xxx
func (t *TemplateVersionAction) Update(ctx context.Context, req *clusterRes.UpdateTemplateVersionReq) error {
	if err := t.checkAccess(ctx); err != nil {
		return err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return err
	}

	templateVersion, err := t.model.GetTemplateVersion(ctx, req.GetId())
	if err != nil {
		return err
	}

	// 检验更新 TemplateVersion 的权限
	if templateVersion.ProjectCode != p.Code {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
	}

	// 检测版本是否重复
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyTemplateName:  templateVersion.TemplateName,
		entity.FieldKeyTemplateSpace: templateVersion.TemplateSpace,
		entity.FieldKeyProjectCode:   p.Code,
		entity.FieldKeyVersion:       req.GetVersion(),
	})
	templateVersions, err := t.model.ListTemplateVersion(ctx, cond)
	if err != nil {
		return err
	}

	// 存在同一个projectCode、不同id、相同的版本号则不能更新
	if len(templateVersions) > 0 && templateVersions[0].ID.Hex() != req.GetId() {
		return errorx.New(errcode.DuplicationNameErr, i18n.GetMsg(ctx, "版本号重复"))
	}

	updateTemplateVersion := entity.M{
		"description": req.GetDescription(),
		"version":     req.GetVersion(),
		"content":     req.GetContent(),
		"creator":     req.GetCreator(),
	}
	if err := t.model.UpdateTemplateVersion(ctx, req.GetId(), updateTemplateVersion); err != nil {
		return err
	}
	return nil
}

// Delete xxx
func (t *TemplateVersionAction) Delete(ctx context.Context, id string) error {
	if err := t.checkAccess(ctx); err != nil {
		return err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return err
	}

	templateVersion, err := t.model.GetTemplateVersion(ctx, id)
	if err != nil {
		return err
	}

	// 检验更新 TemplateVersion 的权限
	if templateVersion.ProjectCode != p.Code {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
	}

	if err := t.model.DeleteTemplateVersion(ctx, id); err != nil {
		return err
	}
	return nil
}
