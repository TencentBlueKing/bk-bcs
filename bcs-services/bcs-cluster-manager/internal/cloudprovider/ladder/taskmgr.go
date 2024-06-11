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

package ladder

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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/ladder/tasks"
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

	// create nodeGroup
	task.works[createNodePoolStep.StepMethod] = tasks.CreateNodePoolTask

	// clean node task
	task.works[removeNodesFromClusterStep.StepMethod] = tasks.RemoveNodesFromClusterTask
	task.works[returnInstanceToResourcePoolStep.StepMethod] = tasks.ReturnInstanceToResourcePoolTask

	// scale node task
	task.works[applyCVMFromResourcePoolStep.StepMethod] = tasks.ApplyCVMFromResourcePoolTask
	task.works[addNodesToClusterStep.StepMethod] = tasks.AddNodesToClusterTask
	task.works[checkClusterNodeStatusStep.StepMethod] = tasks.CheckClusterNodeStatusTask
	task.works[checkClusterNodesInCMDBStep.StepMethod] = tasks.CheckClusterNodesInCMDBTask
	task.works[syncClusterNodesToCMDBStep.StepMethod] = tasks.SyncClusterNodesToCMDBTask
	// task.works[resourcePoolLabelStep.StepMethod] = tasks.SetResourcePoolDeviceLabels

	// delete nodeGroup task
	task.works[checkCleanDBDataStep.StepMethod] = tasks.CheckCleanDBDataTask

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

// GetAllTask register all background task for worker running
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

// BuildCreateNodeGroupTask build create node group task
func (t *Task) BuildCreateNodeGroupTask(group *proto.NodeGroup, opt *cloudprovider.CreateNodeGroupOption) (
	*proto.Task, error) {
	// yunti create nodeGroup steps
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
	// step1. call qcloud create node group
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

// BuildCleanNodesInGroupTask clean specified nodes in NodeGroup
// including remove nodes from NodeGroup, clean data in nodes
func (t *Task) BuildCleanNodesInGroupTask(nodes []*proto.Node, group *proto.NodeGroup, // nolint
	opt *cloudprovider.CleanNodesOption) (*proto.Task, error) {

	// clean nodes in yunti only has two steps:
	// 1. call LeaveClusterAndReturnCVM to clean cluster nodes
	// 2. check cvm state and clean Node
	// because cvms return to yunti resource pool, all clean works are handle by yunti
	// we do little task here

	// validate request params
	if len(nodes) == 0 {
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

	// init task information
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
		Creator:        opt.Operator,
		Updater:        opt.Operator,
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
	cleanTask := &CleanNodesInGroupTaskOption{
		NodeGroup: group,
		Cluster:   opt.Cluster,
		NodeIDs:   nodeIDs,
		NodeIPs:   nodeIPs,
		DeviceIDs: deviceIDs,
		Operator:  opt.Operator,
	}

	// step0: cluster cordon nodes
	cleanTask.BuildCordonNodesStep(task)

	// 业务自定义流程: 缩容前置流程支持 标准运维任务和执行业务job脚本
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

	if group.NodeTemplate != nil && group.NodeTemplate.ScaleInExtraAddons != nil {
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
			return nil, fmt.Errorf("BuildCleanNodesInGroupTask business BuildBkSopsStepAction failed: %v", err)
		}
	}

	// platform clean sops task
	if opt.Cloud != nil && opt.Cloud.NodeGroupManagement != nil && opt.Cloud.NodeGroupManagement.CleanNodesInGroup != nil {
		err := template.BuildSopsFactory{
			StepName: template.SystemInit,
			Cluster:  opt.Cluster,
			Extra: template.ExtraInfo{
				NodeIPList:   strings.Join(nodeIPs, ","),
				NodeOperator: opt.Operator,
				ModuleID:     cloudprovider.GetScaleInModuleID(opt.AsOption, group.NodeTemplate),
				BusinessID:   cloudprovider.GetBusinessID(opt.Cluster, opt.AsOption, group.NodeTemplate, false),
			}}.BuildSopsStep(task, opt.Cloud.NodeGroupManagement.CleanNodesInGroup, true)
		if err != nil {
			return nil, fmt.Errorf("BuildCleanNodesInGroupTask business BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step1: cluster scale-in to clean nodes
	cleanTask.BuildRemoveNodesStep(task)
	// step2: cluster return nodes
	cleanTask.BuildReturnNodesStep(task)

	if len(task.StepSequence) > 0 {
		task.CurrentStep = task.StepSequence[0]
	}

	// Job-type
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.CleanNodeGroupNodesJob.String()
	task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")
	task.CommonParams[cloudprovider.NodeIDsKey.String()] = strings.Join(nodeIDs, ",")

	return task, nil
}

// BuildUpdateDesiredNodesTask scale cluster nodes
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

	// passwd := group.GetLaunchTemplate().GetInitLoginPassword()
	passwd := utils.BuildInstancePwd()
	task.CommonParams[cloudprovider.PasswordKey.String()] = passwd

	updateDesiredNodes := &UpdateDesiredNodesTaskOption{
		NodeGroup: group,
		Cluster:   opt.Cluster,
		Desired:   int(desired),
		Operator:  opt.Operator,
	}
	// step1: apply Instance from resourcePool
	updateDesiredNodes.BuildApplyInstanceStep(task)
	// step2: add Nodes to cluster
	updateDesiredNodes.BuildAddNodesToClusterStep(task)
	// step3: check nodes status
	updateDesiredNodes.BuildCheckClusterNodeStatusStep(task)
	// step4: check nodes if sync to cmdb
	// updateDesiredNodes.BuildCheckClusterNodesInCMDBStep(task)
	updateDesiredNodes.BuildSyncClusterNodesToCMDBStep(task)
	// platform define sops task
	if opt.Cloud != nil && opt.Cloud.NodeGroupManagement != nil &&
		opt.Cloud.NodeGroupManagement.UpdateDesiredNodes != nil {
		err := template.BuildSopsFactory{
			StepName: template.SystemInit,
			Cluster:  opt.Cluster,
			Extra: template.ExtraInfo{
				InstancePasswd: passwd,
				NodeIPList:     "",
				NodeOperator:   opt.Operator,
				GroupCreator:   group.Creator,
				ModuleID: cloudprovider.GetScaleOutModuleID(opt.Cluster, opt.AsOption, group.NodeTemplate,
					true),
				BusinessID: cloudprovider.GetBusinessID(opt.Cluster, opt.AsOption, group.NodeTemplate, true),
			}}.BuildSopsStep(task, opt.Cloud.NodeGroupManagement.UpdateDesiredNodes, false)
		if err != nil {
			return nil, fmt.Errorf("BuildScalingNodesTask platform BuildBkSopsStepAction failed: %v", err)
		}
	}

	// 业务扩容节点后置自定义流程: 支持job后置脚本和标准运维任务
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

	// business define sops task
	if group.NodeTemplate != nil && group.NodeTemplate.ScaleOutExtraAddons != nil {
		err := template.BuildSopsFactory{
			StepName: template.UserAfterInit,
			Cluster:  opt.Cluster,
			Extra: template.ExtraInfo{
				InstancePasswd:  passwd,
				NodeIPList:      "",
				NodeOperator:    opt.Operator,
				ShowSopsUrl:     true,
				TranslateMethod: template.UserPostInit,
			}}.BuildSopsStep(task, group.NodeTemplate.ScaleOutExtraAddons, false)
		if err != nil {
			return nil, fmt.Errorf("BuildScalingNodesTask business BuildBkSopsStepAction failed: %v", err)
		}
	}

	// 混部集群需要执行混部节点流程
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
			return nil, fmt.Errorf("BuildScalingNodesTask BuildBkSopsStepAction failed: %v", err)
		}
	}

	// step3: annotation nodes
	updateDesiredNodes.BuildNodeAnnotationsStep(task)
	// step4: nodes common labels: sZoneID / bizID
	updateDesiredNodes.BuildNodeCommonLabelsStep(task)
	// step5: set resourcePool labels
	updateDesiredNodes.BuildResourcePoolDeviceLabelStep(task)
	// step6: unCordon nodes
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

// BuildDeleteNodeGroupTask delete nodegroup
func (t *Task) BuildDeleteNodeGroupTask(group *proto.NodeGroup, nodes []*proto.Node, // nolint
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
		deviceIDs = append(deviceIDs, node.DeviceID)
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

	cleanNodesTask := CleanNodesInGroupTaskOption{
		NodeGroup: group,
		Cluster:   opt.Cluster,
		NodeIDs:   nodeIDs,
		NodeIPs:   nodeIPs,
		DeviceIDs: deviceIDs,
		Operator:  opt.Operator,
	}

	// first clean nodegroup nodes if node group exist nodes
	if len(nodeIDs) > 0 && len(nodeIPs) > 0 {
		common.BuildCordonNodesTaskStep(task, opt.Cluster.ClusterID, nodeIPs)

		// 业务自定义流程: 缩容前置流程支持 标准运维任务和执行业务job脚本
		if group.NodeTemplate != nil && len(group.NodeTemplate.ScaleInPreScript) > 0 {
			common.BuildJobExecuteScriptStep(task, common.JobExecParas{
				ClusterID: opt.Cluster.ClusterID,
				Content:   group.NodeTemplate.ScaleInPreScript,
				NodeIps:   strings.Join(nodeIPs, ","),
				Operator:  opt.Operator,
				StepName:  common.PreInitStepJob,
				Translate: common.PreInitJob,
			})
		}

		// business define sops task
		if group.NodeTemplate != nil && group.NodeTemplate.ScaleInExtraAddons != nil {
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
				return nil, fmt.Errorf("BuildCleanNodesInGroupTask business BuildBkSopsStepAction failed: %v", err)
			}
		}

		// platform clean sops task
		if opt.Cloud != nil && opt.Cloud.NodeGroupManagement != nil &&
			opt.Cloud.NodeGroupManagement.CleanNodesInGroup != nil {
			err := template.BuildSopsFactory{
				StepName: template.SystemInit,
				Cluster:  opt.Cluster,
				Extra: template.ExtraInfo{
					NodeIPList:   strings.Join(nodeIPs, ","),
					NodeOperator: opt.Operator,
					ModuleID:     cloudprovider.GetScaleInModuleID(opt.AsOption, group.NodeTemplate),
					BusinessID: cloudprovider.GetBusinessID(opt.Cluster, opt.AsOption,
						group.NodeTemplate, false),
				}}.BuildSopsStep(task, opt.Cloud.NodeGroupManagement.CleanNodesInGroup, true)
			if err != nil {
				return nil, fmt.Errorf("BuildCleanNodesInGroupTask business BuildBkSopsStepAction failed: %v", err)
			}
		}

		// step1: cluster scale-in to clean nodes
		cleanNodesTask.BuildRemoveNodesStep(task)
		// step2: cluster return nodes
		cleanNodesTask.BuildReturnNodesStep(task)
	}

	deleteNodeGroupOption := &DeleteNodeGroupOption{
		NodeGroup: group,
		Cluster:   opt.Cluster,
		NodeIDs:   nodeIDs,
		NodeIPs:   nodeIPs,
		Operator:  opt.Operator,
	}
	// step3: previous call successful and delete local storage information
	deleteNodeGroupOption.BuildCheckCleanDBDataStep(task)
	// step4: delete nodeGroup from CA
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

// BuildCreateClusterTask build create cluster task
func (t *Task) BuildCreateClusterTask(cls *proto.Cluster, opt *cloudprovider.CreateClusterOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildImportClusterTask build import cluster task
func (t *Task) BuildImportClusterTask(cls *proto.Cluster, opt *cloudprovider.ImportClusterOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildDeleteClusterTask build delete cluster task
func (t *Task) BuildDeleteClusterTask(cls *proto.Cluster, opt *cloudprovider.DeleteClusterOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildAddNodesToClusterTask build addNodesToCluster task
func (t *Task) BuildAddNodesToClusterTask(cls *proto.Cluster, nodes []*proto.Node,
	opt *cloudprovider.AddNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildRemoveNodesFromClusterTask build removeNodesFromCluster task
func (t *Task) BuildRemoveNodesFromClusterTask(cls *proto.Cluster, nodes []*proto.Node,
	opt *cloudprovider.DeleteNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// BuildMoveNodesToGroupTask build move nodes to group task
func (t *Task) BuildMoveNodesToGroupTask(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.MoveNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
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

	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildSwitchNodeGroupAutoScalingTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]
	task.CommonParams[cloudprovider.JobTypeKey.String()] = cloudprovider.SwitchNodeGroupAutoScalingJob.String()

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

// BuildUpdateAutoScalingOptionTask xxx
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

// BuildSwitchAsOptionStatusTask xxx
func (t *Task) BuildSwitchAsOptionStatusTask(scalingOption *proto.ClusterAutoScalingOption,
	enable bool, opt *cloudprovider.CommonOption) (*proto.Task, error) {
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
