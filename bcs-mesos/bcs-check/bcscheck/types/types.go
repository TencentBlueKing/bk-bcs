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

package types

import (
	commtype "bk-bcs/bcs-common/common/types"

	"golang.org/x/net/context"
)

//HealthSyncData holder for sync data
type HealthSyncData struct {
	Action      SyncDataAction //operation, like Add, Delete, Update
	HealthCheck *HealthCheck
}

type HealthCheck struct {
	ID           string
	OriginID     string
	Type         commtype.BcsHealthCheckType
	TaskMode     HealthTaskMode
	ReporterMode HealthReporterMode

	DelaySeconds        int
	GracePeriodSeconds  int
	IntervalSeconds     int
	TimeoutSeconds      int
	ConsecutiveFailures uint32

	Http *HttpHealthCheck
	Tcp  *TcpHealthCheck

	Operation HealthCheckOperation // operation, like running, stopped
	Status    HealthCheckStatus    //status, like running, failed
	Message   string
}

type WatchPathController struct {
	Path string
	Type ZkPathType
	Data interface{}

	Cxt    context.Context    //context for creating sub context
	Cancel context.CancelFunc //for cancel sub goroutine
}

type HttpHealthCheck struct {
	Ip      string
	Port    int32
	Scheme  string
	Path    string
	Headers map[string]string
}

type TcpHealthCheck struct {
	Ip   string
	Port int32
}

// get clusterID from clusterkeeper
type RequestInfo struct {
	Module string `josn:"module"`
	IP     string `json:"ip"`
}

type RespInfo struct {
	ClusterID string `json:"clusterid"`
	Extension string `json:"extendedinfo"`
}

type Response struct {
	ErrCode int       `json:"errcode"`
	ErrMsg  string    `json:"errmsg"`
	Data    *RespInfo `json:"data"`
}

type HealthCheckType string
type HttpProtocol string
type HttpMethod string
type HealthCheckOperation string
type HealthCheckStatus string
type SyncDataAction string
type HealthTaskMode string
type HealthReporterMode string
type ZkPathType string
type HealthNetworkMode string

const (
	/*HealthCheckOperation*/
	HealthCheckOperationRunning = "running"
	HealthCheckOperationStopped = "stopped"

	/*SyncDataAction*/
	SyncDataActionAdd    = "add"
	SyncDataActionDelete = "delete"
	SyncDataActionUpdate = "update"

	/*HealthCheckStatus*/
	HealthCheckStatusRunning = "running"
	HealthCheckStatusFailed  = "failed"

	/*HealthTaskMode*/
	HealthTaskModeMesos = "mesos"

	/*HealthReporterMode*/
	HealthReporterModeScheduler = "scheduler"

	/*ZkPathType*/
	ZkPathTypeApplication = "application" // zk path, like /blueking/application
	ZkPathTypeNamespace   = "namespace"   // zk path, like /blueking/application/defaultGroup
	ZkPathTypeAppname     = "appname"     // zk path, like /blueking/application/defaultGroup/app-001
	ZkPathTypeTaskgroup   = "taskgroup"   // zk path, like /blueking/application/defaultGroup/app-001/{taskgroup}
	ZkPathTypeTask        = "task"        // zk path, like /blueking/application/defaultGroup/app-001/{taskgroup}/{task}
)
