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

package trans

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/renderer"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

// New 根据 Format 类型，生成不同的 Manifest 转换器
func New(ctx context.Context, rawData map[string]interface{}, clusterID, kind, format string) (Transformer, error) {
	switch format {
	case action.ManifestFormat:
		return &DummyTransformer{manifest: rawData}, nil
	case action.FormDataFormat:
		return &FormDataTransformer{ctx: ctx, formData: rawData, clusterID: clusterID, kind: kind}, nil
	default:
		return nil, errorx.New(errcode.Unsupported, "不受支持的转换器格式：%s", format)
	}
}

// DummyTransformer 无需转换操作的
type DummyTransformer struct {
	manifest map[string]interface{}
}

// ToManifest 转换成 Manifest
func (t *DummyTransformer) ToManifest() (map[string]interface{}, error) {
	return t.manifest, nil
}

// FormDataTransformer 表单数据转 Manifest
type FormDataTransformer struct {
	ctx       context.Context
	formData  map[string]interface{}
	clusterID string
	kind      string
}

// ToManifest 转换成 Manifest
func (t *FormDataTransformer) ToManifest() (map[string]interface{}, error) {
	return renderer.NewManifestRenderer(t.ctx, t.formData, t.clusterID, t.kind).Render()
}
