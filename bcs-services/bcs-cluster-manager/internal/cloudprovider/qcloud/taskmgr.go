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

package qcloud

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/tasks"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/template"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
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

	// init qcloud cluster-manager task, may be call bkops interface to call extra operation

	// import task
	task.works[importClusterNodesTask] = tasks.ImportClusterNodesTask
	task.works[registerClusterKubeConfigTask] = tasks.RegisterClusterKubeConfigTask

	// create cluster task
	task.works[createClusterShieldAlarmTask] = tasks.CreateClusterShieldAlarmTask
	task.works[createTKEClusterTask] = tasks.CreateTkeClusterTask
	task.works[checkTKEClusterStatusTask] = tasks.CheckTkeClusterStatusTask
	task.works[enableTkeClusterVpcCniTask] = tasks.EnableTkeClusterVpcCniTask
	task.works[updateCreateClusterDBInfoTask] = tasks.UpdateCreateClusterDBInfoTask

	// delete cluster task
	task.works[deleteTKEClusterTask] = tasks.DeleteTKEClusterTask
	task.works[cleanClusterDBInfoTask] = tasks.CleanClusterDBInfoTask

	// add node to cluster
	task.works[addNodesShieldAlarmTask] = tasks.AddNodesShieldAlarmTask
	task.works[addNodesToClusterTask] = tasks.AddNodesToClusterTask
	task.works[checkAddNodesStatusTask] = tasks.CheckAddNodesStatusTask
	task.works[updateAddNodeDBInfoTask] = tasks.UpdateNodeDBInfoTask

	// remove node from cluster
	task.works[removeNodesFromClusterTask] = tasks.RemoveNodesFromClusterTask
	task.works[updateRemoveNodeDBInfoTask] = tasks.UpdateRemoveNodeDBInfoTask

	// init qcloud node-group task

	// create nodeGroup task
	task.works[createCloudNodeGroupTask] = tasks.CreateCloudNodeGroupTask
	task.works[checkCloudNodeGroupStatusTask] = tasks.CheckCloudNodeGroupStatusTask
	// task.works[updateCreateNodeGroupDBInfoTask] = tasks.UpdateCreateNodeGroupDBInfoTask

	// delete nodeGroup task
	task.works[deleteNodeGroupTask] = tasks.DeleteCloudNodeGroupTask
	// task.works[updateDeleteNodeGroupDBInfoTask] = tasks.UpdateDeleteNodeGroupDBInfoTask

	// clean node in nodeGroup task
	task.works[cleanNodeGroupNodesTask] = tasks.CleanNodeGroupNodesTask
	task.works[removeHostFromCMDBTask] = common.RemoveHostFromCMDBTask
	// task.works[checkCleanNodeGroupNodesStatusTask] = tasks.CheckCleanNodeGroupNodesStatusTask
	// task.works[updateCleanNodeGroupNodesDBInfoTask] = tasks.UpdateCleanNodeGroupNodesDBInfoTask

	// update desired nodes task
	task.works[applyInstanceMachinesTask] = tasks.ApplyInstanceMachinesTask
	task.works[checkClusterNodesStatusTask] = tasks.CheckClusterNodesStatusTask
	task.works[installGSEAgentTask] = common.InstallGSEAgentTask
	task.works[transferHostModuleTask] = common.TransferHostModule
	// business user define sops

	// move nodes to nodeGroup task

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
	// create cluster currently only has three steps:
	// 0. check if need to generate master instance. you need to call cvm api to produce master instance if necessary.
	//    but we only support add existed instance to cluster as master currently.
	// 1. call qcloud CreateTKECluster to create tke cluster
	// 2. call GetTKECluster to check cluster run status(cluster status: Running Creating Abnormal))
	// 3. update cluster DB info when create cluster successful
	// may be need to call external previous or behind operation by bkops

	// validate request params
	if cls == nil {
		return nil, fmt.Errorf("BuildCreateClusterTask cluster info empty")
	}
	if opt == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("BuildCreateClusterTask TaskOptions is lost")
	}

	nowStr := time.Now().Format(time.RFC3339)
	task := &proto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.CreateCluster),
		TaskName:       "创建TKE集群",
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
	// generate taskName
	taskName := fmt.Sprintf(createClusterTaskTemplate, cls.ClusterID)
	task.CommonParams["taskName"] = taskName

	passwd := utils.BuildInstancePwd()
	task.CommonParams["Password"] = passwd
	// preAction bkops

	// setting all steps details
	// step1: createTKECluster and return clusterID inject common paras
	createStep := &proto.Step{
		Name:       createTKEClusterTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Start:      nowStr,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: createTKEClusterTask,
		TaskName:   "创建集群",
	}
	createStep.Params["ClusterID"] = cls.ClusterID
	createStep.Params["CloudID"] = cls.Provider

	task.Steps[createTKEClusterTask] = createStep
	task.StepSequence = append(task.StepSequence, createTKEClusterTask)

	// step2: check cluster status by clusterID
	checkStep := &proto.Step{
		Name:       checkTKEClusterStatusTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: checkTKEClusterStatusTask,
		TaskName:   "检测集群状态",
	}
	checkStep.Params["ClusterID"] = cls.ClusterID
	checkStep.Params["CloudID"] = cls.Provider

	task.Steps[checkTKEClusterStatusTask] = checkStep
	task.StepSequence = append(task.StepSequence, checkTKEClusterStatusTask)

	// step3: enable vpc-cni network mode when cluster enable vpc-cni
	enableVpcCniStep := &proto.Step{
		Name:       enableTkeClusterVpcCniTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: enableTkeClusterVpcCniTask,
		TaskName:   "开启VPC-CNI网络模式",
	}
	enableVpcCniStep.Params["ClusterID"] = cls.ClusterID
	enableVpcCniStep.Params["CloudID"] = cls.Provider

	task.Steps[enableTkeClusterVpcCniTask] = enableVpcCniStep
	task.StepSequence = append(task.StepSequence, enableTkeClusterVpcCniTask)

	// step3: update DB info by cluster data
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

	// run bk-sops when need to postActions
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil && opt.Cloud.ClusterManagement.CreateCluster != nil {
		action := opt.Cloud.ClusterManagement.CreateCluster
		for i := range action.PostActions {
			plugin, ok := action.Plugins[action.PostActions[i]]
			if ok {
				stepName := cloudprovider.BKSOPTask + "-" + action.PostActions[i]
				step, err := template.GenerateBKopsStep(taskName, stepName, cls, plugin, template.ExtraInfo{
					InstancePasswd: passwd,
					NodeOperator:   opt.Operator,
				})
				if err != nil {
					return nil, fmt.Errorf("BuildCreateClusterTask task failed: %v", err)
				}
				task.Steps[stepName] = step
				task.StepSequence = append(task.StepSequence, stepName)
			}
		}
	}

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildCreateClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams["operator"] = opt.Operator
	task.CommonParams["user"] = opt.Operator
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.CreateClusterJob.String()

	return task, nil
}

// BuildImportClusterTask build import cluster task
func (t *Task) BuildImportClusterTask(cls *proto.Cluster, opt *cloudprovider.ImportClusterOption) (*proto.Task, error) {
	// import cluster currently only has two steps:
	// 0. import cluster: call TKEInterface import cluster master and node instances from cloud(clusterID or kubeConfig)
	// 1. internal install bcs-k8s-watch & agent service; external import qcloud kubeConfig
	// may be need to call external previous or behind operation by bkops

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
		TaskName:       "纳管TKE集群",
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
	// step0: import cluster nodes step
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

	// setting all steps details
	// step1: import cluster registerKubeConfigStep
	registerKubeConfigStep := &proto.Step{
		Name:       registerClusterKubeConfigTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Start:      nowStr,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: registerClusterKubeConfigTask,
		TaskName:   "注册集群kubeConfig认证",
	}
	registerKubeConfigStep.Params["ClusterID"] = cls.ClusterID
	registerKubeConfigStep.Params["CloudID"] = cls.Provider

	task.Steps[registerClusterKubeConfigTask] = registerKubeConfigStep
	task.StepSequence = append(task.StepSequence, registerClusterKubeConfigTask)

	// step2: install cluster watch component
	common.BuildWatchComponentTaskStep(task, cls)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildImportClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams["operator"] = opt.Operator
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.ImportClusterJob.String()

	return task, nil
}

// BuildDeleteClusterTask build deleteCluster task
func (t *Task) BuildDeleteClusterTask(cls *proto.Cluster, opt *cloudprovider.DeleteClusterOption) (*proto.Task, error) {
	// delete cluster has three steps:
	// 1. clean nodeGroup nodes and delete nodeGroup Info
	// 2. call qcloud DeleteTKECluster to delete tke cluster
	// 3. clean DB cluster info and associated data info when delete successful

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
		TaskName:       "删除TKE集群",
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
	task.CommonParams["user"] = opt.Operator

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
	// step1: DeleteTKECluster delete tke cluster
	deleteStep := &proto.Step{
		Name:       deleteTKEClusterTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Start:      nowStr,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: deleteTKEClusterTask,
		TaskName:   "删除集群",
	}
	deleteStep.Params["ClusterID"] = cls.ClusterID
	deleteStep.Params["CloudID"] = cls.Provider
	deleteStep.Params["DeleteMode"] = opt.DeleteMode.String()

	task.Steps[deleteTKEClusterTask] = deleteStep
	task.StepSequence = append(task.StepSequence, deleteTKEClusterTask)

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
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.DeleteClusterJob.String()

	return task, nil
}

// BuildAddNodesToClusterTask build addNodes task
func (t *Task) BuildAddNodesToClusterTask(cls *proto.Cluster, nodes []*proto.Node,
	opt *cloudprovider.AddNodesOption) (*proto.Task, error) {
	// addNodesToCluster has three steps:
	// 1. call qcloud AddExistedInstancesToCluster to add node
	// 2. call qcloud QueryTkeClusterInstances to check instance status(running initializing failed))
	// 3. update node DB info when add node successful
	// may be need to call external previous or behind operation by bkops

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
	nodeIDs := make([]string, 0)
	nodeIPs := make([]string, 0)
	for i := range nodes {
		nodeIPs = append(nodeIPs, nodes[i].InnerIP)
		nodeIDs = append(nodeIDs, nodes[i].NodeID)
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
	taskName := fmt.Sprintf(tkeAddNodeTaskTemplate, cls.ClusterID)
	task.CommonParams["taskName"] = taskName

	passwd := utils.BuildInstancePwd()
	task.CommonParams["Password"] = passwd

	// setting all steps details
	// step1: addNodesToTKECluster add node to cluster
	addStep := &proto.Step{
		Name:       addNodesToClusterTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Start:      nowStr,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: addNodesToClusterTask,
		TaskName:   "添加节点",
	}
	addStep.Params["ClusterID"] = cls.ClusterID
	addStep.Params["CloudID"] = opt.Cloud.CloudID
	addStep.Params["NodeGroupID"] = opt.NodeGroupID
	addStep.Params["InitPasswd"] = opt.InitPassword
	addStep.Params["NodeIPs"] = strings.Join(nodeIPs, ",")
	addStep.Params["NodeIDs"] = strings.Join(nodeIDs, ",")

	task.Steps[addNodesToClusterTask] = addStep
	task.StepSequence = append(task.StepSequence, addNodesToClusterTask)

	// step2: check cluster add node status
	checkStep := &proto.Step{
		Name:       checkAddNodesStatusTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: checkAddNodesStatusTask,
		TaskName:   "检测节点状态",
	}
	checkStep.Params["ClusterID"] = cls.ClusterID
	checkStep.Params["CloudID"] = opt.Cloud.CloudID
	checkStep.Params["NodeGroupID"] = opt.NodeGroupID
	checkStep.Params["NodeIPs"] = strings.Join(nodeIPs, ",")
	checkStep.Params["NodeIDs"] = strings.Join(nodeIDs, ",")

	task.Steps[checkAddNodesStatusTask] = checkStep
	task.StepSequence = append(task.StepSequence, checkAddNodesStatusTask)

	// step3: update DB node info by instanceIP
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
	updateStep.Params["NodeIDs"] = strings.Join(nodeIDs, ",")

	task.Steps[updateAddNodeDBInfoTask] = updateStep
	task.StepSequence = append(task.StepSequence, updateAddNodeDBInfoTask)

	// bk-sops postAction
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil && opt.Cloud.ClusterManagement.AddNodesToCluster != nil {
		action := opt.Cloud.ClusterManagement.AddNodesToCluster

		for i := range action.PostActions {
			plugin, ok := action.Plugins[action.PostActions[i]]
			if ok {
				stepName := cloudprovider.BKSOPTask + "-" + action.PostActions[i]
				step, err := template.GenerateBKopsStep(taskName, stepName, cls, plugin, template.ExtraInfo{
					InstancePasswd: passwd,
					NodeIPList:     strings.Join(nodeIPs, ","),
					NodeOperator:   opt.Operator,
				})
				if err != nil {
					return nil, fmt.Errorf("BuildAddNodesToClusterTask task failed: %v", err)
				}
				task.Steps[stepName] = step
				task.StepSequence = append(task.StepSequence, stepName)
			}
		}
	}

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildAddNodesToClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams["operator"] = opt.Operator
	task.CommonParams["user"] = opt.Operator

	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.AddNodeJob.String()
	task.CommonParams["NodeIPs"] = strings.Join(nodeIPs, ",")
	task.CommonParams["NodeIDs"] = strings.Join(nodeIDs, ",")

	return task, nil
}

// BuildRemoveNodesFromClusterTask build removeNodes task
func (t *Task) BuildRemoveNodesFromClusterTask(cls *proto.Cluster, nodes []*proto.Node,
	opt *cloudprovider.DeleteNodesOption) (*proto.Task, error) {
	// removeNodesFromCluster has two steps:
	// 1. call qcloud DeleteTkeClusterInstance to delete node
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
		nodeIDs []string
	)
	for _, node := range nodes {
		nodeIPs = append(nodeIPs, node.InnerIP)
		nodeIDs = append(nodeIDs, node.NodeID)
	}

	// init task information
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
	taskName := fmt.Sprintf(tkeCleanNodeTaskTemplate, cls.ClusterID)
	task.CommonParams["taskName"] = taskName

	// bk-sops preAction
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil && opt.Cloud.ClusterManagement.DeleteNodesFromCluster !=
		nil {
		action := opt.Cloud.ClusterManagement.DeleteNodesFromCluster

		for i := range action.PreActions {
			plugin, ok := action.Plugins[action.PreActions[i]]
			if ok {
				stepName := cloudprovider.BKSOPTask + "-" + action.PreActions[i]
				step, err := template.GenerateBKopsStep(taskName, stepName, cls, plugin, template.ExtraInfo{
					NodeIPList: strings.Join(nodeIPs, ","),
				})
				if err != nil {
					return nil, fmt.Errorf("BuildRemoveNodesFromClusterTask task failed: %v", err)
				}
				task.Steps[stepName] = step
				task.StepSequence = append(task.StepSequence, stepName)
			}
		}
	}

	// setting all steps details
	// step1: removeNodesFromTKECluster remove nodes
	removeStep := &proto.Step{
		Name:       removeNodesFromClusterTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Start:      nowStr,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: removeNodesFromClusterTask,
		TaskName:   "删除节点",
	}
	removeStep.Params["ClusterID"] = cls.ClusterID
	removeStep.Params["CloudID"] = opt.Cloud.CloudID
	removeStep.Params["DeleteMode"] = opt.DeleteMode
	removeStep.Params["NodeIPs"] = strings.Join(nodeIPs, ",")
	removeStep.Params["NodeIDs"] = strings.Join(nodeIDs, ",")

	task.Steps[removeNodesFromClusterTask] = removeStep
	task.StepSequence = append(task.StepSequence, removeNodesFromClusterTask)

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
	updateDBStep.Params["NodeIDs"] = strings.Join(nodeIDs, ",")

	task.Steps[updateRemoveNodeDBInfoTask] = updateDBStep
	task.StepSequence = append(task.StepSequence, updateRemoveNodeDBInfoTask)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildRemoveNodesFromClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.DeleteNodeJob.String()
	task.CommonParams["NodeIPs"] = strings.Join(nodeIPs, ",")
	task.CommonParams["NodeIDs"] = strings.Join(nodeIDs, ",")

	return task, nil
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
	common.BuildEnsureAutoScalerTaskStep(task, "", group.ClusterID, group.Provider)

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
	// 1. call asg RemoveInstances to clean cluster nodes
	// because cvms return to cloud asg resource pool, all clean works are handle by asg
	// we do little task here

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
	cluster, err := cloudprovider.GetStorageModel().GetCluster(context.Background(), group.ClusterID)
	if err != nil {
		return nil, fmt.Errorf("BuildCleanNodesInGroupTask get cluster %s error, %s", group.ClusterID, err.Error())
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

	// instance passwd
	passwd := group.LaunchTemplate.InitLoginPassword
	task.CommonParams[cloudprovider.PasswordKey.String()] = passwd

	// step1. bcs default steps
	if group.NodeTemplate.BcsScaleInAddons != nil && len(group.NodeTemplate.BcsScaleInAddons.PreActions) > 0 {
		step := &template.BkSopsStepAction{
			TaskName: taskName,
			Actions:  group.NodeTemplate.BcsScaleOutAddons.PostActions,
			Plugins:  group.NodeTemplate.BcsScaleOutAddons.Plugins,
		}
		err := step.BuildBkSopsStepAction(task, opt.Cluster, template.ExtraInfo{
			InstancePasswd: passwd,
			NodeIPList:     strings.Join(nodeIPs, ","),
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
			InstancePasswd: passwd,
			NodeIPList:     strings.Join(nodeIPs, ","),
		})
		if err != nil {
			return nil, fmt.Errorf("BuildCleanNodesInGroupTask ScaleInExtraAddons.PreActions BuildBkSopsStepAction failed: %v",
				err)
		}
	}

	// setting all steps details
	// step3: cluster scaleIn to clean cluster nodes
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

	cleanStep.Params[cloudprovider.NodeIDsKey.String()] = strings.Join(nodeIDs, ",")
	cleanStep.Params[cloudprovider.NodeIDsKey.String()] = strings.Join(nodeIDs, ",")

	task.Steps[cleanNodeGroupNodesTask] = cleanStep
	task.StepSequence = append(task.StepSequence, cleanNodeGroupNodesTask)

	// step4: remove node ip from cmdb
	removeHostStep := &proto.Step{
		Name:         removeHostFromCMDBTask,
		System:       "api",
		Params:       make(map[string]string),
		Retry:        0,
		SkipOnFailed: true,
		Start:        nowStr,
		Status:       cloudprovider.TaskStatusNotStarted,
		// method name is registered name to taskserver
		TaskMethod: removeHostFromCMDBTask,
		TaskName:   cloudprovider.RemoveHostFromCMDBStep.String(),
	}
	removeHostStep.Params[cloudprovider.BKBizIDKey.String()] = cluster.BusinessID
	removeHostStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")
	task.Steps[removeHostFromCMDBTask] = removeHostStep
	task.StepSequence = append(task.StepSequence, removeHostFromCMDBTask)

	// set current step
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
	deleteStep.Params["KeepInstance"] = "true"
	if opt.CleanInstanceInCluster {
		deleteStep.Params["KeepInstance"] = "false"
	}

	task.Steps[deleteNodeGroupTask] = deleteStep
	task.StepSequence = append(task.StepSequence, deleteNodeGroupTask)

	// step2. ensure autoscaler to remove this nodegroup
	if group.EnableAutoscale {
		stepName := fmt.Sprintf("从集群 AutoScaler 移除 NodeGroup[%s]", group.NodeGroupID)
		common.BuildEnsureAutoScalerTaskStep(task, stepName, group.ClusterID, group.Provider)
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

	cluster, err := cloudprovider.GetStorageModel().GetCluster(context.Background(), group.ClusterID)
	if err != nil {
		return nil, fmt.Errorf("BuildUpdateDesiredNodesTask get cluster %s error, %s", group.ClusterID, err.Error())
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

	// step4. business user define flow
	if opt.NodeGroup != nil && opt.NodeGroup.NodeTemplate.ScaleOutExtraAddons != nil &&
		len(opt.NodeGroup.NodeTemplate.ScaleOutExtraAddons.PostActions) > 0 {
		step := &template.BkSopsStepAction{
			TaskName: taskName,
			Actions:  opt.NodeGroup.NodeTemplate.ScaleOutExtraAddons.PostActions,
			Plugins:  opt.NodeGroup.NodeTemplate.ScaleOutExtraAddons.Plugins,
		}

		err := step.BuildBkSopsStepAction(task, opt.Cluster, template.ExtraInfo{
			InstancePasswd: passwd,
			NodeIPList:     "",
		})
		if err != nil {
			return nil, fmt.Errorf("BuildUpdateDesiredNodesTask ExtraAddons.PostActions BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step5: install gse agent
	installGSEAgentStep := &proto.Step{
		Name:         installGSEAgentTask,
		System:       "api",
		Params:       make(map[string]string),
		Retry:        0,
		SkipOnFailed: true,
		Start:        nowStr,
		Status:       cloudprovider.TaskStatusNotStarted,
		// method name is registered name to taskserver
		TaskMethod: installGSEAgentTask,
		TaskName:   cloudprovider.InstallGSEAgentStep.String(),
	}
	installGSEAgentStep.Params[cloudprovider.BKBizIDKey.String()] = cluster.BusinessID
	installGSEAgentStep.Params[cloudprovider.BKCloudIDKey.String()] = strconv.Itoa(int(group.Area.BkCloudID))
	installGSEAgentStep.Params[cloudprovider.PasswordKey.String()] = passwd
	task.Steps[installGSEAgentTask] = installGSEAgentStep
	task.StepSequence = append(task.StepSequence, installGSEAgentTask)

	if group.NodeTemplate != nil && group.NodeTemplate.Module != nil &&
		len(group.NodeTemplate.Module.ScaleOutModuleID) != 0 {
		// step6: transfer host module
		transferHostModuleStep := &proto.Step{
			Name:         transferHostModuleTask,
			System:       "api",
			Params:       make(map[string]string),
			Retry:        0,
			SkipOnFailed: true,
			Start:        nowStr,
			Status:       cloudprovider.TaskStatusNotStarted,
			// method name is registered name to taskserver
			TaskMethod: transferHostModuleTask,
			TaskName:   cloudprovider.TransferHostModuleStep.String(),
		}
		transferHostModuleStep.Params[cloudprovider.BKBizIDKey.String()] = cluster.BusinessID
		transferHostModuleStep.Params[cloudprovider.BKModuleIDKey.String()] = group.NodeTemplate.Module.ScaleOutModuleID
		task.Steps[transferHostModuleTask] = transferHostModuleStep
		task.StepSequence = append(task.StepSequence, transferHostModuleTask)
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
	common.BuildEnsureAutoScalerTaskStep(task, "", group.ClusterID, group.Provider)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildSwitchNodeGroupAutoScalingTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.SwitchNodeGroupAutoScalingJob.String()
	return task, nil
}

// BuildUpdateAutoScalingOptionTask build update auto scaler option task
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
	common.BuildEnsureAutoScalerTaskStep(task, "", scalingOption.ClusterID, scalingOption.Provider)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildUpdateAutoScalingOptionTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.UpdateAutoScalingOptionJob.String()
	return task, nil
}

// BuildSwitchAutoScalingOptionStatusTask build switch auto scaler option status task
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
	common.BuildEnsureAutoScalerTaskStep(task, "", scalingOption.ClusterID, scalingOption.Provider)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildSwitchAutoScalingOptionStatusTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.SwitchAutoScalingOptionStatusJob.String()
	return task, nil
}
