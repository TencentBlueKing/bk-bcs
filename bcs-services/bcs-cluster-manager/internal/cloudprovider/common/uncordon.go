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

package common

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
)

var (
	// UnCordonNodesActionStep 可调度任务
	UnCordonNodesActionStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.UnCordonNodesAction,
		StepName:   "节点设置可调度状态",
	}
	// CordonNodesActionStep 不可调度任务
	CordonNodesActionStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.CordonNodesAction,
		StepName:   "节点设置不可调度状态",
	}
)

// BuildUnCordonNodesTaskStep build uncordon task step
func BuildUnCordonNodesTaskStep(task *proto.Task, clusterID string, nodeIPs []string) {
	unCordonStep := cloudprovider.InitTaskStep(UnCordonNodesActionStep)

	unCordonStep.Params[cloudprovider.ClusterIDKey.String()] = clusterID
	unCordonStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")

	task.Steps[UnCordonNodesActionStep.StepMethod] = unCordonStep
	task.StepSequence = append(task.StepSequence, UnCordonNodesActionStep.StepMethod)
}

// BuildCordonNodesTaskStep build cordon task step
func BuildCordonNodesTaskStep(task *proto.Task, clusterID string, nodeIPs []string) {
	cordonStep := cloudprovider.InitTaskStep(CordonNodesActionStep)

	cordonStep.Params[cloudprovider.ClusterIDKey.String()] = clusterID
	cordonStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")

	task.Steps[CordonNodesActionStep.StepMethod] = cordonStep
	task.StepSequence = append(task.StepSequence, CordonNodesActionStep.StepMethod)
}

type scheduleNodesData struct {
	clusterID string
	nodeIPs   []string
	cordon    bool
}

func updateNodesScheduleStatus(ctx context.Context, data scheduleNodesData) error { // nolint
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())
	// trans nodeIPs to nodeNames: k8s cluster register nodeName not nodeIP
	nodeNames := make([]string, 0)
	nodes, err := k8sOperator.ListClusterNodesByIPsOrNames(ctx, clusterops.ListNodeOption{
		ClusterID: data.clusterID,
		NodeIPs:   data.nodeIPs,
	})
	if err != nil {
		blog.Errorf("updateNodesScheduleStatus[%s] ListClusterNodesByIPsOrNames failed: %v", taskID, err)
		nodeNames = data.nodeIPs
	} else {
		for i := range nodes {
			nodeNames = append(nodeNames, nodes[i].Name)
		}
		blog.Infof("updateNodesScheduleStatus[%s] ListClusterNodesByIPsOrNames successful[%v]", taskID, nodeNames)
	}

	blog.Infof("updateNodesScheduleStatus[%s] nodeNames[%v]", taskID, nodeNames)
	for _, name := range nodeNames {
		err := k8sOperator.ClusterUpdateScheduleNode(context.Background(), clusterops.NodeInfo{
			ClusterID: data.clusterID,
			NodeName:  name,
			Desired:   data.cordon,
		})
		if err != nil {
			blog.Errorf("updateNodesScheduleStatus[%s] ip[%s] failed: %v", taskID, name, err)
			continue
		}

		blog.Infof("updateNodesScheduleStatus[%s] ip[%s] successful", taskID, name)
	}

	return nil
}

// UnCordonNodesTask unCordon cluster nodes
func UnCordonNodesTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start uncordon nodes")
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("UnCordonNodesTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("UnCordonNodesTask[%s]: run step %s, system: %s, old state: %s, params %v",
		taskID, stepName, step.System, step.Status, step.Params)

	// extract parameter
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeIPs := strings.Split(state.Task.CommonParams[cloudprovider.NodeIPsKey.String()], ",")

	if len(clusterID) == 0 || len(nodeIPs) == 0 {
		errMsg := fmt.Sprintf("UnCordonNodesTask[%s] validateParameter failed: clusterID or nodeIPs empty", taskID)
		blog.Errorf(errMsg)
		retErr := fmt.Errorf("UnCordonNodesTask err: %s", errMsg)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	_ = updateNodesScheduleStatus(ctx, scheduleNodesData{
		clusterID: clusterID,
		nodeIPs:   nodeIPs,
		cordon:    false,
	})

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"uncordon nodes successful")

	blog.Infof("UnCordonNodesTask[%s] clusterID[%s] IPs[%v] successful", taskID, clusterID, nodeIPs)
	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("task %s %s update to storage fatal", taskID, stepName)
		return err
	}

	return nil
}

// CordonNodesTask cordon cluster nodes
func CordonNodesTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start cordon cluster nodes")
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CordonNodesTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CordonNodesTask[%s]: run step %s, system: %s, old state: %s, params %v",
		taskID, stepName, step.System, step.Status, step.Params)

	// extract parameter
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeIPs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.NodeIPsKey.String(), ",")

	if len(clusterID) == 0 || len(nodeIPs) == 0 {
		errMsg := fmt.Sprintf("CordonNodesTask[%s] validateParameter failed: clusterID or nodeIPs empty", taskID)
		blog.Errorf(errMsg)
		retErr := fmt.Errorf("CordonNodesTask err: %s", errMsg)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	_ = updateNodesScheduleStatus(ctx, scheduleNodesData{
		clusterID: clusterID,
		nodeIPs:   nodeIPs,
		cordon:    true,
	})

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"cordon cluster nodes successful")

	blog.Infof("CordonNodesTask[%s] clusterID[%s] IPs[%v] successful", taskID, clusterID, nodeIPs)
	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("task %s %s update to storage fatal", taskID, stepName)
		return err
	}

	return nil
}
