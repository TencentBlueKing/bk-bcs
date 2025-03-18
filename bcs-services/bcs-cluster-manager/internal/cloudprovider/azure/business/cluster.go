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

// Package business xxx
package business

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// SyncClusterInfo sync cluster info
func SyncClusterInfo(opt *cloudprovider.GetClusterOption) error {
	if opt.Cluster.ClusterAdvanceSettings == nil {
		opt.Cluster.ClusterAdvanceSettings = &proto.ClusterAdvanceSetting{
			ClusterConnectSetting: &proto.ClusterConnectSetting{
				Internet: &proto.InternetAccessible{},
			},
		}
	} else if opt.Cluster.ClusterAdvanceSettings.ClusterConnectSetting == nil {
		opt.Cluster.ClusterAdvanceSettings.ClusterConnectSetting = &proto.ClusterConnectSetting{
			Internet: &proto.InternetAccessible{},
		}
	} else if opt.Cluster.ClusterAdvanceSettings.ClusterConnectSetting.Internet == nil {
		opt.Cluster.ClusterAdvanceSettings.ClusterConnectSetting.Internet = &proto.InternetAccessible{}
	}

	client, err := api.NewAksServiceImplWithCommonOption(&opt.CommonOption)
	if err != nil {
		return fmt.Errorf("init AksService failed, %v", err)
	}

	cloudCluster, err := client.GetClusterWithName(context.Background(),
		cloudprovider.GetClusterResourceGroup(opt.Cluster), opt.Cluster.SystemID)
	if err != nil {
		return err
	}

	// 同步集群是否为公网
	profile := cloudCluster.Properties.APIServerAccessProfile
	if profile == nil || profile.EnablePrivateCluster == nil || !*profile.EnablePrivateCluster {
		opt.Cluster.ClusterAdvanceSettings.ClusterConnectSetting.IsExtranet = true
	} else {
		opt.Cluster.ClusterAdvanceSettings.ClusterConnectSetting.IsExtranet = false
	}

	// 同步集群白名单
	internet := opt.Cluster.ClusterAdvanceSettings.ClusterConnectSetting.Internet
	internet.PublicAccessCidrs = []string{}
	if profile != nil && len(profile.AuthorizedIPRanges) > 0 {
		for _, ipRange := range profile.AuthorizedIPRanges {
			internet.PublicAccessCidrs = append(internet.PublicAccessCidrs, *ipRange)
		}
	}

	// 同步集群网络插件
	networkProfile := cloudCluster.Properties.NetworkProfile
	if networkProfile != nil {
		if networkProfile.NetworkPluginMode != nil &&
			*networkProfile.NetworkPluginMode == common.ClusterOverlayNetwork {
			opt.Cluster.ClusterAdvanceSettings.NetworkType = common.AzureCniOverlay
			opt.Cluster.NetworkType = common.ClusterOverlayNetwork
		} else {
			opt.Cluster.ClusterAdvanceSettings.NetworkType = common.AzureCniNodeSubnet
			opt.Cluster.NetworkType = common.ClusterUnderlayNetwork
		}
	}

	return nil
}
