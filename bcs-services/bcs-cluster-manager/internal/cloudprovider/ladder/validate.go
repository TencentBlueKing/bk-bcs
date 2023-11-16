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

package ladder

import (
	"fmt"
	"sync"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

var validateMgr sync.Once

func init() {
	validateMgr.Do(func() {
		// init Cluster
		cloudprovider.InitCloudValidateManager(cloudName, &CloudValidate{})
	})
}

// CreateClusterValidate check createCluster operation
func (c *CloudValidate) CreateClusterValidate(req *proto.CreateClusterReq, opt *cloudprovider.CommonOption) error {
	return nil
}

// CloudValidate blueKingCloud validate management implementation
type CloudValidate struct {
}

// CreateCloudAccountValidate create cloud account validate
func (c *CloudValidate) CreateCloudAccountValidate(account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ImportClusterValidate check importCluster operation
func (c *CloudValidate) ImportClusterValidate(req *proto.ImportClusterReq, opt *cloudprovider.CommonOption) error {
	// yunti cloud not import cluster
	return nil
}

// ImportCloudAccountValidate create cloudAccount account validation
func (c *CloudValidate) ImportCloudAccountValidate(req *proto.Account) error {
	// yunti cloud not cloud Account
	return nil
}

// GetCloudRegionZonesValidate xxx
func (c *CloudValidate) GetCloudRegionZonesValidate(
	req *proto.GetCloudRegionZonesRequest, account *proto.Account) error {
	// blueking cloud not cloud Account
	return nil
}

// ListCloudRegionClusterValidate xxx
func (c *CloudValidate) ListCloudRegionClusterValidate(
	req *proto.ListCloudRegionClusterRequest, account *proto.Account) error {
	// blueking cloud not cloud Account
	return nil
}

// ListCloudSubnetsValidate xxx
func (c *CloudValidate) ListCloudSubnetsValidate(req *proto.ListCloudSubnetsRequest, account *proto.Account) error {
	return nil
}

// ListCloudVpcsValidate xxx
func (c *CloudValidate) ListCloudVpcsValidate(req *proto.ListCloudVpcsRequest,
	account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ListSecurityGroupsValidate xxx
func (c *CloudValidate) ListSecurityGroupsValidate(
	req *proto.ListCloudSecurityGroupsRequest, account *proto.Account) error {
	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListSecurityGroupsValidate request lost region info", cloudName)
	}

	return nil
}

// ListKeyPairsValidate list key pairs validate
func (c *CloudValidate) ListKeyPairsValidate(req *proto.ListKeyPairsRequest, account *proto.Account) error {
	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListKeyPairsValidate request lost valid region info", cloudName)
	}

	return nil
}

// ListInstancesValidate xxx
func (c *CloudValidate) ListInstancesValidate(req *proto.ListCloudInstancesRequest, account *proto.Account) error {
	return nil
}

// ListInstanceTypeValidate xxx
func (c *CloudValidate) ListInstanceTypeValidate(
	req *proto.ListCloudInstanceTypeRequest, account *proto.Account) error {
	if c == nil || req == nil {
		return fmt.Errorf("%s ListInstanceTypeValidate request is empty", cloudName)
	}

	if len(req.Region) == 0 || len(req.ProjectID) == 0 {
		return fmt.Errorf("%s ListInstanceTypeValidate request lost valid info", cloudName)
	}

	return nil
}

// ListCloudOsImageValidate xxx
func (c *CloudValidate) ListCloudOsImageValidate(req *proto.ListCloudOsImageRequest, account *proto.Account) error {
	return nil
}

// AddNodesToClusterValidate xxx
func (c *CloudValidate) AddNodesToClusterValidate(req *proto.AddNodesRequest, opt *cloudprovider.CommonOption) error {
	return nil
}

// DeleteNodesFromClusterValidate xxx
func (c *CloudValidate) DeleteNodesFromClusterValidate(
	req *proto.DeleteNodesRequest, opt *cloudprovider.CommonOption) error {
	return nil
}

// CreateNodeGroupValidate xxx
func (c *CloudValidate) CreateNodeGroupValidate(
	req *proto.CreateNodeGroupRequest, opt *cloudprovider.CommonOption) error {
	if len(req.Region) == 0 {
		return fmt.Errorf("%s CreateNodeGroupValidate request lost valid region info", cloudName)
	}
	// simply check instanceType conf info
	/*
		if req.LaunchTemplate.CPU == 0 || req.LaunchTemplate.Mem == 0 {
			return fmt.Errorf("%s CreateNodeGroupValidate validateLaunchTemplate cpu/mem empty", cloudName)
		}
	*/
	if len(req.NodeTemplate.DataDisks) > 4 {
		return fmt.Errorf("nodeGroup max support 4 data disks")
	}

	return nil
}
