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

// Package types xxx
package types

import (
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	schedtypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

// NodeInfo mesos node information
type NodeInfo struct {
	// node ip, example 127.0.0.1
	Ip string
	// clusterid, example BCS-MESOS-10001
	Clusterid string
	// node module, distinguish between switches, example 上海-周浦-M3
	// scheduler detection container best locate different switches
	// so best locate different switches
	Module string
	// idc info, example 上海-周浦
	// every idc deploy three detection containers
	// when all(three) detection containers can't ping, then give an alarm
	Idc string
	// idc unit,对应cmdb机房管理单元
	IdcUnit string
}

// DeployDetection detection record
type DeployDetection struct {
	// clusterid
	Clusterid string
	// deploy idc
	Idc string
	// idc unit
	IdcUnit string
	// deployment json
	Template commtypes.BcsDeployment
	// cluster nodes
	Nodes []*NodeInfo
	// created application info, include status
	// if Application!=nil, then the idc has deployed detection node
	// other else nothing
	Application *schedtypes.Application
	// application's taskgroup
	Pods []*schedtypes.TaskGroup
}

// APIResponse api response
type APIResponse struct {
	Result  bool        `json:"result"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// CmdbHostInfo host info for cmdb
type CmdbHostInfo struct {
	ModuleName string `json:"ModuleName"`
	IDC        string `json:"IDC"`
	ServerRack string `json:"serverRack"`
	IDCUnit    string `json:"IDCUnit"`
}

// DetectionPod detection pod
type DetectionPod struct {
	Ip        string
	Idc       string
	IdcUnit   string
	ClusterId string
}
