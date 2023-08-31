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

package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

const (
	// AddOnCfsChartName cfs
	AddOnCfsChartName = "CFS"
	// AddOnCbsChartName cbs
	AddOnCbsChartName = "CBS"
	// AddOnCosChartName cos
	AddOnCosChartName = "COS"
)

// Application tke app request values
type Application struct {
	Kind       string          `json:"kind"`
	ApiVersion string          `json:"apiVersion"`
	Spec       ApplicationSpec `json:"spec"`
}

// ApplicationSpec spec
type ApplicationSpec struct {
	Chart  Chart  `json:"chart"`
	Values Values `json:"values"`
}

// Chart tke chart info
type Chart struct {
	ChartName    string `json:"chartName"`
	ChartVersion string `json:"chartVersion"`
}

// Values user define parameters
type Values struct {
	Values []string `json:"values"`
}

// BuildApplicationRequestData build tke app chart request values
func BuildApplicationRequestData(chart Chart, values Values) *Application {
	return &Application{
		Kind:       "App",
		ApiVersion: "application.tkestack.io/v1",
		Spec: ApplicationSpec{
			Chart:  chart,
			Values: values,
		},
	}
}

// GetAppChart get chart version
func GetAppChart(ctx context.Context, cmOption *cloudprovider.CommonOption, addonName string) (*Chart, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	cli, err := api.NewTkeClient(cmOption)
	if err != nil {
		blog.Errorf("GetAppChart[%s] failed: %s", taskID, err)
		return nil, err
	}

	version, err := cli.GetTkeAppChartVersionByName("", strings.ToLower(addonName))
	if err != nil {
		blog.Errorf("GetAppChart[%s] failed: %v", taskID, err)
		return nil, err
	}

	return &Chart{
		ChartName:    strings.ToLower(addonName),
		ChartVersion: version,
	}, nil
}

// AppValuesInterface xxx
type AppValuesInterface interface {
	// GetAppChart get app chart
	GetAppChart(ctx context.Context) (*Chart, error)
	// GetAppValues get app values
	GetAppValues(ctx context.Context) *Values
}

// ApplicationBodyInterface xxx
type ApplicationBodyInterface interface {
	Name() string
	GetApplicationRequestBody(ctx context.Context) (string, error)
}

// AddonStorage for handle different storage by name
type AddonStorage struct {
	CmOption  *cloudprovider.CommonOption
	AddonName string
	RootDir   string
}

// GetApplicationRequestBody get request body
func (cbs *AddonStorage) GetApplicationRequestBody(ctx context.Context) (string, error) {
	chart, err := cbs.GetAppChart(ctx)
	if err != nil {
		return "", err
	}
	values := cbs.GetAppValues(ctx)

	app := BuildApplicationRequestData(*chart, *values)
	appString, err := json.Marshal(app)

	return string(appString), err
}

// GetAppChart get addon chart version
func (cbs *AddonStorage) GetAppChart(ctx context.Context) (*Chart, error) {
	return GetAppChart(ctx, cbs.CmOption, cbs.AddonName)
}

// GetAppValues get addon app values
func (cbs *AddonStorage) GetAppValues(ctx context.Context) *Values {
	params := make([]string, 0)
	params = append(params, cbs.getRootDirParam())

	return &Values{Values: params}
}

// Name addon name
func (cbs *AddonStorage) Name() string {
	return cbs.AddonName
}

// addon user-defined params
func (cbs *AddonStorage) getRootDirParam() string {
	if cbs.RootDir == "" {
		cbs.RootDir = common.KubeletRootDirPath
	}
	return fmt.Sprintf("rootdir=%s", cbs.RootDir)
}

// handleTkeDefaultExtensionAddons handle default addon
func handleTkeDefaultExtensionAddons(ctx context.Context, option *cloudprovider.CommonOption) []api.ExtensionAddon {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	addons := make([]api.ExtensionAddon, 0)

	defaultAddons := make([]ApplicationBodyInterface, 0)
	defaultAddons = append(defaultAddons, &AddonStorage{CmOption: option, AddonName: AddOnCbsChartName},
		&AddonStorage{CmOption: option, AddonName: AddOnCfsChartName},
		&AddonStorage{CmOption: option, AddonName: AddOnCosChartName})

	for i := range defaultAddons {
		body, err := defaultAddons[i].GetApplicationRequestBody(ctx)
		if err != nil || body == "" {
			blog.Errorf("handleTkeDefaultExtensionAddons[%s] addon[%s] failed: %v", taskID,
				defaultAddons[i].Name(), err)
			continue
		}

		blog.Infof("handleTkeDefaultExtensionAddons[%s] addon[%s] successful: %v", taskID,
			defaultAddons[i].Name(), body)
		addons = append(addons, api.ExtensionAddon{
			AddonName:  defaultAddons[i].Name(),
			AddonParam: body,
		})
	}

	return addons
}
