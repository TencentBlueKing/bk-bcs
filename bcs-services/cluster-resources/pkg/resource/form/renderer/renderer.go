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
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
)

// ManifestRenderer ...
type ManifestRenderer struct {
	FormData   map[string]interface{}
	ClusterID  string
	Kind       string
	APIVersion string
	Tmpl       *template.Template
	Manifest   map[string]interface{}
}

// NewManifestRenderer ...
func NewManifestRenderer(formData map[string]interface{}, clusterID, kind string) *ManifestRenderer {
	return &ManifestRenderer{FormData: formData, ClusterID: clusterID, Kind: kind, Manifest: map[string]interface{}{}}
}

// Render 渲染表单数据，返回 Manifest
func (r *ManifestRenderer) Render() (map[string]interface{}, error) {
	for _, f := range []func() error{
		// 1. 自动获取资源对应版本
		r.setResAPIVersion,
		// 2. 检查是否支持指定版本表单化
		r.checkResFormRenderable,
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
func (r *ManifestRenderer) setResAPIVersion() error {
	switch r.Kind {
	case res.CJ:
		r.APIVersion = res.DefaultCJGroupVersion
	case res.HPA:
		r.APIVersion = res.DefaultHPAGroupVersion
	default:
		resInfo, err := res.GetGroupVersionResource(res.NewClusterConfig(r.ClusterID), r.Kind, "")
		if err != nil {
			return err
		}
		r.APIVersion = resInfo.Group + "/" + resInfo.Version
	}
	r.FormData["apiVersion"] = r.APIVersion
	return nil
}

// 检查指定资源能否渲染为表单
func (r *ManifestRenderer) checkResFormRenderable() error {
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
	// return mapx.RemoveAllZeroSubMap(r.FormData)
	// TODO 修复无法正确去除零值问题
	return nil
}

// 加载模板并初始化
func (r *ManifestRenderer) initTemplate() error {
	filepath := fmt.Sprintf("%s/%s.yaml", envs.FormTmplFileBaseDir, r.Kind)
	tmplContent, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	// TODO 支持 include，如 metadata，containers 等可以引用一份模板
	// TODO []string 的展开也可以设计一个模板？
	r.Tmpl, err = template.New("tmpl").Funcs(template.FuncMap{
		"split":                  strings.Split,
		"typeMapInSlice":         slice.TypeMapInSlice,
		"filterTypeMapFormSlice": slice.FilterTypeMapFormSlice,
	}).Parse(string(tmplContent))
	return err
}

// 渲染模板并转换成 Map 格式
func (r *ManifestRenderer) renderToMap() error {
	// 渲染，转换并写入数据
	var buf bytes.Buffer
	err := r.Tmpl.Execute(&buf, r.FormData)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(buf.Bytes(), r.Manifest)
}
