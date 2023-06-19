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

package web

import (
	"context"
	"fmt"

	spb "google.golang.org/protobuf/types/known/structpb"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/featureflag"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/validator"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
)

// GenListCObjWebAnnos 生成获取自定义资源列表请求的 WebAnnotations
func GenListCObjWebAnnos(
	ctx context.Context, cobjListRespData map[string]interface{}, crdInfo map[string]interface{}, format string,
) (*spb.Struct, error) {
	kind := crdInfo["kind"].(string)

	annoFuncs := []AnnoFunc{
		NewFeatureFlag(featureflag.FormCreate, validator.IsFormSupportedCObjKinds(kind)),
		NewAdditionalColumns(mapx.GetList(crdInfo, "addColumns")),
	}

	if requireDeletionProtectWebAnno(kind, format) {
		manifests := mapx.GetList(cobjListRespData, "manifest.items")
		annoFuncs = append(annoFuncs, genResListDeleteProtectAnnoFuncs(ctx, manifests, kind)...)
	}

	return NewAnnos(annoFuncs...).ToPbStruct()
}

// GenRetrieveCObjWebAnnos 生成获取单个自定义资源请求的 WebAnnotations
func GenRetrieveCObjWebAnnos(
	ctx context.Context, cobjManifest map[string]interface{}, crdInfo map[string]interface{}, format string,
) (*spb.Struct, error) {
	kind := crdInfo["kind"].(string)

	annoFuncs := []AnnoFunc{
		NewFeatureFlag(featureflag.FormUpdate, validator.IsFormSupportedCObjKinds(kind)),
		NewAdditionalColumns(mapx.GetList(crdInfo, "addColumns")),
	}
	if requireDeletionProtectWebAnno(kind, format) {
		tips := genDeleteProtectTips(ctx, mapx.GetMap(cobjManifest, "manifest"), kind)
		annoFuncs = append(annoFuncs, NewPagePerm(DeleteBtn, PermDetail{Clickable: tips == "", Tip: tips}))
	}

	return NewAnnos(annoFuncs...).ToPbStruct()
}

// requireDeletionProtectWebAnno 判断是否需要提供删除保护的 webAnnotations 信息
func requireDeletionProtectWebAnno(kind, format string) bool {
	return slice.StringInSlice(kind, []string{resCsts.GDeploy, resCsts.HookTmpl, resCsts.GSTS}) &&
		(format == action.DefaultFormat || format == action.ManifestFormat)
}

// genResListDeleteProtectAnnoFuncs 生成 GameDeployment/HookTemplate 各行操作权限的 web 注解（目前主要用于删除保护的标识）
func genResListDeleteProtectAnnoFuncs(ctx context.Context, manifests []interface{}, kind string) []AnnoFunc {
	annoFuncs := []AnnoFunc{}
	for _, manifest := range manifests {
		m := manifest.(map[string]interface{})
		tips := genDeleteProtectTips(ctx, m, kind)
		annoFuncs = append(annoFuncs, NewItemPerm(
			ResUID(mapx.GetStr(m, "metadata.uid")), DeleteBtn, PermDetail{Clickable: tips == "", Tip: tips}),
		)
	}
	return annoFuncs
}

// genDeleteProtectTips 生成删除保护相关的 Tips，为空表示允许删除
func genDeleteProtectTips(ctx context.Context, manifest map[string]interface{}, kind string) string {
	replicas := mapx.GetInt64(manifest, "spec.replicas")
	editMode := mapx.Get(manifest, []string{"metadata", "annotations", resCsts.EditModeAnnoKey}, resCsts.EditModeYaml)

	paths := []string{"metadata", "labels", resCsts.DeletionProtectLabelKey}
	dpPolicy := mapx.Get(manifest, paths, resCsts.DeletionProtectPolicyNotAllow)
	if dpPolicy == resCsts.DeletionProtectPolicyAlways {
		return ""
	}

	var tips string
	if editMode == resCsts.EditModeForm {
		tips = i18n.GetMsg(ctx, "配置信息->删除保护策略->总是允许删除")
	} else {
		tips = fmt.Sprintf(
			i18n.GetMsg(ctx, "标签字段 %s: %s"),
			resCsts.DeletionProtectLabelKey,
			resCsts.DeletionProtectPolicyAlways,
		)
	}
	// 当类型是 GameDeploy/GameSTS，保护策略为 Cascading，实例数为 0 时候是可以删除的
	if (kind == resCsts.GDeploy || kind == resCsts.GSTS) && dpPolicy == resCsts.DeletionProtectPolicyCascading {
		if replicas == 0 {
			return ""
		}
		tips += i18n.GetMsg(ctx, "或确保实例数量为 0")
	}
	return i18n.GetMsg(ctx, "当前实例已添加删除保护功能，若确认要删除，请修改实例") + tips
}
