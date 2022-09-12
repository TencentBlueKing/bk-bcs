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
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

// RecursionMaxNums 模板 include 嵌套最大层数
const RecursionMaxNums = 100

// TmplRandomNameLength 模板随机名称长度
const TmplRandomNameLength = 12

// newTmplFuncMap xxx
// ref: https://github.com/helm/helm/blob/a499b4b179307c267bdf3ec49b880e3dbd2a5591/pkg/engine/funcs.go#L44
func newTmplFuncMap() template.FuncMap {
	f := sprig.TxtFuncMap()

	extra := template.FuncMap{
		"toYaml":                 toYaml,
		"filterMatchKVFormSlice": slice.FilterMatchKVFromSlice,
		"matchKVInSlice":         slice.MatchKVInSlice,
		"i18n":                   i18n.GetMsgWithLang,

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
		// Swallow errors inside of a template.
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}

// initTemplate 模板初始化（含挂载 include 方法等）
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
