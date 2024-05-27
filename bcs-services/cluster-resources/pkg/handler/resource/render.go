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

package resource

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/renderer"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

const (
	// YAMLVarMagic 包含该字符串的 YAML 字符串，表示该字符串是一个变量
	YAMLVarMagic = "{{"
	// YAMLVarMagicPlaceholder 变量占位符
	YAMLVarMagicPlaceholder = "__BCS_VAR__"
)

// FormDataRenderPreview xxx
func (h *Handler) FormDataRenderPreview(
	ctx context.Context, req *clusterRes.FormRenderPreviewReq, resp *clusterRes.CommonResp,
) error {
	// 在 ManifestRenderer 中，对于不存在的创建者/更新者都会新建，因此这里直接指定 UpdateAction 即可
	manifest, err := renderer.NewManifestRenderer(
		ctx, req.FormData.AsMap(), req.ClusterID, "", req.Kind, resCsts.UpdateAction, false,
	).Render()
	if err != nil {
		return errorx.New(errcode.General, i18n.GetMsg(ctx, "预览表单渲染结果失败，请检查您填写的表单配置；错误信息：%s"),
			err.Error())
	}
	resp.Data, err = pbstruct.Map2pbStruct(manifest)
	return err
}

// FormToYAML xxx
func (h *Handler) FormToYAML(ctx context.Context, req *clusterRes.FormToYAMLReq, resp *clusterRes.CommonResp) error {
	var (
		manifest string
		err      error
	)
	// 遍历多个表单模板文件
	for _, v := range req.GetResources() {
		renderManifest, errr := renderer.NewManifestRenderer(
			ctx, v.GetFormData().AsMap(), "", v.GetApiVersion(), v.GetKind(), resCsts.UpdateAction, true,
		).RenderString()
		if errr != nil {
			return errorx.New(errcode.General, i18n.GetMsg(ctx, "预览表单渲染结果失败，请检查您填写的表单配置；错误信息：%s"),
				errr.Error())
		}
		manifest += "\n---\n" + renderManifest
	}
	resp.Data, err = pbstruct.Map2pbStruct(map[string]interface{}{"manifest": manifest})
	return err
}

// YAMLToForm xxx
func (h *Handler) YAMLToForm(ctx context.Context, req *clusterRes.YAMLToFormReq, resp *clusterRes.CommonResp) error {
	manifests := parser.SplitManifests(req.GetYaml())
	formDatas := make([]map[string]interface{}, 0)
	var err error
	defer func() {
		if r := recover(); r != nil {
			resp.Data, err = pbstruct.Map2pbStruct(map[string]interface{}{"resources": nil, "canTransform": false,
				"message": fmt.Sprintf("%v", r)})
		}
	}()

	for _, v := range manifests {
		// 变量不能 yaml 解析，先替换为占位符
		v = strings.ReplaceAll(v, ": "+YAMLVarMagic, ": "+YAMLVarMagicPlaceholder)
		manifest := map[string]interface{}{}
		if errr := yaml.Unmarshal([]byte(v), &manifest); errr != nil {
			resp.Data, err = pbstruct.Map2pbStruct(map[string]interface{}{"resources": nil, "canTransform": false,
				"message": errr.Error()})
			return err
		}
		// 写回变量
		manifest = mapx.ReplaceMapValue(manifest, YAMLVarMagicPlaceholder, YAMLVarMagic)
		kind := mapx.GetStr(manifest, "kind")
		apiVersion := mapx.GetStr(manifest, "apiVersion")
		parseFunc, errr := parser.GetResParseFunc(ctx, kind)
		if errr != nil {
			resp.Data, err = pbstruct.Map2pbStruct(map[string]interface{}{"resources": nil, "canTransform": false,
				"message": errr.Error()})
			return err
		}
		formData := parseFunc(mapx.CleanUpMap(manifest))
		formDatas = append(formDatas, map[string]interface{}{
			"apiVersion": apiVersion, "kind": kind, "formData": formData})
	}
	resp.Data, err = pbstruct.Map2pbStruct(map[string]interface{}{"resources": formDatas, "canTransform": true})
	return err
}
