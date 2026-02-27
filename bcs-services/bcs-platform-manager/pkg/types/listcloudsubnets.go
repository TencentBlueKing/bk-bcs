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

// ListCloudSubnetsRequest list cloud subnets page request
type ListCloudSubnetsRequest struct {
	// CloudID 云信息
	// 最小长度：2
	CloudID string `json:"cloudID" in:"path=cloudID" validate:"min=2"`

	// Region 云地域信息
	Region string `json:"region" in:"query=region" validate:"omitempty"`

	// AccountID 云凭证ID
	AccountID string `json:"accountID" in:"query=accountID" validate:"omitempty"`

	// VpcID 过滤查询指定vpc的信息
	VpcID string `json:"vpcID" in:"query=vpcID" validate:"omitempty"`
}

// ListCloudSubnetsResponse list cloud subnets page response
type ListCloudSubnetsResponse struct {
	Total   uint32             `json:"total"`
	Subnets []ListCloudSubnets `json:"subnets"`
}

// ListCloudSubnets VPC信息
// @Description vpc信息
type ListCloudSubnets struct {
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
