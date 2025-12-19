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
	"path"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/feiin/go-xss"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	crAction "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/helm"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam"
	projectAuth "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/project"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/perm"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
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
		Username:  ctxkey.GetUsernameFromCtx(ctx),
		ProjectID: project.ID,
		TenantID:  ctxkey.GetTenantIDFromCtx(ctx),
	}
	if allow, err := iam.NewProjectPerm().CanView(permCtx); err != nil {
		return err
	} else if !allow {
		return errorx.New(errcode.NoIAMPerm, i18n.GetMsg(ctx, "无项目查看权限"))
	}
	return nil
}

// checkClusterAccess 访问权限检查（如共享集群禁用等）
func (t *TemplateAction) checkClusterAccess(ctx context.Context, clusterID, namespace string,
	manifests []map[string]interface{}) error {
	clusterInfo, err := cluster.GetClusterInfo(ctx, clusterID)
	if err != nil {
		return err
	}
	// 独立集群中，不需要做类似校验
	if clusterInfo.Type == cluster.ClusterTypeSingle {
		return nil
	}
	for _, manifest := range manifests {
		kind := mapx.GetStr(manifest, "kind")
		// 不允许的资源类型，直接抛出错误
		if !slice.StringInSlice(kind, cluster.SharedClusterEnabledNativeKinds) &&
			!slice.StringInSlice(kind, config.G.SharedCluster.EnabledCObjKinds) {
			return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "该请求资源类型 %s 在共享集群中不可用"), kind)
		}
	}
	// 对命名空间进行检查，确保是属于项目
	if err = cli.CheckIsProjNSinSharedCluster(ctx, clusterID, namespace); err != nil {
		return err
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

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode: p.Code,
	})
	// 获取 templatespace id
	templateSpace, err := t.model.ListTemplateSpace(ctx, cond)
	if err != nil {
		return nil, err
	}
	// get version id
	versions, err := t.model.ListTemplateVersion(ctx, operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode:   p.Code,
		entity.FieldKeyTemplateSpace: template.TemplateSpace,
		entity.FieldKeyTemplateName:  template.Name,
		entity.FieldKeyVersion:       template.Version,
	}))
	if err != nil {
		return nil, err
	}

	// template to map
	result := template.ToMap()
	for _, v := range templateSpace {
		if v.Name == template.TemplateSpace {
			result["templateSpaceID"] = v.ID.Hex()
		}
	}
	for _, v := range versions {
		if v.Version == template.Version {
			result["versionID"] = v.ID.Hex()
		}
	}

	return result, nil
}

// List xxx
func (t *TemplateAction) List(ctx context.Context, templateSpaceID string) ([]map[string]interface{}, error) {
	if err := t.checkAccess(ctx); err != nil {
		return nil, err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	templateSpace, err := t.model.GetTemplateSpace(ctx, templateSpaceID)
	if err != nil {
		return nil, err
	}

	// 通过项目编码、文件夹名称检索
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode:   p.Code,
		entity.FieldKeyTemplateSpace: templateSpace.Name,
	})

	template, err := t.model.ListTemplate(ctx, cond)
	if err != nil {
		return nil, err
	}
	// get version id
	versions, err := t.model.ListTemplateVersion(ctx, operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode:   p.Code,
		entity.FieldKeyTemplateSpace: templateSpace.Name,
	}))
	if err != nil {
		return nil, err
	}

	m := make([]map[string]interface{}, 0)
	for _, value := range template {
		mm := value.ToMap()
		for _, v := range versions {
			if v.Version == value.Version && v.TemplateName == value.Name {
				mm["versionID"] = v.ID.Hex()
			}
		}
		m = append(m, mm)
	}
	return m, nil
}

// Create xxx
func (t *TemplateAction) Create(ctx context.Context, req *clusterRes.CreateTemplateMetadataReq) (
	string, string, error) {
	if err := t.checkAccess(ctx); err != nil {
		return "", "", err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return "", "", err
	}

	// 非草稿模板文件需要版本号
	if !req.GetIsDraft() && req.GetVersion() == "" {
		return "", "", errorx.New(errcode.ValidateErr, i18n.GetMsg(ctx, ("版本字段不能为空")))
	}

	templateSpace, err := t.model.GetTemplateSpace(ctx, req.GetTemplateSpaceID())
	if err != nil {
		return "", "", err
	}

	// 检测模板元数据是否重复
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode:   p.Code,
		entity.FieldKeyName:          req.GetName(),
		entity.FieldKeyTemplateSpace: templateSpace.Name,
	})
	templates, err := t.model.ListTemplate(ctx, cond)
	if err != nil {
		return "", "", err
	}

	if len(templates) > 0 {
		return "", "", errorx.New(errcode.DuplicationNameErr, i18n.GetMsg(ctx, "元数据名称重复"))
	}

	userName := ctxkey.GetUsernameFromCtx(ctx)

	renderMode := constants.RenderMode(req.GetRenderMode()).GetRenderMode()
	// 非草稿状态下：创建模板文件版本
	versionID := ""
	if !req.GetIsDraft() {
		// 创建顺序：templateVersion -> template
		templateVersion := &entity.TemplateVersion{
			ProjectCode:   p.Code,
			Description:   req.GetVersionDescription(),
			TemplateName:  req.GetName(),
			TemplateSpace: templateSpace.Name,
			Version:       req.GetVersion(),
			EditFormat:    req.GetEditFormat(),
			Content:       req.GetContent(),
			Creator:       userName,
			RenderMode:    renderMode,
		}
		versionID, err = t.model.CreateTemplateVersion(ctx, templateVersion)
		if err != nil {
			return "", "", err
		}
	}

	template := &entity.Template{
		Name:          req.GetName(),
		ProjectCode:   p.Code,
		Description:   xss.FilterXSS(req.GetDescription(), xss.XssOption{}),
		TemplateSpace: templateSpace.Name,
		ResourceType:  parser.GetResourceTypesFromManifest(req.GetContent()),
		Creator:       userName,
		Updator:       userName,
		Tags:          req.GetTags(),
		VersionMode:   0,
		Version:       req.GetVersion(),
		IsDraft:       req.GetIsDraft(),
		RenderMode:    renderMode,
	}
	// 草稿状态，新增相关字段
	if req.GetIsDraft() {
		template.DraftVersion = req.GetDraftVersion()
		template.DraftContent = req.GetDraftContent()
		template.DraftEditFormat = req.GetDraftEditFormat()
	}

	// 没有记录的情况下直接创建
	templateID, err := t.model.CreateTemplate(ctx, template)
	if err != nil {
		return "", "", err
	}
	return templateID, versionID, nil
}

// Update xxx
func (t *TemplateAction) Update(ctx context.Context, req *clusterRes.UpdateTemplateMetadataReq) error {
	if err := t.checkAccess(ctx); err != nil {
		return err
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

	renderMode := constants.RenderMode(req.GetRenderMode()).GetRenderMode()
	updateTemplate := entity.M{
		"name":            req.GetName(),
		"description":     xss.FilterXSS(req.GetDescription(), xss.XssOption{}),
		"updator":         userName,
		"tags":            req.GetTags(),
		"versionMode":     req.GetVersionMode(),
		"isDraft":         req.GetIsDraft(),
		"draftVersion":    req.GetDraftVersion(),
		"draftContent":    req.GetDraftContent(),
		"draftEditFormat": req.GetDraftEditFormat(),
		"renderMode":      renderMode,
	}
	if req.GetVersionMode() == clusterRes.VersionMode_SpecifyVersion && req.GetVersion() != "" {
		updateTemplate["version"] = req.GetVersion()
	}

	if err = t.model.UpdateTemplate(ctx, req.GetId(), updateTemplate); err != nil {
		return err
	}

	return nil
}

// Delete xxx
func (t *TemplateAction) Delete(ctx context.Context, id string) error {
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

	// 把版本关联的数据也删掉
	if err = t.model.DeleteTemplateVersionBySpecial(
		ctx, p.Code, template.Name, template.TemplateSpace); err != nil {
		return err
	}

	if err = t.model.DeleteTemplate(ctx, id); err != nil {
		return err
	}

	return nil
}

// CreateTemplateSet create template set
func (t *TemplateAction) CreateTemplateSet(ctx context.Context, req *clusterRes.CreateTemplateSetReq) error {
	if err := t.checkAccess(ctx); err != nil {
		return err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return err
	}

	// 获取 templates
	tmps := t.model.ListTemplateVersionFromTemplateIDs(ctx, p.Code, toEntityTemplateIDs(req.GetTemplates()))

	// 组装 chart
	cht := buildChart(tmps, req, ctxkey.GetUsernameFromCtx(ctx))

	// 校验chart helm语法是否正确
	_, err = validAndFillChart(cht, req.GetValues())
	if err != nil {
		return err
	}

	// 上传 chart
	err = helm.UploadChart(ctx, cht, p.Code, req.GetVersion(), req.GetForce())
	if err != nil {
		return err
	}
	return nil
}

func getTemplateContents(ctx context.Context, model store.ClusterResourcesModel, versions []string,
	projectCode string) ([]entity.TemplateDeploy, error) {
	templates := make([]entity.TemplateDeploy, 0)
	for _, v := range versions {
		vv, err := model.GetTemplateVersion(ctx, v)
		if err != nil {
			return nil, err
		}
		if vv.ProjectCode != projectCode {
			return nil, errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
		}
		templates = append(templates, entity.TemplateDeploy{
			TemplateSpace:   vv.TemplateSpace,
			TemplateName:    vv.TemplateName,
			TemplateVersion: vv.Version,
			Content:         vv.Content,
			RenderMode:      vv.RenderMode,
		})
	}
	return templates, nil
}

// ListTemplateFileVariables list template file variables
func (t *TemplateAction) ListTemplateFileVariables(ctx context.Context,
	req *clusterRes.ListTemplateFileVariablesReq) (map[string]interface{}, error) {
	if err := t.checkAccess(ctx); err != nil {
		return nil, err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 获取 templates
	templates, err := getTemplateContents(ctx, t.model, req.GetTemplateVersions(), p.Code)
	if err != nil {
		return nil, err
	}

	// get namespace variables
	clusterVars, err := project.GetVariable(ctx, p.Code, req.GetClusterID(), req.GetNamespace())
	if err != nil {
		return nil, err
	}
	clusterVars = append(clusterVars, project.VariableValue{
		Key: "SYS_CLUSTER_ID", Value: req.GetClusterID(),
	})
	clusterVars = append(clusterVars, project.VariableValue{
		Key: "SYS_CC_APP_ID", Value: p.BusinessID,
	})
	clusterVars = append(clusterVars, project.VariableValue{
		Key: "SYS_PROJECT_ID", Value: p.ID,
	})
	clusterVars = append(clusterVars, project.VariableValue{
		Key: "SYS_NAMESPACE", Value: req.GetNamespace(),
	})
	vars := make([]map[string]interface{}, 0)
	for _, v := range parseMultiTemplateFileVar(templates) {
		value := ""
		for _, vv := range clusterVars {
			if vv.Key == v {
				value = vv.Value
			}
		}
		vars = append(vars, map[string]interface{}{
			"key":      v,
			"value":    value,
			"readonly": false, // 默认可覆盖
		})
	}
	return map[string]interface{}{"vars": vars}, nil
}

// PreviewTemplateFile preview template file
func (t *TemplateAction) PreviewTemplateFile(ctx context.Context, req *clusterRes.DeployTemplateFileReq) (
	map[string]interface{}, error) {
	if err := t.checkAccess(ctx); err != nil {
		return nil, err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 获取 templates
	templates, err := getTemplateContents(ctx, t.model, req.GetTemplateVersions(), p.Code)
	if err != nil {
		return nil, err
	}

	// helm 语法模式 模板文件内容进行helm template 渲染, 简单语法模式自动跳过
	content, errRender := renderTemplateForHelmMode(templates, req.GetValues(), req.GetVariables())
	if errRender != nil {
		return map[string]interface{}{"items": []string{}, "error": errRender.Error()}, nil
	}
	for k, v := range templates {
		helmPath := path.Join(v.TemplateSpace, v.TemplateName)
		if _, ok := content[helmPath]; ok {
			templates[k].Content = content[helmPath]
		}
	}

	// render templates
	manifests, err := t.renderTemplates(ctx, templates, req.GetVariables(), req.GetNamespace())
	if err != nil {
		return map[string]interface{}{"items": []string{}, "error": err.Error()}, nil
	}

	// 鉴权
	if errr := t.checkClusterAccess(ctx, req.GetClusterID(), req.GetNamespace(), manifests); errr != nil {
		return nil, errr
	}

	// dry-run deploy templates
	dryRunMsg := ""
	clusterConf := res.NewClusterConf(req.GetClusterID())
	for _, v := range manifests {
		kind := mapx.GetStr(v, "kind")
		groupVersion := mapx.GetStr(v, "apiVersion")
		if kind == "" {
			continue
		}
		k8sRes, errr := res.GetGroupVersionResource(ctx, clusterConf, kind, groupVersion)
		if errr != nil {
			dryRunMsg = errr.Error()
			break
		}
		_, errr = cli.NewResClient(clusterConf, k8sRes).ApplyWithoutPerm(ctx, v,
			metav1.CreateOptions{DryRun: []string{metav1.DryRunAll}})
		if errr != nil {
			dryRunMsg = errr.Error()
			break
		}
	}

	items, err := convertManifestToString(ctx, manifests, req.GetClusterID())
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"items": items, "error": dryRunMsg}, nil
}

// DeployTemplateFile deploy template file
func (t *TemplateAction) DeployTemplateFile(ctx context.Context, req *clusterRes.DeployTemplateFileReq) (
	map[string]interface{}, error) {
	if err := t.checkAccess(ctx); err != nil {
		return nil, err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 获取 templates
	templates, err := getTemplateContents(ctx, t.model, req.GetTemplateVersions(), p.Code)
	if err != nil {
		return nil, err
	}

	// helm 语法模式 模板文件内容进行helm template 渲染, 简单语法模式自动跳过
	content, errRender := renderTemplateForHelmMode(templates, req.GetValues(), req.GetVariables())
	if errRender != nil {
		return map[string]interface{}{"items": []string{}, "error": errRender.Error()}, nil
	}
	for k, v := range templates {
		helmPath := path.Join(v.TemplateSpace, v.TemplateName)
		if _, ok := content[helmPath]; ok {
			templates[k].Content = content[helmPath]
		}
	}

	// render templates
	manifests, err := t.renderTemplates(ctx, templates, req.GetVariables(), req.GetNamespace())
	if err != nil {
		return nil, err
	}

	// 鉴权
	if errr := t.checkClusterAccess(ctx, req.GetClusterID(), req.GetNamespace(), manifests); errr != nil {
		return nil, errr
	}
	if errr := perm.Validate(ctx, "", crAction.Create, p.ID, req.GetClusterID(), req.GetNamespace()); errr != nil {
		return nil, errr
	}

	// deploy templates
	clusterConf := res.NewClusterConf(req.GetClusterID())
	for _, v := range manifests {
		kind := mapx.GetStr(v, "kind")
		groupVersion := mapx.GetStr(v, "apiVersion")
		if kind == "" {
			continue
		}
		k8sRes, err := res.GetGroupVersionResource(ctx, clusterConf, kind, groupVersion)
		if err != nil {
			return nil, err
		}
		_, err = cli.NewResClient(clusterConf, k8sRes).ApplyWithoutPerm(ctx, v, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	}

	// 更新最新部署版本及最新部署人
	for _, v := range templates {
		err := t.model.UpdateTemplateBySpaceAndName(ctx, p.Code, v.TemplateSpace, v.TemplateName, entity.M{
			"latestDeployVersion": v.TemplateVersion,
			"latestDeployer":      ctxkey.GetUsernameFromCtx(ctx),
		})
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// ConvertTemplateToHelm 普通模板文件转换为 Helm 格式模板文件
func (t *TemplateAction) ConvertTemplateToHelm(
	ctx context.Context, req *clusterRes.ConvertTemplateToHelmReq) (map[string]interface{}, error) {
	if err := t.checkAccess(ctx); err != nil {
		return nil, err
	}

	// 变量如果开头是 .Values. 的则不动，不是的就加上 .Values.
	content := replaceTemplateFileToHelm(req.Content)

	return map[string]interface{}{
		"content": content,
	}, nil
}

// renderTemplates render templates
func (t *TemplateAction) renderTemplates(ctx context.Context, templates []entity.TemplateDeploy,
	vars map[string]string, ns string) ([]map[string]interface{}, error) {
	manifests := make([]map[string]interface{}, 0)
	for i := range templates {
		// helm模式的已经在renderTemplateForHelmMode转过了
		if templates[i].RenderMode != string(constants.HelmRenderMode) {
			templates[i].Content = replaceTemplateFileVar(templates[i].Content, vars)
		}
		mm := parser.SplitManifests(templates[i].Content)
		for _, v := range mm {
			manifest := map[string]interface{}{}
			if errr := yaml.Unmarshal([]byte(v), &manifest); errr != nil {
				return nil, errr
			}
			manifest = mapx.CleanUpMap(manifest)
			manifest = patchTemplateAnnotations(
				manifest, ctxkey.GetUsernameFromCtx(ctx),
				templates[i].TemplateSpace, templates[i].TemplateName, templates[i].TemplateVersion)
			// patch ns
			kind := mapx.GetStr(manifest, "kind")
			if mapx.GetStr(manifest, "metadata.namespace") != "" || isNSRequired(kind) {
				_ = mapx.SetItems(manifest, []string{"metadata", "namespace"}, ns)
			}
			manifests = append(manifests, manifest)
		}
	}
	return manifests, nil
}
