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

// Package helm 获取helm manager client
package helm

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/helmmanager"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
)

const (
	// ReleaseStatusDeployed 已部署
	ReleaseStatusDeployed = "deployed"
	// ReleaseStatusUninstalled 已卸载
	ReleaseStatusUninstalled = "uninstalled"
	// ReleaseStatusSuperseded 已升级
	ReleaseStatusSuperseded = "superseded"
	// ReleaseStatusFailed 安装失败
	ReleaseStatusFailed = "failed"
	// ReleaseStatusUninstalling 卸载中
	ReleaseStatusUninstalling = "uninstalling"
	// ReleaseStatusPendingInstall 安装中
	ReleaseStatusPendingInstall = "pending-install"
	// ReleaseStatusPendingUpgrade 升级中
	ReleaseStatusPendingUpgrade = "pending-upgrade"
	// ReleaseStatusPendingRollback 回滚中
	ReleaseStatusPendingRollback = "pending-rollback"
	// ReleaseStatusFailedInstall 安装失败
	ReleaseStatusFailedInstall = "failed-install"
	// ReleaseStatusFailedUpgrade 升级失败
	ReleaseStatusFailedUpgrade = "failed-upgrade"
	// ReleaseStatusFailedRollback 回滚失败
	ReleaseStatusFailedRollback = "failed-rollback"
	// ReleaseStatusFailedUninstall 卸载失败
	ReleaseStatusFailedUninstall = "failed-uninstall"
	// ReleaseStatusUnknown 未知
	ReleaseStatusUnknown = "unknown"
)

// GetClient 获取helm manager client
func GetClient() (*helmmanager.HelmClientWrapper, func(), error) {
	helmManagerClient, closeFunc, err := helmmanager.GetClient(common.ServiceDomain)
	if err != nil {
		return nil, nil, err
	}
	return helmManagerClient, closeFunc, nil
}

// Install 安装helm chart（异步）
func Install(ctx context.Context, req *helmmanager.InstallReleaseV1Req) (*helmmanager.InstallReleaseV1Resp, error) {
	helmManagerClient, closeFunc, err := GetClient()
	if err != nil {
		return nil, err
	}
	defer closeFunc()

	return helmManagerClient.HelmManagerClient.InstallReleaseV1(ctx, req)
}

// Upgrade 升级helm chart（异步）
func Upgrade(ctx context.Context, req *helmmanager.UpgradeReleaseV1Req) (*helmmanager.UpgradeReleaseV1Resp, error) {
	helmManagerClient, closeFunc, err := GetClient()
	if err != nil {
		return nil, err
	}
	defer closeFunc()

	return helmManagerClient.HelmManagerClient.UpgradeReleaseV1(ctx, req)
}

// SyncInstallOrUpgrade 同步安装或升级
func SyncInstallOrUpgrade(
	ctx context.Context,
	req *helmmanager.UpgradeReleaseV1Req,
) (*helmmanager.GetReleaseDetailV1Resp, error) {
	// 安装或升级
	resp, err := Upgrade(ctx, req)
	if err != nil {
		return nil, err
	}
	blog.Infof("install or upgrade resp: %+v ,req: %+v", resp, req)
	if resp != nil && resp.Code != nil && *resp.Code != common.SuccessCode {
		return nil, fmt.Errorf("install or upgrade failed, code: %d, message: %s", *resp.Code, *resp.Message)
	}
	// 获取详情状态，如果是进行中的状态则等待2s再重试一次
	for i := 0; i < 2; i++ {
		blog.Infof("get release detail, req: %+v", req)
		detail, err := GetReleaseDetail(ctx, &helmmanager.GetReleaseDetailV1Req{
			ProjectCode: req.ProjectCode,
			ClusterID:   req.ClusterID,
			Namespace:   req.Namespace,
			Name:        req.Name,
		})
		if err != nil {
			blog.Errorf("get release detail failed, err: %v", err)
			return nil, fmt.Errorf("get release detail failed, err: %v", err)
		}
		blog.Infof("get release detail, resp: %+v,req: %+v", detail, req)
		// 部署成功
		if detail.Data != nil && *detail.Data.Status == ReleaseStatusDeployed {
			return detail, nil
		}
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("install or upgrade failed")
}

// Uninstall 卸载helm chart（异步）
func Uninstall(
	ctx context.Context,
	req *helmmanager.UninstallReleaseV1Req) (*helmmanager.UninstallReleaseV1Resp, error) {
	helmManagerClient, closeFunc, err := GetClient()
	if err != nil {
		return nil, err
	}
	defer closeFunc()

	return helmManagerClient.HelmManagerClient.UninstallReleaseV1(ctx, req)
}

// GetReleaseDetail 获取helm chart详情
func GetReleaseDetail(
	ctx context.Context,
	req *helmmanager.GetReleaseDetailV1Req,
) (*helmmanager.GetReleaseDetailV1Resp, error) {
	helmManagerClient, closeFunc, err := GetClient()
	if err != nil {
		return nil, err
	}
	defer closeFunc()

	return helmManagerClient.HelmManagerClient.GetReleaseDetailV1(ctx, req)
}
