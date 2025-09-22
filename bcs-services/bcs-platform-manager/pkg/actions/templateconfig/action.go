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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"

	clustermgr "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs/clustermanager"
)

// TemplateConfigAction templateconfig action interface
type TemplateConfigAction interface { // nolint
	CreateTemplateConfig(ctx context.Context, req *clustermanager.CreateTemplateConfigRequest) (bool, error)
	DeleteTemplateConfig(ctx context.Context, req *types.DeleteTemplateConfigReq) (bool, error)
}

// Action action for templateconfig
type Action struct{}

// NewTemplateConfigActionAction new templateconfig action
func NewTemplateConfigActionAction() TemplateConfigAction {
	return &Action{}
}

// CreateTemplateConfig create templateconfig
func (a *Action) CreateTemplateConfig(ctx context.Context,
	req *clustermanager.CreateTemplateConfigRequest) (bool, error) {
	result, err := clustermgr.CreateTemplateConfig(ctx, req)
	if err != nil {
		return false, err
	}

	return result, nil
}

// DeleteTemplateConfig delete templateconfig
func (a *Action) DeleteTemplateConfig(ctx context.Context, req *types.DeleteTemplateConfigReq) (bool, error) {
	result, err := clustermgr.DeleteTemplateConfig(ctx, &clustermanager.DeleteTemplateConfigRequest{
		TemplateConfigID: req.TemplateConfigID,
		BusinessID:       req.BusinessID,
		ProjectID:        req.ProjectID,
	})
	if err != nil {
		return false, err
	}

	return result, nil
}
