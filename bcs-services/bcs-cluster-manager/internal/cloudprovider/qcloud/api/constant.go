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

package api

const (
	// NodeGroupLifeStateCreating node group life state creating
	NodeGroupLifeStateCreating = "creating"
	// NodeGroupLifeStateNormal node group life state normal
	NodeGroupLifeStateNormal = "normal"
	// NodeGroupLifeStateUpdating node group life state updating
	NodeGroupLifeStateUpdating = "updating"
	// NodeGroupLifeStateDeleting node group life state deleting
	NodeGroupLifeStateDeleting = "deleting"
	// NodeGroupLifeStateDeleted node group life state deleted
	NodeGroupLifeStateDeleted = "deleted"
)

// InternetChargeType
const (
	// InternetChargeTypeBandwidthPrepaid 带宽预付费
	InternetChargeTypeBandwidthPrepaid = "BANDWIDTH_PREPAID"
	// InternetChargeTypeBandwidthPostpaidByHour 带宽按小时后付费
	InternetChargeTypeBandwidthPostpaidByHour = "BANDWIDTH_POSTPAID_BY_HOUR"
	// InternetChargeTypeTrafficPostpaidByHour 按流量付费
	InternetChargeTypeTrafficPostpaidByHour = "TRAFFIC_POSTPAID_BY_HOUR"
)

// Cluster Status
const (
	// ClusterStatusRunning running
	ClusterStatusRunning = "Running"
	// ClusterStatusAbnormal abnormal
	ClusterStatusAbnormal = "Abnormal"
)

const (
	// DiskCloudPremium 高性能云硬盘
	DiskCloudPremium = "CLOUD_PREMIUM"
	// DiskCloudSsd SSD云硬盘
	DiskCloudSsd = "CLOUD_SSD"
)

// 实例的最新操作状态
const (
	// SUCCESS success
	SUCCESS = "SUCCESS"
	// OPERATING doing
	OPERATING = "OPERATING"
	// FAILED failed
	FAILED = "FAILED"
)

// 实例状态
const (
	// PENDING 表示创建中
	PENDING = "PENDING"
	// LAUNCHFAILED 表示创建失败
	LAUNCHFAILED = "LAUNCH_FAILED"
	// RUNNING 表示运行中
	RUNNING = "RUNNING"
	// STOPPED 表示关机
	STOPPED = "STOPPED"
	// STARTING 表示开机中
	STARTING = "STARTING"
	// STOPPING 表示关机中
	STOPPING = "STOPPING"
	// REBOOTING 表示重启中
	REBOOTING = "REBOOTING"
	// SHUTDOWN 表示停止待销毁
	SHUTDOWN = "SHUTDOWN"
	// TERMINATING 表示销毁中
	TERMINATING = "TERMINATING"
)
