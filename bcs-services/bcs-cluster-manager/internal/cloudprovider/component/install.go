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

// Package component xxx
package component

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install/addons"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install/helm"
)

// ComponentValues get component values interface
type ComponentValues interface { // nolint
	// GetValues get component values
	GetValues() (string, error)
}

// InstallOptions options for installer
type InstallOptions struct {
	// InstallType install type
	InstallType string
	// ProjectID project info
	ProjectID string
	// component dependent paras
	// ChartName chartName
	ChartName string
	// ReleaseNamespace namespace
	ReleaseNamespace string
	// ReleaseName releaseName
	ReleaseName string
	// IsPublicRepo public repo
	IsPublicRepo bool
	// AddonName addon name
	AddonName string
}

// GetComponentInstaller get component installer
func GetComponentInstaller(opts InstallOptions) (install.Installer, error) {
	var (
		installer install.Installer
		err       error
	)
	switch opts.InstallType {
	case addons.Addons.String():
		installer, err = addons.NewAddonsInstaller(addons.AddonOptions{
			ProjectID: opts.ProjectID,
			AddonName: opts.AddonName,
		}, addons.GetAddonsClient(), false)
	case helm.Helm.String():
		installer, err = helm.NewHelmInstaller(helm.HelmOptions{
			ProjectID:   opts.ProjectID,
			Namespace:   opts.ReleaseNamespace,
			ReleaseName: opts.ReleaseName,
			ChartName:   opts.ChartName,
			IsPublic:    opts.IsPublicRepo,
		}, helm.GetHelmManagerClient(), false)
	default:
		err = fmt.Errorf("installer not support type[%s]", opts.InstallType)
	}
	if err != nil {
		return nil, err
	}

	return installer, nil
}
