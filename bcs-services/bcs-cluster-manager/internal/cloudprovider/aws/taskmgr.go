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

package aws

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/aws/tasks"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/common"
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
	task.works[importClusterNodesStep.StepMethod] = tasks.ImportClusterNodesTask
	task.works[registerClusterKubeConfigStep.StepMethod] = tasks.RegisterClusterKubeConfigTask

	// create nodeGroup task
	task.works[createCloudNodeGroupStep.StepMethod] = tasks.CreateCloudNodeGroupTask
	task.works[checkCloudNodeGroupStatusStep.StepMethod] = tasks.CheckCloudNodeGroupStatusTask

	// delete nodeGroup task
	task.works[deleteNodeGroupStep.StepMethod] = tasks.DeleteCloudNodeGroupTask

	// update desired nodes task
	task.works[applyInstanceMachinesStep.StepMethod] = tasks.ApplyInstanceMachinesTask
	task.works[checkClusterNodesStatusStep.StepMethod] = tasks.CheckClusterNodesStatusTask

	// clean node in nodeGroup task
	task.works[cleanNodeGroupNodesStep.StepMethod] = tasks.CleanNodeGroupNodesTask

	return task
}

// Task represents a task
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

// BuildCreateVirtualClusterTask build create virtual cluster task
func (t *Task) BuildCreateVirtualClusterTask(cls *proto.Cluster,
	opt *cloudprovider.CreateVirtualClusterOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildDeleteVirtualClusterTask build delete virtual cluster task
func (t *Task) BuildDeleteVirtualClusterTask(cls *proto.Cluster,
	opt *cloudprovider.DeleteVirtualClusterOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildCreateClusterTask build create cluster task
func (t *Task) BuildCreateClusterTask(cls *proto.Cluster, opt *cloudprovider.CreateClusterOption) (
	*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildImportClusterTask build import cluster task
func (t *Task) BuildImportClusterTask(cls *proto.Cluster, opt *cloudprovider.ImportClusterOption) (
	*proto.Task, error) {
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
		TaskName:       cloudprovider.ImportClusterTask.String(),
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
	taskName := fmt.Sprintf(importClusterTaskTemplate, cls.ClusterID)
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// setting all steps details
	importCluster := &ImportClusterTaskOption{Cluster: cls}
	// step1: import cluster registerKubeConfigStep
	importCluster.BuildRegisterKubeConfigStep(task)
	// step2: import cluster nodes step
	importCluster.BuildImportClusterNodesStep(task)
	// step3: install cluster watch component
	common.BuildWatchComponentTaskStep(task, cls, "")

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
func (t *Task) BuildDeleteClusterTask(cls *proto.Cluster, opt *cloudprovider.DeleteClusterOption) (
	*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
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
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// setting all steps details
	createNodeGroup := &CreateNodeGroupTaskOption{Group: group}
	// step1. call gke create node group
	createNodeGroup.BuildCreateCloudNodeGroupStep(task)
	// step2. wait gke create node group complete
	createNodeGroup.BuildCheckCloudNodeGroupStatusStep(task)
	// step3. ensure autoscaler in cluster
	common.BuildEnsureAutoScalerTaskStep(task, group.ClusterID, group.Provider)

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

	// setting all steps details
	cleanNodes := &CleanNodeInGroupTaskOption{
		Group:    group,
		NodeIPs:  nodeIPs,
		NodeIds:  nodeIDs,
		Operator: opt.Operator,
	}

	// step1: cluster scaleIn to clean cluster nodes
	common.BuildCordonNodesTaskStep(task, opt.Cluster.ClusterID, nodeIPs)
	// step2: cluster delete nodes
	cleanNodes.BuildCleanNodeGroupNodesStep(task)

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
		TaskName:       cloudprovider.DeleteNodeGroupTask.String(),
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
	deleteNodeGroup := &DeleteNodeGroupTaskOption{Group: group}
	// step1. call gke delete node group
	deleteNodeGroup.BuildDeleteNodeGroupStep(task)
	// step2: update autoscaler component
	common.BuildEnsureAutoScalerTaskStep(task, group.ClusterID, group.Provider)

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
	updateDesired := &UpdateDesiredNodesTaskOption{
		Group:    group,
		Desired:  desired,
		Operator: opt.Operator,
	}
	// step1. call qcloud interface to set desired nodes
	updateDesired.BuildApplyInstanceMachinesStep(task)
	// step2. check cluster nodes and all nodes status is running
	updateDesired.BuildCheckClusterNodeStatusStep(task)
	// install gse agent
	common.BuildInstallGseAgentTaskStep(task, &common.GseInstallInfo{
		ClusterId:  opt.Cluster.ClusterID,
		BusinessId: opt.Cluster.BusinessID,
		CloudArea:  group.GetArea(),
		User:       group.GetLaunchTemplate().GetInitLoginUsername(),
		Passwd:     passwd,
		KeyInfo:    group.GetLaunchTemplate().GetKeyPair(),
		Port:       "",
	})
	// transfer host module
	if group.NodeTemplate != nil && group.NodeTemplate.Module != nil &&
		len(group.NodeTemplate.Module.ScaleOutModuleID) != 0 {
		common.BuildTransferHostModuleStep(task, opt.Cluster.BusinessID, opt.Cluster.GetClusterBasicSettings().GetModule().
			GetWorkerModuleID(), opt.Cluster.GetClusterBasicSettings().GetModule().GetMasterModuleID())
	}
	common.BuildUnCordonNodesTaskStep(task, group.ClusterID, nil)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildUpdateDesiredNodesTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]

	// must set job-type
	task.CommonParams[cloudprovider.ScalingNodesNumKey.String()] = strconv.Itoa(int(desired))
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
		TaskName:       cloudprovider.SwitchAutoScalingOptionStatusTask.String(),
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
	common.BuildEnsureAutoScalerTaskStep(task, group.ClusterID, group.Provider)

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
		TaskName:       cloudprovider.UpdateAutoScalingOptionTask.String(),
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
	common.BuildEnsureAutoScalerTaskStep(task, scalingOption.ClusterID, scalingOption.Provider)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildUpdateAutoScalingOptionTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.UpdateAutoScalingOptionJob.String()
	return task, nil
}

// BuildSwitchAsOptionStatusTask update autoscaling status
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
		TaskName:       cloudprovider.SwitchNodeGroupAutoScalingTask.String(),
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
	common.BuildEnsureAutoScalerTaskStep(task, scalingOption.ClusterID, scalingOption.Provider)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildSwitchAutoScalingOptionStatusTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.SwitchAutoScalingOptionStatusJob.String()
	return task, nil
}

// BuildUpdateNodeGroupTask update nodeGroup task
func (t *Task) BuildUpdateNodeGroupTask(group *proto.NodeGroup, opt *cloudprovider.CommonOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildAddExternalNodeToCluster add external nodes task
func (t *Task) BuildAddExternalNodeToCluster(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.AddExternalNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildDeleteExternalNodeFromCluster delete external node task
func (t *Task) BuildDeleteExternalNodeFromCluster(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.DeleteExternalNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildSwitchClusterNetworkTask switch cluster network mode
func (t *Task) BuildSwitchClusterNetworkTask(cls *proto.Cluster,
	subnet *proto.SubnetSource, opt *cloudprovider.SwitchClusterNetworkOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
