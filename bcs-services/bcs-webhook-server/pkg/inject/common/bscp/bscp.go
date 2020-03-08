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

package bscp

import (
	"encoding/json"
	"fmt"

	"bk-bcs/bcs-common/common/blog"
)

const (
	// SideCarPrefix prefix for bscp sidecar envs
	SideCarPrefix = "BSCP_BCSSIDECAR_"
	// SideCarCfgPath sidecar config path
	SideCarCfgPath = "BSCP_BCSSIDECAR_APPCFG_PATH"

	// SideCarMod sidecar labels for multiple apps
	SideCarMod = "BSCP_BCSSIDECAR_APPINFO_MOD"

	// SideCarBusiness sidecar business name for single app
	SideCarBusiness = "BSCP_BCSSIDECAR_APPINFO_BUSINESS"
	// SideCarApp sidecar app name for single app
	SideCarApp = "BSCP_BCSSIDECAR_APPINFO_APP"
	// SideCarCluster sidecar cluster name for single app
	SideCarCluster = "BSCP_BCSSIDECAR_APPINFO_CLUSTER"
	// SideCarZone sidecar zone name for single app
	SideCarZone = "BSCP_BCSSIDECAR_APPINFO_ZONE"
	// SideCarDc sidecar dc name for single app
	SideCarDc = "BSCP_BCSSIDECAR_APPINFO_DC"
	// SideCarVolumeName shared volume name for sidecar and business container
	SideCarVolumeName = "bscp-sidecar-cfg-shared"
	// AnnotationKey annotation key to enable sidecar injection
	AnnotationKey = "bkbscp.tencent.com/sidecar-injection"
	// AnnotationValue annotation value to enable sidecar injection
	AnnotationValue = "enabled"

	// PatchOperationAdd patch add operation
	PatchOperationAdd = "add"
	// PatchOperationReplace patch replace operation
	PatchOperationReplace = "replace"
	// PatchOperationRemove patch remove operation
	PatchOperationRemove = "remove"
	// PatchPathVolumes volumes path for patch operation
	PatchPathVolumes = "/spec/volumes/%v"
	// PatchPathContainers containers path for patch operation
	PatchPathContainers = "/spec/containers/%v"
)

// AppModInfo is multi app mode information.
type AppModInfo struct {
	// BusinessName business name.
	BusinessName string `json:"business"`

	// AppName app name.
	AppName string `json:"app"`

	// ClusterName cluster name.
	ClusterName string `json:"cluster"`

	// ZoneName zone name.
	ZoneName string `json:"zone"`

	// DC datacenter tag.
	DC string `json:"dc"`

	// Labels sidecar instance KV labels.
	Labels map[string]string `json:"labels,omitempty"`

	// Path is sidecar mod app configs effect path.
	Path string `json:"path"`
}

// ValidateEnvs Validate env
func ValidateEnvs(envMap map[string]string) bool {
	_, okBusiness := envMap[SideCarBusiness]
	_, okApp := envMap[SideCarApp]
	_, okCluster := envMap[SideCarCluster]
	_, okZone := envMap[SideCarZone]
	_, okDc := envMap[SideCarDc]
	if !okBusiness || !okApp || !okCluster || !okZone || !okDc {
		return false
	}
	return true
}

// AddPathIntoAppInfoMode add config path into appMode env value
func AddPathIntoAppInfoMode(envValue string, path string) (string, error) {
	appInfoModes := []*AppModInfo{}
	err := json.Unmarshal([]byte(envValue), &appInfoModes)
	if err != nil {
		blog.Errorf("unmarshal env %s failed, err %s", SideCarMod, err.Error())
		return "", fmt.Errorf("unmarshal env %s failed, err %s", SideCarMod, err.Error())
	}
	if len(appInfoModes) == 0 {
		blog.Errorf("env %s exists, but nothing info in it", SideCarMod)
		return "", fmt.Errorf("env %s exists, but nothing info in it", SideCarMod)
	}
	for _, mod := range appInfoModes {
		if len(mod.BusinessName) == 0 || len(mod.AppName) == 0 ||
			len(mod.ClusterName) == 0 || len(mod.ZoneName) == 0 || len(mod.DC) == 0 {
			blog.Errorf("app info mod is invalid, one or more fields of [business, app, cluster, zone, dc] missing, %+v", mod)
			return "", fmt.Errorf("app info mod is invalid, one or more fields of [business, app, cluster, zone, dc] missing, %+v", mod)
		}
		mod.Path = path
	}
	retBytes, err := json.Marshal(appInfoModes)
	if err != nil {
		blog.Errorf("json encode %v failed, err %s", appInfoModes, err.Error())
		return "", err
	}
	return string(retBytes), nil
}
