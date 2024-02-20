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

// Package templatespace 模板文件文件夹
package templatespace

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam"
	projectAuth "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// TemplateSpaceAction provides the action to manager template space
// nolint
type TemplateSpaceAction struct {
	ctx context.Context

	model store.ClusterResourcesModel
}

// NewTemplateSpaceAction return a new TemplateSpaceAction instance
func NewTemplateSpaceAction(model store.ClusterResourcesModel) *TemplateSpaceAction {
	return &TemplateSpaceAction{
		model: model,
	}
}

func (t *TemplateSpaceAction) checkAccess(ctx context.Context) error {
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
func (t *TemplateSpaceAction) Get(ctx context.Context, id string) (map[string]interface{}, error) {
	if err := t.checkAccess(ctx); err != nil {
		return nil, err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	templateSpace, err := t.model.GetTemplateSpace(ctx, id)
	if err != nil {
		return nil, err
	}

	// 只能查看当前项目视图，或者自己创建的视图，或者公共视图
	if templateSpace.ProjectCode != p.Code {
		return nil, errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
	}

	return templateSpace.ToMap(), nil
}

// List xxx
func (t *TemplateSpaceAction) List(ctx context.Context) ([]map[string]interface{}, error) {
	if err := t.checkAccess(ctx); err != nil {
		return nil, err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 通过项目编码检索
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode: p.Code,
	})

	templateSpace, err := t.model.ListTemplateSpace(ctx, cond)
	if err != nil {
		return nil, err
	}

	m := make([]map[string]interface{}, 0)
	for _, value := range templateSpace {
		m = append(m, value.ToMap())
	}
	return m, nil
}

// Create xxx
func (t *TemplateSpaceAction) Create(ctx context.Context, req *clusterRes.CreateTemplateSpaceReq) (string, error) {
	if err := t.checkAccess(ctx); err != nil {
		return "", err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return "", err
	}

	// 检测是否重复
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode: p.Code,
		entity.FieldKeyName:        req.Name,
	})
	templateSpaces, err := t.model.ListTemplateSpace(ctx, cond)
	if err != nil {
		return "", err
	}

	if len(templateSpaces) > 0 {
		return "", errorx.New(errcode.DuplicationNameErr, i18n.GetMsg(ctx, "文件夹名称重复"))
	}

	templateSpace := &entity.TemplateSpace{
		Name:        req.GetName(),
		ProjectCode: p.Code,
		Description: req.GetDescription(),
	}
	id, err := t.model.CreateTemplateSpace(ctx, templateSpace)
	if err != nil {
		return "", err
	}
	return id, nil
}

// Update xxx
func (t *TemplateSpaceAction) Update(ctx context.Context, req *clusterRes.UpdateTemplateSpaceReq) error {
	if err := t.checkAccess(ctx); err != nil {
		return err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return err
	}

	templateSpace, err := t.model.GetTemplateSpace(ctx, req.GetId())
	if err != nil {
		return err
	}

	// 检验更新 TemplateSpace 的权限
	if templateSpace.ProjectCode != p.Code {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
	}

	// 检测是否重复
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyName:        req.GetName(),
		entity.FieldKeyProjectCode: p.Code,
	})
	templateSpaces, err := t.model.ListTemplateSpace(ctx, cond)
	if err != nil {
		return err
	}

	// 存在同一个projectCode不同id相同的文件夹名称则不能更新
	if len(templateSpaces) > 0 && templateSpaces[0].ID.Hex() != req.GetId() {
		return errorx.New(errcode.DuplicationNameErr, i18n.GetMsg(ctx, "文件夹名称重复"))
	}

	// 文件夹名称变更的情况下，更新顺序 templateversion -> template -> templatespace
	if templateSpace.Name != req.GetName() {
		templateVersion := entity.M{
			"templateSpace": req.GetName(),
		}
		err = t.model.UpdateTemplateVersionBySpecial(
			ctx, templateSpace.ProjectCode, "", templateSpace.Name, templateVersion)
		if err != nil {
			return err
		}

		template := entity.M{
			"templateSpace": req.GetName(),
		}
		err = t.model.UpdateTemplateBySpecial(ctx, templateSpace.ProjectCode, templateSpace.Name, template)
		if err != nil {
			return err
		}

	}

	updateTemplateSpace := entity.M{
		"name":        req.GetName(),
		"description": req.GetDescription(),
	}
	if err = t.model.UpdateTemplateSpace(ctx, req.GetId(), updateTemplateSpace); err != nil {
		return err
	}
	return nil
}

// Delete xxx
func (t *TemplateSpaceAction) Delete(ctx context.Context, id string, isRelateDelete bool) error {
	if err := t.checkAccess(ctx); err != nil {
		return err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return err
	}

	templateSpace, err := t.model.GetTemplateSpace(ctx, id)
	if err != nil {
		return err
	}

	// 检验更新 TemplateSpace 的权限
	if templateSpace.ProjectCode != p.Code {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
	}

	// 是否需要把文件夹关联的数据也删掉， 顺序 templateversion -> template -> templatespace
	if isRelateDelete {
		if err = t.model.DeleteTemplateVersionBySpecial(
			ctx, templateSpace.ProjectCode, "", templateSpace.Name); err != nil {
			return err
		}

		if err = t.model.DeleteTemplateBySpecial(ctx, templateSpace.ProjectCode, templateSpace.Name); err != nil {
			return err
		}
	}

	if err := t.model.DeleteTemplateSpace(ctx, id); err != nil {
		return err
	}
	return nil
}
