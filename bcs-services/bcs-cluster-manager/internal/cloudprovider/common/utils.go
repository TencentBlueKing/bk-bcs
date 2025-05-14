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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var (
	// CreateNamespaceActionStep 创建命名空间任务
	CreateNamespaceActionStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.CreateNamespaceAction,
		StepName:   "创建命名空间",
	}

	// DeleteNamespaceActionStep 删除命名空间任务
	DeleteNamespaceActionStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.DeleteNamespaceAction,
		StepName:   "删除命名空间",
	}

	// CreateResourceQuotaActionStep 创建资源配额任务
	CreateResourceQuotaActionStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.CreateResourceQuotaAction,
		StepName:   "创建资源配额",
	}

	// DeleteResourceQuotaActionStep 删除资源配额任务
	DeleteResourceQuotaActionStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.DeleteResourceQuotaAction,
		StepName:   "删除资源配额",
	}

	// NodeSetLabelsActionStep 节点设置标签任务
	NodeSetLabelsActionStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.SetNodeLabelsAction,
		StepName:   "节点设置通用标签",
	}

	// CheckKubeAgentStatusStep 检测kubeAgent状态
	CheckKubeAgentStatusStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.CheckKubeAgentStatusAction,
		StepName:   "kubeAgent状态检测",
	}

	// NodeSetAnnotationsActionStep 节点设置注解任务
	NodeSetAnnotationsActionStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.SetNodeAnnotationsAction,
		StepName:   "节点设置注解",
	}

	// CheckClusterCleanNodesActionStep 检测下架节点状态
	CheckClusterCleanNodesActionStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.CheckClusterCleanNodesAction,
		StepName:   "检测下架节点状态",
	}
)

// CreateClusterNamespace for cluster create namespace
func CreateClusterNamespace(ctx context.Context, clusterID string, nsInfo NamespaceDetail) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	if len(clusterID) == 0 || len(nsInfo.Namespace) == 0 {
		blog.Errorf("CreateClusterNamespace[%s:%s] resource empty", clusterID, nsInfo.Namespace)
		return fmt.Errorf("cluster or ns resource empty")
	}

	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())
	err := k8sOperator.CreateNamespace(ctx, clusterID, clusterops.NamespaceInfo{
		Name:        nsInfo.Namespace,
		Labels:      nsInfo.Labels,
		Annotations: nsInfo.Annotations,
	})
	if err != nil {
		blog.Errorf("CreateClusterNamespace[%s] resource[%s:%s] failed: %v", taskID, clusterID, nsInfo.Namespace, err)
		return err
	}

	blog.Infof("CreateClusterNamespace[%s] success[%s:%s]", taskID, clusterID, nsInfo.Namespace)

	return nil
}

// DeleteClusterNamespace for cluster delete namespace
func DeleteClusterNamespace(ctx context.Context, clusterID, name string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	if len(clusterID) == 0 || len(name) == 0 {
		blog.Errorf("DeleteClusterNamespace[%s:%s] resource empty", clusterID, name)
		return fmt.Errorf("cluster or ns resource empty")
	}

	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())
	err := k8sOperator.DeleteNamespace(ctx, clusterID, name)
	if err != nil {
		blog.Errorf("DeleteClusterNamespace[%s] resource[%s:%s] failed: %v", taskID, clusterID, name, err)
		return err
	}

	blog.Infof("DeleteClusterNamespace[%s] success[%s:%s]", taskID, clusterID, name)

	return nil
}

// BuildCheckKubeAgentStatusTaskStep build kubeAgent status task step
func BuildCheckKubeAgentStatusTaskStep(task *proto.Task, clusterID string) {
	checkStep := cloudprovider.InitTaskStep(CheckKubeAgentStatusStep)

	checkStep.Params[cloudprovider.ClusterIDKey.String()] = clusterID

	task.Steps[CheckKubeAgentStatusStep.StepMethod] = checkStep
	task.StepSequence = append(task.StepSequence, CheckKubeAgentStatusStep.StepMethod)
}

// CheckKubeAgentStatusTask check cluster kubeAgent status task
func CheckKubeAgentStatusTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start check cluster kubeagent status")
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckKubeAgentStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckKubeAgentStatusTask[%s]: run step %s, system: %s, old state: %s, params %v",
		taskID, stepName, step.System, step.Status, step.Params)

	// extract parameter
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	if len(clusterID) == 0 {
		errMsg := fmt.Sprintf("CheckKubeAgentStatusTask[%s] validateParameter failed: clusterID or "+
			"namespace empty", taskID) // nolint
		blog.Errorf(errMsg)
		retErr := fmt.Errorf("CheckKubeAgentStatusTask err: %s", errMsg)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// check cluster kubeAgent status
	err = checkKubeAgentStatusByClusterID(ctx, clusterID)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("check cluster kubeagent status failed [%s]", err))
		blog.Errorf("CheckKubeAgentStatusTask[%s] failed: %v", taskID, err)
		retErr := fmt.Errorf("CheckKubeAgentStatusTask err: %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CheckKubeAgentStatusTask[%s] clusterID[%s] kubeAgent status successful", taskID, clusterID)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"check cluster kubeagent status successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("task %s %s update to storage fatal", taskID, stepName)
		return err
	}

	return nil
}

func checkKubeAgentStatusByClusterID(ctx context.Context, clusterID string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// wait check delete component status
	timeContext, cancel := context.WithTimeout(ctx, time.Minute*10)
	defer cancel()

	err := loop.LoopDoFunc(timeContext, func() error {
		exist, err := cloudprovider.GetClusterCredentialByClusterID(timeContext, clusterID)
		if err != nil {
			blog.Errorf("checkKubeAgentStatusByClusterID[%s] failed[%s]: %v", taskID, clusterID, err)
			return nil
		}

		blog.Infof("checkKubeAgentStatusByClusterID[%s] cluster[%s]: %v", taskID, clusterID, exist)
		if exist {
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(30*time.Second))
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		blog.Errorf("checkKubeAgentStatusByClusterID[%s] cluster[%s] failed: %v", taskID, clusterID, err)
		return err
	}

	// timeout error
	if errors.Is(err, context.DeadlineExceeded) {
		blog.Infof("checkKubeAgentStatusByClusterID[%s] cluster[%s] timeout failed: %v", taskID, clusterID, err)
		return err
	}

	blog.Infof("checkKubeAgentStatusByClusterID[%s] cluster[%s] success", taskID, clusterID)
	return nil
}

// NamespaceDetail namespace info
type NamespaceDetail struct {
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
}

// BuildCreateNamespaceTaskStep build create namespace task step
func BuildCreateNamespaceTaskStep(task *proto.Task, clusterID string, ns NamespaceDetail) {
	createStep := cloudprovider.InitTaskStep(CreateNamespaceActionStep)

	createStep.Params[cloudprovider.ClusterIDKey.String()] = clusterID
	createStep.Params[cloudprovider.NamespaceKey.String()] = ns.Namespace

	if len(ns.Labels) > 0 {
		createStep.Params[cloudprovider.LabelsKey.String()] = utils.MapToStrings(ns.Labels)
	}
	if len(ns.Annotations) > 0 {
		createStep.Params[cloudprovider.AnnotationsKey.String()] = utils.MapToStrings(ns.Annotations)
	}

	task.Steps[CreateNamespaceActionStep.StepMethod] = createStep
	task.StepSequence = append(task.StepSequence, CreateNamespaceActionStep.StepMethod)
}

// CreateNamespaceTask create cluster namespace
func CreateNamespaceTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start create cluster namespace")
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CreateNamespaceTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CreateNamespaceTask[%s]: run step %s, system: %s, old state: %s, params %v",
		taskID, stepName, step.System, step.Status, step.Params)

	// extract parameter
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	namespace := step.Params[cloudprovider.NamespaceKey.String()]
	labels := cloudprovider.ParseMapFromStepParas(step.Params, cloudprovider.LabelsKey.String())
	annotations := cloudprovider.ParseMapFromStepParas(step.Params, cloudprovider.AnnotationsKey.String())

	if len(clusterID) == 0 || len(namespace) == 0 {
		errMsg := fmt.Sprintf("CreateNamespaceTask[%s] validateParameter failed: clusterID or "+
			"namespace empty", taskID)
		blog.Errorf(errMsg)
		retErr := fmt.Errorf("CreateNamespaceTask err: %s", errMsg)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// check cluster namespace and create namespace when not exist
	err = CreateClusterNamespace(ctx, clusterID, NamespaceDetail{
		Namespace:   namespace,
		Labels:      labels,
		Annotations: annotations,
	})
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("create cluster namespace failed [%s]", err))
		blog.Errorf("CreateNamespaceTask[%s] failed: %v", taskID, err)
		retErr := fmt.Errorf("CreateNamespaceTask err: %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CreateNamespaceTask[%s] clusterID[%s] namespace[%v] successful", taskID, clusterID, namespace)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"create cluster namespace successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("task %s %s update to storage fatal", taskID, stepName)
		return err
	}

	return nil
}

// BuildDeleteNamespaceTaskStep build delete namespace task step
func BuildDeleteNamespaceTaskStep(task *proto.Task, clusterID, name string) {
	deleteStep := cloudprovider.InitTaskStep(DeleteNamespaceActionStep)

	deleteStep.Params[cloudprovider.ClusterIDKey.String()] = clusterID
	deleteStep.Params[cloudprovider.NamespaceKey.String()] = name

	task.Steps[DeleteNamespaceActionStep.StepMethod] = deleteStep
	task.StepSequence = append(task.StepSequence, DeleteNamespaceActionStep.StepMethod)
}

// DeleteNamespaceTask delete cluster namespace
func DeleteNamespaceTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start delete cluster namespace")
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("DeleteNamespaceTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("DeleteNamespaceTask[%s]: run step %s, system: %s, old state: %s, params %v",
		taskID, stepName, step.System, step.Status, step.Params)

	// extract parameter
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	namespace := step.Params[cloudprovider.NamespaceKey.String()]

	if len(clusterID) == 0 || len(namespace) == 0 {
		errMsg := fmt.Sprintf("DeleteNamespaceTask[%s] validateParameter failed: clusterID or "+
			"namespace empty", taskID)
		blog.Errorf(errMsg)
		retErr := fmt.Errorf("DeleteNamespaceTask err: %s", errMsg)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// check cluster namespace and delete namespace when exist
	err = DeleteClusterNamespace(ctx, clusterID, namespace)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("delete cluster namespace failed [%s]", err))
		blog.Errorf("DeleteNamespaceTask[%s] failed: %v", taskID, err)
		retErr := fmt.Errorf("DeleteNamespaceTask err: %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("DeleteNamespaceTask[%s] clusterID[%s] namespace[%v] successful", taskID, clusterID, namespace)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"delete cluster namespace successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("task %s %s update to storage fatal", taskID, stepName)
		return err
	}

	return nil
}

// BuildNodeAnnotationsTaskStep build node annotations (user define labels && common annotations) task step
func BuildNodeAnnotationsTaskStep(task *proto.Task, clusterID string, nodeIPs []string, annotations map[string]string) {
	if len(annotations) == 0 {
		return
	}

	annotationsStep := cloudprovider.InitTaskStep(NodeSetAnnotationsActionStep)

	annotationsStep.Params[cloudprovider.ClusterIDKey.String()] = clusterID
	// annotationsStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")
	annotationsStep.Params[cloudprovider.AnnotationsKey.String()] = utils.MapToStrings(annotations)

	task.Steps[NodeSetAnnotationsActionStep.StepMethod] = annotationsStep
	task.StepSequence = append(task.StepSequence, NodeSetAnnotationsActionStep.StepMethod)
}

// SetNodeAnnotationsTask set cluster nodes annotations
func SetNodeAnnotationsTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start set node annotations")
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("SetNodeAnnotationsTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("SetNodeAnnotationsTask[%s]: run step %s, system: %s, old state: %s, params %v",
		taskID, stepName, step.System, step.Status, step.Params)

	// extract parameter
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeIPs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.NodeIPsKey.String(), ",")
	annotations := cloudprovider.ParseMapFromStepParas(step.Params, cloudprovider.AnnotationsKey.String())

	if len(clusterID) == 0 || len(nodeIPs) == 0 {
		errMsg := fmt.Sprintf("SetNodeAnnotationsTask[%s] validateParameter failed: clusterID or nodeIPs empty", taskID)
		blog.Errorf(errMsg)
		retErr := fmt.Errorf("SetNodeAnnotationsTask err: %s", errMsg)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDAndStepNameForContext(context.Background(), taskID, stepName)
	_ = updateClusterNodesAnnotations(ctx, NodeAnnotationsData{
		clusterID:   clusterID,
		nodeIPs:     nodeIPs,
		annotations: annotations,
	})
	blog.Infof("SetNodeAnnotationsTask[%s] clusterID[%s] IPs[%v] successful", taskID, clusterID, nodeIPs)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"set node annotations successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("task %s %s update to storage fatal", taskID, stepName)
		return err
	}

	return nil
}

// NodeAnnotationsData xxx
type NodeAnnotationsData struct {
	clusterID   string
	nodeIPs     []string
	annotations map[string]string
}

func updateClusterNodesAnnotations(ctx context.Context, data NodeAnnotationsData) error { // nolint
	taskID, stepName := cloudprovider.GetTaskIDAndStepNameFromContext(ctx)

	if len(data.annotations) == 0 {
		blog.Infof("updateClusterNodesAnnotations[%s] clusterID[%s] annotations empty", taskID, data.clusterID)
		return nil
	}

	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())
	// trans nodeIPs to nodeNames: k8s cluster register nodeName not nodeIP
	nodeNames := make([]string, 0)
	nodes, err := k8sOperator.ListClusterNodesByIPsOrNames(ctx, clusterops.ListNodeOption{
		ClusterID: data.clusterID,
		NodeIPs:   data.nodeIPs,
	})
	if err != nil {
		blog.Errorf("updateClusterNodesAnnotations[%s] ListClusterNodesByIPsOrNames failed: %v", taskID, err)
		nodeNames = data.nodeIPs
	} else {
		for i := range nodes {
			nodeNames = append(nodeNames, nodes[i].Name)
		}
		blog.Infof("updateClusterNodesAnnotations[%s] ListClusterNodesByIPsOrNames successful[%v]", taskID, nodeNames)
	}
	blog.Infof("updateClusterNodesAnnotations[%s] ListClusterNodesByIPsOrNames successful[%v]", taskID, nodeNames)

	for _, name := range nodeNames {
		annotations := data.annotations

		if len(annotations) == 0 {
			blog.Infof("updateClusterNodesAnnotations[%s] node[%s] annotations empty", taskID, name)
			continue
		}
		err := k8sOperator.UpdateNodeAnnotations(ctx, data.clusterID, name, annotations, true)
		if err != nil {
			blog.Errorf("updateClusterNodesAnnotations[%s] ip[%s] failed: %v", taskID, name, err)
			continue
		}
		blog.Infof("updateClusterNodesAnnotations[%s] ip[%s] successful", taskID, name)

		cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
			fmt.Sprintf("ip [%s] successful", name))
	}

	return nil
}

// BuildNodeLabelsTaskStep build node labels(user define labels && common labels) task step
func BuildNodeLabelsTaskStep(task *proto.Task, clusterID string, nodeIPs []string, labels map[string]string) {
	labelStep := cloudprovider.InitTaskStep(NodeSetLabelsActionStep)

	labelStep.Params[cloudprovider.ClusterIDKey.String()] = clusterID
	// labelStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")
	if len(labels) > 0 {
		labelStep.Params[cloudprovider.LabelsKey.String()] = utils.MapToStrings(labels)
	}

	task.Steps[NodeSetLabelsActionStep.StepMethod] = labelStep
	task.StepSequence = append(task.StepSequence, NodeSetLabelsActionStep.StepMethod)
}

// SetNodeLabelsTask set cluster nodes labels
func SetNodeLabelsTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start set cluster nodes labels")
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("SetNodeLabelsTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("SetNodeLabelsTask[%s]: run step %s, system: %s, old state: %s, params %v",
		taskID, stepName, step.System, step.Status, step.Params)

	// extract parameter
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeIPs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.NodeIPsKey.String(), ",")
	labels := cloudprovider.ParseMapFromStepParas(step.Params, cloudprovider.LabelsKey.String())

	if len(clusterID) == 0 || len(nodeIPs) == 0 {
		errMsg := fmt.Sprintf("SetNodeLabelsTask[%s] validateParameter failed: clusterID or nodeIPs empty", taskID)
		blog.Errorf(errMsg)
		retErr := fmt.Errorf("SetNodeLabelsTask err: %s", errMsg)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDAndStepNameForContext(context.Background(), taskID, stepName)
	_ = UpdateClusterNodesLabels(ctx, NodeLabelsData{
		ClusterID: clusterID,
		NodeIPs:   nodeIPs,
		Labels:    labels,
	})
	blog.Infof("SetNodeLabelsTask[%s] clusterID[%s] IPs[%v] successful", taskID, clusterID, nodeIPs)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"set cluster nodes labels successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("task %s %s update to storage fatal", taskID, stepName)
		return err
	}

	return nil
}

// NodeLabelsData xxx
type NodeLabelsData struct {
	ClusterID string
	NodeIPs   []string
	Labels    map[string]string
}

// NodeInfo xxx
type NodeInfo struct {
	NodeName   string
	NodeIP     string
	NodeLabels map[string]string
	NodeTaint  []proto.Taint
}

// UpdateClusterNodesLabels update cluster labels
func UpdateClusterNodesLabels(ctx context.Context, data NodeLabelsData) error { // nolint
	taskID, stepName := cloudprovider.GetTaskIDAndStepNameFromContext(ctx)

	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())
	// trans nodeIPs to nodeNames: k8s cluster register nodeName not nodeIP
	nodeNames := make([]NodeInfo, 0)
	nodes, err := k8sOperator.ListClusterNodesByIPsOrNames(ctx, clusterops.ListNodeOption{
		ClusterID: data.ClusterID,
		NodeIPs:   data.NodeIPs,
	})
	if err != nil {
		blog.Errorf("updateClusterNodesLabels[%s] ListClusterNodesByIPsOrNames failed: %v", taskID, err)
		return err
	}
	for i := range nodes {
		nodeNames = append(nodeNames, NodeInfo{
			NodeName: nodes[i].Name,
			NodeIP: func(n *v1.Node) string {
				ipv4s, _ := utils.GetNodeIPAddress(n)
				if len(ipv4s) > 0 {
					return ipv4s[0]
				}

				return ""
			}(nodes[i]),
			NodeLabels: nodes[i].Labels,
		})
	}
	blog.Infof("updateClusterNodesLabels[%s] ListClusterNodesByIPsOrNames successful[%v]", taskID, nodeNames)

	cls, err := cloudprovider.GetStorageModel().GetCluster(ctx, data.ClusterID)
	if err != nil {
		blog.Errorf("updateClusterNodesLabels[%s] GetCluster[%s] failed: %v", taskID, data.ClusterID, err)
	}

	hostsMap, hostIDs, err := GetCmdbNodeDetailInfo(data.NodeIPs)
	if err != nil {
		blog.Errorf("updateClusterNodesLabels[%s] GetCmdbNodeDetailInfo failed: %v", taskID, err)
	}
	hostsTopo, err := GetNodeBizRelation(hostIDs)
	if err != nil {
		blog.Errorf("updateClusterNodesLabels[%s] GetNodeBizRelation failed: %v", taskID, err)
	}

	for _, node := range nodeNames {

		blog.Infof("updateClusterNodesLabels[%s] node[%s] ip[%s] before labels: %v",
			taskID, node.NodeName, node.NodeIP, node.NodeLabels)

		// user defined labels 深拷贝本地标签配置
		labels := make(map[string]string)
		for k, v := range data.Labels {
			labels[k] = v
		}

		// cmdb labels
		h, ok := hostsMap[node.NodeIP]
		if ok {
			labels[utils.SubZoneIDLabelKey] = h.SubZoneID
			labels[utils.AssetIDLabelKey] = h.BkAssetID
			labels[utils.HostIDLabelKey] = fmt.Sprintf("%v", h.BKHostID)
			labels[utils.AgentIDLabelKey] = h.BkAgentID
			labels[utils.CloudAreaLabelKey] = fmt.Sprintf("%v", h.BkCloudID)
			topo, ok1 := hostsTopo[int(h.BKHostID)]
			if ok1 {
				labels[utils.BusinessIDLabelKey] = fmt.Sprintf("%d", topo.BkBizID)
			}
		}

		// mixed labels if cluster is mixed cluster
		if cls != nil && cls.GetIsMixed() {
			labels[utils.MixedNodeLabelKey] = utils.MixedNodeLabelValue
		}

		if len(labels) == 0 {
			blog.Infof("updateClusterNodesLabels[%s] node[%s] labels empty", taskID, node.NodeIP)
			continue
		}

		// merge source node labels
		for k, v := range node.NodeLabels {
			_, exist := labels[k]
			if !exist {
				labels[k] = v
			}
		}

		blog.Infof("updateClusterNodesLabels[%s] node[%s] ip[%s] after labels: %v",
			taskID, node.NodeName, node.NodeIP, labels)

		err := k8sOperator.UpdateNodeLabels(ctx, data.ClusterID, node.NodeName, labels)
		if err != nil {
			blog.Errorf("updateClusterNodesLabels[%s] ip[%s] failed: %v", taskID, node.NodeName, err)
			continue
		}
		blog.Infof("updateClusterNodesLabels[%s] ip[%s] successful", taskID, node.NodeName)

		cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
			fmt.Sprintf("ip [%s] successful", node.NodeName))
	}

	return nil
}

// GetCmdbNodeDetailInfo get cmdb detailed info
func GetCmdbNodeDetailInfo(ips []string) (map[string]cmdb.HostDetailData, []int, error) {
	var (
		hostsMap = make(map[string]cmdb.HostDetailData)
		hostIDs  []int
	)

	cmdbClient := cmdb.GetCmdbClient()
	hosts, err := cmdbClient.QueryAllHostInfoWithoutBiz(ips)
	if err != nil {
		blog.Warnf("QueryAllHostInfoWithoutBiz for %v failed, %s", ips, err.Error())
		return nil, nil, err
	}

	for i := range hosts {
		hostsMap[hosts[i].BKHostInnerIP] = hosts[i]
		hostIDs = append(hostIDs, int(hosts[i].BKHostID))
	}

	return hostsMap, hostIDs, nil
}

// GetNodeBizRelation get nodes cmdb topo
func GetNodeBizRelation(hostIDs []int) (map[int]cmdb.HostBizRelations, error) {
	hostTopo := make(map[int]cmdb.HostBizRelations)

	cmdbClient := cmdb.GetCmdbClient()
	relations, err := cmdbClient.FindHostBizRelations(hostIDs)
	if err != nil {
		blog.Warnf("GetNodeBizRelation for %+v failed: %v", hostIDs, err)
		return nil, err
	}

	for i := range relations {
		hostTopo[relations[i].BkHostID] = relations[i]
	}

	return hostTopo, nil
}

func checkNodeValidatePods(ctx context.Context, clusterID, nodeName string) (bool, error) { // nolint
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	k8sCli, err := clusterops.NewK8SOperator(options.GetGlobalCMOptions(),
		cloudprovider.GetStorageModel()).GetClusterClient(clusterID)
	if err != nil {
		blog.Errorf("checkNodeValidatePods[%s] failed[%s:%s]: %v", taskID, clusterID, nodeName, err)
		return false, err
	}

	listPodOptions := metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	}
	podList, err := k8sCli.CoreV1().Pods(metav1.NamespaceAll).List(ctx, listPodOptions)
	if err != nil {
		return false, err
	}

	// checking business pod
	var warnMessages []string
	for _, item := range podList.Items {
		if !canDelete(item) {
			warnMessages = append(warnMessages, fmt.Sprintf("pod: %s/%s status is %v", item.Namespace,
				item.Name, item.Status.Phase))
		}
	}
	if len(warnMessages) > 0 {
		return false, errors.New(strings.Join(warnMessages, ";"))
	}

	return true, nil
}

func canDelete(pod v1.Pod) bool { // nolint
	// ignore kube-system pod
	if pod.Namespace == metav1.NamespaceSystem ||
		pod.Namespace == utils.BkSystem || pod.Namespace == utils.BCSSystem {
		return true
	}
	// ignore DaemonSet
	for _, item := range pod.OwnerReferences {
		if item.Kind == "DaemonSet" {
			return true
		}
	}
	// ignore completed and terminated pod
	if pod.Status.Phase == v1.PodSucceeded ||
		pod.Status.Phase == v1.PodFailed {
		return true
	}
	return false
}

// ResourceQuotaDetail resource quota info
type ResourceQuotaDetail struct {
	Name        string `json:"name"`
	CpuRequests string `json:"cpuRequests"`
	CpuLimits   string `json:"cpuLimits"`
	MemRequests string `json:"memRequests"`
	MemLimits   string `json:"memLimits"`
}

// BuildCreateResourceQuotaTaskStep build create resource quota task step
func BuildCreateResourceQuotaTaskStep(task *proto.Task, clusterID string, quota ResourceQuotaDetail) {
	createStep := cloudprovider.InitTaskStep(CreateResourceQuotaActionStep)

	createStep.Params[cloudprovider.ClusterIDKey.String()] = clusterID
	createStep.Params[cloudprovider.ResourceQuotaKey.String()] = utils.ToJSONString(quota)

	task.Steps[CreateResourceQuotaActionStep.StepMethod] = createStep
	task.StepSequence = append(task.StepSequence, CreateResourceQuotaActionStep.StepMethod)
}

// CreateResourceQuotaTask create cluster namespace resource quota
func CreateResourceQuotaTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start create cluster namespace resource quota")
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CreateResourceQuotaTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CreateResourceQuotaTask[%s]: run step %s, system: %s, old state: %s, params %v",
		taskID, stepName, step.System, step.Status, step.Params)

	// extract parameter
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	quota := step.Params[cloudprovider.ResourceQuotaKey.String()]

	var quotaObject ResourceQuotaDetail
	err = utils.ToStringObject([]byte(quota), &quotaObject)
	if err != nil {
		errMsg := fmt.Sprintf("CreateResourceQuotaTask[%s] cluster[%s] validateParameter failed: "+
			"quota ToStringObject failed", taskID, clusterID)
		blog.Errorf(errMsg)
		_ = state.UpdateStepFailure(start, stepName, errors.New(errMsg))
		return errors.New(errMsg)
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// check cluster namespace resourceQuota and create namespace resourceQuota when not exist
	err = CreateNamespaceResourceQuota(ctx, clusterID, quotaObject)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("create cluster namespace resource quota failed [%s]", err))
		blog.Errorf("CreateResourceQuotaTask[%s] failed: %v", taskID, err)
		retErr := fmt.Errorf("CreateResourceQuotaTask err: %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CreateResourceQuotaTask[%s] clusterID[%s] namespace[%v] successful",
		taskID, clusterID, quotaObject.Name)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"create cluster namespace resource quota successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("task %s %s update to storage fatal", taskID, stepName)
		return err
	}

	return nil
}

// CreateNamespaceResourceQuota for cluster create namespace resource quota
func CreateNamespaceResourceQuota(ctx context.Context, clusterID string, quota ResourceQuotaDetail) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	if len(clusterID) == 0 || len(quota.Name) == 0 {
		blog.Errorf("CreateNamespaceResourceQuota[%s:%s] resource empty", clusterID, quota.Name)
		return fmt.Errorf("cluster or resourceQuota name empty")
	}

	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())
	err := k8sOperator.CreateResourceQuota(ctx, clusterID, clusterops.ResourceQuotaInfo{
		Name:        quota.Name,
		CpuRequests: quota.CpuRequests,
		CpuLimits:   quota.CpuLimits,
		MemRequests: quota.MemRequests,
		MemLimits:   quota.MemLimits,
	})
	if err != nil {
		blog.Errorf("CreateNamespaceResourceQuota[%s] resource[%s:%s] failed: %v", taskID,
			clusterID, quota.Name, err)
		return err
	}

	blog.Infof("CreateNamespaceResourceQuota[%s] success[%s:%s]", taskID, clusterID, quota.Name)

	return nil
}

// BuildDeleteResourceQuotaTaskStep build delete namespace resource quota task step
func BuildDeleteResourceQuotaTaskStep(task *proto.Task, clusterID, namespace, name string) {
	deleteStep := cloudprovider.InitTaskStep(DeleteResourceQuotaActionStep)

	if namespace == "" {
		namespace = name
	}
	deleteStep.Params[cloudprovider.ClusterIDKey.String()] = clusterID
	deleteStep.Params[cloudprovider.QuotaNameKey.String()] = name
	deleteStep.Params[cloudprovider.NamespaceKey.String()] = namespace

	task.Steps[DeleteResourceQuotaActionStep.StepMethod] = deleteStep
	task.StepSequence = append(task.StepSequence, DeleteResourceQuotaActionStep.StepMethod)
}

// DeleteResourceQuotaTask delete cluster namespace resource quota
func DeleteResourceQuotaTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start delete cluster namespace resource quota")
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("DeleteResourceQuotaTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("DeleteResourceQuotaTask[%s]: run step %s, system: %s, old state: %s, params %v",
		taskID, stepName, step.System, step.Status, step.Params)

	// extract parameter
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	namespace := step.Params[cloudprovider.NamespaceKey.String()]
	quotaName := step.Params[cloudprovider.QuotaNameKey.String()]

	if len(clusterID) == 0 || len(quotaName) == 0 || len(namespace) == 0 {
		errMsg := fmt.Sprintf("DeleteResourceQuotaTask[%s] validateParameter failed: "+
			"clusterID/namespace/quotaName empty", taskID)
		blog.Errorf(errMsg)
		retErr := fmt.Errorf("DeleteResourceQuotaTask err: %s", errMsg)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// check cluster namespace quota and delete namespace quota when exist
	err = DeleteNamespaceResourceQuota(ctx, clusterID, namespace, quotaName)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("delete cluster namespace resource quota failed [%s]", err))
		blog.Errorf("DeleteNamespaceResourceQuota[%s] failed: %v", taskID, err)
		retErr := fmt.Errorf("DeleteNamespaceResourceQuota err: %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("DeleteNamespaceResourceQuota[%s] clusterID[%s] namespace[%v] successful",
		taskID, clusterID, namespace)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"delete cluster namespace resource quota successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("task %s %s update to storage fatal", taskID, stepName)
		return err
	}

	return nil
}

// DeleteNamespaceResourceQuota for cluster delete namespace resource quota
func DeleteNamespaceResourceQuota(ctx context.Context, clusterID, namespace, name string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	if len(clusterID) == 0 || len(name) == 0 || len(namespace) == 0 {
		blog.Errorf("DeleteNamespaceResourceQuota[%s:%s:%s] resource empty", clusterID, namespace, name)
		return fmt.Errorf("cluster/ns/name resource empty")
	}

	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())

	err := k8sOperator.DeleteResourceQuota(ctx, clusterID, namespace, name)
	if err != nil {
		blog.Errorf("DeleteClusterNamespace[%s] resource[%s:%s] failed: %v", taskID, clusterID, name, err)
		return err
	}

	blog.Infof("DeleteClusterNamespace[%s] success[%s:%s]", taskID, clusterID, name)

	return nil
}

// BuildCheckClusterCleanNodesTaskStep build check cluster clean nodes task step
func BuildCheckClusterCleanNodesTaskStep(task *proto.Task, cloudID, clusterID string, nodeNames []string) {
	checkStep := cloudprovider.InitTaskStep(CheckClusterCleanNodesActionStep)

	if len(nodeNames) == 0 {
		return
	}

	checkStep.Params[cloudprovider.CloudIDKey.String()] = cloudID
	checkStep.Params[cloudprovider.ClusterIDKey.String()] = clusterID
	checkStep.Params[cloudprovider.NodeNamesKey.String()] = strings.Join(nodeNames, ",")

	task.Steps[CheckClusterCleanNodesActionStep.StepMethod] = checkStep
	task.StepSequence = append(task.StepSequence, CheckClusterCleanNodesActionStep.StepMethod)
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
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeNames := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params, cloudprovider.NodeNamesKey.String(), ",")

	if len(clusterID) == 0 || len(cloudID) == 0 || len(nodeNames) == 0 {
		blog.Errorf("CheckClusterCleanNodsTask[%s]: check parameter validate failed", taskID)
		retErr := fmt.Errorf("CheckClusterCleanNodsTask check parameters failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
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
		exist, notExist, errLocal := FilterClusterNodesByNodeNames(timeContext, dependInfo, nodeNames)
		if errLocal != nil {
			blog.Errorf("CheckClusterCleanNodsTask[%s] FilterClusterInstanceFromNodesIDs failed: %v",
				taskID, errLocal)
			return nil
		}

		blog.Infof("CheckClusterCleanNodsTask[%s] nodeIDs[%v] exist[%v] notExist[%v]",
			taskID, nodeNames, exist, notExist)

		cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
			fmt.Sprintf("nodeIDs [%v] exist [%v] notExist [%v]", nodeNames, exist, notExist))

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

// FilterClusterNodesByNodeNames filter instanceNames inClusterNodes && notInClusterNodes
func FilterClusterNodesByNodeNames(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	nodeNames []string) ([]string, []string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())

	nodes, err := k8sOperator.ListClusterNodes(context.Background(), info.Cluster.ClusterID)
	if err != nil {
		blog.Errorf("FilterClusterNodesByNodeNames[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, err)
		return nil, nil, err
	}

	var nodeNameMap = make(map[string]*v1.Node, 0)
	for i := range nodes {
		nodeNameMap[nodes[i].Name] = nodes[i]
	}

	var (
		existNodeNames    = make([]string, 0)
		notExistNodeNames = make([]string, 0)
	)

	for _, name := range nodeNames {
		_, ok := nodeNameMap[name]
		if ok {
			existNodeNames = append(existNodeNames, name)
		} else {
			notExistNodeNames = append(notExistNodeNames, name)
		}
	}

	return existNodeNames, notExistNodeNames, nil
}
