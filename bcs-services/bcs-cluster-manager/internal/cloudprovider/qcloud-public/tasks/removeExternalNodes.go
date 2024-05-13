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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud-public/business"
)

// ReturnIDCNodeToResourcePoolTask clean IDCNodes in group task for background running
func ReturnIDCNodeToResourcePoolTask(taskID, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("ReturnIDCNodeToResourcePoolTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("ReturnIDCNodeToResourcePoolTask[%s] task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// extract valid parameter
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	operator := step.Params[cloudprovider.OperatorKey.String()]
	nodeIPList := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params, cloudprovider.NodeIPsKey.String(), ",")
	deviceList := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params, cloudprovider.DeviceIDsKey.String(), ",")

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("ReturnIDCNodeToResourcePoolTask[%s] GetClusterDependBasicInfo for NodeGroup %s to "+
			"clean Node in task %s step %s failed, %s", taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// return IDC device to resource-manager
	orderID, err := destroyIDCDeviceList(ctx, dependInfo, deviceList, operator)
	if err != nil {
		blog.Errorf("ReturnIDCNodeToResourcePoolTask[%s] destroyIDCDeviceList[%v] from NodeGroup %s failed: %v",
			taskID, nodeIPList, nodeGroupID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, err)
		return fmt.Errorf("ReturnIDCNodeToResourcePoolTask destroyIDCDeviceList failed %s", err.Error())
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	state.Task.CommonParams[cloudprovider.OrderIDKey.String()] = orderID

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("task %s %s update to storage fatal", taskID, stepName)
		return err
	}
	return nil
}

// RemoveExternalNodesFromClusterTask remove external node from cluster
func RemoveExternalNodesFromClusterTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("RemoveExternalNodesFromClusterTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("RemoveExternalNodesFromClusterTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// get data info
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	groupID := step.Params[cloudprovider.NodeGroupIDKey.String()]

	ipList := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params, cloudprovider.NodeIPsKey.String(), ",")

	// step login started here
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: groupID,
	})
	if err != nil {
		blog.Errorf("RemoveExternalNodesFromClusterTask[%s]: GetClusterDependBasicInfo for cluster %s in "+
			"task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = business.RemoveExternalNodesFromCluster(ctx, dependInfo, ipList)
	if err != nil {
		blog.Errorf("RemoveExternalNodesFromClusterTask[%s] RemoveExternalNodesFromCluster failed: %v",
			taskID, err)
		retErr := fmt.Errorf("RemoveExternalNodesFromCluster err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("RemoveExternalNodesFromClusterTask[%s] removeNodes[%v]", taskID, ipList)

	// get add external nodes script from cluster
	script, err := business.GetClusterExternalNodeScript(ctx, dependInfo, false)
	if err != nil {
		blog.Errorf("RemoveExternalNodesFromClusterTask[%s]: GetClusterExternalNodeScript for cluster[%s] failed, %s",
			taskID, clusterID, err.Error())
		retErr := fmt.Errorf("GetClusterExternalNodeScript err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	state.Task.CommonParams[cloudprovider.DynamicNodeScriptKey.String()] = script

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("RemoveExternalNodesFromClusterTask[%s] %s update to storage fatal", taskID, stepName)
		return err
	}

	return nil
}
