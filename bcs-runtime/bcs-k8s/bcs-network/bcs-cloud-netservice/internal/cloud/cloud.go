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

// Package cloud is interface for cloud
package cloud

import (
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/types"
)

const (
	// CloudProviderTencent tencent cloud
	CloudProviderTencent = "tencentcloud"
	// CloudProviderAws aws cloud
	CloudProviderAws = "aws"
)

// Interface interface for access
type Interface interface {
	DescribeSubnet(vpcID, region, subnetID string) (*types.CloudSubnet, error)
	DescribeSubnetList(vpcID, region string, subnetIDs []string) ([]*types.CloudSubnet, error)
	QueryEni(eniID string) (*types.EniObject, error)
	QueryEniList(subnetID string) ([]*types.EniObject, error)
	AssignIPToEni(ip, eniID string) (string, error)
	UnassignIPFromEni(ip []string, eniID string) error
	MigrateIP(ip, srcEniID, destEniID string) error
}
