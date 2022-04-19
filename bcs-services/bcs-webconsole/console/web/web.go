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
 *
 */

package web

import (
	"embed"
	"html/template"
	"io/fs"
)

//go:embed templates static
var FS embed.FS

// WebTemplate html 摸版
func WebTemplate() *template.Template {
	tpl := template.Must(template.New("").ParseFS(FS, "templates/*.html"))
	return tpl
}

// WebStatic 静态资源
func WebStatic() fs.FS {
	static, _ := fs.Sub(FS, "static")
	return static
}
