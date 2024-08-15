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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
)

// CheckCleanDBDataTask delete node and nodeGroup data in db
func CheckCleanDBDataTask(taskID string, stepName string) error {
	// step1: delete nodes
	// step2: delete consumer
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckCleanDBDataTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckCleanDBDataTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	group, err := cloudprovider.GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		blog.Errorf("CheckCleanDBDataTask[%s]: get NodeGroup %s to clean Node in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		_ = state.UpdateStepFailure(start, stepName, err)
		return fmt.Errorf("get NodeGroup failed %s", err.Error())
	}

	// step1: delete node db list
	ipList := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params, cloudprovider.NodeIPsKey.String(), ",")
	err = cloudprovider.GetStorageModel().DeleteNodesByIPs(context.Background(), ipList)
	if err != nil {
		blog.Errorf("CheckCleanDBDataTask[%s]: DeleteNodesByIPs failed: %v", taskID, err)
		_ = state.UpdateStepFailure(start, stepName, err)
		return fmt.Errorf("DeleteNodesByIPs failed %s", err.Error())
	}

	// step2: delete resourcePool consumer
	err = utils.DeleteResourcePoolAction(context.Background(), group.ConsumerID)
	if err != nil {
		blog.Errorf("CheckCleanDBDataTask[%s] DeleteResourcePool failed: %v", taskID, err)
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckCleanDBDataTask[%s]: task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}
