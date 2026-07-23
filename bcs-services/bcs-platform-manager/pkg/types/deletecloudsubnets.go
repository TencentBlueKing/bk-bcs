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

// DeleteCloudSubnetsRequest delete cloud subnets cluster request
type DeleteCloudSubnetsRequest struct {
	// CloudID 云信息
	// 最小长度：2
	CloudID string `json:"cloudID" in:"path=cloudID" validate:"min=2"`

	// Region 云地域信息
	Region string `json:"region" in:"query=region" validate:"omitempty"`

	// AccountID 云凭证ID
	AccountID string `json:"accountID" in:"query=accountID" validate:"omitempty"`

	// SubnetID 子网ID
	// 最小长度：1（必填且非空）
	SubnetID string `json:"subnetID" in:"path=subnetID" validate:"min=1"`
}

// DeleteCloudSubnetsResponse delete cloud subnets cluster response
type DeleteCloudSubnetsResponse struct {
}
