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

package google

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/google/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

var validateMgr sync.Once

func init() {
	validateMgr.Do(func() {
		// init Cluster
		cloudprovider.InitCloudValidateManager(cloudName, &CloudValidate{})
	})
}

// CloudValidate google cloud validate management implementation
type CloudValidate struct {
}

// CreateClusterValidate create cluster validate
func (c *CloudValidate) CreateClusterValidate(req *proto.CreateClusterReq, opt *cloudprovider.CommonOption) error {
	if c == nil || req == nil {
		return fmt.Errorf("%s CreateClusterValidate request is empty", cloudName)
	}

	if opt == nil || opt.Account == nil {
		return fmt.Errorf("%s CreateClusterValidate options is empty", cloudName)
	}

	if len(opt.Account.ServiceAccountSecret) == 0 {
		return fmt.Errorf("%s CreateClusterValidate request lost valid crendential info", cloudName)
	}

	if len(opt.Account.GkeProjectID) == 0 {
		return fmt.Errorf("%s CreateClusterValidate request lost valid gkeProjectID info", cloudName)
	}

	return nil
}

// ImportClusterValidate check importCluster operation
func (c *CloudValidate) ImportClusterValidate(req *proto.ImportClusterReq, opt *cloudprovider.CommonOption) error {
	// call cloud interface to check cluster
	if c == nil || req == nil {
		return fmt.Errorf("%s ImportClusterValidate request is empty", cloudName)
	}

	if opt == nil || opt.Account == nil {
		return fmt.Errorf("%s ImportClusterValidate options is empty", cloudName)
	}

	if len(opt.Account.ServiceAccountSecret) == 0 {
		return fmt.Errorf("%s ImportClusterValidate request lost valid crendential info", cloudName)
	}

	if len(opt.Account.GkeProjectID) == 0 {
		return fmt.Errorf("%s ImportClusterValidate request lost valid gkeProjectID info", cloudName)
	}

	if req.CloudMode.CloudID == "" && req.CloudMode.KubeConfig == "" {
		return fmt.Errorf("%s ImportClusterValidate cluster cloudID & kubeConfig empty", cloudName)
	}

	if req.CloudMode.KubeConfig != "" {
		_, err := types.GetKubeConfigFromYAMLBody(false, types.YamlInput{
			FileName:    "",
			YamlContent: req.CloudMode.KubeConfig,
		})
		if err != nil {
			return fmt.Errorf("%s ImportClusterValidate GetKubeConfigFromYAMLBody failed: %v", cloudName, err)
		}
		kubeRet := base64.StdEncoding.EncodeToString([]byte(req.CloudMode.KubeConfig))
		kubeCli, err := clusterops.NewKubeClient(kubeRet)
		if err != nil {
			return fmt.Errorf("%s ImportClusterValidate NewKubeClient failed: %v", cloudName, err)
		}
		_, err = kubeCli.Discovery().ServerVersion()
		if err != nil {
			return fmt.Errorf("%s ImportClusterValidate connect cluster by kubeConfig failed: %v", cloudName, err)
		}

		blog.Infof("%s ImportClusterValidate CloudMode connect cluster ByKubeConfig success", cloudName)
	}

	return nil
}

// CreateCloudAccountValidate create cloudAccount validation
func (c *CloudValidate) CreateCloudAccountValidate(account *proto.Account) error {
	// call cloud interface to check account
	if c == nil || account == nil {
		return fmt.Errorf("%s CreateCloudAccountValidate request is empty", cloudName)
	}

	if len(account.ServiceAccountSecret) == 0 {
		return fmt.Errorf("%s CreateCloudAccountValidate request lost valid crendential info", cloudName)
	}

	sas := &types.GCPServiceAccount{}
	err := json.Unmarshal([]byte(account.ServiceAccountSecret), sas)
	if err != nil {
		return fmt.Errorf("%s CreateCloudAccountValidate decode service account secret failed, %v",
			cloudName, err)
	}
	account.GkeProjectID = sas.ProjectID

	gceCli, err := api.NewComputeServiceClient(&cloudprovider.CommonOption{Account: account})
	if err != nil {
		return err
	}

	// 测试云凭证可用性
	_, err = gceCli.ListRegions(context.Background())
	if err != nil {
		return err
	}

	return nil
}

// ImportCloudAccountValidate create cloudAccount account validation
func (c *CloudValidate) ImportCloudAccountValidate(account *proto.Account) error {
	// call cloud interface to check account
	if c == nil || account == nil {
		return fmt.Errorf("%s ImportCloudAccountValidate request is empty", cloudName)
	}

	if len(account.ServiceAccountSecret) == 0 {
		return fmt.Errorf("%s ImportCloudAccountValidate request lost valid crendential info", cloudName)
	}

	sas := &types.GCPServiceAccount{}
	err := json.Unmarshal([]byte(account.ServiceAccountSecret), sas)
	if err != nil {
		return fmt.Errorf("%s CreateCloudAccountValidate decode service account secret failed, %v",
			cloudName, err)
	}
	account.GkeProjectID = sas.ProjectID

	return nil
}

// GetCloudRegionZonesValidate xxx
func (c *CloudValidate) GetCloudRegionZonesValidate(req *proto.GetCloudRegionZonesRequest,
	account *proto.Account) error {
	// call cloud interface to check account
	if c == nil || account == nil {
		return fmt.Errorf("%s GetCloudRegionZonesValidate request is empty", cloudName)
	}

	if len(account.ServiceAccountSecret) == 0 {
		return fmt.Errorf("%s GetCloudRegionZonesValidate request lost valid crendential info", cloudName)
	}

	if len(account.GkeProjectID) == 0 {
		return fmt.Errorf("%s GetCloudRegionZonesValidate request lost valid gkeProjectID info", cloudName)
	}

	return nil
}

// ListCloudRegionClusterValidate xxx
func (c *CloudValidate) ListCloudRegionClusterValidate(req *proto.ListCloudRegionClusterRequest,
	account *proto.Account) error {
	// call cloud interface to check account
	if c == nil || account == nil {
		return fmt.Errorf("%s ListCloudRegionClusterValidate request is empty", cloudName)
	}

	if len(account.ServiceAccountSecret) == 0 {
		return fmt.Errorf("%s ListCloudRegionClusterValidate request lost valid crendential info", cloudName)
	}

	if len(account.GkeProjectID) == 0 {
		return fmt.Errorf("%s ListCloudRegionClusterValidate request lost valid gkeProjectID info", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListCloudRegionClusterValidate request lost valid region info", cloudName)
	}

	return nil
}

// ListCloudSubnetsValidate xxx
func (c *CloudValidate) ListCloudSubnetsValidate(req *proto.ListCloudSubnetsRequest, account *proto.Account) error {
	// call cloud interface to check account
	if c == nil || account == nil {
		return fmt.Errorf("%s ListCloudSubnetsValidate request is empty", cloudName)
	}

	if len(account.ServiceAccountSecret) == 0 {
		return fmt.Errorf("%s ListCloudSubnetsValidate request lost valid crendential info", cloudName)
	}

	if len(account.GkeProjectID) == 0 {
		return fmt.Errorf("%s ListCloudSubnetsValidate request lost valid gkeProjectID info", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListCloudSubnetsValidate request lost valid region info", cloudName)
	}
	if len(req.VpcID) == 0 {
		return fmt.Errorf("%s ListCloudSubnetsValidate request lost valid vpcID info", cloudName)
	}

	return nil
}

// ListCloudVpcsValidate list vpcs validate
func (c *CloudValidate) ListCloudVpcsValidate(req *proto.ListCloudVpcsRequest,
	account *proto.Account) error {
	if c == nil || account == nil {
		return fmt.Errorf("%s ListCloudVpcsValidate request is empty", cloudName)
	}

	if len(account.ServiceAccountSecret) == 0 {
		return fmt.Errorf("%s ListCloudVpcsValidate request lost valid crendential info", cloudName)
	}

	if len(account.GkeProjectID) == 0 {
		return fmt.Errorf("%s ListCloudVpcsValidate request lost valid gkeProjectID info", cloudName)
	}

	return nil
}

// ListSecurityGroupsValidate xxx
func (c *CloudValidate) ListSecurityGroupsValidate(req *proto.ListCloudSecurityGroupsRequest,
	account *proto.Account) error {
	// call cloud interface to check account
	if c == nil || account == nil {
		return fmt.Errorf("%s ListSecurityGroupsValidate request is empty", cloudName)
	}

	if len(account.ServiceAccountSecret) == 0 {
		return fmt.Errorf("%s ListSecurityGroupsValidate request lost valid crendential info", cloudName)
	}

	if len(account.GkeProjectID) == 0 {
		return fmt.Errorf("%s ListSecurityGroupsValidate request lost valid gkeProjectID info", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListSecurityGroupsValidate request lost valid region info", cloudName)
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
func (c *CloudValidate) ListInstanceTypeValidate(req *proto.ListCloudInstanceTypeRequest,
	account *proto.Account) error {
	// call cloud interface to check account
	if c == nil || account == nil {
		return fmt.Errorf("%s ListInstanceTypeValidate request is empty", cloudName)
	}

	if len(account.ServiceAccountSecret) == 0 {
		return fmt.Errorf("%s ListInstanceTypeValidate request lost valid crendential info", cloudName)
	}

	if len(account.GkeProjectID) == 0 {
		return fmt.Errorf("%s ListInstanceTypeValidate request lost valid gkeProjectID info", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListInstanceTypeValidate request lost valid region info", cloudName)
	}

	return nil
}

// ListCloudOsImageValidate xxx
func (c *CloudValidate) ListCloudOsImageValidate(req *proto.ListCloudOsImageRequest, account *proto.Account) error {
	// call cloud interface to check account
	if c == nil || account == nil {
		return fmt.Errorf("%s ListCloudOsImageValidate request is empty", cloudName)
	}

	if len(account.ServiceAccountSecret) == 0 {
		return fmt.Errorf("%s ListCloudOsImageValidate request lost valid crendential info", cloudName)
	}

	if len(account.GkeProjectID) == 0 {
		return fmt.Errorf("%s ListCloudOsImageValidate request lost valid gkeProjectID info", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListCloudOsImageValidate request lost valid region info", cloudName)
	}

	return nil
}

// CreateNodeGroupValidate xxx
func (c *CloudValidate) CreateNodeGroupValidate(req *proto.CreateNodeGroupRequest,
	opt *cloudprovider.CommonOption) error {
	// call cloud interface to check account
	if c == nil || opt.Account == nil {
		return fmt.Errorf("%s CreateNodeGroupValidate request is empty", cloudName)
	}

	if len(opt.Account.ServiceAccountSecret) == 0 {
		return fmt.Errorf("%s CreateNodeGroupValidate request lost valid crendential info", cloudName)
	}

	if len(opt.Account.GkeProjectID) == 0 {
		return fmt.Errorf("%s CreateNodeGroupValidate request lost valid gkeProjectID info", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s CreateNodeGroupValidate request lost valid region info", cloudName)
	}

	return nil
}

// AddNodesToClusterValidate xxx
func (c *CloudValidate) AddNodesToClusterValidate(req *proto.AddNodesV2Request, opt *cloudprovider.CommonOption) error {
	return nil
}

// DeleteNodesFromClusterValidate xxx
func (c *CloudValidate) DeleteNodesFromClusterValidate(
	req *proto.DeleteNodesRequest, opt *cloudprovider.CommonOption) error {
	return nil
}
