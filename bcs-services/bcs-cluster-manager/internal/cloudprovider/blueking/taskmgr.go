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

// Package blueking xxx
package blueking

import (
	"fmt"
	"strconv"
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

	// create resource pool
	task.works[createNodePoolStep.StepMethod] = tasks.CreateNodePoolTask

	// delete resource pool
	task.works[deleteNodePoolStep.StepMethod] = tasks.DeleteNodePoolTask

	// apply nodes from resource pool
	task.works[applyNodesFromResourcePoolStep.StepMethod] = tasks.ApplyNodesFromResourcePoolTask

	// return nodes to resource pool
	task.works[returnNodesToResourcePoolStep.StepMethod] = tasks.ReturnNodesToResourcePoolTask

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
func (t *Task) BuildCreateClusterTask(cls *proto.Cluster, opt *cloudprovider.CreateClusterOption) ( // nolint
	*proto.Task, error) {
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
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil &&
		opt.Cloud.ClusterManagement.CreateCluster != nil {
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
	taskName := fmt.Sprintf(importClusterTaskTemplate, cls.ClusterID)
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// setting all steps details
	importNodesTask := &ImportClusterTaskOption{Cluster: cls}

	// step0: 集群节点导入方式需要 安装websocket模式的kubeAgent，打通集群连接
	if len(opt.CloudMode.GetNodeIps()) > 0 && opt.Cloud != nil && opt.Cloud.ClusterManagement != nil &&
		opt.Cloud.ClusterManagement.ImportCluster != nil {
		err := template.BuildSopsFactory{
			StepName: template.SystemInit,
			Cluster:  cls,
			Extra: template.ExtraInfo{
				BusinessID:   cls.BusinessID,
				Operator:     opt.Operator,
				NodeOperator: opt.Operator,
				NodeIPList:   strings.Join(opt.CloudMode.GetNodeIps(), ","),
			}}.BuildSopsStep(task, opt.Cloud.ClusterManagement.ImportCluster, false)
		if err != nil {
			return nil, fmt.Errorf("BuildImportClusterTask BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step0: import cluster nodes
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

	// step1: call bkops operation preAction to delete cluster
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

	// step2: clean master nodes
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil &&
		opt.Cloud.ClusterManagement.DeleteNodesFromCluster != nil {
		step := &template.BkSopsStepAction{
			TaskName: template.SystemInit,
			Actions:  opt.Cloud.ClusterManagement.DeleteNodesFromCluster.PreActions,
			Plugins:  opt.Cloud.ClusterManagement.DeleteNodesFromCluster.Plugins,
		}
		err := step.BuildBkSopsStepAction(task, cls, template.ExtraInfo{
			NodeIPList:      strings.Join(cloudprovider.GetClusterMasterIPList(cls), ","),
			NodeOperator:    opt.Operator,
			BusinessID:      cls.BusinessID,
			Operator:        opt.Operator,
			TranslateMethod: removeNodesFromClusterStep.StepMethod,
		})
		if err != nil {
			return nil, fmt.Errorf("BuildRemoveNodesFromClusterTask BuildBkSopsStepAction failed: %v", err)
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
// nolint funlen
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

	addNodesTask := AddNodesTaskOption{
		Cluster:      cls,
		Cloud:        opt.Cloud,
		NodeIps:      nodeIPs,
		NodeTemplate: opt.NodeTemplate,
	}

	// step1: call bkops operation
	// preAction bcs sops task
	if err := t.addNodesSystemInitBkSops(task, cls, nodeIPs, opt); err != nil {
		return nil, err
	}

	// step2: 业务后置自定义流程: 支持标准运维任务 或者 后置脚本
	t.addNodesExecUserScript(task, cls, nodeIPs, opt)

	// business post define sops task or script
	if err := t.addNodesUserPostBkSops(task, cls, nodeIPs, opt); err != nil {
		return nil, err
	}

	// step3: 混部集群需要执行混部节点流程
	if err := t.addNodesMixedClsBkSops(task, cls, nodeIPs, opt); err != nil {
		return nil, err
	}

	// step4: 设置节点标签
	addNodesTask.BuildNodeLabelsStep(task)
	// step5: 设置节点污点
	addNodesTask.BuildNodeTaintsTaskStep(task)
	// step6: 设置节点注解
	addNodesTask.BuildNodeAnnotationsStep(task)

	// step7: update DB node info by instanceIP
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

// addNodesSystemInitBkSops 构建系统初始化步骤
func (t *Task) addNodesSystemInitBkSops(
	task *proto.Task, cls *proto.Cluster, nodeIPs []string, opt *cloudprovider.AddNodesOption) error {
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil &&
		opt.Cloud.ClusterManagement.AddNodesToCluster != nil && !opt.NodeTemplate.GetSkipSystemInit() {
		step := &template.BkSopsStepAction{
			TaskName: template.SystemInit,
			Actions:  opt.Cloud.ClusterManagement.AddNodesToCluster.PreActions,
			Plugins:  opt.Cloud.ClusterManagement.AddNodesToCluster.Plugins,
		}
		err := step.BuildBkSopsStepAction(task, cls, template.ExtraInfo{
			NodeIPList:      strings.Join(nodeIPs, ","),
			NodeOperator:    opt.Operator,
			ModuleID:        cloudprovider.GetScaleOutModuleID(cls, nil, opt.NodeTemplate, false),
			BusinessID:      cloudprovider.GetBusinessID(cls, nil, opt.NodeTemplate, true),
			Operator:        opt.Operator,
			TranslateMethod: addNodesToClusterStep.StepMethod,
		})
		if err != nil {
			return fmt.Errorf("BuildAddNodesToClusterTask BuildBkSopsStepAction failed: %v", err)
		}
	}
	return nil
}

// addNodesExecUserScript 构建用户脚本步骤
func (t *Task) addNodesExecUserScript(
	task *proto.Task, cls *proto.Cluster, nodeIPs []string, opt *cloudprovider.AddNodesOption) {
	if opt.NodeTemplate != nil && len(opt.NodeTemplate.UserScript) > 0 {
		common.BuildJobExecuteScriptStep(task, common.JobExecParas{
			ClusterID:        cls.ClusterID,
			Content:          opt.NodeTemplate.UserScript,
			NodeIps:          strings.Join(nodeIPs, ","),
			Operator:         opt.Operator,
			StepName:         common.PostInitStepJob,
			AllowSkipJobTask: opt.NodeTemplate.AllowSkipScaleOutWhenFailed,
			Translate:        common.PostInitJob,
		})
	}
}

// addNodesUserPostBkSops 构建用户后置步骤
func (t *Task) addNodesUserPostBkSops(
	task *proto.Task, cls *proto.Cluster, nodeIPs []string, opt *cloudprovider.AddNodesOption) error {
	if opt.NodeTemplate != nil && opt.NodeTemplate.ScaleOutExtraAddons != nil {
		err := template.BuildSopsFactory{
			StepName: template.UserAfterInit,
			Cluster:  cls,
			Extra: template.ExtraInfo{
				NodeIPList:      strings.Join(nodeIPs, ","),
				NodeOperator:    opt.Operator,
				ShowSopsUrl:     true,
				TranslateMethod: template.UserPostInit,
			}}.BuildSopsStep(task, opt.NodeTemplate.ScaleOutExtraAddons, false)
		if err != nil {
			return fmt.Errorf("BuildScalingNodesTask business BuildBkSopsStepAction failed: %v", err)
		}
	}
	return nil
}

// addNodesMixedClsBkSops 构建混部集群步骤
func (t *Task) addNodesMixedClsBkSops(
	task *proto.Task, cls *proto.Cluster, nodeIPs []string, opt *cloudprovider.AddNodesOption) error {
	if cls.GetIsMixed() && opt.Cloud != nil && opt.Cloud.ClusterManagement != nil &&
		opt.Cloud.ClusterManagement.CommonMixedAction != nil {
		err := template.BuildSopsFactory{
			StepName: template.NodeMixedInitCh,
			Cluster:  cls,
			Extra: template.ExtraInfo{
				NodeIPList:      strings.Join(nodeIPs, ","),
				ShowSopsUrl:     true,
				TranslateMethod: template.NodeMixedInit,
			}}.BuildSopsStep(task, opt.Cloud.ClusterManagement.CommonMixedAction, false)
		if err != nil {
			return fmt.Errorf("BuildScalingNodesTask BuildBkSopsStepAction failed: %v", err)
		}
	}
	return nil
}

// BuildRemoveNodesFromClusterTask build removeNodes task
// nolint: funlen
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

	// step1: 业务自定义缩容流程: 支持 缩容节点前置脚本和前置标准运维流程
	if opt.NodeTemplate != nil && len(opt.NodeTemplate.ScaleInPreScript) > 0 {
		common.BuildJobExecuteScriptStep(task, common.JobExecParas{
			ClusterID:        cls.ClusterID,
			Content:          opt.NodeTemplate.ScaleInPreScript,
			NodeIps:          strings.Join(nodeIPs, ","),
			Operator:         opt.Operator,
			StepName:         common.PreInitStepJob,
			AllowSkipJobTask: opt.NodeTemplate.AllowSkipScaleInWhenFailed,
			Translate:        common.PreInitJob,
		})
	}

	// business define sops task
	if opt.NodeTemplate != nil && opt.NodeTemplate.ScaleInExtraAddons != nil {
		err := template.BuildSopsFactory{
			StepName: template.UserPreInit,
			Cluster:  cls,
			Extra: template.ExtraInfo{
				NodeIPList:      strings.Join(nodeIPs, ","),
				NodeOperator:    opt.Operator,
				ShowSopsUrl:     true,
				TranslateMethod: template.UserBeforeInit,
			}}.BuildSopsStep(task, opt.NodeTemplate.ScaleInExtraAddons, true)
		if err != nil {
			return nil, fmt.Errorf("BuildRemoveNodesFromClusterTask business "+
				"BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step2: build bkops task
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
			BusinessID:      cloudprovider.GetBusinessID(cls, nil, opt.NodeTemplate, false),
			ModuleID:        cloudprovider.GetScaleInModuleID(nil, opt.NodeTemplate),
			Operator:        opt.Operator,
			TranslateMethod: removeNodesFromClusterStep.StepMethod,
		})
		if err != nil {
			return nil, fmt.Errorf("BuildRemoveNodesFromClusterTask BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step3: update node DB info
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
// nolint funlen
func (t *Task) BuildCleanNodesInGroupTask(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.CleanNodesOption) (*proto.Task, error) {
	// clean nodeGroup nodes in cloud only has four steps:
	// step1: cordon nodes
	// step2: user define processes
	// step3: delete nodes in cluster
	// step4: return nodes to resource pool

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
		nodeIPs            = make([]string, 0)
		nodeIDs, deviceIDs = make([]string, 0), make([]string, 0)
	)
	for _, node := range nodes {
		nodeIPs = append(nodeIPs, node.InnerIP)
		nodeIDs = append(nodeIDs, node.NodeID)
		deviceIDs = append(deviceIDs, node.DeviceID)
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
		NodeIPList:     nodeIPs,
	}
	// generate taskName
	taskName := fmt.Sprintf(cleanNodeGroupNodesTaskTemplate, group.ClusterID, group.Name)
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// setting all steps details
	cleanNodeGroupNodes := &CleanNodesInGroupTaskOption{
		Group:     group,
		Cluster:   opt.Cluster,
		NodeIPs:   nodeIPs,
		NodeIDs:   nodeIDs,
		DeviceIDs: deviceIDs,
		Operator:  opt.Operator,
	}

	// step1: cordon nodes
	common.BuildCordonNodesTaskStep(task, opt.Cluster.GetClusterID(), nodeIPs)

	// step2: business user define processes, bksops or job
	if group.NodeTemplate != nil && len(group.NodeTemplate.ScaleInPreScript) > 0 {
		common.BuildJobExecuteScriptStep(task, common.JobExecParas{
			ClusterID:        opt.Cluster.ClusterID,
			Content:          group.NodeTemplate.ScaleInPreScript,
			NodeIps:          strings.Join(nodeIPs, ","),
			Operator:         opt.Operator,
			StepName:         common.PreInitStepJob,
			AllowSkipJobTask: group.NodeTemplate.AllowSkipScaleInWhenFailed,
			Translate:        common.PreInitJob,
		})
	}
	if group.NodeTemplate != nil && group.NodeTemplate.ScaleInExtraAddons != nil &&
		len(group.NodeTemplate.ScaleInExtraAddons.PreActions) > 0 {
		err := template.BuildSopsFactory{
			StepName: template.UserPreInit,
			Cluster:  opt.Cluster,
			Extra: template.ExtraInfo{
				NodeIPList:      strings.Join(nodeIPs, ","),
				NodeOperator:    opt.Operator,
				ShowSopsUrl:     true,
				TranslateMethod: template.UserBeforeInit,
			}}.BuildSopsStep(task, group.NodeTemplate.ScaleInExtraAddons, true)
		if err != nil {
			return nil, fmt.Errorf("BuildCleanNodesInGroupTask ScaleInExtraAddons.PreActions "+
				"BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step3: build sops task to delete node from cluster
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil &&
		opt.Cloud.ClusterManagement.DeleteNodesFromCluster != nil {
		step := &template.BkSopsStepAction{
			TaskName: template.SystemInit,
			Actions:  opt.Cloud.ClusterManagement.DeleteNodesFromCluster.PreActions,
			Plugins:  opt.Cloud.ClusterManagement.DeleteNodesFromCluster.Plugins,
		}
		err := step.BuildBkSopsStepAction(task, opt.Cluster, template.ExtraInfo{
			NodeIPList:      strings.Join(nodeIPs, ","),
			NodeOperator:    opt.Operator,
			BusinessID:      cloudprovider.GetBusinessID(opt.Cluster, opt.AsOption, group.NodeTemplate, false),
			ModuleID:        cloudprovider.GetScaleInModuleID(opt.AsOption, group.NodeTemplate),
			Operator:        opt.Operator,
			TranslateMethod: removeNodesFromClusterStep.StepMethod,
		})
		if err != nil {
			return nil, fmt.Errorf("BuildCleanNodesInGroupTask BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step4: return nodes to resource pool
	cleanNodeGroupNodes.BuildReturnNodesToResourcePoolStep(task)

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

	if opt == nil || len(opt.Operator) == 0 || opt.Cloud == nil || opt.Cluster == nil {
		return nil, fmt.Errorf("BuildDeleteNodeGroupTask TaskOptions is lost")
	}

	var (
		nodeIPs, nodeIDs, deviceIDs = make([]string, 0), make([]string, 0), make([]string, 0)
	)
	for _, node := range nodes {
		nodeIPs = append(nodeIPs, node.InnerIP)
		nodeIDs = append(nodeIDs, node.NodeID)
		deviceIDs = append(deviceIDs, node.DeviceID) // nolint staticcheck (this result of append is never used)
	}

	// init task information
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
		Creator:        opt.Operator,
		Updater:        opt.Operator,
		LastUpdate:     nowStr,
		CommonParams:   make(map[string]string),
		ForceTerminate: false,
		NodeGroupID:    group.NodeGroupID,
	}
	// generate taskName
	taskName := fmt.Sprintf(deleteNodeGroupTaskTemplate, group.ClusterID, group.Name)
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	deleteNodeGroupOption := &DeleteNodeGroupOption{
		NodeGroup: group,
		Cluster:   opt.Cluster,
		NodeIDs:   nodeIDs,
		NodeIPs:   nodeIPs,
		Operator:  opt.Operator,
	}
	// step1: previous call successful and delete local storage information
	deleteNodeGroupOption.BuildDeleteCloudNodeGroupStep(task)
	// step2: delete nodeGroup from CA
	common.BuildEnsureAutoScalerTaskStep(task, group.ClusterID, group.Provider)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildDeleteNodeGroupTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]

	// Job-type
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.DeleteNodeGroupJob.String()
	task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")
	task.CommonParams[cloudprovider.NodeIDsKey.String()] = strings.Join(nodeIDs, ",")

	return task, nil
}

// BuildCreateNodeGroupTask build create node group task
func (t *Task) BuildCreateNodeGroupTask(group *proto.NodeGroup, opt *cloudprovider.CreateNodeGroupOption) (
	*proto.Task, error) {
	// bluekingCloud create nodeGroup steps
	// step1: create underlying resourcePool and update nodeGroup relative info
	// step2: deploy node group to cluster

	// validate request params
	if group == nil {
		return nil, fmt.Errorf("BuildCreateNodeGroupTask group info empty")
	}
	if opt == nil || opt.Cluster == nil {
		return nil, fmt.Errorf("BuildCreateNodeGroupTask TaskOptions is lost option or cluster")
	}
	err := opt.PoolInfo.Validate()
	if err != nil {
		return nil, err
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
	createNodeGroupTask := &CreateNodeGroupOption{NodeGroup: group,
		Cluster: opt.Cluster, PoolProvider: opt.PoolInfo.Provider, PoolID: opt.PoolInfo.ResourcePoolID}
	// step1. call self resourcePool to create node group
	createNodeGroupTask.BuildCreateCloudNodeGroupStep(task)
	// step2. ensure autoscaler(安装/更新CA组件) in cluster
	common.BuildEnsureAutoScalerTaskStep(task, group.ClusterID, group.Provider)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildCreateNodeGroupTask task StepSequence empty")
	}

	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.CreateNodeGroupJob.String()

	return task, nil
}

// BuildUpdateNodeGroupTask when update nodegroup, we need to create background task,
func (t *Task) BuildUpdateNodeGroupTask(group *proto.NodeGroup, opt *cloudprovider.CommonOption) (*proto.Task, error) {
	// validate request params
	if group == nil {
		return nil, fmt.Errorf("BuildUpdateNodeGroupTask group info empty")
	}
	if opt == nil {
		return nil, fmt.Errorf("BuildUpdateNodeGroupTask TaskOptions is lost")
	}

	nowStr := time.Now().Format(time.RFC3339)
	task := &proto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.UpdateNodeGroup),
		TaskName:       cloudprovider.UpdateNodeGroupTask.String(),
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
	taskName := fmt.Sprintf(updateNodeGroupTaskTemplate, group.ClusterID, group.NodeGroupID)
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// setting all steps details
	// step1. ensure auto scaler
	common.BuildEnsureAutoScalerTaskStep(task, group.ClusterID, group.Provider)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildUpdateNodeGroupTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.UpdateNodeGroupJob.String()
	return task, nil
}

// BuildMoveNodesToGroupTask when create cluster, we need to create background task,
func (t *Task) BuildMoveNodesToGroupTask(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.MoveNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildUpdateDesiredNodesTask build update desired nodes task
// nolint funlen
func (t *Task) BuildUpdateDesiredNodesTask(desired uint32, group *proto.NodeGroup,
	opt *cloudprovider.UpdateDesiredNodeOption) (*proto.Task, error) {
	// UpdateDesiredNodesTask has five steps:
	// 1. call resource interface to apply for Instances
	// 2. call sops task to add nodes to cluster
	// 3. add annotation to nodes
	// 4. add labels to nodes
	// 5. unCordon nodes

	// validate request params
	if desired == 0 {
		return nil, fmt.Errorf("BuildUpdateDesiredNodesTask desired nodes is zero")
	}
	if group == nil {
		return nil, fmt.Errorf("BuildUpdateDesiredNodesTask group info is empty")
	}
	if opt == nil || len(opt.Operator) == 0 || opt.Cluster == nil {
		return nil, fmt.Errorf("BuildUpdateDesiredNodesTask TaskOptions is lost")
	}

	// init task information
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
		Creator:        opt.Operator,
		Updater:        opt.Operator,
		LastUpdate:     nowStr,
		CommonParams:   make(map[string]string),
		ForceTerminate: false,
		NodeGroupID:    group.NodeGroupID,
	}

	// generate taskName
	taskName := fmt.Sprintf(updateNodeGroupDesiredNodeTemplate, group.ClusterID, group.Name)
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// setting all steps details
	updateDesiredNodes := &UpdateDesiredNodesTaskOption{
		Group:    group,
		Cluster:  opt.Cluster,
		Cloud:    opt.Cloud,
		Desired:  desired,
		Operator: opt.Operator,
	}

	// step1: apply for Instance from resourcePool
	updateDesiredNodes.BuildApplyInstanceStep(task)

	// 检测是否存在bkcc/检测是否安装agent/空闲检查等流程

	// step2: call sops task to add nodes to cluster: install agent & trans module
	if err := t.updateDesiredNodesSysInitBkSops(task, group, opt); err != nil {
		return nil, err
	}

	// transfer host module
	moduleID := cloudprovider.GetTransModuleInfo(opt.Cluster, opt.AsOption, opt.NodeGroup)
	if moduleID != "" {
		common.BuildTransferHostModuleStep(task, opt.Cluster.BusinessID, moduleID, "")
	}

	// step3: 业务扩容节点后置自定义流程: 支持job后置脚本和标准运维任务
	t.updateDesiredNodesUserScript(task, group, opt)

	// business define sops task
	if err := t.updateDesiredNodesUserPostBkSops(task, group, opt); err != nil {
		return nil, err
	}

	// 混部集群需要执行混部节点流程
	if err := t.updateDesiredNodesMixedClsBkSops(task, opt); err != nil {
		return nil, err
	}

	// step4: annotation nodes
	updateDesiredNodes.BuildNodeAnnotationsStep(task)
	// step5: nodes common labels: sZoneID / bizID
	updateDesiredNodes.BuildNodeCommonLabelsStep(task)
	// set node taint
	common.BuildNodeTaintsTaskStep(task, opt.Cluster.ClusterID, nil, cloudprovider.GetTaintsByNg(opt.NodeGroup))
	// step6: set resourcePool labels
	updateDesiredNodes.BuildResourcePoolDeviceLabelStep(task)
	// step7: unCordon nodes
	updateDesiredNodes.BuildUnCordonNodesStep(task)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildUpdateDesiredNodesTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]

	// set common parameters && JobType
	task.CommonParams[cloudprovider.ClusterIDKey.String()] = group.ClusterID
	task.CommonParams[cloudprovider.ScalingNodesNumKey.String()] = strconv.Itoa(int(desired))
	// Job-type
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.UpdateNodeGroupDesiredNodeJob.String()
	task.CommonParams[cloudprovider.ManualKey.String()] = strconv.FormatBool(opt.Manual)

	return task, nil
}

// updateDesiredNodesSysInitBkSops 构建更新期望节点数的系统初始化步骤
func (t *Task) updateDesiredNodesSysInitBkSops(
	task *proto.Task, group *proto.NodeGroup, opt *cloudprovider.UpdateDesiredNodeOption) error {
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil &&
		opt.Cloud.ClusterManagement.AddNodesToCluster != nil && !group.GetNodeTemplate().GetSkipSystemInit() {
		step := &template.BkSopsStepAction{
			TaskName: template.SystemInit,
			Actions:  opt.Cloud.ClusterManagement.AddNodesToCluster.PreActions,
			Plugins:  opt.Cloud.ClusterManagement.AddNodesToCluster.Plugins,
		}
		err := step.BuildBkSopsStepAction(task, opt.Cluster, template.ExtraInfo{
			NodeIPList:   "",
			NodeOperator: opt.Operator,
			BusinessID:   cloudprovider.GetBusinessID(opt.Cluster, opt.AsOption, group.NodeTemplate, true),
			ModuleID: cloudprovider.GetScaleOutModuleID(opt.Cluster, opt.AsOption, group.NodeTemplate,
				true),
			Operator:        opt.Operator,
			TranslateMethod: addNodesToClusterStep.StepMethod,
		})
		if err != nil {
			return fmt.Errorf("BuildUpdateDesiredNodesTask BuildBkSopsStepAction failed: %v", err)
		}
	}
	return nil
}

// updateDesiredNodesUserScript 构建更新期望节点数的用户脚本步骤
func (t *Task) updateDesiredNodesUserScript(
	task *proto.Task, group *proto.NodeGroup, opt *cloudprovider.UpdateDesiredNodeOption) {
	if group.NodeTemplate != nil && len(group.NodeTemplate.UserScript) > 0 {
		common.BuildJobExecuteScriptStep(task, common.JobExecParas{
			ClusterID:        group.ClusterID,
			Content:          group.NodeTemplate.UserScript,
			NodeIps:          "",
			Operator:         opt.Operator,
			StepName:         common.PostInitStepJob,
			AllowSkipJobTask: group.NodeTemplate.GetAllowSkipScaleOutWhenFailed(),
			Translate:        common.PostInitJob,
		})
	}
}

// updateDesiredNodesUserPostBkSops 构建更新期望节点数的用户后置步骤
func (t *Task) updateDesiredNodesUserPostBkSops(
	task *proto.Task, group *proto.NodeGroup, opt *cloudprovider.UpdateDesiredNodeOption) error {
	if group.NodeTemplate != nil && group.NodeTemplate.ScaleOutExtraAddons != nil {
		err := template.BuildSopsFactory{
			StepName: template.UserAfterInit,
			Cluster:  opt.Cluster,
			Extra: template.ExtraInfo{
				NodeIPList:      "",
				NodeOperator:    opt.Operator,
				ShowSopsUrl:     true,
				TranslateMethod: template.UserPostInit,
			}}.BuildSopsStep(task, group.NodeTemplate.ScaleOutExtraAddons, false)
		if err != nil {
			return fmt.Errorf("BuildUpdateDesiredNodesTask business BuildBkSopsStepAction failed: %v", err)
		}
	}
	return nil
}

// updateDesiredNodesMixedClsBkSops 构建更新期望节点数的混部集群步骤
func (t *Task) updateDesiredNodesMixedClsBkSops(task *proto.Task, opt *cloudprovider.UpdateDesiredNodeOption) error {
	if opt.Cluster.GetIsMixed() && opt.Cloud != nil && opt.Cloud.ClusterManagement != nil &&
		opt.Cloud.ClusterManagement.CommonMixedAction != nil {
		err := template.BuildSopsFactory{
			StepName: template.NodeMixedInitCh,
			Cluster:  opt.Cluster,
			Extra: template.ExtraInfo{
				NodeIPList:      "",
				ShowSopsUrl:     true,
				TranslateMethod: template.NodeMixedInit,
			}}.BuildSopsStep(task, opt.Cloud.ClusterManagement.CommonMixedAction, false)
		if err != nil {
			return fmt.Errorf("BuildUpdateDesiredNodesTask BuildBkSopsStepAction failed: %v", err)
		}
	}
	return nil
}

// BuildSwitchNodeGroupAutoScalingTask switch nodegroup auto scaling
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
		TaskName:       cloudprovider.SwitchNodeGroupAutoScalingTask.String(),
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
	taskName := fmt.Sprintf(switchNodeGroupAutoScalingTaskTemplate, group.ClusterID, group.Name)
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

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

// BuildSwitchAsOptionStatusTask switch auto scaling option status
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
		TaskName:       cloudprovider.SwitchAutoScalingOptionStatusTask.String(),
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
