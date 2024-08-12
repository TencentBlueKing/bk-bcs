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

package blueking

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/blueking/tasks"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/template"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
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
	task.works[importClusterNodesStep.StepMethod] = tasks.ImportClusterNodesTask

	// create cluster task
	task.works[updateCreateClusterDBInfoStep.StepMethod] = tasks.UpdateCreateClusterDBInfoTask

	// delete cluster task
	task.works[cleanClusterDBInfoStep.StepMethod] = tasks.CleanClusterDBInfoTask

	// add node to cluster
	task.works[updateAddNodeDBInfoStep.StepMethod] = tasks.UpdateAddNodeDBInfoTask

	// remove node from cluster
	task.works[updateRemoveNodeDBInfoStep.StepMethod] = tasks.UpdateRemoveNodeDBInfoTask

	return task
}

// Task background task manager
type Task struct {
	works map[string]interface{}
}

// Name get task cloudName
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

	taskName := fmt.Sprintf(createClusterTaskTemplate, cls.ClusterID)
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// step1: call bkops preAction operation 创建集群
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil && opt.Cloud.ClusterManagement.CreateCluster != nil {
		step := &template.BkSopsStepAction{
			TaskName: template.SystemInit,
			Actions:  opt.Cloud.ClusterManagement.CreateCluster.PreActions,
			Plugins:  opt.Cloud.ClusterManagement.CreateCluster.Plugins,
		}
		err := step.BuildBkSopsStepAction(task, cls, template.ExtraInfo{
			NodeIPList:      strings.Join(opt.WorkerNodes, ","),
			BusinessID:      cls.BusinessID,
			NodeOperator:    opt.Operator,
			Operator:        opt.Operator,
			TranslateMethod: createClusterStep.StepMethod,
		})
		if err != nil {
			return nil, fmt.Errorf("BuildCreateClusterTask BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step2: call bksops add nodes to cluster 上架节点
	if len(opt.WorkerNodes) > 0 && opt.Cloud != nil && opt.Cloud.ClusterManagement != nil &&
		opt.Cloud.ClusterManagement.AddNodesToCluster != nil {
		step := &template.BkSopsStepAction{
			TaskName: template.SystemInit,
			Actions:  opt.Cloud.ClusterManagement.AddNodesToCluster.PreActions,
			Plugins:  opt.Cloud.ClusterManagement.AddNodesToCluster.Plugins,
		}
		err := step.BuildBkSopsStepAction(task, cls, template.ExtraInfo{
			NodeIPList:      strings.Join(opt.WorkerNodes, ","),
			NodeOperator:    opt.Operator,
			BusinessID:      cls.BusinessID,
			Operator:        opt.Operator,
			TranslateMethod: addNodesToClusterStep.StepMethod,
		})
		if err != nil {
			return nil, fmt.Errorf("BuildCreateClusterTask BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step3: update cluster DB info and associated data
	createClusterTask := &CreateClusterTaskOption{Cluster: cls, WorkerNodes: opt.WorkerNodes}
	createClusterTask.BuildUpdateClusterDbInfoStep(task)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildCreateClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.CreateClusterJob.String()

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

	// setting all steps details
	// step0: create cluster shield alarm step
	importNodesTask := &ImportClusterTaskOption{Cluster: cls}
	importNodesTask.BuildImportClusterNodesStep(task)

	// step1: install cluster watch component
	if options.GetEditionInfo().IsCommunicationEdition() || options.GetEditionInfo().IsEnterpriseEdition() {
		common.BuildWatchComponentTaskStep(task, cls, "")
	}
	// step2: install image pull secret addon if config
	common.BuildInstallImageSecretAddonTaskStep(task, cls)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildImportClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.OperatorKey.String()] = opt.Operator
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.ImportClusterJob.String()

	return task, nil
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

	// step1: call bkops operation preAction
	// validate bkops config
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil && opt.Cloud.ClusterManagement.DeleteCluster != nil {
		step := &template.BkSopsStepAction{
			TaskName: template.SystemInit,
			Actions:  opt.Cloud.ClusterManagement.DeleteCluster.PreActions,
			Plugins:  opt.Cloud.ClusterManagement.DeleteCluster.Plugins,
		}
		err := step.BuildBkSopsStepAction(task, cls, template.ExtraInfo{
			BusinessID:      cls.BusinessID,
			Operator:        opt.Operator,
			TranslateMethod: deleteClusterStep.StepMethod,
		})
		if err != nil {
			return nil, fmt.Errorf("BuildDeleteClusterTask BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step2: update cluster DB info and associated data
	cleanClusterTask := &DeleteClusterTaskOption{Cluster: cls}
	cleanClusterTask.BuildCleanClusterDbInfoStep(task)

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
		TaskName:       cloudprovider.AddNodesToClusterTask.String(),
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
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// step1: call bkops operation
	// preAction bcs sops task
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil && opt.Cloud.ClusterManagement.AddNodesToCluster != nil {
		step := &template.BkSopsStepAction{
			TaskName: template.SystemInit,
			Actions:  opt.Cloud.ClusterManagement.AddNodesToCluster.PreActions,
			Plugins:  opt.Cloud.ClusterManagement.AddNodesToCluster.Plugins,
		}
		err := step.BuildBkSopsStepAction(task, cls, template.ExtraInfo{
			NodeIPList:      strings.Join(nodeIPs, ","),
			NodeOperator:    opt.Operator,
			BusinessID:      cls.BusinessID,
			Operator:        opt.Operator,
			TranslateMethod: addNodesToClusterStep.StepMethod,
		})
		if err != nil {
			return nil, fmt.Errorf("BuildAddNodesToClusterTask BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step2: update DB node info by instanceIP
	addNodesTask := AddNodesTaskOption{
		Cluster: cls,
		Cloud:   opt.Cloud,
		NodeIps: nodeIPs,
	}
	addNodesTask.BuildUpdateAddNodeDbInfoStep(task)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildAddNodesToClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.OperatorKey.String()] = opt.Operator
	task.CommonParams[cloudprovider.UserKey.String()] = opt.Operator

	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.AddNodeJob.String()
	task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")

	return task, nil
}

// BuildRemoveNodesFromClusterTask build removeNodes task
func (t *Task) BuildRemoveNodesFromClusterTask(cls *proto.Cluster, nodes []*proto.Node,
	opt *cloudprovider.DeleteNodesOption) (*proto.Task, error) {
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

	// init task information
	nowStr := time.Now().Format(time.RFC3339)
	task := &proto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.RemoveNodesFromCluster),
		TaskName:       cloudprovider.RemoveNodesFromClusterTask.String(),
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
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// step1: build bkops task
	// validate bkops config
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil &&
		opt.Cloud.ClusterManagement.DeleteNodesFromCluster != nil {
		step := &template.BkSopsStepAction{
			TaskName: template.SystemInit,
			Actions:  opt.Cloud.ClusterManagement.DeleteNodesFromCluster.PreActions,
			Plugins:  opt.Cloud.ClusterManagement.DeleteNodesFromCluster.Plugins,
		}
		err := step.BuildBkSopsStepAction(task, cls, template.ExtraInfo{
			NodeIPList:      strings.Join(nodeIPs, ","),
			NodeOperator:    opt.Operator,
			BusinessID:      cls.BusinessID,
			Operator:        opt.Operator,
			TranslateMethod: removeNodesFromClusterStep.StepMethod,
		})
		if err != nil {
			return nil, fmt.Errorf("BuildAddNodesToClusterTask BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step2: update node DB info
	removeNodesTask := &RemoveNodesTaskOption{Cluster: cls, NodeIps: nodeIPs}
	removeNodesTask.BuildUpdateRemoveNodeDbInfoStep(task)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildRemoveNodesFromClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]

	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.DeleteNodeJob.String()
	task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")

	return task, nil
}

// BuildCleanNodesInGroupTask clean specified nodes in NodeGroup
// including remove nodes from NodeGroup, clean data in nodes
func (t *Task) BuildCleanNodesInGroupTask(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.CleanNodesOption) (*proto.Task, error) {
	// build task step1: move nodes out of nodegroup
	//step2: delete nodes in cluster
	//step3: delete nodes record in local storage
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildDeleteNodeGroupTask when delete nodegroup, we need to create background
// task to clean all nodes in nodegroup, release all resource in cloudprovider,
// finnally delete nodes information in local storage.
// @param group: need to delete
func (t *Task) BuildDeleteNodeGroupTask(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.DeleteNodeGroupOption) (*proto.Task, error) {
	return nil, nil
}

// BuildCreateNodeGroupTask build create node group task
func (t *Task) BuildCreateNodeGroupTask(group *proto.NodeGroup, opt *cloudprovider.CreateNodeGroupOption) (
	*proto.Task, error) {
	return nil, nil
}

// BuildUpdateNodeGroupTask when update nodegroup, we need to create background task,
func (t *Task) BuildUpdateNodeGroupTask(group *proto.NodeGroup, opt *cloudprovider.CommonOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildMoveNodesToGroupTask when create cluster, we need to create background task,
func (t *Task) BuildMoveNodesToGroupTask(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.MoveNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildUpdateDesiredNodesTask build update desired nodes task
func (t *Task) BuildUpdateDesiredNodesTask(desired uint32, group *proto.NodeGroup,
	opt *cloudprovider.UpdateDesiredNodeOption) (*proto.Task, error) {
	return nil, nil
}

// BuildSwitchNodeGroupAutoScalingTask switch nodegroup auto scaling
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

// BuildSwitchClusterNetworkTask switch cluster network mode
func (t *Task) BuildSwitchClusterNetworkTask(cls *proto.Cluster,
	subnet *proto.SubnetSource, opt *cloudprovider.SwitchClusterNetworkOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
