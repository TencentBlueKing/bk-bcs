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
 *
 */

package bkapi

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// BKAPIInstaller is the bk-api helm installer
type BKAPIInstaller struct {
	chartName        string
	isPublicRepo     bool
	releaseName      string
	releaseNamespace string
	projectID        string
	// 已安装应用的 ID，用于更新应用
	appID  int
	client *BCSAppClient
	debug  bool
}

// NewBKAPIInstaller creates a new bk-api helm installer
func NewBKAPIInstaller(projectID, chartName, releaseName, releaseNamespace string,
	isPublicRepo bool, client *BCSAppClient, debug bool) *BKAPIInstaller {
	in := &BKAPIInstaller{
		chartName:        chartName,
		releaseName:      releaseName,
		releaseNamespace: releaseNamespace,
		isPublicRepo:     isPublicRepo,
		projectID:        projectID,
		client:           client,
		debug:            debug,
	}
	return in
}

var _ install.Installer = &BKAPIInstaller{}

// IsInstalled returns whether the app is installed
func (h *BKAPIInstaller) IsInstalled(clusterID string) (bool, error) {
	if h.debug {
		return true, nil
	}
	resp, err := h.client.ListApps(h.projectID, clusterID, h.releaseNamespace, 1, 1000, 0)
	if err != nil {
		blog.Errorf("[BKAPIInstaller] list apps failed, err: %s", err.Error())
		return false, err
	}
	if resp == nil {
		blog.Errorf("[BKAPIInstaller] list apps failed, resp is empty")
		return false, fmt.Errorf("list apps failed, resp is empty")
	}
	if resp.Code != 0 {
		blog.Errorf("[BKAPIInstaller] list apps failed, code: %d, message: %s", resp.Code, resp.Message)
		return false, fmt.Errorf("list apps failed, code: %d, message: %s", resp.Code, resp.Message)
	}
	for _, v := range resp.Data.Results {
		if v.Name == h.releaseName && v.ClusterID == clusterID && v.Namespace == h.releaseNamespace {
			h.appID = v.ID
			return true, nil
		}
	}
	return false, nil
}

// Install installs the app
func (h *BKAPIInstaller) Install(clusterID, values string) error {
	if h.debug {
		return nil
	}
	// get namespace id
	nsID, err := h.getNamespaceID(clusterID)
	if err != nil {
		return err
	}
	// get chart id
	chartID, err := h.getChartID()
	if err != nil {
		return err
	}

	// create app
	req := &CreateAppRequest{
		ProjectID:     h.projectID,
		Answers:       []string{},
		Name:          h.releaseName,
		ClusterID:     clusterID,
		ChartVersion:  chartID,
		NamespaceInfo: nsID,
		ValueFile:     values,
		CmdFlags:      install.DefaultCmdFlag,
	}
	resp, err := h.client.CreateApp(req)
	if err != nil {
		blog.Errorf("[BKAPIInstaller] create app failed, err: %s", err.Error())
		return err
	}
	if resp == nil {
		blog.Errorf("[BKAPIInstaller] create app failed, resp is empty")
		return fmt.Errorf("create app failed, resp is empty")
	}
	if resp.Code != 0 {
		blog.Errorf("[BKAPIInstaller] create app failed, code: %d, message: %s", resp.Code, resp.Message)
		return fmt.Errorf("create app failed, code: %d, message: %s", resp.Code, resp.Message)
	}
	return nil
}
func (h *BKAPIInstaller) getNamespaceID(clusterID string) (int, error) {
	nsList, err := h.client.ListNamespace(h.projectID, clusterID)
	if err != nil {
		blog.Errorf("[BKAPIInstaller] list namespace failed, err: %s", err.Error())
		return 0, err
	}
	if nsList == nil {
		blog.Errorf("[BKAPIInstaller] list namespace failed, resp is empty")
		return 0, fmt.Errorf("list namespace failed, resp is empty")
	}
	if nsList.Code != 0 {
		blog.Errorf("[BKAPIInstaller] list namespace failed, code: %d, message: %s", nsList.Code, nsList.Message)
		return 0, fmt.Errorf("list namespace failed, code: %d, message: %s", nsList.Code, nsList.Message)
	}

	var nsID int
	for _, v := range nsList.Data {
		if v.ClusterID == clusterID && v.Name == h.releaseNamespace {
			blog.Infof("[BKAPIInstaller] list namespace success: %v", h.releaseNamespace)
			nsID = v.ID
		}
	}
	if nsID == 0 {
		// create namespace
		resp, err := h.client.CreateNamespace(h.projectID, clusterID, CreateNamespaceRequest{
			Name: h.releaseNamespace,
		})
		if err != nil {
			blog.Errorf("[BKAPIInstaller] create namespace[%s] failed: %v", h.releaseNamespace, err)
			return 0, fmt.Errorf("create namespace failed: %v", err)
		}

		blog.Infof("[BKAPIInstaller] getNamespaceID CreateNamespace success: %+v", resp.Data)
		nsID = int(resp.Data.ID)

		return nsID, nil
	}

	return nsID, nil
}

func (h *BKAPIInstaller) getChartID() (int, error) {
	charts, err := h.client.ListCharts(h.projectID)
	if err != nil {
		blog.Errorf("[BKAPIInstaller] list charts failed, err: %s", err.Error())
		return 0, err
	}
	if charts == nil {
		blog.Errorf("[BKAPIInstaller] list charts failed, resp is empty")
		return 0, fmt.Errorf("list charts failed, resp is empty")
	}
	if charts.Code != 0 {
		blog.Errorf("[BKAPIInstaller] list charts failed, code: %d, message: %s", charts.Code, charts.Message)
		return 0, fmt.Errorf("list charts failed, code: %d, message: %s", charts.Code, charts.Message)
	}
	for _, v := range charts.Data {
		if h.isPublicRepo && v.Repository.Name != "public-repo" {
			continue
		}
		if v.Name == h.chartName {
			return v.DefaultChartVersion.ID, nil
		}
	}
	blog.Errorf("[BKAPIInstaller] list charts failed, chart %s not found", h.chartName)
	return 0, fmt.Errorf("list charts failed, chart %s not found", h.chartName)
}

// Upgrade upgrades the app
func (h *BKAPIInstaller) Upgrade(clusterID, values string) error {
	if h.debug {
		return nil
	}
	ok, err := h.IsInstalled(clusterID)
	if err != nil {
		blog.Errorf("[BKAPIInstaller] check app installed failed, err: %s", err.Error())
		return err
	}
	if !ok {
		return fmt.Errorf("app %s not installed", h.releaseName)
	}
	// 等待应用正常
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*30)
	defer cancel()
	err = loop.LoopDoFunc(ctx, func() error {
		// get app
		app, errApp := h.client.GetApp(h.projectID, h.appID)
		if errApp != nil {
			blog.Errorf("[BKAPIInstaller] get app failed, err: %s", errApp.Error())
			return errApp
		}
		if app == nil {
			return fmt.Errorf("get app failed, resp is empty")
		}
		if app.Code != 0 {
			return fmt.Errorf("get app failed, code: %d, message: %s, requestID: %s", app.Code, app.Message,
				app.RequestID)
		}
		if !app.Data.TransitioningOn {
			return loop.EndLoop
		}
		blog.Warnf("[BKAPIInstaller] app is on transitioning, waiting, %s", utils.ToJSONString(app.Data))
		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("[BKAPIInstaller] check app installed failed, err: %s", err.Error())
		return err
	}

	// update app
	req := &UpdateAppRequest{
		ProjectID: h.projectID,
		AppID:     h.appID,
		Answers:   []string{},
		// 不更新版本
		UpgradeVersion: -1,
		ValueFile:      values,
		CmdFlags:       install.DefaultCmdFlag,
	}
	resp, err := h.client.UpdateApp(req)
	if err != nil {
		blog.Errorf("[BKAPIInstaller] update app failed, err: %s", err.Error())
		return err
	}
	if resp == nil {
		blog.Errorf("[BKAPIInstaller] update app failed, resp is empty")
		return fmt.Errorf("update app failed, resp is empty")
	}
	if resp.Code != 0 {
		blog.Errorf("[BKAPIInstaller] update app failed, code: %d, message: %s", resp.Code, resp.Message)
		return fmt.Errorf("update app failed, code: %d, message: %s, requestID: %s", resp.Code, resp.Message,
			resp.RequestID)
	}
	return nil
}

// Uninstall uninstalls the app
func (h *BKAPIInstaller) Uninstall(clusterID string) error {
	if h.debug {
		return nil
	}

	// get project cluster appID
	ok, err := h.IsInstalled(clusterID)
	if err != nil {
		blog.Errorf("[BKAPIInstaller] check app installed failed, err: %s", err.Error())
		return err
	}
	if !ok {
		blog.Infof("app %s not installed", h.releaseName)
		return nil
	}
	// delete app
	err = h.client.DeleteApp(&DeleteAppRequest{
		ProjectID: h.projectID,
		AppID:     h.appID,
	})
	if err != nil {
		blog.Errorf("[BKAPIInstaller] delete app failed, err: %s", err.Error())
		return err
	}

	blog.Infof("[BKAPIInstaller] delete app successful[%s:%v]", h.projectID, h.appID)
	return nil
}

// CheckAppStatus check app install status
func (h *BKAPIInstaller) CheckAppStatus(clusterID string, timeout time.Duration, pre bool) (bool, error) {
	if h.debug {
		return true, nil
	}

	// get project cluster appID
	ok, err := h.IsInstalled(clusterID)
	if err != nil {
		blog.Errorf("[BKAPIInstaller] check app installed failed, err: %s", err.Error())
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
		app, errApp := h.client.GetApp(h.projectID, h.appID)
		if errApp != nil {
			blog.Errorf("[BKAPIInstaller] get app failed, err: %s", errApp.Error())
			return errApp
		}
		if app == nil {
			return fmt.Errorf("get app failed, resp is empty")
		}
		if app.Code != 0 {
			return fmt.Errorf("get app failed, code: %d, message: %s, requestID: %s", app.Code, app.Message,
				app.RequestID)
		}
		// 运行中
		if app.Data.TransitioningOn {
			blog.Warnf("[BKAPIInstaller] app is on transitioning, waiting, %s", utils.ToJSONString(app.Data))
			return nil
		}
		// 应用正常
		if app.Data.TransitioningResult {
			return loop.EndLoop
		}
		// 应用异常
		return fmt.Errorf("check app failed, error: %s", app.Data.TransitioningMessage)
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("[BKAPIInstaller] check app installed failed, err: %s", err.Error())
		return false, err
	}

	blog.Infof("[BKAPIInstaller] app install successful[%s:%v]", h.projectID, h.appID)
	return true, nil
}

// Close clean operation
func (h *BKAPIInstaller) Close() {
	return
}
