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

package helm

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install"

	k8scorecliset "k8s.io/client-go/kubernetes"
)

var (
	// Helm helmInstall
	Helm install.InstallerType = "helm"
)

// HelmInstaller is the helm installer
type HelmInstaller struct {
	chartRepo        string
	chartName        string
	releaseName      string
	releaseNamespace string
	client           *k8scorecliset.Clientset
}

// NewHelmInstaller creates a new helm installer
func NewHelmInstaller(chartRepo, chartName, releaseName, releaseNamespace string,
	client *k8scorecliset.Clientset) (*HelmInstaller, error) {
	hi := &HelmInstaller{
		chartRepo:        chartRepo,
		chartName:        chartName,
		releaseName:      releaseName,
		releaseNamespace: releaseNamespace,
		client:           client,
	}
	return hi, nil
}

var _ install.Installer = &HelmInstaller{}

// IsInstalled returns whether the app is installed
func (h *HelmInstaller) IsInstalled(clusterID string) (bool, error) {
	return false, nil
}

// Install installs the app
func (h *HelmInstaller) Install(clusterID, values string) error {
	return nil
}

// Upgrade upgrades the app
func (h *HelmInstaller) Upgrade(clusterID, values string) error {
	return nil
}

// Uninstall uninstalls the app
func (h *HelmInstaller) Uninstall(clusterID string) error {
	return nil
}

// CheckAppStatus check app status
func (h *HelmInstaller) CheckAppStatus(clusterID string, timeout time.Duration) (bool, error) {
	return false, nil
}
