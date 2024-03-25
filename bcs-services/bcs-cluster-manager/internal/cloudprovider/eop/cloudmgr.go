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

// Package eop xxx
package eop

import (
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	putils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

var cloudInfoMgr sync.Once

func init() {
	cloudInfoMgr.Do(func() {
		// init Cluster
		cloudprovider.InitCloudInfoManager(cloudName, &CloudInfoManager{})
	})
}

// CloudInfoManager TKE management cluster info
type CloudInfoManager struct {
}

// InitCloudClusterDefaultInfo init cloud cluster default configInfo
func (c *CloudInfoManager) InitCloudClusterDefaultInfo(cls *cmproto.Cluster,
	opt *cloudprovider.InitClusterConfigOption) error {
	if c == nil || cls == nil {
		return fmt.Errorf("%s InitCloudClusterDefaultInfo request is empty", cloudName)
	}

	if opt == nil || opt.Cloud == nil {
		return fmt.Errorf("%s InitCloudClusterDefaultInfo option is empty", cloudName)
	}

	if len(cls.ManageType) == 0 {
		cls.ManageType = common.ClusterManageTypeIndependent
	}

	// cluster cloud basic setting
	clusterCloudDefaultBasicSetting(cls, opt.Cloud, opt.ClusterVersion)
	// cluster cloud advanced setting
	clusterCloudDefaultAdvancedSetting(cls, opt.Cloud, opt.ClusterVersion)

	if cls.NetworkSettings == nil {
		cls.NetworkSettings = &cmproto.NetworkSetting{}
	}

	return nil
}

// SyncClusterCloudInfo sync cluster metadata
func (c *CloudInfoManager) SyncClusterCloudInfo(cls *cmproto.Cluster,
	opt *cloudprovider.SyncClusterCloudInfoOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// UpdateClusterCloudInfo update cluster info by cloud
func (c *CloudInfoManager) UpdateClusterCloudInfo(cls *cmproto.Cluster) error {
	if c == nil || cls == nil {
		return fmt.Errorf("%s UpdateClusterCloudInfo request is empty", cloudName)
	}

	return nil
}

func clusterCloudDefaultBasicSetting(cls *cmproto.Cluster, cloud *cmproto.Cloud, version string) {
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

func clusterCloudDefaultAdvancedSetting(cls *cmproto.Cluster, cloud *cmproto.Cloud, version string) {
	runtimeInfo, err := putils.GetCloudDefaultRuntimeVersion(cloud, version)
	if err != nil {
		blog.Errorf("clusterCloudDefaultAdvancedSetting[%s] getCloudDefaultRuntimeInfo "+
			"failed: %v", cloud.CloudID, err)
		runtimeInfo = &cmproto.RunTimeInfo{
			ContainerRuntime: common.ContainerdRuntime,
			RuntimeVersion:   "1.6.9",
		}
	}

	if cls.ClusterAdvanceSettings == nil {
		cls.ClusterAdvanceSettings = &cmproto.ClusterAdvanceSetting{
			IPVS:             true,
			ContainerRuntime: runtimeInfo.ContainerRuntime,
			RuntimeVersion:   runtimeInfo.RuntimeVersion,
		}
	} else {
		if cls.ClusterAdvanceSettings.ContainerRuntime == "" {
			cls.ClusterAdvanceSettings.ContainerRuntime = runtimeInfo.ContainerRuntime
		}
		if cls.ClusterAdvanceSettings.RuntimeVersion == "" {
			cls.ClusterAdvanceSettings.RuntimeVersion = runtimeInfo.RuntimeVersion
		}
	}
}
