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

package discovery

import "github.com/Tencent/bk-bcs/bcs-services/bcs-service-prometheus/types"

const (
	//DefaultBcsModuleLabelKey label key
	DefaultBcsModuleLabelKey = "bcs_module"
	// DiscoveryFileName promethus file name
	DiscoveryFileName = "_sd_config.json"
)

const (
	// CadvisorModule name
	CadvisorModule = "cadvisor"
	// NodeexportModule name
	NodeexportModule = "node_export"
)

// Discovery interface for prometheus discovery
type Discovery interface {
	//start
	Start() error

	// get prometheus service discovery config
	GetPrometheusSdConfig(module string) ([]*types.PrometheusSdConfig, error)

	// get prometheus sd config file path
	GetPromSdConfigFile(module string) string

	//register event handle function
	RegisterEventFunc(handleFunc EventHandleFunc)
}

// EventHandleFunc event handler for callback
type EventHandleFunc func(dInfo Info)

// Info information
type Info struct {
	//mesosModules: commtypes.BCS_MODULE_SCHEDULER, commtypes.BCS_MODULE_MESOSDATAWATCH ...
	//serviceModules: commtypes.BCS_MODULE_APISERVER, commtypes.BCS_MODULE_STORAGE, commtypes.BCS_MODULE_NETSERVICE ...
	//nodeModules: discovery.CadvisorModule, discovery.NodeexportModule
	//serviceMonitor: ServiceMonitor
	Module string
	//changed key
	Key string
}
