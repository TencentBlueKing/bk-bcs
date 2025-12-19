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
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/coreos/go-semver/semver"
	"github.com/feiin/go-xss"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam"
	projectAuth "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
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

// GetTemplateResources xxx
func (t *TemplateVersionAction) GetTemplateResources(
	ctx context.Context, in *clusterRes.GetTemplateResourcesReq) (map[string]interface{}, error) {

	if err := t.checkAccess(ctx); err != nil {
		return nil, err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	if in.TemplateSpace == "" {
		return parser.ParseResourcesFromManifest(nil, in.Kind), nil
	}

	templateSpace, err := t.model.GetTemplateSpace(ctx, in.TemplateSpace)
	if err != nil {
		return parser.ParseLablesFromManifest(nil, in.Kind, ""), nil
	}

	templates, err := t.model.ListTemplate(ctx, operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode:   p.Code,
		entity.FieldKeyTemplateSpace: templateSpace.Name,
	}))
	if err != nil {
		return nil, err
	}

	versions := make([]*entity.TemplateVersion, 0)
	for _, v := range templates {
		version, err := t.model.ListTemplateVersion(ctx, operator.NewLeafCondition(operator.Eq, operator.M{
			entity.FieldKeyProjectCode:   p.Code,
			entity.FieldKeyTemplateSpace: v.TemplateSpace,
			entity.FieldKeyTemplateName:  v.Name,
			entity.FieldKeyVersion:       v.Version,
		}))
		if err != nil {
			continue
		}
		if len(version) == 0 {
			continue
		}
		versions = append(versions, version[0])
	}

	return parser.ParseResourcesFromManifest(versions, in.Kind), nil
}

// GetTemplateAssociateLabels xxx
func (t *TemplateVersionAction) GetTemplateAssociateLabels(
	ctx context.Context, in *clusterRes.GetTemplateAssociateLabelsReq) (map[string]interface{}, error) {

	if err := t.checkAccess(ctx); err != nil {
		return nil, err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	//  associateName 格式为 templateName/workloadName
	names := strings.SplitN(in.AssociateName, "/", 2)
	if len(names) != 2 {
		return parser.ParseLablesFromManifest(nil, in.Kind, ""), nil
	}

	templateSpace, err := t.model.GetTemplateSpace(ctx, in.TemplateSpace)
	if err != nil {
		return parser.ParseLablesFromManifest(nil, in.Kind, ""), nil
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode:   p.Code,
		entity.FieldKeyTemplateSpace: templateSpace.Name,
		entity.FieldKeyTemplateName:  names[0],
	})

	templateVersions, err := t.model.ListTemplateVersion(ctx, cond)
	if err != nil {
		return nil, err
	}

	return parser.ParseLablesFromManifest(templateVersions, in.Kind, names[1]), nil
}

// GetTemplateAssociatePorts xxx
func (t *TemplateVersionAction) GetTemplateAssociatePorts(
	ctx context.Context, in *clusterRes.GetTemplateAssociatePortsReq) (map[string]interface{}, error) {

	if err := t.checkAccess(ctx); err != nil {
		return nil, err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	//  associateName 格式为 templateName/workloadName
	names := strings.SplitN(in.AssociateName, "/", 2)
	if len(names) != 2 {
		return parser.ParsePortsFromManifest(nil, in.Kind, "", ""), nil
	}

	templateSpace, err := t.model.GetTemplateSpace(ctx, in.TemplateSpace)
	if err != nil {
		return parser.ParsePortsFromManifest(nil, in.Kind, "", ""), nil
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode:   p.Code,
		entity.FieldKeyTemplateSpace: templateSpace.Name,
		entity.FieldKeyTemplateName:  names[0],
	})

	templateVersions, err := t.model.ListTemplateVersion(ctx, cond)
	if err != nil {
		return nil, err
	}

	return parser.ParsePortsFromManifest(templateVersions, in.Kind, names[1], in.Protocol), nil
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

	// 区分语义化版本及非语义话版本
	var semVersion []*entity.TemplateVersion
	var nonSemVersion []*entity.TemplateVersion
	for i, v := range templateVersion {
		_, err = semver.NewVersion(v.Version)
		if err != nil {
			nonSemVersion = append(nonSemVersion, templateVersion[i])
			continue
		}
		semVersion = append(semVersion, templateVersion[i])
	}
	sort.Sort(entity.VersionsSortByVersion(semVersion))
	// 非语义化版本按时间倒序排序
	sort.Sort(entity.VersionsSortByCreateAt(nonSemVersion))
	semVersion = append(semVersion, nonSemVersion...)
	templateVersion = semVersion

	m := make([]map[string]interface{}, 0)
	// append draft version
	if tmp.IsDraft {
		draftVersion := &entity.TemplateVersion{
			ProjectCode:   tmp.ProjectCode,
			TemplateSpace: tmp.TemplateSpace,
			TemplateName:  tmp.Name,
			Version:       tmp.DraftVersion,
			Content:       tmp.DraftContent,
			EditFormat:    tmp.DraftEditFormat,
			Creator:       tmp.Updator,
			CreateAt:      tmp.UpdateAt,
			Draft:         true,
			RenderMode:    tmp.RenderMode,
		}
		m = append(m, draftVersion.ToMap())
	}
	for _, value := range templateVersion {
		if value.Version == tmp.Version {
			value.Latest = true
		}
		value.LatestDeployVersion = tmp.LatestDeployVersion
		value.LatestDeployer = tmp.LatestDeployer
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
	renderMode := constants.RenderMode(req.GetRenderMode())
	if len(templateVersions) > 0 {
		// 版本存在且不允许覆盖的情况下直接返回
		if !req.GetForce() {
			return "", errorx.New(errcode.DuplicationNameErr, i18n.GetMsg(ctx, "版本已存在"))
		}

		// 允许覆盖的情况下直接覆盖
		updateTemplateVersion := entity.M{
			"description": xss.FilterXSS(req.GetDescription(), xss.XssOption{}),
			"version":     req.GetVersion(),
			"content":     req.GetContent(),
			"creator":     userName,
			"renderMode":  renderMode.GetRenderMode(),
		}
		if err = t.model.UpdateTemplateVersion(ctx, templateVersions[0].ID.Hex(), updateTemplateVersion); err != nil {
			return "", err
		}

		return templateVersions[0].ID.Hex(), nil
	}

	templateVersion := &entity.TemplateVersion{
		ProjectCode:   p.Code,
		Description:   xss.FilterXSS(req.GetDescription(), xss.XssOption{}),
		TemplateName:  tmp.Name,
		TemplateSpace: tmp.TemplateSpace,
		Version:       req.GetVersion(),
		EditFormat:    req.GetEditFormat(),
		Content:       req.GetContent(),
		Creator:       userName,
		RenderMode:    renderMode.GetRenderMode(),
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
