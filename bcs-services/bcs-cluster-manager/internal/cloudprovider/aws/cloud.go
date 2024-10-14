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

package aws

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"sync"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/aws/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var cloudInfoMgr sync.Once

func init() {
	cloudInfoMgr.Do(func() {
		// init Cluster
		cloudprovider.InitCloudInfoManager(cloudName, &CloudInfoManager{})
	})
}

// CloudInfoManager management cluster info
type CloudInfoManager struct {
}

// InitCloudClusterDefaultInfo init cluster defaultConfig
func (c *CloudInfoManager) InitCloudClusterDefaultInfo(cls *cmproto.Cluster,
	opt *cloudprovider.InitClusterConfigOption) error {
	// call aws interface to init cluster defaultConfig
	if c == nil || cls == nil {
		return fmt.Errorf("%s InitCloudClusterDefaultInfo request is empty", cloudName)
	}

	if opt == nil || opt.Cloud == nil {
		return fmt.Errorf("%s InitCloudClusterDefaultInfo option is empty", cloudName)
	}

	if len(cls.ManageType) == 0 {
		cls.ManageType = common.ClusterManageTypeIndependent
	}
	if len(cls.ClusterCategory) == 0 {
		cls.ClusterCategory = common.Builder
	}

	return nil
}

// SyncClusterCloudInfo get cluster cloudInfo by clusterID or kubeConfig
func (c *CloudInfoManager) SyncClusterCloudInfo(cls *cmproto.Cluster,
	opt *cloudprovider.SyncClusterCloudInfoOption) error {
	if c == nil || cls == nil {
		return fmt.Errorf("%s SyncClusterCloudInfo request is empty", cloudName)
	}

	if opt == nil || opt.Cloud == nil {
		return fmt.Errorf("%s SyncClusterCloudInfo option is empty", cloudName)
	}

	// get cloud cluster

	client, err := api.NewEksClient(opt.Common)
	if err != nil {
		return err
	}

	cluster, err := client.GetEksCluster(opt.ImportMode.CloudID)
	if err != nil {
		return err
	}

	kubeConfig, err := api.GetClusterKubeConfig(opt.Common, cluster)
	if err != nil {
		return fmt.Errorf("SyncClusterCloudInfo GetClusterKubeConfig failed: %v", err)
	}

	cls.KubeConfig = kubeConfig
	cls.SystemID = *cluster.Name
	cls.VpcID = *cluster.ResourcesVpcConfig.VpcId

	ec2Client, err := api.GetEc2Client(opt.Common)
	if err != nil {
		return err
	}

	res, err := ec2Client.DescribeVpcs(&ec2.DescribeVpcsInput{
		VpcIds: []*string{&cls.VpcID},
	})
	if err != nil {
		return err
	}

	var ipv4Cidr, ipv6Cidr string
	if len(res.Vpcs) == 1 {
		ipv4Cidr = *res.Vpcs[0].CidrBlock
		if len(res.Vpcs[0].Ipv6CidrBlockAssociationSet) > 0 {
			ipv6Cidr = *res.Vpcs[0].Ipv6CidrBlockAssociationSet[0].Ipv6CidrBlock
		}
	}

	// cluster cloud basic setting
	clusterBasicSettingByEks(cls, cluster)

	// cluster cloud network setting
	clusterNetworkSettingByEks(cls, cluster, ipv4Cidr, ipv6Cidr)

	return nil
}

func clusterBasicSettingByEks(cls *cmproto.Cluster, cluster *eks.Cluster) {
	cls.ClusterBasicSettings = &cmproto.ClusterBasicSetting{
		Version:     *cluster.Version,
		VersionName: *cluster.Version,
	}
	// if cluster.NodeConfig != nil {
	// 	cls.ClusterBasicSettings.OS = cluster.NodeConfig.ImageType
	// }
}

func clusterNetworkSettingByEks(cls *cmproto.Cluster, cluster *eks.Cluster, ipv4Cidr, ipv6Cidr string) {
	if cluster.KubernetesNetworkConfig.ServiceIpv4Cidr != nil {
		maxServiceNum, _ := utils.ConvertCIDRToStep(*cluster.KubernetesNetworkConfig.ServiceIpv4Cidr)
		cls.NetworkSettings = &cmproto.NetworkSetting{
			ClusterIPv4CIDR: ipv4Cidr,
			ServiceIPv4CIDR: *cluster.KubernetesNetworkConfig.ServiceIpv4Cidr,
			MaxServiceNum:   maxServiceNum,
		}
	} else if cluster.KubernetesNetworkConfig.ServiceIpv6Cidr != nil {
		cls.NetworkSettings = &cmproto.NetworkSetting{
			ClusterIPv6CIDR: ipv6Cidr,
			ServiceIPv6CIDR: *cluster.KubernetesNetworkConfig.ServiceIpv6Cidr,
		}
	}

	// if cluster.DefaultMaxPodsConstraint != nil {
	// 	cls.NetworkSettings.MaxNodePodNum = uint32(cluster.DefaultMaxPodsConstraint.MaxPodsPerNode)
	// }
}

// UpdateClusterCloudInfo update cluster cloud info
func (c *CloudInfoManager) UpdateClusterCloudInfo(cls *cmproto.Cluster) error {
	return nil
}
