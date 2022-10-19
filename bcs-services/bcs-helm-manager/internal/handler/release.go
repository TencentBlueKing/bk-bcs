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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// ListRelease provide the actions to do list release
func (hm *HelmManager) ListRelease(ctx context.Context,
	req *helmmanager.ListReleaseReq, resp *helmmanager.ListReleaseResp) error {

	defer recorder(ctx, "ListRelease", req, resp)()
	action := actionRelease.NewListReleaseAction(hm.releaseHandler)
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

// ListReleaseV1 provide the actions to do list release
func (hm *HelmManager) ListReleaseV1(ctx context.Context,
	req *helmmanager.ListReleaseV1Req, resp *helmmanager.ListReleaseV1Resp) error {

	defer recorder(ctx, "ListReleaseV1", req, resp)()
	action := actionRelease.NewListReleaseAction(hm.releaseHandler)

	listReq := &helmmanager.ListReleaseReq{
		ClusterID: req.ClusterID,
		Namespace: req.Namespace,
		Name:      req.Name,
		Page:      req.Page,
		Size:      req.Size,
	}
	listResp := &helmmanager.ListReleaseResp{}
	err := action.Handle(ctx, listReq, listResp)
	resp.Code = listResp.Code
	resp.Message = listResp.Message
	resp.Result = listResp.Result
	resp.Data = listResp.Data
	return err
}

// GetReleaseDetailV1 provide the actions to do get release detail
func (hm *HelmManager) GetReleaseDetailV1(ctx context.Context,
	req *helmmanager.GetReleaseDetailV1Req, resp *helmmanager.GetReleaseDetailV1Resp) error {

	defer recorder(ctx, "GetReleaseDetailV1", req, resp)()
	action := actionRelease.NewGetReleaseDetailAction(hm.model, hm.releaseHandler)

	getReq := &helmmanager.GetReleaseDetailReq{
		ClusterID: req.ClusterID,
		Namespace: req.Namespace,
		Name:      req.Name,
	}
	getResp := &helmmanager.GetReleaseDetailResp{}
	err := action.Handle(ctx, getReq, getResp)
	resp.Code = getResp.Code
	resp.Message = getResp.Message
	resp.Result = getResp.Result
	resp.Data = getResp.Data
	return err
}

// InstallReleaseV1 provide the actions to do install release
func (hm *HelmManager) InstallReleaseV1(ctx context.Context,
	req *helmmanager.InstallReleaseV1Req, resp *helmmanager.InstallReleaseV1Resp) error {

	defer recorder(ctx, "InstallReleaseV1", req, resp)()
	action := actionRelease.NewInstallReleaseAction(hm.model, hm.platform, hm.releaseHandler)

	installReq := &helmmanager.InstallReleaseReq{
		Name:       req.Name,
		Namespace:  req.Namespace,
		ClusterID:  req.ClusterID,
		ProjectID:  req.ProjectCode,
		Repository: req.Repository,
		Chart:      req.Chart,
		Version:    req.Version,
		Operator:   common.GetStringP(auth.GetUserFromCtx(ctx)),
		Values:     req.Values,
		Args:       req.Args,
	}
	installResp := &helmmanager.InstallReleaseResp{}
	err := action.Handle(ctx, installReq, installResp)
	resp.Code = installResp.Code
	resp.Message = installResp.Message
	resp.Result = installResp.Result
	return err
}

// UninstallReleaseV1 provide the actions to do uninstall release
func (hm *HelmManager) UninstallReleaseV1(ctx context.Context,
	req *helmmanager.UninstallReleaseV1Req, resp *helmmanager.UninstallReleaseV1Resp) error {

	defer recorder(ctx, "UninstallReleaseV1", req, resp)()
	action := actionRelease.NewUninstallReleaseAction(hm.model, hm.platform, hm.releaseHandler)

	uninstallReq := &helmmanager.UninstallReleaseReq{
		ClusterID: req.ClusterID,
		Name:      req.Name,
		Namespace: req.Namespace,
		Operator:  common.GetStringP(auth.GetUserFromCtx(ctx)),
	}
	uninstallResp := &helmmanager.UninstallReleaseResp{}
	err := action.Handle(ctx, uninstallReq, uninstallResp)
	resp.Code = uninstallResp.Code
	resp.Message = uninstallResp.Message
	resp.Result = uninstallResp.Result
	return err
}

// UpgradeReleaseV1 provide the actions to do upgrade release
func (hm *HelmManager) UpgradeReleaseV1(ctx context.Context,
	req *helmmanager.UpgradeReleaseV1Req, resp *helmmanager.UpgradeReleaseV1Resp) error {

	defer recorder(ctx, "UpgradeReleaseV1", req, resp)()
	action := actionRelease.NewUpgradeReleaseAction(hm.model, hm.platform, hm.releaseHandler)

	upgradeReq := &helmmanager.UpgradeReleaseReq{
		Name:       req.Name,
		Namespace:  req.Namespace,
		ClusterID:  req.ClusterID,
		ProjectID:  req.ProjectCode,
		Repository: req.Repository,
		Chart:      req.Chart,
		Version:    req.Version,
		Operator:   common.GetStringP(auth.GetUserFromCtx(ctx)),
		Values:     req.Values,
		Args:       req.Args,
	}
	upgradeResp := &helmmanager.UpgradeReleaseResp{}
	err := action.Handle(ctx, upgradeReq, upgradeResp)
	resp.Code = upgradeResp.Code
	resp.Message = upgradeResp.Message
	resp.Result = upgradeResp.Result
	return err
}

// RollbackReleaseV1 provide the actions to do rollback release
func (hm *HelmManager) RollbackReleaseV1(ctx context.Context,
	req *helmmanager.RollbackReleaseV1Req, resp *helmmanager.RollbackReleaseV1Resp) error {

	defer recorder(ctx, "RollbackReleaseV1", req, resp)()
	action := actionRelease.NewRollbackReleaseAction(hm.model, hm.platform, hm.releaseHandler)

	rollbackReq := &helmmanager.RollbackReleaseReq{
		ClusterID: req.ClusterID,
		Namespace: req.Namespace,
		Name:      req.Name,
		Revision:  req.Revision,
		Operator:  common.GetStringP(auth.GetUserFromCtx(ctx)),
	}
	rollbackResp := &helmmanager.RollbackReleaseResp{}
	err := action.Handle(ctx, rollbackReq, rollbackResp)
	resp.Code = rollbackResp.Code
	resp.Message = rollbackResp.Message
	resp.Result = rollbackResp.Result
	return err
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

// GetReleaseStatus provide the actions to do get release status
func (hm *HelmManager) GetReleaseStatus(ctx context.Context,
	req *helmmanager.GetReleaseStatusReq, resp *helmmanager.CommonListResp) error {

	defer recorder(ctx, "GetReleaseStatus", req, resp)()
	action := actionRelease.NewGetReleaseStatusAction(hm.releaseHandler)
	return action.Handle(ctx, req, resp)
}
