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
	"encoding/json"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
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
		r.renderToMap,
	} {
		if err := f(); err != nil {
			return nil, err
		}
	}
	return r.Manifest, nil
}

// 获取资源对应 APIVersion 并更新 Renderer 配置
func (r *ManifestRenderer) setAPIVersion() error {
	switch r.Kind {
	case res.CJ:
		r.APIVersion = res.DefaultCJGroupVersion
	case res.HPA:
		r.APIVersion = res.DefaultHPAGroupVersion
	default:
		resInfo, err := res.GetGroupVersionResource(r.ctx, res.NewClusterConfig(r.ClusterID), r.Kind, "")
		if err != nil {
			return errorx.New(errcode.General, "获取资源 APIVersion 信息失败：%v", err)
		}
		r.APIVersion = resInfo.Group + "/" + resInfo.Version
	}
	r.FormData["apiVersion"] = r.APIVersion
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
	r.Tmpl, err = initTemplate(envs.FormFileBaseDir + "/tmpl/*")
	return err
}

// 渲染模板并转换成 Map 格式
func (r *ManifestRenderer) renderToMap() error {
	// 渲染，转换并写入数据（模板名称格式：{r.Kind}.yaml）
	var buf bytes.Buffer
	err := r.Tmpl.ExecuteTemplate(&buf, r.Kind+".yaml", r.FormData)
	if err != nil {
		log.Warn(r.ctx, "渲染模板失败：%v", err)
		return errorx.New(errcode.General, "渲染模板失败：%v", err)
	}
	return yaml.Unmarshal(buf.Bytes(), r.Manifest)
}

// SchemaRenderer 渲染并加载表单 Schema 模板
type SchemaRenderer struct {
	Kind   string
	Tmpl   *template.Template
	Schema map[string]interface{}
}

// NewSchemaRenderer ...
func NewSchemaRenderer(kind string) *SchemaRenderer {
	return &SchemaRenderer{Kind: kind, Schema: map[string]interface{}{}}
}

// Render ...
func (r *SchemaRenderer) Render() (ret map[string]interface{}, err error) {
	// 1. 检查指定资源类型是否支持表单化
	if _, ok := FormRenderSupportedResAPIVersion[r.Kind]; !ok {
		return nil, errorx.New(errcode.Unsupported, "资源类型 %s 不支持表单化", r.Kind)
	}

	// 2. 加载模板并初始化
	r.Tmpl, err = initTemplate(envs.FormFileBaseDir + "/schema/*")
	if err != nil {
		return nil, errorx.New(errcode.General, "加载模板失败：%v", err)
	}

	// 3. 渲染模板并转换成 Map 格式（模板名称格式：{r.Kind}.json）
	var buf bytes.Buffer
	err = r.Tmpl.ExecuteTemplate(&buf, r.Kind+".json", nil)
	if err != nil {
		return nil, errorx.New(errcode.General, "渲染模板失败：%v", err)
	}
	err = json.Unmarshal(buf.Bytes(), &r.Schema)
	return r.Schema, err
}

// 模板初始化（含挂载 include 方法等）
func initTemplate(tmplPattern string) (*template.Template, error) {
	funcMap := newTmplFuncMap()
	tmpl, err := template.New(stringx.Rand(TmplRandomNameLength, "")).Funcs(funcMap).ParseGlob(tmplPattern)
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
