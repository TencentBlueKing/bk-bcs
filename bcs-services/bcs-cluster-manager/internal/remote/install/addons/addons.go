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

// Package addons for addon
package addons

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/helmmanager"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install"
)

var (
	// Addons install
	Addons install.InstallerType = "addons"
)

// AddonsInstaller is the addons installer
type AddonsInstaller struct { // nolint
	projectID string
	clusterID string
	addonName string
	debug     bool

	client helmmanager.ClusterAddonsClient
	close  func()
}

// AddonOptions xxx
type AddonOptions struct { // nolint
	ProjectID string
	ClusterID string
	AddonName string
}

// NewAddonsInstaller creates a new addon installer
func NewAddonsInstaller(opts AddonOptions, client *AddonsClient,
	debug bool) (*AddonsInstaller, error) {
	hi := &AddonsInstaller{
		projectID: opts.ProjectID,
		clusterID: opts.ClusterID,
		addonName: opts.AddonName,
		debug:     debug,
	}

	cli, conClose, err := client.GetAddonsClient()
	if err != nil {
		blog.Errorf("NewAddonsInstaller GetAddonsClient failed: %v", err)
		return nil, err
	}
	hi.client = cli
	hi.close = conClose

	return hi, nil
}

var _ install.Installer = &AddonsInstaller{}

// IsInstalled returns whether the app is installed
func (a *AddonsInstaller) IsInstalled(ctx context.Context, clusterID string) (bool, error) {
	if a.debug {
		return true, nil
	}

	resp, err := a.getAddonDetail(ctx, clusterID)
	if err != nil {
		blog.Errorf("[AddonsInstaller] GetAddonsDetail failed, err: %s", err.Error())
		return false, err
	}
	// not found addon
	if resp.Code != nil && *resp.Code != 0 {
		blog.Errorf("[AddonsInstaller] GetAddonsDetail failed, code: %d, message: %s", *resp.Code, *resp.Message)
		return false, nil
	}
	blog.Infof("[AddonsInstaller] [%s:%s] GetAddonsDetail success[%s:%s] status: %s",
		a.projectID, a.clusterID, *resp.Data.Namespace, *resp.Data.Name, *resp.Data.Status)

	if resp.Data.Status == nil || *resp.Data.Status == "" {
		return false, nil
	}

	return true, nil
}

func (a *AddonsInstaller) getAddonDetail(
	ctx context.Context, clusterId string) (*helmmanager.GetAddonsDetailResp, error) {
	start := time.Now()
	resp, err := a.client.GetAddonsDetail(ctx, &helmmanager.GetAddonsDetailReq{
		ProjectCode: &a.projectID,
		ClusterID:   &clusterId,
		Name:        &a.addonName,
	})
	if err != nil {
		metrics.ReportLibRequestMetric("addons", "GetAddonsDetail", "grpc", metrics.LibCallStatusErr, start)
		blog.Errorf("GetAddonsDetail failed, err: %s", err.Error())
		return nil, err
	}
	metrics.ReportLibRequestMetric("addons", "GetAddonsDetail", "grpc", metrics.LibCallStatusOK, start)

	if resp == nil {
		blog.Errorf("[AddonsInstaller] GetAddonsDetail failed, resp is empty")
		return nil, fmt.Errorf("GetAddonsDetail failed, resp is empty")
	}

	return resp, nil
}

// Install installs the app
func (a *AddonsInstaller) Install(ctx context.Context, clusterID, values string) error {
	if a.debug {
		return nil
	}

	addonResp, err := a.getAddonDetail(ctx, clusterID)
	if err != nil || (addonResp.Code != nil && *addonResp.Code != 0) {
		return fmt.Errorf("[AddonsInstaller] InstallAddons failed: %v", err)
	}

	start := time.Now()
	resp, err := a.client.UpgradeAddons(ctx, &helmmanager.UpgradeAddonsReq{
		ProjectCode: &a.projectID,
		ClusterID:   &clusterID,
		Name:        &a.addonName,
		Version:     addonResp.Data.Version,
	})
	if err != nil {
		metrics.ReportLibRequestMetric("addons", "UpgradeAddons", "grpc", metrics.LibCallStatusErr, start)
		blog.Errorf("[AddonsInstaller] InstallAddons failed, err: %s", err.Error())
		return err
	}
	metrics.ReportLibRequestMetric("addons", "UpgradeAddons", "grpc", metrics.LibCallStatusOK, start)

	if resp == nil {
		blog.Errorf("[AddonsInstaller] InstallAddons failed, resp is empty")
		return fmt.Errorf("InstallAddons failed, resp is empty")
	}

	if (resp.Code != nil && *resp.Code != 0) || (resp.Result != nil && !*resp.Result) {
		blog.Errorf("[AddonsInstaller] InstallAddons failed, code: %d, message: %s", resp.Code, *resp.Message)
		return fmt.Errorf("InstallAddons failed, code: %d, message: %s", *resp.Code, *resp.Message)
	}

	blog.Errorf("[AddonsInstaller] InstallAddons[%s:%s] success[%s]", a.projectID, clusterID, a.addonName)

	return nil
}

// Upgrade upgrades the app
func (a *AddonsInstaller) Upgrade(ctx context.Context, clusterID, values string) error {
	return nil
}

// Uninstall uninstalls the app
func (a *AddonsInstaller) Uninstall(ctx context.Context, clusterID string) error {
	if a.debug {
		return nil
	}

	start := time.Now()
	// delete addon
	resp, err := a.client.UninstallAddons(ctx, &helmmanager.UninstallAddonsReq{
		ProjectCode: &a.projectID,
		ClusterID:   &clusterID,
		Name:        &a.addonName,
	})
	if err != nil {
		metrics.ReportLibRequestMetric("addons", "UninstallAddons", "grpc", metrics.LibCallStatusErr, start)
		blog.Errorf("[AddonsInstaller] delete addon failed, err: %s", err.Error())
		return err
	}
	metrics.ReportLibRequestMetric("addons", "UninstallAddons", "grpc", metrics.LibCallStatusOK, start)

	if (resp.Code != nil && *resp.Code != 0) || (resp.Result != nil && !*resp.Result) {
		blog.Errorf("[AddonsInstaller] UninstallAddons failed, code: %d, message: %s", *resp.Code, *resp.Message)
		return fmt.Errorf("UninstallAddons failed, code: %d, message: %s, requestID: %s", *resp.Code,
			*resp.Message, *resp.RequestID)
	}

	blog.Infof("[AddonsInstaller] delete addon successful[%s:%s]", clusterID, a.addonName)
	return nil
}

// CheckAppStatus check app install status
func (a *AddonsInstaller) CheckAppStatus(
	ctx context.Context, clusterID string, timeout time.Duration, pre bool) (bool, error) {
	if a.debug {
		return true, nil
	}
	return true, nil
}

// Close clean operation
func (a *AddonsInstaller) Close() {
	if a.close != nil {
		a.close()
	}
}
