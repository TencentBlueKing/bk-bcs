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

// Package dbprivilege  x
package dbprivilege

import (
	"fmt"
)

// DbPrivOptions options for db privilege plugin
type DbPrivOptions struct {
	KubeMaster         string `json:"kube_master"`
	Kubeconfig         string `json:"kubeconfig"`
	NetworkType        string `json:"network_type"`
	ExternalSysType    string `json:"external_sys_type"`
	ExternalSysConfig  string `json:"external_sys_config"`
	InitContainerImage string `json:"init_container_image"`
	ServicePort        int    `json:"service_port"`
	ServiceName        string `json:"service_name"`
	ServiceNamespace   string `json:"service_namespace"`
	ServiceServerPort  int    `json:"service_server_port"`
	DbmOptimizeEnabled bool   `json:"dbm_enabled"`
	TicketTimer        int    `json:"ticket_timer"`
}

// Validate validate options
func (dpo *DbPrivOptions) Validate() error {
	if len(dpo.NetworkType) == 0 {
		dpo.NetworkType = NetworkTypeOverlay
	}
	if dpo.NetworkType != NetworkTypeOverlay &&
		dpo.NetworkType != NetworkTypeUnderlay {
		return fmt.Errorf("invalid network_type %s", dpo.NetworkType)
	}
	if len(dpo.ExternalSysConfig) == 0 {
		return fmt.Errorf("external_sys_config cannot be empty")
	}
	if len(dpo.InitContainerImage) == 0 {
		return fmt.Errorf("init_container_image cannot be empty")
	}
	return nil
}
