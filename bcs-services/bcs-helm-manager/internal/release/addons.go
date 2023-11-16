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

// Package release xxx
package release

import (
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// Addons define cluster add-ons struct,
// cluster add-ons is similar of helm, but more simply.
type Addons struct {
	Name             string
	ChartName        string `yaml:"chartName"`
	Description      string
	DocsLink         string `yaml:"docsLink"`
	Namespace        string
	DefaultValues    string   `yaml:"defaultValues"`
	StopValues       string   `yaml:"stopValues"`
	SupportedActions []string `yaml:"supportedActions"`
}

// AddonsSlice add-ons slice
type AddonsSlice struct {
	Addons []*Addons `yaml:"addons"`
}

// FindByName find add-ons by name
func (a AddonsSlice) FindByName(name string) *Addons {
	for i := range a.Addons {
		if a.Addons[i].Name == name {
			return a.Addons[i]
		}
	}
	return nil
}

// ToAddonsProto trans addons to proto struct
func (a Addons) ToAddonsProto() *helmmanager.Addons {
	return &helmmanager.Addons{
		Name:             &a.Name,
		ChartName:        &a.ChartName,
		Description:      &a.Description,
		Logo:             common.GetStringP(""),
		DocsLink:         &a.DocsLink,
		Version:          common.GetStringP(""),
		CurrentVersion:   common.GetStringP(""),
		Namespace:        &a.Namespace,
		DefaultValues:    &a.DefaultValues,
		CurrentValues:    common.GetStringP(""),
		Status:           common.GetStringP(""),
		Message:          common.GetStringP(""),
		SupportedActions: a.SupportedActions,
		ReleaseName:      common.GetStringP(a.ReleaseName()),
	}
}

// ReleaseName return release name
func (a Addons) ReleaseName() string {
	return strings.ToLower(a.Name)
}

// CanStop check addons can stop
func (a Addons) CanStop() bool {
	for _, v := range a.SupportedActions {
		if v == "stop" {
			return true
		}
	}
	return false
}

// CanUpgrade check addons can upgrade
func (a Addons) CanUpgrade() bool {
	for _, v := range a.SupportedActions {
		if v == "upgrade" {
			return true
		}
	}
	return false
}

// CanConfig check addons can config
func (a Addons) CanConfig() bool {
	for _, v := range a.SupportedActions {
		if v == "config" {
			return true
		}
	}
	return false
}
