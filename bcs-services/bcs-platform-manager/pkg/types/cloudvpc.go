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

// CreateCloudVPCReq create cloud vpc request
type CreateCloudVPCReq struct {
	CloudID       string `json:"cloudID"`
	NetworkType   string `json:"networkType"`
	Region        string `json:"region"`
	RegionName    string `json:"regionName"`
	VpcName       string `json:"vpcName"`
	VpcID         string `json:"vpcID"`
	Available     string `json:"available"`
	Extra         string `json:"extra"`
	Creator       string `json:"creator"`
	ReservedIPNum uint32 `json:"reservedIPNum"`
	BusinessID    string `json:"businessID"`
	Overlay       *Cidr  `json:"overlay"`
	Underlay      *Cidr  `json:"underlay"`
}

// Cidr cidr
type Cidr struct {
	Cidrs         []*CidrState `json:"cidrs"`
	ReservedIPNum uint32       `json:"reservedIPNum"`
	ReservedCidrs []string     `json:"reservedCidrs"`
}

// CidrState cidr state
type CidrState struct {
	Cidr  string `json:"cidr"`
	Block bool   `json:"block"`
}

// UpdateCloudVPCReq update cloud vpc request
type UpdateCloudVPCReq struct {
	CloudID       string  `json:"cloudID"`
	NetworkType   string  `json:"networkType"`
	Region        string  `json:"region"`
	RegionName    string  `json:"regionName"`
	VpcName       string  `json:"vpcName"`
	VpcID         string  `json:"vpcID"`
	Available     string  `json:"available"`
	Updater       string  `json:"updater"`
	ReservedIPNum *uint32 `json:"reservedIPNum"`
	BusinessID    *string `json:"businessID"`
	Overlay       *Cidr   `json:"overlay"`
	Underlay      *Cidr   `json:"underlay"`
}

// Subnet subnet
type Subnet struct {
	VpcID                   string       `json:"vpcID"`
	SubnetID                string       `json:"subnetID"`
	SubnetName              string       `json:"subnetName"`
	CidrRange               string       `json:"cidrRange"`
	Zone                    string       `json:"zone"`
	AvailableIPAddressCount uint64       `json:"availableIPAddressCount"`
	ZoneName                string       `json:"zoneName"`
	Cluster                 *ClusterInfo `json:"cluster"`
}

// ClusterInfo cluster info
type ClusterInfo struct {
	ClusterName string `json:"clusterName"`
	ClusterID   string `json:"clusterID"`
}

// GetCloudVPCRecommendCIDRReq get cloud vpc recommend cidr request
type GetCloudVPCRecommendCIDRReq struct {
	CloudID     string `json:"cloudID" in:"query=cloudID"`
	Region      string `json:"region" in:"query=region"`
	AccountID   string `json:"accountID" in:"query=accountID"`
	VpcID       string `json:"vpcID" in:"query=vpcID"`
	NetworkType string `json:"networkType" in:"query=networkType"`
	Mask        uint32 `json:"mask" in:"query=mask"`
}

// GetCloudVPCRecommendCIDRResp get cloud vpc recommend cidr response
type GetCloudVPCRecommendCIDRResp struct {
	Cidrs []string `json:"cidrs"`
}
