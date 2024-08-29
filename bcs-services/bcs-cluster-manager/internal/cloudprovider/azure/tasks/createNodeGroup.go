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

package tasks

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
)

// CreateCloudNodeGroupTask 创建节点池 - create cloud node group task
func CreateCloudNodeGroupTask(taskID string, stepName string) error {
	start := time.Now()
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName) // get task and task current step
	if err != nil {
		return err
	}
	if step == nil { // previous step successful when retry task
		return nil
	}
	// extract parameter
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID) // inject taskID
	if len(clusterID) == 0 || len(nodeGroupID) == 0 || len(cloudID) == 0 {  // check validate
		blog.Errorf("CreateCloudNodeGroupTask[%s]: check parameter validate failed", taskID)
		retErr := fmt.Errorf("CreateCloudNodeGroupTask check parameters failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: GetClusterDependBasicInfo failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("CreateCloudNodeGroupTask GetClusterDependBasicInfo failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	if dependInfo.NodeGroup.AutoScaling == nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: nodegroup %s in task %s step %s has no autoscaling group",
			taskID, nodeGroupID, taskID, stepName)
		retErr := fmt.Errorf("get autoScalingID err, %v", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// create Agent Pool(代理节点池)
	if err = createAgentPool(ctx, dependInfo); err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: call createAgentPool[%s] api in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call createAgentPool[%s] api err, %s", nodeGroupID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// 设置节点密码、购买系统盘、数据盘等
	if err = setVmSets(ctx, dependInfo); err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: call setVmSets[%s] api in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call setVmSets[%s] api err, %s", nodeGroupID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update nodeGroup cloudNodeGroupID
	if err = updateCloudNodeGroupIDInNodeGroup(nodeGroupID, dependInfo.NodeGroup); err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: updateCloudNodeGroupIDInNodeGroup[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call updateCloudNodeGroupIDInNodeGroup[%s] api err, %s",
			nodeGroupID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CreateCloudNodeGroupTask[%s]: call CreateClusterNodePool updateCloudNodeGroupIDInNodeGroup"+
		" successful", taskID)

	if state.Task.CommonParams == nil { // update response information to task common params
		state.Task.CommonParams = make(map[string]string)
	}

	if err = state.UpdateStepSucc(start, stepName); err != nil { // update step
		return errors.Wrapf(err, "CreateCloudNodeGroupTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
	}

	return nil
}

// CheckCloudNodeGroupStatusTask 检查节点池创建状态 - check cloud node group status task
func CheckCloudNodeGroupStatusTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		return nil
	}
	// extract parameter
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	// check validate
	if len(clusterID) == 0 || len(nodeGroupID) == 0 || len(cloudID) == 0 {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: check parameter validate failed", taskID)
		retErr := fmt.Errorf("CheckCloudNodeGroupStatusTask check parameters failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: GetClusterDependBasicInfo failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("CheckCloudNodeGroupStatusTask GetClusterDependBasicInfo failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	if dependInfo.NodeGroup.AutoScaling == nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: nodegroup %s in task %s step %s has no autoscaling group",
			taskID, nodeGroupID, taskID, stepName)
		retErr := fmt.Errorf("get autoScalingID err, %v", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check nodeGroup
	if err = checkNodeGroup(ctx, dependInfo); err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: call checkNodeGroup[%s] api in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call checkNodeGroup[%s] api err, %s", nodeGroupID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CheckCloudNodeGroupStatusTask nodeGroup status:%s", dependInfo.NodeGroup.Status)

	// update bcs nodeGroup
	if err = cloudprovider.GetStorageModel().UpdateNodeGroup(ctx, dependInfo.NodeGroup); err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: updateNodeGroupCloudArgsID[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call UpdateNodeGroup[%s] api err, %s",
			nodeGroupID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s] task %s %s update to storage fatal", taskID, taskID,
			stepName)
		return err
	}
	return nil
}

// createAgentPool 创建节点池
func createAgentPool(rootCtx context.Context, info *cloudprovider.CloudDependBasicInfo) error {
	group := info.NodeGroup
	cluster := info.Cluster

	taskID := cloudprovider.GetTaskIDFromContext(rootCtx)

	// new Azure client
	client, err := api.NewAksServiceImplWithCommonOption(info.CmOption)
	if err != nil {
		return errors.Wrapf(err, "call NewAgentPoolClientWithOpt[%s] falied", taskID)
	}

	// wait node group state to normal
	ctx, cancel := context.WithTimeout(rootCtx, 20*time.Minute)
	defer cancel()

	// set cloudNodeGroupID
	setCloudNodeGroupID(group)

	// create node group
	pool := new(armcontainerservice.AgentPool)
	if err = client.NodeGroupToAgentPool(group, pool); err != nil {
		return errors.Wrapf(err, "createAgentPool[%s]: call NodeGroupToAgentPool failed", taskID)
	}
	_, err = client.CreatePoolAndReturn(ctx, pool, cloudprovider.GetClusterResourceGroup(info.Cluster),
		cluster.SystemID, group.CloudNodeGroupID)
	if err != nil {
		return errors.Wrapf(err, "createAgentPool[%s]: call CreatePoolAndReturn[%s][%s] falied", taskID,
			cluster.ClusterID, group.CloudNodeGroupID)
	}
	blog.Infof("CreateCloudNodeGroupTask[%s]: call CreateClusterNodePool successful", taskID)

	// set default value for nodegroup
	if group.AutoScaling != nil && group.AutoScaling.VpcID == "" {
		group.AutoScaling.VpcID = cluster.VpcID // 继承集群vpc
	}
	if group.LaunchTemplate != nil {
		if group.LaunchTemplate.InstanceChargeType == "" {
			group.LaunchTemplate.InstanceChargeType = "POSTPAID_BY_HOUR" // 按小时后付费
		}
	}

	return nil
}

// setVmSets 设置虚拟机规模集，购买系统盘、数据盘、节点密码与用户名
func setVmSets(rootCtx context.Context, info *cloudprovider.CloudDependBasicInfo) error {
	group := info.NodeGroup
	cluster := info.Cluster
	lc := group.LaunchTemplate

	taskID := cloudprovider.GetTaskIDFromContext(rootCtx)
	if lc == nil || (lc.SystemDisk == nil && len(lc.DataDisks) == 0 && len(lc.InitLoginUsername) == 0 &&
		len(lc.InitLoginPassword) == 0) {
		return nil
	}

	client, err := api.NewAksServiceImplWithCommonOption(info.CmOption) // new Azure client
	if err != nil {
		return errors.Wrapf(err, "call NewAgentPoolClientWithOpt[%s] falied", taskID)
	}

	nodeGroupResource, ok := cluster.ExtraInfo[common.NodeResourceGroup]
	if !ok {
		ctx, cancel := context.WithTimeout(rootCtx, 30*time.Second)
		defer cancel()

		var cloudCluster *armcontainerservice.ManagedCluster

		clsResourceGroup := cloudprovider.GetClusterResourceGroup(info.Cluster)
		cloudCluster, err = client.GetCluster(ctx, info, clsResourceGroup)
		if err != nil {
			return errors.Wrapf(err, "setVmSets[%s]: call GetCluster falied", taskID)
		}
		nodeGroupResource = *cloudCluster.Properties.NodeResourceGroup
	}

	netGroupResource, ok2 := cluster.ExtraInfo[common.NetworkResourceGroup]
	if !ok2 {
		return fmt.Errorf("setVmSets[%s] get netGroupResource failed", taskID)
	}

	ctx, cancel := context.WithTimeout(rootCtx, 30*time.Second)
	defer cancel()
	set, err := client.MatchNodeGroup(ctx, nodeGroupResource, info.NodeGroup.CloudNodeGroupID)
	if err != nil {
		return errors.Wrapf(err, "setVmSets[%s]: call MatchNodeGroup[%s][%s] falied", taskID,
			cluster.ClusterID, group.CloudNodeGroupID)
	}

	_, err = updateVmss(ctx, client, group, set, nodeGroupResource, netGroupResource)
	if err != nil {
		return errors.Wrapf(err, "setVmSets[%s]: call UpdateVmss[%s][%s] falied", taskID,
			cluster.ClusterID, group.CloudNodeGroupID)
	}

	blog.Infof("setVmSets[%s]: %s set node password or purchase data disk and system disk successful",
		taskID, *set.Name)

	return nil
}

// updateCloudNodeGroupIDInNodeGroup set nodegroup cloudNodeGroupID
func updateCloudNodeGroupIDInNodeGroup(nodeGroupID string, newGroup *proto.NodeGroup) error {
	group, err := cloudprovider.GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		return err
	}

	group.CloudNodeGroupID = newGroup.CloudNodeGroupID
	if group.AutoScaling != nil && group.AutoScaling.VpcID == "" {
		group.AutoScaling.VpcID = newGroup.AutoScaling.VpcID
	}
	if group.LaunchTemplate != nil {
		group.LaunchTemplate.InstanceChargeType = newGroup.LaunchTemplate.InstanceChargeType
	}

	if err = cloudprovider.GetStorageModel().UpdateNodeGroup(context.Background(), group); err != nil {
		return err
	}
	return nil
}

// checkNodeGroup 校验节点池
func checkNodeGroup(rootCtx context.Context, info *cloudprovider.CloudDependBasicInfo) error {
	group := info.NodeGroup
	cluster := info.Cluster
	var pool *armcontainerservice.AgentPool
	taskID := cloudprovider.GetTaskIDFromContext(rootCtx)
	// new Azure client
	client, err := api.NewAksServiceImplWithCommonOption(info.CmOption)
	if err != nil {
		return errors.Wrapf(err, "call NewAgentPoolClientWithOpt[%s] falied", taskID)
	}

	// wait node group state to normal
	ctx, cancel := context.WithTimeout(rootCtx, 30*time.Second)
	defer cancel()
	err = loop.LoopDoFunc(ctx, func() error {
		pool, err = client.GetPoolAndReturn(ctx, cloudprovider.GetClusterResourceGroup(info.Cluster),
			cluster.SystemID, group.CloudNodeGroupID)
		if err != nil {
			blog.Errorf("checkNodeGroup[%s] poll GetAgentPool[%s][%s] failed: %v", taskID, cluster.SystemID,
				group.CloudNodeGroupID, err)
			return nil
		}
		if pool == nil {
			return nil
		}
		switch {
		case *pool.Properties.ProvisioningState == api.NormalState:
			return loop.EndLoop
		case *pool.Properties.ProvisioningState == api.CreatingState:
			blog.Infof("checkNodeGroup[%s] poll GetAgentPool[%s] still creating, status[%s]", taskID,
				group.CloudNodeGroupID, *pool.Properties.ProvisioningState)
			return nil
		default:
			return nil
		}
	}, loop.LoopInterval(5*time.Second))
	if err != nil {
		return errors.Wrapf(err, "checkNodeGroup[%s] poll GetPoolAndReturn failed", taskID)
	}

	return cloudDataToNodeGroup(rootCtx, pool, client, info)
}

// cloudDataToNodeGroup 对齐
func cloudDataToNodeGroup(rootCtx context.Context, pool *armcontainerservice.AgentPool, client api.AksService,
	info *cloudprovider.CloudDependBasicInfo) error {
	var (
		group       = info.NodeGroup
		cluster     = info.Cluster
		taskID      = cloudprovider.GetTaskIDFromContext(rootCtx)
		ctx, cancel = context.WithTimeout(rootCtx, 30*time.Second)
	)
	defer cancel()
	if len(group.CloudNodeGroupID) == 0 {
		setCloudNodeGroupID(group)
	}
	if len(group.AutoScaling.MultiZoneSubnetPolicy) == 0 {
		group.AutoScaling.MultiZoneSubnetPolicy = "PRIORITY"
	}
	if len(group.AutoScaling.ScalingMode) == 0 {
		group.AutoScaling.ScalingMode = "CLASSIC_SCALING"
	}

	// 尝试获取nodeGroupResource
	nodeGroupResource, ok := info.Cluster.ExtraInfo[common.NodeResourceGroup]
	if !ok {
		cloudCluster, err := client.GetCluster(ctx, info, nodeGroupResource)
		if err != nil {
			return errors.Wrapf(err, "checkNodeGroup[%s]: call GetCluster falied", taskID)
		}
		nodeGroupResource = *cloudCluster.Properties.NodeResourceGroup
	}
	// 查询节点池与VMSSs映射
	ctx, cancel = context.WithTimeout(rootCtx, 30*time.Second)
	defer cancel()
	set, err := client.MatchNodeGroup(ctx, nodeGroupResource, info.NodeGroup.CloudNodeGroupID)
	if err != nil {
		return errors.Wrapf(err, "checkNodeGroup[%s]: call MatchNodeGroup[%s][%s] falied", taskID,
			cluster.ClusterID, group.CloudNodeGroupID)
	}
	// 字段对齐
	_ = client.SetToNodeGroup(set, group)
	_ = client.AgentPoolToNodeGroup(pool, group)
	syncSku(rootCtx, client, group)
	setModuleInfo(group, cluster.BusinessID)
	group.ClusterID = cluster.ClusterID
	// 镜像信息/dockerGraphPath节点运行时/runtime(运行时版本信息，RunTimeInfo)
	return nil
}

func syncSku(rootCtx context.Context, client api.AksService, group *proto.NodeGroup) {
	// wait node group state to normal
	ctx, cancel := context.WithTimeout(rootCtx, 30*time.Second)
	defer cancel()
	skus, err := client.ListResourceByLocation(ctx, group.Region)
	if err != nil {
		return
	}
	idx := -1
	for i, sku := range skus {
		if *sku.Name == group.LaunchTemplate.InstanceType {
			idx = i
			break
		}
	}
	if idx == -1 {
		return
	}
	for _, capability := range skus[idx].Capabilities {
		if *capability.Name == "vCPUs" {
			x, _ := strconv.ParseUint(*capability.Value, 10, 32)
			group.LaunchTemplate.CPU = uint32(x)
		} else if *capability.Name == "MemoryGB" {
			x, _ := strconv.ParseUint(*capability.Value, 10, 32)
			group.LaunchTemplate.Mem = uint32(x)
		} else if *capability.Name == "GPUs" {
			x, _ := strconv.ParseUint(*capability.Value, 10, 32)
			group.LaunchTemplate.GPU = uint32(x)
		}
	}
}
