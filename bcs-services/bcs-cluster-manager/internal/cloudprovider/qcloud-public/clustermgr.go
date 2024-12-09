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
	"sort"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n"
	"github.com/avast/retry-go"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud-public/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
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
		return nil, fmt.Errorf("%s CreateCluster cluster is empty", cloudName)
	}

	if opt == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("%s CreateCluster cluster opt or cloud is empty", cloudName)
	}

	if opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("%s CreateCluster opt lost valid crendential info", cloudName)
	}

	// GetTaskManager for nodegroup manager initialization
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
		return nil, fmt.Errorf("%s CreateVirtualCluster cluster is empty", cloudName)
	}

	if opt == nil || opt.Cloud == nil || opt.HostCluster == nil {
		return nil, fmt.Errorf("%s CreateVirtualCluster opt/cloud/hostCluster is empty", cloudName)
	}

	if opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("%s CreateVirtualCluster lost credential info", cloudName)
	}

	// GetTaskManager for nodegroup manager initialization
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
		return nil, fmt.Errorf("%s DeleteVirtualCluster cluster is empty", cloudName)
	}

	if opt == nil || opt.Cloud == nil || opt.HostCluster == nil || (opt.Namespace == nil || opt.Namespace.Name == "") {
		return nil, fmt.Errorf("%s DeleteVirtualCluster opt/cloud/hostCluster is empty", cloudName)
	}

	if opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("%s DeleteVirtualCluster lost credential info", cloudName)
	}

	// GetTaskManager for nodegroup manager initialization
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
		return nil, fmt.Errorf("%s ImportCluster cluster is empty", cloudName)
	}

	if opt == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("%s ImportCluster cluster opt or cloud is empty", cloudName)
	}

	if opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("%s CreateCluster opt lost valid crendential info", cloudName)
	}

	// GetTaskManager for nodegroup manager initialization
	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when ImportCluster %d failed, %s",
			opt.Cloud.CloudID, cls.ClusterName, err.Error(),
		)
		return nil, err
	}

	// get tke cluster masterIPs && nodeIPs
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

	// QueryTkeClusterAllInstances query all cluster instances
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
		return nil, fmt.Errorf("%s DeleteCluster cluster is empty", cloudName)
	}

	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 ||
		len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("%s DeleteCluster cluster lost oprion", cloudName)
	}

	// GetTaskManager for nodegroup manager initialization
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

func getCloudCluster(cloudID string, opt *cloudprovider.CommonOption) (*tke.Cluster, error) {
	cli, err := api.NewTkeClient(opt)
	if err != nil {
		blog.Errorf("%s getCloudCluster NewTkeClient failed: %v", cloudName, err)
		return nil, err
	}
	// GetTKECluster get tke cluster info
	cls, err := cli.GetTKECluster(cloudID)
	if err != nil {
		blog.Errorf("%s getCloudCluster GetTKECluster failed: %v", cloudName, err)
		return nil, err
	}

	return cls, err
}

// checkClusterOsNameInWhiteImages check cluster osName if it is white image osName
func checkClusterOsNameInWhiteImages(cls *proto.Cluster, opt *cloudprovider.CommonOption) bool {
	if cls == nil {
		blog.Errorf("checkClusterOsNameInWhiteImages failed: %v", "cluster nil")
		return false
	}

	osName := cls.GetClusterBasicSettings().GetOS()
	if osName == "" {
		blog.Errorf("checkClusterOsNameInWhiteImages[%s] failed: cluster OS empty", cls.ClusterID)
		return false
	}

	// osName maybe is imageID
	if osName != "" {
		nodeMgr := &NodeManager{}
		// GetImageInfoByImageID get image by image
		image, errGet := nodeMgr.GetImageInfoByImageID(osName, opt)
		if errGet != nil {
			blog.Errorf("%s checkClusterOsNameInWhiteImages GetImageInfoByImageID failed: %v", cloudName, errGet)
		} else {
			osName = image.OsName
		}
	}

	blog.Infof("checkClusterOsNameInWhiteImages[%s] osName[%s]", cls.ClusterID, osName)
	return utils.StringInSlice(osName, utils.WhiteImageOsName)
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

// update ClusterInfo
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

	//

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

// trans TKEClusterToCloudCluster
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

func skipGlobalRouterCIDR(cls *proto.Cluster) bool { // nolint
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

	if options.GetEditionInfo().IsCommunicationEdition() || options.GetEditionInfo().IsEnterpriseEdition() {
		return true, nil
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

// AddSubnetsToCluster cluster add subnet
func (c *Cluster) AddSubnetsToCluster(ctx context.Context, subnet *proto.SubnetSource,
	opt *cloudprovider.AddSubnetsToClusterOption) error {
	if opt == nil || opt.Cluster == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 ||
		len(opt.Account.SecretKey) == 0 {
		return fmt.Errorf("AddSubnetsToCluster lost cloud accoount")
	}
	if subnet == nil || (len(subnet.GetNew()) == 0 && len(subnet.Existed.GetIds()) == 0) {
		return fmt.Errorf("AddSubnetsToCluster subnet data empty")
	}

	subnetIds := make([]string, 0)
	if len(subnet.GetExisted().GetIds()) > 0 {
		subnetIds = append(subnetIds, subnet.GetExisted().GetIds()...)
	}

	allocateSubnets, err := business.AllocateClusterVpcCniSubnets(ctx,
		opt.Cluster.GetClusterID(), opt.Cluster.GetVpcID(), subnet.GetNew(), &opt.CommonOption)
	if err != nil {
		blog.Errorf("AddSubnetsToCluster AllocateClusterVpcCniSubnets failed: %v", err)
		return err
	}
	subnetIds = append(subnetIds, allocateSubnets...)

	// add subnets to cluster
	return addSubnetsToCluster(opt.Cluster, subnetIds, &opt.CommonOption)
}

func addSubnetsToCluster(cls *proto.Cluster, subnets []string, opt *cloudprovider.CommonOption) error {
	client, err := api.NewTkeClient(opt)
	if err != nil {
		return err
	}

	err = client.AddVpcCniSubnets(&api.AddVpcCniSubnetsInput{
		ClusterID: cls.GetSystemID(),
		VpcID:     cls.GetVpcID(),
		SubnetIDs: subnets,
	})
	if err != nil {
		return err
	}
	if cls.GetNetworkSettings().GetEniSubnetIDs() == nil {
		cls.GetNetworkSettings().EniSubnetIDs = make([]string, 0)
	}
	cls.GetNetworkSettings().EniSubnetIDs = append(cls.GetNetworkSettings().EniSubnetIDs, subnets...)

	return cloudprovider.GetStorageModel().UpdateCluster(context.Background(), cls)
}

// GetMasterSuggestedMachines get master suggested machines
func (c *Cluster) GetMasterSuggestedMachines(level, vpcId string,
	opt *cloudprovider.GetMasterSuggestedMachinesOption) ([]*proto.InstanceTemplateConfig, error) {
	var (
		clusterLevel     = ClusterLevel(level)
		instanceTemplate = make([]*proto.InstanceTemplateConfig, 0)
	)

	machineConfig := clusterLevel.GetCpuMemConfig(opt.Cpu, opt.Mem)

	mtZones, zoneInstanceTypes, err := getZoneMachineTypes(machineConfig.Cpu, machineConfig.Mem, opt.CommonOption)
	if err != nil {
		blog.Errorf("GetMasterSuggestedMachines getZoneMachineTypes failed: %v", err)
		return nil, err
	}
	blog.Infof("GetMasterSuggestedMachines machineTypeZone: %v", mtZones)

	snZones, zoneSubnetMap, err := getZoneSubnets(vpcId, opt.CommonOption)
	if err != nil {
		blog.Errorf("GetMasterSuggestedMachines getZoneSubnets failed: %v", err)
		return nil, err
	}
	blog.Infof("GetMasterSuggestedMachines subnetZone: %v", snZones)

	// 机型和子网zone求交集
	existZones, _ := utils.SplitExistString(mtZones, snZones)
	if len(existZones) == 0 {
		return nil, fmt.Errorf("available subnet is empty")
	}
	blog.Infof("GetMasterSuggestedMachines machineTypeZone & subnetZone: %v", existZones)

	// 过滤指定zone机型
	filterZones := func() []string {
		if len(opt.Zones) == 0 {
			return existZones
		}

		exist, _ := utils.SplitExistString(existZones, opt.Zones)
		return exist
	}()
	if len(filterZones) == 0 {
		return nil, fmt.Errorf("available subnet is empty")
	}
	blog.Infof("GetMasterSuggestedMachines filterZone: %v", filterZones)

	allocateMachines := utils.AllocateMachinesToAZs(clusterLevel.GetMasterCnt(), len(filterZones))

	// machineCnt < zone || machineCnt > zone
	for cnt := range allocateMachines {
		if len(allocateMachines[cnt]) == 0 {
			continue
		}
		zone := filterZones[cnt]
		insCnt := len(allocateMachines[cnt])

		instanceTemplate = append(instanceTemplate, &proto.InstanceTemplateConfig{
			Zone: zone,
			SubnetID: func() string {
				subnets, ok := zoneSubnetMap[zone]
				if ok && len(subnets) > 0 {
					return subnets[0].SubnetID
				}
				return ""
			}(),
			ApplyNum:       uint32(insCnt),
			CPU:            zoneInstanceTypes[zone][0].GetCpu(),
			Mem:            zoneInstanceTypes[zone][0].GetMemory(),
			GPU:            zoneInstanceTypes[zone][0].GetGpu(),
			InstanceType:   zoneInstanceTypes[zone][0].NodeType,
			SystemDisk:     clusterLevel.GetSystemDisk(),
			CloudDataDisks: []*proto.CloudDataDisk{clusterLevel.GetDataDisk()},
		})

		blog.Infof("zone[%s] allocate[%v] instances", filterZones[cnt], insCnt)
	}

	return instanceTemplate, nil
}

func getZoneSubnets(vpcId string, opt cloudprovider.CommonOption) ([]string, map[string][]*proto.Subnet, error) {
	var (
		snZones = make([]string, 0)
	)

	vpcCli := &VPCManager{}
	subnets, err := vpcCli.ListSubnets(vpcId, "", &cloudprovider.ListNetworksOption{CommonOption: opt})
	if err != nil {
		blog.Errorf("getZoneSubnets ListSubnets failed: %v", err)
		return nil, nil, err
	}

	zoneSubnetMap := make(map[string][]*proto.Subnet, 0)
	for i := range subnets {
		// subnet default available when availableIpCount > 5
		if subnets[i].AvailableIPAddressCount >= 5 {
			if zoneSubnetMap[subnets[i].Zone] == nil {
				zoneSubnetMap[subnets[i].Zone] = make([]*proto.Subnet, 0)
			}
			zoneSubnetMap[subnets[i].Zone] = append(zoneSubnetMap[subnets[i].Zone], subnets[i])
		}
	}

	for zone := range zoneSubnetMap {
		snZones = append(snZones, zone)
	}
	return snZones, zoneSubnetMap, nil
}

func getZoneMachineTypes(
	cpu, mem int, opt cloudprovider.CommonOption) ([]string, map[string][]*proto.InstanceType, error) {
	var (
		mtZones = make([]string, 0)
	)

	nodeCli := &NodeManager{}
	instanceTypes, err := nodeCli.getCloudInstanceType(cloudprovider.InstanceInfo{
		Cpu:    uint32(cpu),
		Memory: uint32(mem),
	}, &opt)
	if err != nil {
		blog.Errorf("getZoneMachineTypes getCloudInstanceType failed: %v", err)
		return nil, nil, err
	}

	// 按照区域机型排序
	zoneInstanceTypes := make(map[string][]*proto.InstanceType, 0)
	for i := range instanceTypes {
		if instanceTypes[i].GetStatus() == icommon.InstanceSoldOut {
			continue
		}

		for j := range instanceTypes[i].Zones {
			_, exist := zoneInstanceTypes[instanceTypes[i].Zones[j]]
			if !exist {
				if zoneInstanceTypes[instanceTypes[i].Zones[j]] == nil {
					zoneInstanceTypes[instanceTypes[i].Zones[j]] = make([]*proto.InstanceType, 0)
				}
				zoneInstanceTypes[instanceTypes[i].Zones[j]] = append(zoneInstanceTypes[instanceTypes[i].Zones[j]],
					instanceTypes[i])
				continue
			}

			zoneInstanceTypes[instanceTypes[i].Zones[j]] = append(zoneInstanceTypes[instanceTypes[i].Zones[j]],
				instanceTypes[i])
		}
	}
	// 按照价格排序
	for zone := range zoneInstanceTypes {
		mtZones = append(mtZones, zone)
		sort.Sort(utils.InstanceTypeSlice(zoneInstanceTypes[zone]))
	}

	return mtZones, zoneInstanceTypes, nil
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

	status, err := client.GetClusterEndpointStatus(clusterID, isExtranet)
	if err != nil {
		return false, err
	}

	blog.Infof("cluster endpoint status: %s", status)

	if !status.Created() {
		return false, fmt.Errorf("cluster endpoint status is not created")
	}

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
	if (cluster.ManageType == icommon.ClusterManageTypeManaged ||
		cluster.ManageType == icommon.ClusterManageTypeIndependent) && !utils.ExistRunningNodes(nodes) {
		blog.Infof("CheckIfGetNodesFromCluster[%s] successful", cluster.ClusterID)
		return false
	}

	return true
}

// SwitchClusterNetwork switch cluster network mode
func (c *Cluster) SwitchClusterNetwork(
	cls *proto.Cluster, subnet *proto.SubnetSource, opt *cloudprovider.SwitchClusterNetworkOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// CheckClusterNetworkStatus get cluster network
func (c *Cluster) CheckClusterNetworkStatus(cloudID string,
	opt *cloudprovider.CheckClusterNetworkStatusOption) (bool, error) {
	return false, cloudprovider.ErrCloudNotImplemented
}

// UpdateCloudKubeConfig update cluster kubeconfig to clustercredential
func (c *Cluster) UpdateCloudKubeConfig(kubeConfig string,
	opt *cloudprovider.UpdateCloudKubeConfigOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

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
