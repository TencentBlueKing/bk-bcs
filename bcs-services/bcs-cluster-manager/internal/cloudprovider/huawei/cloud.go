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

// Package huawei xxx
package huawei

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
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
	client, err := api.NewCceClient(opt.Common)
	if err != nil {
		return err
	}

	cluster, err := client.GetCceCluster(opt.ImportMode.CloudID)
	if err != nil {
		return err
	}

	kubeConfig, err := api.GetClusterKubeConfig(client, opt.ImportMode.CloudID)
	if err != nil {
		return fmt.Errorf("SyncClusterCloudInfo GetClusterKubeConfig failed: %v", err)
	}

	cls.KubeConfig = kubeConfig
	cls.SystemID = cluster.Metadata.Name
	cls.VpcID = cluster.Spec.HostNetwork.Vpc

	// cluster cloud basic setting
	clusterBasicSettingByCCE(cls, cluster)

	// cluster cloud network setting
	err = clusterNetworkSettingByCCE(cls, cluster)
	if err != nil {
		blog.Errorf("SyncClusterCloudInfo clusterNetworkSettingByCCE failed: %v", err)
	}

	return nil
}

// UpdateClusterCloudInfo update cluster info by cloud
func (c *CloudInfoManager) UpdateClusterCloudInfo(cls *cmproto.Cluster) error {
	return nil
}

func clusterBasicSettingByCCE(cls *cmproto.Cluster, cluster *model.ShowClusterResponse) {
	cls.ClusterBasicSettings = &cmproto.ClusterBasicSetting{}

	if cluster.Spec != nil {
		cls.ClusterBasicSettings.Version = *cluster.Spec.Version
		cls.ClusterBasicSettings.VersionName = *cluster.Spec.Version

		if cluster.Spec.Type != nil {
			cls.ClusterBasicSettings.OS = cluster.Spec.Type.Value()
		}
	}
}

func clusterNetworkSettingByCCE(cls *cmproto.Cluster, cluster *model.ShowClusterResponse) error {
	cls.NetworkSettings = &cmproto.NetworkSetting{}

	if cluster.Spec != nil {
		if cluster.Spec.ContainerNetwork != nil {
			if cluster.Spec.ContainerNetwork.Cidr != nil {
				cls.NetworkSettings = &cmproto.NetworkSetting{
					ClusterIPv4CIDR: *cluster.Spec.ContainerNetwork.Cidr,
					ServiceIPv4CIDR: *cluster.Spec.ContainerNetwork.Cidr,
				}
			}

			if cluster.Spec.ContainerNetwork.Cidrs != nil && len(*cluster.Spec.ContainerNetwork.Cidrs) > 0 {
				cls.NetworkSettings.ClusterIPv4CIDR = (*cluster.Spec.ContainerNetwork.Cidrs)[0].Cidr
				cls.NetworkSettings.ServiceIPv4CIDR = (*cluster.Spec.ContainerNetwork.Cidrs)[0].Cidr
			}
		}

		if cluster.Spec.ExtendParam != nil && cluster.Spec.ExtendParam.AlphaCceFixPoolMask != nil {
			num, err := strconv.ParseInt(*cluster.Spec.ExtendParam.AlphaCceFixPoolMask, 10, 32)
			if err != nil {
				return err
			}

			cls.NetworkSettings.MaxNodePodNum = uint32(num)
		}
	}

	return nil
}
