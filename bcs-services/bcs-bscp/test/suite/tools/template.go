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

package main

import (
	"html/template"
	"io"
)

var tp *defaultTemplate

type defaultTemplate struct {
	engine *template.Template
}

func init() {
	t := template.New("template")
	t = template.Must(t.Parse(htmlTemplate))

	tp = &defaultTemplate{
		engine: t,
	}
}

func (t defaultTemplate) render(wr io.Writer, results []*StatisticalResults) {
	t.engine.Execute(wr, map[string][]*StatisticalResults{ // nolint error not checked
		"Results": results,
	})
}

// htmlTemplate test results file statistical report html template.
const htmlTemplate = `
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<style type="text/css">
	table.gridtable {
		font-family: verdana,arial,sans-serif;
		font-size:12px;
		color:#333333;
		border-width: 1px;
		border-color: #666666;
		border-collapse: collapse;
	}
	table.gridtable th {
		border-width: 1px;
		padding: 8px;
		border-style: solid;
		border-color: #666666;
		background-color: #dedede;
	}
	table.gridtable td {
		border-width: 1px;
		padding: 8px;
		border-style: solid;
		border-color: #666666;
		background-color: #ffffff;
	}
</style>

<table class="gridtable">
	<thead>
		<tr>
			<th>Title</th>
			<th>Total</th>
			<th>Succeed</th>
			<th>Failed</th>
			<th>Failed Info</th>
		</tr>
	</thead>

	<tbody>
		{{ with .Results }}
		{{- range . }}
		<tr>
			<td>{{.Title}}</td>
			<td>{{.Total}}</td>
			<td>{{.Succeed}}</td>
			<td>
				{{ if eq .Failed 0 }}
				  	<p style='color:blue;'>0</p>
				{{ else }}
					<p style='color:red;'>{{.Failed}}</p>
				{{ end }}
			</td>

			<td>
			<table class="gridtable">
			{{ with .FailedInfos }}
			{{- range . }}
				<tr style='color:red;'>
					<td>{{.Line}}</td>
					<td>{{.Message}}</td>
					<td>{{.Total}}</td>
				</tr>
			{{- end }}
			{{ end }}
			</table>
			</td>
		</tr>
		{{- end }}
		{{ end }}
	</tbody>
</table>
<br>
`
