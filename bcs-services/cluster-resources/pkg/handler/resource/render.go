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

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/renderer"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// FormDataRenderPreview xxx
func (h *Handler) FormDataRenderPreview(
	ctx context.Context, req *clusterRes.FormRenderPreviewReq, resp *clusterRes.CommonResp,
) error {
	manifest, err := renderer.NewManifestRenderer(ctx, req.FormData.AsMap(), req.ClusterID, req.Kind).Render()
	if err != nil {
		return err
	}
	resp.Data, err = pbstruct.Map2pbStruct(manifest)
	return err
}
