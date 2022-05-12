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
	"fmt"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

// ManifestRenderer 渲染并加载资源配置模板
type ManifestRenderer struct {
	ctx        context.Context
	FormData   map[string]interface{}
	ClusterID  string
	Kind       string
	APIVersion string
	Tmpl       *template.Template
	Manifest   map[string]interface{}
}

// NewManifestRenderer ...
func NewManifestRenderer(ctx context.Context, formData map[string]interface{}, clusterID, kind string) *ManifestRenderer {
	return &ManifestRenderer{ctx: ctx, FormData: formData, ClusterID: clusterID, Kind: kind, Manifest: map[string]interface{}{}}
}

// Render 渲染表单数据，返回 Manifest
func (r *ManifestRenderer) Render() (map[string]interface{}, error) {
	for _, f := range []func() error{
		// 1. 自动获取资源对应版本
		r.setAPIVersion,
		// 2. 检查是否支持指定版本表单化
		r.checkRenderable,
		// 3. 数据清洗，去除表单默认值等
		r.cleanFormData,
		// 4. 加载模板并初始化
		r.initTemplate,
		// 5. 渲染模板并转换格式
		r.render2Map,
		// 6. 添加 EditMode Label 标识
		r.setEditMode,
	} {
		if err := f(); err != nil {
			return nil, err
		}
	}
	return r.Manifest, nil
}

// 获取资源对应 APIVersion 并更新 Renderer 配置
func (r *ManifestRenderer) setAPIVersion() error {
	// 以 FormData 中的 ApiVersion 为准，若为空，则自动填充 preferred version
	r.APIVersion = mapx.GetStr(r.FormData, "metadata.apiVersion")
	if r.APIVersion == "" {
		resInfo, err := res.GetGroupVersionResource(r.ctx, res.NewClusterConfig(r.ClusterID), r.Kind, "")
		if err != nil {
			return errorx.New(errcode.General, "获取资源 APIVersion 信息失败：%v", err)
		}
		r.APIVersion = resInfo.Version
		if resInfo.Group != "" {
			r.APIVersion = resInfo.Group + "/" + resInfo.Version
		}
		if err = mapx.SetItems(r.FormData, "metadata.apiVersion", r.APIVersion); err != nil {
			return err
		}
	}
	return nil
}

// 检查指定资源能否渲染为表单
func (r *ManifestRenderer) checkRenderable() error {
	supportedAPIVersions, ok := FormRenderSupportedResAPIVersion[r.Kind]
	if !ok {
		return errorx.New(errcode.Unsupported, "资源类型 %s 不支持表单化", r.Kind)
	}
	if !slice.StringInSlice(r.APIVersion, supportedAPIVersions) {
		return errorx.New(
			errcode.Unsupported,
			"资源类型 %s APIVersion %s 不在受支持的版本列表 %v 中，请改用 Yaml 模式而非表单化",
			r.Kind, r.APIVersion, supportedAPIVersions,
		)
	}
	return nil
}

// 清理表单数据，如去除默认值等
func (r *ManifestRenderer) cleanFormData() error {
	// 默认值清理规则：某子表单中均为初始的零值，则认为未被修改，不应作为配置下发
	if isEmptyMap := mapx.RemoveZeroSubItem(r.FormData); isEmptyMap {
		return errorx.New(errcode.General, "数据清洗零值结果为空集合")
	}
	return nil
}

// 加载模板并初始化
func (r *ManifestRenderer) initTemplate() (err error) {
	r.Tmpl, err = initTemplate(envs.FormTmplFileBaseDir+"/manifest/", "*")
	return err
}

// 渲染模板并转换成 Map 格式
func (r *ManifestRenderer) render2Map() error {
	// 渲染，转换并写入数据（模板名称格式：{r.Kind}.yaml）
	var buf bytes.Buffer
	err := r.Tmpl.ExecuteTemplate(&buf, r.Kind+".yaml", r.FormData)
	if err != nil {
		log.Warn(r.ctx, "渲染模板失败：%v", err)
		return errorx.New(errcode.General, "渲染模板失败：%v", err)
	}
	return yaml.Unmarshal(buf.Bytes(), r.Manifest)
}

// 添加 EditMode Label 标识
func (r *ManifestRenderer) setEditMode() error {
	// 若原始配置中没有 labels，则默认新建
	if labels, _ := mapx.GetItems(r.Manifest, "metadata.labels"); labels == nil {
		_ = mapx.SetItems(r.Manifest, "metadata.labels", map[string]interface{}{})
	}
	return mapx.SetItems(r.Manifest, []string{"metadata", "labels", res.EditModeLabelKey}, res.EditModeForm)
}

const (
	// RandomSuffixLength 资源名称随机后缀长度
	RandomSuffixLength = 8
	// SuffixCharset 后缀可选字符集（小写 + 数字）
	SuffixCharset = "abcdefghijklmnopqrstuvwxyz1234567890"
)

// SchemaRenderer 渲染并加载表单 Schema 模板
type SchemaRenderer struct {
	Kind    string
	Context map[string]interface{}
}

// NewSchemaRenderer ...
func NewSchemaRenderer(kind, namespace string) *SchemaRenderer {
	// 若没有指定命名空间，则使用 default
	if namespace == "" {
		namespace = "default"
	}
	// 避免名称重复，每次默认添加随机后缀
	randSuffix := stringx.Rand(RandomSuffixLength, SuffixCharset)
	return &SchemaRenderer{
		Kind: kind,
		Context: map[string]interface{}{
			"kind":      kind,
			"namespace": namespace,
			"resName":   fmt.Sprintf("%s-%s", strings.ToLower(kind), randSuffix),
		},
	}
}

// Render ...
func (r *SchemaRenderer) Render() (ret map[string]interface{}, err error) {
	// 1. 检查指定资源类型是否支持表单化
	if _, ok := FormRenderSupportedResAPIVersion[r.Kind]; !ok {
		return nil, errorx.New(errcode.Unsupported, "资源类型 %s 不支持表单化", r.Kind)
	}

	// 表单模板 Schema 包含原始 Schema + Layout 信息，两者格式不同，因此分别加载
	schema := map[string]interface{}{}
	if err = r.renderSubTypeTmpl2Map("schema", &schema); err != nil {
		return nil, err
	}

	layout := []interface{}{}
	if err = r.renderSubTypeTmpl2Map("layout", &layout); err != nil {
		return nil, err
	}

	return map[string]interface{}{"schema": schema, "layout": layout, "rules": genSchemaRules()}, nil
}

func (r *SchemaRenderer) renderSubTypeTmpl2Map(subType string, ret interface{}) error {
	// 1. 加载模板并初始化
	tmpl, err := initTemplate(fmt.Sprintf("%s/%s/", envs.FormTmplFileBaseDir, subType), "*")
	if err != nil {
		return errorx.New(errcode.General, "加载模板失败：%v", err)
	}

	// 2. 渲染模板并转换成 Map 格式（模板名称格式：{r.Kind}.yaml）
	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, r.Kind+".yaml", r.Context)
	if err != nil {
		return errorx.New(errcode.General, "渲染模板失败：%v", err)
	}
	return yaml.Unmarshal(buf.Bytes(), ret)
}

// 模板初始化（含挂载 include 方法等）
func initTemplate(baseDir, tmplPattern string) (*template.Template, error) {
	funcMap := newTmplFuncMap()
	tmpl, err := template.New(
		stringx.Rand(TmplRandomNameLength, ""),
	).Funcs(funcMap).ParseFS(os.DirFS(baseDir), tmplPattern)
	if err != nil {
		return nil, err
	}

	// Add the 'include' function here, so we can close over t.
	// ref: https://github.com/helm/helm/blob/3d1bc72827e4edef273fb3d8d8ded2a25fa6f39d/pkg/engine/engine.go#L107
	includedNames := map[string]int{}
	funcMap["include"] = func(name string, data interface{}) (string, error) {
		var buf strings.Builder
		if v, ok := includedNames[name]; ok {
			if v > RecursionMaxNums {
				return "", errorx.New(errcode.Unsupported, "rendering template has a nested ref name: %s", name)
			}
			includedNames[name]++
		} else {
			includedNames[name] = 1
		}
		err := tmpl.ExecuteTemplate(&buf, name, data)
		includedNames[name]--
		return buf.String(), err
	}

	return tmpl.Funcs(funcMap), nil
}

// 生成 JsonSchema 校验规则
func genSchemaRules() map[string]interface{} {
	return map[string]interface{}{
		"required": map[string]interface{}{
			"validator": "{{ $self.value != '' }}",
			"message":   "值不能为空",
		},
		"nameRegex": map[string]interface{}{
			"validator": "/^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/",
			"message":   "仅支持小写字母，数字及 '-' 且需以字母数字开头和结尾",
		},
		"numberRegex": map[string]interface{}{
			"validator": "/^[0-9]*$/",
			"message":   "仅可包含数字字符",
		},
		"maxLength64": map[string]interface{}{
			"validator": "{{ $self.value.length < 64 }}",
			"message":   "超过长度限制（64）",
		},
		"maxLength128": map[string]interface{}{
			"validator": "{{ $self.value.length < 128 }}",
			"message":   "超过长度限制（128）",
		},
		"maxLength250": map[string]interface{}{
			"validator": "{{ $self.value.length < 250 }}",
			"message":   "超过长度限制（250）",
		},
		// 规则：https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set
		"labelKeyRegex": map[string]interface{}{
			"validator": "/^[a-z0-9A-Z]([-_a-z0-9A-Z]*[a-z0-9A-Z])?((\\.|\\/)[a-z0-9A-Z]([-_a-z0-9A-Z]*[a-z0-9A-Z])?)*$/",
			"message":   "仅支持字母，数字，'-'，'_' 及 '/' 且需以字母数字开头和结尾",
		},
		// NOTE 标签值允许为空
		"labelValRegex": map[string]interface{}{
			"validator": "/(^$|^[a-z0-9A-Z]([-_a-z0-9A-Z]*[a-z0-9A-Z])?(\\.[a-z0-9A-Z]([-_a-z0-9A-Z]*[a-z0-9A-Z])?)*$)/",
			"message":   "需以字母数字开头和结尾，可包含 '-'，'_'，'.' 和字母数字",
		},
	}
}
