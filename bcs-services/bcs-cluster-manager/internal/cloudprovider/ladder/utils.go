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

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/common"
)

var (
	cloudName = "yunti"
)

const (
	// createNodeGroupTaskTemplate bk-sops add task template
	createNodeGroupTaskTemplate = "yunti-create node group: %s/%s"
	// deleteNodeGroupTaskTemplate bk-sops add task template
	deleteNodeGroupTaskTemplate = "yunti-delete node group: %s/%s"
	// updateNodeGroupTaskTemplate bk-sops add task template
	updateNodeGroupTaskTemplate = "yunti-update node group: %s/%s"
	// updateNodeGroupDesiredNode bk-sops add task template
	updateNodeGroupDesiredNodeTemplate = "yunti-update node group desired node: %s/%s"
	// cleanNodeGroupNodesTaskTemplate bk-sops add task template
	cleanNodeGroupNodesTaskTemplate = "yunti-remove node group nodes: %s/%s"

	// switchNodeGroupAutoScalingTaskTemplate bk-sops add task template
	switchNodeGroupAutoScalingTaskTemplate = "yunti-switch node group auto scaling: %s/%s"
	// updateAutoScalingOptionTemplate bk-sops add task template
	updateAutoScalingOptionTemplate = "yunti-update auto scaling option: %s"
	// switchAutoScalingOptionStatusTemplate bk-sops add task template
	switchAutoScalingOptionStatusTemplate = "yunti-switch auto scaling option status: %s"
)

var (
	// BuildCleanNodesInGroupTask: stepName and stepMethod
	removeNodesFromClusterStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-RemoveNodesFromClusterTask", cloudName),
		StepName:   "下架集群节点",
	}
	returnInstanceToResourcePoolStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-ReturnInstanceToResourcePoolTask", cloudName),
		StepName:   "回收节点",
	}

	// BuildUpdateDesiredNodesTask: stepName and stepMethod
	applyCVMFromResourcePoolStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-%s", cloudName, cloudprovider.ApplyInstanceMachinesTask),
		StepName:   "申请节点任务",
	}
	addNodesToClusterStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-AddNodesToClusterTask", cloudName),
		StepName:   "上架集群节点",
	}
	checkClusterNodeStatusStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CheckClusterNodeStatusStep", cloudName),
		StepName:   "检测节点状态",
	}
	syncClusterNodesToCMDBStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-SyncClusterNodesToCMDBStep", cloudName),
		StepName:   "同步节点至bkcc",
	}

	// delete nodeGroup task: stepName and stepMethod
	checkCleanDBDataStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CheckCleanDBDataTask", cloudName),
		StepName:   "清理节点组数据",
	}

	// create nodeGroup task: stepName and stepMethod
	createNodePoolStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CreateNodePoolTask", cloudName),
		StepName:   "创建资源池",
	}
)

// CleanNodesInGroupTaskOption for build CleanNodesInGroupTask step
type CleanNodesInGroupTaskOption struct {
	NodeGroup *proto.NodeGroup
	Cluster   *proto.Cluster
	NodeIDs   []string
	NodeIPs   []string
	DeviceIDs []string
	Operator  string
}

// BuildRemoveNodesStep xxx
func (cn *CleanNodesInGroupTaskOption) BuildRemoveNodesStep(task *proto.Task) {
	removeNodesStep := cloudprovider.InitTaskStep(removeNodesFromClusterStep)

	removeNodesStep.Params[cloudprovider.ClusterIDKey.String()] = cn.NodeGroup.ClusterID
	removeNodesStep.Params[cloudprovider.NodeGroupIDKey.String()] = cn.NodeGroup.NodeGroupID
	// yunti cloud use cluster provider
	removeNodesStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider
	removeNodesStep.Params[cloudprovider.OperatorKey.String()] = cn.Operator

	removeNodesStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(cn.NodeIPs, ",")
	removeNodesStep.Params[cloudprovider.NodeIDsKey.String()] = strings.Join(cn.NodeIDs, ",")
	removeNodesStep.Params[cloudprovider.DeviceIDsKey.String()] = strings.Join(cn.DeviceIDs, ",")

	task.Steps[removeNodesFromClusterStep.StepMethod] = removeNodesStep
	task.StepSequence = append(task.StepSequence, removeNodesFromClusterStep.StepMethod)
}

// BuildReturnNodesStep xxx
func (cn *CleanNodesInGroupTaskOption) BuildReturnNodesStep(task *proto.Task) {
	returnNodesStep := cloudprovider.InitTaskStep(returnInstanceToResourcePoolStep)

	returnNodesStep.Params[cloudprovider.ClusterIDKey.String()] = cn.NodeGroup.ClusterID
	returnNodesStep.Params[cloudprovider.NodeGroupIDKey.String()] = cn.NodeGroup.NodeGroupID
	// yunti cloud use cluster provider
	returnNodesStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider
	returnNodesStep.Params[cloudprovider.OperatorKey.String()] = cn.Operator

	returnNodesStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(cn.NodeIPs, ",")
	returnNodesStep.Params[cloudprovider.NodeIDsKey.String()] = strings.Join(cn.NodeIDs, ",")
	returnNodesStep.Params[cloudprovider.DeviceIDsKey.String()] = strings.Join(cn.DeviceIDs, ",")

	task.Steps[returnInstanceToResourcePoolStep.StepMethod] = returnNodesStep
	task.StepSequence = append(task.StepSequence, returnInstanceToResourcePoolStep.StepMethod)
}

// BuildCordonNodesStep 设置节点不可调度状态
func (cn *CleanNodesInGroupTaskOption) BuildCordonNodesStep(task *proto.Task) {
	cordonStep := cloudprovider.InitTaskStep(common.CordonNodesActionStep)
	cordonStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID

	task.Steps[common.CordonNodesActionStep.StepMethod] = cordonStep
	task.StepSequence = append(task.StepSequence, common.CordonNodesActionStep.StepMethod)
}

// UpdateDesiredNodesTaskOption xxx
type UpdateDesiredNodesTaskOption struct {
	NodeGroup *proto.NodeGroup
	Cluster   *proto.Cluster
	Desired   int
	Operator  string
}

// BuildApplyInstanceStep xxx
func (ud *UpdateDesiredNodesTaskOption) BuildApplyInstanceStep(task *proto.Task) {
	applyInstanceStep := cloudprovider.InitTaskStep(applyCVMFromResourcePoolStep)

	applyInstanceStep.Params[cloudprovider.ClusterIDKey.String()] = ud.NodeGroup.ClusterID
	applyInstanceStep.Params[cloudprovider.NodeGroupIDKey.String()] = ud.NodeGroup.NodeGroupID
	// yunti cluster-manager by qcloud, thus use cluster provider
	applyInstanceStep.Params[cloudprovider.CloudIDKey.String()] = ud.Cluster.Provider
	applyInstanceStep.Params[cloudprovider.ScalingNodesNumKey.String()] = strconv.Itoa(ud.Desired)
	applyInstanceStep.Params[cloudprovider.OperatorKey.String()] = ud.Operator

	task.Steps[applyCVMFromResourcePoolStep.StepMethod] = applyInstanceStep
	task.StepSequence = append(task.StepSequence, applyCVMFromResourcePoolStep.StepMethod)
}

// BuildAddNodesToClusterStep xxx
func (ud *UpdateDesiredNodesTaskOption) BuildAddNodesToClusterStep(task *proto.Task) {
	newStepInfo := addNodesToClusterStep
	if ud.NodeGroup != nil && ud.NodeGroup.NodeTemplate != nil && len(ud.NodeGroup.NodeTemplate.PreStartUserScript) > 0 {
		newStepInfo.StepName += "(包含前置初始化)"
	}

	addNodesStep := cloudprovider.InitTaskStep(newStepInfo)
	addNodesStep.Params[cloudprovider.NodeGroupIDKey.String()] = ud.NodeGroup.NodeGroupID
	addNodesStep.Params[cloudprovider.CloudIDKey.String()] = ud.Cluster.Provider
	addNodesStep.Params[cloudprovider.ClusterIDKey.String()] = ud.NodeGroup.ClusterID
	addNodesStep.Params[cloudprovider.ScalingNodesNumKey.String()] = strconv.Itoa(ud.Desired)
	addNodesStep.Params[cloudprovider.OperatorKey.String()] = ud.Operator

	task.Steps[newStepInfo.StepMethod] = addNodesStep
	task.StepSequence = append(task.StepSequence, newStepInfo.StepMethod)
}

// BuildCheckClusterNodeStatusStep check cluster nodes status step
func (ud *UpdateDesiredNodesTaskOption) BuildCheckClusterNodeStatusStep(task *proto.Task) {
	checkNodeStep := cloudprovider.InitTaskStep(checkClusterNodeStatusStep)

	checkNodeStep.Params[cloudprovider.NodeGroupIDKey.String()] = ud.NodeGroup.NodeGroupID
	checkNodeStep.Params[cloudprovider.CloudIDKey.String()] = ud.Cluster.Provider
	checkNodeStep.Params[cloudprovider.ClusterIDKey.String()] = ud.NodeGroup.ClusterID
	checkNodeStep.Params[cloudprovider.OperatorKey.String()] = ud.Operator

	task.Steps[checkClusterNodeStatusStep.StepMethod] = checkNodeStep
	task.StepSequence = append(task.StepSequence, checkClusterNodeStatusStep.StepMethod)
}

// BuildSyncClusterNodesToCMDBStep sync cluster nodes to cmdb step
func (ud *UpdateDesiredNodesTaskOption) BuildSyncClusterNodesToCMDBStep(task *proto.Task) {
	syncCmdbStep := cloudprovider.InitTaskStep(syncClusterNodesToCMDBStep)

	syncCmdbStep.Params[cloudprovider.CloudIDKey.String()] = ud.Cluster.Provider
	syncCmdbStep.Params[cloudprovider.ClusterIDKey.String()] = ud.NodeGroup.ClusterID
	syncCmdbStep.Params[cloudprovider.NodeGroupIDKey.String()] = ud.NodeGroup.NodeGroupID
	syncCmdbStep.Params[cloudprovider.OperatorKey.String()] = ud.Operator

	task.Steps[syncClusterNodesToCMDBStep.StepMethod] = syncCmdbStep
	task.StepSequence = append(task.StepSequence, syncClusterNodesToCMDBStep.StepMethod)
}

// BuildNodeAnnotationsStep set node annotations
func (ud *UpdateDesiredNodesTaskOption) BuildNodeAnnotationsStep(task *proto.Task) {
	if ud.NodeGroup == nil || ud.NodeGroup.NodeTemplate == nil || len(ud.NodeGroup.NodeTemplate.Annotations) == 0 {
		return
	}
	common.BuildNodeAnnotationsTaskStep(task, ud.Cluster.ClusterID, nil, ud.NodeGroup.NodeTemplate.Annotations)
}

// BuildNodeCommonLabelsStep set node common labels
func (ud *UpdateDesiredNodesTaskOption) BuildNodeCommonLabelsStep(task *proto.Task) {
	common.BuildNodeLabelsTaskStep(task, ud.Cluster.ClusterID, nil, nil)
}

// BuildResourcePoolDeviceLabelStep set devices labels
func (ud *UpdateDesiredNodesTaskOption) BuildResourcePoolDeviceLabelStep(task *proto.Task) {
	common.BuildResourcePoolLabelTaskStep(task, ud.Cluster.ClusterID)
}

// BuildUnCordonNodesStep 设置节点可调度状态
func (ud *UpdateDesiredNodesTaskOption) BuildUnCordonNodesStep(task *proto.Task) {
	unCordonStep := cloudprovider.InitTaskStep(common.UnCordonNodesActionStep)

	unCordonStep.Params[cloudprovider.ClusterIDKey.String()] = ud.Cluster.ClusterID

	task.Steps[common.UnCordonNodesActionStep.StepMethod] = unCordonStep
	task.StepSequence = append(task.StepSequence, common.UnCordonNodesActionStep.StepMethod)
}

// CreateNodeGroupOption xxx
type CreateNodeGroupOption struct {
	NodeGroup    *proto.NodeGroup
	Cluster      *proto.Cluster
	PoolProvider string
	PoolID       string
}

// BuildCreateCloudNodeGroupStep xxx
func (cn *CreateNodeGroupOption) BuildCreateCloudNodeGroupStep(task *proto.Task) {
	createStep := cloudprovider.InitTaskStep(createNodePoolStep)

	createStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	createStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider
	createStep.Params[cloudprovider.NodeGroupIDKey.String()] = cn.NodeGroup.NodeGroupID
	createStep.Params[cloudprovider.PoolProvider.String()] = cn.PoolProvider
	createStep.Params[cloudprovider.PoolID.String()] = cn.PoolID

	task.Steps[createNodePoolStep.StepMethod] = createStep
	task.StepSequence = append(task.StepSequence, createNodePoolStep.StepMethod)
}

// DeleteNodeGroupOption xxx
type DeleteNodeGroupOption struct {
	NodeGroup *proto.NodeGroup
	Cluster   *proto.Cluster
	NodeIDs   []string
	NodeIPs   []string
	Operator  string
}

// BuildCheckCleanDBDataStep xxx
func (dn *DeleteNodeGroupOption) BuildCheckCleanDBDataStep(task *proto.Task) {
	checkStep := cloudprovider.InitTaskStep(checkCleanDBDataStep)

	checkStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(dn.NodeIPs, ",")
	checkStep.Params[cloudprovider.NodeGroupIDKey.String()] = dn.NodeGroup.NodeGroupID

	task.Steps[checkCleanDBDataStep.StepMethod] = checkStep
	task.StepSequence = append(task.StepSequence, checkCleanDBDataStep.StepMethod)
}
