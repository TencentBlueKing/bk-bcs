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

// Package addon xxx
package addon

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install/addons"
)

// GetAddonInstaller addon installer
func GetAddonInstaller(projectID string, addonName string) (install.Installer, error) {
	return component.GetComponentInstaller(component.InstallOptions{
		InstallType: addons.Addons.String(),
		ProjectID:   projectID,
		AddonName:   addonName,
	})
}
