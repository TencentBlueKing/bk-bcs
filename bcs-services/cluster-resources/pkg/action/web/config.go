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

	spb "google.golang.org/protobuf/types/known/structpb"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/featureflag"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// GenListConfigWebAnnos 生成获取配置类资源列表请求的 WebAnnotations
func GenListConfigWebAnnos(ctx context.Context, configListRespData map[string]interface{}) (*spb.Struct, error) {
	annoFuncs := []AnnoFunc{
		NewFeatureFlag(featureflag.FormCreate, true),
	}
	manifests := mapx.GetList(configListRespData, "manifest.items")
	annoFuncs = append(annoFuncs, genResListImmutableAnnoFuncs(ctx, manifests)...)
	return NewAnnos(annoFuncs...).ToPbStruct()
}

// GenRetrieveConfigWebAnnos 生成获取单个配置类资源请求的 WebAnnotations
func GenRetrieveConfigWebAnnos(ctx context.Context, configRespData map[string]interface{}) (*spb.Struct, error) {
	tips := genImmutableTips(ctx, mapx.GetMap(configRespData, "manifest"))
	annoFuncs := []AnnoFunc{
		NewFeatureFlag(featureflag.FormUpdate, true),
		NewPagePerm(UpdateBtn, PermDetail{Clickable: tips == "", Tip: tips}),
	}
	return NewAnnos(annoFuncs...).ToPbStruct()
}

// genResListImmutableAnnoFuncs 生成 ConfigMap/Secret 各行操作权限的 web 注解（目前主要用于不可变更的标识）
func genResListImmutableAnnoFuncs(ctx context.Context, manifests []interface{}) []AnnoFunc {
	annoFuncs := []AnnoFunc{}
	for _, manifest := range manifests {
		m := manifest.(map[string]interface{})
		tips := genImmutableTips(ctx, m)
		annoFuncs = append(annoFuncs, NewItemPerm(
			ResUID(mapx.GetStr(m, "metadata.uid")), UpdateBtn, PermDetail{Clickable: tips == "", Tip: tips}),
		)
	}
	return annoFuncs
}

// genImmutableTips 生成不可变更相关的 Tips
func genImmutableTips(ctx context.Context, manifest map[string]interface{}) string {
	if mapx.GetBool(manifest, "immutable") {
		return i18n.GetMsg(ctx, "当前资源已设置为不可变更状态，无法编辑")
	}
	return ""
}
