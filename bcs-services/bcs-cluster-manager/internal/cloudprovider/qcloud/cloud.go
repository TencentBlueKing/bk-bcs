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

package qcloud

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"
)

var cloudInfoMgr sync.Once

func init() {
	cloudInfoMgr.Do(func() {
		//init Cluster
		cloudprovider.InitCloudInfoManager(cloudName, &CloudInfoManager{})
	})
}

// CloudInfoManager blueKingCloud management cluster info
type CloudInfoManager struct {
}

// ImportClusterValidate check importCluster operation
func (c *CloudInfoManager) InitCloudClusterDefaultInfo(cls *cmproto.Cluster, opt *cloudprovider.InitClusterConfigOption) error {
	// call qcloud interface to init cluster defaultConfig
	if c == nil || cls == nil {
		return fmt.Errorf("%s InitCloudClusterDefaultInfo request is empty", cloudName)
	}

	if opt == nil || opt.Cloud == nil {
		return fmt.Errorf("%s InitCloudClusterDefaultInfo option is empty", cloudName)
	}

	// cluster cloud advanced setting
	clusterCloudDefaultAdvancedSetting(cls)
	// cluster cloud node setting
	clusterCloudDefaultNodeSetting(cls)
	// cluster cloud basic setting
	clusterCloudDefaultBasicSetting(cls, opt.Cloud, opt.ClusterVersion)

	if cls.NetworkSettings.CidrStep <= 0 {
		switch cls.Environment {
		case common.Prod:
			cls.NetworkSettings.CidrStep = 4096
		default:
			cls.NetworkSettings.CidrStep = 2048
		}
	}

	return nil
}

// SyncClusterCloudInfo get cluster cloudInfo by clusterID or kubeConfig
func (c *CloudInfoManager) SyncClusterCloudInfo(cls *cmproto.Cluster, opt *cloudprovider.SyncClusterCloudInfoOption) error {
	// call qcloud interface to init cluster defaultConfig
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
	cls.SystemID = *cluster.ClusterId
	cls.VpcID = *cluster.ClusterNetworkSettings.VpcId

	// cluster cloud basic setting
	clusterBasicSettingByQCloud(cls, cluster)
	// cluster cloud node setting
	clusterCloudDefaultNodeSetting(cls)
	// cluster cloud advanced setting
	clusterAdvancedSettingByQCloud(cls, cluster)

	// cluster cloud network setting
	err = clusterNetworkSettingByQCloud(cls, cluster)
	if err != nil {
		blog.Errorf("SyncClusterCloudInfo clusterNetworkSettingByQCloud failed: %v", err)
	}

	return nil
}

func getCloudCluster(opt *cloudprovider.SyncClusterCloudInfoOption) (*tke.Cluster, error) {
	var (
		cloudID = opt.ImportMode.CloudID
		err     error
	)
	if cloudID == "" {
		cloudID, err = getCloudIDByKubeConfig(opt)
		if err != nil {
			return nil, err
		}
	}

	cli, err := api.NewTkeClient(opt.Common)
	if err != nil {
		return nil, err
	}

	return cli.GetTKECluster(cloudID)
}

// kubeConfig cluster name must be cloud clusterID
func getCloudIDByKubeConfig(opt *cloudprovider.SyncClusterCloudInfoOption) (string, error) {
	config, err := types.GetKubeConfigFromYAMLBody(false, types.YamlInput{
		FileName:    "",
		YamlContent: opt.ImportMode.KubeConfig,
	})
	if err != nil {
		return "", fmt.Errorf("%s getCloudIDByKubeConfig GetKubeConfigFromYAMLBody failed: %v", cloudName, err)
	}
	return config.Clusters[0].Name, nil
}

func clusterAdvancedSettingByQCloud(cls *cmproto.Cluster, cluster *tke.Cluster) {
	cls.ClusterAdvanceSettings = &cmproto.ClusterAdvanceSetting{
		IPVS:             *cluster.ClusterNetworkSettings.Ipvs,
		ContainerRuntime: *cluster.ContainerRuntime,
		RuntimeVersion:   common.DockerRuntimeVersion,
		ExtraArgs:        common.DefaultClusterConfig,
	}
}

func clusterBasicSettingByQCloud(cls *cmproto.Cluster, cluster *tke.Cluster) {
	cls.ClusterBasicSettings = &cmproto.ClusterBasicSetting{
		OS:          *cluster.ClusterOs,
		Version:     *cluster.ClusterVersion,
		VersionName: *cluster.ClusterVersion,
	}
}

func clusterNetworkSettingByQCloud(cls *cmproto.Cluster, cluster *tke.Cluster) error {
	property := *cluster.Property
	propertyInfo := make(map[string]interface{})
	err := json.Unmarshal([]byte(property), &propertyInfo)
	if err != nil {
		return err
	}

	var multiCIDRList []string
	if v, ok := propertyInfo["EnableMultiClusterCIDR"]; ok && v.(bool) {
		multiCIDRs := propertyInfo["MultiClusterCIDR"]
		multiCIDRList = strings.Split(multiCIDRs.(string), ",")
	}

	masterCIDR := *cluster.ClusterNetworkSettings.ClusterCIDR
	step, err := utils.ConvertCIDRToStep(masterCIDR)
	if err != nil {
		return err
	}

	cls.NetworkSettings = &cmproto.NetworkSetting{
		ClusterIPv4CIDR:  masterCIDR,
		MaxNodePodNum:    uint32(*cluster.ClusterNetworkSettings.MaxNodePodNum),
		MaxServiceNum:    uint32(*cluster.ClusterNetworkSettings.MaxClusterServiceNum),
		MultiClusterCIDR: multiCIDRList,
		CidrStep:         step,
	}

	return nil
}

func clusterCloudDefaultAdvancedSetting(cls *cmproto.Cluster) {
	if cls.ClusterAdvanceSettings == nil {
		cls.ClusterAdvanceSettings = &cmproto.ClusterAdvanceSetting{
			IPVS:             true,
			ContainerRuntime: common.DockerContainerRuntime,
			RuntimeVersion:   common.DockerRuntimeVersion,
			ExtraArgs:        common.DefaultClusterConfig,
		}

	} else {
		if cls.ClusterAdvanceSettings.ContainerRuntime == "" {
			cls.ClusterAdvanceSettings.ContainerRuntime = common.DockerContainerRuntime
		}
		if cls.ClusterAdvanceSettings.RuntimeVersion == "" {
			cls.ClusterAdvanceSettings.RuntimeVersion = common.DockerRuntimeVersion
		}
		if cls.ClusterAdvanceSettings.ExtraArgs == nil {
			cls.ClusterAdvanceSettings.ExtraArgs = common.DefaultClusterConfig
		}
	}
}

func clusterCloudDefaultNodeSetting(cls *cmproto.Cluster) {
	if cls.NodeSettings == nil {
		cls.NodeSettings = &cmproto.NodeSetting{
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

func clusterCloudDefaultBasicSetting(cls *cmproto.Cluster, cloud *cmproto.Cloud, version string) {
	defaultOSImage := common.DefaultImageName
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
