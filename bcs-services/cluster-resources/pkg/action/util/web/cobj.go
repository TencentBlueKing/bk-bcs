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
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/validator"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// GenCObjListWebAnnos 生成获取自定义资源列表请求的 WebAnnotations
func GenCObjListWebAnnos(
	ctx context.Context, cobjListRespData map[string]interface{}, crdInfo map[string]interface{}, format string,
) (*spb.Struct, error) {
	kind := crdInfo["kind"].(string)

	annoFuncs := []AnnoFunc{
		NewFeatureFlag(featureflag.FormCreate, validator.IsFormSupportedCObjKinds(kind)),
		NewAdditionalColumns(crdInfo["addColumns"].([]interface{})),
	}

	// GameDeployment 支持删除保护，如果格式是 manifest 会逐个检查能否删除，并补充不可删除的 tips
	if kind == res.GDeploy && (format == action.DefaultFormat || format == action.ManifestFormat) {
		gdeployList := mapx.GetList(cobjListRespData, "manifest.items")
		annoFuncs = append(annoFuncs, genGDeployListOpAnnoFuncs(ctx, gdeployList)...)
	}

	return NewAnnos(annoFuncs...).ToPbStruct()
}

// 生成 GameDeployment 各行操作权限的 web 注解（目前主要用于删除保护的标识）
func genGDeployListOpAnnoFuncs(ctx context.Context, gdeployList []interface{}) []AnnoFunc {
	annoFuncs := []AnnoFunc{}
	for _, gdeploy := range gdeployList {
		gd := gdeploy.(map[string]interface{})
		uid := ResUID(mapx.GetStr(gd, "metadata.uid"))
		replicas := mapx.GetInt64(gd, "spec.replicas")
		editMode := mapx.Get(gd, []string{"metadata", "annotations", res.EditModeAnnoKey}, res.EditModeYaml)

		paths := []string{"metadata", "labels", res.DeletionProtectLabelKey}
		deletionProtectPolicy := mapx.Get(gd, paths, res.GDeployDeletionProtectPolicyNotAllow)
		if deletionProtectPolicy == res.GDeployDeletionProtectPolicyAlways {
			annoFuncs = append(annoFuncs, NewItemPerm(uid, DeleteBtn, PermDetail{Clickable: true}))
			continue
		}

		// 剩下的都是不允许删除的情况（除非 Cascading && replicas == 0）
		tips, clickable := "", false
		if editMode == res.EditModeForm {
			tips = i18n.GetMsg(ctx, "配置信息->删除保护策略->总是允许删除")
		} else {
			tips = fmt.Sprintf(
				i18n.GetMsg(ctx, "标签字段 %s: %s"),
				res.DeletionProtectLabelKey,
				res.GDeployDeletionProtectPolicyAlways,
			)
		}
		if deletionProtectPolicy == res.GDeployDeletionProtectPolicyCascading {
			if replicas == 0 {
				tips, clickable = "", true
			} else {
				tips += i18n.GetMsg(ctx, "或确保实例数量为 0")
			}
		}
		if tips != "" {
			tips = i18n.GetMsg(ctx, "当前实例已添加删除保护功能，若确认要删除，请修改实例") + tips
		}
		annoFuncs = append(annoFuncs, NewItemPerm(uid, DeleteBtn, PermDetail{Clickable: clickable, Tip: tips}))
	}
	return annoFuncs
}
