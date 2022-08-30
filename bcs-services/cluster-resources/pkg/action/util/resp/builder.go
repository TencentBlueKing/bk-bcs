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

package resp

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// NewRespDataBuilder 根据 Format 类型，生成不同的 Retrieve 请求响应数据生成器
func NewRespDataBuilder(ctx context.Context, manifest map[string]interface{}, kind, format string) (DataBuilder,
	error) {
	switch format {
	case action.DefaultFormat, action.ManifestFormat:
		return &ManifestRespBuilder{ctx: ctx, manifest: manifest, kind: kind}, nil
	case action.FormDataFormat:
		return &FormDataRespBuilder{ctx: ctx, manifest: manifest, kind: kind}, nil
	case action.SelectItemsFormat:
		return &SelectItemsRespBuilder{ctx: ctx, manifest: manifest, kind: kind}, nil
	default:
		return nil, errorx.New(errcode.Unsupported, i18n.GetMsg(ctx, "不受支持的生成器格式：%s"), format)
	}
}

// ManifestRespBuilder 提供 manifest && manifestExt
type ManifestRespBuilder struct {
	ctx      context.Context
	manifest map[string]interface{}
	kind     string
}

// BuildList xxx
func (b *ManifestRespBuilder) BuildList() (map[string]interface{}, error) {
	manifestExt := map[string]interface{}{}
	formatFunc := formatter.GetFormatFunc(b.kind)
	// 遍历列表中的每个资源，生成 manifestExt
	for _, item := range b.manifest["items"].([]interface{}) {
		uid, _ := mapx.GetItems(item.(map[string]interface{}), "metadata.uid")
		manifestExt[uid.(string)] = formatFunc(item.(map[string]interface{}))
	}
	return map[string]interface{}{"manifest": b.manifest, "manifestExt": manifestExt}, nil
}

// Build xxx
func (b *ManifestRespBuilder) Build() (map[string]interface{}, error) {
	return map[string]interface{}{
		"manifest":    b.manifest,
		"manifestExt": formatter.GetFormatFunc(b.kind)(b.manifest),
	}, nil
}

// FormDataRespBuilder 表单数据转 Manifest
type FormDataRespBuilder struct {
	ctx      context.Context
	manifest map[string]interface{}
	kind     string
}

// BuildList xxx
func (b *FormDataRespBuilder) BuildList() (map[string]interface{}, error) {
	return nil, errorx.New(errcode.Unsupported, "FormDataRespBuilder.BuildList is unsupported")
}

// Build xxx
func (b *FormDataRespBuilder) Build() (map[string]interface{}, error) {
	parseFunc, err := parser.GetResParseFunc(b.ctx, b.kind)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"formData": parseFunc(b.manifest),
	}, nil
}

// SelectItemsRespBuilder 下拉框数据生成器
type SelectItemsRespBuilder struct {
	ctx      context.Context
	manifest map[string]interface{}
	kind     string
}

// BuildList xxx
func (b *SelectItemsRespBuilder) BuildList() (map[string]interface{}, error) {
	// 取每个 K8S 资源的名称，作为下拉框选项
	selectItems := []interface{}{}
	for _, item := range mapx.GetList(b.manifest, "items") {
		name := mapx.GetStr(item.(map[string]interface{}), "metadata.name")
		selectItems = append(selectItems, map[string]interface{}{
			"label": name, "value": name,
		})
	}
	return map[string]interface{}{"selectItems": selectItems}, nil
}

// Build xxx
func (b *SelectItemsRespBuilder) Build() (map[string]interface{}, error) {
	return nil, errorx.New(errcode.Unsupported, "SelectItemsRespBuilder.Build is unsupported")
}
