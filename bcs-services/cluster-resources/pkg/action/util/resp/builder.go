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
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

// NewGetRespDataBuilder 根据 Format 类型，生成不同的 Retrieve 请求响应数据生成器
func NewGetRespDataBuilder(manifest map[string]interface{}, kind, format string) (GetRespDataBuilder, error) {
	switch format {
	case action.ManifestFormat:
		return &ManifestRespBuilder{manifest: manifest, kind: kind}, nil
	case action.FormDataFormat:
		return &FormDataRespBuilder{manifest: manifest, kind: kind}, nil
	default:
		return nil, errorx.New(errcode.Unsupported, "不受支持的生成器格式：%s", format)
	}
}

// ManifestRespBuilder 提供 manifest && manifestExt
type ManifestRespBuilder struct {
	manifest map[string]interface{}
	kind     string
}

// Do ...
func (b *ManifestRespBuilder) Do() (map[string]interface{}, error) {
	return map[string]interface{}{
		"manifest":    b.manifest,
		"manifestExt": formatter.GetFormatFunc(b.kind)(b.manifest),
	}, nil
}

// FormDataRespBuilder 表单数据转 Manifest
type FormDataRespBuilder struct {
	manifest map[string]interface{}
	kind     string
}

// Do ...
func (b *FormDataRespBuilder) Do() (map[string]interface{}, error) {
	parseFunc, err := parser.GetResParseFunc(b.kind)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"formData": parseFunc(b.manifest),
	}, nil
}
