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

package blueking

import (
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

var validateMgr sync.Once

func init() {
	validateMgr.Do(func() {
		// init Cluster
		cloudprovider.InitCloudValidateManager(cloudName, &CloudValidate{})
	})
}

// CloudValidate blueKingCloud validate management implementation
type CloudValidate struct {
}

// CreateClusterValidate create cluster validate
func (c *CloudValidate) CreateClusterValidate(req *proto.CreateClusterReq, opt *cloudprovider.CommonOption) error {
	// kubernetes version
	if len(req.ClusterBasicSettings.Version) == 0 {
		return fmt.Errorf("%s CreateClusterValidate lost kubernetes version in request", cloudName)
	}

	// check masterIP
	if req.ManageType == common.ClusterManageTypeIndependent && len(req.Master) == 0 {
		return fmt.Errorf("%s CreateClusterValidate lost kubernetes cluster masterIP", cloudName)
	}

	// default not handle systemReinstall
	req.SystemReinstall = true

	// auto generate master nodes
	if req.AutoGenerateMasterNodes && len(req.Instances) == 0 {
		return fmt.Errorf("%s CreateClusterValidate invalid instanceTemplate config "+
			"when AutoGenerateMasterNodes=true", cloudName)
	}

	// use existed instances
	if !req.AutoGenerateMasterNodes {
		switch req.ManageType {
		case common.ClusterManageTypeManaged:
			if len(req.Nodes) == 0 {
				return fmt.Errorf("%s CreateClusterValidate invalid node config "+
					"when AutoGenerateMasterNodes false in MANAGED_CLUSTER", cloudName)
			}
		default:
			if len(req.Master) == 0 {
				return fmt.Errorf("%s CreateClusterValidate invalid master config "+
					"when AutoGenerateMasterNodes false in INDEPENDENT_CLUSTER", cloudName)
			}
		}
	}

	return nil
}

// CreateCloudAccountValidate create cloud account validate
func (c *CloudValidate) CreateCloudAccountValidate(account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ImportClusterValidate check importCluster operation
func (c *CloudValidate) ImportClusterValidate(req *proto.ImportClusterReq, opt *cloudprovider.CommonOption) error {
	// call blueking interface to create cluster
	if c == nil || req == nil {
		return fmt.Errorf("%s ImportClusterValidate request is empty", cloudName)
	}

	if opt == nil {
		return fmt.Errorf("%s ImportClusterValidate options is empty", cloudName)
	}

	if req.CloudMode.KubeConfig == "" {
		return fmt.Errorf("%s ImportClusterValidate cluster kubeConfig empty", cloudName)
	}

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

	version, err := kubeCli.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("%s ImportClusterValidate connect cluster by kubeConfig failed: %v", cloudName, err)
	}
	req.Version = version.String()

	blog.Infof("%s ImportClusterValidate CloudMode connect cluster ByKubeConfig success", cloudName)

	return nil
}

// ImportCloudAccountValidate create cloudAccount account validation
func (c *CloudValidate) ImportCloudAccountValidate(req *proto.Account) error {
	// blueking cloud not cloud Account
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
	return nil
}

// ListKeyPairsValidate list key pairs validate
func (c *CloudValidate) ListKeyPairsValidate(req *proto.ListKeyPairsRequest, account *proto.Account) error {
	return nil
}

// ListInstancesValidate xxx
func (c *CloudValidate) ListInstancesValidate(req *proto.ListCloudInstancesRequest, account *proto.Account) error {
	return nil
}

// ListInstanceTypeValidate xxx
func (c *CloudValidate) ListInstanceTypeValidate(
	req *proto.ListCloudInstanceTypeRequest, account *proto.Account) error {
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
	return nil
}
