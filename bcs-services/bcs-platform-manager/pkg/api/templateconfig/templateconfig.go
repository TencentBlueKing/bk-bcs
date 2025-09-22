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

// Package templateconfig templateconfig operate
package templateconfig

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"

	actions "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/actions/templateconfig"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
)

// CreateTemplateConfig 创建TemplateConfig
// @Summary 创建TemplateConfig
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /templateConfig [post]
func CreateTemplateConfig(ctx context.Context, req *clustermanager.CreateTemplateConfigRequest) (*bool, error) {
	result, err := actions.NewTemplateConfigActionAction().CreateTemplateConfig(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteTemplateConfig 删除TemplateConfig
// @Summary 删除TemplateConfig
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /templateConfig/{templateConfigID} [delete]
func DeleteTemplateConfig(ctx context.Context, req *types.DeleteTemplateConfigReq) (*bool, error) {
	result, err := actions.NewTemplateConfigActionAction().DeleteTemplateConfig(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
