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

package eop

import (
	"fmt"
	"sync"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var validateMgr sync.Once

func init() {
	validateMgr.Do(func() {
		// init Cluster
		cloudprovider.InitCloudValidateManager(cloudName, &CloudValidate{})
	})
}

// CloudValidate eopCloud validate management implementation
type CloudValidate struct {
}

var availableVersions = []string{"1.19.16", "1.20.15", "1.25.6"}

// CreateClusterValidate create cluster validate
func (c *CloudValidate) CreateClusterValidate(req *proto.CreateClusterReq, opt *cloudprovider.CommonOption) error {
	if len(req.ClusterBasicSettings.Version) == 0 {
		return fmt.Errorf("%s CreateClusterValidate lost kubernetes version in request", cloudName)
	}

	if !utils.StringInSlice(req.ClusterBasicSettings.Version, availableVersions) {
		return fmt.Errorf("%s CreateClusterValidate not supportted kubernetes version %s in request",
			cloudName, req.ClusterBasicSettings.Version)
	}

	if len(req.Instances) == 0 {
		return fmt.Errorf("%s CreateClusterValidate master nodes not set", cloudName)
	}

	if req.Instances[0].NodeRole != common.NodeRoleMaster {
		return fmt.Errorf("%s CreateClusterValidate nodeRole should be %s", cloudName, common.NodeRoleMaster)
	}

	if len(req.NodeGroups) == 0 {
		return fmt.Errorf("%s CreateClusterValidate node group not set", cloudName)
	}

	return nil
}

// CreateCloudAccountValidate create cloud account validate
func (c *CloudValidate) CreateCloudAccountValidate(account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ImportClusterValidate check importCluster operation
func (c *CloudValidate) ImportClusterValidate(req *proto.ImportClusterReq, opt *cloudprovider.CommonOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ImportCloudAccountValidate create cloudAccount account validation
func (c *CloudValidate) ImportCloudAccountValidate(account *proto.Account) error {
	// call eopCloud interface to check account
	if c == nil || account == nil {
		return fmt.Errorf("%s ImportCloudAccountValidate request is empty", cloudName)
	}

	if len(account.SecretID) == 0 || len(account.SecretKey) == 0 {
		return fmt.Errorf("%s ImportCloudAccountValidate request lost valid crendential info", cloudName)
	}

	return nil
}

// GetCloudRegionZonesValidate xxx
func (c *CloudValidate) GetCloudRegionZonesValidate(req *proto.GetCloudRegionZonesRequest,
	account *proto.Account) error {
	// call eopCloud interface to check account
	if c == nil || req == nil {
		return fmt.Errorf("%s GetCloudRegionZonesValidate request is empty", cloudName)
	}

	if len(req.Region) == 0 || len(req.CloudID) == 0 {
		return fmt.Errorf("%s GetCloudRegionZonesValidate request lost valid region info", cloudName)
	}

	if account == nil || len(account.SecretID) == 0 || len(account.SecretKey) == 0 {
		return fmt.Errorf("%s GetCloudRegionZonesValidate request lost valid crendential info", cloudName)
	}

	return nil
}

// ListCloudRegionClusterValidate xxx
func (c *CloudValidate) ListCloudRegionClusterValidate(req *proto.ListCloudRegionClusterRequest,
	account *proto.Account) error {
	// call eopCloud interface to check account
	if c == nil || account == nil {
		return fmt.Errorf("%s ListCloudRegionClusterValidate request is empty", cloudName)
	}

	if len(account.SecretID) == 0 || len(account.SecretKey) == 0 {
		return fmt.Errorf("%s ListCloudRegionClusterValidate request lost valid crendential info", cloudName)
	}

	return nil
}

// ListCloudVpcsValidate xxx
func (c *CloudValidate) ListCloudVpcsValidate(req *proto.ListCloudVpcsRequest,
	account *proto.Account) error {
	if c == nil {
		return fmt.Errorf("%s ListCloudVpcsValidate request is empty", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListCloudVpcsValidate request lost valid region info", cloudName)
	}

	return nil
}

// ListCloudSubnetsValidate xxx
func (c *CloudValidate) ListCloudSubnetsValidate(req *proto.ListCloudSubnetsRequest,
	account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ListSecurityGroupsValidate xxx
func (c *CloudValidate) ListSecurityGroupsValidate(req *proto.ListCloudSecurityGroupsRequest,
	account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ListKeyPairsValidate list key pairs validate
func (c *CloudValidate) ListKeyPairsValidate(req *proto.ListKeyPairsRequest, account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ListInstancesValidate xxx
func (c *CloudValidate) ListInstancesValidate(req *proto.ListCloudInstancesRequest, account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ListInstanceTypeValidate xxx
func (c *CloudValidate) ListInstanceTypeValidate(req *proto.ListCloudInstanceTypeRequest,
	account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ListCloudOsImageValidate xxx
func (c *CloudValidate) ListCloudOsImageValidate(req *proto.ListCloudOsImageRequest, account *proto.Account) error {
	// call eopCloud interface to check account
	if c == nil || account == nil {
		return fmt.Errorf("%s ListCloudOsImageValidate request is empty", cloudName)
	}

	if len(account.SecretID) == 0 || len(account.SecretKey) == 0 {
		return fmt.Errorf("%s ListCloudOsImageValidate request lost valid crendential info", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListCloudOsImageValidate request lost valid region info", cloudName)
	}

	return nil
}

// AddNodesToClusterValidate xxx
func (c *CloudValidate) AddNodesToClusterValidate(req *proto.AddNodesRequest, opt *cloudprovider.CommonOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// DeleteNodesFromClusterValidate xxx
func (c *CloudValidate) DeleteNodesFromClusterValidate(req *proto.DeleteNodesRequest,
	opt *cloudprovider.CommonOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// CreateNodeGroupValidate xxx
func (c *CloudValidate) CreateNodeGroupValidate(req *proto.CreateNodeGroupRequest,
	opt *cloudprovider.CommonOption) error {
	return cloudprovider.ErrCloudNotImplemented
}
