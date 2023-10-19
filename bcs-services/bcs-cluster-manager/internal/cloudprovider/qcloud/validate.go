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
 *
 */

package qcloud

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"strconv"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

var validateMgr sync.Once

func init() {
	validateMgr.Do(func() {
		// init Cluster
		cloudprovider.InitCloudValidateManager("qcloud", &CloudValidate{})
	})
}

// CloudValidate qcloud validate management implementation
type CloudValidate struct {
}

// CreateClusterValidate create cluster validate
func (c *CloudValidate) CreateClusterValidate(req *proto.CreateClusterReq, opt *cloudprovider.CommonOption) error {
	// kubernetes version
	if len(req.ClusterBasicSettings.Version) == 0 {
		return fmt.Errorf("%s CreateClusterValidate lost kubernetes version in request", cloudName)
	}

	// check masterIP
	if len(req.Master) == 0 {
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
	if !req.AutoGenerateMasterNodes && len(req.Master) == 0 {
		return fmt.Errorf("%s CreateClusterValidate invalid master config "+
			"when AutoGenerateMasterNodes=false", cloudName)
	}

	// check cidr
	if len(req.NetworkSettings.ClusterIPv4CIDR) > 0 {
		cidr, err := cloudprovider.GetStorageModel().GetTkeCidr(
			context.Background(), req.VpcID, req.NetworkSettings.ClusterIPv4CIDR)
		if err != nil {
			blog.Errorf("get cluster cidr[%s:%s] info failed: %v",
				req.VpcID, req.NetworkSettings.ClusterIPv4CIDR, err)
			return err
		}
		if cidr.Status == common.TkeCidrStatusUsed || cidr.Cluster != "" {
			errMsg := fmt.Errorf("create cluster cidr[%s:%s] already used by cluster(%s)",
				req.VpcID, req.NetworkSettings.ClusterIPv4CIDR, cidr.Cluster)
			return errMsg
		}
	}
	// check vpc-cni
	if req.NetworkSettings.EnableVPCCni {
		if req.NetworkSettings.SubnetSource == nil {
			return fmt.Errorf("networkSetting.SubnetSource cannot be empty when enable vpc-cni")
		}
		subnetIDs := make([]string, 0)
		switch {
		case req.NetworkSettings.SubnetSource.Existed != nil:
			if len(req.NetworkSettings.SubnetSource.Existed.Ids) == 0 {
				return fmt.Errorf("existed subet ids cannot be empty")
			}
			subnetIDs = req.NetworkSettings.SubnetSource.Existed.Ids
		case req.NetworkSettings.SubnetSource.New != nil:
			// apply vpc cidr subnet by mask and zone
			return fmt.Errorf("current not support apply vpc subnet cidr when vpc-cni mode")
		}
		req.NetworkSettings.EniSubnetIDs = subnetIDs

		if req.NetworkSettings.IsStaticIpMode && req.NetworkSettings.ClaimExpiredSeconds <= 0 {
			req.NetworkSettings.ClaimExpiredSeconds = 300
		}
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

	if len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
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

		_, err = cli.GetTKECluster(req.CloudMode.CloudID)
		if err != nil {
			return fmt.Errorf("%s ImportClusterValidate GetTKECluster[%s] failed: %v", cloudName,
				req.CloudMode.CloudID, err)
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
	if c == nil || account == nil {
		return fmt.Errorf("%s GetCloudRegionZonesValidate request is empty", cloudName)
	}

	if len(account.SecretID) == 0 || len(account.SecretKey) == 0 {
		return fmt.Errorf("%s GetCloudRegionZonesValidate request lost valid crendential info", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s GetCloudRegionZonesValidate request lost valid region info", cloudName)
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

// ListCloudVPCV2Validate xxx
func (c *CloudValidate) ListCloudVPCV2Validate(req *proto.ListCloudVPCV2Request, account *proto.Account) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ListCloudSubnetsValidate xxx
func (c *CloudValidate) ListCloudSubnetsValidate(req *proto.ListCloudSubnetsRequest, account *proto.Account) error {
	// call qcloud interface to check account
	if c == nil || account == nil {
		return fmt.Errorf("%s ListCloudSubnetsValidate request is empty", cloudName)
	}

	if len(account.SecretID) == 0 || len(account.SecretKey) == 0 {
		return fmt.Errorf("%s ListCloudSubnetsValidate request lost valid crendential info", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListCloudSubnetsValidate request lost valid region info", cloudName)
	}
	if len(req.VpcID) == 0 {
		return fmt.Errorf("%s ListCloudSubnetsValidate request lost valid vpcID info", cloudName)
	}

	return nil
}

// ListSecurityGroupsValidate xxx
func (c *CloudValidate) ListSecurityGroupsValidate(req *proto.ListCloudSecurityGroupsRequest,
	account *proto.Account) error {
	// call qcloud interface to check account
	if c == nil || account == nil {
		return fmt.Errorf("%s ListSecurityGroupsValidate request is empty", cloudName)
	}

	if len(account.SecretID) == 0 || len(account.SecretKey) == 0 {
		return fmt.Errorf("%s ListSecurityGroupsValidate request lost valid crendential info", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListSecurityGroupsValidate request lost valid region info", cloudName)
	}

	return nil
}

// ListInstanceTypeValidate xxx
func (c *CloudValidate) ListInstanceTypeValidate(req *proto.ListCloudInstanceTypeRequest,
	account *proto.Account) error {
	// call qcloud interface to check account
	if c == nil || account == nil {
		return fmt.Errorf("%s ListInstanceTypeValidate request is empty", cloudName)
	}

	if len(account.SecretID) == 0 || len(account.SecretKey) == 0 {
		return fmt.Errorf("%s ListInstanceTypeValidate request lost valid crendential info", cloudName)
	}

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListInstanceTypeValidate request lost valid region info", cloudName)
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

// CreateNodeGroupValidate xxx
func (c *CloudValidate) CreateNodeGroupValidate(req *proto.CreateNodeGroupRequest,
	opt *cloudprovider.CommonOption) error {

	if len(req.Region) == 0 {
		return fmt.Errorf("%s ListCloudOsImageValidate request lost valid region info", cloudName)
	}

	// simply check instanceType conf info
	if req.LaunchTemplate.CPU == 0 || req.LaunchTemplate.Mem == 0 {
		return fmt.Errorf("validateLaunchTemplate cpu/mem empty")
	}
	// check internetAccess conf info
	if req.LaunchTemplate.InternetAccess != nil {
		bandwidth, _ := strconv.Atoi(req.LaunchTemplate.InternetAccess.InternetMaxBandwidth)
		if req.LaunchTemplate.InternetAccess.PublicIPAssigned && bandwidth == 0 {
			return fmt.Errorf("validateLaunchTemplate internetAccess bandwidth empty")
		}
	}

	return nil
}
