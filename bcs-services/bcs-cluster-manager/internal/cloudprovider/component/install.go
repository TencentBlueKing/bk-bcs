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

package component

import (
	"fmt"

	cmoptions "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install/bkapi"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install/helm"
)

// ComponentValues get component values interface
type ComponentValues interface {
	// GetValues get component values
	GetValues() (string, error)
}

// InstallOptions options for installer
type InstallOptions struct {
	InstallType string
	ProjectID   string
	// component dependent paras
	// chartName
	ChartName string
	// namespace
	ReleaseNamespace string
	// releaseName
	ReleaseName string
	// public repo
	IsPublicRepo bool
}

// GetComponentInstaller get component installer
func GetComponentInstaller(opts InstallOptions) (install.Installer, error) {
	var (
		installer install.Installer
		err       error
	)
	switch opts.InstallType {
	case bkapi.BcsApp.String():
		client, debug := GetBCSAppClient()
		installer = bkapi.NewBKAPIInstaller(opts.ProjectID, opts.ChartName, opts.ReleaseName, opts.ReleaseNamespace,
			opts.IsPublicRepo, client, debug)
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

// GetBCSAppClient get installer init client
func GetBCSAppClient() (*bkapi.BCSAppClient, bool) {
	op := cmoptions.GetGlobalCMOptions()
	client := bkapi.NewBCSAppClient(
		op.BCSAppConfig.Server,
		op.BCSAppConfig.AppCode,
		op.BCSAppConfig.AppSecret,
		op.BCSAppConfig.BkUserName,
		op.BCSAppConfig.Debug,
	)
	// installer if close by debug
	debug := false
	if !op.BCSAppConfig.Enable {
		debug = true
	}

	return client, debug
}
