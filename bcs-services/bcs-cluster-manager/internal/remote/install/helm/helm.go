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

// Package helm for helm
package helm

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/helmmanager"
	"github.com/avast/retry-go"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var (
	// Helm helmInstall
	Helm install.InstallerType = "helm"
)

// HelmInstaller is the helm installer
type HelmInstaller struct { // nolint
	projectID        string
	clusterID        string
	releaseNamespace string
	releaseName      string

	chartName    string
	isPublicRepo bool
	repo         string
	debug        bool

	client helmmanager.HelmManagerClient
	close  func()
}

// HelmOptions xxx
type HelmOptions struct { // nolint
	ProjectID   string
	ClusterID   string
	Namespace   string
	ReleaseName string
	ChartName   string
	IsPublic    bool
}

// NewHelmInstaller creates a new helm installer
func NewHelmInstaller(opts HelmOptions, client *HelmClient,
	debug bool) (*HelmInstaller, error) {
	hi := &HelmInstaller{
		projectID:        opts.ProjectID,
		clusterID:        opts.ClusterID,
		releaseNamespace: opts.Namespace,
		releaseName:      opts.ReleaseName,
		chartName:        opts.ChartName,
		isPublicRepo:     opts.IsPublic,
		debug:            debug,
	}

	cli, conClose, err := client.GetHelmManagerClient()
	if err != nil {
		blog.Errorf("NewHelmInstaller GetHelmManagerClient failed: %v", err)
		return nil, err
	}
	hi.client = cli
	hi.close = conClose

	return hi, nil
}

var _ install.Installer = &HelmInstaller{}

// IsInstalled returns whether the app is installed
func (h *HelmInstaller) IsInstalled(clusterID string) (bool, error) {
	if h.debug {
		return true, nil
	}

	resp, err := h.client.GetReleaseDetailV1(context.Background(), &helmmanager.GetReleaseDetailV1Req{
		ProjectCode: h.projectID,
		ClusterID:   clusterID,
		Namespace:   h.releaseNamespace,
		Name:        h.releaseName,
	})
	if err != nil {
		blog.Errorf("[HelmInstaller] GetReleaseDetail failed, err: %s", err.Error())
		return false, err
	}
	if resp == nil {
		blog.Errorf("[HelmInstaller] GetReleaseDetail failed, resp is empty")
		return false, fmt.Errorf("GetReleaseDetail failed, resp is empty")
	}
	// not found release
	if resp.Code != 0 {
		blog.Errorf("[HelmInstaller] GetReleaseDetail failed, code: %d, message: %s", resp.Code, resp.Message)
		return false, nil
	}

	blog.Infof("[HelmInstaller] [%s:%s] GetReleaseDetail success[%s:%s] status: %s",
		resp.Data.Chart, resp.Data.ChartVersion, resp.Data.Namespace, resp.Data.Name, resp.Data.Status)

	return true, nil
}

func (h *HelmInstaller) getChartLatestVersion(project string, repo, chart string) (string, error) {
	resp, err := h.client.GetChartDetailV1(context.Background(), &helmmanager.GetChartDetailV1Req{
		ProjectCode: project,
		RepoName:    repo,
		Name:        chart,
	})
	if err != nil {
		blog.Errorf("[HelmInstaller] getChartLatestVersion failed: %v", err)
		return "", err
	}

	if resp.Code != 0 || !resp.Result {
		blog.Errorf("[HelmInstaller] getChartLatestVersion[%s] failed: %v", resp.RequestID, resp.Message)
		return "", err
	}

	return resp.Data.LatestVersion, nil
}

func (h *HelmInstaller) setRepo() {
	// default use public-repo
	if h.isPublicRepo || h.repo == "" {
		h.repo = PubicRepo
	}
}

// Install installs the app
func (h *HelmInstaller) Install(clusterID, values string) error {
	if h.debug {
		return nil
	}

	h.setRepo()
	// get chart latest version
	version, err := h.getChartLatestVersion(h.projectID, h.repo, h.chartName)
	if err != nil {
		blog.Errorf("[HelmInstaller] getChartLatestVersion failed: %v", err)
		return err
	}

	// create app
	req := &helmmanager.InstallReleaseV1Req{
		ProjectCode: h.projectID,
		ClusterID:   clusterID,
		Namespace:   h.releaseNamespace,
		Name:        h.releaseName,
		Repository:  h.repo,
		Chart:       h.chartName,
		Version:     version,
		Values:      []string{values},
		Args:        install.DefaultArgsFlag,
	}

	resp := &helmmanager.InstallReleaseV1Resp{}
	err = retry.Do(func() error {
		resp, err = h.client.InstallReleaseV1(context.Background(), req)
		if err != nil {
			blog.Errorf("[HelmInstaller] InstallRelease failed, err: %s", err.Error())
			return err
		}
		if resp == nil {
			blog.Errorf("[HelmInstaller] InstallRelease failed, resp is empty")
			return fmt.Errorf("InstallRelease failed, resp is empty")
		}

		if resp.Code != 0 || !resp.Result {
			blog.Errorf("[HelmInstaller] InstallRelease failed, code: %d, message: %s", resp.Code, resp.Message)
			return fmt.Errorf("InstallRelease failed, code: %d, message: %s", resp.Code, resp.Message)
		}

		return nil
	}, retry.Attempts(retryCount), retry.Delay(defaultTimeOut), retry.DelayType(retry.FixedDelay))
	if err != nil {
		return fmt.Errorf("call api HelmInstaller InstallRelease failed: %v, resp: %s", err, utils.ToJSONString(resp))
	}

	return nil
}

// Upgrade upgrades the app
func (h *HelmInstaller) Upgrade(clusterID, values string) error {
	if h.debug {
		return nil
	}

	// upgrade need app status deployed
	ok, err := h.CheckAppStatus(clusterID, time.Minute*10, true)
	if err != nil {
		blog.Errorf("[HelmInstaller] Upgrade CheckAppStatus failed: %v", err)
		return err
	}
	if !ok {
		return fmt.Errorf("[HelmInstaller] Upgrade release %s status acnormal", h.releaseName)
	}

	h.setRepo()
	// get chart latest version
	/*
		version, err := h.getChartLatestVersion(h.projectID, h.repo, h.chartName)
		if err != nil {
			blog.Errorf("[HelmInstaller] getChartLatestVersion failed: %v", err)
			return err
		}
	*/

	// update app: default not update chart version
	req := &helmmanager.UpgradeReleaseV1Req{
		ProjectCode: h.projectID,
		ClusterID:   clusterID,
		Namespace:   h.releaseNamespace,
		Name:        h.releaseName,
		Repository:  h.repo,
		Chart:       h.chartName,
		//Version:     version,
		Values: []string{values},
		Args:   install.DefaultArgsFlag,
	}

	resp, err := h.client.UpgradeReleaseV1(context.Background(), req)
	if err != nil {
		blog.Errorf("[HelmInstaller] UpgradeRelease failed, err: %s", err.Error())
		return err
	}
	if resp == nil {
		blog.Errorf("[HelmInstaller] UpgradeRelease failed, resp is empty")
		return fmt.Errorf("UpgradeRelease failed, resp is empty")
	}
	if resp.Code != 0 {
		blog.Errorf("[HelmInstaller] UpgradeRelease failed, code: %d, message: %s", resp.Code, resp.Message)
		return fmt.Errorf("UpgradeRelease failed, code: %d, message: %s, requestID: %s", resp.Code, resp.Message,
			resp.RequestID)
	}

	return nil
}

// Uninstall uninstalls the app
func (h *HelmInstaller) Uninstall(clusterID string) error {
	if h.debug {
		return nil
	}

	// get project cluster release
	ok, err := h.IsInstalled(clusterID)
	if err != nil {
		blog.Errorf("[HelmInstaller] check app installed failed, err: %s", err.Error())
		return err
	}
	if !ok {
		blog.Infof("app %s not installed", h.releaseName)
		return nil
	}

	// delete app
	resp, err := h.client.UninstallReleaseV1(context.Background(), &helmmanager.UninstallReleaseV1Req{
		ProjectCode: h.projectID,
		Name:        h.releaseName,
		Namespace:   h.releaseNamespace,
		ClusterID:   clusterID,
	})
	if err != nil {
		blog.Errorf("[HelmInstaller] delete app failed, err: %s", err.Error())
		return err
	}
	if resp.Code != 0 {
		blog.Errorf("[HelmInstaller] UninstallRelease failed, code: %d, message: %s", resp.Code, resp.Message)
		return fmt.Errorf("UninstallRelease failed, code: %d, message: %s, requestID: %s", resp.Code, resp.Message,
			resp.RequestID)
	}

	blog.Infof("[HelmInstaller] delete app successful[%s:%s:%v]", clusterID, h.releaseNamespace, h.releaseName)
	return nil
}

// CheckAppStatus check app install status
func (h *HelmInstaller) CheckAppStatus(clusterID string, timeout time.Duration, pre bool) (bool, error) {
	if h.debug {
		return true, nil
	}

	// get project cluster appID
	ok, err := h.IsInstalled(clusterID)
	if err != nil {
		blog.Errorf("[HelmInstaller] check app installed failed, err: %s", err.Error())
		return false, err
	}
	if !ok {
		blog.Errorf("app %s not installed", h.releaseName)
		return false, fmt.Errorf("app %s not installed", h.releaseName)
	}

	// 等待应用正常
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err = loop.LoopDoFunc(ctx, func() error {
		// get app
		resp, err := h.client.GetReleaseDetailV1(ctx, &helmmanager.GetReleaseDetailV1Req{ // nolint
			ProjectCode: h.projectID,
			ClusterID:   clusterID,
			Namespace:   h.releaseNamespace,
			Name:        h.releaseName,
		})
		if err != nil {
			blog.Errorf("[HelmInstaller] GetReleaseDetail failed, err: %s", err.Error())
			return err
		}
		if resp == nil {
			return fmt.Errorf("[HelmInstaller] GetReleaseDetail failed, resp is empty")
		}
		if resp.Code != 0 {
			return fmt.Errorf("[HelmInstaller] GetReleaseDetail failed, code: %d, message: %s, requestID: %s",
				resp.Code, resp.Message, resp.RequestID)
		}

		blog.Infof("[HelmInstaller] GetReleaseDetail status: %s", resp.Data.Status)

		// 前置检查
		if pre {
			switch resp.Data.Status {
			case DeployedInstall, DeployedRollback, DeployedUpgrade, FailedInstall,
				FailedRollback, FailedUpgrade, FailedState, FailedUninstall:
				return loop.EndLoop
			default:
			}

			blog.Warnf("[HelmInstaller] GetReleaseDetail[%v] is on transitioning, waiting, %s", pre,
				utils.ToJSONString(resp.Data))
			return nil
		}

		// 后置检查

		// 成功状态 / 失败状态 则终止
		switch resp.Data.Status {
		case DeployedInstall, DeployedRollback, DeployedUpgrade:
			return loop.EndLoop
		case FailedInstall, FailedRollback, FailedUpgrade, FailedState:
			return fmt.Errorf("[HelmInstaller] CheckAppStatus[%s] failed: %s", resp.RequestID, resp.Data.Status)
		default:
		}

		blog.Warnf("[HelmInstaller] GetReleaseDetail[%v] is on transitioning, waiting, %s", pre,
			utils.ToJSONString(resp.Data))
		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("[HelmInstaller] GetReleaseDetail installed failed, err: %s", err.Error())
		return false, err
	}

	blog.Infof("[HelmInstaller] app install successful[%s:%s:%v]", clusterID, h.releaseNamespace, h.releaseName)
	return true, nil
}

// Close clean operation
func (h *HelmInstaller) Close() {
	if h.close != nil {
		h.close()
	}
}
