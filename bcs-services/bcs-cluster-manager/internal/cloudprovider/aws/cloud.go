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

package aws

import (
	"fmt"
	"sync"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/aws/api"

	"github.com/aws/aws-sdk-go/service/eks"
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

	cluster, err := client.GetEksCluster(cls.ClusterName)
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

	// cluster cloud basic setting
	clusterBasicSettingByGKE(cls, cluster)

	// cluster cloud network setting
	clusterNetworkSettingByGKE(cls, cluster)

	return nil
}

func clusterBasicSettingByGKE(cls *cmproto.Cluster, cluster *eks.Cluster) {
	cls.ClusterBasicSettings = &cmproto.ClusterBasicSetting{
		Version:     *cluster.Version,
		VersionName: *cluster.Version,
	}
	// if cluster.NodeConfig != nil {
	// 	cls.ClusterBasicSettings.OS = cluster.NodeConfig.ImageType
	// }
}

func clusterNetworkSettingByGKE(cls *cmproto.Cluster, cluster *eks.Cluster) {
	cls.NetworkSettings = &cmproto.NetworkSetting{
		ClusterIPv4CIDR: *cluster.KubernetesNetworkConfig.ServiceIpv4Cidr,
		ServiceIPv4CIDR: *cluster.KubernetesNetworkConfig.ServiceIpv6Cidr,
	}
	// if cluster.DefaultMaxPodsConstraint != nil {
	// 	cls.NetworkSettings.MaxNodePodNum = uint32(cluster.DefaultMaxPodsConstraint.MaxPodsPerNode)
	// }
}
