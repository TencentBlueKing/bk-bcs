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
	corev1 "k8s.io/api/core/v1"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var (
	// CheckNodePodsActionStep 节点业务Pods检查
	CheckNodePodsActionStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.CheckNodePodsAction,
		StepName:   "检测节点业务Pods",
	}
)

// BuildCheckNodePodsTaskStep 检测节点是否存在业务Pods
func BuildCheckNodePodsTaskStep(task *proto.Task, clusterId string,
	nodeNames []string, opts []cloudprovider.StepOption) {
	checkNodePodsStep := cloudprovider.InitTaskStep(CheckNodePodsActionStep, opts...)

	checkNodePodsStep.Params[cloudprovider.ClusterIDKey.String()] = clusterId
	checkNodePodsStep.Params[cloudprovider.NodeNamesKey.String()] = strings.Join(nodeNames, ",")

	task.Steps[CheckNodePodsActionStep.StepMethod] = checkNodePodsStep
	task.StepSequence = append(task.StepSequence, CheckNodePodsActionStep.StepMethod)
}

// CheckNodePodsTask check node pods
func CheckNodePodsTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start to check cluster node business pods")
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

	// step login started here && extract parameter && check validate
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeNames := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params, cloudprovider.NodeNamesKey.String(), ",")

	if len(clusterID) == 0 || len(nodeNames) == 0 {
		blog.Errorf("CheckNodePodsTask[%s] check parameter validate failed", taskID)
		retErr := fmt.Errorf("CheckNodePodsTask check parameters failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	nodePods, err := checkNodesBusinessPods(ctx, clusterID, nodeNames)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("checkNodePods failed: %s", err.Error()))
		blog.Errorf("CheckNodePodsTask[%s] checkNodesBusinessPods failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("CheckNodePodsTask checkNodesBusinessPods failed")
		if step.GetSkipOnFailed() {
			_ = state.SkipFailure(start, stepName, retErr)
			return nil
		}
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	errors := utils.NewMultiError()
	for name, pods := range nodePods {
		blog.Errorf("CheckNodePodsTask[%s] nodeName[%s] failed: %d", taskID, name, len(pods))

		if len(pods) > 0 {
			errors.Append(fmt.Errorf("node[%s] exist business podNum[%v]", name, len(pods)))
		}
	}
	if errors.HasErrors() {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("checkNodePods failed: %s, please drain business pods on the node", errors.Error()))
		blog.Errorf("CheckNodePodsTask[%s] checkNodesBusinessPods failed: %s", taskID, errors.Error())
		retErr := fmt.Errorf("CheckNodePodsTask failed: nodes exist business pod")
		if step.GetSkipOnFailed() {
			_ = state.SkipFailure(start, stepName, retErr)
			return nil
		}
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"check cluster node business pods successful")

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckNodePodsTask[%s] %s update to storage fatal", taskID, stepName)
		return err
	}

	return nil
}

func checkNodesBusinessPods(ctx context.Context, clusterId string, nodeNames []string) (
	map[string][]*corev1.Pod, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())
	var (
		nodePods  = make(map[string][]*corev1.Pod, 0)
		mulErrors = utils.NewMultiError()
	)

	var filterNamespaces []string
	if len(options.GetGlobalCMOptions().CommonConfig.SystemNameSpaces) > 0 {
		filterNamespaces = strings.Split(options.GetGlobalCMOptions().CommonConfig.SystemNameSpaces, ",")
	}

	for _, name := range nodeNames {

		// check node if exist
		_, err := k8sOperator.GetClusterNode(ctx, clusterops.QueryNodeOption{
			ClusterID: clusterId,
			NodeName:  name,
		})
		if err != nil && !strings.Contains(err.Error(), "not found") {
			mulErrors.Append(fmt.Errorf("node[%s] err %v", name, err))
			continue
		}
		if err != nil && strings.Contains(err.Error(), "not found") {
			blog.Infof("checkNodesBusinessPods[%s] node[%s] not found", taskID, name)
			continue
		}

		// get nodes pods
		pods, err := k8sOperator.GetNodePods(ctx, clusterops.GetNodePodsOption{
			ClusterID:        clusterId,
			NodeName:         name,
			FilterNamespaces: filterNamespaces,
		})
		if err != nil {
			mulErrors.Append(fmt.Errorf("node[%s] err %v", name, err))
			continue
		}
		nodePods[name] = pods

		blog.Errorf("checkNodesBusinessPods[%s] nodeName[%v:%v] success", taskID, name, len(pods))
	}

	if mulErrors.HasErrors() {
		return nil, mulErrors
	}

	return nodePods, nil
}
