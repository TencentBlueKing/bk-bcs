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

package qcloud

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud-public/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
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

// CreateCloudAccountValidate create cloud account validate
func (c *CloudValidate) CreateCloudAccountValidate(account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// CreateClusterValidate check createCluster operation
func (c *CloudValidate) CreateClusterValidate(req *proto.CreateClusterReq, opt *cloudprovider.CommonOption) error {
	// call qcloud interface to check cluster
	if c == nil || req == nil || opt == nil {
		return fmt.Errorf("%s CreateClusterValidate request&options is empty", cloudName)
	}

	if opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return fmt.Errorf("%s CreateClusterValidate opt lost valid crendential info", cloudName)
	}

	// kubernetes version
	if len(req.ClusterBasicSettings.Version) == 0 {
		return fmt.Errorf("lost kubernetes version in request")
	}
	// default not handle systemReinstall
	req.SystemReinstall = true

	// cluster type
	switch req.ManageType {
	// 托管集群
	case common.ClusterManageTypeManaged:
		if req.AutoGenerateMasterNodes {
			_, nodes := business.GetMasterNodeTemplateConfig(req.Instances)
			if len(nodes) == 0 {
				return fmt.Errorf("instance template empty when auto generate worker nodes in MANAGED_CLUSTER")
			}
			break
		}

		if len(req.Nodes) == 0 {
			return fmt.Errorf("lost kubernetes cluster masterIP when use exited worker nodes")
		}
	// 独立集群
	case common.ClusterManageTypeIndependent:
		if req.AutoGenerateMasterNodes {
			masters, nodes := business.GetMasterNodeTemplateConfig(req.Instances)
			if len(masters) == 0 || len(nodes) == 0 {
				return fmt.Errorf("instance template empty when auto generate master nodes")
			}
			break
		}

		if len(req.Master) == 0 || len(req.Nodes) == 0 {
			return fmt.Errorf("lost kubernetes cluster masterIP when use exited master nodes")
		}
	default:
		return fmt.Errorf("%s not supported cluster type[%s]", cloudName, req.ManageType)
	}

	// cluster category
	if len(req.ClusterCategory) == 0 {
		req.ClusterCategory = common.Builder
	}

	return nil
}

// ImportClusterValidate check importCluster operation
func (c *CloudValidate) ImportClusterValidate(req *proto.ImportClusterReq, opt *cloudprovider.CommonOption) error {
	// call qcloud interface to check cluster
	if c == nil || req == nil {
		return fmt.Errorf("%s ImportClusterValidate request is empty", cloudName)
	}

	if opt == nil {
		return fmt.Errorf("%s ImportClusterValidate options is empty", cloudName)
	}

	if opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return fmt.Errorf("%s ImportClusterValidate opt lost valid crendential info", cloudName)
	}

	if req.CloudMode.CloudID == "" && req.CloudMode.KubeConfig == "" {
		return fmt.Errorf("%s ImportClusterValidate cluster cloudID & kubeConfig empty", cloudName)
	}

	if req.CloudMode.CloudID != "" {
		cli, err := api.NewTkeClient(opt)
		if err != nil {
			return fmt.Errorf("%s ImportClusterValidate getTKEClient failed: %v", cloudName, err)
		}

		tkeCluster, err := cli.GetTKECluster(req.CloudMode.CloudID)
		if err != nil {
			return fmt.Errorf("%s ImportClusterValidate GetTKECluster[%s] failed: %v", cloudName,
				req.CloudMode.CloudID, err)
		}

		// 托管集群导入必须存在节点, 存在节点时才能打通集群链路
		if *tkeCluster.ClusterType == common.ClusterManageTypeManaged && *tkeCluster.ClusterNodeNum == 0 {
			return fmt.Errorf("%s ImportClusterValidate ManageTypeCluster[%s] must exist worker nodes",
				req.CloudMode.CloudID, cloudName)
		}

		blog.Infof("%s ImportClusterValidate CloudMode CloudID[%s] success", cloudName, req.CloudMode.CloudID)
		return nil
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
	// call qcloud interface to check account
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
	// call qcloud interface to check account
	if c == nil || req == nil {
		return fmt.Errorf("%s GetCloudRegionZonesValidate request is empty", cloudName)
	}

	if len(req.Region) == 0 || len(req.CloudID) == 0 {
		return fmt.Errorf("%s GetCloudRegionZonesValidate request lost valid region info", cloudName)
	}

	if options.GetEditionInfo().IsInnerEdition() {
		return nil
	}

	if account == nil || len(account.SecretID) == 0 || len(account.SecretKey) == 0 {
		return fmt.Errorf("%s GetCloudRegionZonesValidate request lost valid crendential info", cloudName)
	}

	return nil
}

// ListCloudRegionClusterValidate xxx
func (c *CloudValidate) ListCloudRegionClusterValidate(req *proto.ListCloudRegionClusterRequest,
	account *proto.Account) error {
	// call qcloud interface to check account
	if c == nil || account == nil {
		return fmt.Errorf("%s ListCloudRegionClusterValidate request is empty", cloudName)
	}

	if len(account.SecretID) == 0 || len(account.SecretKey) == 0 {
		return fmt.Errorf("%s ListCloudRegionClusterValidate request lost valid crendential info", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListCloudRegionClusterValidate request lost valid region info", cloudName)
	}

	return nil
}

// ListCloudSubnetsValidate xxx
func (c *CloudValidate) ListCloudSubnetsValidate(req *proto.ListCloudSubnetsRequest,
	account *proto.Account) error {
	// call qcloud interface to check account
	if c == nil {
		return fmt.Errorf("%s ListCloudSubnetsValidate request is empty", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListCloudSubnetsValidate request lost valid region info", cloudName)
	}
	if len(req.VpcID) == 0 {
		return fmt.Errorf("%s ListCloudSubnetsValidate request lost valid vpcID info", cloudName)
	}

	return nil
}

// ListCloudVpcsValidate xxx
func (c *CloudValidate) ListCloudVpcsValidate(req *proto.ListCloudVpcsRequest,
	account *proto.Account) error {
	// call qcloud interface to check account
	if c == nil {
		return fmt.Errorf("%s ListCloudVpcsValidate request is empty", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListCloudVpcsValidate request lost valid region info", cloudName)
	}

	return nil
}

// ListSecurityGroupsValidate xxx
func (c *CloudValidate) ListSecurityGroupsValidate(req *proto.ListCloudSecurityGroupsRequest,
	account *proto.Account) error {
	// call qcloud interface to check account
	if c == nil || req == nil {
		return fmt.Errorf("%s ListSecurityGroupsValidate request is empty", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListSecurityGroupsValidate request lost valid region info", cloudName)
	}

	if options.GetEditionInfo().IsInnerEdition() {
		return nil
	}

	if account == nil || len(account.SecretID) == 0 || len(account.SecretKey) == 0 {
		return fmt.Errorf("%s ListSecurityGroupsValidate request lost valid crendential info", cloudName)
	}

	return nil
}

// ListKeyPairsValidate list key pairs validate
func (c *CloudValidate) ListKeyPairsValidate(req *proto.ListKeyPairsRequest, account *proto.Account) error {
	// call qcloud interface to check account
	if c == nil || req == nil {
		return fmt.Errorf("%s ListKeyPairsValidate request is empty", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListKeyPairsValidate request lost valid region info", cloudName)
	}

	if options.GetEditionInfo().IsInnerEdition() {
		return nil
	}

	if account == nil || len(account.SecretID) == 0 || len(account.SecretKey) == 0 {
		return fmt.Errorf("%s ListKeyPairsValidate request lost valid crendential info", cloudName)
	}

	return nil
}

// ListInstancesValidate xxx
func (c *CloudValidate) ListInstancesValidate(req *proto.ListCloudInstancesRequest, account *proto.Account) error {
	if c == nil || req == nil {
		return fmt.Errorf("%s ListInstancesValidate request is empty", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListInstancesValidate request lost valid region info", cloudName)
	}

	if options.GetEditionInfo().IsInnerEdition() {
		return nil
	}

	if account == nil || len(account.SecretID) == 0 || len(account.SecretKey) == 0 {
		return fmt.Errorf("%s ListInstancesValidate request lost valid crendential info", cloudName)
	}

	return nil
}

// ListInstanceTypeValidate xxx
func (c *CloudValidate) ListInstanceTypeValidate(
	req *proto.ListCloudInstanceTypeRequest, account *proto.Account) error {
	// call qcloud interface to check account
	if c == nil || req == nil {
		return fmt.Errorf("%s ListInstanceTypeValidate request is empty", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListInstanceTypeValidate request lost valid region info", cloudName)
	}

	if options.GetEditionInfo().IsInnerEdition() {
		if len(req.ProjectID) == 0 {
			return fmt.Errorf("%s ListInstanceTypeValidate request lost valid info", cloudName)
		}

		return nil
	}

	if account == nil || len(account.SecretID) == 0 || len(account.SecretKey) == 0 {
		return fmt.Errorf("%s ListInstanceTypeValidate request lost valid crendential info", cloudName)
	}

	return nil
}

// ListCloudOsImageValidate xxx
func (c *CloudValidate) ListCloudOsImageValidate(req *proto.ListCloudOsImageRequest, account *proto.Account) error {
	// call qcloud interface to check account
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
	// call qcloud interface to check account
	if c == nil || req == nil {
		return fmt.Errorf("%s AddNodesToClusterValidate request is empty", cloudName)
	}

	if req.IsExternalNode && req.NodeGroupID == "" {
		return fmt.Errorf("%s AddNodesToClusterValidate must be depent NodeGroup", cloudName)
	}

	return nil
}

// DeleteNodesFromClusterValidate xxx
func (c *CloudValidate) DeleteNodesFromClusterValidate(req *proto.DeleteNodesRequest,
	opt *cloudprovider.CommonOption) error {
	// call qcloud interface to check account
	if c == nil || req == nil {
		return fmt.Errorf("%s DeleteNodesFromClusterValidate request is empty", cloudName)
	}

	if req.IsExternalNode && req.NodeGroupID == "" {
		return fmt.Errorf("%s DeleteNodesFromClusterValidate must be depent NodeGroup", cloudName)
	}

	return nil
}

// CreateNodeGroupValidate xxx
func (c *CloudValidate) CreateNodeGroupValidate(req *proto.CreateNodeGroupRequest,
	opt *cloudprovider.CommonOption) error {

	if len(req.Region) == 0 {
		return fmt.Errorf("%s CreateNodeGroupValidate request lost valid region info", cloudName)
	}

	/*
		// simply check instanceType conf info
		if req.LaunchTemplate.CPU == 0 || req.LaunchTemplate.Mem == 0 {
			return fmt.Errorf("validateLaunchTemplate cpu/mem empty")
		}
	*/

	// check internetAccess conf info
	if req.LaunchTemplate.InternetAccess != nil {
		bandwidth, _ := strconv.Atoi(req.LaunchTemplate.InternetAccess.InternetMaxBandwidth)
		if req.LaunchTemplate.InternetAccess.PublicIPAssigned && bandwidth == 0 {
			return fmt.Errorf("validateLaunchTemplate internetAccess bandwidth empty")
		}
	}

	return nil
}
