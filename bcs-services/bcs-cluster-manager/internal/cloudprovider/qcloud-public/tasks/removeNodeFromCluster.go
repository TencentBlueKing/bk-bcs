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
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud-public/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// RemoveNodesFromClusterTask remove node from cluster
func RemoveNodesFromClusterTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("RemoveNodesFromClusterTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("RemoveNodesFromClusterTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// get data info
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	// get nodes IDs and IPs
	ipList := strings.Split(step.Params[cloudprovider.NodeIPsKey.String()], ",")
	idList := strings.Split(step.Params[cloudprovider.NodeIDsKey.String()], ",")
	if len(idList) != len(ipList) {
		blog.Errorf("RemoveNodesFromClusterTask[%s] [inner fatal] task %s step %s NodeID %d is not equal to "+
			"InnerIP %d, fatal", taskID, taskID, stepName,
			len(idList), len(ipList))
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("NodeID & InnerIP params err"))
		return fmt.Errorf("task %s parameter err", taskID)
	}
	// delete node mode:
	mode := step.Params[cloudprovider.DeleteModeKey.String()]
	if mode == "" || !utils.StringInSlice(mode, []string{api.Terminate.String(), api.Retain.String()}) {
		mode = api.Retain.String()
	}

	// step login started here
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("RemoveNodesFromClusterTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	deleteResult, err := business.RemoveNodesFromCluster(ctx, dependInfo, mode, idList)
	if err != nil {
		blog.Errorf("RemoveNodesFromClusterTask[%s] RemoveNodesFromCluster failed: %v",
			taskID, err)
		retErr := fmt.Errorf("RemoveNodesFromCluster err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("RemoveNodesFromClusterTask[%s] deletedInstance[%v]", taskID, deleteResult)

	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	state.Task.CommonParams[cloudprovider.SuccessClusterNodeIDsKey.String()] = strings.Join(deleteResult, ",")

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("RemoveNodesFromClusterTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// UpdateRemoveNodeDBInfoTask update remove node DB info
func UpdateRemoveNodeDBInfoTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("UpdateRemoveNodeDBInfoTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("UpdateRemoveNodeDBInfoTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// extract valid info
	success := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.SuccessClusterNodeIDsKey.String(), ",")
	bizIdStr := step.Params[cloudprovider.BKBizIDKey.String()]

	terminateNodes := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params,
		cloudprovider.TerminateChargeNodes.String(), ",")

	retainNodes := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params,
		cloudprovider.RetainChargeNodes.String(), ",")

	if len(success) > 0 {
		for i := range success {
			err = cloudprovider.GetStorageModel().DeleteNode(context.Background(), success[i])
			if err != nil {
				blog.Errorf("UpdateRemoveNodeDBInfoTask[%s] task %s deleteNodeByNodeID failed: %v", taskID, taskID, err)
			}
		}
	}

	// trans host module by cloud nodes chargeType
	biz, _ := strconv.Atoi(bizIdStr)
	if len(terminateNodes) > 0 {
		err = common.RemoveHostFromCmdb(context.Background(), biz, strings.Join(terminateNodes, ","))
		if err != nil {
			blog.Errorf("UpdateRemoveNodeDBInfoTask[%s] RemoveHostFromCmdb failed: %v", taskID, err)
		}
	}
	if len(retainNodes) > 0 {
		err = common.TransBizNodeModule(context.Background(), biz, 0, retainNodes)
		if err != nil {
			blog.Errorf("UpdateRemoveNodeDBInfoTask[%s] RemoveHostFromCmdb failed: %v", taskID, err)
		}
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateNodeDBInfoTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}
