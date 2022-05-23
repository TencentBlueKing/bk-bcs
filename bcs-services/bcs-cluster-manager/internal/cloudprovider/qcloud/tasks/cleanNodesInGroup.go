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

package tasks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
)

// CleanNodeGroupNodesTask clean node group nodes task
func CleanNodeGroupNodesTask(taskID string, stepName string) error {
	start := time.Now()
	//get task information and validate
	state, step, err := getStateAndStep(taskID, "CleanNodeGroupNodesTask", stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// step login started here
	cloudID := step.Params["CloudID"]
	nodeGroupID := step.Params["NodeGroupID"]
	nodesIDs := step.Params["NodesIDs"]
	group, err := cloudprovider.GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		blog.Errorf("CleanNodeGroupNodesTask[%s]: get nodegroup for %s failed", taskID, nodeGroupID)
		retErr := fmt.Errorf("get nodegroup information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get cloud and project info
	cloud, cluster, err := actions.GetCloudAndCluster(cloudprovider.GetStorageModel(), cloudID, group.ClusterID)
	if err != nil {
		blog.Errorf("CleanNodeGroupNodesTask[%s]: get cloud for nodegroup %s in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get dependency resource for cloudprovider operation
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("CleanNodeGroupNodesTask[%s]: get credential for nodegroup %s in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud credential err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cmOption.Region = group.Region

	// create node group
	asCli, err := api.NewASClient(cmOption)
	if err != nil {
		blog.Errorf("CleanNodeGroupNodesTask[%s]: get as client for nodegroup[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud as client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}
	if group.AutoScaling == nil || group.AutoScaling.AutoScalingID == "" {
		blog.Errorf("CleanNodeGroupNodesTask[%s]: nodegroup %s in task %s step %s has no autoscaling group",
			taskID, nodeGroupID, taskID, stepName)
		retErr := fmt.Errorf("get autoScalingID err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	err = asCli.RemoveInstances(group.AutoScaling.AutoScalingID, strings.Split(nodesIDs, ","))
	if err != nil {
		blog.Errorf("CleanNodeGroupNodesTask[%s]: call RemoveInstances[%s] api in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call RemoveInstances[%s] api err, %s", nodeGroupID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CleanNodeGroupNodesTask[%s]: call RemoveInstances successful", taskID)

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CleanNodeGroupNodesTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

// CheckCleanNodeGroupNodesStatusTask ckeck clean node group nodes status task
func CheckCleanNodeGroupNodesStatusTask(taskID string, stepName string) error {
	start := time.Now()
	//get task information and validate
	state, step, err := getStateAndStep(taskID, "CheckCleanNodeGroupNodesStatusTask", stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// step login started here
	nodeGroupID := step.Params["NodeGroupID"]
	cloudID := step.Params["CloudID"]

	group, err := cloudprovider.GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		blog.Errorf("CheckCleanNodeGroupNodesStatusTask[%s]: get nodegroup for %s failed", taskID, nodeGroupID)
		retErr := fmt.Errorf("get nodegroup information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cloud, cluster, err := actions.GetCloudAndCluster(cloudprovider.GetStorageModel(), cloudID, group.ClusterID)
	if err != nil {
		blog.Errorf("CheckCleanNodeGroupNodesStatusTask[%s]: get cloud/cluster for nodegroup %s in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get dependency resource for cloudprovider operation
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("CheckCleanNodeGroupNodesStatusTask[%s]: get credential for nodegroup %s in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud credential err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cmOption.Region = group.Region

	// get qcloud client
	cli, err := api.NewTkeClient(cmOption)
	if err != nil {
		blog.Errorf("CheckCleanNodeGroupNodesStatusTask[%s]: get tke client for nodegroup[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// wait node group state to normal
	ctx, cancel := context.WithTimeout(context.TODO(), 20*time.Minute)
	defer cancel()

	// wait all nodes to be ready
	err = cloudprovider.LoopDoFunc(ctx, func() error {
		np, err := cli.DescribeClusterNodePoolDetail(group.ClusterID, group.CloudNodeGroupID)
		if err != nil {
			blog.Errorf("taskID[%s] CheckCleanNodeGroupNodesStatusTask[%s/%s] failed: %v", taskID, group.ClusterID,
				group.CloudNodeGroupID, err)
			return nil
		}
		if np == nil || np.NodeCountSummary == nil {
			return nil
		}
		if np.NodeCountSummary.ManuallyAdded == nil || np.NodeCountSummary.AutoscalingAdded == nil {
			return nil
		}
		allNormalNodesCount := *np.NodeCountSummary.ManuallyAdded.Normal + *np.NodeCountSummary.AutoscalingAdded.Normal
		switch {
		case *np.DesiredNodesNum == allNormalNodesCount:
			return cloudprovider.EndLoop
		default:
			return nil
		}
	}, cloudprovider.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("taskID[%s] DescribeClusterNodePoolDetail failed: %v", taskID, err)
		return err
	}
	return nil
}

// UpdateCleanNodeGroupNodesDBInfoTask update clean node group nodes db info task
func UpdateCleanNodeGroupNodesDBInfoTask(taskID string, stepName string) error {
	start := time.Now()
	//get task information and validate
	state, step, err := getStateAndStep(taskID, "UpdateCleanNodeGroupNodesDBInfoTask", stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// step login started here
	nodeGroupID := step.Params["NodeGroupID"]
	cloudID := step.Params["CloudID"]

	group, err := cloudprovider.GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		blog.Errorf("UpdateCleanNodeGroupNodesDBInfoTask[%s]: get nodegroup for %s failed", taskID, nodeGroupID)
		retErr := fmt.Errorf("get nodegroup information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cloud, cluster, err := actions.GetCloudAndCluster(cloudprovider.GetStorageModel(), cloudID, group.ClusterID)
	if err != nil {
		blog.Errorf("UpdateCleanNodeGroupNodesDBInfoTask[%s]: get cloud/cluster for nodegroup %s in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get dependency resource for cloudprovider operation
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("UpdateCleanNodeGroupNodesDBInfoTask[%s]: get credential for nodegroup %s in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud credential err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cmOption.Region = group.Region

	// get qcloud client
	cli, err := api.NewTkeClient(cmOption)
	if err != nil {
		blog.Errorf("UpdateCleanNodeGroupNodesDBInfoTask[%s]: get tke client for nodegroup[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	np, err := cli.DescribeClusterNodePoolDetail(group.ClusterID, group.CloudNodeGroupID)
	if err != nil {
		blog.Errorf("taskID[%s] DescribeClusterNodePoolDetail[%s/%s] failed: %v", taskID, group.ClusterID,
			group.CloudNodeGroupID, err)
		retErr := fmt.Errorf("DescribeClusterNodePoolDetail err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return nil
	}

	// TODO update nodes info
	err = updateNodeGroupDesiredSize(nodeGroupID, uint32(*np.DesiredNodesNum))
	if err != nil {
		blog.Errorf("taskID[%s] updateNodeGroupDesiredSize[%s/%d] failed: %v", taskID, nodeGroupID,
			*np.DesiredNodesNum, err)
		retErr := fmt.Errorf("updateNodeGroupDesiredSize err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return nil
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateCleanNodeGroupNodesDBInfoTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}
