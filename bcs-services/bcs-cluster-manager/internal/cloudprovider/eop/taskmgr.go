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

// BuildCreateClusterTask build create cluster task
func (t *Task) BuildCreateClusterTask(
	cls *proto.Cluster, opt *cloudprovider.CreateClusterOption) (*proto.Task, error) {
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
		TaskName:       "创建ECK集群",
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

	// setting all steps details
	// step1: createTKECluster and return clusterID inject common paras
	createStep := &proto.Step{
		Name:       createECKClusterStep.StepMethod,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Start:      nowStr,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: createECKClusterStep.StepMethod,
		TaskName:   createECKClusterStep.StepName,
	}
	createStep.Params["clusterID"] = cls.ClusterID
	createStep.Params["cloudID"] = cls.Provider
	createStep.Params[cloudprovider.NodeGroupIDKey.String()] = strings.Join(opt.NodeGroupIDs, ",")

	task.Steps[createECKClusterStep.StepMethod] = createStep
	task.StepSequence = append(task.StepSequence, createECKClusterStep.StepMethod)

	// step2: check cluster status by clusterID
	checkStep := &proto.Step{
		Name:       checkECKClusterStatusStep.StepMethod,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: checkECKClusterStatusStep.StepMethod,
		TaskName:   checkECKClusterStatusStep.StepName,
	}
	checkStep.Params["clusterID"] = cls.ClusterID
	checkStep.Params["cloudID"] = cls.Provider

	task.Steps[checkECKClusterStatusStep.StepMethod] = checkStep
	task.StepSequence = append(task.StepSequence, checkECKClusterStatusStep.StepMethod)

	// step3: check cluster nodegroup status by clusterID
	checkNgStep := &proto.Step{
		Name:       checkECKNodeGroupsStatusStep.StepMethod,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: checkECKNodeGroupsStatusStep.StepMethod,
		TaskName:   checkECKNodeGroupsStatusStep.StepName,
	}
	checkNgStep.Params["clusterID"] = cls.ClusterID
	checkNgStep.Params["cloudID"] = cls.Provider
	checkNgStep.Params[cloudprovider.NodeGroupIDKey.String()] = strings.Join(opt.NodeGroupIDs, ",")

	task.Steps[checkECKNodeGroupsStatusStep.StepMethod] = checkNgStep
	task.StepSequence = append(task.StepSequence, checkECKNodeGroupsStatusStep.StepMethod)

	// step4: update cluster nodegroup info to DB
	updateNgStep := &proto.Step{
		Name:       updateECKNodeGroupsToDBStep.StepMethod,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: updateECKNodeGroupsToDBStep.StepMethod,
		TaskName:   updateECKNodeGroupsToDBStep.StepName,
	}
	updateNgStep.Params["clusterID"] = cls.ClusterID
	updateNgStep.Params["cloudID"] = cls.Provider

	task.Steps[updateECKNodeGroupsToDBStep.StepMethod] = updateNgStep
	task.StepSequence = append(task.StepSequence, updateECKNodeGroupsToDBStep.StepMethod)

	// step5: check cluster node status by clusterID
	checkNodeStep := &proto.Step{
		Name:       checkCreateClusterNodeStatusStep.StepMethod,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: checkCreateClusterNodeStatusStep.StepMethod,
		TaskName:   checkCreateClusterNodeStatusStep.StepName,
	}
	checkNodeStep.Params["clusterID"] = cls.ClusterID
	checkNodeStep.Params["cloudID"] = cls.Provider

	task.Steps[checkCreateClusterNodeStatusStep.StepMethod] = checkNodeStep
	task.StepSequence = append(task.StepSequence, checkCreateClusterNodeStatusStep.StepMethod)

	// step6: update cluster node info to DB
	updateNodeStep := &proto.Step{
		Name:       updateECKNodesToDBStep.StepMethod,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: updateECKNodesToDBStep.StepMethod,
		TaskName:   updateECKNodesToDBStep.StepName,
	}
	updateNodeStep.Params["clusterID"] = cls.ClusterID
	updateNodeStep.Params["cloudID"] = cls.Provider

	task.Steps[updateECKNodesToDBStep.StepMethod] = updateNodeStep
	task.StepSequence = append(task.StepSequence, updateECKNodesToDBStep.StepMethod)

	// step7: update cluster node info to DB
	registerStep := &proto.Step{
		Name:       registerManageClusterKubeConfigStep.StepMethod,
		System:     "api",
		Params:     make(map[string]string),
		Retry:      0,
		Status:     cloudprovider.TaskStatusNotStarted,
		TaskMethod: registerManageClusterKubeConfigStep.StepMethod,
		TaskName:   registerManageClusterKubeConfigStep.StepName,
	}
	registerStep.Params["clusterID"] = cls.ClusterID
	registerStep.Params["cloudID"] = cls.Provider

	task.Steps[registerManageClusterKubeConfigStep.StepMethod] = registerStep
	task.StepSequence = append(task.StepSequence, registerManageClusterKubeConfigStep.StepMethod)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildCreateClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.OperatorKey.String()] = opt.Operator
	task.CommonParams["user"] = opt.Operator
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.CreateClusterJob.String()

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

// BuildSwitchAutoScalingOptionStatusTask build switch auto scaler option status task
func (t *Task) BuildSwitchAutoScalingOptionStatusTask(scalingOption *proto.ClusterAutoScalingOption, enable bool,
	opt *cloudprovider.CommonOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
