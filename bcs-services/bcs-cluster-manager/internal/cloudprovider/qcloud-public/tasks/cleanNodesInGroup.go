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
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/avast/retry-go"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud-public/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// CleanNodeGroupNodesTask clean node group nodes task
func CleanNodeGroupNodesTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start clean nodegroup nodes")
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
	nodeIDs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.NodeIDsKey.String(), ",")

	if len(clusterID) == 0 || len(nodeGroupID) == 0 || len(cloudID) == 0 || len(nodeIDs) == 0 {
		blog.Errorf("CleanNodeGroupNodesTask[%s]: check parameter validate failed", taskID)
		retErr := fmt.Errorf("CleanNodeGroupNodesTask check parameters failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
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
	ctx := cloudprovider.WithTaskIDAndStepNameForContext(context.Background(), taskID, stepName)

	// 按量计费节点池 销毁节点; 包年包月节点池 移除节点,需要用户手动回收
	switch dependInfo.NodeGroup.GetLaunchTemplate().GetInstanceChargeType() {
	case icommon.PREPAID:
		deleteResult, errLocal := business.RemoveNodesFromCluster(ctx, dependInfo, cloudprovider.Terminate.String(), nodeIDs)
		if errLocal != nil {
			cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
				fmt.Sprintf("remove nodes from cluster failed [%s]", errLocal))
			blog.Errorf("CleanNodeGroupNodesTask[%s] RemoveNodesFromCluster failed: %v",
				taskID, errLocal)
			retErr := fmt.Errorf("RemoveNodesFromCluster err, %s", errLocal.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
		blog.Infof("CleanNodeGroupNodesTask[%s] deletedInstance[%v]", taskID, deleteResult)
	default:
		err = removeAsgInstances(ctx, dependInfo, nodeIDs)
		if err != nil {
			cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
				fmt.Sprintf("remove asg instances failed [%s]", err))
			blog.Errorf("CleanNodeGroupNodesTask[%s] nodegroup %s removeAsgInstances failed: %v",
				taskID, nodeGroupID, err)
			retErr := fmt.Errorf("removeAsgInstances err, %v", err)
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"clean nodegroup nodes successful")

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
		activityID, err := asCli.RemoveInstances(asgID, validateInstances) // nolint
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

// CheckClusterCleanNodsTask check cluster clean nodes task
func CheckClusterCleanNodsTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start check cluster clean nodes")
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
	nodeIDs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.NodeIDsKey.String(), ",")

	if len(clusterID) == 0 || len(cloudID) == 0 || len(nodeIDs) == 0 {
		blog.Errorf("CheckClusterCleanNodsTask[%s]: check parameter validate failed", taskID)
		retErr := fmt.Errorf("CheckClusterCleanNodsTask check parameters failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("CheckClusterCleanNodsTask[%s]: GetClusterDependBasicInfo failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("CheckClusterCleanNodsTask GetClusterDependBasicInfo failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// wait check delete component status
	timeContext, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	err = loop.LoopDoFunc(timeContext, func() error {
		exist, notExist, err := business.FilterClusterInstanceFromNodesIDs(timeContext, dependInfo, nodeIDs) // nolint
		if err != nil {
			blog.Errorf("CheckClusterCleanNodsTask[%s] FilterClusterInstanceFromNodesIDs failed: %v",
				taskID, err)
			return nil
		}

		blog.Infof("CheckClusterCleanNodsTask[%s] nodeIDs[%v] exist[%v] notExist[%v]",
			taskID, nodeIDs, exist, notExist)

		cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
			fmt.Sprintf("nodeIDs [%v] exist [%v] notExist [%v]", nodeIDs, exist, notExist))

		if len(exist) == 0 {
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(30*time.Second))

	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		blog.Errorf("CheckClusterCleanNodsTask[%s] cluster[%s] failed: %v", taskID, clusterID, err)
	}

	// timeout error
	if errors.Is(err, context.DeadlineExceeded) {
		blog.Infof("CheckClusterCleanNodsTask[%s] cluster[%s] timeout failed: %v", taskID, clusterID, err)
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"check cluster clean nodes successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckClusterCleanNodsTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

// CheckCleanNodeGroupNodesStatusTask check clean node group nodes status task
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
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("CheckCleanNodeGroupNodesStatusTask[%s]: GetClusterDependBasicInfo for nodegroup %s in "+
			"task %s step %s failed, %s", // nolint
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get qcloud client
	cli, err := api.NewTkeClient(dependInfo.CmOption)
	if err != nil {
		blog.Errorf("CheckCleanNodeGroupNodesStatusTask[%s]: get tke client for nodegroup[%s] in "+
			"task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// wait node group state to normal
	ctx, cancel := context.WithTimeout(context.TODO(), 20*time.Minute)
	defer cancel()

	// wait all nodes to be ready
	err = loop.LoopDoFunc(ctx, func() error {
		np, errPool := cli.DescribeClusterNodePoolDetail(dependInfo.Cluster.SystemID,
			dependInfo.NodeGroup.CloudNodeGroupID)
		if errPool != nil {
			blog.Errorf("taskID[%s] CheckCleanNodeGroupNodesStatusTask[%s/%s] failed: %v", taskID,
				dependInfo.NodeGroup.ClusterID,
				dependInfo.NodeGroup.CloudNodeGroupID, errPool)
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
			return loop.EndLoop
		default:
			return nil
		}
	}, loop.LoopInterval(10*time.Second))
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
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("CheckCleanNodeGroupNodesStatusTask[%s]: GetClusterDependBasicInfo for nodegroup %s in "+
			"task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get qcloud client
	cli, err := api.NewTkeClient(dependInfo.CmOption)
	if err != nil {
		blog.Errorf("UpdateCleanNodeGroupNodesDBInfoTask[%s]: get tke client for nodegroup[%s] in "+
			"task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	np, err := cli.DescribeClusterNodePoolDetail(dependInfo.Cluster.SystemID, dependInfo.NodeGroup.CloudNodeGroupID)
	if err != nil {
		blog.Errorf("taskID[%s] DescribeClusterNodePoolDetail[%s/%s] failed: %v", taskID,
			dependInfo.NodeGroup.ClusterID,
			dependInfo.NodeGroup.CloudNodeGroupID, err)
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
		blog.Errorf("UpdateCleanNodeGroupNodesDBInfoTask[%s] task %s %s update to storage fatal", taskID,
			taskID, stepName)
		return err
	}

	return nil
}
