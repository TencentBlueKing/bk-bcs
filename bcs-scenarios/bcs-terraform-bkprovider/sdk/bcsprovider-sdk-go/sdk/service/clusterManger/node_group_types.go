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

// Package clusterManger cluster-service
package clusterManger

import "github.com/golang/protobuf/jsonpb"

/*
	节点池
*/

const (
	// createNodeGroupApi post
	createNodeGroupApi = "/clustermanager/v1/nodegroup"

	// deleteNodeGroupApi delete ( nodeGroupID )
	deleteNodeGroupApi = "/clustermanager/v1/nodegroup/%s"

	// updateNodeGroupApi put ( nodeGroupID )
	updateNodeGroupApi = "/clustermanager/v1/nodegroup/%s"

	// updateGroupDesiredNodeApi post ( nodeGroupID )
	updateGroupDesiredNodeApi = "/clustermanager/v1/nodegroup/%s/desirednode"

	// updateGroupMinMaxSizeApi post ( nodeGroupID )
	updateGroupMinMaxSizeApi = "/clustermanager/v1/nodegroup/%s/boundsize"

	// getNodeGroupApi get ( nodeGroupID )
	getNodeGroupApi = "/clustermanager/v1/nodegroup/%s"

	// listClusterNodeGroupApi get ( clusterID )
	listClusterNodeGroupApi = "/clustermanager/v1/clusters/%s/nodegroups"
)

const (
	// dockerGraphPath default docker graphPath
	dockerGraphPath = "/data/bcs/service/docker"
)

// 可用区子网模式
const (
	// Priority 在高优先级的子网与可用区创建实例(默认)
	Priority = "PRIORITY"

	// Equality 所有可用区、子网机会均衡(打散)
	Equality = "EQUALITY"
)

// 重试策略
const (
	// ImmediateRetry 立即重试, 在较短时间内快速重试, 连续失败超过一定次数（5次）后不再重试.(默认)
	ImmediateRetry = "IMMEDIATE_RETRY"

	// IncrementalIntervals 间隔递增重试, 随着连续失败次数的增加, 重试间隔逐渐增大, 重试间隔从秒级到1天不等.
	IncrementalIntervals = "INCREMENTAL_INTERVALS"

	// NoRetry 不进行重试, 直到再次收到用户调用或者告警信息后才会重试.
	NoRetry = "NO_RETRY"
)

// 扩容模式
const (
	// ClassicScaling 扩容时创建新实例,缩容时销毁实例 (默认)
	ClassicScaling = "CLASSIC_SCALING"

	// WakeUpStoppedScaling 缩容时关机不销毁, 扩容时优先唤醒关机实例
	WakeUpStoppedScaling = "WAKE_UP_STOPPED_SCALING"
)

// 实例计费模式
const (
	// Prepaid 表示预付费, 即包年包月(默认)
	Prepaid = "PREPAID"

	// PostpaidByHour 表示后付费, 即按量计费
	PostpaidByHour = "POSTPAID_BY_HOUR"

	// Spotpaid 表示竞价实例付费.
	Spotpaid = "SPOTPAID"
)

var (
	// pbMarshaller 创建一个jsonpb.Marshaler
	pbMarshaller = new(jsonpb.Marshaler)
)
