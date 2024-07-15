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

package auth

import (
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
)

// ActionPermissions action 对应权限中心的权限
var ActionPermissions = map[string]string{
	// Repo
	"HelmManager.CreateRepository":   project.CanViewProjectOperation,
	"HelmManager.CreatePersonalRepo": project.CanViewProjectOperation,
	"HelmManager.UpdateRepository":   project.CanEditProjectOperation,
	"HelmManager.GetRepository":      project.CanViewProjectOperation,
	"HelmManager.DeleteRepository":   project.CanEditProjectOperation,
	"HelmManager.ListRepository":     project.CanViewProjectOperation,

	// Chart
	"HelmManager.ListChartV1":        project.CanViewProjectOperation,
	"HelmManager.GetChartDetailV1":   project.CanViewProjectOperation,
	"HelmManager.ListChartVersionV1": project.CanViewProjectOperation,
	"HelmManager.GetVersionDetailV1": project.CanViewProjectOperation,
	"HelmManager.DeleteChart":        project.CanEditProjectOperation,
	"HelmManager.DeleteChartVersion": project.CanEditProjectOperation,
	"HelmManager.DownloadChart":      project.CanViewProjectOperation,
	"HelmManager.UploadChart":        project.CanViewProjectOperation,
	"HelmManager.GetChartRelease":    project.CanViewProjectOperation,

	// Release
	"HelmManager.ListReleaseV1":          project.CanViewProjectOperation,
	"HelmManager.ListReleaseV2":          project.CanViewProjectOperation,
	"HelmManager.GetReleaseDetailV1":     namespace.CanViewNamespaceScopedResourceOperation,
	"HelmManager.InstallReleaseV1":       namespace.CanCreateNamespaceScopedResourceOperation,
	"HelmManager.UninstallReleaseV1":     namespace.CanDeleteNamespaceScopedResourceOperation,
	"HelmManager.UpgradeReleaseV1":       namespace.CanUpdateNamespaceScopedResourceOperation,
	"HelmManager.RollbackReleaseV1":      namespace.CanUpdateNamespaceScopedResourceOperation,
	"HelmManager.ReleasePreview":         project.CanViewProjectOperation,
	"HelmManager.GetReleaseHistory":      namespace.CanViewNamespaceScopedResourceOperation,
	"HelmManager.GetReleaseManifest":     namespace.CanViewNamespaceScopedResourceOperation,
	"HelmManager.GetReleaseStatus":       project.CanViewProjectOperation,
	"HelmManager.GetReleaseDetailExtend": project.CanViewProjectOperation,
	"HelmManager.GetReleasePods":         project.CanViewProjectOperation,
	"HelmManager.ImportClusterRelease":   project.CanViewProjectOperation,

	// Addons
	"ClusterAddons.ListAddons":      project.CanViewProjectOperation,
	"ClusterAddons.GetAddonsDetail": project.CanViewProjectOperation,
	"ClusterAddons.InstallAddons":   project.CanViewProjectOperation,
	"ClusterAddons.UpgradeAddons":   project.CanViewProjectOperation,
	"ClusterAddons.StopAddons":      project.CanViewProjectOperation,
	"ClusterAddons.UninstallAddons": project.CanViewProjectOperation,
}
