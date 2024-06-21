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

// Package common xxx
package common

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

var (
	addNodesShieldAlarmStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.AddNodesShieldAlarmAction,
		StepName:   "屏蔽机器告警",
	}
)

// BuildShieldAlertTaskStep 屏蔽节点告警
func BuildShieldAlertTaskStep(task *proto.Task, clusterId string) {
	shieldStep := cloudprovider.InitTaskStep(addNodesShieldAlarmStep, cloudprovider.WithStepSkipFailed(true))
	shieldStep.Params[cloudprovider.ClusterIDKey.String()] = clusterId

	task.Steps[addNodesShieldAlarmStep.StepMethod] = shieldStep
	task.StepSequence = append(task.StepSequence, addNodesShieldAlarmStep.StepMethod)
}

// AddNodesShieldAlarmTask shield nodes alarm
func AddNodesShieldAlarmTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("AddNodesShieldAlarmTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("AddNodesShieldAlarmTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// extract valid info
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]

	ipList := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.GetCommonParams(),
		cloudprovider.NodeIPsKey.String(), ",")
	if len(ipList) == 0 {
		blog.Errorf("AddNodesShieldAlarmTask[%s]: get cluster IPList/clusterID empty", taskID)
		retErr := fmt.Errorf("AddNodesShieldAlarmTask: get cluster IPList/clusterID empty")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cluster, err := cloudprovider.GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		blog.Errorf("AddNodesShieldAlarmTask[%s]: get cluster for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("get cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = cloudprovider.ShieldHostAlarm(ctx, cluster.BusinessID, ipList)
	if err != nil {
		blog.Errorf("AddNodesShieldAlarmTask[%s] ShieldHostAlarmConfig failed: %v", taskID, err)
	} else {
		blog.Infof("AddNodesShieldAlarmTask[%s] ShieldHostAlarmConfig success", taskID)
	}

	// update step
	if errLocal := state.UpdateStepSucc(start, stepName); errLocal != nil {
		blog.Errorf("AddNodesShieldAlarmTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return errLocal
	}
	return nil
}
