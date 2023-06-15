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

// Package resource xxx
package resource

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action"
	respUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resp"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// EnabledSelectItemsAPIResKind 允许使用 SelectItems API 的资源类型，若有需要可扩展
var EnabledSelectItemsAPIResKind = []string{resCsts.Deploy, resCsts.GDeploy, resCsts.STS, resCsts.GSTS}

// GetResSelectItems 为前端下拉框提供数据的 API，目前主要是 HPA 的 Schema 使用
// 可以考虑其他的资源类型也走这个 API 而不是指定资源的 List API + Format（selectItems）
func (h *Handler) GetResSelectItems(
	ctx context.Context, req *clusterRes.GetResSelectItemsReq, resp *clusterRes.CommonResp,
) (err error) {
	if req.Namespace == "" {
		return errorx.New(errcode.ValidateErr, i18n.GetMsg(ctx, "需要指定命名空间"))
	}
	if !slice.StringInSlice(req.Kind, EnabledSelectItemsAPIResKind) {
		return errorx.New(errcode.ValidateErr, i18n.GetMsg(ctx, "当前资源类型 %s 不受支持"), req.Kind)
	}
	resp.Data, err = respUtil.BuildListAPIResp(
		ctx, respUtil.ListParams{
			ClusterID:    req.ClusterID,
			ResKind:      req.Kind,
			GroupVersion: "",
			Namespace:    req.Namespace,
			Format:       action.SelectItemsFormat,
			Scene:        req.Scene,
		}, metav1.ListOptions{},
	)
	return err
}
