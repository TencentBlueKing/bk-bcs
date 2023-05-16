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

package google

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/google/api"

	"google.golang.org/api/container/v1"
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
	cls.VpcID = cluster.NetworkConfig.Network
	// 记录gke集群发布类型
	if cluster.ReleaseChannel != nil {
		if cls.ExtraInfo == nil {
			cls.ExtraInfo = make(map[string]string, 0)
		}
		cls.ExtraInfo["releaseChannel"] = cluster.ReleaseChannel.Channel
	}
	// 区分gke集群是zone级别还是region级别
	if len(strings.Split(cluster.Location, "-")) == 2 {
		cls.ExtraInfo["locationType"] = "regions"
	} else if len(strings.Split(cluster.Location, "-")) == 3 {
		cls.ExtraInfo["locationType"] = "zones"
	}
	kubeConfig, err := api.GetClusterKubeConfig(context.Background(), opt.Common.Account.ServiceAccountSecret,
		opt.Common.Account.GkeProjectID, cls.Region, cls.SystemID)
	if err != nil {
		return fmt.Errorf("SyncClusterCloudInfo GetClusterKubeConfig failed: %v", err)
	}
	cls.KubeConfig = kubeConfig
	// cluster cloud basic setting
	clusterBasicSettingByGKE(cls, cluster)

	// cluster cloud network setting
	err = clusterNetworkSettingByGKE(cls, cluster)
	if err != nil {
		blog.Errorf("SyncClusterCloudInfo clusterNetworkSettingByGKE failed: %v", err)
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

func clusterBasicSettingByGKE(cls *cmproto.Cluster, cluster *container.Cluster) {
	cls.ClusterBasicSettings = &cmproto.ClusterBasicSetting{
		Version:     cluster.CurrentMasterVersion,
		VersionName: cluster.CurrentMasterVersion,
	}
	if cluster.NodeConfig != nil {
		cls.ClusterBasicSettings.OS = cluster.NodeConfig.ImageType
	}
}

func clusterNetworkSettingByGKE(cls *cmproto.Cluster, cluster *container.Cluster) error {
	cls.NetworkSettings = &cmproto.NetworkSetting{
		ClusterIPv4CIDR: cluster.ClusterIpv4Cidr,
		ServiceIPv4CIDR: cluster.ServicesIpv4Cidr,
	}
	if cluster.DefaultMaxPodsConstraint != nil {
		cls.NetworkSettings.MaxNodePodNum = uint32(cluster.DefaultMaxPodsConstraint.MaxPodsPerNode)
	}

	return nil
}
