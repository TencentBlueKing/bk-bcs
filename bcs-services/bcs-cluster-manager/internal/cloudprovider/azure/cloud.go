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

// Package azure xxx
package azure

import (
	"context"
	"fmt"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v3"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

var cloudInfoMgr sync.Once

func init() {
	cloudInfoMgr.Do(func() {
		// init Cluster
		cloudprovider.InitCloudInfoManager(cloudName, &CloudInfoManager{})
	})
}

// CloudInfoManager Azure management cluster info
type CloudInfoManager struct {
}

// InitCloudClusterDefaultInfo init cluster defaultConfig
func (c *CloudInfoManager) InitCloudClusterDefaultInfo(cls *proto.Cluster,
	opt *cloudprovider.InitClusterConfigOption) error {
	if c == nil || cls == nil {
		return fmt.Errorf("%s InitCloudClusterDefaultInfo request is empty", cloudName)
	}

	if opt == nil || opt.Cloud == nil {
		return fmt.Errorf("%s InitCloudClusterDefaultInfo option is empty", cloudName)
	}

	if cls.GetClusterAdvanceSettings() == nil {
		return fmt.Errorf("initCloudCluster advanced setting empty")
	}

	// if !utils.StringInSlice(cls.ClusterAdvanceSettings.NetworkType,
	// 	[]string{common.AzureCniOverlay, common.AzureCniNodeSubnet}) {
	// 	return fmt.Errorf("initCloudCluster not supported networkPlugin[%s]", cls.ClusterAdvanceSettings.NetworkType)
	// }

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
	cluster, err := getCloudCluster(opt, opt.ImportMode.GetResourceGroup())
	if err != nil {
		return fmt.Errorf("SyncClusterCloudInfo failed: %v", err)
	}
	cls.SystemID = *cluster.Name
	// cluster cloud basic setting
	clusterBasicSettingByAzure(cls, cluster, opt)
	// cluster cloud advance setting
	clusterAdvanceSettingByAzure(cls, cluster)
	// cluster cloud network setting
	err = clusterNetworkSettingByAzure(cls, cluster)
	if err != nil {
		blog.Errorf("SyncClusterCloudInfo clusterNetworkSettingByAzure failed: %v", err)
	}

	return nil
}

// UpdateClusterCloudInfo update cluster info by cloud
func (c *CloudInfoManager) UpdateClusterCloudInfo(cls *proto.Cluster) error {
	// call azure interface to init cluster defaultConfig
	if c == nil || cls == nil {
		return fmt.Errorf("%s UpdateClusterCloudInfo request is empty", cloudName)
	}

	return nil
}

func getCloudCluster(opt *cloudprovider.SyncClusterCloudInfoOption,
	resourceGroupName string) (*armcontainerservice.ManagedCluster, error) {
	client, err := api.NewAksServiceImplWithCommonOption(opt.Common)
	if err != nil {
		return nil, fmt.Errorf("%s getCloudCluster NewContainerServiceClient failed: %v", cloudName, err)
	}
	mc, err := client.GetClusterWithName(context.Background(), resourceGroupName,
		opt.ImportMode.CloudID)
	if err != nil {
		return nil, fmt.Errorf("%s getCloudCluster GetCluster failed: %v", cloudName, err)
	}
	return mc, nil
}

func clusterBasicSettingByAzure(cls *proto.Cluster, cluster *armcontainerservice.ManagedCluster,
	opt *cloudprovider.SyncClusterCloudInfoOption) {
	clusterOs := ""
	if len(cluster.Properties.AgentPoolProfiles) > 0 {
		p := cluster.Properties.AgentPoolProfiles
		clusterOs = string(*p[0].OSSKU)
	}
	cls.ClusterBasicSettings = &proto.ClusterBasicSetting{
		OS:          clusterOs,
		Version:     *cluster.Properties.CurrentKubernetesVersion,
		VersionName: *cluster.Properties.CurrentKubernetesVersion,
		Area:        opt.Area,
	}
}

// clusterAdvanceSettingByAzure 同步集群高级设置
func clusterAdvanceSettingByAzure(cls *proto.Cluster, cluster *armcontainerservice.ManagedCluster) {
	cls.ClusterAdvanceSettings = &proto.ClusterAdvanceSetting{
		ClusterConnectSetting: &proto.ClusterConnectSetting{
			Internet: &proto.InternetAccessible{},
		},
	}

	profile := cluster.Properties.APIServerAccessProfile
	if profile != nil {
		internet := cls.ClusterAdvanceSettings.ClusterConnectSetting.Internet
		// 同步集群白名单
		if len(profile.AuthorizedIPRanges) > 0 {
			for _, cidr := range profile.AuthorizedIPRanges {
				internet.PublicAccessCidrs = append(internet.PublicAccessCidrs, *cidr)
			}
		}

		// 同步集群是否为公网集群
		if profile.EnablePrivateCluster == nil {
			cls.ClusterAdvanceSettings.ClusterConnectSetting.IsExtranet = true
		} else if *profile.EnablePrivateCluster {
			cls.ClusterAdvanceSettings.ClusterConnectSetting.IsExtranet = true
		}
	} else {
		cls.ClusterAdvanceSettings.ClusterConnectSetting.IsExtranet = true
	}

	// 同步集群网络插件
	networkProfile := cluster.Properties.NetworkProfile
	if networkProfile != nil {
		if networkProfile.NetworkPluginMode != nil &&
			*networkProfile.NetworkPluginMode == common.ClusterOverlayNetwork {
			cls.ClusterAdvanceSettings.NetworkType = common.AzureCniOverlay
			cls.NetworkType = common.ClusterOverlayNetwork
		} else {
			cls.ClusterAdvanceSettings.NetworkType = common.AzureCniNodeSubnet
			cls.NetworkType = common.ClusterUnderlayNetwork
		}
	}
}

func clusterNetworkSettingByAzure(cls *proto.Cluster, cluster *armcontainerservice.ManagedCluster) error { // nolint
	cidrs := cluster.Properties.NetworkProfile.PodCidrs
	podCidrs := make([]string, len(cidrs))
	for i := range cidrs {
		podCidrs[i] = *cidrs[i]
	}
	cls.NetworkSettings = &proto.NetworkSetting{
		ServiceIPv4CIDR:  *cluster.Properties.NetworkProfile.ServiceCidr,
		MultiClusterCIDR: podCidrs,
	}
	// 有时cluster不会返回pod cidr
	if cluster.Properties.NetworkProfile.PodCidr != nil {
		cls.NetworkSettings.ClusterIPv4CIDR = *cluster.Properties.NetworkProfile.PodCidr
	}

	return nil
}
