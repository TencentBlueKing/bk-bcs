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

package handler

import (
	"context"

	actionRelease "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/actions/release"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// ListReleaseV1 provide the actions to do list release
func (hm *HelmManager) ListReleaseV1(ctx context.Context,
	req *helmmanager.ListReleaseV1Req, resp *helmmanager.ListReleaseV1Resp) error {

	defer recorder(ctx, "ListReleaseV1", req, resp)()
	action := actionRelease.NewListReleaseV1Action(hm.model, hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// ListReleaseV2 provide the actions to do list release
func (hm *HelmManager) ListReleaseV2(ctx context.Context,
	req *helmmanager.ListReleaseV1Req, resp *helmmanager.ListReleaseV1Resp) error {

	defer recorder(ctx, "ListReleaseV2", req, resp)()
	action := actionRelease.NewListReleaseV2Action(hm.model, hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// GetReleaseDetailV1 provide the actions to do get release detail
func (hm *HelmManager) GetReleaseDetailV1(ctx context.Context,
	req *helmmanager.GetReleaseDetailV1Req, resp *helmmanager.GetReleaseDetailV1Resp) error {

	defer recorder(ctx, "GetReleaseDetailV1", req, resp)()
	action := actionRelease.NewGetReleaseDetailV1Action(hm.model, hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// InstallReleaseV1 provide the actions to do install release
func (hm *HelmManager) InstallReleaseV1(ctx context.Context,
	req *helmmanager.InstallReleaseV1Req, resp *helmmanager.InstallReleaseV1Resp) error {

	defer recorder(ctx, "InstallReleaseV1", req, resp)()
	action := actionRelease.NewInstallReleaseV1Action(hm.model, hm.platform, hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// UninstallReleaseV1 provide the actions to do uninstall release
func (hm *HelmManager) UninstallReleaseV1(ctx context.Context,
	req *helmmanager.UninstallReleaseV1Req, resp *helmmanager.UninstallReleaseV1Resp) error {

	defer recorder(ctx, "UninstallReleaseV1", req, resp)()
	action := actionRelease.NewUninstallReleaseV1Action(hm.model, hm.platform, hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// UpgradeReleaseV1 provide the actions to do upgrade release
func (hm *HelmManager) UpgradeReleaseV1(ctx context.Context,
	req *helmmanager.UpgradeReleaseV1Req, resp *helmmanager.UpgradeReleaseV1Resp) error {

	defer recorder(ctx, "UpgradeReleaseV1", req, resp)()
	action := actionRelease.NewUpgradeReleaseV1Action(hm.model, hm.platform, hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// RollbackReleaseV1 provide the actions to do rollback release
func (hm *HelmManager) RollbackReleaseV1(ctx context.Context,
	req *helmmanager.RollbackReleaseV1Req, resp *helmmanager.RollbackReleaseV1Resp) error {

	defer recorder(ctx, "RollbackReleaseV1", req, resp)()
	action := actionRelease.NewRollbackReleaseV1Action(hm.model, hm.platform, hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// ReleasePreview provide the actions to get release preview
func (hm *HelmManager) ReleasePreview(ctx context.Context,
	req *helmmanager.ReleasePreviewReq, resp *helmmanager.ReleasePreviewResp) error {

	defer recorder(ctx, "ReleasePreview", req, resp)()
	action := actionRelease.NewReleasePreviewAction(hm.model, hm.platform, hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// GetReleaseHistory provide the actions to do get release history
func (hm *HelmManager) GetReleaseHistory(ctx context.Context,
	req *helmmanager.GetReleaseHistoryReq, resp *helmmanager.GetReleaseHistoryResp) error {

	defer recorder(ctx, "GetReleaseHistory", req, resp)()
	action := actionRelease.NewGetReleaseHistoryAction(hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// GetReleaseManifest provide the actions to do get release manifest
func (hm *HelmManager) GetReleaseManifest(ctx context.Context,
	req *helmmanager.GetReleaseManifestReq, resp *helmmanager.GetReleaseManifestResp) error {

	defer recorder(ctx, "GetReleaseManifest", req, resp)()
	action := actionRelease.NewGetReleaseManifestAction(hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// GetReleaseStatus provide the actions to do get release status
func (hm *HelmManager) GetReleaseStatus(ctx context.Context,
	req *helmmanager.GetReleaseStatusReq, resp *helmmanager.CommonListResp) error {

	defer recorder(ctx, "GetReleaseStatus", req, resp)()
	action := actionRelease.NewGetReleaseStatusAction(hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// GetReleaseDetailExtend provide the actions to do get release detail extend
func (hm *HelmManager) GetReleaseDetailExtend(
	ctx context.Context, req *helmmanager.GetReleaseDetailExtendReq, resp *helmmanager.CommonResp) error {
	defer recorder(ctx, "GetReleaseDetailExtend", req, resp)()
	action := actionRelease.NewGetReleaseExtendAction(hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// GetReleasePods provide the actions to do get release pods
func (hm *HelmManager) GetReleasePods(ctx context.Context,
	req *helmmanager.GetReleasePodsReq, resp *helmmanager.CommonListResp) error {

	defer recorder(ctx, "GetReleasePods", req, resp)()
	action := actionRelease.NewGetReleasePodsAction(hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}

// ImportClusterRelease provide the actions to import cluster releases
func (hm *HelmManager) ImportClusterRelease(ctx context.Context,
	req *helmmanager.ImportClusterReleaseReq, resp *helmmanager.ImportClusterReleaseResp) error {

	defer recorder(ctx, "ImportClusterRelease", req, nil)()
	action := actionRelease.NewImportClusterReleaseAction(hm.model, hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}
