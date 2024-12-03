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

package qcloud

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n"
	"github.com/avast/retry-go"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/tasks"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var clsMgr sync.Once

func init() {
	clsMgr.Do(func() {
		// init Node
		cloudprovider.InitClusterManager(cloudName, &Cluster{})
	})
}

// Cluster tke management implementation
type Cluster struct {
}

// build task or handle data

// CreateCluster create kubenretes cluster according cloudprovider
func (c *Cluster) CreateCluster(cls *proto.Cluster, opt *cloudprovider.CreateClusterOption) (*proto.Task, error) {
	// call qcloud interface to create cluster
	if cls == nil {
		return nil, fmt.Errorf("qcloud CreateCluster cluster is empty")
	}

	if opt == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("qcloud CreateCluster cluster opt or cloud is empty")
	}

	if opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("qcloud CreateCluster opt lost valid crendential info")
	}

	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when CreateCluster %d failed, %s",
			opt.Cloud.CloudID, cls.ClusterName, err.Error(),
		)
		return nil, err
	}

	// build create cluster task
	task, err := mgr.BuildCreateClusterTask(cls, opt)
	if err != nil {
		blog.Errorf("build CreateCluster task for cluster %s with cloudprovider %s failed, %s",
			cls.ClusterName, cls.Provider, err.Error(),
		)
		return nil, err
	}

	return task, nil
}

// CreateVirtualCluster create virtual cluster by cloud provider
func (c *Cluster) CreateVirtualCluster(cls *proto.Cluster,
	opt *cloudprovider.CreateVirtualClusterOption) (*proto.Task, error) {
	if cls == nil {
		return nil, fmt.Errorf("qcloud CreateVirtualCluster cluster is empty")
	}

	if opt == nil || opt.Cloud == nil || opt.HostCluster == nil {
		return nil, fmt.Errorf("qcloud CreateVirtualCluster opt/cloud/hostCluster is empty")
	}

	if opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("qcloud CreateVirtualCluster lost credential info")
	}

	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when CreateVirtualCluster %d failed, %s",
			opt.Cloud.CloudID, cls.ClusterName, err.Error(),
		)
		return nil, err
	}

	// build create virtual cluster task
	task, err := mgr.BuildCreateVirtualClusterTask(cls, opt)
	if err != nil {
		blog.Errorf("build CreateVirtualCluster task for cluster %s with cloudprovider %s failed, %s",
			cls.ClusterName, cls.Provider, err.Error(),
		)
		return nil, err
	}

	return task, nil
}

// DeleteVirtualCluster delete virtual cluster
func (c *Cluster) DeleteVirtualCluster(cls *proto.Cluster,
	opt *cloudprovider.DeleteVirtualClusterOption) (*proto.Task, error) {
	if cls == nil {
		return nil, fmt.Errorf("qcloud DeleteVirtualCluster cluster is empty")
	}

	if opt == nil || opt.Cloud == nil || opt.HostCluster == nil || (opt.Namespace == nil || opt.Namespace.Name == "") {
		return nil, fmt.Errorf("qcloud DeleteVirtualCluster opt/cloud/hostCluster is empty")
	}

	if opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("qcloud DeleteVirtualCluster lost credential info")
	}

	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when DeleteVirtualCluster %d failed, %s",
			opt.Cloud.CloudID, cls.ClusterName, err.Error(),
		)
		return nil, err
	}

	// build delete virtual cluster task
	task, err := mgr.BuildDeleteVirtualClusterTask(cls, opt)
	if err != nil {
		blog.Errorf("build DeleteVirtualCluster task for cluster %s with cloudprovider %s failed, %s",
			cls.ClusterName, cls.Provider, err.Error(),
		)
		return nil, err
	}

	return task, nil
}

// ImportCluster import cluster according cloudprovider
func (c *Cluster) ImportCluster(cls *proto.Cluster, opt *cloudprovider.ImportClusterOption) (*proto.Task, error) {
	// call qcloud interface to create cluster
	if cls == nil {
		return nil, fmt.Errorf("qcloud ImportCluster cluster is empty")
	}

	if opt == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("qcloud ImportCluster cluster opt or cloud is empty")
	}

	if opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("qcloud CreateCluster opt lost valid crendential info")
	}

	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when ImportCluster %d failed, %s",
			opt.Cloud.CloudID, cls.ClusterName, err.Error(),
		)
		return nil, err
	}

	_, nodeIPs, err := getClusterInstancesByClusterID(cls.SystemID, &opt.CommonOption)
	if err != nil {
		blog.Errorf("get cloud/cluster %s/%s nodes when ImportCluster %d failed, %s",
			opt.Cloud.CloudID, cls.SystemID, cls.ClusterName, err.Error(),
		)
		return nil, err
	}

	opt.NodeIPs = nodeIPs

	// build import cluster task
	task, err := mgr.BuildImportClusterTask(cls, opt)
	if err != nil {
		blog.Errorf("build ImportCluster task for cluster %s with cloudprovider %s failed, %s",
			cls.ClusterName, cls.Provider, err.Error(),
		)
		return nil, err
	}

	return task, nil
}

// get tke cluster masterIPs && nodeIPs
func getClusterInstancesByClusterID(clusterID string, option *cloudprovider.CommonOption) ([]string, []string, error) {
	tkeCli, err := api.NewTkeClient(option)
	if err != nil {
		return nil, nil, err
	}

	instancesList, err := tkeCli.QueryTkeClusterAllInstances(context.Background(), clusterID, nil)
	if err != nil {
		return nil, nil, err
	}

	var (
		masterIPs, nodeIPs = make([]string, 0), make([]string, 0)
	)
	for _, ins := range instancesList {
		switch ins.InstanceRole {
		case api.MASTER_ETCD.String():
			masterIPs = append(masterIPs, ins.InstanceIP)
		case api.WORKER.String():
			nodeIPs = append(nodeIPs, ins.InstanceIP)
		default:
			continue
		}
	}

	return masterIPs, nodeIPs, nil
}

// DeleteCluster delete kubenretes cluster according cloudprovider
func (c *Cluster) DeleteCluster(cls *proto.Cluster, opt *cloudprovider.DeleteClusterOption) (*proto.Task, error) {
	if cls == nil {
		return nil, fmt.Errorf("qcloud DeleteCluster cluster is empty")
	}

	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 ||
		len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("qcloud DeleteCluster cluster lost oprion")
	}

	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when DeleteCluster %d failed, %s",
			opt.Cloud.CloudID, cls.ClusterName, err.Error(),
		)
		return nil, err
	}

	// build delete cluster task
	task, err := mgr.BuildDeleteClusterTask(cls, opt)
	if err != nil {
		blog.Errorf("build DeleteCluster task for cluster %s with cloudprovider %s failed, %s",
			cls.ClusterName, cls.Provider, err.Error(),
		)
		return nil, err
	}

	return task, nil
}

// GetCluster get kubernetes cluster detail information according cloudprovider
func (c *Cluster) GetCluster(cloudID string, opt *cloudprovider.GetClusterOption) (*proto.Cluster, error) {
	if cloudID == "" || opt == nil || opt.Cluster == nil {
		return nil, fmt.Errorf("%s GetCluster valid info empty", cloudName)
	}
	if opt.Account == nil || len(opt.Account.SecretID) == 0 ||
		len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("%s GetCluster lost credential info", cloudName)
	}

	return updateClusterInfo(cloudID, opt)
}

// getCloudCluster get tke cloud cluster
func getCloudCluster(cloudID string, opt *cloudprovider.CommonOption) (*tke.Cluster, error) {
	cli, err := api.NewTkeClient(opt)
	if err != nil {
		blog.Errorf("%s getCloudCluster NewTkeClient failed: %v", cloudName, err)
		return nil, err
	}
	cls, err := cli.GetTKECluster(cloudID)
	if err != nil {
		blog.Errorf("%s getCloudCluster GetTKECluster failed: %v", cloudName, err)
		return nil, err
	}

	return cls, err
}

// checkIfWhiteImageOsNames check cluster osName if it is white image osName
func checkIfWhiteImageOsNames(opt *cloudprovider.ClusterGroupOption) bool {
	if opt == nil || opt.Cluster == nil {
		blog.Errorf("checkIfWhiteImageOsNames failed: %v", "option empty")
		return false
	}

	if opt.Cluster.SystemID == "" {
		blog.Errorf("checkIfWhiteImageOsNames[%s] failed: systemID empty", opt.Cluster.ClusterID)
		return false
	}

	cls, err := getCloudCluster(opt.Cluster.SystemID, &opt.CommonOption)
	if err != nil {
		blog.Errorf("%s checkIfWhiteImageOsNames[%s] getCloudCluster failed: %v",
			cloudName, opt.Cluster.ClusterID, err)
		return false
	}

	// NOCC:ineffassign/assign(误报)
	osName := ""
	if opt.Group != nil && opt.Group.NodeTemplate != nil && opt.Group.NodeTemplate.NodeOS != "" {
		osName = opt.Group.NodeTemplate.NodeOS
		blog.Infof("checkIfWhiteImageOsNames[%s] osName[%s]", opt.Cluster.ClusterID, osName)
		return utils.StringInSlice(osName, utils.WhiteImageOsName)
	}

	if cls.ImageId != nil && *cls.ImageId != "" {
		nodeMgr := &NodeManager{}
		image, errGet := nodeMgr.GetImageInfoByImageID(*cls.ImageId, &opt.CommonOption)
		if errGet != nil {
			blog.Errorf("%s checkIfWhiteImageOsNames GetImageInfoByImageID failed: %v", cloudName, errGet)
			osName = *cls.ClusterOs
		} else {
			osName = image.OsName
		}
	} else {
		osName = *cls.ClusterOs
	}

	blog.Infof("checkIfWhiteImageOsNames[%s] osName[%s]", opt.Cluster.ClusterID, osName)
	return utils.StringInSlice(osName, utils.WhiteImageOsName)
}

// clusterSupportNodeNum cluster support node num
func clusterSupportNodeNum(tkeCls *tke.Cluster, cluster *proto.Cluster) (uint32, uint32, uint32) {
	var (
		ipNum          uint32
		clusterCidrNum uint32
	)
	cidrs, err := business.GetCidrsFromCluster(tkeCls)
	if err != nil {
		blog.Errorf("clusterSupportNodeNum failed: %v", err)
		return 0, 0, 0
	}
	for i := range cidrs {
		if utils.StringInSlice(cidrs[i].Type, []string{utils.ClusterCIDR, utils.MultiClusterCIDR}) {
			clusterCidrNum++
			num, _ := cidrs[i].GetIPNum()
			ipNum += num
		}
	}

	// 已经存在的节点数量
	clusterNodeNum := *tkeCls.ClusterNodeNum
	if *tkeCls.ClusterType == icommon.ClusterManageTypeIndependent {
		clusterNodeNum += *tkeCls.ClusterMaterNodeNum
	}

	// 集群可添加节点数
	maxClusterNodeNum := float64(uint64(ipNum)-*tkeCls.ClusterNetworkSettings.MaxClusterServiceNum) /
		float64(*tkeCls.ClusterNetworkSettings.MaxNodePodNum)

	// 剩余可支持的节点数量
	step := getClusterCidrStep(cluster)

	surplusNodeNum := float64((business.GrBcsMaxClusterCidrNum-clusterCidrNum)*step) /
		float64(*tkeCls.ClusterNetworkSettings.MaxNodePodNum)

	return uint32(clusterNodeNum), uint32(maxClusterNodeNum) - uint32(clusterNodeNum), uint32(surplusNodeNum)
}

// updateClusterInfo update cluster info
func updateClusterInfo(cloudID string, opt *cloudprovider.GetClusterOption) (*proto.Cluster, error) {
	cls, err := getCloudCluster(cloudID, &opt.CommonOption)
	if err != nil {
		blog.Errorf("%s updateClusterInfo getCloudCluster failed: %v", cloudName, err)
		return nil, err
	}

	opt.Cluster.ManageType = *cls.ClusterType

	if opt.Cluster.ClusterAdvanceSettings != nil {
		opt.Cluster.ClusterAdvanceSettings.ContainerRuntime = *cls.ContainerRuntime
		opt.Cluster.ClusterAdvanceSettings.RuntimeVersion = *cls.RuntimeVersion
		opt.Cluster.ClusterAdvanceSettings.NetworkType = getTkeClusterNetworkType(cls)
	}
	if opt.Cluster.ClusterBasicSettings != nil {
		opt.Cluster.ClusterBasicSettings.Version = *cls.ClusterVersion
		opt.Cluster.ClusterBasicSettings.VersionName = *cls.ClusterVersion
		opt.Cluster.ClusterBasicSettings.ClusterLevel = *cls.ClusterLevel
		opt.Cluster.ClusterBasicSettings.IsAutoUpgradeClusterLevel = *cls.AutoUpgradeClusterLevel
	}
	if opt.Cluster.ExtraInfo == nil {
		opt.Cluster.ExtraInfo = make(map[string]string)
	}

	if cls.ImageId != nil && *cls.ImageId != "" {
		nodeMgr := &NodeManager{}
		image, errGet := nodeMgr.GetImageInfoByImageID(*cls.ImageId, &opt.CommonOption)
		if errGet != nil {
			blog.Errorf("%s updateClusterInfo GetImageInfoByImageID failed: %v", cloudID, errGet)
			opt.Cluster.ClusterBasicSettings.OS = *cls.ClusterOs
		} else {
			opt.Cluster.ClusterBasicSettings.OS = image.OsName
		}
		opt.Cluster.ExtraInfo[icommon.ImageProvider] = icommon.PrivateImageProvider
	} else {
		opt.Cluster.ClusterBasicSettings.OS = *cls.ClusterOs
		opt.Cluster.ExtraInfo[icommon.ImageProvider] = icommon.PublicImageProvider
	}

	// 计算集群可支持节点容量
	currentNodeNum, supNodeNum, maxNodeNum := clusterSupportNodeNum(cls, opt.Cluster)
	opt.Cluster.ExtraInfo[icommon.ClusterCurNodeNum] = fmt.Sprintf("%v", currentNodeNum)
	opt.Cluster.ExtraInfo[icommon.ClusterSupNodeNum] = fmt.Sprintf("%v", supNodeNum)
	opt.Cluster.ExtraInfo[icommon.ClusterMaxNodeNum] = fmt.Sprintf("%v", maxNodeNum)

	if opt.Cluster.NetworkSettings == nil {
		opt.Cluster.NetworkSettings = &proto.NetworkSetting{}
	}
	if opt.Cluster.NetworkSettings.SubnetSource == nil {
		opt.Cluster.NetworkSettings.SubnetSource = &proto.SubnetSource{}
	}

	// 集群VPC-CNI模式子网信息
	if !utils.StringInSlice(opt.Cluster.GetNetworkSettings().GetStatus(),
		[]string{icommon.StatusInitialization, icommon.TaskStatusFailure}) {
		opt.Cluster.NetworkSettings.EnableVPCCni = business.GetClusterVpcCniStatus(cls)
	}
	if opt.Cluster.NetworkSettings.GetNetworkMode() == "" {
		opt.Cluster.NetworkSettings.NetworkMode = api.TKERouteEni
	}
	opt.Cluster.NetworkSettings.EniSubnetIDs = business.GetClusterVpcCniSubnets(cls)

	return opt.Cluster, nil
}

// ListCluster get cloud cluster list by region
func (c *Cluster) ListCluster(opt *cloudprovider.ListClusterOption) ([]*proto.CloudClusterInfo, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 ||
		len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("qcloud ListCluster cluster lost operation")
	}

	cli, err := api.NewTkeClient(&opt.CommonOption)
	if err != nil {
		return nil, err
	}
	tkeClusters, err := cli.ListTKECluster()
	if err != nil {
		return nil, err
	}

	return transTKEClusterToCloudCluster(opt.Region, tkeClusters), nil
}

// transTKEClusterToCloudCluster trans cluster
func transTKEClusterToCloudCluster(region string, clusters []*tke.Cluster) []*proto.CloudClusterInfo {
	cloudClusterList := make([]*proto.CloudClusterInfo, 0)
	for _, cls := range clusters {
		cloudClusterList = append(cloudClusterList, &proto.CloudClusterInfo{
			ClusterID:          *cls.ClusterId,
			ClusterName:        *cls.ClusterName,
			ClusterDescription: *cls.ClusterDescription,
			ClusterVersion:     *cls.ClusterVersion,
			ClusterOS:          *cls.ClusterOs,
			ClusterType:        *cls.ClusterType,
			ClusterStatus:      *cls.ClusterStatus,
			Location:           region,
		})
	}

	return cloudClusterList
}

// GetNodesInCluster get all nodes belong to cluster according cloudprovider
func (c *Cluster) GetNodesInCluster(cls *proto.Cluster, opt *cloudprovider.GetNodesOption) ([]*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// AddNodesToCluster add new node to cluster according cloudprovider
func (c *Cluster) AddNodesToCluster(cls *proto.Cluster, nodes []*proto.Node,
	opt *cloudprovider.AddNodesOption) (*proto.Task, error) {
	if cls == nil {
		return nil, fmt.Errorf("qcloud AddNodesToCluster cluster is empty")
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("qcloud AddNodesToCluster nodes is empty")
	}

	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 ||
		len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("qcloud AddNodesToCluster cluster lost operation")
	}

	if opt.Operator == "" || opt.Cloud == nil {
		return nil, fmt.Errorf("qcloud AddNodesToCluster cluster lost operator|cloud")
	}

	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when AddNodesToCluster %d failed, %s",
			opt.Cloud.CloudID, cls.ClusterName, err.Error(),
		)
		return nil, err
	}

	// build add nodes to cluster task
	task, err := mgr.BuildAddNodesToClusterTask(cls, nodes, opt)
	if err != nil {
		blog.Errorf("build AddNodesToCluster task for cluster %s with cloudprovider %s failed, %s",
			cls.ClusterName, cls.Provider, err.Error(),
		)
		return nil, err
	}

	return task, nil
}

// DeleteNodesFromCluster delete specified nodes from cluster according cloudprovider
func (c *Cluster) DeleteNodesFromCluster(cls *proto.Cluster, nodes []*proto.Node,
	opt *cloudprovider.DeleteNodesOption) (*proto.Task, error) {
	if cls == nil {
		return nil, fmt.Errorf("qcloud DeleteNodesFromCluster cluster is empty")
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("qcloud DeleteNodesFromCluster nodes is empty")
	}

	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 ||
		len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("qcloud DeleteNodesFromCluster cluster lost operation")
	}

	if opt.Operator == "" || opt.Cloud == nil {
		return nil, fmt.Errorf("qcloud DeleteNodesFromCluster cluster lost operator|cloud")
	}

	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when DeleteNodesFromCluster %d failed, %s",
			opt.Cloud.CloudID, cls.ClusterName, err.Error(),
		)
		return nil, err
	}

	// build delete nodes from cluster task
	task, err := mgr.BuildRemoveNodesFromClusterTask(cls, nodes, opt)
	if err != nil {
		blog.Errorf("build DeleteNodesFromCluster task for cluster %s with cloudprovider %s failed, %s",
			cls.ClusterName, cls.Provider, err.Error(),
		)
		return nil, err
	}

	return task, nil
}

// skipGlobalRouterCIDR skip global router cidr
func skipGlobalRouterCIDR(cls *proto.Cluster) bool {
	if cls.ExtraInfo != nil {
		v, ok := cls.ExtraInfo[api.GlobalRouteCIDRCheck]
		if ok && v == "true" {
			return true
		}
	}

	return false
}

// CheckClusterCidrAvailable check cluster CIDR nodesNum when add nodes
func (c *Cluster) CheckClusterCidrAvailable(cls *proto.Cluster,
	opt *cloudprovider.CheckClusterCIDROption) (bool, error) {
	if cls == nil || opt == nil || opt.ExternalNode {
		return true, nil
	}

	// skip clusterCidr autoScale about some scene
	if skipGlobalRouterCIDR(cls) {
		blog.Infof("CheckClusterCidrAvailable skipGlobalRouterCIDR successful")
		return true, nil
	}

	_, ipNum, err := getClusterCidrAvailableIPNum(cls.ClusterID, cls.SystemID, &opt.CommonOption)
	if err != nil {
		return false, err
	}
	sumIPNum := uint32(opt.IncomingNodeCnt) * cls.NetworkSettings.MaxNodePodNum
	blog.Infof("cluster[%s] cloud[%s] CheckClusterCidrAvailable for incomingNodes[%v] availableIPCount[%v] "+
		"needIPCount[%v] addNodeCnt[%v]", cls.ClusterID, cloudName, opt.IncomingNodeCnt, ipNum, sumIPNum, opt.IncomingNodeCnt)

	if ipNum >= sumIPNum {
		return true, nil
	}

	cidrList, err := autoScaleClusterCidr(opt.CommonOption, cls, sumIPNum-ipNum)
	if err != nil {
		return false, err
	}

	cls.NetworkSettings.MultiClusterCIDR = append(cls.NetworkSettings.MultiClusterCIDR, cidrList...)
	err = cloudprovider.GetStorageModel().UpdateCluster(context.Background(), cls)
	if err != nil {
		blog.Errorf("CheckClusterCidrAvailable cluster[%s] update multiClusterCDR failed: %v", cls.ClusterID, err)
	}

	return true, nil
}

// EnableExternalNodeSupport enable cluster support external node
func (c *Cluster) EnableExternalNodeSupport(cls *proto.Cluster, opt *cloudprovider.EnableExternalNodeOption) error {
	if cls == nil {
		return fmt.Errorf("qcloud EnableExternalNodeSupport cluster is empty")
	}
	validate := func(opt *cloudprovider.EnableExternalNodeOption) error {
		if opt == nil || opt.Operator == "" || opt.EnablePara == nil {
			return fmt.Errorf("qcloud EnableExternalNodeSupport lost valid paras")
		}
		if opt.EnablePara.NetworkType == "" || opt.EnablePara.SubnetId == "" || opt.EnablePara.ClusterCIDR == "" {
			return fmt.Errorf("qcloud EnableExternalNodeSupport enableexternal paras empty")
		}

		return nil
	}
	err := validate(opt)
	if err != nil {
		return err
	}
	cli, err := api.NewTkeClient(&opt.CommonOption)
	if err != nil {
		return err
	}
	err = retry.Do(func() error {
		err := cli.EnableExternalNodeSupport(cls.SystemID, api.EnableExternalNodeConfig{ // nolint
			NetworkType: opt.EnablePara.NetworkType,
			ClusterCIDR: opt.EnablePara.ClusterCIDR,
			SubnetId:    opt.EnablePara.SubnetId,
			Enabled:     opt.EnablePara.Enabled,
		})
		if err != nil {
			blog.Errorf("qcloud EnableExternalNodeSupport[%s] failed: %v", cls.ClusterID, err)
			return err
		}

		return nil
	}, retry.Attempts(3))
	if err != nil {
		blog.Errorf("qcloud EnableExternalNodeSupport[%s] failed: %v", cls.ClusterID, err)
		return err
	}

	return nil
}

// ListOsImage list image os
func (c *Cluster) ListOsImage(provider string, opt *cloudprovider.CommonOption) ([]*proto.OsImage, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 ||
		len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("qcloud ListOsImage lost authoration")
	}

	images := make([]*proto.OsImage, 0)

	cli, err := api.NewTkeClient(opt)
	if err != nil {
		return nil, err
	}
	cloudImages, err := cli.DescribeOsImages(provider, opt)
	if err != nil {
		return nil, err
	}

	for _, image := range cloudImages {
		if image == nil || *image.Status == "offline" {
			continue
		}

		images = append(images, &proto.OsImage{
			ImageID: *image.ImageId,
			Alias:   *image.Alias,
			Arch:    *image.Arch,
			OsCustomizeType: func() string {
				if image.OsCustomizeType == nil {
					return ""
				}
				return *image.OsCustomizeType
			}(),
			OsName: *image.OsName,
			SeriesName: func() string {
				if image.SeriesName == nil {
					return ""
				}
				return *image.SeriesName
			}(),
			Status:   *image.Status,
			Provider: provider,
		})
	}

	return images, nil
}

// ListProjects list cloud projects
func (c *Cluster) ListProjects(opt *cloudprovider.CommonOption) ([]*proto.CloudProject, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 ||
		len(opt.Account.SecretKey) == 0 {
		return nil, fmt.Errorf("qcloud ListProjects lost authoration")
	}

	projects := make([]*proto.CloudProject, 0)

	cli, err := api.NewTagClient(opt)
	if err != nil {
		return nil, err
	}
	cloudProjects, err := cli.ListProjects()
	if err != nil {
		return nil, err
	}

	for _, pro := range cloudProjects {
		projects = append(projects, &proto.CloudProject{
			ProjectID:   *pro.ProjectId,
			ProjectName: *pro.ProjectName,
		})
	}

	return projects, nil
}

// AppendCloudNodeInfo append cloud node detailed info
func (c *Cluster) AppendCloudNodeInfo(ctx context.Context,
	nodes []*proto.ClusterNode, opt *cloudprovider.CommonOption) error {

	zoneIdMap, zoneMap, err := business.GetZoneInfoByRegion(opt)
	if err != nil {
		blog.Errorf("AppendCloudNodeInfo GetZoneInfoByRegion failed: %v", err)
		return err
	}
	// 获取语言
	lang := i18n.LanguageFromCtx(ctx)
	// get node zoneName
	for i := range nodes {
		zone, ok := zoneIdMap[nodes[i].ZoneName]
		if ok {
			nodes[i].ZoneName = zone.GetZoneName()
			if lang != utils.ZH {
				nodes[i].ZoneName = zone.GetZone()
			}
			continue
		}
		zone, ok = zoneMap[nodes[i].ZoneID]
		if ok {
			nodes[i].ZoneName = zone.GetZoneName()
			if lang != utils.ZH {
				nodes[i].ZoneName = zone.GetZone()
			}
			continue
		}
	}

	return nil
}

// mergeSubnetSource merge subnets
func mergeSubnetSource(originSubs, newSubs []*proto.NewSubnet) []*proto.NewSubnet {
	if originSubs == nil {
		originSubs = make([]*proto.NewSubnet, 0)
	}

	originSubsMap := make(map[string]*proto.NewSubnet, 0)
	for i := range originSubs {
		originSubsMap[originSubs[i].GetZone()] = originSubs[i]
	}

	for i := range newSubs {
		zone := newSubs[i].GetZone()

		sub, ok := originSubsMap[zone]
		if ok {
			sub.IpCnt += newSubs[i].GetIpCnt()
		} else {
			originSubs = append(originSubs, &proto.NewSubnet{
				Zone:  zone,
				IpCnt: newSubs[i].GetIpCnt(),
			})
		}
	}

	return originSubs
}

// AddSubnetsToCluster add subnets to cluster
func (c *Cluster) AddSubnetsToCluster(ctx context.Context, subnet *proto.SubnetSource,
	opt *cloudprovider.AddSubnetsToClusterOption) error {
	if opt == nil || opt.Cluster == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 ||
		len(opt.Account.SecretKey) == 0 {
		return fmt.Errorf("AddSubnetsToCluster lost cloud accoount")
	}
	if subnet == nil || len(subnet.GetNew()) == 0 {
		return fmt.Errorf("AddSubnetsToCluster subnet data empty")
	}

	// 检测当前集群子网资源使用率, 如果使用率达标则继续扩容, 不达标则拒绝扩容
	zoneSubnetRatio, _, _, err := business.GetClusterCurrentVpcCniSubnets(*opt.Cluster, false)
	if err != nil {
		return fmt.Errorf("AddSubnetsToCluster failed: %v", err)
	}

	goalRatio := opt.Cloud.GetNetworkInfo().GetUnderlayRatio()
	for i := range subnet.GetNew() {
		zoneRatio, ok := zoneSubnetRatio[subnet.GetNew()[i].GetZone()]
		if ok && zoneRatio.Ratio < float64(goalRatio) {
			return fmt.Errorf("zone[%s] usage lt goalRatio %+v", subnet.GetNew()[i].GetZone(), goalRatio)
		}
	}

	newClusterSubnets := mergeSubnetSource(opt.Cluster.GetNetworkSettings().GetSubnetSource().GetNew(), subnet.GetNew())
	if opt.Cluster.NetworkSettings.SubnetSource == nil {
		opt.Cluster.NetworkSettings.SubnetSource = &proto.SubnetSource{}
	}

	opt.Cluster.NetworkSettings.SubnetSource.New = newClusterSubnets
	return cloudprovider.UpdateCluster(opt.Cluster)
}

// GetMasterSuggestedMachines get master suggested machines
func (c *Cluster) GetMasterSuggestedMachines(level, vpcId string,
	opt *cloudprovider.GetMasterSuggestedMachinesOption) ([]*proto.InstanceTemplateConfig, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// CheckClusterEndpointStatus check cluster endpoint status
func (c *Cluster) CheckClusterEndpointStatus(clusterID string, isExtranet bool,
	opt *cloudprovider.CheckEndpointStatusOption) (bool, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 ||
		len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return false, fmt.Errorf("qcloud CheckClusterEndpointStatus lost authoration")
	}

	client, err := api.NewTkeClient(&opt.CommonOption)
	if err != nil {
		return false, err
	}

	// GetClusterEndpointStatus endpoint status
	status, err := client.GetClusterEndpointStatus(clusterID, isExtranet)
	if err != nil {
		return false, err
	}
	blog.Infof("cluster endpoint status: %s", status)

	if !status.Created() {
		return false, fmt.Errorf("cluster endpoint status is not created")
	}

	// get cluster kubeconfig
	kubeConfig, err := client.GetTKEClusterKubeConfig(clusterID, isExtranet)
	if err != nil {
		return false, err
	}

	data, err := base64.StdEncoding.DecodeString(kubeConfig)
	if err != nil {
		return false, fmt.Errorf("decode kube config failed: %v", err)
	}

	_, err = cloudprovider.GetCRDByKubeConfig(string(data))
	if err != nil {
		return false, err
	}

	return true, nil
}

// CheckIfGetNodesFromCluster check cluster if can get nodes from k8s
func (c *Cluster) CheckIfGetNodesFromCluster(ctx context.Context, cluster *proto.Cluster,
	nodes []*proto.ClusterNode) bool {
	// managed cluster
	if cluster.ManageType == icommon.ClusterManageTypeManaged && !utils.ExistRunningNodes(nodes) {
		blog.Infof("CheckIfGetNodesFromCluster[%s] successful", cluster.ClusterID)
		return false
	}

	return true
}

// SwitchClusterNetwork switch cluster network mode
func (c *Cluster) SwitchClusterNetwork(
	cls *proto.Cluster, subnet *proto.SubnetSource, opt *cloudprovider.SwitchClusterNetworkOption) (*proto.Task, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 ||
		len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("qcloud SwitchClusterNetwork lost authoration")
	}

	// GetTaskManager for cluster manager initialization
	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when SwitchClusterNetwork %d failed, %s",
			opt.Cloud.CloudID, cls.ClusterName, err.Error(),
		)
		return nil, err
	}

	// build create cluster task
	task, err := mgr.BuildSwitchClusterNetworkTask(cls, subnet, opt)
	if err != nil {
		blog.Errorf("build SwitchClusterNetwork task for cluster %s with cloudProvider %s failed, %s",
			cls.ClusterName, cls.Provider, err.Error(),
		)
		return nil, err
	}
	return task, nil
}

// CheckClusterNetworkStatus check cluster network status
func (c *Cluster) CheckClusterNetworkStatus(clusterId string,
	opt *cloudprovider.CheckClusterNetworkStatusOption) (bool, error) {
	if clusterId == "" {
		return false, fmt.Errorf("cluster[%s] cloud systemId empty", opt.Cluster.ClusterID)
	}

	// get cloud cluster
	cls, err := getCloudCluster(clusterId, &opt.CommonOption)
	if err != nil {
		blog.Errorf("Get Cluster %s failed, %s", clusterId, err.Error())
		return false, err
	}

	if opt.Cluster.GetNetworkSettings().GetSubnetSource() == nil {
		opt.Cluster.GetNetworkSettings().SubnetSource = &proto.SubnetSource{}
	}

	switch opt.Disable {
	case true:
		// 底层集群已经关闭
		if !business.GetClusterVpcCniStatus(cls) {
			opt.Cluster.NetworkSettings.EnableVPCCni = false
			opt.Cluster.NetworkSettings.EniSubnetIDs = nil
			opt.Cluster.NetworkSettings.SubnetSource.New = nil
			opt.Cluster.NetworkSettings.Status = icommon.StatusRunning

			return false, nil
		}

		if !opt.Cluster.GetNetworkSettings().GetEnableVPCCni() &&
			opt.Cluster.GetNetworkSettings().GetStatus() != icommon.TaskStatusFailure {
			return false,
				fmt.Errorf("cluster %s/%s already close vpc-cni", opt.Cluster.ClusterID, opt.Cluster.ClusterName)
		}

		// check subnets usage when close vpc-cni
		opt.Cluster.NetworkSettings.EniSubnetIDs = nil
		opt.Cluster.NetworkSettings.SubnetSource.New = nil
	default:
		if business.GetClusterVpcCniStatus(cls) {
			opt.Cluster.NetworkSettings.EnableVPCCni = true
			opt.Cluster.NetworkSettings.EniSubnetIDs = business.GetClusterVpcCniSubnets(cls)
			opt.Cluster.NetworkSettings.Status = icommon.StatusRunning

			zoneSubs, _, errLocal := business.GetClusterSubnetsZoneUsage(&opt.CommonOption,
				business.GetClusterVpcCniSubnets(cls), true)
			if errLocal != nil {
				return false, errLocal
			}

			opt.Cluster.NetworkSettings.SubnetSource.New = func() []*proto.NewSubnet {
				newSubnets := make([]*proto.NewSubnet, 0)
				for zone, sub := range zoneSubs {
					newSubnets = append(newSubnets, &proto.NewSubnet{
						Zone:  zone,
						IpCnt: uint32(sub.TotalIps),
					})
				}

				return newSubnets
			}()

			return false, nil
		}

		if opt.Cluster.GetNetworkSettings().GetEnableVPCCni() &&
			opt.Cluster.GetNetworkSettings().GetStatus() != icommon.TaskStatusFailure {
			return false,
				fmt.Errorf("cluster %s/%s already open vpc-cni", opt.Cluster.ClusterID, opt.Cluster.ClusterName)
		}
		opt.Cluster.NetworkSettings.IsStaticIpMode = opt.IsStaticIpMode
		opt.Cluster.NetworkSettings.ClaimExpiredSeconds = opt.ClaimExpiredSeconds
		opt.Cluster.NetworkSettings.SubnetSource = opt.SubnetSource
		if opt.Cluster.NetworkSettings.GetClaimExpiredSeconds() <= 0 {
			opt.Cluster.NetworkSettings.ClaimExpiredSeconds = 300
		}
	}

	return true, nil
}

// UpdateCloudKubeConfig update cluster kubeconfig to clustercredential
func (c *Cluster) UpdateCloudKubeConfig(kubeConfig string,
	opt *cloudprovider.UpdateCloudKubeConfigOption) error {
	if kubeConfig == "" {
		// 开启admin权限, 并生成kubeconfig
		clusterKube, connectKube, err := tasks.OpenClusterAdminKubeConfig(
			context.Background(), &cloudprovider.CloudDependBasicInfo{
				Cluster:  opt.Cluster,
				CmOption: &opt.CommonOption,
			})
		if err != nil {
			return err
		}
		blog.Infof("UpdateCloudKubeConfig[%s] openClusterAdminKubeConfig[%s] [%s] success",
			opt.Cluster.ClusterID, clusterKube, connectKube)

		kubeBytes, err := base64.StdEncoding.DecodeString(clusterKube)
		if err != nil {
			return err
		}
		kubeConfig = string(kubeBytes)
	}

	// update cluster credential
	config, err := types.GetKubeConfigFromYAMLBody(false, types.YamlInput{
		YamlContent: kubeConfig,
	})
	if err != nil {
		return err
	}

	err = cloudprovider.UpdateClusterCredentialByConfig(opt.Cluster.ClusterID, config)
	if err != nil {
		return err
	}

	return nil
}

// getClusterCidrAvailableIPNum get global router ip num
func getClusterCidrAvailableIPNum(clusterId, tkeId string, option *cloudprovider.CommonOption) (uint32, uint32, error) {
	return business.GetClusterGrIPSurplus(option, clusterId, tkeId)
}

// TKE cluster exist master clusterCIDR and multiCIDRList, multiCIDRList add length 9 CIDRs at most.
// when scale tke cluster CIDRs at present,
// BCS use [step, ..., step, xxx, xxx] 7 step rules, xxx need to manually assign
func autoScaleClusterCidr(option cloudprovider.CommonOption, cls *proto.Cluster, needIPNum uint32) ([]string, error) {
	// not allow when assign full multiCIDR
	if len(cls.NetworkSettings.MultiClusterCIDR) >= utils.MultiClusterCIDRCnt {
		return nil, fmt.Errorf("cluster[%s] scaleNodes exceed max cdir number[%v]", cls.ClusterID,
			utils.MultiClusterCIDRCnt)
	}

	// auto scale cidr resource when addNodes previous clusters may be not set cidrStep
	defaultCidrStep := getClusterCidrStep(cls)

	// surPlusIPNum if enough
	surPlusIPNum := getSurplusCidrNum(cls.NetworkSettings.MultiClusterCIDR, defaultCidrStep)
	blog.Infof("cluster[%s] cloud[%s] CheckClusterCidrAvailable surPlusIPCount[%v] needIPCount[%v]",
		cls.ClusterID, cloudName, surPlusIPNum, needIPNum)

	if surPlusIPNum < needIPNum {
		return nil, fmt.Errorf("cluster[%s] scaleNodes exceed max cdir number", cls.ClusterID)
	}

	// calculate mask
	var (
		maskIPNum = make([]uint32, 0)
		sumIPSum  uint32
	)
	for _, segNum := range getSurplusCidrList(cls.NetworkSettings.MultiClusterCIDR, defaultCidrStep) {
		sumIPSum += segNum
		maskIPNum = append(maskIPNum, utils.CalMaskLen(float64(segNum)))

		if sumIPSum >= needIPNum {
			break
		}
	}
	blog.Infof("cluster[%s] cloud[%s] CheckClusterCidrAvailable maskIPNum[%v]",
		cls.ClusterID, cloudName, maskIPNum)

	return business.AddGrCidrsToCluster(&option, cls.GetVpcID(), cls, maskIPNum, nil)
}

// getClusterCidrStep get cluster cidr step
func getClusterCidrStep(cls *proto.Cluster) uint32 {
	defaultCidrStep := cls.NetworkSettings.CidrStep

	if defaultCidrStep <= 0 {
		defaultCidrStep = func() uint32 {
			if cls.Environment == icommon.Prod {
				return 4096
			}

			return 2048
		}()
	}

	return defaultCidrStep
}

// getTkeClusterNetworkType get tke cluster networkType
func getTkeClusterNetworkType(cluster *tke.Cluster) string {
	property := *cluster.Property

	propertyInfo := make(map[string]interface{})
	err := json.Unmarshal([]byte(property), &propertyInfo)
	if err != nil {
		return ""
	}
	nType, ok := propertyInfo["NetworkType"]
	if ok {
		v, ok1 := nType.(string)
		if ok1 {
			return v
		}
	}

	return ""
}

// getSurplusCidrList xxx
func getSurplusCidrList(mulList []string, step uint32) []uint32 {
	cidrList := make([]uint32, 0)

	for i := len(mulList); i < utils.MultiClusterCIDRCnt; i++ {
		cidrList = append(cidrList, step)
	}

	return cidrList
}

// getSurplusCidrNum xxx
func getSurplusCidrNum(mulList []string, step uint32) uint32 {
	surplusCidrCnt := utils.MultiClusterCIDRCnt - len(mulList)

	return step * uint32(surplusCidrCnt)
}
