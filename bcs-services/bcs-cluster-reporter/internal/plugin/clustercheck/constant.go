/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package clustercheck xxx
package clustercheck

const (
	pluginName                               = "clustercheck"
	ClusterAvailabilityCheckMetricName       = "cluster_availability"
	ClusterVersionMetricName                 = "cluster_version"
	ClusterCheckDurationMeticName            = "cluster_check_duration_seconds"
	ClusterApiserverCertExpirationMetricName = "cluster_apiserver_cert_expiration"

	// Status
	NormalStatus                   = "ok"
	AboutToExpireStatus            = "expire_soon"
	ClusterAvailabilityPanicStatus = "panic"

	AvailabilityConfigFailStatus          = "config_fail"
	AvailabilityClusterFailStatus         = "connect_cluster_fail"
	AvailabilityNamespaceFailStatus       = "namespace_fail"
	AvailabilityWorkloadExistStatus       = "workload_exist"
	AvailabilityCreateWorkloadErrorStatus = "create_workload_fail"
	AvailabilityCreatePodTimeoutStatus    = "create_pod_timeout"
	AvailabilitySchedulePodTimeoutStatus  = "schedule_pod_timeout"
	AvailabilityTimeOffsetStatus          = "time_offset"
	AvailabilityWatchErrorStatus          = "watch_fail"
	AvailabilityNoNodeErrorStatus         = "node_fail"

	// Detail
	AboutToExpireDetail       = "AboutToExpireDetail"
	ClusterAvailabilityDetail = "ClusterAvailabilityDetail"

	ClusterVersionLabel                     = "ClusterVersionLabel"
	ClusterVersionItem                      = "ClusterVersionItem"
	ClusterApiserverCertExpirationCheckItem = "ClusterApiserverCertExpiration"
	ClusterAvailabilityItem                 = "ClusterAvailabilityItem"
	ClusterLatencyItem                      = "ClusterLatencyItem"
	ApiserverTarget                         = "apiserver"

	workloadToPod      = "create_pod"
	workloadToSchedule = "schedule_pod"
	worloadToRunning   = "start_pod"

	workloadToPodItem      = "workloadToPodTarget"
	workloadToScheduleItem = "workloadToSchedule"
	worloadToRunningItem   = "worloadToRunning"
)

var (
	ChinenseStringMap = map[string]string{
		// status
		NormalStatus:                          "正常",
		ClusterAvailabilityPanicStatus:        ClusterAvailabilityPanicStatus,
		AvailabilityWorkloadExistStatus:       "workload已存在",
		AvailabilityCreateWorkloadErrorStatus: "创建workload失败",
		AvailabilityTimeOffsetStatus:          "apiserver时间偏移",
		AvailabilityWatchErrorStatus:          "watch失败",

		pluginName:                              "集群控制面检查",
		AboutToExpireDetail:                     "%s Apiserver 的证书将在 %d 秒内过期",
		ClusterAvailabilityDetail:               "%s 的黑盒监控检测结果异常: %s",
		ClusterVersionLabel:                     "集群版本",
		ClusterVersionItem:                      "集群版本",
		ClusterApiserverCertExpirationCheckItem: "apiserver证书过期时间",
		ClusterAvailabilityItem:                 "集群黑盒监控",
		ClusterLatencyItem:                      "黑盒监控延迟",

		workloadToPodItem:      "创建pod",
		workloadToScheduleItem: "调度pod",
		worloadToRunningItem:   "执行pod",

		ApiserverTarget: ApiserverTarget,
	}

	EnglishStringMap = map[string]string{
		// status
		NormalStatus:                          NormalStatus,
		ClusterAvailabilityPanicStatus:        ClusterAvailabilityPanicStatus,
		AvailabilityWorkloadExistStatus:       AvailabilityWorkloadExistStatus,
		AvailabilityCreateWorkloadErrorStatus: AvailabilityCreateWorkloadErrorStatus,
		AvailabilityTimeOffsetStatus:          AvailabilityTimeOffsetStatus,
		AvailabilityWatchErrorStatus:          AvailabilityWatchErrorStatus,

		pluginName:                              pluginName,
		AboutToExpireDetail:                     "%s Apiserver cert is about to expiration in %d seconds, ",
		ClusterAvailabilityDetail:               "%s blackbox check result is %s",
		ClusterVersionLabel:                     "cluster version",
		ClusterVersionItem:                      "cluster version",
		ClusterApiserverCertExpirationCheckItem: "apiserver cert expiration",
		ClusterAvailabilityItem:                 "cluster blackbox check",
		ClusterLatencyItem:                      "blackbox check latency",

		workloadToPodItem:      "create pod",
		workloadToScheduleItem: "schedule pod",
		worloadToRunningItem:   "excute pod",

		ApiserverTarget: ApiserverTarget,
	}

	StringMap = ChinenseStringMap
)
