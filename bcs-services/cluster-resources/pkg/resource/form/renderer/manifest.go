/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package renderer

import (
	"bytes"
	"context"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/validator"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ManifestRenderer 渲染并加载资源配置模板
type ManifestRenderer struct {
	ctx        context.Context
	formData   map[string]interface{}
	clusterID  string
	kind       string
	apiVersion string
	tmpl       *template.Template
	manifest   map[string]interface{}
}

// NewManifestRenderer xxx
func NewManifestRenderer(
	ctx context.Context, formData map[string]interface{}, clusterID, kind string,
) *ManifestRenderer {
	return &ManifestRenderer{
		ctx: ctx, formData: formData, clusterID: clusterID, kind: kind, manifest: map[string]interface{}{},
	}
}

// Render 渲染表单数据，返回 Manifest
func (r *ManifestRenderer) Render() (map[string]interface{}, error) {
	for _, f := range []func() error{
		// 1. 获取资源对应版本
		r.setVersionAndKind,
		// 2. 校验表单数据
		r.validate,
		// 3. 添加 EditMode 注解标识
		r.setEditMode,
		// 4. 数据清洗，去除表单默认值等
		r.cleanFormData,
		// 5. 加载模板并初始化
		r.initTemplate,
		// 6. 渲染模板并转换格式
		r.render2Map,
	} {
		if err := f(); err != nil {
			return nil, err
		}
	}
	return r.manifest, nil
}

// setVersionAndKind 获取资源对应 APIVersion && Kind 并更新 Renderer 配置
func (r *ManifestRenderer) setVersionAndKind() (err error) {
	// 以 FormData 中的 ApiVersion 为准，若为空，则自动填充 preferred version
	r.apiVersion = mapx.GetStr(r.formData, "metadata.apiVersion")
	if r.apiVersion == "" {
		if r.apiVersion, err = res.GetResPreferredVersion(r.ctx, r.clusterID, r.kind); err != nil {
			return err
		}
		if err = mapx.SetItems(r.formData, "metadata.apiVersion", r.apiVersion); err != nil {
			return err
		}
	}
	// 预设资源 Kind
	if err = mapx.SetItems(r.formData, "metadata.kind", r.kind); err != nil {
		return err
	}
	return nil
}

// validate 校验表单数据
func (r *ManifestRenderer) validate() error {
	return validator.New(r.ctx, r.formData, r.apiVersion, r.kind).Validate()
}

// setEditMode 添加 EditMode Annotations 标识
func (r *ManifestRenderer) setEditMode() error {
	// 若 annotations 中有 editMode key，则刷新为 FormMode
	annotations := mapx.GetList(r.formData, "metadata.annotations")
	for _, anno := range annotations {
		if anno.(map[string]interface{})["key"] == res.EditModeAnnoKey {
			anno.(map[string]interface{})["value"] = res.EditModeForm
			return nil
		}
	}
	// 如果没有对应的 key，则新增
	annotations = append(annotations, map[string]interface{}{"key": res.EditModeAnnoKey, "value": res.EditModeForm})
	return mapx.SetItems(r.formData, "metadata.annotations", annotations)
}

// cleanFormData 清理表单数据，如去除默认值等
func (r *ManifestRenderer) cleanFormData() error {
	// 默认值清理规则：某子表单中均为初始的零值，则认为未被修改，不应作为配置下发
	if isEmptyMap := mapx.RemoveZeroSubItem(r.formData); isEmptyMap {
		return errorx.New(errcode.General, i18n.GetMsg(r.ctx, "数据清洗零值结果为空集合"))
	}
	return nil
}

// initTemplate 加载模板并初始化
func (r *ManifestRenderer) initTemplate() (err error) {
	r.tmpl, err = initTemplate(envs.FormTmplFileBaseDir+"/manifest/", "*")
	return err
}

// render2Map 渲染模板并转换成 Map 格式
func (r *ManifestRenderer) render2Map() error {
	// 渲染，转换并写入数据（模板名称格式：{r.kind}.yaml）
	var buf bytes.Buffer
	err := r.tmpl.ExecuteTemplate(&buf, r.kind+".yaml", r.formData)
	if err != nil {
		log.Warn(r.ctx, "failed to render template：%v", err)
		return errorx.New(errcode.General, i18n.GetMsg(r.ctx, "渲染模板失败：%v"), err)
	}
	return yaml.Unmarshal(buf.Bytes(), r.manifest)
}
