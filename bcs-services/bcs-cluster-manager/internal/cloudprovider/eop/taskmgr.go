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

package eop

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/eop/tasks"
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

	// create cluster task
	task.works[createECKClusterStep.StepMethod] = tasks.CreateECKClusterTask
	task.works[checkECKClusterStatusStep.StepMethod] = tasks.CheckECKClusterStatusTask
	task.works[checkECKNodeGroupsStatusStep.StepMethod] = tasks.CheckECKNodesGroupStatusTask
	task.works[updateECKNodeGroupsToDBStep.StepMethod] = tasks.UpdateECKNodesGroupToDBTask
	task.works[checkCreateClusterNodeStatusStep.StepMethod] = tasks.CheckECKClusterNodesStatusTask
	task.works[updateECKNodesToDBStep.StepMethod] = tasks.UpdateECKNodesToDBTask
	task.works[registerManageClusterKubeConfigStep.StepMethod] = tasks.RegisterClusterKubeConfigTask

	// delete cluster task
	task.works[deleteECKClusterStep.StepMethod] = tasks.DeleteECKClusterTask
	task.works[cleanClusterDBInfoStep.StepMethod] = tasks.CleanClusterDBInfoTask

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
func (t *Task) BuildCreateClusterTask(
	cls *proto.Cluster, opt *cloudprovider.CreateClusterOption) (*proto.Task, error) {
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
		TaskName:       cloudprovider.CreateClusterTask.String(),
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
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// setting all steps details
	createClusterTask := &CreateClusterTaskOption{Cluster: cls, NodeGroupIDs: opt.NodeGroupIDs}
	// step0: createCluster and return clusterID inject common paras
	createClusterTask.BuildCreateClusterStep(task)
	// step1: check cluster status by clusterID
	createClusterTask.BuildCheckClusterStatusStep(task)
	// step2: check cluster nodes status
	createClusterTask.BuildCheckNodeGroupsStatusStep(task)
	// step3: update nodegroups to DB
	createClusterTask.BuildUpdateNodeGroupsToDBStep(task)
	// step4: check cluster nodegroups status
	createClusterTask.BuildCheckClusterNodesStatusStep(task)
	// step5: update nodes to DB
	createClusterTask.BuildUpdateNodesToDBStep(task)
	// step6: register managed cluster kubeConfig
	createClusterTask.BuildRegisterClsKubeConfigStep(task)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildCreateClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.OperatorKey.String()] = opt.Operator
	task.CommonParams[cloudprovider.UserKey.String()] = opt.Operator
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.CreateClusterJob.String()
	if len(opt.WorkerNodes) > 0 {
		task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(opt.WorkerNodes, ",")
	}

	return task, nil
}

// BuildImportClusterTask build import cluster task
func (t *Task) BuildImportClusterTask(
	cls *proto.Cluster, opt *cloudprovider.ImportClusterOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildDeleteClusterTask build deleteCluster task
func (t *Task) BuildDeleteClusterTask(
	cls *proto.Cluster, opt *cloudprovider.DeleteClusterOption) (*proto.Task, error) {
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
		TaskName:       cloudprovider.DeleteClusterTask.String(),
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

	// setting all steps details
	deleteCluster := &DeleteClusterTaskOption{
		Cluster: cls,
	}
	// step1: deleteECKClusterTask delete eck cluster
	deleteCluster.BuildDeleteECKClusterStep(task)
	// step2: update cluster DB info and associated data
	deleteCluster.BuildCleanClusterDBInfoStep(task)

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

// BuildUpdateNodeGroupTask when update nodegroup, we need to create background task,
func (t *Task) BuildUpdateNodeGroupTask(group *proto.NodeGroup, opt *cloudprovider.CommonOption) (*proto.Task, error) {
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

// BuildSwitchAsOptionStatusTask switch auto scaling option status
func (t *Task) BuildSwitchAsOptionStatusTask(scalingOption *proto.ClusterAutoScalingOption, enable bool,
	opt *cloudprovider.CommonOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildAddExternalNodeToCluster add external to cluster
func (t *Task) BuildAddExternalNodeToCluster(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.AddExternalNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildDeleteExternalNodeFromCluster remove external node from cluster
func (t *Task) BuildDeleteExternalNodeFromCluster(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.DeleteExternalNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
