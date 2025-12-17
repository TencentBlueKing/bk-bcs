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

// Package options xxx
package options

const (
	// BcsSubnetResourceQueueName Bcs子网资源任务队列名称
	BcsSubnetResourceQueueName = "bcssubnetresource"
	// VpcIPMonitorQueueName VPC IP监测任务队列名称
	VpcIPMonitorQueueName = "vpcipmonitor"
	// VpcOverlayNoticeQueueName VPC overlay ip通知任务队列名称
	VpcOverlayNoticeQueueName = "vpcoverlaynotice"
)

// A list of task types.
const (
	// TypeBcsSubnetResource Bcs子网资源任务类型
	TypeBcsSubnetResource = "bcssubnet:resource"
	// TypeVpcIPMonitor VPC IP监测任务类型
	TypeVpcIPMonitor = "vpcip:monitor"
	// TypeVpcOverlayNotice VPC overlay ip通知任务
	TypeVpcOverlayNotice = "vpcoverlay:notice"
)
