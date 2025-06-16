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

// Package util xxx
package util

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

var (
	HtmlEmailTemplate = `<!DOCTYPE html
PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">

<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">　　
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
body{ margin :0;padding: 0;font-family: "微软雅黑";line-height: 24px; }
div img table h4{margin :0;}
table { font-family: arial, sans-serif; border-collapse: collapse; width: 100%; }
td, th { border: 1px solid #c3c3c3; text-align: left; padding: 8px; }
tr:nth-child(even) { background-color: #dddddd; }
tfoot { background-color: #ffff00; }
.banner{
background-image:url('https://p.qpic.cn/xingzhengzhushou/610796532/45f14853-5747-4e96-963d-05f69d839b06/0');
background-repeat:no-repeat ;
background-size:100% 100%;
background-repeat:no-repeat;
height:300px;
text-align:center;
}
.h4{ color:#ffffff; margin :0; padding-top:200px;}
</style>
</head>

<body style="margin: 0; padding: 0;">
<div>
<img width="100%"
src="https://p.qpic.cn/xingzhengzhushou/610796532/45f14853-5747-4e96-963d-05f69d839b06/0">
<h4 color="#ffffff" margin="0" padding-top="200px">xxxxx</h4>
</div>

{{ range .}}
<h1> {{ .Title }}</h2>
<table>
	<tr>
		{{ range .Headers}}
		<th>{{ . }}</th>
		{{ end}}
	</tr>
	{{ range .Data}}
		<tr>
			{{ range . }}
			<td>{{ . }}</td>
			{{ end}}
		</tr>
	{{ end}}
</table>
{{ end}}

<table>
<tr>
<th> 11</th>
<td> 22<td>
<th> 11</th>
<td> 22<td>
<th> 11</th>
<td> 22<td>
</tr>
<tr>
<th> 11</th>
<td> 22<td>
</tr>
</table>

<table>
<tr>
		<th>Header 1</th>
		<th>Header 2</th>
		<th>Header 3</th>
		<th>Header 4</th>
	</tr>
	<tr>
		<td>Row 1, Cell 1</td>
		<td>Row 1, Cell 2</td>
		<td>Row 1, Cell 3</td>
		<!-- 第一行有3个单元格 -->
	</tr>
	<tr>
		<td>Row 2, Cell 1</td>
		<td colspan="2">Row 2, Cell 2 (spanning 2 columns)</td>
		<!-- 第二行有2个单元格，第二个单元格跨越2列 -->
	</tr>
	<tr>
		<td>Row 3, Cell 1</td>
		<td>Row 3, Cell 2</td>
		<td>Row 3, Cell 3</td>
		<td>Row 3, Cell 4</td>
		<!-- 第三行有4个单元格 -->
	</tr>
	<tr>
		<td>Row 4, Cell 1</td>
		<!-- 第四行只有1个单元格 -->
	</tr>
</table>
</body>

</html>`
)

// HtmlEmail XXX
type HtmlEmail struct {
	content string
	doc     *goquery.Document
}

// HtmlTable xxx
type HtmlTable struct {
	Title   string
	Headers []string
	Data    [][]string
}

// NewHtmlEmail xxx
func NewHtmlEmail() *HtmlEmail {
	// test goquery
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(`<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">　　
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
body{ margin :0;padding: 0;font-family: "微软雅黑";line-height: 12px; font-size: 12px}
div img table h4{margin :0;}
table { margin :30;font-family: arial, sans-serif; border-collapse: collapse; width: 100%; }
td, th { border: 1px solid #c3c3c3; text-align: left; padding: 4px; }
tr:nth-child(even) { background-color: #dddddd; }
tfoot { background-color: #ffff00; }
.h4{ color:#ffffff; margin :0; padding-top:200px;}

</style>
</head>
<body>
</body>
</html>`))
	//doc.Find("p").AfterHtml("<p>Text for article #2222222</p>")
	//doc.Find("body").AppendHtml("<p>Text for article #333333</p>")

	return &HtmlEmail{
		doc: doc,
	}
}

// GetText xxx
func (h *HtmlEmail) GetText() string {
	ret, err := h.doc.Html()
	if err != nil {
		return err.Error()
	}
	return ret
}

// Append xxx
func (h *HtmlEmail) Append(html string) {
	h.doc.Find("body").AppendHtml(html)
}

// AppendNodes xxx
func (h *HtmlEmail) AppendNodes(ns ...*html.Node) {
	h.doc.Find("body").AppendNodes(ns...)
}
