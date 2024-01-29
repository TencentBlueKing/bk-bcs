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

// Package blueking xxx
package blueking

import (
	"fmt"
	"sync"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var cloudInfoMgr sync.Once

func init() {
	cloudInfoMgr.Do(func() {
		// init Cluster
		cloudprovider.InitCloudInfoManager(cloudName, &CloudInfoManager{})
	})
}

// CloudInfoManager blueKingCloud management cluster info
type CloudInfoManager struct {
}

// InitCloudClusterDefaultInfo check importCluster operation
func (c *CloudInfoManager) InitCloudClusterDefaultInfo(cls *proto.Cluster,
	opt *cloudprovider.InitClusterConfigOption) error {
	// call blueking interface to init cluster defaultConfig
	if c == nil || cls == nil {
		return fmt.Errorf("%s InitCloudClusterDefaultInfo request is empty", cloudName)
	}

	if opt == nil || opt.Cloud == nil {
		return fmt.Errorf("%s InitCloudClusterDefaultInfo option is empty", cloudName)
	}

	// cluster node setting
	clusterCloudDefaultNodeSetting(cls)
	// cluster basic setting
	clusterCloudDefaultBasicSetting(cls, opt.Cloud, opt.ClusterVersion)

	return nil
}

// SyncClusterCloudInfo get cluster cloudInfo by clusterID or kubeConfig
func (c *CloudInfoManager) SyncClusterCloudInfo(cls *proto.Cluster,
	opt *cloudprovider.SyncClusterCloudInfoOption) error {
	// call blueking interface to init cluster defaultConfig
	if c == nil || cls == nil {
		return fmt.Errorf("%s SyncClusterCloudInfo request is empty", cloudName)
	}

	if opt == nil || opt.Cloud == nil {
		return fmt.Errorf("%s SyncClusterCloudInfo option is empty", cloudName)
	}

	// cluster cloud basic setting
	clusterCloudDefaultBasicSetting(cls, opt.Cloud, opt.ClusterVersion)
	// cluster cloud node setting
	clusterCloudDefaultNodeSetting(cls)
	// cluster cloud default advanced setting
	clusterCloudAdvancedSetting(cls)
	// cluster cloud default networking setting
	err := clusterCloudNetworkSetting(cls)
	if err != nil {
		return err
	}

	return nil
}

// UpdateClusterCloudInfo update cluster info by cloud
func (c *CloudInfoManager) UpdateClusterCloudInfo(cls *proto.Cluster) error {
	// call qcloud interface to init cluster defaultConfig
	if c == nil || cls == nil {
		return fmt.Errorf("%s UpdateClusterCloudInfo request is empty", cloudName)
	}

	return nil
}

func clusterCloudNetworkSetting(cls *proto.Cluster) error {
	if cls == nil || cls.NetworkSettings == nil {
		return nil
	}

	switch cls.NetworkSettings.ClusterIpType {
	case utils.IPV4:
		if cls.NetworkSettings.ClusterIPv4CIDR == "" || cls.NetworkSettings.ServiceIPv4CIDR == "" {
			return fmt.Errorf("networkIpv4Type[%s] cidr empty", cls.NetworkSettings.ClusterIpType)
		}

		cls.NetworkSettings.CidrStep, _ = utils.ConvertCIDRToStep(cls.NetworkSettings.ClusterIPv4CIDR)
	case utils.IPV6:
		if cls.NetworkSettings.ClusterIPv6CIDR == "" || cls.NetworkSettings.ServiceIPv6CIDR == "" {
			return fmt.Errorf("networkIpv6Type[%s] cidr empty", cls.NetworkSettings.ClusterIpType)
		}

		cls.NetworkSettings.CidrStep, _ = utils.ConvertCIDRToStep(cls.NetworkSettings.ClusterIPv6CIDR)
	case utils.DualStack:
		if cls.NetworkSettings.ClusterIPv4CIDR == "" || cls.NetworkSettings.ServiceIPv4CIDR == "" ||
			cls.NetworkSettings.ClusterIPv6CIDR == "" || cls.NetworkSettings.ServiceIPv6CIDR == "" {
			return fmt.Errorf("networkIpv6Type[%s] cidr empty", cls.NetworkSettings.ClusterIpType)
		}

		cls.NetworkSettings.CidrStep, _ = utils.ConvertCIDRToStep(cls.NetworkSettings.ClusterIPv4CIDR)
	default:
		return fmt.Errorf("not supported networkIpType[%s]", cls.NetworkSettings.ClusterIPv4CIDR)
	}

	return nil
}

func clusterCloudAdvancedSetting(cls *proto.Cluster) {
	if cls == nil || cls.ClusterAdvanceSettings == nil {
		return
	}

	if cls.ClusterAdvanceSettings.ContainerRuntime == "" {
		cls.ClusterAdvanceSettings.ContainerRuntime = common.DockerContainerRuntime
		cls.ClusterAdvanceSettings.RuntimeVersion = common.DockerRuntimeSelfVersion
	}
	if cls.ClusterAdvanceSettings.NetworkType == "" {
		cls.ClusterAdvanceSettings.NetworkType = common.Flannel
	}
}

func clusterCloudDefaultNodeSetting(cls *proto.Cluster) {
	if cls.NodeSettings == nil {
		cls.NodeSettings = &proto.NodeSetting{
			DockerGraphPath: common.DockerGraphPath,
			MountTarget:     common.MountTarget,
			UnSchedulable:   1,
		}
	} else {
		if cls.NodeSettings.DockerGraphPath == "" {
			cls.NodeSettings.DockerGraphPath = common.DockerGraphPath
		}
		if cls.NodeSettings.MountTarget == "" {
			cls.NodeSettings.MountTarget = common.MountTarget
		}
		if cls.NodeSettings.UnSchedulable == 0 {
			cls.NodeSettings.UnSchedulable = 1
		}
	}
}

func clusterCloudDefaultBasicSetting(cls *proto.Cluster, cloud *proto.Cloud, version string) {
	defaultOSImage := common.DefaultImageName
	if len(cloud.OsManagement.AvailableVersion) > 0 {
		defaultOSImage = cloud.OsManagement.AvailableVersion[0]
	}
	if version == "" && len(cloud.ClusterManagement.AvailableVersion) > 0 {
		version = cloud.ClusterManagement.AvailableVersion[0]
	}

	if cls.ClusterBasicSettings == nil {
		cls.ClusterBasicSettings = &proto.ClusterBasicSetting{
			OS:          defaultOSImage,
			Version:     version,
			VersionName: version,
		}
	} else {
		if cls.ClusterBasicSettings.OS == "" {
			cls.ClusterBasicSettings.OS = defaultOSImage
		}
		cls.ClusterBasicSettings.Version = version
		cls.ClusterBasicSettings.VersionName = version
	}
}
