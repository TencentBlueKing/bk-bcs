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

// Package handler xxx
package handler

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/actions/addons"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// ListAddons provide the actions to do list addons
func (ah *AddonsHandler) ListAddons(ctx context.Context,
	req *helmmanager.ListAddonsReq, resp *helmmanager.ListAddonsResp) error {

	defer recorder(ctx, "ListAddons", req, resp)()
	action := addons.NewListAddonsAction(ah.model, *ah.addons, ah.platform, ah.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// GetAddonsDetail provide the actions to do get addons detail
func (ah *AddonsHandler) GetAddonsDetail(ctx context.Context,
	req *helmmanager.GetAddonsDetailReq, resp *helmmanager.GetAddonsDetailResp) error {

	defer recorder(ctx, "GetAddonsDetail", req, resp)()
	action := addons.NewGetAddonsDetailAction(ah.model, *ah.addons, ah.platform, ah.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// InstallAddons provide the actions to do install addons
func (ah *AddonsHandler) InstallAddons(ctx context.Context,
	req *helmmanager.InstallAddonsReq, resp *helmmanager.InstallAddonsResp) error {

	defer recorder(ctx, "InstallAddons", req, resp)()
	action := addons.NewInstallAddonsAction(ah.model, *ah.addons, ah.platform, ah.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// UpgradeAddons provide the actions to do upgrade addons
func (ah *AddonsHandler) UpgradeAddons(ctx context.Context,
	req *helmmanager.UpgradeAddonsReq, resp *helmmanager.UpgradeAddonsResp) error {

	defer recorder(ctx, "UpgradeAddons", req, resp)()
	action := addons.NewUpgradeAddonsAction(ah.model, *ah.addons, ah.platform, ah.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// StopAddons provide the actions to do stop addons
func (ah *AddonsHandler) StopAddons(ctx context.Context,
	req *helmmanager.StopAddonsReq, resp *helmmanager.StopAddonsResp) error {

	defer recorder(ctx, "StopAddons", req, resp)()
	action := addons.NewStopAddonsAction(ah.model, *ah.addons, ah.platform, ah.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// UninstallAddons provide the actions to do uninstall addons
func (ah *AddonsHandler) UninstallAddons(ctx context.Context,
	req *helmmanager.UninstallAddonsReq, resp *helmmanager.UninstallAddonsResp) error {

	defer recorder(ctx, "UninstallAddons", req, resp)()
	action := addons.NewUninstallAddonsAction(ah.model, *ah.addons, ah.platform, ah.releaseHandler)
	return action.Handle(ctx, req, resp)
}
