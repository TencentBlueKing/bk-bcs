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
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/utils/pointer"

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

// UninstallIstioComponent 通用的istio组件卸载函数
func UninstallIstioComponent(ctx context.Context, clusterID, componentName, projectCode, meshID string) error {
	resp, err := Uninstall(ctx, &helmmanager.UninstallReleaseV1Req{
		ProjectCode: pointer.String(projectCode),
		ClusterID:   pointer.String(clusterID),
		Name:        pointer.String(componentName),
		Namespace:   pointer.String(common.IstioNamespace),
	})
	if err != nil {
		blog.Errorf("[%s]helm uninstall %s failed, clusterID: %s, err: %s",
			meshID, componentName, clusterID, err)
		return fmt.Errorf("uninstall %s failed: %s", componentName, err)
	}
	if resp.Result != nil && !*resp.Result {
		blog.Errorf("[%s]helm uninstall %s failed, meshID: %s, clusterID: %s, resp message: %s",
			componentName, meshID, clusterID, *resp.Message)
		return fmt.Errorf("uninstall %s failed: %s", componentName, *resp.Message)
	}

	// 查询是否删除成功 查询详情 每隔5s查询一次 直到删除成功，超时2min
	timeout := time.NewTimer(2 * time.Minute)
	defer timeout.Stop()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			blog.Errorf("[%s]uninstall %s timeout, clusterID: %s",
				meshID, componentName, clusterID)
			return fmt.Errorf("uninstall %s timeout for cluster %s", componentName, clusterID)
		case <-ticker.C:
			// 查询 release 是否存在
			detail, err := GetReleaseDetail(ctx, &helmmanager.GetReleaseDetailV1Req{
				ProjectCode: pointer.String(projectCode),
				ClusterID:   pointer.String(clusterID),
				Name:        pointer.String(componentName),
				Namespace:   pointer.String(common.IstioNamespace),
			})
			if err != nil {
				blog.Errorf("[%s]get %s release status failed, clusterID: %s, err: %v",
					meshID, componentName, clusterID, err)
				return fmt.Errorf("get %s release status failed: %v", componentName, err)
			}
			if detail != nil && detail.Message != nil && *detail.Message == driver.ErrReleaseNotFound.Error() {
				return nil
			}
		}
	}
}

// InstallComponentOption istio安装组件参数
type InstallComponentOption struct {
	ChartVersion  string
	ClusterID     string
	ComponentName string
	ChartName     string
	ProjectCode   string
	MeshID        string
	NetworkID     string
	ChartRepo     string
}

// InstallComponent 通用安装istio组件方法
func InstallComponent(
	ctx context.Context,
	opt *InstallComponentOption,
	valuesGenFunc func() (string, error),
) error {
	values, err := valuesGenFunc()
	if err != nil {
		return fmt.Errorf("gen %s values failed: %s", opt.ComponentName, err)
	}
	blog.Infof("install %s values: %s for cluster: %s, mesh: %s, network: %s",
		opt.ComponentName, values, opt.ClusterID, opt.MeshID, opt.NetworkID)

	resp, err := Install(ctx, &helmmanager.InstallReleaseV1Req{
		ProjectCode: pointer.String(opt.ProjectCode),
		ClusterID:   pointer.String(opt.ClusterID),
		Name:        pointer.String(opt.ComponentName),
		Namespace:   pointer.String(common.IstioNamespace),
		Chart:       pointer.String(opt.ChartName),
		Repository:  pointer.String(opt.ChartRepo),
		Version:     pointer.String(opt.ChartVersion),
		Values:      []string{values},
		Args:        []string{"--wait"},
	})
	if err != nil {
		blog.Errorf("install %s failed, err: %s", opt.ComponentName, err)
		return fmt.Errorf("install %s failed: %s", opt.ComponentName, err)
	}
	if resp.Result != nil && !*resp.Result {
		blog.Errorf("install %s failed, err: %s", opt.ComponentName, *resp.Message)
		return fmt.Errorf("install %s failed: %s", opt.ComponentName, *resp.Message)
	}
	// 查询是否安装成功
	timeout := time.NewTimer(2 * time.Minute)
	defer timeout.Stop()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			blog.Errorf("install %s timeout for cluster %s", opt.ComponentName, opt.ClusterID)
			return fmt.Errorf("install %s timeout for cluster %s", opt.ComponentName, opt.ClusterID)
		case <-ticker.C:
			release, err := GetReleaseDetail(ctx, &helmmanager.GetReleaseDetailV1Req{
				ProjectCode: pointer.String(opt.ProjectCode),
				ClusterID:   pointer.String(opt.ClusterID),
				Name:        pointer.String(opt.ComponentName),
				Namespace:   pointer.String(common.IstioNamespace),
			})
			blog.Infof("[loop]get %s release: %+v, err: %s, cluster: %s", opt.ComponentName, release, err, opt.ClusterID)
			if err != nil {
				blog.Errorf("get %s release failed, err: %s", opt.ComponentName, err)
				return fmt.Errorf("get %s release failed: %s", opt.ComponentName, err)
			}
			if release.Data != nil && release.Data.Status != nil {
				if *release.Data.Status == ReleaseStatusDeployed {
					blog.Infof("install %s success for cluster %s", opt.ComponentName, opt.ClusterID)
					return nil
				}
			}
		}
	}
}
