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

// Package google xxx
package google

import (
	"context"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/encrypt"
	"strings"
	"sync"

	container "google.golang.org/api/container/v1"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/google/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
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
	cluster, err := getCloudCluster(opt, opt.ImportMode.CloudID)
	if err != nil {
		return fmt.Errorf("SyncClusterCloudInfo failed: %v", err)
	}
	cls.SystemID = cluster.Name

	cls.VpcID = cluster.Network
	// 记录gke集群发布类型
	if cluster.ReleaseChannel != nil {
		if cls.ExtraInfo == nil {
			cls.ExtraInfo = make(map[string]string, 0)
		}
		cls.ExtraInfo[api.GKEClusterReleaseChannel] = cluster.ReleaseChannel.Channel
	}

	// gke集群 region级别 zone级别
	clusterType := common.Regions
	if len(strings.Split(opt.Common.Region, "-")) == 3 {
		clusterType = common.Zones
	}
	cls.ExtraInfo[api.GKEClusterLocationType] = clusterType
	cls.ExtraInfo[api.GKEClusterLocations] = strings.Join(cluster.Locations, ",")

	kubeConfig, err := api.GetClusterKubeConfig(context.Background(), opt.Common.Account.ServiceAccountSecret,
		opt.Common.Account.GkeProjectID, cls.Region, clusterType, cls.SystemID)
	if err != nil {
		return fmt.Errorf("SyncClusterCloudInfo GetClusterKubeConfig failed: %v", err)
	}
	cls.KubeConfig, _ = encrypt.Encrypt(nil, kubeConfig)

	// cluster cloud basic setting
	clusterBasicSettingByGKE(cls, cluster, opt)
	// cluster cloud network setting
	clusterNetworkSettingByGKE(cls, cluster)
	// cluster cloud advanced setting
	clusterAdvanceSettingByGKE(cls, cluster)

	return nil
}

// UpdateClusterCloudInfo update cluster info by cloud
func (c *CloudInfoManager) UpdateClusterCloudInfo(cls *cmproto.Cluster) error {
	// call qcloud interface to init cluster defaultConfig
	if c == nil || cls == nil {
		return fmt.Errorf("%s UpdateClusterCloudInfo request is empty", cloudName)
	}

	return nil
}

func getCloudCluster(opt *cloudprovider.SyncClusterCloudInfoOption, clusterName string) (
	*container.Cluster, error) {
	cli, err := api.NewContainerServiceClient(opt.Common)
	if err != nil {
		return nil, fmt.Errorf("%s getCloudCluster NewContainerServiceClient failed: %v", cloudName, err)
	}
	cluster, err := cli.GetCluster(context.Background(), clusterName)
	if err != nil {
		return nil, fmt.Errorf("%s getCloudCluster GetCluster failed: %v", cloudName, err)
	}
	return cluster, nil
}

func clusterBasicSettingByGKE(cls *cmproto.Cluster, cluster *container.Cluster,
	opt *cloudprovider.SyncClusterCloudInfoOption) {
	cls.ClusterBasicSettings = &cmproto.ClusterBasicSetting{
		Version:     cluster.CurrentMasterVersion,
		VersionName: cluster.CurrentMasterVersion,
		Area:        opt.Area,
	}
	if cluster.NodeConfig != nil {
		cls.ClusterBasicSettings.OS = cluster.NodeConfig.ImageType
	}
}

func clusterNetworkSettingByGKE(cls *cmproto.Cluster, cluster *container.Cluster) {
	cls.NetworkSettings = &cmproto.NetworkSetting{
		ClusterIPv4CIDR: cluster.ClusterIpv4Cidr,
		ServiceIPv4CIDR: cluster.ServicesIpv4Cidr,
	}
	if cluster.DefaultMaxPodsConstraint != nil {
		cls.NetworkSettings.MaxNodePodNum = uint32(cluster.DefaultMaxPodsConstraint.MaxPodsPerNode)
	}
}

func clusterAdvanceSettingByGKE(cls *cmproto.Cluster, cluster *container.Cluster) { // nolint
	cls.ClusterAdvanceSettings = &cmproto.ClusterAdvanceSetting{
		IPVS: true,
	}
}
