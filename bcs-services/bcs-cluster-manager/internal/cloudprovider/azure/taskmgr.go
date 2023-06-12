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

package azure

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/tasks"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/template"
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
	task.works[deleteAKSKEClusterTask] = tasks.DeleteAKSClusterTask
	task.works[cleanClusterDBInfoTask] = tasks.CleanClusterDBInfoTask

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
		TaskName:       "纳管AKS集群",
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
		TaskName:       "删除AKS集群",
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
	// step1: deleteAKSKEClusterTask delete aks cluster
	deleteStep := &proto.Step{
		Name:       deleteAKSKEClusterTask,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Start:      nowStr,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: deleteAKSKEClusterTask,
		TaskName:   "删除集群",
	}
	deleteStep.Params[cloudprovider.ClusterIDKey.String()] = cls.ClusterID
	deleteStep.Params[cloudprovider.CloudIDKey.String()] = cls.Provider
	deleteStep.Params[cloudprovider.DeleteModeKey.String()] = opt.DeleteMode.String()

	task.Steps[deleteAKSKEClusterTask] = deleteStep
	task.StepSequence = append(task.StepSequence, deleteAKSKEClusterTask)

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

// BuildCreateNodeGroupTask build create node group task - 创建节点池
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
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

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
	createStep.Params[cloudprovider.CloudIDKey.String()] = group.Provider
	createStep.Params[cloudprovider.ClusterIDKey.String()] = group.ClusterID
	createStep.Params[cloudprovider.NodeGroupIDKey.String()] = group.NodeGroupID

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
	checkStep.Params[cloudprovider.CloudIDKey.String()] = group.Provider
	checkStep.Params[cloudprovider.ClusterIDKey.String()] = group.ClusterID
	checkStep.Params[cloudprovider.NodeGroupIDKey.String()] = group.NodeGroupID

	task.Steps[checkCloudNodeGroupStatusTask] = checkStep
	task.StepSequence = append(task.StepSequence, checkCloudNodeGroupStatusTask)

	// step3. ensure autoscaler in cluster
	common.BuildEnsureAutoScalerTaskStep(task, ensureAutoScalerTask, group.ClusterID, group.Provider)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildCreateNodeGroupTask task StepSequence empty")
	}

	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.CreateNodeGroupJob.String()

	return task, nil
}

// BuildCleanNodesInGroupTask clean specified nodes in NodeGroup
// including remove nodes from NodeGroup, clean data in nodes - 缩容，不保留节点
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
	cleanStep.Params[cloudprovider.CloudIDKey.String()] = group.Provider
	cleanStep.Params[cloudprovider.ClusterIDKey.String()] = group.ClusterID
	cleanStep.Params[cloudprovider.NodeGroupIDKey.String()] = group.NodeGroupID

	task.CommonParams[cloudprovider.NodeIDsKey.String()] = strings.Join(nodeIDs, ",")
	task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")

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
// @param group: need to delete - 删除节点池
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
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

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
	deleteStep.Params[cloudprovider.ClusterIDKey.String()] = group.ClusterID
	deleteStep.Params[cloudprovider.NodeGroupIDKey.String()] = group.NodeGroupID
	deleteStep.Params[cloudprovider.CloudIDKey.String()] = group.Provider
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

// BuildMoveNodesToGroupTask build move nodes to group task - 节点移入节点池
func (t *Task) BuildMoveNodesToGroupTask(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.MoveNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildUpdateDesiredNodesTask build update desired nodes task - 扩容节点
func (t *Task) BuildUpdateDesiredNodesTask(desired uint32, group *proto.NodeGroup,
	opt *cloudprovider.UpdateDesiredNodeOption) (*proto.Task, error) {
	// validate request params
	if desired == 0 {
		return nil, errors.New("BuildUpdateDesiredNodesTask desired is zero")
	}
	if group == nil {
		return nil, errors.New("BuildUpdateDesiredNodesTask group info empty")
	}
	if opt == nil || len(opt.Operator) == 0 || opt.Cluster == nil {
		return nil, errors.New("BuildUpdateDesiredNodesTask TaskOptions is lost")
	}

	cluster, err := cloudprovider.GetStorageModel().GetCluster(context.Background(), group.ClusterID)
	if err != nil {
		return nil, errors.Wrapf(err, "BuildUpdateDesiredNodesTask get cluster %s error.", group.ClusterID)
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
	// step1. scale up to node pool
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
	applyInstanceStep.Params[cloudprovider.OperatorKey.String()] = opt.Operator
	applyInstanceStep.Params[cloudprovider.CloudIDKey.String()] = group.Provider
	applyInstanceStep.Params[cloudprovider.ClusterIDKey.String()] = group.ClusterID
	applyInstanceStep.Params[cloudprovider.NodeGroupIDKey.String()] = group.NodeGroupID
	applyInstanceStep.Params[cloudprovider.ScalingKey.String()] = strconv.Itoa(int(desired))

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
	checkClusterNodeStatusStep.Params[cloudprovider.CloudIDKey.String()] = group.Provider
	checkClusterNodeStatusStep.Params[cloudprovider.ClusterIDKey.String()] = group.ClusterID
	checkClusterNodeStatusStep.Params[cloudprovider.NodeGroupIDKey.String()] = group.NodeGroupID

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
		err = step.BuildBkSopsStepAction(task, opt.Cluster, template.ExtraInfo{
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

		err = step.BuildBkSopsStepAction(task, opt.Cluster, template.ExtraInfo{
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
	installGSEAgentStep.Params[cloudprovider.PasswordKey.String()] = passwd
	installGSEAgentStep.Params[cloudprovider.UsernameKey.String()] = group.LaunchTemplate.InitLoginUsername
	installGSEAgentStep.Params[cloudprovider.BKBizIDKey.String()] = cluster.BusinessID
	installGSEAgentStep.Params[cloudprovider.BKCloudIDKey.String()] = strconv.Itoa(int(group.Area.BkCloudID))
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

// BuildSwitchNodeGroupAutoScalingTask ensure auto scaler status and update nodegroup status to normal - 开启/关闭 节点池
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
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// setting all steps details
	// step1. ensure auto scaler
	common.BuildEnsureAutoScalerTaskStep(task, ensureAutoScalerTask, group.ClusterID, group.Provider)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildSwitchNodeGroupAutoScalingTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.SwitchNodeGroupAutoScalingJob.String()
	return task, nil
}

// BuildUpdateAutoScalingOptionTask update auto scaling option - 更新集群自动扩缩容配置
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
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// setting all steps details
	// step1. ensure auto scaler
	common.BuildEnsureAutoScalerTaskStep(task, ensureAutoScalerTask, scalingOption.ClusterID, scalingOption.Provider)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildUpdateAutoScalingOptionTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.UpdateAutoScalingOptionJob.String()
	return task, nil
}

// BuildSwitchAsOptionStatusTask switch auto scaling option status - 开启/关闭集群自动扩缩容
func (t *Task) BuildSwitchAsOptionStatusTask(scalingOption *proto.ClusterAutoScalingOption, enable bool,
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
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// setting all steps details
	// step1. ensure auto scaler
	common.BuildEnsureAutoScalerTaskStep(task, ensureAutoScalerTask, scalingOption.ClusterID, scalingOption.Provider)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildSwitchAutoScalingOptionStatusTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.SwitchAutoScalingOptionStatusJob.String()
	return task, nil
}
