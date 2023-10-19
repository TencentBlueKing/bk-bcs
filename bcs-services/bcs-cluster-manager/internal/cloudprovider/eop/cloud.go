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

package eop

import (
	"fmt"
	"sync"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

var cloudInfoMgr sync.Once

func init() {
	cloudInfoMgr.Do(func() {
		//init Cluster
		cloudprovider.InitCloudInfoManager(cloudName, &CloudInfoManager{})
	})
}

// CloudInfoManager TKE management cluster info
type CloudInfoManager struct {
}

// InitCloudClusterDefaultInfo init cloud cluster default configInfo
func (c *CloudInfoManager) InitCloudClusterDefaultInfo(cls *cmproto.Cluster, opt *cloudprovider.InitClusterConfigOption) error {
	if c == nil || cls == nil {
		return fmt.Errorf("%s InitCloudClusterDefaultInfo request is empty", cloudName)
	}

	if opt == nil || opt.Cloud == nil {
		return fmt.Errorf("%s InitCloudClusterDefaultInfo option is empty", cloudName)
	}

	if len(cls.ManageType) == 0 {
		cls.ManageType = common.ClusterManageTypeIndependent
	}

	if cls.Environment == "" {
		cls.Environment = common.Prod
	}

	if cls.EngineType == "" {
		cls.EngineType = common.ClusterEngineTypeK8s
	}

	if !cls.IsExclusive {
		cls.IsExclusive = true
	}

	if cls.ClusterType == "" {
		cls.ClusterType = common.ClusterTypeSingle
	}

	if len(cls.Template) > 0 {
		cls.Template[0].NodeRole = "MASTER_ETCD"
	}
	// cluster cloud basic setting
	clusterCloudDefaultBasicSetting(cls, opt.Cloud, opt.ClusterVersion)
	// cluster cloud advanced setting
	clusterCloudDefaultAdvancedSetting(cls)

	if cls.NetworkSettings == nil {
		cls.NetworkSettings = &cmproto.NetworkSetting{
			MaxNodePodNum:   128,
			ClusterIPv4CIDR: "172.16.0.0/16",
			ServiceIPv4CIDR: "192.168.0.0/16",
		}
	} else {
		if cls.NetworkSettings.MaxNodePodNum == 0 {
			cls.NetworkSettings.MaxNodePodNum = 128
		}
		if cls.NetworkSettings.ClusterIPv4CIDR == "" {
			cls.NetworkSettings.ClusterIPv4CIDR = "172.16.0.0/16"
		}
		if cls.NetworkSettings.ServiceIPv4CIDR == "" {
			cls.NetworkSettings.ServiceIPv4CIDR = "192.168.0.0/16"
		}
	}

	return nil
}

// SyncClusterCloudInfo sync cluster metadata
func (c *CloudInfoManager) SyncClusterCloudInfo(cls *cmproto.Cluster, opt *cloudprovider.SyncClusterCloudInfoOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

func clusterCloudDefaultBasicSetting(cls *cmproto.Cluster, cloud *cmproto.Cloud, version string) {
	if version == "" {
		cls.ClusterBasicSettings.Version = "1.25.6"
	}

	defaultOSImage := common.DefaultECKImageName
	if len(cloud.OsManagement.AvailableVersion) > 0 {
		defaultOSImage = cloud.OsManagement.AvailableVersion[0]
	}

	if cls.ClusterBasicSettings == nil {
		cls.ClusterBasicSettings = &cmproto.ClusterBasicSetting{
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

func clusterCloudDefaultAdvancedSetting(cls *cmproto.Cluster) {
	if cls.ClusterAdvanceSettings == nil {
		cls.ClusterAdvanceSettings = &cmproto.ClusterAdvanceSetting{
			IPVS:             true,
			ContainerRuntime: "containerd",
			RuntimeVersion:   "1.6.9",
		}
	} else {
		cls.ClusterAdvanceSettings.IPVS = true
		if cls.ClusterAdvanceSettings.ContainerRuntime == "" {
			cls.ClusterAdvanceSettings.ContainerRuntime = "containerd"
		}
		if cls.ClusterAdvanceSettings.RuntimeVersion == "" {
			cls.ClusterAdvanceSettings.RuntimeVersion = "1.6.9"
		}
	}
}
