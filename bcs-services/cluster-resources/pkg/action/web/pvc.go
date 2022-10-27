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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/featureflag"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// 至多展示的 Pod 名称数量
const mountPodNameMaxDisplayNum = 2

// GenListPVCWebAnnos 生成获取配置类资源列表请求的 WebAnnotations
func GenListPVCWebAnnos(
	ctx context.Context, clusterID, namespace string, pvcListRespData map[string]interface{},
) (*spb.Struct, error) {
	annoFuncs := []AnnoFunc{
		NewFeatureFlag(featureflag.FormCreate, true),
	}
	manifests := mapx.GetList(pvcListRespData, "manifest.items")
	annoFuncs = append(annoFuncs, genResListPVCMountAnnoFuncs(ctx, clusterID, namespace, manifests)...)
	return NewAnnos(annoFuncs...).ToPbStruct()
}

// GenRetrievePVCWebAnnos 生成获取单个配置类资源请求的 WebAnnotations
func GenRetrievePVCWebAnnos(
	ctx context.Context, clusterID, namespace string, pvcRespData map[string]interface{},
) (*spb.Struct, error) {
	pvcMountInfo := cli.NewPodCliByClusterID(ctx, clusterID).GetPVCMountInfo(ctx, namespace, metav1.ListOptions{})
	tips := genPVCMountTips(ctx, mapx.GetStr(pvcRespData, "manifest.metadata.name"), pvcMountInfo)
	annoFuncs := []AnnoFunc{
		NewFeatureFlag(featureflag.FormUpdate, true),
		NewPagePerm(DeleteBtn, PermDetail{Clickable: tips == "", Tip: tips}),
	}
	return NewAnnos(annoFuncs...).ToPbStruct()
}

// genResListPVCMountAnnoFuncs 生成 ConfigMap/Secret 各行操作权限的 web 注解（目前主要用于不可变更的标识）
func genResListPVCMountAnnoFuncs(ctx context.Context, clusterID, namespace string, manifests []interface{}) []AnnoFunc {
	pvcMountInfo := cli.NewPodCliByClusterID(ctx, clusterID).GetPVCMountInfo(ctx, namespace, metav1.ListOptions{})

	annoFuncs := []AnnoFunc{}
	for _, manifest := range manifests {
		m := manifest.(map[string]interface{})
		tips := genPVCMountTips(ctx, mapx.GetStr(m, "metadata.name"), pvcMountInfo)
		annoFuncs = append(annoFuncs, NewItemPerm(
			ResUID(mapx.GetStr(m, "metadata.uid")), DeleteBtn, PermDetail{Clickable: tips == "", Tip: tips}),
		)
	}
	return annoFuncs
}

// genPVCMountTips 生成 PVC 已经被挂载相关的 Tips
func genPVCMountTips(ctx context.Context, pvcName string, pvcMountInfo map[string][]string) string {
	if podNames, ok := pvcMountInfo[pvcName]; ok {
		tips := i18n.GetMsg(ctx, "无法删除 PersistentVolumeClaim，原因：")
		if len(podNames) <= mountPodNameMaxDisplayNum {
			return tips + fmt.Sprintf(i18n.GetMsg(ctx, "已被 Pods %v 挂载"), podNames)
		}
		return tips + fmt.Sprintf(i18n.GetMsg(ctx, "已经被 %s 等共计 %d 个 Pod 挂载"), podNames[0], len(podNames))
	}
	return ""
}
