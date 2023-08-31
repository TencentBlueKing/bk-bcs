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

package azure

import (
	"context"
	"fmt"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
)

var cloudInfoMgr sync.Once

func init() {
	cloudInfoMgr.Do(func() {
		//init Cluster
		cloudprovider.InitCloudInfoManager(cloudName, &CloudInfoManager{})
	})
}

// CloudInfoManager Azure management cluster info
type CloudInfoManager struct {
}

// InitCloudClusterDefaultInfo init cluster defaultConfig
func (c *CloudInfoManager) InitCloudClusterDefaultInfo(cls *proto.Cluster,
	opt *cloudprovider.InitClusterConfigOption) error {
	return nil
}

// SyncClusterCloudInfo get cluster cloudInfo by clusterID or kubeConfig
func (c *CloudInfoManager) SyncClusterCloudInfo(cls *proto.Cluster,
	opt *cloudprovider.SyncClusterCloudInfoOption) error {
	if c == nil || cls == nil {
		return fmt.Errorf("%s SyncClusterCloudInfo request is empty", cloudName)
	}
	if opt == nil || opt.Cloud == nil {
		return fmt.Errorf("%s SyncClusterCloudInfo option is empty", cloudName)
	}
	// get cloud cluster
	cluster, err := getCloudCluster(opt)
	if err != nil {
		return fmt.Errorf("SyncClusterCloudInfo failed: %v", err)
	}
	cls.SystemID = *cluster.Name
	// cluster cloud basic setting
	clusterBasicSettingByQCloud(cls, cluster)
	// cluster cloud network setting
	err = clusterNetworkSettingByQCloud(cls, cluster)
	if err != nil {
		blog.Errorf("SyncClusterCloudInfo clusterNetworkSettingByQCloud failed: %v", err)
	}

	return nil
}

func getCloudCluster(opt *cloudprovider.SyncClusterCloudInfoOption) (*armcontainerservice.ManagedCluster, error) {
	client, err := api.NewAksServiceImplWithCommonOption(opt.Common)
	if err != nil {
		return nil, fmt.Errorf("%s getCloudCluster NewContainerServiceClient failed: %v", cloudName, err)
	}
	mc, err := client.GetClusterWithName(context.Background(), opt.Common.Account.ResourceGroupName,
		opt.ImportMode.CloudID)
	if err != nil {
		return nil, fmt.Errorf("%s getCloudCluster GetCluster failed: %v", cloudName, err)
	}
	return mc, nil
}

func clusterBasicSettingByQCloud(cls *proto.Cluster, cluster *armcontainerservice.ManagedCluster) {
	clusterOs := ""
	if len(cluster.Properties.AgentPoolProfiles) > 0 {
		p := cluster.Properties.AgentPoolProfiles
		clusterOs = string(*p[0].OSSKU)
	}
	cls.ClusterBasicSettings = &proto.ClusterBasicSetting{
		OS:          clusterOs,
		Version:     *cluster.Properties.CurrentKubernetesVersion,
		VersionName: *cluster.Properties.CurrentKubernetesVersion,
	}
}

func clusterNetworkSettingByQCloud(cls *proto.Cluster, cluster *armcontainerservice.ManagedCluster) error {
	cidrs := cluster.Properties.NetworkProfile.PodCidrs
	podCidrs := make([]string, len(cidrs))
	for i := range cidrs {
		podCidrs[i] = *cidrs[i]
	}
	cls.NetworkSettings = &proto.NetworkSetting{
		ClusterIPv4CIDR:  *cluster.Properties.NetworkProfile.PodCidr,
		ServiceIPv4CIDR:  *cluster.Properties.NetworkProfile.ServiceCidr,
		MultiClusterCIDR: podCidrs,
	}
	return nil
}
