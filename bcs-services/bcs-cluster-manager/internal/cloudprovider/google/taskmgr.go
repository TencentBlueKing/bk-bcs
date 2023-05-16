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
 *
 */

package google

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/google/tasks"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/template"

	"github.com/google/uuid"
)

var taskMgr sync.Once

func init() {
	taskMgr.Do(func() {
		cloudprovider.InitTaskManager(cloudName, newtask())
	})
}

func newtask() *Task {
	task := &Task{
		works: make(map[string]interface{}),
	}

	// import task
	task.works[importClusterNodesTask] = tasks.ImportClusterNodesTask
	task.works[registerClusterKubeConfigTask] = tasks.RegisterClusterKubeConfigTask

	// delete cluster task
	task.works[deleteGKEClusterTask] = tasks.DeleteGKEClusterTask
	task.works[cleanClusterDBInfoTask] = tasks.CleanClusterDBInfoTask

	// create nodeGroup task
	task.works[createCloudNodeGroupTask] = tasks.CreateCloudNodeGroupTask
	task.works[checkCloudNodeGroupStatusTask] = tasks.CheckCloudNodeGroupStatusTask

	// delete nodeGroup task
	task.works[deleteNodeGroupTask] = tasks.DeleteCloudNodeGroupTask

	// autoScaler task
	task.works[ensureAutoScalerTask] = tasks.EnsureAutoScalerTask

	// update desired nodes task
	task.works[applyInstanceMachinesTask] = tasks.ApplyInstanceMachinesTask
	task.works[checkClusterNodesStatusTask] = tasks.CheckClusterNodesStatusTask

	// clean node in nodeGroup task
	task.works[cleanNodeGroupNodesTask] = tasks.CleanNodeGroupNodesTask

	return task
}

// Task background task manager
type Task struct {
	works map[string]interface{}
}

// Name get cloudName
func (t *Task) Name() string {
	return cloudName
}

// GetAllTask register all backgroup task for worker running
func (t *Task) GetAllTask() map[string]interface{} {
	return t.works
}

// BuildCreateClusterTask build create cluster task
func (t *Task) BuildCreateClusterTask(cls *proto.Cluster, opt *cloudprovider.CreateClusterOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildImportClusterTask build import cluster task
func (t *Task) BuildImportClusterTask(cls *proto.Cluster, opt *cloudprovider.ImportClusterOption) (*proto.Task, error) {
	// validate request params
	if cls == nil {
		return nil, fmt.Errorf("BuildImportClusterTask cluster info empty")
	}
	if opt == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("BuildImportClusterTask TaskOptions is lost")
	}

	// init task information
	nowStr := time.Now().Format(time.RFC3339)
	task := &proto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.ImportCluster),
		TaskName:       "纳管GKE集群",
		Status:         cloudprovider.TaskStatusInit,
		Message:        "task initializing",
		Start:          nowStr,
		Steps:          make(map[string]*proto.Step),
		StepSequence:   make([]string, 0),
		ClusterID:      cls.ClusterID,
		ProjectID:      cls.ProjectID,
		Creator:        opt.Operator,
		Updater:        opt.Operator,
		LastUpdate:     nowStr,
		CommonParams:   make(map[string]string),
		ForceTerminate: false,
	}

	// preAction bkops

	// setting all steps details
	// step1: import cluster registerKubeConfigStep
	registerKubeConfigStep := &proto.Step{
		Name:       registerClusterKubeConfigTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      3,
		Start:      nowStr,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: registerClusterKubeConfigTask,
		TaskName:   "注册集群kubeConfig认证",
	}
	registerKubeConfigStep.Params[cloudprovider.ClusterIDKey.String()] = cls.ClusterID
	registerKubeConfigStep.Params[cloudprovider.CloudIDKey.String()] = cls.Provider

	task.Steps[registerClusterKubeConfigTask] = registerKubeConfigStep
	task.StepSequence = append(task.StepSequence, registerClusterKubeConfigTask)

	// setting all steps details
	// step2: import cluster nodes step
	importNodesStep := &proto.Step{
		Name:       importClusterNodesTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      3,
		Start:      nowStr,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: importClusterNodesTask,
		TaskName:   "导入集群节点",
	}
	importNodesStep.Params[cloudprovider.ClusterIDKey.String()] = cls.ClusterID
	importNodesStep.Params[cloudprovider.CloudIDKey.String()] = cls.Provider

	task.Steps[importClusterNodesTask] = importNodesStep
	task.StepSequence = append(task.StepSequence, importClusterNodesTask)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildImportClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.OperatorKey.String()] = opt.Operator
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.ImportClusterJob.String()

	return task, nil
}

// BuildDeleteClusterTask build deleteCluster task
func (t *Task) BuildDeleteClusterTask(cls *proto.Cluster, opt *cloudprovider.DeleteClusterOption) (*proto.Task, error) {
	// validate request params
	if cls == nil {
		return nil, fmt.Errorf("BuildDeleteClusterTask cluster info empty")
	}
	if opt == nil || opt.Operator == "" || opt.Cloud == nil || opt.Cluster == nil {
		return nil, fmt.Errorf("BuildDeleteClusterTask TaskOptions is lost")
	}

	// init task information
	nowStr := time.Now().Format(time.RFC3339)
	task := &proto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.DeleteCluster),
		TaskName:       "删除GKE集群",
		Status:         cloudprovider.TaskStatusInit,
		Message:        "task initializing",
		Start:          nowStr,
		Steps:          make(map[string]*proto.Step),
		StepSequence:   make([]string, 0),
		ClusterID:      cls.ClusterID,
		ProjectID:      cls.ProjectID,
		Creator:        opt.Operator,
		Updater:        opt.Operator,
		LastUpdate:     nowStr,
		CommonParams:   make(map[string]string),
		ForceTerminate: false,
	}
	taskName := fmt.Sprintf(deleteClusterTaskTemplate, cls.ClusterID)
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName
	task.CommonParams[cloudprovider.OperatorKey.String()] = opt.Operator

	// run bk-sops, current only depend on bksops create task and only need to create one task
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil && opt.Cloud.ClusterManagement.DeleteCluster != nil {
		action := opt.Cloud.ClusterManagement.DeleteCluster

		for i := range action.PreActions {
			plugin, ok := action.Plugins[action.PreActions[i]]
			if ok {
				stepName := cloudprovider.BKSOPTask + "-" + action.PreActions[i]
				step, err := template.GenerateBKopsStep(taskName, stepName, cls, plugin, template.ExtraInfo{})
				if err != nil {
					return nil, fmt.Errorf("BuildDeleteClusterTask task failed: %v", err)
				}
				task.Steps[stepName] = step
				task.StepSequence = append(task.StepSequence, stepName)
			}
		}
	}

	// setting all steps details
	// step1: deleteGKEClusterTask delete gke cluster
	deleteStep := &proto.Step{
		Name:       deleteGKEClusterTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Start:      nowStr,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: deleteGKEClusterTask,
		TaskName:   "删除集群",
	}
	deleteStep.Params[cloudprovider.ClusterIDKey.String()] = cls.ClusterID
	deleteStep.Params[cloudprovider.CloudIDKey.String()] = cls.Provider
	deleteStep.Params[cloudprovider.DeleteModeKey.String()] = opt.DeleteMode.String()

	task.Steps[deleteGKEClusterTask] = deleteStep
	task.StepSequence = append(task.StepSequence, deleteGKEClusterTask)

	// step2: update cluster DB info and associated data
	updateStep := &proto.Step{
		Name:       cleanClusterDBInfoTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: cleanClusterDBInfoTask,
		TaskName:   "更新任务状态",
	}
	updateStep.Params[cloudprovider.ClusterIDKey.String()] = cls.ClusterID
	updateStep.Params[cloudprovider.CloudIDKey.String()] = cls.Provider

	task.Steps[cleanClusterDBInfoTask] = updateStep
	task.StepSequence = append(task.StepSequence, cleanClusterDBInfoTask)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildDeleteClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.DeleteClusterJob.String()

	return task, nil
}

// BuildAddNodesToClusterTask build addNodes task
func (t *Task) BuildAddNodesToClusterTask(cls *proto.Cluster, nodes []*proto.Node,
	opt *cloudprovider.AddNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildRemoveNodesFromClusterTask build removeNodes task
func (t *Task) BuildRemoveNodesFromClusterTask(cls *proto.Cluster, nodes []*proto.Node,
	opt *cloudprovider.DeleteNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildCreateNodeGroupTask build create node group task
func (t *Task) BuildCreateNodeGroupTask(group *proto.NodeGroup, opt *cloudprovider.CreateNodeGroupOption) (
	*proto.Task, error) {
	// validate request params
	if group == nil {
		return nil, fmt.Errorf("BuildCreateNodeGroupTask group info empty")
	}
	if opt == nil {
		return nil, fmt.Errorf("BuildCreateNodeGroupTask TaskOptions is lost")
	}

	nowStr := time.Now().Format(time.RFC3339)
	task := &proto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.CreateNodeGroup),
		TaskName:       cloudprovider.CreateNodeGroupTask.String(),
		Status:         cloudprovider.TaskStatusInit,
		Message:        "task initializing",
		Start:          nowStr,
		Steps:          make(map[string]*proto.Step),
		StepSequence:   make([]string, 0),
		ClusterID:      group.ClusterID,
		ProjectID:      group.ProjectID,
		Creator:        group.Creator,
		Updater:        group.Updater,
		LastUpdate:     nowStr,
		CommonParams:   make(map[string]string),
		ForceTerminate: false,
		NodeGroupID:    group.NodeGroupID,
	}
	// generate taskName
	taskName := fmt.Sprintf(createNodeGroupTaskTemplate, group.ClusterID, group.Name)
	task.CommonParams["taskName"] = taskName

	// setting all steps details
	// step1. call qcloud create node group
	createStep := &proto.Step{
		Name:       createCloudNodeGroupTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Start:      nowStr,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: createCloudNodeGroupTask,
		TaskName:   "创建 NodeGroup",
	}
	createStep.Params["ClusterID"] = group.ClusterID
	createStep.Params["NodeGroupID"] = group.NodeGroupID
	createStep.Params["CloudID"] = group.Provider

	task.Steps[createCloudNodeGroupTask] = createStep
	task.StepSequence = append(task.StepSequence, createCloudNodeGroupTask)

	// step2. wait qcloud create node group complete
	checkStep := &proto.Step{
		Name:       checkCloudNodeGroupStatusTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: checkCloudNodeGroupStatusTask,
		TaskName:   "检测 NodeGroup 状态",
	}
	checkStep.Params["ClusterID"] = group.ClusterID
	checkStep.Params["NodeGroupID"] = group.NodeGroupID
	checkStep.Params["CloudID"] = group.Provider

	task.Steps[checkCloudNodeGroupStatusTask] = checkStep
	task.StepSequence = append(task.StepSequence, checkCloudNodeGroupStatusTask)

	// step3. ensure autoscaler in cluster
	ensureCAStep := &proto.Step{
		Name:       ensureAutoScalerTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      3,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: ensureAutoScalerTask,
		TaskName:   "开启自动伸缩组件",
	}
	ensureCAStep.Params["ClusterID"] = group.ClusterID
	ensureCAStep.Params["NodeGroupID"] = group.NodeGroupID
	ensureCAStep.Params["CloudID"] = group.Provider

	task.Steps[ensureAutoScalerTask] = ensureCAStep
	task.StepSequence = append(task.StepSequence, ensureAutoScalerTask)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildCreateNodeGroupTask task StepSequence empty")
	}

	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.CreateNodeGroupJob.String()

	return task, nil
}

// BuildCleanNodesInGroupTask clean specified nodes in NodeGroup
// including remove nodes from NodeGroup, clean data in nodes
func (t *Task) BuildCleanNodesInGroupTask(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.CleanNodesOption) (*proto.Task, error) {

	// clean nodeGroup nodes in cloud only has two steps:
	// 1. call MIG deleteInstances to delete cluster nodes

	// validate request params
	if nodes == nil {
		return nil, fmt.Errorf("BuildCleanNodesInGroupTask nodes info empty")
	}
	if group == nil {
		return nil, fmt.Errorf("BuildCleanNodesInGroupTask group info empty")
	}
	if opt == nil || len(opt.Operator) == 0 || opt.Cluster == nil {
		return nil, fmt.Errorf("BuildCleanNodesInGroupTask TaskOptions is lost")
	}

	var (
		nodeIPs, nodeIDs = make([]string, 0), make([]string, 0)
	)
	for _, node := range nodes {
		nodeIPs = append(nodeIPs, node.InnerIP)
		nodeIDs = append(nodeIDs, node.NodeID)
	}

	nowStr := time.Now().Format(time.RFC3339)
	task := &proto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.CleanNodeGroupNodes),
		TaskName:       cloudprovider.CleanNodesInGroupTask.String(),
		Status:         cloudprovider.TaskStatusInit,
		Message:        "task initializing",
		Start:          nowStr,
		Steps:          make(map[string]*proto.Step),
		StepSequence:   make([]string, 0),
		ClusterID:      group.ClusterID,
		ProjectID:      group.ProjectID,
		Creator:        group.Creator,
		Updater:        group.Updater,
		LastUpdate:     nowStr,
		CommonParams:   make(map[string]string),
		ForceTerminate: false,
		NodeGroupID:    group.NodeGroupID,
	}
	// generate taskName
	taskName := fmt.Sprintf(cleanNodeGroupNodesTaskTemplate, group.ClusterID, group.Name)
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// step1. bcs default steps
	if group.NodeTemplate.BcsScaleInAddons != nil && len(group.NodeTemplate.BcsScaleInAddons.PreActions) > 0 {
		step := &template.BkSopsStepAction{
			TaskName: taskName,
			Actions:  group.NodeTemplate.BcsScaleOutAddons.PostActions,
			Plugins:  group.NodeTemplate.BcsScaleOutAddons.Plugins,
		}
		err := step.BuildBkSopsStepAction(task, opt.Cluster, template.ExtraInfo{
			//InstancePasswd: passwd,
			NodeIPList: strings.Join(nodeIPs, ","),
		})
		if err != nil {
			return nil, fmt.Errorf("BuildCleanNodesInGroupTask BcsScaleInAddons.PreActions BuildBkSopsStepAction failed: %v",
				err)
		}
	}

	// step2. business user define flow
	if group.NodeTemplate.ScaleInExtraAddons != nil && len(group.NodeTemplate.ScaleInExtraAddons.PreActions) > 0 {
		step := &template.BkSopsStepAction{
			TaskName: taskName,
			Actions:  group.NodeTemplate.ScaleInExtraAddons.PreActions,
			Plugins:  group.NodeTemplate.ScaleInExtraAddons.Plugins,
		}

		err := step.BuildBkSopsStepAction(task, opt.Cluster, template.ExtraInfo{
			//InstancePasswd: passwd,
			NodeIPList: strings.Join(nodeIPs, ","),
		})
		if err != nil {
			return nil, fmt.Errorf("BuildCleanNodesInGroupTask ScaleInExtraAddons.PreActions BuildBkSopsStepAction failed: %v",
				err)
		}
	}

	// setting all steps details
	// step1: cluster scaleIn to clean cluster nodes
	cleanStep := &proto.Step{
		Name:   cleanNodeGroupNodesTask,
		System: "api",
		Params: make(map[string]string),
		Retry:  0,
		Start:  nowStr,
		Status: cloudprovider.TaskStatusNotStarted,
		// method name is registered name to taskserver
		TaskMethod: cleanNodeGroupNodesTask,
		TaskName:   cloudprovider.CleanNodeGroupNodesStep.String(),
	}
	cleanStep.Params[cloudprovider.ClusterIDKey.String()] = group.ClusterID
	cleanStep.Params[cloudprovider.NodeGroupIDKey.String()] = group.NodeGroupID
	cleanStep.Params[cloudprovider.CloudIDKey.String()] = group.Provider

	cleanStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")
	cleanStep.Params[cloudprovider.NodeIDsKey.String()] = strings.Join(nodeIDs, ",")

	task.Steps[cleanNodeGroupNodesTask] = cleanStep
	task.StepSequence = append(task.StepSequence, cleanNodeGroupNodesTask)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildCleanNodesInGroupTask task StepSequence empty")
	}

	task.CurrentStep = task.StepSequence[0]

	// set global task paras
	task.CommonParams[cloudprovider.NodeIDsKey.String()] = strings.Join(nodeIDs, ",")
	task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")

	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.CleanNodeGroupNodesJob.String()
	return task, nil
}

// BuildDeleteNodeGroupTask when delete nodegroup, we need to create background
// task to clean all nodes in nodegroup, release all resource in cloudprovider,
// finnally delete nodes information in local storage.
// @param group: need to delete
func (t *Task) BuildDeleteNodeGroupTask(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.DeleteNodeGroupOption) (*proto.Task, error) {
	// validate request params
	if group == nil {
		return nil, fmt.Errorf("BuildDeleteNodeGroupTask group info empty")
	}
	if opt == nil {
		return nil, fmt.Errorf("BuildDeleteNodeGroupTask TaskOptions is lost")
	}

	nowStr := time.Now().Format(time.RFC3339)
	task := &proto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.DeleteNodeGroup),
		TaskName:       "删除 NodeGroup",
		Status:         cloudprovider.TaskStatusInit,
		Message:        "task initializing",
		Start:          nowStr,
		Steps:          make(map[string]*proto.Step),
		StepSequence:   make([]string, 0),
		ClusterID:      group.ClusterID,
		ProjectID:      group.ProjectID,
		Creator:        group.Creator,
		Updater:        group.Updater,
		LastUpdate:     nowStr,
		CommonParams:   make(map[string]string),
		ForceTerminate: false,
		NodeGroupID:    group.NodeGroupID,
	}
	// generate taskName
	taskName := fmt.Sprintf(deleteNodeGroupTaskTemplate, group.ClusterID, group.Name)
	task.CommonParams["taskName"] = taskName

	// setting all steps details
	// step1. call qcloud delete node group
	deleteStep := &proto.Step{
		Name:       deleteNodeGroupTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Start:      nowStr,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: deleteNodeGroupTask,
		TaskName:   "删除云 NodeGroup",
	}
	deleteStep.Params["ClusterID"] = group.ClusterID
	deleteStep.Params["NodeGroupID"] = group.NodeGroupID
	deleteStep.Params["CloudID"] = group.Provider

	task.Steps[deleteNodeGroupTask] = deleteStep
	task.StepSequence = append(task.StepSequence, deleteNodeGroupTask)

	// step2. ensure autoscaler to remove this nodegroup
	if group.EnableAutoscale {
		ensureAutoScalerStep := &proto.Step{
			Name:       ensureAutoScalerTask,
			System:     "api",
			Params:     make(map[string]string),
			Retry:      3,
			Start:      nowStr,
			Status:     cloudprovider.TaskStatusNotStarted,
			TaskMethod: ensureAutoScalerTask,
			TaskName:   fmt.Sprintf("从集群 AutoScaler 移除 NodeGroup[%s]", group.NodeGroupID),
		}
		ensureAutoScalerStep.Params["ClusterID"] = group.ClusterID
		ensureAutoScalerStep.Params["NodeGroupID"] = group.NodeGroupID
		ensureAutoScalerStep.Params["CloudID"] = group.Provider

		task.Steps[ensureAutoScalerTask] = ensureAutoScalerStep
		task.StepSequence = append(task.StepSequence, ensureAutoScalerTask)
	}

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildDeleteNodeGroupTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.DeleteNodeGroupJob.String()
	return task, nil
}

// BuildMoveNodesToGroupTask build move nodes to group task
func (t *Task) BuildMoveNodesToGroupTask(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.MoveNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildUpdateDesiredNodesTask build update desired nodes task
func (t *Task) BuildUpdateDesiredNodesTask(desired uint32, group *proto.NodeGroup,
	opt *cloudprovider.UpdateDesiredNodeOption) (*proto.Task, error) {
	// validate request params
	if desired == 0 {
		return nil, fmt.Errorf("BuildUpdateDesiredNodesTask desired is zero")
	}
	if group == nil {
		return nil, fmt.Errorf("BuildUpdateDesiredNodesTask group info empty")
	}
	if opt == nil || len(opt.Operator) == 0 || opt.Cluster == nil {
		return nil, fmt.Errorf("BuildUpdateDesiredNodesTask TaskOptions is lost")
	}

	// generate main task
	nowStr := time.Now().Format(time.RFC3339)
	task := &proto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.UpdateNodeGroupDesiredNode),
		TaskName:       cloudprovider.UpdateDesiredNodesTask.String(),
		Status:         cloudprovider.TaskStatusInit,
		Message:        "task initializing",
		Start:          nowStr,
		Steps:          make(map[string]*proto.Step),
		StepSequence:   make([]string, 0),
		ClusterID:      group.ClusterID,
		ProjectID:      group.ProjectID,
		Creator:        group.Creator,
		Updater:        group.Updater,
		LastUpdate:     nowStr,
		CommonParams:   make(map[string]string),
		ForceTerminate: false,
		NodeGroupID:    group.NodeGroupID,
	}
	// generate taskName
	taskName := fmt.Sprintf(updateNodeGroupDesiredNodeTemplate, group.ClusterID, group.Name)
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	passwd := group.LaunchTemplate.InitLoginPassword
	task.CommonParams[cloudprovider.PasswordKey.String()] = passwd

	// setting all steps details
	// step1. call qcloud interface to set desired nodes
	applyInstanceStep := &proto.Step{
		Name:       applyInstanceMachinesTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Start:      nowStr,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: applyInstanceMachinesTask,
		TaskName:   cloudprovider.ApplyInstanceMachinesStep.String(),
	}
	applyInstanceStep.Params[cloudprovider.ClusterIDKey.String()] = group.ClusterID
	applyInstanceStep.Params[cloudprovider.NodeGroupIDKey.String()] = group.NodeGroupID
	applyInstanceStep.Params[cloudprovider.CloudIDKey.String()] = group.Provider
	applyInstanceStep.Params[cloudprovider.ScalingKey.String()] = strconv.Itoa(int(desired))
	applyInstanceStep.Params[cloudprovider.OperatorKey.String()] = opt.Operator

	task.Steps[applyInstanceMachinesTask] = applyInstanceStep
	task.StepSequence = append(task.StepSequence, applyInstanceMachinesTask)

	// step2. check cluster nodes and all nodes status is running
	checkClusterNodeStatusStep := &proto.Step{
		Name:       checkClusterNodesStatusTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: checkClusterNodesStatusTask,
		TaskName:   cloudprovider.CheckClusterNodesStatusStep.String(),
	}
	checkClusterNodeStatusStep.Params[cloudprovider.ClusterIDKey.String()] = group.ClusterID
	checkClusterNodeStatusStep.Params[cloudprovider.NodeGroupIDKey.String()] = group.NodeGroupID
	checkClusterNodeStatusStep.Params[cloudprovider.CloudIDKey.String()] = group.Provider

	task.Steps[checkClusterNodesStatusTask] = checkClusterNodeStatusStep
	task.StepSequence = append(task.StepSequence, checkClusterNodesStatusTask)

	// step3. bcs default steps
	if opt.NodeGroup != nil && opt.NodeGroup.NodeTemplate.BcsScaleOutAddons != nil &&
		len(opt.NodeGroup.NodeTemplate.BcsScaleOutAddons.PostActions) > 0 {
		step := &template.BkSopsStepAction{
			TaskName: taskName,
			Actions:  opt.NodeGroup.NodeTemplate.BcsScaleOutAddons.PostActions,
			Plugins:  opt.NodeGroup.NodeTemplate.BcsScaleOutAddons.Plugins,
		}
		err := step.BuildBkSopsStepAction(task, opt.Cluster, template.ExtraInfo{
			InstancePasswd: passwd,
			NodeIPList:     "",
		})
		if err != nil {
			return nil, fmt.Errorf("BuildUpdateDesiredNodesTask BcsAddons.PostActions BuildBkSopsStepAction failed: %v", err)
		}
	}

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildUpdateDesiredNodesTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]

	// must set job-type
	task.CommonParams[cloudprovider.ScalingKey.String()] = strconv.Itoa(int(desired))
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.UpdateNodeGroupDesiredNodeJob.String()
	return task, nil
}

// BuildSwitchNodeGroupAutoScalingTask ensure auto scaler status and update nodegroup status to normal
func (t *Task) BuildSwitchNodeGroupAutoScalingTask(group *proto.NodeGroup, enable bool,
	opt *cloudprovider.SwitchNodeGroupAutoScalingOption) (*proto.Task, error) {
	// validate request params
	if group == nil {
		return nil, fmt.Errorf("BuildSwitchNodeGroupAutoScalingTask nodegroup info empty")
	}
	if opt == nil {
		return nil, fmt.Errorf("BuildSwitchNodeGroupAutoScalingTask TaskOptions is lost")
	}

	nowStr := time.Now().Format(time.RFC3339)
	task := &proto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.SwitchNodeGroupAutoScaling),
		TaskName:       "开启/关闭 节点池",
		Status:         cloudprovider.TaskStatusInit,
		Message:        "task initializing",
		Start:          nowStr,
		Steps:          make(map[string]*proto.Step),
		StepSequence:   make([]string, 0),
		ClusterID:      group.ClusterID,
		NodeGroupID:    group.NodeGroupID,
		ProjectID:      group.ProjectID,
		Creator:        group.Creator,
		Updater:        group.Updater,
		LastUpdate:     nowStr,
		CommonParams:   make(map[string]string),
		ForceTerminate: false,
	}
	// generate taskName
	taskName := fmt.Sprintf(switchNodeGroupAutoScalingTaskTemplate, group.ClusterID, group.Name)
	task.CommonParams["taskName"] = taskName

	// setting all steps details
	// step1. ensure auto scaler
	ensureStep := &proto.Step{
		Name:       ensureAutoScalerTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      3,
		Start:      nowStr,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: ensureAutoScalerTask,
		TaskName:   "安装/更新 AutoScaler",
	}
	ensureStep.Params["ClusterID"] = group.ClusterID
	ensureStep.Params["NodeGroupID"] = group.NodeGroupID
	ensureStep.Params["CloudID"] = group.Provider

	task.Steps[ensureAutoScalerTask] = ensureStep
	task.StepSequence = append(task.StepSequence, ensureAutoScalerTask)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildSwitchNodeGroupAutoScalingTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.SwitchNodeGroupAutoScalingJob.String()
	return task, nil
}

// BuildUpdateAutoScalingOptionTask update auto scaling option
func (t *Task) BuildUpdateAutoScalingOptionTask(scalingOption *proto.ClusterAutoScalingOption,
	opt *cloudprovider.UpdateScalingOption) (*proto.Task, error) {
	// validate request params
	if scalingOption == nil {
		return nil, fmt.Errorf("BuildUpdateAutoScalingOptionTask scaling option info empty")
	}
	if opt == nil {
		return nil, fmt.Errorf("BuildUpdateAutoScalingOptionTask TaskOptions is lost")
	}

	nowStr := time.Now().Format(time.RFC3339)
	task := &proto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.UpdateAutoScalingOption),
		TaskName:       "更新集群自动扩缩容配置",
		Status:         cloudprovider.TaskStatusInit,
		Message:        "task initializing",
		Start:          nowStr,
		Steps:          make(map[string]*proto.Step),
		StepSequence:   make([]string, 0),
		ClusterID:      scalingOption.ClusterID,
		ProjectID:      scalingOption.ProjectID,
		Creator:        scalingOption.Creator,
		Updater:        scalingOption.Updater,
		LastUpdate:     nowStr,
		CommonParams:   make(map[string]string),
		ForceTerminate: false,
	}
	// generate taskName
	taskName := fmt.Sprintf(updateAutoScalingOptionTemplate, scalingOption.ClusterID)
	task.CommonParams["taskName"] = taskName

	// setting all steps details
	// step1. ensure auto scaler
	ensureStep := &proto.Step{
		Name:       ensureAutoScalerTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      3,
		Start:      nowStr,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: ensureAutoScalerTask,
		TaskName:   "安装/更新 AutoScaler",
	}
	ensureStep.Params["ClusterID"] = scalingOption.ClusterID
	ensureStep.Params["CloudID"] = scalingOption.Provider

	task.Steps[ensureAutoScalerTask] = ensureStep
	task.StepSequence = append(task.StepSequence, ensureAutoScalerTask)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildUpdateAutoScalingOptionTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.UpdateAutoScalingOptionJob.String()
	return task, nil
}

// BuildSwitchAutoScalingOptionStatusTask switch auto scaling option status
func (t *Task) BuildSwitchAutoScalingOptionStatusTask(scalingOption *proto.ClusterAutoScalingOption, enable bool,
	opt *cloudprovider.CommonOption) (*proto.Task, error) {
	// validate request params
	if scalingOption == nil {
		return nil, fmt.Errorf("BuildSwitchAutoScalingOptionStatusTask scalingOption info empty")
	}
	if opt == nil {
		return nil, fmt.Errorf("BuildSwitchAutoScalingOptionStatusTask TaskOptions is lost")
	}

	nowStr := time.Now().Format(time.RFC3339)
	task := &proto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.SwitchAutoScalingOptionStatus),
		TaskName:       "开启/关闭集群自动扩缩容",
		Status:         cloudprovider.TaskStatusInit,
		Message:        "task initializing",
		Start:          nowStr,
		Steps:          make(map[string]*proto.Step),
		StepSequence:   make([]string, 0),
		ClusterID:      scalingOption.ClusterID,
		ProjectID:      scalingOption.ProjectID,
		Creator:        scalingOption.Creator,
		Updater:        scalingOption.Updater,
		LastUpdate:     nowStr,
		CommonParams:   make(map[string]string),
		ForceTerminate: false,
	}
	// generate taskName
	taskName := fmt.Sprintf(switchAutoScalingOptionStatusTemplate, scalingOption.ClusterID)
	task.CommonParams["taskName"] = taskName

	// setting all steps details
	// step1. ensure auto scaler
	ensureStep := &proto.Step{
		Name:       ensureAutoScalerTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      3,
		Start:      nowStr,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: ensureAutoScalerTask,
		TaskName:   "安装/更新 AutoScaler",
	}
	ensureStep.Params["ClusterID"] = scalingOption.ClusterID
	ensureStep.Params["CloudID"] = scalingOption.Provider

	task.Steps[ensureAutoScalerTask] = ensureStep
	task.StepSequence = append(task.StepSequence, ensureAutoScalerTask)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildSwitchAutoScalingOptionStatusTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.SwitchAutoScalingOptionStatusJob.String()
	return task, nil
}
