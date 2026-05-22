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

// Package types pod types
package types

// ListCloudVpcsPageRequest list cloud vpcs page request
type ListCloudVpcsPageRequest struct {
	// CloudID 云信息
	// 最小长度：2
	CloudID string `json:"cloudID" in:"path=cloudID" validate:"min=2"`

	// Region 云地域信息
	Region string `json:"region" in:"query=region" validate:"omitempty"`

	// AccountID 云凭证ID
	AccountID string `json:"accountID" in:"query=accountID" validate:"omitempty"`

	// VpcID 过滤查询指定vpc的信息
	VpcID string `json:"vpcID" in:"query=vpcID" validate:"omitempty"`

	// ResourceGroupName Azure Cloud中Vpc所属的resource group名称
	ResourceGroupName string `json:"resourceGroupName" in:"query=resourceGroupName" validate:"omitempty"`

	// VpcName vpc名称
	VpcName string `json:"vpcName" in:"query=vpcName" validate:"omitempty"`

	// Offset 查询偏移量
	// 最小值：0
	Offset uint32 `json:"offset" in:"query=offset" validate:"gte=0"`

	// Limit 查询限制数量
	// 最大值：5000
	Limit uint32 `json:"limit" in:"query=limit" validate:"lte=5000"`
}

// ListCloudVpcsPageResponse list cloud vpcs page response
type ListCloudVpcsPageResponse struct {
	Total     uint32      `json:"total"`
	CloudVpcs []CloudVpcs `json:"cloudVpcs"`
}

// CloudVpcs VPC信息
// @Description vpc信息
type CloudVpcs struct {
	// VpcName vpc名称
	VpcName string `json:"vpcName"`

	// VpcID vpcID
	VpcID string `json:"vpcID"`

	// Region 云地域信息
	Region string `json:"region"`

	// OverlayCidr Overlay CIDR列表
	OverlayCidr []string `json:"overlayCidr"`

	// AvailableOverlayIpNum 可用Overlay IP数量
	AvailableOverlayIpNum uint32 `json:"availableOverlayIpNum"`

	// AvailableOverlayCidr 可用Overlay CIDR列表
	AvailableOverlayCidr []string `json:"availableOverlayCidr"`

	// TotalOverlayIpNum Overlay IP总数
	TotalOverlayIpNum uint32 `json:"totalOverlayIpNum"`

	// OverlayIpUsageRate Overlay IP使用率
	OverlayIpUsageRate float64 `json:"overlayIpUsageRate"`

	// UnderlayCidr underlay CIDR列表
	UnderlayCidr []string `json:"underlayCidr"`

	// AvailableUnderlayIpNum 可用Underlay IP数量
	AvailableUnderlayIpNum uint32 `json:"availableUnderlayIpNum"`

	// AvailableUnderlayCidr 可用Underlay CIDR列表
	AvailableUnderlayCidr []string `json:"availableUnderlayCidr"`

	// TotalUnderlayIpNum Underlay IP总数
	TotalUnderlayIpNum uint32 `json:"totalUnderlayIpNum"`

	// UnderlayIpUsageRate Underlay IP使用率
	UnderlayIpUsageRate float64 `json:"underlayIpUsageRate"`

	// CreateTime 创建时间
	CreateTime string `json:"createTime"`

	// OverlayIPCidr Overlay IP CIDR列表（嵌套结构体）
	OverlayIPCidr []OverlayIPCidr `json:"overlayIPCidr"`
}

// OverlayIPCidr Overlay IP CIDR 信息
// @Description Overlay IP CIDR 详情
type OverlayIPCidr struct {
	// Cidr Overlay IP CIDR
	Cidr string `json:"cidr"`

	// IpNum Overlay IP数量
	IpNum uint32 `json:"ipNum"`
}
