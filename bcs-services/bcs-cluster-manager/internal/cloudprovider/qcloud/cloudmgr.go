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

// Package qcloud xxx
package qcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	putils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
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

// InitCloudClusterDefaultInfo check importCluster operation
func (c *CloudInfoManager) InitCloudClusterDefaultInfo(cls *cmproto.Cluster,
	opt *cloudprovider.InitClusterConfigOption) error {
	// call qcloud interface to init cluster defaultConfig
	if c == nil || cls == nil {
		return fmt.Errorf("%s InitCloudClusterDefaultInfo request is empty", cloudName)
	}

	if opt == nil || opt.Cloud == nil {
		return fmt.Errorf("%s InitCloudClusterDefaultInfo option is empty", cloudName)
	}

	if len(cls.ManageType) == 0 {
		cls.ManageType = common.ClusterManageTypeIndependent
	}
	if len(cls.ClusterCategory) == 0 {
		cls.ClusterCategory = common.Builder
	}

	// cluster cloud basic setting
	clusterCloudDefaultBasicSetting(cls, opt.Cloud, opt.ClusterVersion)
	// cluster cloud advanced setting
	clusterCloudDefaultAdvancedSetting(cls, opt.Cloud, opt.ClusterVersion)
	// cluster cloud node setting
	clusterCloudDefaultNodeSetting(cls, true)
	// cluster cloud network setting
	err := clusterCloudNetworkSetting(cls)
	if err != nil {
		return err
	}

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
func (c *CloudInfoManager) SyncClusterCloudInfo(cls *cmproto.Cluster,
	opt *cloudprovider.SyncClusterCloudInfoOption) error {
	// call qcloud interface to init cluster defaultConfig
	if c == nil || cls == nil {
		return fmt.Errorf("%s SyncClusterCloudInfo request is empty", cloudName)
	}

	if opt == nil || opt.Cloud == nil {
		return fmt.Errorf("%s SyncClusterCloudInfo option is empty", cloudName)
	}

	// get cloud cluster
	tkeCluster, masterNodes, err := getCloudClusterInfo(opt)
	if err != nil {
		return fmt.Errorf("SyncClusterCloudInfo failed: %v", err)
	}

	cls.SystemID = *tkeCluster.ClusterId
	cls.VpcID = *tkeCluster.ClusterNetworkSettings.VpcId
	cls.Master = masterNodes
	cls.ManageType = *tkeCluster.ClusterType

	// cluster cloud basic setting
	clusterBasicSettingByQCloud(cls, tkeCluster)
	// cluster cloud node setting
	clusterCloudDefaultNodeSetting(cls, false)
	// cluster cloud advanced setting
	clusterAdvancedSettingByQCloud(cls, tkeCluster)

	// cluster cloud network setting
	err = clusterNetworkSettingByQCloud(cls, tkeCluster)
	if err != nil {
		blog.Errorf("SyncClusterCloudInfo clusterNetworkSettingByQCloud failed: %v", err)
	}

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

func getCloudClusterInfo(opt *cloudprovider.SyncClusterCloudInfoOption) (
	*tke.Cluster, map[string]*cmproto.Node, error) {
	var (
		cloudID = opt.ImportMode.CloudID
		err     error
	)
	if cloudID == "" {
		cloudID, err = getCloudIDByKubeConfig(opt)
		if err != nil {
			return nil, nil, err
		}
	}
	tkeCluster, err := getTkeCluster(opt, cloudID)
	if err != nil {
		return nil, nil, err
	}

	switch *tkeCluster.ClusterType {
	case common.ClusterManageTypeManaged:
		return tkeCluster, nil, nil
	default:
	}

	masterNodes, err := getClusterMasterNodes(opt, tkeCluster)
	if err != nil {
		return nil, nil, err
	}

	return tkeCluster, masterNodes, nil
}

func getTkeCluster(opt *cloudprovider.SyncClusterCloudInfoOption, cloudID string) (*tke.Cluster, error) {
	tkeCli, err := api.NewTkeClient(opt.Common)
	if err != nil {
		return nil, err
	}

	return tkeCli.GetTKECluster(cloudID)
}

func getClusterMasterNodes(opt *cloudprovider.SyncClusterCloudInfoOption,
	cluster *tke.Cluster) (map[string]*cmproto.Node, error) {
	tkeCli, err := api.NewTkeClient(opt.Common)
	if err != nil {
		return nil, err
	}

	instancesList, err := tkeCli.QueryTkeClusterAllInstances(context.Background(), *cluster.ClusterId, nil)
	if err != nil {
		return nil, err
	}

	var (
		masterIPs = make([]string, 0)
	)
	for _, ins := range instancesList {
		switch ins.InstanceRole {
		case api.MASTER_ETCD.String():
			masterIPs = append(masterIPs, ins.InstanceIP)
		default:
			continue
		}
	}

	masterNodes := make(map[string]*cmproto.Node)
	nodes, err := transInstanceIPToNodes(masterIPs, &cloudprovider.ListNodesOption{
		Common:       opt.Common,
		ClusterVPCID: *cluster.ClusterNetworkSettings.VpcId,
	})
	if err != nil {
		return nil, err
	}
	for _, node := range nodes {
		node.Status = common.StatusRunning
		masterNodes[node.InnerIP] = node
	}

	return masterNodes, nil
}

func transInstanceIPToNodes(ipList []string, opt *cloudprovider.ListNodesOption) ([]*cmproto.Node, error) {
	nodeMgr := NodeManager{}
	nodes, err := nodeMgr.ListNodesByIP(ipList, &cloudprovider.ListNodesOption{
		Common:       opt.Common,
		ClusterVPCID: opt.ClusterVPCID,
	})
	if err != nil {
		return nil, err
	}

	return nodes, nil
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
		RuntimeVersion: func() string {
			if cluster != nil && cluster.RuntimeVersion != nil {
				return *cluster.RuntimeVersion
			}
			return common.DockerRuntimeVersion
		}(),
		ExtraArgs: common.DefaultClusterConfig,
	}
}

func clusterBasicSettingByQCloud(cls *cmproto.Cluster, cluster *tke.Cluster) {
	cls.ClusterBasicSettings = &cmproto.ClusterBasicSetting{
		OS:                        *cluster.ClusterOs,
		Version:                   *cluster.ClusterVersion,
		VersionName:               *cluster.ClusterVersion,
		ClusterLevel:              *cluster.ClusterLevel,
		IsAutoUpgradeClusterLevel: *cluster.AutoUpgradeClusterLevel,
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

func clusterCloudDefaultAdvancedSetting(cls *cmproto.Cluster, cloud *cmproto.Cloud, version string) {
	runtimeInfo, err := putils.GetCloudDefaultRuntimeVersion(cloud, version)
	if err != nil {
		blog.Errorf("clusterCloudDefaultAdvancedSetting[%s] getCloudDefaultRuntimeInfo "+
			"failed: %v", cloud.CloudID, err)
		runtimeInfo = &cmproto.RunTimeInfo{
			ContainerRuntime: common.DockerContainerRuntime,
			RuntimeVersion:   common.DockerRuntimeVersion,
		}
	}

	if cls.ClusterAdvanceSettings == nil {
		cls.ClusterAdvanceSettings = &cmproto.ClusterAdvanceSetting{
			IPVS:             true,
			ContainerRuntime: runtimeInfo.ContainerRuntime,
			RuntimeVersion:   runtimeInfo.RuntimeVersion,
			ExtraArgs:        common.DefaultClusterConfig,
		}
	} else {
		if cls.ClusterAdvanceSettings.ContainerRuntime == "" {
			cls.ClusterAdvanceSettings.ContainerRuntime = runtimeInfo.ContainerRuntime
		}
		if cls.ClusterAdvanceSettings.RuntimeVersion == "" {
			cls.ClusterAdvanceSettings.RuntimeVersion = runtimeInfo.RuntimeVersion
		}
		if cls.ClusterAdvanceSettings.ExtraArgs == nil {
			cls.ClusterAdvanceSettings.ExtraArgs = common.DefaultClusterConfig
		}
	}
}

func clusterCloudNetworkSetting(cls *cmproto.Cluster) error {
	if cls.GetNetworkSettings() == nil {
		return fmt.Errorf("initCloudCluster network setting empty")
	}

	if cls.GetNetworkSettings().GetEnableVPCCni() {
		if cls.NetworkSettings.SubnetSource == nil ||
			(len(cls.NetworkSettings.SubnetSource.New) == 0 && cls.NetworkSettings.SubnetSource.Existed == nil) {
			return fmt.Errorf("network[%s] subnet resource empty", common.VpcCni)
		}

		// 固定IP模式下, IP默认回收时间
		if cls.NetworkSettings.IsStaticIpMode && cls.NetworkSettings.ClaimExpiredSeconds < 300 {
			cls.NetworkSettings.ClaimExpiredSeconds = 300
		}
	}

	return nil
}

func clusterCloudDefaultNodeSetting(cls *cmproto.Cluster, defaultNodeConfig bool) {
	if cls.NodeSettings == nil {
		cls.NodeSettings = &cmproto.NodeSetting{
			DockerGraphPath: common.DockerGraphPath,
			MountTarget:     common.MountTarget,
			UnSchedulable:   1,
		}
		if defaultNodeConfig {
			cls.NodeSettings.ExtraArgs = common.DefaultNodeConfig
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
		if cls.ClusterAdvanceSettings.ExtraArgs == nil && defaultNodeConfig {
			cls.NodeSettings.ExtraArgs = common.DefaultNodeConfig
		}
	}
}

func clusterCloudDefaultBasicSetting(cls *cmproto.Cluster, cloud *cmproto.Cloud, version string) {
	defaultOSImage := common.DefaultImageName
	// platform default image name
	if len(cloud.OsManagement.AvailableVersion) > 0 {
		defaultOSImage = cloud.OsManagement.AvailableVersion[0]
	}

	if version == "" {
		version = cloud.ClusterManagement.AvailableVersion[0]
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
