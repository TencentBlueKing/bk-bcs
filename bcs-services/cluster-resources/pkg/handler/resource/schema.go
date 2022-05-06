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
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/renderer"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// GetResFormSchema ...
func (h *Handler) GetResFormSchema(
	_ context.Context, req *clusterRes.GetResFormSchemaReq, resp *clusterRes.CommonResp,
) error {
	schema, err := renderer.NewSchemaRenderer(req.Kind).Render()
	if err != nil {
		return err
	}
	resp.Data, err = pbstruct.Map2pbStruct(schema)
	return err
}

// GetFormSupportedAPIVersions ...
func (h *Handler) GetFormSupportedAPIVersions(
	_ context.Context, req *clusterRes.GetFormSupportedApiVersionsReq, resp *clusterRes.CommonListResp,
) (err error) {
	supportedAPIVersions, ok := renderer.FormRenderSupportedResAPIVersion[req.Kind]
	if !ok {
		return errorx.New(errcode.Unsupported, "资源类型 %s 不支持表单化", req.Kind)
	}
	versions := []map[string]interface{}{
		{"label": "Preferred Version", "value": ""},
	}
	for _, ver := range supportedAPIVersions {
		versions = append(versions, map[string]interface{}{"label": ver, "value": ver})
	}
	resp.Data, err = pbstruct.MapSlice2ListValue(versions)
	return err
}
