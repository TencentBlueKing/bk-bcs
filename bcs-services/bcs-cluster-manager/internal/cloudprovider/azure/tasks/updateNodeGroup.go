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
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v3"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
)

// errors
var (
	nodePoolScaleUpErr = errors.New("the status of the aks node pool is scale up, and currently " + // nolint
		"no operations can be performed on it")
	nodePoolUpdatingErr = errors.New("the aks node pool status is in the process of being updating " + // nolint
		"and no operations can be performed on it right now")
)

// UpdateAKSNodeGroupTask call azure interface to update cluster node group
func UpdateAKSNodeGroupTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("UpdateAKSNodeGroupTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("UpdateAKSNodeGroupTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// get step paras
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]

	// get dependent basic info
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("UpdateAKSNodeGroupTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err) // nolint
		retErr := fmt.Errorf("get cloud/project information failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cluster := dependInfo.Cluster
	group := dependInfo.NodeGroup

	// create aks client
	client, err := api.NewAksServiceImplWithCommonOption(dependInfo.CmOption)
	if err != nil {
		blog.Errorf("UpdateAKSNodeGroupTask[%s]: updateAgentPool[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err)
		retErr := fmt.Errorf("create aks client failed, err: %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	if group.NodeGroupID == "" || group.ClusterID == "" {
		blog.Errorf("nodegroup id or cluster id is empty")
		retErr := fmt.Errorf("nodegroup id or cluster id is empty")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update agent pool
	if err = updateAgentPoolProperties(client, cluster, group); err != nil {
		blog.Errorf("UpdateAKSNodeGroupTask[%s]: updateAgentPool[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err)
		retErr := fmt.Errorf("call updateAgentPoolProperties failed, err: %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update virtual machine scale set
	if err = updateVMSSProperties(client, group); err != nil {
		blog.Errorf("UpdateAKSNodeGroupTask[%s]: updateAgentPool[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err)
		retErr := fmt.Errorf("call updateVMSSProperties failed, err: %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateAKSClusterTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

func updateAgentPoolProperties(client api.AksService, cluster *proto.Cluster,
	group *proto.NodeGroup) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := client.GetPoolAndReturn(ctx, cloudprovider.GetClusterResourceGroup(cluster),
		cluster.SystemID, group.CloudNodeGroupID)
	if err != nil {
		return errors.Wrapf(err, "UpdateNodeGroup: call GetAgentPool api failed")
	}

	if err = checkPoolState(pool); err != nil { // 更新前检查节点池的状态
		return errors.Wrapf(err, "nodeGroupID: %s unable to update agent pool", group.NodeGroupID)
	}

	// 更新 pool
	api.SetAgentPoolFromNodeGroup(group, pool)

	// update agent pool
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if _, err = client.UpdatePoolAndReturn(ctx, pool, cloudprovider.GetClusterResourceGroup(cluster),
		cluster.SystemID, *pool.Name); err != nil {
		return errors.Wrapf(err, "UpdateNodeGroup: call UpdateAgentPool api failed")
	}

	return nil
}

// checkPoolState 更新前，检查节点池的状态
// 如果节点池正在 "更新中" 或 "扩容中"，将无法对其进行操作
func checkPoolState(pool *armcontainerservice.AgentPool) error {
	state := *pool.Properties.ProvisioningState
	if state == api.UpdatingState {
		return errors.Wrapf(nodePoolUpdatingErr, "cloudNodeGroupID: %s", *pool.Name)
	}
	if state == api.ScalingState {
		return errors.Wrapf(nodePoolScaleUpErr, "cloudNodeGroupID: %s", *pool.Name)
	}
	return nil
}

func updateVMSSProperties(client api.AksService, group *proto.NodeGroup) error {
	if group.LaunchTemplate == nil || len(group.LaunchTemplate.UserData) == 0 {
		return nil
	}

	asg := group.AutoScaling
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	set, err := client.GetSetWithName(ctx, asg.AutoScalingName, asg.AutoScalingID)
	if err != nil {
		return errors.Wrapf(err, "UpdateNodeGroup: call GetSetWithName api failed")
	}

	if group.LaunchTemplate != nil && len(group.LaunchTemplate.UserData) != 0 {
		set.Properties.VirtualMachineProfile.UserData = to.Ptr(group.LaunchTemplate.UserData)
	}
	// 镜像引用-暂时置空处理，若不置空会导致无法更新set
	api.SetImageReferenceNull(set)

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if _, err = client.UpdateSetWithName(ctx, set, asg.AutoScalingName, asg.AutoScalingID); err != nil {
		return errors.Wrapf(err, "UpdateNodeGroup: call UpdateSetWithName api failed")
	}

	return nil
}
