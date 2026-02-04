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

// CreateCloudSubnetsRequest create cloud subnets cluster request
type CreateCloudSubnetsRequest struct {
	// CloudID 云信息
	// 最小长度：2
	CloudID string `json:"cloudID" in:"path=cloudID" validate:"min=2"`

	// VpcID vpcID信息
	// 最小长度：2 | 最大长度：100 | 格式：仅包含数字、字母、短横线（-）
	VpcID string `json:"vpcID" validate:"min=2,max=100"`

	// Region 云地域信息
	Region string `json:"region" validate:"omitempty"`

	// AccountID 云凭证ID
	AccountID string `json:"accountID" validate:"omitempty"`

	// SubnetName 子网名称
	SubnetName string `json:"subnetName" validate:"omitempty"`

	// CidrBlock 子网CIDR
	CidrBlock string `json:"cidrBlock" validate:"omitempty"`

	// Zone 可用区
	Zone string `json:"zone" validate:"omitempty"`
}

// CreateCloudSubnetsResponse create cloud subnets cluster response
type CreateCloudSubnetsResponse struct {
	Subnet CloudSubnets `json:"subnet"`
}

// CloudSubnets VPC信息
// @Description vpc信息
type CloudSubnets struct {
	VpcID                   string      `json:"vpcID"`
	SubnetID                string      `json:"subnetID"`
	SubnetName              string      `json:"subnetName"`
	CidrRange               string      `json:"cidrRange"`
	Ipv6CidrRange           string      `json:"ipv6CidrRange"`
	Zone                    string      `json:"zone"`
	AvailableIPAddressCount uint64      `json:"availableIPAddressCount"`
	ZoneName                string      `json:"zoneName"`
	Cluster                 ClusterInfo `json:"cluster"`
	HwNeutronSubnetID       string      `json:"hwNeutronSubnetID"`
	TotalIpAddressCount     uint64      `json:"totalIpAddressCount"`
}

// ClusterInfo 集群信息
type ClusterInfo struct {
	ClusterName string `json:"clusterName"`
	ClusterID   string `json:"clusterID"`
}
