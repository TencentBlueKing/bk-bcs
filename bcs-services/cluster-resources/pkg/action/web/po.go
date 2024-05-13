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

package web

import (
	"context"

	spb "google.golang.org/protobuf/types/known/structpb"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	nsAuth "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/namespace"
)

// GenNodePodListWebAnnos 生成 ListPoWithNodeName 请求的 WebAnnotations
func GenNodePodListWebAnnos(
	ctx context.Context, podList []map[string]interface{}, projectID, clusterID string, namespaces []string,
) (*spb.Struct, error) {
	username := ctx.Value(ctxkey.UsernameKey).(string)
	scopePerms, err := nsAuth.FetchBatchNSScopedResMultiActPerm(
		username, projectID, clusterID, namespaces, "pods",
	)
	if err != nil {
		return nil, err
	}

	annoFuncs := []AnnoFunc{}
	permAction2WebObjMap := map[string]ObjName{
		nsAuth.NamespaceScopedView:   DetailBtn,
		nsAuth.NamespaceScopedUpdate: UpdateBtn,
		nsAuth.NamespaceScopedDelete: DeleteBtn,
	}
	for _, po := range podList {
		for action, webObj := range permAction2WebObjMap {
			clickable := false
			if scopePerm, ok := scopePerms[po["namespace"].(string)]; ok {
				clickable = scopePerm[action]
			}
			annoFuncs = append(annoFuncs, NewItemPerm(
				ResUID(po["uid"].(string)), webObj, PermDetail{Clickable: clickable},
			))
		}
	}
	return NewAnnos(annoFuncs...).ToPbStruct()
}
