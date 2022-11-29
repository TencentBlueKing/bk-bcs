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

package resource

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/renderer"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// FormDataRenderPreview xxx
func (h *Handler) FormDataRenderPreview(
	ctx context.Context, req *clusterRes.FormRenderPreviewReq, resp *clusterRes.CommonResp,
) error {
	// 在 ManifestRenderer 中，对于不存在的创建者/更新者都会新建，因此这里直接指定 UpdateAction 即可
	manifest, err := renderer.NewManifestRenderer(
		ctx, req.FormData.AsMap(), req.ClusterID, req.Kind, resCsts.UpdateAction,
	).Render()
	if err != nil {
		return errorx.New(errcode.General, i18n.GetMsg(ctx, "预览表单渲染结果失败，请检查您填写的表单配置；错误信息：%w"), err)
	}
	resp.Data, err = pbstruct.Map2pbStruct(manifest)
	return err
}
