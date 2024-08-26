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

const (
	// listCloudInstanceTypeApi get ( cloudID )
	listCloudInstanceTypeApi = "/clustermanager/v1/clouds/%s/instancetypes"

	// listCloudSubnetsApi get ( cloudID )
	listCloudSubnetsApi = "/clustermanager/v1/clouds/%s/subnets"

	// listCloudSecurityGroupsApi get ( cloudID )
	listCloudSecurityGroupsApi = "/clustermanager/v1/clouds/%s/securitygroups"

	// getCloudRegionsApi get ( cloudID )
	getCloudRegionsApi = "/clustermanager/v1/clouds/%s/regions"

	// getCloudRegionZonesApi get ( cloudID )
	getCloudRegionZonesApi = "/clustermanager/v1/clouds/%s/zones"

	// listCloudVpcsApi get ( cloudID )
	listCloudVpcsApi = "/clustermanager/v1/clouds/%s/vpcs"

	// listCloudProjectsApi get ( cloudID )
	listCloudProjectsApi = "/clustermanager/v1/clouds/%s/projects"

	// listCloudOsImageApi get ( cloudID )
	listCloudOsImageApi = "/clustermanager/v1/clouds/%s/osimage"

	// listKeypairsApi get ( cloudID )
	listKeypairsApi = "/clustermanager/v1/clouds/%s/keypairs"

	// getCloudAccountTypeApi get ( cloudID )
	getCloudAccountTypeApi = "/clustermanager/v1/clouds/%s/accounttype"

	// getCloudBandwidthPackagesApi get ( cloudID )
	getCloudBandwidthPackagesApi = "/clustermanager/v1/clouds/%s/bwps"
)

// 镜像类型
const (
	// PublicImage 公共镜像
	PublicImage = "PUBLIC_IMAGE"

	// PrivateImage 私有镜像
	PrivateImage = "PRIVATE_IMAGE"

	// SharedImage 共享镜像
	SharedImage = "SHARED_IMAGE"

	// MarketImage 市场镜像
	MarketImage = "MARKET_IMAGE"

	// All 所有镜像
	All = "ALL"
)

// 磁盘类型
const (
	// CloudPremium 高性能云硬盘
	CloudPremium = "CLOUD_PREMIUM"

	// CloudSSD SSD云硬盘
	CloudSSD = "CLOUD_SSD"

	// CloudHSSD 增强型SSD云硬盘
	CloudHSSD = "CLOUD_HSSD"

	// CloudTSSD 极速型SSD云硬盘
	CloudTSSD = "CLOUD_TSSD"

	// CloudBSSD 通用型SSD云硬盘
	CloudBSSD = "CLOUD_BSSD"
)

// 弹性公网IP的网络计费模式
const (
	// BandwidthPrepaidByMonth  表示包月带宽预付费。
	BandwidthPrepaidByMonth = "BANDWIDTH_PREPAID_BY_MONTH"

	// TrafficPostpaidByHour   表示按小时流量后付费。
	TrafficPostpaidByHour = "TRAFFIC_POSTPAID_BY_HOUR"

	// BandwidthPostpaidByHour  表示按小时带宽后付费。
	BandwidthPostpaidByHour = "BANDWIDTH_POSTPAID_BY_HOUR"

	// BandwidthPackage   表示共享带宽包。
	BandwidthPackage = "BANDWIDTH_PACKAGE"
)
