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

package huawei

import (
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
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

// CloudValidate qcloud validate management implementation
type CloudValidate struct {
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

	if opt.Account.SecretID == "" && opt.Account.SecretKey == "" {
		return fmt.Errorf("%s ImportClusterValidate request lost valid crendential info", cloudName)
	}

	if len(opt.Region) == 0 {
		return fmt.Errorf("%s ImportClusterValidate request lost valid region info", cloudName)
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

// ImportCloudAccountValidate create cloudAccount account validation
func (c *CloudValidate) ImportCloudAccountValidate(account *proto.Account) error {
	// call cloud interface to check account
	if c == nil || account == nil {
		return fmt.Errorf("%s ImportCloudAccountValidate request is empty", cloudName)
	}

	if account.SecretID == "" || account.SecretKey == "" {
		return fmt.Errorf("%s ImportCloudAccountValidate request lost valid crendential info", cloudName)
	}

	return nil
}

// GetCloudRegionZonesValidate xxx
func (c *CloudValidate) GetCloudRegionZonesValidate(req *proto.GetCloudRegionZonesRequest,
	account *proto.Account) error {
	// call cloud interface to check account
	if c == nil || account == nil {
		return fmt.Errorf("%s GetCloudRegionZonesValidate request is empty", cloudName)
	}

	if account.SecretID == "" || account.SecretKey == "" {
		return fmt.Errorf("%s GetCloudRegionZonesValidate request lost valid crendential info", cloudName)
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

	if account.SecretID == "" && account.SecretKey == "" {
		return fmt.Errorf("%s ListCloudRegionClusterValidate request lost valid crendential info", cloudName)
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

	if account.SecretID == "" && account.SecretKey == "" {
		return fmt.Errorf("%s ListCloudSubnetsValidate request lost valid crendential info", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListCloudSubnetsValidate request lost valid region info", cloudName)
	}

	return nil
}

// ListSecurityGroupsValidate xxx
func (c *CloudValidate) ListSecurityGroupsValidate(req *proto.ListCloudSecurityGroupsRequest,
	account *proto.Account) error {
	// call cloud interface to check account
	return cloudprovider.ErrCloudNotImplemented
}

// ListInstanceTypeValidate xxx
func (c *CloudValidate) ListInstanceTypeValidate(req *proto.ListCloudInstanceTypeRequest,
	account *proto.Account) error {
	// call cloud interface to check account
	return cloudprovider.ErrCloudNotImplemented
}

// ListCloudOsImageValidate xxx
func (c *CloudValidate) ListCloudOsImageValidate(req *proto.ListCloudOsImageRequest, account *proto.Account) error {
	// call cloud interface to check account
	return cloudprovider.ErrCloudNotImplemented
}

// CreateNodeGroupValidate xxx
func (c *CloudValidate) CreateNodeGroupValidate(req *proto.CreateNodeGroupRequest,
	opt *cloudprovider.CommonOption) error {
	// call cloud interface to check account
	return cloudprovider.ErrCloudNotImplemented
}

// AddNodesToClusterValidate xxx
func (c *CloudValidate) AddNodesToClusterValidate(req *proto.AddNodesRequest, opt *cloudprovider.CommonOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// CreateCloudAccountValidate create cloud account validate
func (c *CloudValidate) CreateCloudAccountValidate(account *proto.Account) error {
	// call cloud interface to check account
	if c == nil || account == nil {
		return fmt.Errorf("%s CreateCloudAccountValidate request is empty", cloudName)
	}

	if len(account.SecretID) == 0 || len(account.SecretKey) == 0 {
		return fmt.Errorf("%s CreateCloudAccountValidate request lost valid crendential info", cloudName)
	}

	client, err := api.GetIamClient(&cloudprovider.CommonOption{Account: account})
	if err != nil {
		return err
	}

	req := model.KeystoneListRegionsRequest{}
	_, err = client.KeystoneListRegions(&req)
	if err != nil {
		return err
	}

	return nil
}

// CreateClusterValidate check createCluster operation
func (c *CloudValidate) CreateClusterValidate(req *proto.CreateClusterReq, opt *cloudprovider.CommonOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// DeleteNodesFromClusterValidate xxx
func (c *CloudValidate) DeleteNodesFromClusterValidate(req *proto.DeleteNodesRequest,
	opt *cloudprovider.CommonOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ListCloudVpcsValidate xxx
func (c *CloudValidate) ListCloudVpcsValidate(req *proto.ListCloudVpcsRequest,
	account *proto.Account) error {
	if c == nil || account == nil {
		return fmt.Errorf("%s ListCloudVpcsValidate request is empty", cloudName)
	}

	if len(account.SecretID) == 0 || len(account.SecretKey) == 0 {
		return fmt.Errorf("%s ListCloudVpcsValidate request lost valid crendential info", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListCloudVpcsValidate request lost valid region info", cloudName)
	}

	return nil
}

// ListInstancesValidate xxx
func (c *CloudValidate) ListInstancesValidate(req *proto.ListCloudInstancesRequest, account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ListKeyPairsValidate list key pairs validate
func (c *CloudValidate) ListKeyPairsValidate(req *proto.ListKeyPairsRequest, account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}
