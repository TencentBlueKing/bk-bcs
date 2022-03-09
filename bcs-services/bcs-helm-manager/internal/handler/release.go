/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package handler

import (
	"context"

	actionRelease "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/actions/release"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// ListRelease provide the actions to do list release
func (hm *HelmManager) ListRelease(ctx context.Context,
	req *helmmanager.ListReleaseReq, resp *helmmanager.ListReleaseResp) error {

	defer recorder(ctx, "ListRelease", req, resp)()
	action := actionRelease.NewListReleaseAction(hm.model, hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// GetReleaseDetail provide the actions to do get release detail
func (hm *HelmManager) GetReleaseDetail(ctx context.Context,
	req *helmmanager.GetReleaseDetailReq, resp *helmmanager.GetReleaseDetailResp) error {

	defer recorder(ctx, "GetReleaseDetail", req, resp)()
	action := actionRelease.NewGetReleaseDetailAction(hm.model, hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// InstallRelease provide the actions to do install release
func (hm *HelmManager) InstallRelease(ctx context.Context,
	req *helmmanager.InstallReleaseReq, resp *helmmanager.InstallReleaseResp) error {

	defer recorder(ctx, "InstallRelease", req, resp)()
	action := actionRelease.NewInstallReleaseAction(hm.model, hm.platform, hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// UninstallRelease provide the actions to do uninstall release
func (hm *HelmManager) UninstallRelease(ctx context.Context,
	req *helmmanager.UninstallReleaseReq, resp *helmmanager.UninstallReleaseResp) error {

	defer recorder(ctx, "UninstallRelease", req, resp)()
	action := actionRelease.NewUninstallReleaseAction(hm.model, hm.platform, hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// UpgradeRelease provide the actions to do upgrade release
func (hm *HelmManager) UpgradeRelease(ctx context.Context,
	req *helmmanager.UpgradeReleaseReq, resp *helmmanager.UpgradeReleaseResp) error {

	defer recorder(ctx, "UpgradeRelease", req, resp)()
	action := actionRelease.NewUpgradeReleaseAction(hm.model, hm.platform, hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// RollbackRelease provide the actions to do rollback release
func (hm *HelmManager) RollbackRelease(ctx context.Context,
	req *helmmanager.RollbackReleaseReq, resp *helmmanager.RollbackReleaseResp) error {

	defer recorder(ctx, "RollbackRelease", req, resp)()
	action := actionRelease.NewRollbackReleaseAction(hm.model, hm.platform, hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}
