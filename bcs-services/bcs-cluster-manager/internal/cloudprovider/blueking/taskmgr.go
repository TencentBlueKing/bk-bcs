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

package blueking

import (
	"fmt"
	"strings"
	"sync"
	"time"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/blueking/tasks"
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

	// init blueking cluster-manager task, may be call bkops interface to call extra operation

	// import cluster task
	task.works[importClusterNodesTask] = tasks.ImportClusterNodesTask
	// create cluster task
	task.works[updateCreateClusterDBInfoTask] = tasks.UpdateCreateClusterDBInfoTask
	// delete cluster task
	task.works[cleanClusterDBInfoTask] = tasks.CleanClusterDBInfoTask
	// add node to cluster
	task.works[updateAddNodeDBInfoTask] = tasks.UpdateAddNodeDBInfoTask
	// remove node from cluster
	task.works[updateRemoveNodeDBInfoTask] = tasks.UpdateRemoveNodeDBInfoTask

	return task
}

//Task background task manager
type Task struct {
	works map[string]interface{}
}

// Name get task cloudName
func (t *Task) Name() string {
	return cloudName
}

//GetAllTask register all backgroup task for worker running
func (t *Task) GetAllTask() map[string]interface{} {
	return t.works
}

// BuildCreateClusterTask build create cluster task
func (t *Task) BuildCreateClusterTask(cls *proto.Cluster, opt *cloudprovider.CreateClusterOption) (*proto.Task, error) {
	// create cluster currently only has three steps:
	// 1. call blueking CreateTKECluster bksops to create cluster
	// 2. update cluster DB info when create cluster successful
	// may be need to call external previous or behind operation by bkops
	// validate request params
	if cls == nil {
		return nil, fmt.Errorf("BuildCreateClusterTask cluster info empty")
	}
	if opt == nil || opt.Cloud == nil || opt.Operator == "" {
		return nil, fmt.Errorf("BuildCreateClusterTask TaskOptions is lost")
	}

	// init task information
	nowStr := time.Now().Format(time.RFC3339)
	task := &proto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.CreateCluster),
		TaskName:       "创建blueking集群",
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

	taskName := fmt.Sprintf(createClusterTaskTemplate, cls.ClusterID)
	task.CommonParams["taskName"] = taskName

	// step1: call bkops preAction operation
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil && opt.Cloud.ClusterManagement.CreateCluster != nil {
		action := opt.Cloud.ClusterManagement.CreateCluster

		if len(action.PreActions) == 0 {
			errMsg := fmt.Sprintf("cloud clusterManagerment createCluster preActions empty")
			return nil, fmt.Errorf("%s BuildCreateClusterTask failed: %v", cloudName, errMsg)
		}

		for i := range action.PreActions {
			plugin, ok := action.Plugins[action.PreActions[i]]
			if ok {
				stepName := cloudprovider.BKSOPTask + "-" + action.PreActions[i]
				step, err := template.GenerateBKopsStep(taskName, stepName, cls, plugin, template.ExtraInfo{
					BusinessID: cls.BusinessID,
					Operator:   opt.Operator,
				})
				if err != nil {
					return nil, fmt.Errorf("BuildCreateClusterTask task failed: %v", err)
				}
				task.Steps[stepName] = step
				task.StepSequence = append(task.StepSequence, stepName)
			}
		}
	}

	// step2: update cluster DB info and associated data
	updateStep := &proto.Step{
		Name:       updateCreateClusterDBInfoTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: updateCreateClusterDBInfoTask,
		TaskName:   "更新任务状态",
	}
	updateStep.Params["ClusterID"] = cls.ClusterID
	updateStep.Params["CloudID"] = cls.Provider

	task.Steps[updateCreateClusterDBInfoTask] = updateStep
	task.StepSequence = append(task.StepSequence, updateCreateClusterDBInfoTask)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildCreateClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams["JobType"] = cloudprovider.CreateClusterJob.String()

	return task, nil
}

// BuildImportClusterTask build import cluster task
func (t *Task) BuildImportClusterTask(cls *proto.Cluster, opt *cloudprovider.ImportClusterOption) (*proto.Task, error) {
	// import cluster currently only has two steps:
	// 0. import cluster: call blueking import cluster master and node instances from kubeconfig
	// 1. install bcs-k8s-watch & agent service
	// may be need to call external previous or behind operation by bkops

	// validate request params
	if cls == nil {
		return nil, fmt.Errorf("BuildImportClusterTask cluster info empty")
	}
	if opt == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("BuildImportClusterTask TaskOptions is lost")
	}

	nowStr := time.Now().Format(time.RFC3339)
	task := &proto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.ImportCluster),
		TaskName:       "纳管蓝鲸云集群",
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
	// step0: create cluster shield alarm step
	importNodesStep := &proto.Step{
		Name:       importClusterNodesTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Start:      nowStr,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: importClusterNodesTask,
		TaskName:   "导入集群节点",
	}
	importNodesStep.Params["ClusterID"] = cls.ClusterID
	importNodesStep.Params["CloudID"] = cls.Provider

	task.Steps[importClusterNodesTask] = importNodesStep
	task.StepSequence = append(task.StepSequence, importClusterNodesTask)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildImportClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams["operator"] = opt.Operator
	task.CommonParams["JobType"] = cloudprovider.ImportClusterJob.String()

	return task, nil
}

// BuildDeleteClusterTask build deleteCluster task
func (t *Task) BuildDeleteClusterTask(cls *proto.Cluster, opt *cloudprovider.DeleteClusterOption) (*proto.Task, error) {
	// delete cluster has two steps:
	// 1. call blueking bkops interface to delete cluster
	// 2. clean DB cluster info and associated data info when delete successful
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
		TaskName:       "删除blueking集群",
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
	task.CommonParams["taskName"] = taskName

	// step1: call bkops preAction operation
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil && opt.Cloud.ClusterManagement.DeleteCluster != nil {
		action := opt.Cloud.ClusterManagement.DeleteCluster

		if len(action.PreActions) == 0 {
			errMsg := fmt.Sprintf("cloud clusterManagerment deleteCluster preActions empty")
			return nil, fmt.Errorf("%s BuildDeleteClusterTask failed: %v", cloudName, errMsg)
		}

		for i := range action.PreActions {
			plugin, ok := action.Plugins[action.PreActions[i]]
			if ok {
				stepName := cloudprovider.BKSOPTask + "-" + action.PreActions[i]
				step, err := template.GenerateBKopsStep(taskName, stepName, cls, plugin, template.ExtraInfo{
					BusinessID: cls.BusinessID,
					Operator:   opt.Operator,
				})
				if err != nil {
					return nil, fmt.Errorf("BuildDeleteClusterTask task failed: %v", err)
				}
				task.Steps[stepName] = step
				task.StepSequence = append(task.StepSequence, stepName)
			}
		}
	}

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
	updateStep.Params["ClusterID"] = cls.ClusterID
	updateStep.Params["CloudID"] = cls.Provider

	task.Steps[cleanClusterDBInfoTask] = updateStep
	task.StepSequence = append(task.StepSequence, cleanClusterDBInfoTask)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildDeleteClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams["JobType"] = cloudprovider.DeleteClusterJob.String()

	return task, nil
}

// BuildAddNodesToClusterTask build addNodes task
func (t *Task) BuildAddNodesToClusterTask(cls *proto.Cluster, nodes []*proto.Node, opt *cloudprovider.AddNodesOption) (*proto.Task, error) {
	// addNodesToCluster has only two steps:
	// 1. call bkops interface to add nodes to cluster
	// 2. update DB operation

	// validate request params
	if cls == nil {
		return nil, fmt.Errorf("BuildAddNodesToClusterTask cluster info empty")
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("BuildAddNodesToClusterTask lost nodes info")
	}

	if opt == nil || opt.Cloud == nil || opt.Operator == "" {
		return nil, fmt.Errorf("BuildAddNodesToClusterTask TaskOptions is lost")
	}

	// format node IPs
	nodeIPs := make([]string, 0)
	for i := range nodes {
		nodeIPs = append(nodeIPs, nodes[i].InnerIP)
	}

	// init task information
	nowStr := time.Now().Format(time.RFC3339)
	task := &proto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.AddNodesToCluster),
		TaskName:       "集群添加节点任务",
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
		NodeIPList:     nodeIPs,
	}
	taskName := fmt.Sprintf(addClusterNodesTaskTemplate, cls.ClusterID)
	task.CommonParams["taskName"] = taskName

	// step1: call bkops operation
	// validate bkops config
	if opt.Cloud == nil || opt.Cloud.ClusterManagement == nil || opt.Cloud.ClusterManagement.AddNodesToCluster == nil {
		errMsg := fmt.Sprintf("cloud clusterManagement or addNodesToCluster config empty")
		return nil, fmt.Errorf("%s BuildAddNodesToClusterTask failed: %v", cloudName, errMsg)
	}
	action := opt.Cloud.ClusterManagement.AddNodesToCluster
	if len(action.PreActions) == 0 {
		errMsg := fmt.Sprintf("cloud clusterManagerment addNodesToCluster preActions empty")
		return nil, fmt.Errorf("%s BuildAddNodesToClusterTask failed: %v", cloudName, errMsg)
	}

	// attention: bksops only need to generate one task
	for i := range action.PreActions {
		plugin, ok := action.Plugins[action.PreActions[i]]
		if ok {
			stepName := cloudprovider.BKSOPTask + "-" + action.PreActions[i]
			step, err := template.GenerateBKopsStep(taskName, stepName, cls, plugin, template.ExtraInfo{
				NodeIPList:   strings.Join(nodeIPs, ","),
				NodeOperator: opt.Operator,
				BusinessID:   cls.BusinessID,
				Operator:     opt.Operator,
			})
			if err != nil {
				return nil, fmt.Errorf("BuildAddNodesToClusterTask task failed: %v", err)
			}
			task.Steps[stepName] = step
			task.StepSequence = append(task.StepSequence, stepName)
		}
	}

	// step2: update DB node info by instanceIP
	updateStep := &proto.Step{
		Name:       updateAddNodeDBInfoTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: updateAddNodeDBInfoTask,
		TaskName:   "更新任务状态",
	}
	updateStep.Params["ClusterID"] = cls.ClusterID
	updateStep.Params["CloudID"] = opt.Cloud.CloudID
	updateStep.Params["NodeIPs"] = strings.Join(nodeIPs, ",")

	task.Steps[updateAddNodeDBInfoTask] = updateStep
	task.StepSequence = append(task.StepSequence, updateAddNodeDBInfoTask)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildAddNodesToClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams["operator"] = opt.Operator
	task.CommonParams["user"] = opt.Operator

	task.CommonParams["JobType"] = cloudprovider.AddNodeJob.String()
	task.CommonParams["NodeIPs"] = strings.Join(nodeIPs, ",")

	return task, nil
}

// BuildRemoveNodesFromClusterTask build removeNodes task
func (t *Task) BuildRemoveNodesFromClusterTask(cls *proto.Cluster, nodes []*proto.Node, opt *cloudprovider.DeleteNodesOption) (*proto.Task, error) {
	// removeNodesFromCluster has two steps:
	// 1. call blueking bkops to delete node
	// 2. update node DB info when delete node successful
	// may be need to call external previous or behind operation by bkops

	// validate request params
	if cls == nil {
		return nil, fmt.Errorf("BuildRemoveNodesFromClusterTask cluster info empty")
	}
	if opt == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("BuildRemoveNodesFromClusterTask TaskOptions is lost")
	}

	// format all nodes InnerIP
	var (
		nodeIPs []string
	)
	for _, node := range nodes {
		nodeIPs = append(nodeIPs, node.InnerIP)
	}

	//init task information
	nowStr := time.Now().Format(time.RFC3339)
	task := &proto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.RemoveNodesFromCluster),
		TaskName:       "集群删除节点任务",
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
		NodeIPList:     nodeIPs,
	}
	taskName := fmt.Sprintf(deleteClusterNodesTaskTemplate, cls.ClusterID)
	task.CommonParams["taskName"] = taskName

	// step1: build bkops task
	// validate bkops config
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil && opt.Cloud.ClusterManagement.DeleteNodesFromCluster != nil {
		action := opt.Cloud.ClusterManagement.DeleteNodesFromCluster

		for i := range action.PreActions {
			plugin, ok := action.Plugins[action.PreActions[i]]
			if !ok {
				errMsg := fmt.Sprintf("cloud clusterManagerment removeNodesFromCluster preActions %s not exist", action.PreActions[i])
				return nil, fmt.Errorf("%s BuildRemoveNodesFromClusterTask failed: %v", cloudName, errMsg)
			}
			stepName := cloudprovider.BKSOPTask + "-" + action.PreActions[i]
			step, err := template.GenerateBKopsStep(taskName, stepName, cls, plugin, template.ExtraInfo{
				NodeIPList: strings.Join(nodeIPs, ","),
				BusinessID: cls.BusinessID,
				Operator:   opt.Operator,
			})
			if err != nil {
				return nil, fmt.Errorf("BuildRemoveNodesFromClusterTask task failed: %v", err)
			}
			task.Steps[stepName] = step
			task.StepSequence = append(task.StepSequence, stepName)
		}
	}

	// step2: update node DB info
	updateDBStep := &proto.Step{
		Name:       updateRemoveNodeDBInfoTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: updateRemoveNodeDBInfoTask,
		TaskName:   "更新任务状态",
	}
	updateDBStep.Params["ClusterID"] = cls.ClusterID
	updateDBStep.Params["CloudID"] = cls.Provider
	updateDBStep.Params["NodeIPs"] = strings.Join(nodeIPs, ",")

	task.Steps[updateRemoveNodeDBInfoTask] = updateDBStep
	task.StepSequence = append(task.StepSequence, updateRemoveNodeDBInfoTask)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildRemoveNodesFromClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]

	task.CommonParams["JobType"] = cloudprovider.DeleteNodeJob.String()
	task.CommonParams["NodeIPs"] = strings.Join(nodeIPs, ",")

	return task, nil
}

//BuildCleanNodesInGroupTask clean specified nodes in NodeGroup
// including remove nodes from NodeGroup, clean data in nodes
func (t *Task) BuildCleanNodesInGroupTask(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.TaskOptions) (*proto.Task, error) {
	//build task step1: move nodes out of nodegroup
	//step2: delete nodes in cluster
	//step3: delete nodes record in local storage
	return nil, cloudprovider.ErrCloudNotImplemented
}

//BuildScalingNodesTask when scaling nodes, we need to create background
// task to verify scaling status and update new nodes to local storage
func (t *Task) BuildScalingNodesTask(scaling uint32, group *proto.NodeGroup, opt *cloudprovider.TaskOptions) (*proto.Task, error) {
	//validate request params
	return nil, nil
}

//BuildDeleteNodeGroupTask when delete nodegroup, we need to create background
//task to clean all nodes in nodegroup, release all resource in cloudprovider,
//finnally delete nodes information in local storage.
//@param group: need to delete
func (t *Task) BuildDeleteNodeGroupTask(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.DeleteNodeGroupOption) (*proto.Task, error) {
	return nil, nil
}
