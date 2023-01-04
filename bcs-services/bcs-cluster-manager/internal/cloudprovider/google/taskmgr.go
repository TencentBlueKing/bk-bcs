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
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildCleanNodesInGroupTask clean specified nodes in NodeGroup
// including remove nodes from NodeGroup, clean data in nodes
func (t *Task) BuildCleanNodesInGroupTask(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.CleanNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildDeleteNodeGroupTask when delete nodegroup, we need to create background
// task to clean all nodes in nodegroup, release all resource in cloudprovider,
// finnally delete nodes information in local storage.
// @param group: need to delete
func (t *Task) BuildDeleteNodeGroupTask(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.DeleteNodeGroupOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildMoveNodesToGroupTask build move nodes to group task
func (t *Task) BuildMoveNodesToGroupTask(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.MoveNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildUpdateDesiredNodesTask build update desired nodes task
func (t *Task) BuildUpdateDesiredNodesTask(desired uint32, group *proto.NodeGroup,
	opt *cloudprovider.UpdateDesiredNodeOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildSwitchNodeGroupAutoScalingTask ensure auto scaler status and update nodegroup status to normal
func (t *Task) BuildSwitchNodeGroupAutoScalingTask(group *proto.NodeGroup, enable bool,
	opt *cloudprovider.SwitchNodeGroupAutoScalingOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildUpdateAutoScalingOptionTask update auto scaling option
func (t *Task) BuildUpdateAutoScalingOptionTask(scalingOption *proto.ClusterAutoScalingOption,
	opt *cloudprovider.UpdateScalingOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildSwitchAutoScalingOptionStatusTask switch auto scaling option status
func (t *Task) BuildSwitchAutoScalingOptionStatusTask(scalingOption *proto.ClusterAutoScalingOption, enable bool,
	opt *cloudprovider.CommonOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
