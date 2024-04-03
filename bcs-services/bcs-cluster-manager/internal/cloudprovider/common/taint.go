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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	cutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var (
	// NodeSetTaintsActionStep 节点设置污点任务
	NodeSetTaintsActionStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.SetNodeTaintsAction,
		StepName:   "节点设置通用污点",
	}

	// RemoveClusterNodesInnerTaintStep 删除内置污点
	RemoveClusterNodesInnerTaintStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.RemoveClusterNodesInnerTaintAction,
		StepName:   "删除预置的污点",
	}
)

// BuildRemoveClusterNodesInnerTaintTaskStep 删除预置的污点
func BuildRemoveClusterNodesInnerTaintTaskStep(task *proto.Task, group *proto.NodeGroup) {
	removeTaintStep := cloudprovider.InitTaskStep(RemoveClusterNodesInnerTaintStep)

	removeTaintStep.Params[cloudprovider.ClusterIDKey.String()] = group.ClusterID
	removeTaintStep.Params[cloudprovider.NodeGroupIDKey.String()] = group.NodeGroupID
	removeTaintStep.Params[cloudprovider.CloudIDKey.String()] = group.Provider

	task.Steps[RemoveClusterNodesInnerTaintStep.StepMethod] = removeTaintStep
	task.StepSequence = append(task.StepSequence, RemoveClusterNodesInnerTaintStep.StepMethod)
}

// RemoveClusterNodesInnerTaintTask removes cluster nodes taint
func RemoveClusterNodesInnerTaintTask(taskID string, stepName string) error {
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

	// step login started here
	// extract parameter && check validate
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	// inject success nodesNames
	nodeNames := strings.Split(state.Task.CommonParams[cloudprovider.NodeNamesKey.String()], ",")

	if len(clusterID) == 0 || len(nodeGroupID) == 0 || len(cloudID) == 0 || len(nodeNames) == 0 {
		blog.Errorf("RemoveClusterNodesTaintTask[%s]: check parameter validate failed", taskID)
		retErr := fmt.Errorf("RemoveClusterNodesTaintTask check parameters failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("RemoveClusterNodesTaintTask[%s]: GetClusterDependBasicInfo failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("RemoveClusterNodesTaintTask GetClusterDependBasicInfo failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = removeClusterNodesTaint(ctx, dependInfo.Cluster.ClusterID, nodeNames)
	if err != nil {
		blog.Errorf("RemoveClusterNodesTaintTask[%s]: removeClusterNodesTaint failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("RemoveClusterNodesTaintTask removeClusterNodesTaint failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("RemoveClusterNodesTaintTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

func removeClusterNodesTaint(ctx context.Context, clusterID string, nodeNames []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())
	kubeCli, err := k8sOperator.GetClusterClient(clusterID)
	if err != nil {
		return err
	}

	for _, ins := range nodeNames {
		node, errLocal := kubeCli.CoreV1().Nodes().Get(context.Background(), ins, metav1.GetOptions{})
		if errLocal != nil {
			blog.Errorf("removeClusterNodesTaint[%s] nodeName[%s] failed: %v", taskID, ins, err)
			continue
		}

		newTaints := make([]corev1.Taint, 0)
		for _, taint := range node.Spec.Taints {
			if taint.Key != cutils.BCSNodeGroupTaintKey {
				newTaints = append(newTaints, taint)
			}
		}
		node.Spec.Taints = newTaints

		_, errLocal = kubeCli.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
		if errLocal != nil {
			blog.Errorf("removeClusterNodesTaint[%s] nodeName[%s] failed: %v", taskID, ins, errLocal)
			continue
		}

		blog.Errorf("removeClusterNodesTaint[%s] nodeName[%s] success", taskID, ins)
	}

	return nil
}

// BuildNodeTaintsTaskStep build node taints(user define taints) task step
func BuildNodeTaintsTaskStep(task *proto.Task, clusterID string, nodeIPs []string, taints []*proto.Taint) {
	if len(taints) == 0 {
		return
	}

	taintStep := cloudprovider.InitTaskStep(NodeSetTaintsActionStep)

	taintStep.Params[cloudprovider.ClusterIDKey.String()] = clusterID
	taintStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")

	taintBytes, _ := json.Marshal(taints)
	taintStep.Params[cloudprovider.TaintsKey.String()] = string(taintBytes)

	task.Steps[NodeSetTaintsActionStep.StepMethod] = taintStep
	task.StepSequence = append(task.StepSequence, NodeSetTaintsActionStep.StepMethod)
}

// SetNodeTaintsTask set cluster nodes taints
func SetNodeTaintsTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("SetNodeTaintsTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("SetNodeTaintsTask[%s]: run step %s, system: %s, old state: %s, params %v",
		taskID, stepName, step.System, step.Status, step.Params)

	// extract parameter
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeIPs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.NodeIPsKey.String(), ",")

	taintBytes := step.Params[cloudprovider.TaintsKey.String()]

	var taints []*proto.Taint
	err = json.Unmarshal([]byte(taintBytes), &taints)
	if err != nil {
		errMsg := fmt.Sprintf("SetNodeTaintsTask[%s] validateParameter failed: taints error", taskID)
		blog.Errorf(errMsg)
		retErr := fmt.Errorf("SetNodeTaintsTask err: %v", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	if len(clusterID) == 0 || len(nodeIPs) == 0 {
		errMsg := fmt.Sprintf("SetNodeTaintsTask[%s] validateParameter failed: clusterID or nodeIPs empty", taskID)
		blog.Errorf(errMsg)
		retErr := fmt.Errorf("SetNodeTaintsTask err: %s", errMsg)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	_ = UpdateClusterNodesTaints(ctx, NodeTaintData{
		ClusterID: clusterID,
		NodeIPs:   nodeIPs,
		Taints:    taints,
	})
	blog.Infof("SetNodeTaintsTask[%s] clusterID[%s] IPs[%v] successful", taskID, clusterID, nodeIPs)

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("task %s %s update to storage fatal", taskID, stepName)
		return err
	}

	return nil
}

// NodeTaintData Node data
type NodeTaintData struct {
	ClusterID string
	NodeNames []string
	NodeIPs   []string
	Taints    []*proto.Taint
}

// UpdateClusterNodesTaints update cluster taints
func UpdateClusterNodesTaints(ctx context.Context, data NodeTaintData) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())

	// trans nodeIPs to nodeNames: k8s cluster register nodeName not nodeIP
	nodeNames := make([]NodeInfo, 0)
	nodes, err := k8sOperator.ListClusterNodesByIPsOrNames(ctx, clusterops.ListNodeOption{
		ClusterID: data.ClusterID,
		NodeIPs:   data.NodeIPs,
		NodeNames: data.NodeNames,
	})
	if err != nil {
		blog.Errorf("UpdateClusterNodesTaints[%s] ListClusterNodesByIPsOrNames failed: %v", taskID, err)
		return err
	}
	for i := range nodes {
		nodeNames = append(nodeNames, NodeInfo{
			NodeName: nodes[i].Name,
			NodeIP: func(n *corev1.Node) string {
				ipv4s, _ := utils.GetNodeIPAddress(n)
				if len(ipv4s) > 0 {
					return ipv4s[0]
				}

				return ""
			}(nodes[i]),
			NodeTaint: func() []proto.Taint {
				var nodeTaints []proto.Taint
				for _, taint := range nodes[i].Spec.Taints {
					nodeTaints = append(nodeTaints, proto.Taint{
						Key:    taint.Key,
						Value:  taint.Value,
						Effect: string(taint.Effect),
					})
				}

				return nodeTaints
			}(),
		})
	}
	blog.Infof("UpdateClusterNodesTaints[%s] ListClusterNodesByIPsOrNames successful[%v]", taskID, nodeNames)

	for _, node := range nodeNames {
		// user defined labels
		taints := data.Taints
		if taints == nil {
			taints = make([]*proto.Taint, 0)
		}

		// merge source node labels
		for i := range node.NodeTaint {
			taints = append(taints, &proto.Taint{
				Key:    node.NodeTaint[i].Key,
				Value:  node.NodeTaint[i].Value,
				Effect: node.NodeTaint[i].Effect,
			})
		}
		err := k8sOperator.UpdateNodeTaints(ctx, data.ClusterID, node.NodeName, utils.TaintToK8sTaint(taints))
		if err != nil {
			blog.Errorf("UpdateClusterNodesTaints[%s] ip[%s] failed: %v", taskID, node.NodeName, err)
			continue
		}
		blog.Infof("UpdateClusterNodesTaints[%s] ip[%s] successful", taskID, node.NodeName)
	}

	return nil
}
