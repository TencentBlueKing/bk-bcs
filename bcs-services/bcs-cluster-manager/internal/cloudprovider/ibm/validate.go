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

package ibm

import (
	"sync"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

var cloudValidateMgr sync.Once

func init() {
	cloudValidateMgr.Do(func() {
		cloudprovider.InitCloudValidateManager(cloudName, &CloudValidateManager{})
	})
}

// CloudValidateManager validate for ibm cloud resource
type CloudValidateManager struct{}

// CreateClusterValidate create cluster validate
func (c *CloudValidateManager) CreateClusterValidate(req *proto.CreateClusterReq,
	opt *cloudprovider.CommonOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ImportClusterValidate import cluster validate
func (c *CloudValidateManager) ImportClusterValidate(req *proto.ImportClusterReq,
	opt *cloudprovider.CommonOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// AddNodesToClusterValidate validate
func (c *CloudValidateManager) AddNodesToClusterValidate(req *proto.AddNodesV2Request,
	opt *cloudprovider.CommonOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// DeleteNodesFromClusterValidate validate
func (c *CloudValidateManager) DeleteNodesFromClusterValidate(req *proto.DeleteNodesRequest,
	opt *cloudprovider.CommonOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// CreateCloudAccountValidate create cloud account validate
func (c *CloudValidateManager) CreateCloudAccountValidate(account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ImportCloudAccountValidate import cloud account validate
func (c *CloudValidateManager) ImportCloudAccountValidate(account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// GetCloudRegionZonesValidate get cloud region zones validate
func (c *CloudValidateManager) GetCloudRegionZonesValidate(req *proto.GetCloudRegionZonesRequest,
	account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ListCloudRegionClusterValidate get cloud region zones validate
func (c *CloudValidateManager) ListCloudRegionClusterValidate(req *proto.ListCloudRegionClusterRequest,
	account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ListCloudSubnetsValidate list subnets validate
func (c *CloudValidateManager) ListCloudSubnetsValidate(req *proto.ListCloudSubnetsRequest,
	account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ListCloudVpcsValidate list vpcs validate
func (c *CloudValidateManager) ListCloudVpcsValidate(req *proto.ListCloudVpcsRequest,
	account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ListSecurityGroupsValidate list SecurityGroups validate
func (c *CloudValidateManager) ListSecurityGroupsValidate(req *proto.ListCloudSecurityGroupsRequest,
	account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ListKeyPairsValidate list key pairs validate
func (c *CloudValidateManager) ListKeyPairsValidate(req *proto.ListKeyPairsRequest,
	account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ListInstanceTypeValidate list instance type validate
func (c *CloudValidateManager) ListInstanceTypeValidate(req *proto.ListCloudInstanceTypeRequest,
	account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ListCloudOsImageValidate list tke image os validate
func (c *CloudValidateManager) ListCloudOsImageValidate(req *proto.ListCloudOsImageRequest,
	account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// CreateNodeGroupValidate create node group validate
func (c *CloudValidateManager) CreateNodeGroupValidate(req *proto.CreateNodeGroupRequest,
	opt *cloudprovider.CommonOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ListInstancesValidate list instance type validate
func (c *CloudValidateManager) ListInstancesValidate(req *proto.ListCloudInstancesRequest,
	account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// AllowCrossBizNodes allow cross biz nodes
func (c *CloudValidateManager) AllowCrossBizNodes(cluster *proto.Cluster) bool {
	return false
}
