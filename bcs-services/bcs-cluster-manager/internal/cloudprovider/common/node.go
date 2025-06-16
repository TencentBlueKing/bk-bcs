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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// NodeDrainPodTask running node drain pod job and wait for results
func NodeDrainPodTask(taskID string, stepName string) error { // nolint
	// step1: get para by taskID
	// step2: for range node use goroutine run k8s command to drain pod
	// step3: record success and failed result

	message := fmt.Sprintf("start run node drain pod task, [TaskID:%s][Step:%s] ", taskID, stepName)
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName, message)

	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("NodeDrainPodTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("NodeDrainPodTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// get node drain pod common parameter
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeNames := step.Params[cloudprovider.NodeNamesKey.String()]
	drainer := step.Params[cloudprovider.DrainHelperKey.String()]

	taskName := state.Task.CommonParams[cloudprovider.TaskNameKey.String()]

	message = fmt.Sprintf(
		"run node drain pod task parameters: [TaskID:%s][Step:%s]\n"+
			"Parameters:\n"+
			"- ClusterID: %s\n"+
			"- NodeNames: %s\n"+
			"- Drainer: %s\n"+
			"- TaskName: %s",
		taskID,
		stepName,
		clusterID,
		nodeNames,
		drainer,
		taskName,
	)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName, message)

	if nodeNames == "" || drainer == "" || clusterID == "" || taskName == "" {
		errMsg := fmt.Sprintf("NodeDrainPodTask[%s] validateParameter task[%s] step[%s] failed", taskID, taskID, stepName)
		blog.Errorf(errMsg)
		retErr := fmt.Errorf("NodeDrainPodTask err, %s", errMsg)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	drainerHelper := clusterops.DrainHelper{}
	err = json.Unmarshal([]byte(drainer), &drainerHelper)
	if err != nil {
		errMsg := fmt.Sprintf("NodeDrainPodTask[%s] unmarshal drainer failed[%v]", taskID, err)
		blog.Errorf(errMsg)
		retErr := fmt.Errorf("NodeDrainPodTask err, %s", errMsg)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	if drainerHelper.Timeout == 0 {
		drainerHelper.Timeout = 60 * 10
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"run node drain pod task start ExecNodeDrainPodTask")

	// inject taskID and StepName
	ctx := cloudprovider.WithTaskIDAndStepNameForContext(context.Background(), taskID, stepName)

	timeOutCtx, cancel := context.WithTimeout(ctx, time.Minute*60)
	defer cancel()

	// exec node drain pod task
	ret := ExecNodeDrainPodTask(timeOutCtx, CreateNodeDrainPodTaskParas{
		DrainHelper: drainerHelper,
		NodeNames:   strings.Split(nodeNames, ","),
		ClusterID:   clusterID,
	})

	var successNodes, failedNodes, partFailedNodes []string
	for _, r := range ret {
		switch r.Result {
		case NodeDrainPodResultSuccess:
			message = fmt.Sprintf("node[%s] drain pod in cluster[%s] success", r.NodeName, clusterID)
			cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName, message)
			successNodes = append(successNodes, r.NodeName)
		case NodeDrainPodResultFailed:
			message = fmt.Sprintf("node[%s] drain pod in cluster[%s] failed, error: [%s]",
				r.NodeName, clusterID, r.Err.Error())
			cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName, message)
			failedNodes = append(failedNodes, r.NodeName)
		case NodeDrainPodResultPartFailed:
			message = fmt.Sprintf("node[%s] drain pod in cluster[%s] part failed, info: [%s]",
				r.NodeName, clusterID, r.Err.Error())
			cloudprovider.GetStorageModel().CreateTaskStepLogWarn(context.Background(), taskID, stepName, message)
			partFailedNodes = append(partFailedNodes, r.NodeName)
		}
	}

	message = fmt.Sprintf("all node drain pod task finished, [TaskID:%s][Step:%s]",
		taskID, stepName)
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName, message)

	successNodesStr := strings.Join(successNodes, ",")
	failedNodesStr := strings.Join(failedNodes, ",")
	partFailedNodesStr := strings.Join(partFailedNodes, ",")

	message = fmt.Sprintf("successNodes: [%s]; partFailedNodes: [%s]; failedNodes: [%s]",
		successNodesStr, partFailedNodesStr, failedNodesStr)
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName, message)

	if len(failedNodes) > 0 || len(partFailedNodes) > 0 {
		retErr := fmt.Errorf("failedNodes: [%s], partFailedNodes: [%s]",
			failedNodesStr, partFailedNodesStr)

		if len(successNodes) > 0 || len(partFailedNodes) > 0 {
			_ = state.UpdateStepPartFailure(start, stepName, retErr)
		} else {
			_ = state.UpdateStepFailure(start, stepName, retErr)
		}
		return retErr
	}

	_ = state.UpdateStepSucc(start, stepName)
	return nil
}

// CreateNodeDrainPodTaskParas create node drain task paras
type CreateNodeDrainPodTaskParas struct {
	DrainHelper clusterops.DrainHelper
	NodeNames   []string
	ClusterID   string
}

const (
	// NodeDrainPodResultSuccess node drain pod result success
	NodeDrainPodResultSuccess = "success"
	// NodeDrainPodResultFailed node drain pod result failed
	NodeDrainPodResultFailed = "failed"
	// NodeDrainPodResultPartFailed node drain pod result part failed
	NodeDrainPodResultPartFailed = "partial_failed"
)

// NodeDrainPodResult record node drain pod result
type NodeDrainPodResult struct {
	NodeName string
	Result   string
	Err      error
}

// ExecNodeDrainPodTask exec node drain pod task
func ExecNodeDrainPodTask(ctx context.Context, paras CreateNodeDrainPodTaskParas) []NodeDrainPodResult {
	taskID, stepName := cloudprovider.GetTaskIDAndStepNameFromContext(ctx)

	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())

	barrier := utils.NewRoutinePool(50)
	defer barrier.Close()

	resultChan := make(chan NodeDrainPodResult, len(paras.NodeNames))

	for i := range paras.NodeNames {
		barrier.Add(1)
		go func(node string) {
			defer barrier.Done()

			message := fmt.Sprintf("node[%s] start drain pod in cluster[%s] ", node, paras.ClusterID)
			cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName, message)

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(paras.DrainHelper.Timeout)*time.Second)
			defer cancel()

			if err := k8sOperator.ClusterUpdateScheduleNode(ctx, clusterops.NodeInfo{
				ClusterID: paras.ClusterID,
				NodeName:  node,
				Desired:   true,
			}); err != nil {
				blog.Errorf("drainClusterNodes[%s] ClusterUpdateScheduleNode failed in cluster %s, err %s",
					node, paras.ClusterID, err.Error())

				cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
					fmt.Sprintf("node[%s] drain pod task finished", node))
				resultChan <- NodeDrainPodResult{
					NodeName: node,
					Result:   NodeDrainPodResultFailed,
					Err:      err,
				}
				return
			}

			if err := k8sOperator.DrainNode(ctx, paras.ClusterID, node, paras.DrainHelper); err != nil {
				blog.Errorf("drainClusterNodes[%s] failed in cluster %s, err %s",
					node, paras.ClusterID, err.Error())

				cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
					fmt.Sprintf("node[%s] drain pod task finished", node))
				resultChan <- NodeDrainPodResult{
					NodeName: node,
					Result:   NodeDrainPodResultPartFailed,
					Err:      err,
				}
				return
			}

			cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
				fmt.Sprintf("node[%s] drain pod task finished", node))
			resultChan <- NodeDrainPodResult{
				NodeName: node,
				Result:   NodeDrainPodResultSuccess,
			}
		}(paras.NodeNames[i])
	}
	barrier.Wait()
	close(resultChan)

	var ret []NodeDrainPodResult
	for result := range resultChan {
		ret = append(ret, result)
	}

	return ret
}
