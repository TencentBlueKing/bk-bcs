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

package qcloud

import (
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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
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
	task.works[importClusterNodesStep.StepMethod] = tasks.ImportClusterNodesTask
	task.works[registerClusterKubeConfigStep.StepMethod] = tasks.RegisterClusterKubeConfigTask

	// create cluster task
	task.works[createClusterShieldAlarmStep.StepMethod] = tasks.CreateClusterShieldAlarmTask
	task.works[createTKEClusterStep.StepMethod] = tasks.CreateTkeClusterTask
	task.works[checkTKEClusterStatusStep.StepMethod] = tasks.CheckTkeClusterStatusTask
	task.works[enableTkeClusterVpcCniStep.StepMethod] = tasks.EnableTkeClusterVpcCniTask
	task.works[checkCreateClusterNodeStatusStep.StepMethod] = tasks.CheckCreateClusterNodeStatusTask
	task.works[registerManageClusterKubeConfigStep.StepMethod] = tasks.RegisterManageClusterKubeConfigTask
	task.works[updateCreateClusterDBInfoStep.StepMethod] = tasks.UpdateCreateClusterDBInfoTask

	// delete cluster task
	task.works[deleteTKEClusterStep.StepMethod] = tasks.DeleteTKEClusterTask
	task.works[cleanClusterDBInfoStep.StepMethod] = tasks.CleanClusterDBInfoTask

	// add node to cluster
	task.works[addNodesShieldAlarmStep.StepMethod] = tasks.AddNodesShieldAlarmTask
	task.works[addNodesToClusterStep.StepMethod] = tasks.AddNodesToClusterTask
	task.works[checkAddNodesStatusStep.StepMethod] = tasks.CheckAddNodesStatusTask
	task.works[updateAddNodeDBInfoStep.StepMethod] = tasks.UpdateNodeDBInfoTask

	// remove node from cluster
	task.works[removeNodesFromClusterStep.StepMethod] = tasks.RemoveNodesFromClusterTask
	task.works[updateRemoveNodeDBInfoStep.StepMethod] = tasks.UpdateRemoveNodeDBInfoTask

	// add external node to cluster
	task.works[getExternalNodeScriptStep.StepMethod] = tasks.GetExternalNodeScriptTask

	// remove external node from cluster
	task.works[removeExternalNodesFromClusterStep.StepMethod] = tasks.RemoveExternalNodesFromClusterTask

	// init qcloud node-group task

	// autoScaler task
	// task.works[ensureAutoScalerStep.StepMethod] = tasks.EnsureAutoScalerTask

	// create nodeGroup task
	task.works[createCloudNodeGroupStep.StepMethod] = tasks.CreateCloudNodeGroupTask
	task.works[checkCloudNodeGroupStatusStep.StepMethod] = tasks.CheckCloudNodeGroupStatusTask
	// task.works[updateCreateNodeGroupDBInfoTask] = tasks.UpdateCreateNodeGroupDBInfoTask

	// delete nodeGroup task
	task.works[deleteNodeGroupStep.StepMethod] = tasks.DeleteCloudNodeGroupTask
	// task.works[updateDeleteNodeGroupDBInfoTask] = tasks.UpdateDeleteNodeGroupDBInfoTask

	// clean node in nodeGroup task
	task.works[cleanNodeGroupNodesStep.StepMethod] = tasks.CleanNodeGroupNodesTask
	task.works[checkClusterCleanNodsStep.StepMethod] = tasks.CheckClusterCleanNodsTask
	task.works[returnIDCNodeToResourcePoolStep.StepMethod] = tasks.ReturnIDCNodeToResourcePoolTask
	// task.works[checkCleanNodeGroupNodesStatusTask] = tasks.CheckCleanNodeGroupNodesStatusTask
	// task.works[updateCleanNodeGroupNodesDBInfoTask] = tasks.UpdateCleanNodeGroupNodesDBInfoTask

	// update desired nodes task
	task.works[applyInstanceMachinesStep.StepMethod] = tasks.ApplyInstanceMachinesTask
	task.works[checkClusterNodesStatusStep.StepMethod] = tasks.CheckClusterNodesStatusTask

	task.works[applyExternalNodeMachinesStep.StepMethod] = tasks.ApplyExternalNodeMachinesTask
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
// NOCC:CCN_threshold(工具误报:),golint/fnsize(设计如此:)
func (t *Task) BuildCreateClusterTask(cls *proto.Cluster, opt *cloudprovider.CreateClusterOption) ( // nolint
	*proto.Task, error) {
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

	// init instance passwd
	passwd := utils.BuildInstancePwd()
	if opt.InitPassword != "" {
		passwd = opt.InitPassword
	}
	task.CommonParams[cloudprovider.PasswordKey.String()] = passwd

	// setting all steps details
	createClusterTask := &CreateClusterTaskOption{Cluster: cls, Nodes: opt.Nodes, NodeTemplate: opt.NodeTemplate}
	// step0: create cluster shield alarm step
	createClusterTask.BuildShieldAlertStep(task)
	// step1: createTKECluster and return clusterID inject common paras
	createClusterTask.BuildCreateClusterStep(task)
	// step2: check cluster status by clusterID
	createClusterTask.BuildCheckClusterStatusStep(task)
	// step3: check cluster nodes status
	createClusterTask.BuildCheckClusterNodesStatusStep(task)

	// step4: register managed cluster kubeConfig
	createClusterTask.BuildRegisterClsKubeConfigStep(task)

	// step5: 系统初始化 postAction bkops, platform run default steps
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil && opt.Cloud.ClusterManagement.CreateCluster != nil {
		err := template.BuildSopsFactory{
			StepName: template.SystemInit,
			Cluster:  cls,
			Extra: template.ExtraInfo{
				InstancePasswd: passwd,
				NodeOperator:   opt.Operator,
				NodeIPList:     strings.Join(opt.Nodes, ","),
			}}.BuildSopsStep(task, opt.Cloud.ClusterManagement.CreateCluster, false)
		if err != nil {
			return nil, fmt.Errorf("BuildCreateClusterTask BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step6: 业务后置自定义流程: 支持标准运维任务 或者 后置脚本
	if len(opt.Nodes) > 0 && opt.NodeTemplate != nil && len(opt.NodeTemplate.UserScript) > 0 {
		common.BuildJobExecuteScriptStep(task, common.JobExecParas{
			ClusterID: cls.ClusterID,
			Content:   opt.NodeTemplate.UserScript,
			NodeIps:   strings.Join(opt.Nodes, ","),
			Operator:  opt.Operator,
			StepName:  common.PostInitStepJob,
		})
	}

	// business post define sops task or script
	if len(opt.Nodes) > 0 && opt.NodeTemplate != nil && opt.NodeTemplate.ScaleOutExtraAddons != nil {
		err := template.BuildSopsFactory{
			StepName: template.UserAfterInit,
			Cluster:  cls,
			Extra: template.ExtraInfo{
				InstancePasswd: passwd,
				NodeIPList:     strings.Join(opt.Nodes, ","),
				NodeOperator:   opt.Operator,
				ShowSopsUrl:    true,
			}}.BuildSopsStep(task, opt.NodeTemplate.ScaleOutExtraAddons, false)
		if err != nil {
			return nil, fmt.Errorf("BuildScalingNodesTask business BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step7: enable vpc-cni network mode when cluster enable vpc-cni
	createClusterTask.BuildEnableVpcCniStep(task)
	// step8: update DB info by cluster data
	createClusterTask.BuildUpdateTaskStatusStep(task)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildCreateClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.OperatorKey.String()] = opt.Operator
	task.CommonParams[cloudprovider.UserKey.String()] = opt.Operator
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.CreateClusterJob.String()
	if len(opt.Nodes) > 0 {
		task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(opt.Nodes, ",")
	}

	return task, nil
}

// BuildCreateVirtualClusterTask build create virtual cluster task
func (t *Task) BuildCreateVirtualClusterTask(cls *proto.Cluster,
	opt *cloudprovider.CreateVirtualClusterOption) (*proto.Task, error) {
	// create virtual cluster by host cluster namespace
	// 1. hostCluster create namespace or exist in cluster
	// 2. hostCluster deploy vcluster/agent component
	// 3. wait subCluster kube-agent deployed
	// 4. subCluster deploy k8s-watch component

	// validate request params
	if cls == nil {
		return nil, fmt.Errorf("BuildCreateVirtualClusterTask cluster info empty")
	}
	if opt == nil || opt.Cloud == nil || opt.HostCluster == nil || opt.Namespace == nil {
		return nil, fmt.Errorf("BuildCreateVirtualClusterTask TaskOptions is lost")
	}

	nowStr := time.Now().Format(time.RFC3339)
	task := &proto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.CreateVirtualCluster),
		TaskName:       cloudprovider.CreateVirtualClusterTask.String(),
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
	taskName := fmt.Sprintf(createVirtualClusterTaskTemplate, cls.ClusterID)
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// setting all steps details
	createTask := CreateVirtualClusterTask{
		Cluster:     cls,
		HostCluster: opt.HostCluster,
		Namespace:   opt.Namespace,
	}
	createTask.BuildCreateNamespaceStep(task)
	createTask.BuildCreateResourceQuotaStep(task)
	createTask.BuildInstallVclusterStep(task)
	createTask.BuildCheckAgentStatusStep(task)
	createTask.BuildInstallWatchStep(task)

	// step6: 系统初始化 postAction bkops, platform run default steps
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil && opt.Cloud.ClusterManagement.CreateCluster != nil {
		err := template.BuildSopsFactory{
			StepName: template.SystemInit,
			Cluster:  cls,
			Extra: template.ExtraInfo{
				NodeOperator: opt.Operator,
			}}.BuildSopsStep(task, opt.Cloud.ClusterManagement.CreateCluster, false)
		if err != nil {
			return nil, fmt.Errorf("BuildCreateVirtualClusterTask BuildBkSopsStepAction failed: %v", err)
		}
	}

	createTask.BuildUpdateTaskStatusStep(task)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildCreateVirtualClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.OperatorKey.String()] = opt.Operator
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.CreateVirtualClusterJob.String()

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
	importTask := ImportClusterTaskOption{Cluster: cls}
	// step0: import cluster nodes step
	importTask.BuildImportClusterNodesStep(task)

	if options.GetEditionInfo().IsCommunicationEdition() ||
		options.GetEditionInfo().IsEnterpriseEdition() {
		// setting all steps details
		// step1: import cluster registerKubeConfigStep
		importTask.BuildRegisterKubeConfigStep(task)
		// step2: install cluster watch component
		common.BuildWatchComponentTaskStep(task, cls, "")
	}

	if options.GetEditionInfo().IsInnerEdition() {
		importTask.BuildRegisterClusterKubeConfigStep(task)
	}

	// importCluster sops task
	// run bk-sops, current only depend on bksops create task and only need to create one task
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil && opt.Cloud.ClusterManagement.ImportCluster != nil {
		err := template.BuildSopsFactory{
			StepName: template.SystemInit,
			Cluster:  cls,
			Extra: template.ExtraInfo{
				NodeOperator: opt.Operator,
				NodeIPList:   strings.Join(opt.NodeIPs, ","),
			}}.BuildSopsStep(task, opt.Cloud.ClusterManagement.ImportCluster, false)
		if err != nil {
			return nil, fmt.Errorf("BuildCreateClusterTask BuildBkSopsStepAction failed: %v", err)
		}
	}

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildImportClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.OperatorKey.String()] = opt.Operator
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.ImportClusterJob.String()

	return task, nil
}

// BuildDeleteVirtualClusterTask build delete virtual cluster task
func (t *Task) BuildDeleteVirtualClusterTask(cls *proto.Cluster,
	opt *cloudprovider.DeleteVirtualClusterOption) (*proto.Task, error) {
	// delete cluster has three steps:
	// 1. delete virtual cluster
	// 2. delete vcluster namespace in hostCluster
	// 3. delete cluster relative data && cluster credential

	// validate request params
	if cls == nil {
		return nil, fmt.Errorf("BuildDeleteVirtualClusterTask cluster info empty")
	}
	if opt == nil || opt.Operator == "" || opt.Cloud == nil || opt.HostCluster == nil || opt.Namespace == nil {
		return nil, fmt.Errorf("BuildDeleteVirtualClusterTask TaskOptions is lost")
	}

	// init task information
	nowStr := time.Now().Format(time.RFC3339)
	task := &proto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.DeleteVirtualCluster),
		TaskName:       cloudprovider.DeleteVirtualClusterTask.String(),
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
	taskName := fmt.Sprintf(deleteVirtualClusterTaskTemplate, cls.ClusterID)
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName
	task.CommonParams[cloudprovider.UserKey.String()] = opt.Operator

	// setting all steps details
	deleteVirtualClusterTask := &DeleteVirtualClusterTaskOption{
		Cluster:     cls,
		Cloud:       opt.Cloud,
		HostCluster: opt.HostCluster,
		Namespace:   opt.Namespace,
	}
	deleteVirtualClusterTask.BuildUninstallVClusterStep(task)
	deleteVirtualClusterTask.BuildDeleteNamespaceStep(task)
	deleteVirtualClusterTask.BuildCleanClusterDBInfoStep(task)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildDeleteVirtualClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.DeleteVirtualClusterJob.String()
	task.CommonParams[cloudprovider.OperatorKey.String()] = opt.Operator

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
	task.CommonParams[cloudprovider.UserKey.String()] = opt.Operator

	// deleteTKECluster sops preActions
	// run bk-sops, current only depend on bksops create task and only need to create one task
	if cloudprovider.IsInDependentCluster(opt.Cluster) && opt.Cloud != nil && opt.Cloud.ClusterManagement != nil &&
		opt.Cloud.ClusterManagement.DeleteCluster != nil {
		err := template.BuildSopsFactory{
			StepName: template.SystemInit,
			Cluster:  cls,
			Extra: template.ExtraInfo{
				NodeOperator: opt.Operator,
			}}.BuildSopsStep(task, opt.Cloud.ClusterManagement.DeleteCluster, true)
		if err != nil {
			return nil, fmt.Errorf("BuildDeleteClusterTask BuildBkSopsStepAction failed: %v", err)
		}
	}

	// setting all steps details
	deleteClusterTask := &DeleteClusterTaskOption{
		Cluster:    cls,
		DeleteMode: opt.DeleteMode.String(),
	}
	// step1: DeleteTKECluster delete tke cluster
	deleteClusterTask.BuildDeleteTKEClusterStep(task)
	// step2: update cluster DB info and associated data
	deleteClusterTask.BuildCleanClusterDBInfoStep(task)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildDeleteClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.DeleteClusterJob.String()
	task.CommonParams[cloudprovider.OperatorKey.String()] = opt.Operator

	return task, nil
}

// BuildAddNodesToClusterTask build addNodes task
// NOCC:CCN_threshold(工具误报:),golint/fnsize(设计如此:)
func (t *Task) BuildAddNodesToClusterTask(cls *proto.Cluster, nodes []*proto.Node, // nolint
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
	taskName := fmt.Sprintf(tkeAddNodeTaskTemplate, cls.ClusterID)
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// init instance passwd
	passwd := utils.BuildInstancePwd()
	if opt.InitPassword != "" {
		passwd = opt.InitPassword
	}
	task.CommonParams[cloudprovider.PasswordKey.String()] = passwd

	// setting all steps details
	addNodesTask := &AddNodesToClusterTaskOption{
		Cluster:      cls,
		Cloud:        opt.Cloud,
		NodeTemplate: opt.NodeTemplate,
		NodeIPs:      nodeIPs,
		NodeIDs:      nodeIDs,
		PassWd:       passwd,
		Operator:     opt.Operator,
		NodeSchedule: opt.NodeSchedule,
	}
	// step0: addNodes shield nodes alarm
	addNodesTask.BuildShieldAlertStep(task)
	// step1: addNodesToTKECluster add node to cluster
	addNodesTask.BuildAddNodesToClusterStep(task)
	// step2: check cluster add node status
	addNodesTask.BuildCheckAddNodesStatusStep(task)
	// step3: update DB node info by instanceIP
	addNodesTask.BuildUpdateAddNodeDBInfoStep(task)

	// postAction bk-sops task
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil && opt.Cloud.ClusterManagement.AddNodesToCluster != nil {
		err := template.BuildSopsFactory{
			StepName: template.SystemInit,
			Cluster:  cls,
			Extra: template.ExtraInfo{
				InstancePasswd: passwd,
				NodeIPList:     strings.Join(nodeIPs, ","),
				NodeOperator:   opt.Operator,
				ModuleID:       cloudprovider.GetScaleOutModuleID(cls, nil, opt.NodeTemplate, false),
				BusinessID:     cloudprovider.GetBusinessID(nil, opt.NodeTemplate, true),
			}}.BuildSopsStep(task, opt.Cloud.ClusterManagement.AddNodesToCluster, false)
		if err != nil {
			return nil, fmt.Errorf("BuildAddNodesToClusterTask BuildBkSopsStepAction failed: %v", err)
		}
	}

	// 业务后置自定义流程: 支持标准运维任务 或者 后置脚本
	if opt.NodeTemplate != nil && len(opt.NodeTemplate.UserScript) > 0 {
		common.BuildJobExecuteScriptStep(task, common.JobExecParas{
			ClusterID:        cls.ClusterID,
			Content:          opt.NodeTemplate.UserScript,
			NodeIps:          strings.Join(nodeIPs, ","),
			Operator:         opt.Operator,
			StepName:         common.PostInitStepJob,
			AllowSkipJobTask: opt.NodeTemplate.AllowSkipScaleOutWhenFailed,
		})
	}

	// business post define sops task or script
	if opt.NodeTemplate != nil && opt.NodeTemplate.ScaleOutExtraAddons != nil {
		err := template.BuildSopsFactory{
			StepName: template.UserAfterInit,
			Cluster:  cls,
			Extra: template.ExtraInfo{
				InstancePasswd: passwd,
				NodeIPList:     strings.Join(nodeIPs, ","),
				NodeOperator:   opt.Operator,
				ShowSopsUrl:    true,
			}}.BuildSopsStep(task, opt.NodeTemplate.ScaleOutExtraAddons, false)
		if err != nil {
			return nil, fmt.Errorf("BuildScalingNodesTask business BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step4: 若需要则设置节点注解
	addNodesTask.BuildNodeAnnotationsStep(task)
	// step5: 设置平台公共标签
	addNodesTask.BuildNodeLabelsStep(task)
	// step6: 设置节点可调度状态
	addNodesTask.BuildUnCordonNodesStep(task)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildAddNodesToClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.OperatorKey.String()] = opt.Operator
	task.CommonParams[cloudprovider.UserKey.String()] = opt.Operator

	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.AddNodeJob.String()
	task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")
	task.CommonParams[cloudprovider.NodeIDsKey.String()] = strings.Join(nodeIDs, ",")

	return task, nil
}

// BuildRemoveNodesFromClusterTask build removeNodes task
// NOCC:CCN_threshold(工具误报:),golint/fnsize(设计如此:)
func (t *Task) BuildRemoveNodesFromClusterTask(cls *proto.Cluster, nodes []*proto.Node, // nolint
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
	// generate taskName
	taskName := fmt.Sprintf(tkeCleanNodeTaskTemplate, cls.ClusterID)
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// setting all steps details
	removeNodesTask := &RemoveNodesFromClusterTaskOption{
		Cluster:    cls,
		Cloud:      opt.Cloud,
		DeleteMode: opt.DeleteMode,
		NodeIPs:    nodeIPs,
		NodeIDs:    nodeIDs,
	}

	// step0: cordon nodes
	removeNodesTask.BuildCordonNodesStep(task)

	// 业务自定义缩容流程: 支持 缩容节点前置脚本和前置标准运维流程
	if opt.NodeTemplate != nil && len(opt.NodeTemplate.ScaleInPreScript) > 0 {
		common.BuildJobExecuteScriptStep(task, common.JobExecParas{
			ClusterID:        cls.ClusterID,
			Content:          opt.NodeTemplate.ScaleInPreScript,
			NodeIps:          strings.Join(nodeIPs, ","),
			Operator:         opt.Operator,
			StepName:         common.PreInitStepJob,
			AllowSkipJobTask: opt.NodeTemplate.AllowSkipScaleInWhenFailed,
		})
	}
	// business define sops task
	if opt.NodeTemplate != nil && opt.NodeTemplate.ScaleInExtraAddons != nil {
		err := template.BuildSopsFactory{
			StepName: template.UserPreInit,
			Cluster:  cls,
			Extra: template.ExtraInfo{
				NodeIPList:   strings.Join(nodeIPs, ","),
				NodeOperator: opt.Operator,
				ShowSopsUrl:  true,
			}}.BuildSopsStep(task, opt.NodeTemplate.ScaleInExtraAddons, true)
		if err != nil {
			return nil, fmt.Errorf("BuildRemoveNodesFromClusterTask business "+
				"BuildBkSopsStepAction failed: %v", err)
		}
	}

	// preAction platform sops task
	if opt.Cloud != nil && opt.Cloud.ClusterManagement != nil &&
		opt.Cloud.ClusterManagement.DeleteNodesFromCluster != nil {
		err := template.BuildSopsFactory{
			StepName: template.SystemInit,
			Cluster:  cls,
			Extra: template.ExtraInfo{
				NodeIPList:   strings.Join(nodeIPs, ","),
				NodeOperator: opt.Operator,
				ModuleID:     cloudprovider.GetScaleInModuleID(nil, opt.NodeTemplate),
				BusinessID:   cloudprovider.GetBusinessID(nil, opt.NodeTemplate, false),
			}}.BuildSopsStep(task, opt.Cloud.ClusterManagement.DeleteNodesFromCluster, true)
		if err != nil {
			return nil, fmt.Errorf("BuildRemoveNodesFromClusterTask BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step1: removeNodesFromTKECluster remove nodes
	removeNodesTask.BuildRemoveNodesFromClusterStep(task)
	// step2: update node DB info
	removeNodesTask.BuildUpdateRemoveNodeDBInfoStep(task)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildRemoveNodesFromClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.DeleteNodeJob.String()
	task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")
	task.CommonParams[cloudprovider.NodeIDsKey.String()] = strings.Join(nodeIDs, ",")
	task.CommonParams[cloudprovider.OperatorKey.String()] = opt.Operator

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
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// setting all steps details
	createNodeGroupTask := &CreateNodeGroupTaskOption{Group: group}
	// step1. call qcloud create node group
	createNodeGroupTask.BuildCreateCloudNodeGroupStep(task)
	// step2. wait qcloud create node group complete
	createNodeGroupTask.BuildCheckCloudNodeGroupStatusStep(task)
	// step3. ensure autoscaler(安装/更新CA组件) in cluster
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
// NOCC:CCN_threshold(工具误报:),golint/fnsize(设计如此:)
func (t *Task) BuildCleanNodesInGroupTask(nodes []*proto.Node, group *proto.NodeGroup, // nolint
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

	var (
		nodeIPs, nodeIDs, deviceIDs = make([]string, 0), make([]string, 0), make([]string, 0)
	)
	for _, node := range nodes {
		nodeIPs = append(nodeIPs, node.InnerIP)
		nodeIDs = append(nodeIDs, node.NodeID)
		deviceIDs = append(deviceIDs, node.DeviceID)
	}

	isExternal := cloudprovider.IsExternalNodePool(group)

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

	// instance passwd
	passwd := group.LaunchTemplate.InitLoginPassword
	task.CommonParams[cloudprovider.PasswordKey.String()] = passwd

	// setting all steps details
	cleanNodeGroupNodes := &CleanNodeInGroupTaskOption{
		Group:     group,
		NodeIPs:   nodeIPs,
		NodeIds:   nodeIDs,
		DeviceIDs: deviceIDs,
		Operator:  opt.Operator,
	}

	// step0: cordon nodes
	cleanNodeGroupNodes.BuildCordonNodesStep(task)

	// step1. business user define flow
	if group.NodeTemplate != nil && len(group.NodeTemplate.ScaleInPreScript) > 0 {
		common.BuildJobExecuteScriptStep(task, common.JobExecParas{
			ClusterID:        opt.Cluster.ClusterID,
			Content:          group.NodeTemplate.ScaleInPreScript,
			NodeIps:          strings.Join(nodeIPs, ","),
			Operator:         opt.Operator,
			StepName:         common.PreInitStepJob,
			AllowSkipJobTask: group.NodeTemplate.AllowSkipScaleInWhenFailed,
		})
	}

	if group.NodeTemplate != nil && group.NodeTemplate.ScaleInExtraAddons != nil &&
		len(group.NodeTemplate.ScaleInExtraAddons.PreActions) > 0 {
		err := template.BuildSopsFactory{
			StepName: template.UserPreInit,
			Cluster:  opt.Cluster,
			Extra: template.ExtraInfo{
				InstancePasswd: passwd,
				NodeIPList:     strings.Join(nodeIPs, ","),
				NodeOperator:   opt.Operator,
				ShowSopsUrl:    true,
			}}.BuildSopsStep(task, group.NodeTemplate.ScaleInExtraAddons, true)
		if err != nil {
			return nil, fmt.Errorf("BuildCleanNodesInGroupTask ScaleInExtraAddons.PreActions "+
				"BuildBkSopsStepAction failed: %v", err)
		}
	}

	// platform clean sops task
	if !isExternal && opt.Cloud != nil && opt.Cloud.NodeGroupManagement != nil &&
		opt.Cloud.NodeGroupManagement.CleanNodesInGroup != nil {
		err := template.BuildSopsFactory{
			StepName: template.SystemInit,
			Cluster:  opt.Cluster,
			Extra: template.ExtraInfo{
				InstancePasswd: passwd,
				NodeIPList:     strings.Join(nodeIPs, ","),
				NodeOperator:   opt.Operator,
				ModuleID:       cloudprovider.GetScaleInModuleID(opt.AsOption, group.NodeTemplate),
				BusinessID:     cloudprovider.GetBusinessID(opt.AsOption, group.NodeTemplate, false),
			}}.BuildSopsStep(task, opt.Cloud.NodeGroupManagement.CleanNodesInGroup, true)
		if err != nil {
			return nil, fmt.Errorf("BuildCleanNodesInGroupTask business BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step1: cluster scaleIn to clean cluster nodes
	if !isExternal {
		cleanNodeGroupNodes.BuildCleanNodeGroupNodesStep(task)
		cleanNodeGroupNodes.BuildCheckClusterCleanNodsStep(task)
		common.BuildRemoveHostStep(task, opt.Cluster.BusinessID, nodeIPs)
	} else {
		cleanNodeGroupNodes.BuildRemoveExternalNodesStep(task)
		// externalNodes platform bk sops task 系统初始化
		if opt.Cloud != nil && opt.Cloud.NodeGroupManagement != nil &&
			opt.Cloud.NodeGroupManagement.DeleteExternalNodesFromCluster != nil {
			err := template.BuildSopsFactory{
				Cluster: opt.Cluster,
				Extra: template.ExtraInfo{
					NodeIPList:         strings.Join(nodeIPs, ","),
					NodeOperator:       opt.Operator,
					ExternalNodeScript: "",
					ModuleID:           cloudprovider.GetScaleInModuleID(opt.AsOption, group.NodeTemplate),
					BusinessID:         cloudprovider.GetBusinessID(opt.AsOption, group.NodeTemplate, false),
					NodeGroupID:        group.NodeGroupID,
				}}.BuildSopsStep(task, opt.Cloud.NodeGroupManagement.DeleteExternalNodesFromCluster, false)
			if err != nil {
				return nil, fmt.Errorf("BuildCleanNodesInGroupTask BuildBkSopsStepAction failed: %v", err)
			}
		}
		// 归还第三方节点机器
		cleanNodeGroupNodes.BuildReturnIDCNodeToResPoolStep(task)
	}

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
	deleteNodeGroupTask := &DeleteNodeGroupTaskOption{
		Group:                  group,
		CleanInstanceInCluster: opt.CleanInstanceInCluster,
	}
	// step1. call qcloud delete node group
	deleteNodeGroupTask.BuildDeleteNodeGroupStep(task)

	// step2. ensure autoscaler to remove this nodegroup
	if group.EnableAutoscale {
		common.BuildEnsureAutoScalerTaskStep(task, group.ClusterID, group.Provider)
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

func getTransModuleInfo(asOption *proto.ClusterAutoScalingOption, group *proto.NodeGroup) string {
	if group != nil && group.NodeTemplate != nil && group.NodeTemplate.Module != nil &&
		len(group.NodeTemplate.Module.ScaleOutModuleID) != 0 {
		return group.NodeTemplate.Module.ScaleOutModuleID
	}

	return asOption.GetModule().GetScaleOutModuleID()
}

// BuildUpdateDesiredNodesTask build update desired nodes task
// NOCC:CCN_threshold(工具误报:),golint/fnsize(设计如此:)
func (t *Task) BuildUpdateDesiredNodesTask(desired uint32, group *proto.NodeGroup, // nolint
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

	// normal or external nodePool
	isExternal := cloudprovider.IsExternalNodePool(group)

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
	updateDesiredNodesTask := &UpdateDesiredNodesTaskOption{
		Group:    group,
		Desired:  desired,
		Operator: opt.Operator,
	}
	// first: apply instance && add nodes to cluster
	if isExternal {
		// step1. call resource interface to apply externalNodes
		updateDesiredNodesTask.BuildApplyExternalNodeMachinesStep(task)
		// step2. get external nodes script
		updateDesiredNodesTask.BuildGetExternalNodeScriptStep(task)
	} else {
		// step1. call qcloud interface to set desired nodes
		updateDesiredNodesTask.BuildApplyInstanceMachinesStep(task)
		// step2. check cluster nodes and all nodes status is running
		updateDesiredNodesTask.BuildCheckClusterNodeStatusStep(task)
		// install gse agent
		common.BuildInstallGseAgentTaskStep(task, opt.Cluster, group, passwd, func() string {
			exist := checkIfWhiteImageOsNames(&cloudprovider.ClusterGroupOption{
				CommonOption: opt.CommonOption,
				Cluster:      opt.Cluster,
				Group:        opt.NodeGroup,
			})
			if exist {
				return fmt.Sprintf("%v", utils.ConnectPort)
			}

			return ""
		}())
		// transfer host module
		moduleID := getTransModuleInfo(opt.AsOption, opt.NodeGroup)
		if moduleID != "" {
			common.BuildTransferHostModuleStep(task, opt.Cluster.BusinessID, moduleID)
		}
	}

	// step3. platform define sops task
	if !isExternal && opt.Cloud != nil && opt.Cloud.NodeGroupManagement != nil &&
		opt.Cloud.NodeGroupManagement.UpdateDesiredNodes != nil {
		err := template.BuildSopsFactory{
			Cluster: opt.Cluster,
			Extra: template.ExtraInfo{
				InstancePasswd: passwd,
				NodeIPList:     "",
				NodeOperator:   opt.Operator,
				ModuleID: cloudprovider.GetScaleOutModuleID(opt.Cluster, opt.AsOption, group.NodeTemplate,
					true),
				BusinessID: cloudprovider.GetBusinessID(opt.AsOption, group.NodeTemplate, true),
			}}.BuildSopsStep(task, opt.Cloud.NodeGroupManagement.UpdateDesiredNodes, false)
		if err != nil {
			return nil, fmt.Errorf("BuildScalingNodesTask platform BuildBkSopsStepAction failed: %v", err)
		}
	}

	// external nodes postAction bk-sops task
	if isExternal && !group.GetNodeTemplate().GetSkipSystemInit() && opt.Cloud != nil &&
		opt.Cloud.NodeGroupManagement != nil && opt.Cloud.NodeGroupManagement.AddExternalNodesToCluster != nil {
		err := template.BuildSopsFactory{
			Cluster: opt.Cluster,
			Extra: template.ExtraInfo{
				NodeIPList:         "",
				NodeOperator:       opt.Operator,
				ExternalNodeScript: "",
				ModuleID: cloudprovider.GetScaleOutModuleID(opt.Cluster, opt.AsOption, group.NodeTemplate,
					false),
				BusinessID:  cloudprovider.GetBusinessID(opt.AsOption, group.NodeTemplate, true),
				NodeGroupID: group.NodeGroupID,
			}}.BuildSopsStep(task, opt.Cloud.NodeGroupManagement.AddExternalNodesToCluster, false)
		if err != nil {
			return nil, fmt.Errorf("BuildScalingNodesTask BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step4. business define sops task 支持脚本和标准运维流程
	if group.NodeTemplate != nil && len(group.NodeTemplate.UserScript) > 0 {
		common.BuildJobExecuteScriptStep(task, common.JobExecParas{
			ClusterID:        group.ClusterID,
			Content:          group.NodeTemplate.UserScript,
			NodeIps:          "",
			Operator:         opt.Operator,
			StepName:         common.PostInitStepJob,
			AllowSkipJobTask: group.NodeTemplate.GetAllowSkipScaleOutWhenFailed(),
		})
	}

	if group.NodeTemplate != nil && group.NodeTemplate.ScaleOutExtraAddons != nil {
		err := template.BuildSopsFactory{
			StepName: template.UserAfterInit,
			Cluster:  opt.Cluster,
			Extra: template.ExtraInfo{
				InstancePasswd:     passwd,
				NodeIPList:         "",
				NodeOperator:       opt.Operator,
				ShowSopsUrl:        true,
				ExternalNodeScript: "",
				NodeGroupID:        group.NodeGroupID,
			}}.BuildSopsStep(task, group.NodeTemplate.ScaleOutExtraAddons, false)
		if err != nil {
			return nil, fmt.Errorf("BuildScalingNodesTask business BuildBkSopsStepAction failed: %v", err)
		}
	}

	// set external node labels
	updateDesiredNodesTask.BuildNodeLabelsStep(task)

	// step4: set node annotations
	common.BuildNodeAnnotationsTaskStep(task, opt.Cluster.ClusterID, nil,
		cloudprovider.GetAnnotationsByNg(opt.NodeGroup))

	// step5: set resourcePool labels
	common.BuildResourcePoolLabelTaskStep(task, opt.Cluster.ClusterID)

	// step6. set node scheduler by nodeIPs
	updateDesiredNodesTask.BuildUnCordonNodesStep(task)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildUpdateDesiredNodesTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]

	// must set job-type
	task.CommonParams[cloudprovider.ScalingNodesNumKey.String()] = strconv.Itoa(int(desired))
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.UpdateNodeGroupDesiredNodeJob.String()
	task.CommonParams[cloudprovider.ManualKey.String()] = strconv.FormatBool(opt.Manual)

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
	// step2. update node group info in DB
	// switchNodeGroupTask.BuildUpdateNodeGroupAutoScalingDBStep(task)

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

// BuildSwitchAsOptionStatusTask build switch auto scaler option status task - 开启/关闭集群自动扩缩容
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
// NOCC:CCN_threshold(工具误报:),golint/fnsize(设计如此:)
func (t *Task) BuildAddExternalNodeToCluster(group *proto.NodeGroup, nodes []*proto.Node, // nolint
	opt *cloudprovider.AddExternalNodesOption) (*proto.Task, error) {
	// AddExternalNodeToCluster has three steps:
	// 1. call qcloud getExternalNodeScript get addNodes script
	// 2. call bksops add nodes to cluster
	// may be need to call external previous or behind operation by bkops

	// validate request params
	if group == nil {
		return nil, fmt.Errorf("BuildAddExternalNodeToCluster group info empty")
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("BuildAddExternalNodeToCluster lost nodes info")
	}

	if opt == nil || opt.Cloud == nil || opt.Operator == "" {
		return nil, fmt.Errorf("BuildAddExternalNodeToCluster TaskOptions is lost")
	}

	nodeIPs := make([]string, 0)
	for i := range nodes {
		nodeIPs = append(nodeIPs, nodes[i].InnerIP)
	}

	// init task information
	nowStr := time.Now().Format(time.RFC3339)
	task := &proto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.AddExternalNodesToCluster),
		TaskName:       cloudprovider.AddExternalNodesToClusterTask.String(),
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
		NodeIPList:     nodeIPs,
	}
	taskName := fmt.Sprintf(tkeAddExternalNodeTaskTemplate, group.ClusterID)
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// setting all steps details
	addExternalNodesTask := &AddExternalNodesToClusterTaskOption{
		Group:   group,
		NodeIPs: nodeIPs,
		Cluster: opt.Cluster,
	}
	// step1: addNodesToTKECluster add node to cluster
	addExternalNodesTask.BuildGetExternalNodeScriptStep(task)

	// postAction bk-sops task
	if !group.GetNodeTemplate().GetSkipSystemInit() && opt.Cloud != nil &&
		opt.Cloud.NodeGroupManagement != nil && opt.Cloud.NodeGroupManagement.AddExternalNodesToCluster != nil {
		err := template.BuildSopsFactory{
			Cluster: opt.Cluster,
			Extra: template.ExtraInfo{
				NodeIPList:         strings.Join(nodeIPs, ","),
				NodeOperator:       opt.Operator,
				ExternalNodeScript: "",
				ModuleID: cloudprovider.GetScaleOutModuleID(opt.Cluster, nil, group.NodeTemplate,
					false),
				BusinessID:  cloudprovider.GetBusinessID(nil, group.NodeTemplate, true),
				NodeGroupID: group.NodeGroupID,
			}}.BuildSopsStep(task, opt.Cloud.NodeGroupManagement.AddExternalNodesToCluster, false)
		if err != nil {
			return nil, fmt.Errorf("BuildAddExternalNodeToCluster BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step3. business define sops task 支持脚本和标准运维流程
	if group.NodeTemplate != nil && len(group.NodeTemplate.UserScript) > 0 {
		common.BuildJobExecuteScriptStep(task, common.JobExecParas{
			ClusterID:        group.ClusterID,
			Content:          group.NodeTemplate.UserScript,
			NodeIps:          strings.Join(nodeIPs, ","),
			Operator:         opt.Operator,
			StepName:         common.PostInitStepJob,
			AllowSkipJobTask: group.NodeTemplate.GetAllowSkipScaleOutWhenFailed(),
		})
	}

	if group.NodeTemplate != nil && group.NodeTemplate.ScaleOutExtraAddons != nil {
		err := template.BuildSopsFactory{
			StepName: template.UserAfterInit,
			Cluster:  opt.Cluster,
			Extra: template.ExtraInfo{
				InstancePasswd:     "",
				NodeIPList:         strings.Join(nodeIPs, ","),
				NodeOperator:       opt.Operator,
				ShowSopsUrl:        true,
				ExternalNodeScript: "",
				NodeGroupID:        group.NodeGroupID,
			}}.BuildSopsStep(task, group.NodeTemplate.ScaleOutExtraAddons, false)
		if err != nil {
			return nil, fmt.Errorf("BuildScalingNodesTask business BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step3: 设置节点labels
	addExternalNodesTask.BuildNodeLabelsStep(task)
	// step4: 设置节点可调度状态
	addExternalNodesTask.BuildUnCordonNodesStep(task)

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildAddExternalNodeToCluster task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.OperatorKey.String()] = opt.Operator
	task.CommonParams[cloudprovider.UserKey.String()] = opt.Operator

	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.AddExternalNodeJob.String()
	task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")

	return task, nil
}

// BuildDeleteExternalNodeFromCluster remove external node from cluster
func (t *Task) BuildDeleteExternalNodeFromCluster(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.DeleteExternalNodesOption) (*proto.Task, error) {
	// DeleteExternalNodeFromCluster has two steps:
	// 1. call qcloud DeleteExternalNodes
	// 2. call bksops clean node
	// may be need to call external previous or behind operation by bkops

	// validate request params
	if group == nil {
		return nil, fmt.Errorf("BuildDeleteExternalNodeFromCluster cluster info empty")
	}
	if opt == nil || opt.Cloud == nil || opt.Cluster == nil {
		return nil, fmt.Errorf("BuildDeleteExternalNodeFromCluster TaskOptions is lost")
	}

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
		TaskType:       cloudprovider.GetTaskType(cloudName, cloudprovider.RemoveExternalNodesFromCluster),
		TaskName:       cloudprovider.RemoveExternalNodesFromClusterTask.String(),
		Status:         cloudprovider.TaskStatusInit,
		Message:        "task initializing",
		Start:          nowStr,
		Steps:          make(map[string]*proto.Step),
		StepSequence:   make([]string, 0),
		ClusterID:      opt.Cluster.ClusterID,
		ProjectID:      opt.Cluster.ProjectID,
		Creator:        opt.Operator,
		Updater:        opt.Operator,
		LastUpdate:     nowStr,
		CommonParams:   make(map[string]string),
		ForceTerminate: false,
		NodeIPList:     nodeIPs,
	}
	// generate taskName
	taskName := fmt.Sprintf(tkeCleanExternalNodeTaskTemplate, opt.Cluster.ClusterID)
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	// setting all steps details
	removeExternalNodesTask := &RemoveExternalNodesFromClusterTaskOption{
		Cluster: opt.Cluster,
		Group:   group,
		NodeIPs: nodeIPs,
	}
	// step0: cordon nodes
	removeExternalNodesTask.BuildCordonNodesStep(task)

	// step1: RemoveExternalNodes remove nodes
	removeExternalNodesTask.BuildRemoveExternalNodesStep(task)

	// step2: preAction platform sops task
	if opt.Cloud != nil && opt.Cloud.NodeGroupManagement != nil &&
		opt.Cloud.NodeGroupManagement.DeleteExternalNodesFromCluster != nil {
		err := template.BuildSopsFactory{
			Cluster: opt.Cluster,
			Extra: template.ExtraInfo{
				NodeIPList:         strings.Join(nodeIPs, ","),
				NodeOperator:       opt.Operator,
				ExternalNodeScript: "",
				ModuleID:           cloudprovider.GetScaleInModuleID(nil, group.NodeTemplate),
				BusinessID:         cloudprovider.GetBusinessID(nil, group.NodeTemplate, false),
				NodeGroupID:        group.NodeGroupID,
			}}.BuildSopsStep(task, opt.Cloud.NodeGroupManagement.DeleteExternalNodesFromCluster, false)
		if err != nil {
			return nil, fmt.Errorf("BuildDeleteExternalNodeFromCluster BuildBkSopsStepAction failed: %v", err)
		}
	}

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildDeleteExternalNodeFromCluster task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.DeleteExternalNodeJob.String()
	task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")

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
