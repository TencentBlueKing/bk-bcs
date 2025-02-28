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

// Package renderer xxx
package renderer

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

// RecursionMaxNums 模板 include 嵌套最大层数
const RecursionMaxNums = 100

// TmplRandomNameLength 模板随机名称长度
const TmplRandomNameLength = 12

// ref: https://github.com/helm/helm/blob/a499b4b179307c267bdf3ec49b880e3dbd2a5591/pkg/engine/funcs.go#L44
func newTmplFuncMap() template.FuncMap {
	f := sprig.TxtFuncMap()

	extra := template.FuncMap{
		// 功能类方法
		"toYaml":                 toYaml,
		"toJson":                 toJson,
		"filterMatchKVFormSlice": slice.FilterMatchKVFromSlice,
		"matchKVInSlice":         slice.MatchKVInSlice,
		"i18n":                   i18n.GetMsgWithLang,
		"genDockerConfigJson":    genDockerConfigJSON,
		"contains":               contains,
		"toInt":                  toInt,

		// 辅助类方法
		"isNSRequired":             isNSRequired,
		"isLabelRequired":          isLabelRequired,
		"isLabelVisible":           isLabelVisible,
		"isLabelAsSelector":        isLabelAsSelector,
		"isLabelEditDisabled":      isLabelEditDisabled,
		"hasLabelSelector":         hasLabelSelector,
		"isAnnoVisible":            isAnnoVisible,
		"canRenderResVersion":      canRenderResVersion,
		"genWorkloadSelectorLabel": genWorkloadSelectorLabel,

		// This is a placeholder for the "include" function, which is late-bound to a template.
		// By declaring it here, we preserve the integrity of the linter.
		"include": func(string, interface{}) string { return "not implemented" },
	}

	for k, v := range extra {
		f[k] = v
	}
	return f
}

// toYaml takes an interface, marshals it to yaml, and returns a string. It will
// always return a string, even on marshal error (empty string).
func toYaml(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside a template.
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}

// toJson takes an interface, marshals it to json, and returns a string. It will
// always return a string, even on marshal error (empty string).
func toJson(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		// Swallow errors inside a template.
		return ""
	}
	return string(data)
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
		err = tmpl.ExecuteTemplate(&buf, name, data)
		includedNames[name]--
		return buf.String(), err
	}

	return tmpl.Funcs(funcMap), nil
}

// 指定资源类型是否必须填写命名空间
func isNSRequired(kind string) bool {
	return !slice.StringInSlice(kind, []string{resCsts.PV, resCsts.SC})
}

// 指定资源类型是否必须填写 labels
func isLabelRequired(kind string) bool {
	return !slice.StringInSlice(kind, []string{
		resCsts.HookTmpl, resCsts.Ing, resCsts.SVC, resCsts.EP, resCsts.HPA,
		resCsts.Po, resCsts.CM, resCsts.Secret, resCsts.PV, resCsts.PVC, resCsts.SC,
	})
}

// 指定资源类型是否展示 labels
func isLabelVisible(kind string) bool {
	return !slice.StringInSlice(kind, []string{resCsts.HookTmpl})
}

// 判断对于当前这种资源类型，labels 是否会被用于 selector 的配置，主要是Workloads 类型（除 Pod 外）
func isLabelAsSelector(kind string) bool {
	return slice.StringInSlice(kind, []string{
		resCsts.Deploy, resCsts.STS, resCsts.DS, resCsts.CJ, resCsts.Job, resCsts.GDeploy, resCsts.GSTS,
	})
}

// 判断对于当前这种资源类型，是否有 labelSelector
func hasLabelSelector(kind string) bool {
	return slice.StringInSlice(kind, []string{
		resCsts.Deploy, resCsts.STS, resCsts.DS, resCsts.GDeploy, resCsts.GSTS,
	})
}

// genWorkloadSelectorLabel 生成 Workload 类型资源的 selector label
func genWorkloadSelectorLabel(v string) string {
	value := fmt.Sprintf("[{key: 'workload.bcs.tencent.io/workloadSelector', value: '%s'}]", v)
	return value
}

// 判断资源类型和使用场景下，能否编辑 labels
func isLabelEditDisabled(kind, action string) bool {
	if action == "create" {
		return false
	}
	// 因设计原因，labels 也会作为某些资源 selector 的配置，因此不允许更新时候修改，后续可能会评估开放
	return isLabelAsSelector(kind)
}

// 指定资源类型是否展示 annotations
func isAnnoVisible(kind string) bool {
	return !slice.StringInSlice(kind, []string{resCsts.HookTmpl})
}

// 指定资源类型能否渲染 metadata.resourceVersion
func canRenderResVersion(kind string) bool {
	return !slice.StringInSlice(kind, resCsts.RemoveResVersionKinds)
}

// 生成 dockerconfigjson
func genDockerConfigJSON(registry, username, password string) string {
	c, err := json.Marshal(map[string]interface{}{
		"auths": map[string]interface{}{
			registry: map[string]interface{}{
				"username": username,
				"password": password,
			},
		},
	})
	if err != nil {
		return err.Error()
	}
	return string(c)
}

// contains 判断字符串是否包含子串
func contains(str, substr string) bool {
	return strings.Contains(str, substr)
}

// toInt 字符串转 int
func toInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}
