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
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
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
	task.works[createAKSClusterStep.StepMethod] = tasks.CreateAKSClusterTask
	task.works[checkAKSClusterStatusStep.StepMethod] = tasks.CheckAKSClusterStatusTask
	task.works[checkAKSNodeGroupsStatusStep.StepMethod] = tasks.CheckAKSNodeGroupsStatusTask
	task.works[updateAKSNodeGroupsToDBStep.StepMethod] = tasks.UpdateAKSNodesGroupToDBTask
	task.works[checkCreateClusterNodeStatusStep.StepMethod] = tasks.CheckAKSClusterNodesStatusTask
	task.works[updateAKSNodesToDBStep.StepMethod] = tasks.UpdateAKSNodesToDBTask
	task.works[registerAKSClusterKubeConfigStep.StepMethod] = tasks.RegisterAKSClusterKubeConfigTask

	// import cluster task
	task.works[importClusterNodesStep.StepMethod] = tasks.ImportClusterNodesTask
	task.works[registerClusterKubeConfigStep.StepMethod] = tasks.RegisterClusterKubeConfigTask

	// delete cluster task
	task.works[deleteAKSClusterStep.StepMethod] = tasks.DeleteAKSClusterTask
	task.works[cleanClusterDBInfoStep.StepMethod] = tasks.CleanClusterDBInfoTask

	// create nodeGroup task
	task.works[createCloudNodeGroupStep.StepMethod] = tasks.CreateCloudNodeGroupTask
	task.works[checkCloudNodeGroupStatusStep.StepMethod] = tasks.CheckCloudNodeGroupStatusTask

	// delete nodeGroup task
	task.works[deleteNodeGroupStep.StepMethod] = tasks.DeleteCloudNodeGroupTask

	// clean node in nodeGroup task
	task.works[cleanNodeGroupNodesStep.StepMethod] = tasks.CleanNodeGroupNodesTask

	// update desired nodes task
	task.works[applyInstanceMachinesStep.StepMethod] = tasks.ApplyInstanceMachinesTask
	task.works[checkClusterNodesStatusStep.StepMethod] = tasks.CheckClusterNodesStatusTask

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
	// 1. call azure CreateAKSCluster to create tke cluster
	// 2. call GetAKSCluster to check cluster run status(cluster status: Running Creating Abnormal))
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

	// setting all steps details
	createClusterTask := &CreateClusterTaskOption{Cluster: cls, NodeGroupIDs: opt.NodeGroupIDs}

	// step0: createAKSCluster and return clusterID inject common paras
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
	// step7: install cluster watch component
	common.BuildWatchComponentTaskStep(task, cls, "")
	// step8: 若需要则设置节点注解
	common.BuildNodeAnnotationsTaskStep(task, cls.ClusterID, nil, func() map[string]string {
		if opt.NodeTemplate != nil && len(opt.NodeTemplate.GetAnnotations()) > 0 {
			return opt.NodeTemplate.GetAnnotations()
		}
		return nil
	}())

	// step9: install gse agent
	common.BuildInstallGseAgentTaskStep(task, &common.GseInstallInfo{
		ClusterId:          cls.ClusterID,
		BusinessId:         cls.BusinessID,
		CloudArea:          cls.GetClusterBasicSettings().GetArea(),
		User:               cls.GetNodeSettings().GetWorkerLogin().GetInitLoginUsername(),
		Passwd:             cls.GetNodeSettings().GetWorkerLogin().GetInitLoginPassword(),
		KeyInfo:            cls.GetNodeSettings().GetWorkerLogin().GetKeyPair(),
		AllowReviseCloudId: icommon.True,
	}, cloudprovider.WithStepAllowSkip(true))

	// step10: 业务后置自定义流程: 支持标准运维任务 或者 后置脚本
	if opt.NodeTemplate != nil && len(opt.NodeTemplate.UserScript) > 0 {
		common.BuildJobExecuteScriptStep(task, common.JobExecParas{
			ClusterID: cls.ClusterID,
			Content:   opt.NodeTemplate.UserScript,
			// dynamic node ips
			NodeIps:   "",
			Operator:  opt.Operator,
			StepName:  common.PostInitStepJob,
			Translate: common.PostInitJob,
		})
	}
	// business post define sops task or script
	if opt.NodeTemplate != nil && opt.NodeTemplate.ScaleOutExtraAddons != nil {
		err := template.BuildSopsFactory{
			StepName: template.UserAfterInit,
			Cluster:  cls,
			Extra: template.ExtraInfo{
				// dynamic node ips
				NodeIPList:      "",
				NodeOperator:    opt.Operator,
				ShowSopsUrl:     true,
				TranslateMethod: template.UserPostInit,
			}}.BuildSopsStep(task, opt.NodeTemplate.ScaleOutExtraAddons, false)
		if err != nil {
			return nil, fmt.Errorf("BuildCreateClusterTask business BuildBkSopsStepAction failed: %v", err)
		}
	}

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildCreateClusterTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.OperatorKey.String()] = opt.Operator
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.CreateClusterJob.String()

	if len(opt.WorkerNodes) > 0 {
		task.CommonParams[cloudprovider.WorkerNodeIPsKey.String()] = strings.Join(opt.WorkerNodes, ",")
	}
	if len(opt.MasterNodes) > 0 {
		task.CommonParams[cloudprovider.MasterNodeIPsKey.String()] = strings.Join(opt.MasterNodes, ",")
	}

	return task, nil
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
		Cluster:    cls,
		DeleteMode: opt.DeleteMode.String(),
	}
	// step1: deleteAKSClusterTask delete aks cluster
	deleteCluster.BuildDeleteAKSClusterStep(task)
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
	createNodeGroup := &CreateNodeGroupTaskOption{Group: group}
	// step1. call aks create node group
	createNodeGroup.BuildCreateCloudNodeGroupStep(task)
	// step2. wait aks create node group complete
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
// including remove nodes from NodeGroup, clean data in nodes - 缩容，不保留节点
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
	cluster, err := cloudprovider.GetStorageModel().GetCluster(context.Background(), group.ClusterID)
	if err != nil {
		return nil, fmt.Errorf("BuildCleanNodesInGroupTask get cluster %s error, %s", group.ClusterID, err.Error())
	}

	var (
		nodeIPs, nodeIDs, nodeNames = make([]string, 0), make([]string, 0), make([]string, 0)
	)
	for _, node := range nodes {
		nodeIPs = append(nodeIPs, node.InnerIP)
		// nodeIDs for azure vmSet remove vm
		nodeIDs = append(nodeIDs, node.NodeID)
		if node.NodeName != "" {
			nodeNames = append(nodeNames, node.NodeName)
		}
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

	// instance passwd
	passwd := group.LaunchTemplate.InitLoginPassword
	task.CommonParams[cloudprovider.PasswordKey.String()] = passwd

	// setting all steps details
	cleanNodes := &CleanNodeInGroupTaskOption{
		Group:    group,
		NodeIPs:  nodeIPs,
		NodeIds:  nodeIDs,
		Operator: opt.Operator,
	}

	// step1: cluster scaleIn to clean cluster nodes
	common.BuildCordonNodesTaskStep(task, opt.Cluster.ClusterID, nodeIPs)

	// step2. business user define flow
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
				InstancePasswd: "",
				NodeIPList:     strings.Join(nodeIPs, ","),
				NodeOperator:   opt.Operator,
				ShowSopsUrl:    true,
			}}.BuildSopsStep(task, group.NodeTemplate.ScaleInExtraAddons, true)
		if err != nil {
			return nil, fmt.Errorf("BuildCleanNodesInGroupTask ScaleInExtraAddons.PreActions "+
				"BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step3: cluster delete nodes
	cleanNodes.BuildCleanNodeGroupNodesStep(task)

	// step4: check deleted node status
	common.BuildCheckClusterCleanNodesTaskStep(task, group.Provider, opt.Cluster.ClusterID, nodeNames)

	// step4: remove node ip from cmdb
	common.BuildRemoveHostStep(task, cluster.BusinessID, nodeIPs)

	// set current step
	task.CurrentStep = task.StepSequence[0]

	// set global task paras
	task.CommonParams[cloudprovider.NodeIDsKey.String()] = strings.Join(nodeIDs, ",")
	task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")
	task.CommonParams[cloudprovider.NodeNamesKey.String()] = strings.Join(nodeNames, ",")

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
	// step1. call aks delete node group
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

// BuildMoveNodesToGroupTask build move nodes to group task - 节点移入节点池
func (t *Task) BuildMoveNodesToGroupTask(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.MoveNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildUpdateDesiredNodesTask build update desired nodes task - 扩容节点
func (t *Task) BuildUpdateDesiredNodesTask(desired uint32, group *proto.NodeGroup, // nolint
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
	// step1. call azure interface to set desired nodes
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
	moduleID := cloudprovider.GetTransModuleInfo(opt.Cluster, opt.AsOption, opt.NodeGroup)
	if moduleID != "" {
		common.BuildTransferHostModuleStep(task, opt.Cluster.BusinessID, moduleID, "")
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

	// step4: set node annotations
	common.BuildNodeAnnotationsTaskStep(task, opt.Cluster.ClusterID, nil,
		cloudprovider.GetAnnotationsByNg(opt.NodeGroup))

	// step5: remove nodes inner taints
	// common.BuildRemoveClusterNodesInnerTaintTaskStep(task, group)

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
		TaskName:       cloudprovider.SwitchNodeGroupAutoScalingTask.String(),
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

// BuildAddExternalNodeToCluster xxx
func (t *Task) BuildAddExternalNodeToCluster(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.AddExternalNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildDeleteExternalNodeFromCluster xxx
func (t *Task) BuildDeleteExternalNodeFromCluster(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.DeleteExternalNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildUpdateNodeGroupTask when update nodegroup, we need to create background task,
func (t *Task) BuildUpdateNodeGroupTask(group *proto.NodeGroup, opt *cloudprovider.CommonOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
