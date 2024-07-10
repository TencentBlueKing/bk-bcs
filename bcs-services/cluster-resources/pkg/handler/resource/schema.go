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

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/renderer"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/validator"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// GetMultiResFormSchema xxx
func (h *Handler) GetMultiResFormSchema(
	ctx context.Context, req *clusterRes.GetMultiResFormSchemaReq, resp *clusterRes.CommonListResp,
) error {
	result := make([]map[string]interface{}, 0)
	var err error
	for _, v := range req.GetResourceTypes() {
		schema, ierr := renderer.NewSchemaRenderer(ctx, "", v.ApiVersion, v.Kind, "", "create", true).Render()
		if ierr != nil {
			return ierr
		}
		result = append(result, schema)
	}
	resp.Data, err = pbstruct.MapSlice2ListValue(result)
	return err
}

// GetResFormSchema xxx
func (h *Handler) GetResFormSchema(
	ctx context.Context, req *clusterRes.GetResFormSchemaReq, resp *clusterRes.CommonResp,
) error {
	schema, err := renderer.NewSchemaRenderer(ctx, req.ClusterID, "", req.Kind, req.Namespace, req.Action, false).
		Render()
	if err != nil {
		return err
	}
	resp.Data, err = pbstruct.Map2pbStruct(schema)
	return err
}

// GetFormSupportedAPIVersions xxx
func (h *Handler) GetFormSupportedAPIVersions(
	ctx context.Context, req *clusterRes.GetFormSupportedApiVersionsReq, resp *clusterRes.CommonResp,
) (err error) {
	supportedAPIVersions, ok := validator.FormSupportedResAPIVersion[req.Kind]
	if !ok {
		return errorx.New(errcode.Unsupported, i18n.GetMsg(ctx, "资源类型 `%s` 不支持表单化"), req.Kind)
	}
	versions := []map[string]interface{}{}
	for _, ver := range supportedAPIVersions {
		versions = append(versions, map[string]interface{}{"label": ver, "value": ver})
	}
	resp.Data, err = pbstruct.Map2pbStruct(map[string]interface{}{"selectItems": versions})
	return err
}
