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
	"github.com/avast/retry-go"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// CleanNodeGroupNodesTask clean node group nodes task
func CleanNodeGroupNodesTask(taskID string, stepName string) error {
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

	// extract parameter && check validate
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeIDs := strings.Split(state.Task.CommonParams[cloudprovider.NodeIDsKey.String()], ",")

	if len(clusterID) == 0 || len(nodeGroupID) == 0 || len(cloudID) == 0 || len(nodeIDs) == 0 {
		blog.Errorf("CleanNodeGroupNodesTask[%s]: check parameter validate failed", taskID)
		retErr := fmt.Errorf("CleanNodeGroupNodesTask check parameters failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(clusterID, cloudID, nodeGroupID)
	if err != nil {
		blog.Errorf("CleanNodeGroupNodesTask[%s]: GetClusterDependBasicInfo failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("CleanNodeGroupNodesTask GetClusterDependBasicInfo failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	if dependInfo.NodeGroup.AutoScaling == nil || dependInfo.NodeGroup.AutoScaling.AutoScalingID == "" {
		blog.Errorf("CleanNodeGroupNodesTask[%s]: nodegroup %s in task %s step %s has no autoscaling group",
			taskID, nodeGroupID, taskID, stepName)
		retErr := fmt.Errorf("get autoScalingID err, %v", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = removeAsgInstances(ctx, dependInfo, nodeIDs)
	if err != nil {
		blog.Errorf("CleanNodeGroupNodesTask[%s] nodegroup %s removeAsgInstances failed: %v",
			taskID, nodeGroupID, err)
		retErr := fmt.Errorf("removeAsgInstances err, %v", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CleanNodeGroupNodesTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

func removeAsgInstances(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, nodeIDs []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	asgID, err := getAsgIDByNodePool(ctx, info)
	if err != nil {
		return fmt.Errorf("removeAsgInstances[%s] getAsgIDByNodePool failed: %v", taskID, err)
	}

	// create node group
	asCli, err := api.NewASClient(info.CmOption)
	if err != nil {
		blog.Errorf("removeAsgInstances[%s] get as client failed: %v", taskID, err.Error())
		return err
	}

	// check instances if exist
	var (
		instanceIDList, validateInstances = make([]string, 0), make([]string, 0)
	)
	asgInstances, err := asCli.DescribeAutoScalingInstances(asgID)
	if err != nil {
		blog.Errorf("removeAsgInstances[%s] DescribeAutoScalingInstances[%s] failed: %v", taskID, asgID, err.Error())
		return err
	}
	for _, ins := range asgInstances {
		instanceIDList = append(instanceIDList, *ins.InstanceID)
	}
	for _, id := range nodeIDs {
		if utils.StringInSlice(id, instanceIDList) {
			validateInstances = append(validateInstances, id)
		}
	}
	if len(validateInstances) == 0 {
		blog.Infof("removeAsgInstances[%s] validateInstances is empty", taskID)
		return nil
	}

	blog.Infof("removeAsgInstances[%s] validateInstances[%v]", taskID, validateInstances)
	err = retry.Do(func() error {
		activityID, err := asCli.RemoveInstances(asgID, validateInstances)
		if err != nil {
			blog.Errorf("removeAsgInstances[%s] RemoveInstances failed: %v", taskID, err)
			return err
		}

		blog.Infof("removeAsgInstances[%s] RemoveInstances[%v] successful[%s]", taskID, nodeIDs, activityID)
		return nil
	}, retry.Attempts(3))

	if err != nil {
		return err
	}

	return nil
}

// CheckCleanNodeGroupNodesStatusTask ckeck clean node group nodes status task
func CheckCleanNodeGroupNodesStatusTask(taskID string, stepName string) error {
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
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
		blog.Errorf(
			"CheckCleanNodeGroupNodesStatusTask[%s]: get cloud/cluster for nodegroup %s in task %s step %s failed, %s",
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
		np, errPool := cli.DescribeClusterNodePoolDetail(cluster.SystemID, group.CloudNodeGroupID)
		if errPool != nil {
			blog.Errorf("taskID[%s] CheckCleanNodeGroupNodesStatusTask[%s/%s] failed: %v", taskID, group.ClusterID,
				group.CloudNodeGroupID, errPool)
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
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
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
		blog.Errorf(
			"UpdateCleanNodeGroupNodesDBInfoTask[%s]: get cloud/cluster for nodegroup %s in task %s step %s failed, %s",
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

	np, err := cli.DescribeClusterNodePoolDetail(cluster.SystemID, group.CloudNodeGroupID)
	if err != nil {
		blog.Errorf("taskID[%s] DescribeClusterNodePoolDetail[%s/%s] failed: %v", taskID, group.ClusterID,
			group.CloudNodeGroupID, err)
		retErr := fmt.Errorf("DescribeClusterNodePoolDetail err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return nil
	}

	// will do update nodes info
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
