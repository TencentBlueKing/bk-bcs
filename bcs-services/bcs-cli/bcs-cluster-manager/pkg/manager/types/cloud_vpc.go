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

package types

// CreateCloudVPCReq 创建云私有网络request
type CreateCloudVPCReq struct {
	CloudID     string `json:"cloudID"`
	NetworkType string `json:"networkType"`
	Region      string `json:"region"`
	VPCName     string `json:"vpcName"`
	VPCID       string `json:"vpcID"`
	Available   string `json:"available"`
}

// UpdateCloudVPCReq 更新云私有网络request
type UpdateCloudVPCReq struct {
	CloudID     string `json:"cloudID"`
	NetworkType string `json:"networkType"`
	Region      string `json:"region"`
	RegionName  string `json:"regionName"`
	VPCName     string `json:"vpcName"`
	VPCID       string `json:"vpcID"`
	Available   string `json:"available"`
	Updater     string `json:"updater"`
}

// DeleteCloudVPCReq 删除云私有网络request
type DeleteCloudVPCReq struct {
	CloudID string `json:"cloudID"`
	VPCID   string `json:"vpcID"`
}

// ListCloudVPCReq 查询云私有网络request
type ListCloudVPCReq struct {
	NetworkType string `json:"networkType"`
}

// ListCloudRegionsReq 查询云区域列表request
type ListCloudRegionsReq struct {
	CloudID string `json:"cloudID"`
}

// GetVPCCidrReq 查询私有网络cidr request
type GetVPCCidrReq struct {
	VPCID string `json:"vpcID"`
}

// ListCloudVPCResp 查询云私有网络列表response
type ListCloudVPCResp struct {
	Data []*CloudVPC `json:"data"`
}

// ListCloudRegionsResp 查询云区域列表response
type ListCloudRegionsResp struct {
	Data []*CloudRegion `json:"data"`
}

// GetVPCCidrResp 查询私有网络cidr response
type GetVPCCidrResp struct {
	Data []*VPCCidr `json:"data"`
}

// CloudVPC 云私有网络信息
type CloudVPC struct {
	CloudID        string `json:"cloudID"`
	Region         string `json:"region"`
	RegionName     string `json:"regionName"`
	NetworkType    string `json:"networkType"`
	VPCID          string `json:"vpcID"`
	VPCName        string `json:"vpcName"`
	Available      string `json:"available"`
	Extra          string `json:"extra"`
	ReservedIPNum  uint32 `json:"reservedIPNum"`
	AvailableIPNum uint32 `json:"availableIPNum"`
}

// CloudRegion 云区域信息
type CloudRegion struct {
	CloudID    string `json:"cloudID"`
	Region     string `json:"region"`
	RegionName string `json:"regionName"`
}

// VPCCidr 私有网络cidr
type VPCCidr struct {
	VPC      string `json:"vpc"`
	Cidr     string `json:"cidr"`
	IPNumber uint32 `json:"ipNumber"`
	Status   string `json:"status"`
}

// CloudVPCMgr 云私有网络管理接口
type CloudVPCMgr interface {
	// Create 创建云VPC管理信息
	Create(CreateCloudVPCReq) error
	// Update 更新云vpc信息
	Update(UpdateCloudVPCReq) error
	// Delete 删除特定cloud vpc信息
	Delete(DeleteCloudVPCReq) error
	// List 查询Cloud VPC列表
	List(ListCloudVPCReq) (ListCloudVPCResp, error)
	// ListCloudRegions 根据cloudID获取所属cloud的地域信息
	ListCloudRegions(ListCloudRegionsReq) (ListCloudRegionsResp, error)
	// GetVPCCidr 根据vpcID获取所属vpc的cidr信息
	GetVPCCidr(GetVPCCidrReq) (GetVPCCidrResp, error)
}
