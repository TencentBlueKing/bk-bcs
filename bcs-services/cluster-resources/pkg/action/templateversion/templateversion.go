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
	"sort"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam"
	projectAuth "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser"
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

// GetContent xxx
func (t *TemplateVersionAction) GetContent(ctx context.Context, templateSpace, templateName, version string) (
	map[string]interface{}, error) {
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
		entity.FieldKeyTemplateSpace: templateSpace,
		entity.FieldKeyTemplateName:  templateName,
		entity.FieldKeyVersion:       version,
	})

	templateVersion, err := t.model.ListTemplateVersion(ctx, cond)
	if err != nil {
		return nil, err
	}

	if len(templateVersion) < 1 {
		return nil, nil
	}

	return templateVersion[0].ToMap(), nil
}

// List xxx
func (t *TemplateVersionAction) List(
	ctx context.Context, templateID string) ([]map[string]interface{}, error) {
	if err := t.checkAccess(ctx); err != nil {
		return nil, err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	tmp, err := t.model.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, err
	}
	if tmp.ProjectCode != p.Code {
		return nil, errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
	}

	// 检测条件
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode:   p.Code,
		entity.FieldKeyTemplateName:  tmp.Name,
		entity.FieldKeyTemplateSpace: tmp.TemplateSpace,
	})

	templateVersion, err := t.model.ListTemplateVersion(ctx, cond)
	if err != nil {
		return nil, err
	}
	sort.Sort(entity.VersionsSortByVersion(templateVersion))

	m := make([]map[string]interface{}, 0)
	// append draft version
	if tmp.IsDraft {
		draftVersion := &entity.TemplateVersion{
			ProjectCode:   tmp.ProjectCode,
			TemplateSpace: tmp.TemplateSpace,
			TemplateName:  tmp.Name,
			Version:       tmp.DraftVersion,
			Content:       tmp.DraftContent,
			Creator:       tmp.Updator,
			CreateAt:      tmp.UpdateAt,
			Draft:         true,
		}
		m = append(m, draftVersion.ToMap())
	}
	for _, value := range templateVersion {
		if value.Version == tmp.Version {
			value.Latest = true
		}
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

	tmp, err := t.model.GetTemplate(ctx, req.GetTemplateID())
	if err != nil {
		return "", err
	}

	// 检测是否重复
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode:   p.Code,
		entity.FieldKeyTemplateName:  tmp.Name,
		entity.FieldKeyTemplateSpace: tmp.TemplateSpace,
		entity.FieldKeyVersion:       req.GetVersion(),
	})
	templateVersions, err := t.model.ListTemplateVersion(ctx, cond)
	if err != nil {
		return "", err
	}

	userName := ctxkey.GetUsernameFromCtx(ctx)
	if len(templateVersions) > 0 {
		// 版本存在且不允许覆盖的情况下直接返回
		if !req.GetForce() {
			return "", errorx.New(errcode.DuplicationNameErr, i18n.GetMsg(ctx, "版本已存在"))
		}

		// 允许覆盖的情况下直接覆盖
		updateTemplateVersion := entity.M{
			"description": req.GetDescription(),
			"version":     req.GetVersion(),
			"content":     req.GetContent(),
			"creator":     userName,
		}
		if err = t.model.UpdateTemplateVersion(ctx, templateVersions[0].ID.Hex(), updateTemplateVersion); err != nil {
			return "", err
		}

		return templateVersions[0].ID.Hex(), nil
	}

	templateVersion := &entity.TemplateVersion{
		ProjectCode:   p.Code,
		Description:   req.GetDescription(),
		TemplateName:  tmp.Name,
		TemplateSpace: tmp.TemplateSpace,
		Version:       req.GetVersion(),
		EditFormat:    req.GetEditFormat(),
		Content:       req.GetContent(),
		Creator:       userName,
	}
	id, err := t.model.CreateTemplateVersion(ctx, templateVersion)
	if err != nil {
		return "", err
	}

	updateTemplate := make(entity.M, 0)
	// 如果草稿态的情况下，创建版本解除草稿态
	if tmp.IsDraft {
		updateTemplate["isDraft"] = false
		updateTemplate["draftVersion"] = ""
		updateTemplate["draftContent"] = ""
	}

	// update template lastet version
	if tmp.VersionMode == int(clusterRes.VersionMode_LatestUpdateTime) {
		updateTemplate["version"] = req.GetVersion()
		updateTemplate["resourceType"] = parser.GetResourceTypesFromManifest(req.GetContent())
		updateTemplate["updator"] = userName
	}

	if len(updateTemplate) != 0 {
		if err = t.model.UpdateTemplate(ctx, req.GetTemplateID(), updateTemplate); err != nil {
			return "", err
		}
	}
	return id, nil
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
