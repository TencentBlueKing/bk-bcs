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

// Package template 模板文件元数据
package template

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/feiin/go-xss"

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

// TemplateAction provides the action to manager template file
// nolint
type TemplateAction struct {
	ctx context.Context

	model store.ClusterResourcesModel
}

// NewTemplateAction return a new TemplateAction instance
func NewTemplateAction(model store.ClusterResourcesModel) *TemplateAction {
	return &TemplateAction{
		model: model,
	}
}

func (t *TemplateAction) checkAccess(ctx context.Context) error {
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
func (t *TemplateAction) Get(ctx context.Context, id string) (map[string]interface{}, error) {
	if err := t.checkAccess(ctx); err != nil {
		return nil, err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	template, err := t.model.GetTemplate(ctx, id)
	if err != nil {
		return nil, err
	}

	// 只能查看当前项目模板文件元数据
	if template.ProjectCode != p.Code {
		return nil, errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
	}

	return template.ToMap(), nil
}

// List xxx
func (t *TemplateAction) List(ctx context.Context, templateSpace string) ([]map[string]interface{}, error) {
	if err := t.checkAccess(ctx); err != nil {
		return nil, err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 通过项目编码、文件夹名称检索
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode:   p.Code,
		entity.FieldKeyTemplateSpace: templateSpace,
	})

	template, err := t.model.ListTemplate(ctx, cond)
	if err != nil {
		return nil, err
	}

	m := make([]map[string]interface{}, 0)
	for _, value := range template {
		m = append(m, value.ToMap())
	}
	return m, nil
}

// Create xxx
func (t *TemplateAction) Create(ctx context.Context, req *clusterRes.CreateTemplateMetadataReq) (string, error) {
	if err := t.checkAccess(ctx); err != nil {
		return "", err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return "", err
	}

	// 检测模板元数据是否重复
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode:   p.Code,
		entity.FieldKeyName:          req.GetName(),
		entity.FieldKeyTemplateSpace: req.GetTemplateSpace(),
	})
	templates, err := t.model.ListTemplate(ctx, cond)
	if err != nil {
		return "", err
	}

	if len(templates) > 0 {
		return "", errorx.New(errcode.DuplicationNameErr, i18n.GetMsg(ctx, "元数据名称重复"))
	}

	userName := ctxkey.GetUsernameFromCtx(ctx)

	// 创建顺序：templateVersion -> template
	templateVersion := &entity.TemplateVersion{
		ProjectCode:   p.Code,
		Description:   xss.FilterXSS(req.GetVersionDescription(), xss.XssOption{}),
		TemplateName:  req.Name,
		TemplateSpace: req.TemplateSpace,
		Version:       req.Version,
		Content:       req.Content,
		Creator:       userName,
	}

	_, err = t.model.CreateTemplateVersion(ctx, templateVersion)
	if err != nil {
		return "", err
	}

	template := &entity.Template{
		Name:          req.GetName(),
		ProjectCode:   p.Code,
		Description:   xss.FilterXSS(req.GetDescription(), xss.XssOption{}),
		TemplateSpace: req.GetTemplateSpace(),
		ResourceType:  req.GetResourceType(),
		Creator:       userName,
		Updator:       userName,
		Tags:          req.GetTags(),
		VersionMode:   0,
		Version:       req.GetVersion(),
	}
	templateId, err := t.model.CreateTemplate(ctx, template)
	if err != nil {
		return "", err
	}

	return templateId, nil
}

// Update xxx
func (t *TemplateAction) Update(ctx context.Context, req *clusterRes.UpdateTemplateMetadataReq) error {
	if err := t.checkAccess(ctx); err != nil {
		return err
	}

	// 校验版本模式
	if req.GetVersionMode() != 0 && req.GetVersionMode() != 1 {
		return errorx.New(errcode.ValidateErr, i18n.GetMsg(ctx, "版本模式校验失败"))
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return err
	}

	template, err := t.model.GetTemplate(ctx, req.GetId())
	if err != nil {
		return err
	}

	userName := ctxkey.GetUsernameFromCtx(ctx)
	// 检验更新 Template 的权限
	if template.ProjectCode != p.Code {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
	}

	// 检测是否重复
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyName:          req.GetName(),
		entity.FieldKeyProjectCode:   p.Code,
		entity.FieldKeyTemplateSpace: template.TemplateSpace,
	})
	templates, err := t.model.ListTemplate(ctx, cond)
	if err != nil {
		return err
	}

	// 存在同一个projectCode不同id相同的元数据名称则不能更新
	if len(templates) > 0 && templates[0].ID.Hex() != req.GetId() {
		return errorx.New(errcode.DuplicationNameErr, i18n.GetMsg(ctx, "元数据名称重复"))
	}

	// 如果元数据名称有更新需要先更新版本集合
	if template.Name != req.GetName() {
		updateTemplateVersion := entity.M{
			"templateName":  req.GetName(),
			"templateSpace": template.TemplateSpace,
		}
		if err = t.model.UpdateTemplateVersionBySpecial(
			ctx, p.Code, template.Name, template.TemplateSpace, updateTemplateVersion); err != nil {
			return err
		}
	}

	updateTemplate := entity.M{
		"name":         req.GetName(),
		"description":  xss.FilterXSS(req.GetDescription(), xss.XssOption{}),
		"resourceType": req.GetResourceType(),
		"updator":      userName,
		"tags":         req.GetTags(),
		"versionMode":  req.GetVersionMode(),
		"version":      req.GetVersion(),
	}
	if err = t.model.UpdateTemplate(ctx, req.GetId(), updateTemplate); err != nil {
		return err
	}

	return nil
}

// Delete xxx
func (t *TemplateAction) Delete(ctx context.Context, id string, isRelateDelete bool) error {
	if err := t.checkAccess(ctx); err != nil {
		return err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return err
	}

	template, err := t.model.GetTemplate(ctx, id)
	if err != nil {
		return err
	}

	// 检验删除 Template 的权限
	if template.ProjectCode != p.Code {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
	}

	// 是否需要把版本关联的数据也删掉
	if isRelateDelete {
		if err = t.model.DeleteTemplateVersionBySpecial(
			ctx, p.Code, template.Name, template.TemplateSpace); err != nil {
			return err
		}
	}

	if err = t.model.DeleteTemplate(ctx, id); err != nil {
		return err
	}

	return nil
}
