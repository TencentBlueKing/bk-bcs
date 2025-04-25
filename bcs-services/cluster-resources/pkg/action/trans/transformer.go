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

// Package trans xxx
package trans

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/renderer"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

// New 根据 Format 类型，生成不同的 Manifest 转换器
func New(
	ctx context.Context, rawData map[string]interface{}, clusterID, kind, uAction, format string,
) (Transformer, error) {
	switch format {
	case action.DefaultFormat, action.ManifestFormat:
		return &DummyTransformer{ctx: ctx, action: uAction, kind: kind, manifest: rawData}, nil
	case action.FormDataFormat:
		return &FormDataTransformer{ctx: ctx, formData: rawData, clusterID: clusterID, kind: kind, action: uAction}, nil
	default:
		return nil, errorx.New(errcode.Unsupported, i18n.GetMsg(ctx, "不受支持的转换器格式：%s"), format)
	}
}

// DummyTransformer 无需转换操作的
type DummyTransformer struct {
	ctx      context.Context
	manifest map[string]interface{}
	kind     string
	action   string
}

// ToManifest 转换成 Manifest
func (t *DummyTransformer) ToManifest() (map[string]interface{}, error) {
	// 对注解进行检查，如果不存在注解，则初始化以便于后续填充信息
	annos := mapx.GetMap(t.manifest, "metadata.annotations")
	if len(annos) == 0 {
		_ = mapx.SetItems(t.manifest, "metadata.annotations", map[string]interface{}{})
	}

	// 使用原生 Manifest 作为创建 / 更新配置时，检查 editMode，如果值不为空，则设置为 yaml，
	// 避免出现使用 yaml 模式后依然使用表单进行编辑导致的表单未支持字段配置丢失的情况
	paths := []string{"metadata", "annotations", resCsts.EditModeAnnoKey}
	if editMode := mapx.GetStr(t.manifest, paths); editMode != "" {
		if err := mapx.SetItems(t.manifest, paths, resCsts.EditModeYaml); err != nil {
			return nil, err
		}
	}

	// 检查是否需要强制移除 resourceVersion，仅针对特定类型的资源
	if slice.StringInSlice(t.kind, resCsts.RemoveResVersionKinds) {
		delete(t.manifest["metadata"].(map[string]interface{}), "resourceVersion")
	}

	username := stringx.ReplaceIllegalChars(t.ctx.Value(ctxkey.UsernameKey).(string))
	creatorPaths := []string{"metadata", "annotations", resCsts.CreatorAnnoKey}
	creatorLabelPaths := []string{"metadata", "labels", resCsts.CreatorAnnoKey}
	updaterPaths := []string{"metadata", "annotations", resCsts.UpdaterAnnoKey}
	updaterLabelPaths := []string{"metadata", "labels", resCsts.UpdaterAnnoKey}
	switch t.action {
	// 若操作为创建，则保存资源的创建/编辑者信息
	case resCsts.CreateAction:
		_ = mapx.SetItems(t.manifest, creatorPaths, username)
		_ = mapx.SetItems(t.manifest, creatorLabelPaths, username)
		_ = mapx.SetItems(t.manifest, updaterPaths, username)
		_ = mapx.SetItems(t.manifest, updaterLabelPaths, username)
	// 若操作为更新，则更新资源的编辑者信息
	case resCsts.UpdateAction:
		_ = mapx.SetItems(t.manifest, updaterPaths, username)
		_ = mapx.SetItems(t.manifest, updaterLabelPaths, username)
	}

	return t.manifest, nil
}

// FormDataTransformer 表单数据转 Manifest
type FormDataTransformer struct {
	ctx       context.Context
	formData  map[string]interface{}
	clusterID string
	kind      string
	action    string
}

// ToManifest 转换成 Manifest
func (t *FormDataTransformer) ToManifest() (map[string]interface{}, error) {
	// ManifestRenderer Render 会标识 EditMode == form
	return renderer.NewManifestRenderer(t.ctx, t.formData, t.clusterID, "", t.kind, t.action, false).Render()
}
