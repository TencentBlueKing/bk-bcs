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

// ListCloudVpcClusterRequest list cloud vpc cluster request
type ListCloudVpcClusterRequest struct {
	// CloudID 云信息
	// 最小长度：2
	CloudID string `json:"cloudID" in:"path=cloudID" validate:"min=2"`

	// Region 云地域信息
	// 最小长度：2
	Region string `json:"region" in:"query=region" validate:"min=2"`

	// AccountID 云凭证ID
	AccountID string `json:"accountID" in:"query=accountID" validate:"omitempty"`

	// VpcID 过滤查询指定vpc的信息
	// 最小长度：2 | 最大长度：100 | 格式：仅包含数字、字母、短横线（-）
	VpcID string `json:"vpcID" in:"path=vpcID" validate:"min=2,max=100"`

	// Offset 查询偏移量
	// 最小值：0
	Offset uint32 `json:"offset" in:"query=offset" validate:"gte=0"`

	// Limit 查询限制数量
	// 最大值：5000
	Limit uint32 `json:"limit" in:"query=limit" validate:"lte=5000"`
}

// ListCloudVpcClusterResponse list cloud vpc cluster response
type ListCloudVpcClusterResponse struct {
	Total        uint32         `json:"total"`
	CloudCluster []CloudCluster `json:"cloudCluster"`
}

// CloudCluster VPC信息
// @Description vpc信息
type CloudCluster struct {
	// ClusterID 集群ID
	ClusterID string `json:"clusterID"`

	// OverlayIPCidr Overlay IP CIDR列表（嵌套结构体）
	OverlayIPCidr []OverlayIPCidr `json:"overlayIPCidr"`
}
