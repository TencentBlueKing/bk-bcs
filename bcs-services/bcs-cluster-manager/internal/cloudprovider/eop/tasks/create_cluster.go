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

// Package tasks xxx
package tasks

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/eop/api"
	providerutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/encrypt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// CreateECKClusterTask call eopcloud interface to create cluster
func CreateECKClusterTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CreateECKClusterTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CreateECKClusterTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeGroupIDs := step.Params[cloudprovider.NodeGroupIDKey.String()]

	// depend basic info
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CreateECKClusterTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s step %s failed %s",
			taskID, clusterID, taskID, stepName, err)
		retErr := fmt.Errorf("get cloud/project information failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// split node groups
	nodeGroups := make([]*proto.NodeGroup, 0)
	for _, ngID := range strings.Split(nodeGroupIDs, ",") {
		nodeGroup, errGet := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if errGet != nil {
			blog.Errorf("CreateECKClusterTask[%s]: GetNodeGroupByGroupID for cluster %s in task %s step %s failed %s",
				taskID, clusterID, taskID, stepName, err)
			retErr := fmt.Errorf("get nodegroup information failed, %s", err)
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
		nodeGroups = append(nodeGroups, nodeGroup)
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// create cluster task
	cls, err := createECKCluster(ctx, dependInfo, nodeGroups)
	if err != nil {
		blog.Errorf("CreateECKClusterTask[%s] createECKCluster for cluster[%s] failed, %s",
			taskID, clusterID, err.Error())
		retErr := fmt.Errorf("createECKCluster err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	// inject cloud cluster id
	state.Task.CommonParams[cloudprovider.CloudSystemID.String()] = cls.ClusterId

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateECKClusterTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

// createECKCluster create eck cluster
func createECKCluster(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, groups []*proto.NodeGroup) (
	*api.CreateClusterReObj, error) {
	var err error
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// init eck client
	eckCli, err := api.NewCTClient(info.CmOption)
	if err != nil {
		blog.Errorf("createECKCluster[%s]: get ECK client for cluster[%s] failed, %s",
			taskID, info.Cluster.ClusterID, err.Error())
		retErr := fmt.Errorf("get eck client err, %s", err.Error())
		return nil, retErr
	}

	// vpc list
	vpcs, err := eckCli.ListVpcs(info.Cluster.Region)
	if err != nil {
		blog.Errorf("createECKCluster ListVpcs failed, %v", err)
		return nil, err
	}
	vpcId, subnetId, err := getVpcIdAndSubnetId(vpcs, info.Cluster)
	if err != nil {
		blog.Errorf("createECKCluster getVpcIdAndSubnetId failed, %v", err)
		return nil, err
	}

	// generateCreateECKReq build create cluster request
	req, err := generateCreateECKReq(info, vpcId, subnetId, groups)
	if err != nil {
		blog.Errorf("createECKCluster eck client CreateCluster failed, %v", err)
		return nil, err
	}

	// create cluster
	eckCluster, err := eckCli.CreateCluster(req)
	if err != nil {
		blog.Errorf("createECKCluster eck client CreateCluster failed, %v", err)
		return nil, err
	}

	// uodate cluster cloud id
	err = cloudprovider.UpdateClusterSystemID(info.Cluster.ClusterID, eckCluster.ClusterId)
	if err != nil {
		blog.Errorf("createECKCluster[%s] updateClusterSystemID[%s] failed %s",
			taskID, info.Cluster.ClusterID, err.Error())
		retErr := fmt.Errorf("call createECKCluster updateClusterSystemID[%s] api err: %s",
			info.Cluster.ClusterID, err.Error())
		return nil, retErr
	}
	blog.Infof("createECKCluster[%s] call CreateTKECluster updateClusterSystemID successful", taskID)

	return eckCluster, nil
}

// generateCreateECKReq build create cluster request
func generateCreateECKReq(info *cloudprovider.CloudDependBasicInfo, vpcId, subnetId uint32, groups []*proto.NodeGroup) (
	*api.CreateClusterRequest, error) {
	var err error
	compnents := make([]*api.Component, 0)
	compnents = append(compnents, &api.Component{Name: "prometheus"}, &api.Component{Name: "node-local-dns"})
	blog.Infof("----------- generateCreateECKReq k8s version is %s", info.Cluster.ClusterBasicSettings.Version)

	// build create cluster request
	req := &api.CreateClusterRequest{
		Components: compnents,
		ContainerRuntime: &api.ContainerRuntime{
			Name:    info.Cluster.ClusterAdvanceSettings.ContainerRuntime,
			Version: info.Cluster.ClusterAdvanceSettings.RuntimeVersion,
		},
		CustomName:             info.Cluster.ClusterName,
		Description:            info.Cluster.Description,
		EnableDeleteProtection: info.Cluster.ClusterAdvanceSettings.DeletionProtection,
		IPNum:                  info.Cluster.NetworkSettings.MaxNodePodNum,
		K8sVersion:             info.Cluster.ClusterBasicSettings.Version,
		KubeProxyMode:          generateKubeProxyMode(info.Cluster),
		Labels:                 generateLabels(info.Cluster),
		PodCidr:                info.Cluster.NetworkSettings.ClusterIPv4CIDR,
		ServiceCidr:            info.Cluster.NetworkSettings.ServiceIPv4CIDR,
		SlbConfig: &api.SlbConfig{
			AllocEip: true,
			BwSize:   5,
		},
		WorkerNodes: nil,
	}

	// generateK8SExtension build k8s extension
	req.K8SExtension, err = generateK8SExtension(info.Cluster)
	if err != nil {
		blog.Errorf("createECKCluster generateK8SExtension failed, %v", err)
		return nil, err
	}

	// generateMasterNodes build master nodes
	req.MasterNodes, err = generateMasterNodes(info.Cluster, vpcId, subnetId)
	if err != nil {
		blog.Errorf("createECKCluster generateMasterNodes failed, %v", err)
		return nil, err
	}

	// generateWorkerNodes build worker nodes
	req.WorkerNodes, err = generateWorkerNodes(info.Cluster, vpcId, subnetId, groups)
	if err != nil {
		blog.Errorf("createECKCluster generateWorkerNodes failed, %v", err)
		return nil, err
	}

	return req, nil
}

// generateK8SExtension build k8s extension
func generateK8SExtension(cls *proto.Cluster) (string, error) {
	extension, ok := cls.ExtraInfo[common.CloudClusterTypeKey]
	if !ok {
		return common.CloudClusterTypeEdge, nil
	}
	if extension != common.CloudClusterTypeEdge && extension != common.CloudClusterTypeNative {
		return "", fmt.Errorf("invalid CloudClusterType %s", extension)
	}

	return extension, nil
}

// generate Kube Proxy Mode
func generateKubeProxyMode(cls *proto.Cluster) string {
	if cls.ClusterAdvanceSettings.IPVS {
		return "KUBEPROXYMODE_IPVS"
	}

	return "KUBEPROXYMODE_IPTABLES"
}

// generateLabels build labels
func generateLabels(cls *proto.Cluster) []*api.Label {
	labels := make([]*api.Label, 0)
	for k, v := range cls.Labels {
		labels = append(labels, &api.Label{Key: k, Value: v})
	}

	return labels
}

// generateWorkerNodes build worker nodes
func generateWorkerNodes(cls *proto.Cluster, vpcId, subnetId uint32, groups []*proto.NodeGroup) (
	*api.WorkerNode, error) {
	workerNodes := make([]*api.WorkerNode, 0)

	// build worker nodes
	for _, ng := range groups {
		if ng.AutoScaling == nil {
			return nil, fmt.Errorf("empty AutoScaling for cluster %s NodeGroup %s", cls.ClusterID, ng.Name)
		}

		if ng.LaunchTemplate == nil {
			return nil, fmt.Errorf("empty LaunchTemplate for cluster %s NodeGroup %s", cls.ClusterID, ng.Name)
		}

		if ng.LaunchTemplate.InitLoginPassword == "" {
			return nil, fmt.Errorf("empty InitLoginPassword for cluster %s NodeGroup %s", cls.ClusterID, ng.Name)
		}

		if ng.LaunchTemplate.SystemDisk == nil {
			return nil, fmt.Errorf("empty system disk for cluster %s NodeGroup %s", cls.ClusterID, ng.Name)
		}
		systemDiskSize, err := strconv.Atoi(ng.LaunchTemplate.SystemDisk.DiskSize)
		if err != nil {
			return nil, fmt.Errorf("invalid system disk size %s", ng.LaunchTemplate.SystemDisk.DiskSize)
		}

		passwd, err := encrypt.Decrypt(nil, ng.LaunchTemplate.InitLoginPassword)
		if err != nil {
			return nil, fmt.Errorf("generateWorkerNodes for cluster %s NodeGroup %s decrypt password failed, %v",
				cls.ClusterID, ng.Name, err)
		}

		dataDisks := make([]*api.Disk, 0)
		for _, d := range ng.LaunchTemplate.DataDisks {
			size, err := strconv.Atoi(d.DiskSize)
			if err != nil {
				return nil, fmt.Errorf("invalid data disk size %s for cluster %s NodeGroup %s",
					d.DiskSize, cls.ClusterID, ng.Name)
			}
			dataDisks = append(dataDisks, &api.Disk{
				Count:    1,
				IOType:   api.DisKIOTypeNormal,
				Size:     uint32(size),
				DiskType: d.DiskType,
			})
		}

		workerNodes = append(workerNodes, &api.WorkerNode{
			DataDisks:     dataDisks,
			ImageName:     cls.ClusterBasicSettings.OS,
			MountLastDisk: false,
			NetworkInfo: &api.NetworkInfo{
				SubnetId: subnetId,
				VpcId:    vpcId,
			},
			NodeCode:     cls.Region,
			NodePoolName: ng.NodeGroupID,
			Num:          ng.AutoScaling.DesiredSize,
			Password:     passwd,
			SystemDisk: &api.Disk{
				Count:    1,
				IOType:   api.DisKIOTypeNormal,
				Size:     uint32(systemDiskSize),
				DiskType: ng.LaunchTemplate.SystemDisk.DiskType,
			},
			VmInstanceName: ng.LaunchTemplate.InstanceType,
		})
	}

	return workerNodes[0], nil
}

// generateMasterNodes build master nodes
func generateMasterNodes(cls *proto.Cluster, vpcId, subnetId uint32) (*api.MasterNode, error) {
	if cls.Template[0].InitLoginPassword == "" {
		return nil, fmt.Errorf("empty InitLoginPassword for cluster %s", cls.ClusterID)
	}

	// build system disks
	if cls.Template[0].SystemDisk == nil {
		return nil, fmt.Errorf("empty system disk for cluster %s", cls.ClusterID)
	}
	systemDiskSize, err := strconv.Atoi(cls.Template[0].SystemDisk.DiskSize)
	if err != nil {
		return nil, fmt.Errorf("invalid system disk size %s for cluster %s",
			cls.Template[0].SystemDisk.DiskSize, cls.ClusterID)
	}

	// parse passwd
	passwd, err := encrypt.Decrypt(nil, cls.Template[0].InitLoginPassword)
	if err != nil {
		return nil, fmt.Errorf("generateWorkerNodes for cluster %s decrypt password failed, %v",
			cls.ClusterID, err)
	}

	// data disks
	dataDisks := make([]*api.Disk, 0)
	for _, d := range cls.Template[0].DataDisks {
		size, err := strconv.Atoi(d.DiskSize)
		if err != nil {
			return nil, fmt.Errorf("invalid data disk size %s", d.DiskSize)
		}
		dataDisks = append(dataDisks, &api.Disk{
			Count:    1,
			IOType:   api.DisKIOTypeNormal,
			Size:     uint32(size),
			DiskType: d.DiskType,
		})
	}

	// build master nodes
	masterNode := &api.MasterNode{
		DataDisks:     dataDisks,
		ImageName:     cls.ClusterBasicSettings.OS,
		MountLastDisk: false,
		NetworkInfo: &api.NetworkInfo{
			SubnetId: subnetId,
			VpcId:    vpcId,
		},
		NodeCode: cls.Region,
		Num:      cls.Template[0].ApplyNum,
		Password: passwd,
		SystemDisk: &api.Disk{
			Count:    1,
			IOType:   api.DisKIOTypeNormal,
			Size:     uint32(systemDiskSize),
			DiskType: cls.Template[0].SystemDisk.DiskType,
		},
		VmInstanceName: cls.Template[0].InstanceType,
	}

	return masterNode, nil
}

// getVpcIdAndSubnetId get vpc/subnet
func getVpcIdAndSubnetId(vpcs []*api.Vpc, cls *proto.Cluster) (uint32, uint32, error) {
	var vpcId, subnetId uint32
	for _, vpc := range vpcs {
		if vpc.Name == cls.Template[0].VpcID {
			vpcId = vpc.VpcId
			for _, subnet := range vpc.Subnets {
				if subnet.Name == cls.Template[0].SubnetID {
					subnetId = subnet.SubnetId
					break
				}
			}
			break
		}
	}

	if vpcId == 0 || subnetId == 0 {
		return vpcId, subnetId, fmt.Errorf("unavailable vpcId or subnetId")
	}

	return vpcId, subnetId, nil
}

// CheckECKClusterStatusTask check cluster create status
func CheckECKClusterStatusTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckECKClusterStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckECKClusterStatusTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	systemID := state.Task.CommonParams[cloudprovider.CloudSystemID.String()]

	// get depend basic info
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckECKClusterStatusTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s step %s "+
			"failed %s", taskID, clusterID, taskID, stepName, err)
		retErr := fmt.Errorf("get cloud/project information failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = checkClusterStatus(ctx, dependInfo, systemID)
	if err != nil {
		blog.Errorf("CheckECKClusterStatusTask[%s] checkClusterStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("checkClusterStatus[%s] timeout|abnormal", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckECKClusterStatusTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}

	return nil
}

// checkClusterStatus check cluster status
func checkClusterStatus(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, systemID string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// get eopCloud client
	cli, err := api.NewCTClient(info.CmOption)
	if err != nil {
		blog.Errorf("checkClusterStatus[%s] get eck client failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("get cloud eck client err, %s", err.Error())
		return retErr
	}

	var (
		failed = false
	)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// loop cluster status
	err = loop.LoopDoFunc(ctx, func() error {
		cluster, errGet := cli.GetCluster(systemID)
		if errGet != nil {
			blog.Errorf("checkClusterStatus[%s] failed: %v", taskID, errGet)
			return nil
		}

		blog.Infof("checkClusterStatus[%s] cluster[%s] current status[%s]", taskID,
			info.Cluster.ClusterID, cluster.State)

		switch cluster.State {
		case api.ClusterStatusRunning:
			return loop.EndLoop
		case api.ClusterStatusCreateFailed:
			failed = true
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("checkClusterStatus[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, err)
		return err
	}

	// failed status
	if failed {
		blog.Errorf("checkClusterStatus[%s] GeteckCluster[%s] failed: abnormal", taskID, info.Cluster.ClusterID)
		retErr := fmt.Errorf("cluster[%s] status abnormal", info.Cluster.ClusterID)
		return retErr
	}

	return nil
}

// CheckECKNodesGroupStatusTask check cluster nodes status
func CheckECKNodesGroupStatusTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckECKNodesGroupStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckECKNodesGroupStatusTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeGroupIDs := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params,
		cloudprovider.NodeGroupIDKey.String(), ",")
	systemID := state.Task.CommonParams[cloudprovider.CloudSystemID.String()]

	// get depend basic info
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckECKNodesGroupStatusTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed %s", taskID, clusterID, taskID, stepName, err)
		retErr := fmt.Errorf("get cloud/project information failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	addSuccessNodeGroups, addFailureNodeGroups, err := checkNodesGroupStatus(ctx, dependInfo, systemID, nodeGroupIDs)
	if err != nil {
		blog.Errorf("CheckECKNodesGroupStatusTask[%s] checkNodesGroupStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("CheckECKNodesGroupStatusTask[%s] timeout|abnormal", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	if len(addFailureNodeGroups) > 0 {
		state.Task.CommonParams[cloudprovider.FailedNodeGroupIDsKey.String()] =
			strings.Join(addFailureNodeGroups, ",")
	}

	if len(addSuccessNodeGroups) == 0 {
		blog.Errorf("CheckECKNodesGroupStatusTask[%s] nodegroups init failed", taskID)
		retErr := fmt.Errorf("节点池初始化失败, 请联系管理员")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	state.Task.CommonParams[cloudprovider.SuccessNodeGroupIDsKey.String()] =
		strings.Join(addSuccessNodeGroups, ",")

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckECKNodesGroupStatusTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}

	return nil
}

// checkNodesGroupStatus check nodeGroup status
func checkNodesGroupStatus(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	systemID string, nodeGroupIDs []string) ([]string, []string, error) {

	var (
		addSuccessNodeGroups = make([]string, 0)
		addFailureNodeGroups = make([]string, 0)
	)

	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// get node groups
	nodeGroups := make([]*proto.NodeGroup, 0)
	for _, ngID := range nodeGroupIDs {
		nodeGroup, err := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if err != nil {
			return nil, nil, fmt.Errorf("checkNodesGroupStatus GetNodeGroupByGroupID failed, %s", err.Error())
		}
		nodeGroups = append(nodeGroups, nodeGroup)
	}
	// get eopCloud client
	cli, err := api.NewCTClient(info.CmOption)
	if err != nil {
		blog.Errorf("checkNodesGroupStatus[%s] get eck client failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("get cloud eck client err, %s", err.Error())
		return nil, nil, retErr
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	running, failure := make([]string, 0), make([]string, 0)
	// loop cluster groups status
	err = loop.LoopDoFunc(ctx, func() error {
		index := 0
		for _, ng := range nodeGroups {
			eckNodePools, errQuery := cli.ListNodePool(&api.ListNodePoolReq{
				ClusterID:    systemID,
				NodePoolName: ng.NodeGroupID,
			})
			if errQuery != nil {
				blog.Errorf("checkNodesGroupStatus[%s] failed: %v", taskID, err)
				return nil
			}
			if len(eckNodePools) == 0 {
				blog.Errorf("checkNodesGroupStatus[%s] ListNodePool[%s] not found", taskID, ng.NodeGroupID)
				return nil
			}

			blog.Infof("checkNodesGroupStatus[%s] nodeGroup[%s] status %s",
				taskID, ng.NodeGroupID, eckNodePools[0].State)

			switch eckNodePools[0].State {
			case api.NodePoolStatusActive:
				running = append(running, ng.NodeGroupID)
				index++
			case api.NodePoolStatusCreateFailed:
				failure = append(failure, ng.NodeGroupID)
				index++
			}
		}
		if index == len(nodeGroups) {
			addSuccessNodeGroups = running
			addFailureNodeGroups = failure
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("checkNodesGroupStatus[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, err)
		return nil, nil, err
	}

	blog.Infof("checkNodesGroupStatus[%s] success[%v] failure[%v]",
		taskID, addSuccessNodeGroups, addFailureNodeGroups)

	return addSuccessNodeGroups, addFailureNodeGroups, nil
}

// UpdateECKNodesGroupToDBTask update ECK nodepool
func UpdateECKNodesGroupToDBTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("UpdateECKNodesGroupToDBTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("UpdateECKNodesGroupToDBTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	addSuccessNodeGroupIDs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.SuccessNodeGroupIDsKey.String(), ",")
	addFailedNodeGroupIDs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.FailedNodeGroupIDsKey.String(), ",")

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("UpdateECKNodesGroupToDBTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed %s", taskID, clusterID, taskID, stepName, err)
		retErr := fmt.Errorf("get cloud/project information failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = updateNodeGroups(ctx, dependInfo, addFailedNodeGroupIDs, addSuccessNodeGroupIDs)
	if err != nil {
		blog.Errorf("UpdateECKNodesGroupToDBTask[%s] checkNodesGroupStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("UpdateECKNodesGroupToDBTask[%s] timeout|abnormal", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateECKNodesGroupToDBTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}

	return nil
}

// updateNodeGroups update node groups
func updateNodeGroups(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	addFailedNodeGroupIDs, addSuccessNodeGroupIDs []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	if len(addFailedNodeGroupIDs) > 0 {
		for _, ngID := range addFailedNodeGroupIDs {
			err := cloudprovider.UpdateNodeGroupStatus(ngID, common.StatusCreateNodeGroupFailed)
			if err != nil {
				return fmt.Errorf("updateNodeGroups updateNodeGroupStatus[%s] failed, %v", ngID, err)
			}
		}
	}

	for _, ngID := range addSuccessNodeGroupIDs {
		nodeGroup, err := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if err != nil {
			return fmt.Errorf("updateNodeGroups GetNodeGroupByGroupID failed, %s", err.Error())
		}

		// get eopCloud client
		cli, err := api.NewCTClient(info.CmOption)
		if err != nil {
			blog.Errorf("updateNodeGroups[%s] get eck client failed: %s", taskID, err.Error())
			return fmt.Errorf("get cloud eck client err, %s", err.Error())
		}

		eckNodePools, err := cli.ListNodePool(&api.ListNodePoolReq{
			ClusterID:    info.Cluster.SystemID,
			NodePoolName: ngID,
		})
		if err != nil {
			blog.Errorf("updateNodeGroups[%s] ListNodePool failed: %v", taskID, err)
			return nil
		}

		nodeGroup.CloudNodeGroupID = eckNodePools[0].NodePoolId
		nodeGroup.Status = common.StatusRunning
		err = cloudprovider.GetStorageModel().UpdateNodeGroup(context.Background(), nodeGroup)
		if err != nil {
			return fmt.Errorf("updateNodeGroups UpdateNodeGroup failed, %s", err.Error())
		}
	}

	return nil
}

// CheckECKClusterNodesStatusTask check cluster nodes status
func CheckECKClusterNodesStatusTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckECKClusterNodesStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckECKClusterNodesStatusTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	systemID := state.Task.CommonParams[cloudprovider.CloudSystemID.String()]
	nodeGroupIDs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.SuccessNodeGroupIDsKey.String(), ",")

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckECKClusterNodesStatusTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	addSuccessNodes, addFailureNodes, err := checkClusterNodesStatus(ctx, dependInfo, systemID, nodeGroupIDs)
	if err != nil {
		blog.Errorf("CheckECKClusterStatusTask[%s] checkClusterStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("checkClusterStatus[%s] timeout|abnormal", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	if len(addFailureNodes) > 0 {
		state.Task.CommonParams[cloudprovider.FailedClusterNodeIDsKey.String()] = strings.Join(addFailureNodes, ",")
	}
	if len(addSuccessNodes) == 0 {
		blog.Errorf("CheckCreateClusterNodeStatusTask[%s] nodes init failed", taskID)
		retErr := fmt.Errorf("节点初始化失败, 请联系管理员")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	state.Task.CommonParams[cloudprovider.SuccessClusterNodeIDsKey.String()] = strings.Join(addSuccessNodes, ",")

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckCreateClusterNodeStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// checkClusterNodesStatus check cluster nodes status
func checkClusterNodesStatus(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, // nolint
	systemID string, nodeGroupIDs []string) ([]string, []string, error) {
	var (
		totalNodesNum   uint32
		addSuccessNodes = make([]string, 0)
		addFailureNodes = make([]string, 0)
	)

	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	nodePoolList := make([]string, 0)
	for _, ngID := range nodeGroupIDs {
		nodeGroup, err := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if err != nil {
			return nil, nil, fmt.Errorf("get nodegroup information failed, %s", err.Error())
		}
		totalNodesNum += nodeGroup.AutoScaling.DesiredSize
		nodePoolList = append(nodePoolList, nodeGroup.CloudNodeGroupID)
	}

	// get eopCloud client
	cli, err := api.NewCTClient(info.CmOption)
	if err != nil {
		blog.Errorf("checkClusterNodesStatus[%s] get eck client failed: %s", taskID, err.Error())
		return nil, nil, fmt.Errorf("get cloud eck client err, %s", err.Error())
	}

	result, err := cli.ListNodePool(&api.ListNodePoolReq{
		ClusterID:            systemID,
		NodePoolName:         api.MasterNodePoolName,
		RetainSystemNodePool: true,
	})
	if err != nil {
		blog.Errorf("checkClusterNodesStatus[%s] failed: %v", taskID, err)
		return nil, nil, err
	}
	if len(result) == 0 {
		blog.Errorf("checkClusterNodesStatus[%s] failed, master node pool not found", taskID)
		return nil, nil, fmt.Errorf("checkClusterNodesStatus[%s] failed, master node pool not found", taskID)
	}

	// add master node num
	totalNodesNum += info.Cluster.Template[0].ApplyNum
	// master节点的节点池由ecp自动创建, 有默认名称
	nodePoolList = append(nodePoolList, result[0].NodePoolId)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// loop cluster status
	err = loop.LoopDoFunc(ctx, func() error {
		// ListNodes lists ECK cluster nodes
		nodes, errGet := cli.ListNodes(&api.ListNodeReq{
			ClusterID: systemID,
			Page:      1,
			PerPage:   totalNodesNum,
		})
		if errGet != nil {
			blog.Errorf("checkClusterNodesStatus[%s] failed: %v", taskID, errGet)
			return nil
		}

		blog.Infof("checkClusterNodesStatus[%s] expected nodes %d , current nodes %d ",
			taskID, totalNodesNum, len(nodes))

		running, failure := make([]string, 0), make([]string, 0)
		index := 0
		for _, node := range nodes {
			if !utils.StringInSlice(node.NodePoolId, nodePoolList) {
				continue
			}
			blog.Infof("checkClusterNodesStatus[%s] node[%s] state %s", taskID, node.NodeName, node.State)
			switch node.State {
			case api.NodeStatusRunning:
				index++
				running = append(running, node.NodeName)
			case api.NodeStatusUnknown:
				failure = append(failure, node.NodeName)
				index++
			}
		}

		if index == int(totalNodesNum) {
			addSuccessNodes = running
			addFailureNodes = failure
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(10*time.Second))
	// other error
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		blog.Errorf("checkClusterNodesStatus[%s] ListNodes failed: %v", taskID, err)
		return nil, nil, err
	}
	// timeout error
	if errors.Is(err, context.DeadlineExceeded) {
		blog.Errorf("checkClusterNodesStatus[%s] ListNodes failed: %v", taskID, err)

		running, failure := make([]string, 0), make([]string, 0)
		// ListNodes lists ECK cluster nodes
		nodes, errQuery := cli.ListNodes(&api.ListNodeReq{
			ClusterID: systemID,
			Page:      1,
			PerPage:   totalNodesNum,
		})
		if errQuery != nil {
			blog.Errorf("checkClusterNodesStatus[%s] ListNodes failed: %v", taskID, errQuery)
			return nil, nil, errQuery
		}
		for _, n := range nodes {
			blog.Infof("checkClusterNodesStatus[%s] instance[%s] status[%s]", taskID, n.NodeName, n.State)
			switch n.State {
			case api.NodeStatusRunning:
				running = append(running, n.NodeName)
			default:
				failure = append(failure, n.NodeName)
			}
		}
		addSuccessNodes = running
		addFailureNodes = failure
	}
	blog.Infof("checkClusterNodesStatus[%s] success[%v] failure[%v]", taskID, addSuccessNodes, addFailureNodes)

	// set cluster node status
	for _, n := range addFailureNodes {
		err = cloudprovider.UpdateNodeStatus(false, n, common.StatusAddNodesFailed)
		if err != nil {
			blog.Errorf("checkClusterNodesStatus[%s] UpdateNodeStatus[%s] failed: %v", taskID, n, err)
		}
	}

	return addSuccessNodes, addFailureNodes, nil
}

// UpdateECKNodesToDBTask update ECK nodes
func UpdateECKNodesToDBTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("UpdateNodesToDBTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("UpdateNodesToDBTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	// ParseNodeIpOrIdFromCommonMap parse nodeIDs or nodeIPs by chart
	nodes := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.SuccessClusterNodeIDsKey.String(), ",")
	// ParseNodeIpOrIdFromCommonMap parse nodeIDs or nodeIPs by chart
	nodeGroupIDs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.SuccessNodeGroupIDsKey.String(), ",")

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("UpdateNodesToDBTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = updateNodeToDB(ctx, dependInfo, nodes, nodeGroupIDs)
	if err != nil {
		blog.Errorf("UpdateNodesToDBTask[%s] checkNodesGroupStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("UpdateNodesToDBTask[%s] timeout|abnormal", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// sync clusterData to pass-cc
	providerutils.SyncClusterInfoToPassCC(taskID, dependInfo.Cluster)

	// sync cluster perms
	providerutils.AuthClusterResourceCreatorPerm(ctx, dependInfo.Cluster.ClusterID,
		dependInfo.Cluster.ClusterName, dependInfo.Cluster.Creator)

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateNodesToDBTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}

	return nil
}

// update node to DB
func updateNodeToDB(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, nodes, nodeGroupIDs []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	nodeGroups := make([]*proto.NodeGroup, 0)
	for _, ngID := range nodeGroupIDs {
		// GetNodeGroupByGroupID get nodeGroup info
		nodeGroup, err := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if err != nil {
			return fmt.Errorf("get nodegroup information failed, %s", err.Error())
		}
		nodeGroups = append(nodeGroups, nodeGroup)
	}

	// get eopCloud client
	cli, err := api.NewCTClient(info.CmOption)
	if err != nil {
		blog.Errorf("updateNodeToDB[%s] get eck client failed: %s", taskID, err.Error())
		return fmt.Errorf("updateNodeToDB get cloud eck client err, %s", err.Error())
	}

	for _, n := range nodes {
		// ListNodes lists ECK cluster nodes
		result, errGet := cli.ListNodes(&api.ListNodeReq{
			ClusterID: info.Cluster.SystemID,
			NodeNames: n,
		})
		if errGet != nil {
			blog.Errorf("updateNodeToDB[%s] ListNodes failed: %v", taskID, errGet)
			return nil
		}

		if result[0].Role == api.NodeRoleMaster {
			node := &proto.Node{
				NodeID:       result[0].InstanceId,
				InnerIP:      result[0].InnerIp,
				InstanceType: info.Cluster.Template[0].InstanceType,
				ClusterID:    info.Cluster.ClusterID,
				VPC:          info.Cluster.Template[0].VpcID,
				Region:       info.Cluster.Template[0].Region,
				NodeName:     result[0].NodeName,
				Status:       common.StatusRunning,
			}
			// create node
			err = cloudprovider.GetStorageModel().CreateNode(context.Background(), node)
			if err != nil {
				return fmt.Errorf("updateNodeToDB CreateNode[%s] failed, %v", node.NodeName, err)
			}
			continue
		}

		for _, ng := range nodeGroups {
			if result[0].NodePoolId == ng.CloudNodeGroupID {
				node := &proto.Node{
					NodeID:       result[0].InstanceId,
					InnerIP:      result[0].InnerIp,
					InstanceType: ng.LaunchTemplate.InstanceType,
					NodeGroupID:  ng.NodeGroupID,
					ClusterID:    info.Cluster.ClusterID,
					VPC:          ng.AutoScaling.VpcID,
					Region:       ng.Region,
					NodeName:     result[0].NodeName,
					Status:       common.StatusRunning,
				}
				// create node
				err = cloudprovider.GetStorageModel().CreateNode(context.Background(), node)
				if err != nil {
					return fmt.Errorf("updateNodeToDB CreateNode[%s] failed, %v", node.NodeName, err)
				}
				break
			}
		}
	}

	return nil
}

// RegisterClusterKubeConfigTask register cluster kubeconfig
func RegisterClusterKubeConfigTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("RegisterClusterKubeConfigTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("RegisterClusterKubeConfigTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	// handler logic
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("RegisterClusterKubeConfigTask[%s] GetClusterDependBasicInfo in task %s step %s failed, %s",
			taskID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	err = importClusterCredential(ctx, dependInfo)
	if err != nil {
		blog.Errorf("RegisterClusterKubeConfigTask[%s] importClusterCredential failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("importClusterCredential failed %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	blog.Infof("RegisterClusterKubeConfigTask[%s] importClusterCredential success", taskID)

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("RegisterClusterKubeConfigTask[%s:%s] update to storage fatal", taskID, stepName)
		return err
	}

	return nil
}

// import Cluster Credential
func importClusterCredential(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// get eopCloud client
	cli, err := api.NewCTClient(info.CmOption)
	if err != nil {
		blog.Errorf("importClusterCredential[%s] get eck client failed: %s", taskID, err.Error())
		return fmt.Errorf("importClusterCredential get cloud eck client err, %s", err.Error())
	}

	// GetKubeConfig gets kubeconfig
	result, err := cli.GetKubeConfig(info.Cluster.SystemID)
	if err != nil {
		return fmt.Errorf("importClusterCredential[%s] GetKubeConfig failed, %v", taskID, err)
	}

	if result.ExternalKubeConfig == nil {
		return fmt.Errorf("importClusterCredential[%s] GetKubeConfig failed, empty ExternalKubeConfig", taskID)
	}

	// GetKubeConfigFromYAMLBody get kubeConfig from YAML file
	kubeConfig, err := types.GetKubeConfigFromYAMLBody(false, types.YamlInput{
		FileName:    "",
		YamlContent: result.ExternalKubeConfig.Content,
	})
	if err != nil {
		return fmt.Errorf("importClusterCredential[%s] GetKubeConfigFromYAMLBody failed, %v", taskID, err)
	}
	// UpdateClusterCredentialByConfig update clusterCredential by kubeConfig
	err = cloudprovider.UpdateClusterCredentialByConfig(info.Cluster.ClusterID, kubeConfig)
	if err != nil {
		return err
	}

	return nil
}
