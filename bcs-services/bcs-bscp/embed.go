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

package bscp

import (
	"embed"
	"html/template"
	"io/fs"
	"strings"
)

//go:embed ui/dist
var frontendAssets embed.FS

// Config 前端配置
type Config struct {
	Docs map[string]string `json:"docs"`
}

// WebStatic 静态资源
func WebStatic() fs.FS {
	static, err := fs.Sub(frontendAssets, "ui/dist")
	if err != nil {
		panic(err)
	}
	return static
}

// WebFaviconPath 站点 icon 路径
func WebFaviconPath() string {
	entrys, err := frontendAssets.ReadDir("ui/dist/static")
	if err != nil {
		panic(err)
	}
	for _, v := range entrys {
		if v.IsDir() {
			continue
		}
		if strings.Contains(v.Name(), "favicon") {
			return "/static/" + v.Name()
		}
	}
	panic("favicon not found")
}

// WebTemplate html 摸版
func WebTemplate() *template.Template {
	tpl := template.Must(template.New("").ParseFS(frontendAssets, "ui/dist/*.html"))
	return tpl
}
