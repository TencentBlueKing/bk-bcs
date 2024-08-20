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

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/common"
)

var (
	cloudName = "blueking"
)

const (
	// deleteClusterNodesTaskTemplate bk-sops delete clusterNodes task template
	deleteClusterNodesTaskTemplate = "blueking-remove nodes: %s"
	// addClusterNodesTaskTemplate bk-sops add clusterNodes task template
	addClusterNodesTaskTemplate = "blueking-add nodes: %s"
	// deleteClusterTaskTemplate bk-sops delete cluster task template
	deleteClusterTaskTemplate = "blueking-delete cluster: %s"
	// createClusterTaskTemplate bk-sops delete cluster task template
	createClusterTaskTemplate = "blueking-create cluster: %s"

	// createNodeGroupTaskTemplate bk-sops add task template
	createNodeGroupTaskTemplate = "blueking-create node group: %s/%s"
	// deleteNodeGroupTaskTemplate bk-sops add task template
	deleteNodeGroupTaskTemplate = "blueking-delete node group: %s/%s"
	// updateNodeGroupDesiredNodeTemplate bk-sops add task template
	updateNodeGroupDesiredNodeTemplate = "blueking-update node group desired node: %s/%s"
	// cleanNodeGroupNodesTaskTemplate bk-sops add task template
	cleanNodeGroupNodesTaskTemplate = "blueking-remove node group nodes: %s/%s"
	// updateNodeGroupTaskTemplate bk-sops add task template
	updateNodeGroupTaskTemplate = "blueking-update node group: %s/%s"
	// updateAutoScalingOptionTemplate bk-sops add task template
	updateAutoScalingOptionTemplate = "blueking-update auto scaling option: %s"
	// switchAutoScalingOptionStatusTemplate bk-sops add task template
	switchAutoScalingOptionStatusTemplate = "blueking-switch auto scaling option status: %s"
	// switchNodeGroupAutoScalingTaskTemplate bk-sops add task template
	switchNodeGroupAutoScalingTaskTemplate = "blueking-switch node group auto scaling: %s/%s"
)

var (
	// create cluster task steps
	createClusterStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CreateCluster", cloudName),
		StepName:   "创建集群",
	}

	// delete cluster task steps
	deleteClusterStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-DeleteCluster", cloudName),
		StepName:   "删除集群",
	}

	// cluster add nodes task steps
	addNodesToClusterStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-AddNodesToCluster", cloudName),
		StepName:   "集群上架节点",
	}

	// cluster remove nodes task steps
	removeNodesFromClusterStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-RemoveNodesFromCluster", cloudName),
		StepName:   "集群下架节点",
	}

	// import cluster task steps
	importClusterNodesStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-ImportClusterNodesTask", cloudName),
		StepName:   "导入集群节点",
	}

	// create cluster task steps
	updateCreateClusterDBInfoStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-UpdateCreateClusterDBInfoTask", cloudName),
		StepName:   "更新集群任务状态",
	}

	// delete cluster task steps
	cleanClusterDBInfoStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CleanClusterDBInfoTask", cloudName),
		StepName:   "清理集群数据",
	}

	// add cluster nodes task steps
	updateAddNodeDBInfoStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-UpdateAddNodeDBInfoTask", cloudName),
		StepName:   "更新任务状态",
	}

	// delete cluster nodes task steps
	updateRemoveNodeDBInfoStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-UpdateRemoveNodeDBInfoTask", cloudName),
		StepName:   "更新任务状态",
	}

	// create nodeGroup task: stepName and stepMethod
	createNodePoolStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CreateNodePoolTask", cloudName),
		StepName:   "创建资源池",
	}

	// delete nodeGroup task: stepName and stepMethod
	deleteNodePoolStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CheckCleanDBDataTask", cloudName),
		StepName:   "清理节点组数据",
	}

	// update desired nodes task
	applyNodesFromResourcePoolStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-%s", cloudName, cloudprovider.ApplyInstanceMachinesTask),
		StepName:   "申请节点任务",
	}

	// clean nodes in group task
	returnNodesToResourcePoolStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-ReturnNodesToResourcePoolTask", cloudName),
		StepName:   "回收节点",
	}
)

// CreateClusterTaskOption for build create cluster step
type CreateClusterTaskOption struct {
	Cluster     *proto.Cluster
	WorkerNodes []string
}

// BuildUpdateClusterDbInfoStep xxx
func (cn *CreateClusterTaskOption) BuildUpdateClusterDbInfoStep(task *proto.Task) {
	updateStep := cloudprovider.InitTaskStep(updateCreateClusterDBInfoStep)

	updateStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	updateStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider
	if len(cn.WorkerNodes) > 0 {
		updateStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(cn.WorkerNodes, ",")
	}

	task.Steps[updateCreateClusterDBInfoStep.StepMethod] = updateStep
	task.StepSequence = append(task.StepSequence, updateCreateClusterDBInfoStep.StepMethod)
}

// ImportClusterTaskOption for build import cluster step
type ImportClusterTaskOption struct {
	Cluster *proto.Cluster
}

// BuildImportClusterNodesStep xxx
func (in *ImportClusterTaskOption) BuildImportClusterNodesStep(task *proto.Task) {
	importNodesStep := cloudprovider.InitTaskStep(importClusterNodesStep)

	importNodesStep.Params[cloudprovider.ClusterIDKey.String()] = in.Cluster.ClusterID
	importNodesStep.Params[cloudprovider.CloudIDKey.String()] = in.Cluster.Provider

	task.Steps[importClusterNodesStep.StepMethod] = importNodesStep
	task.StepSequence = append(task.StepSequence, importClusterNodesStep.StepMethod)
}

// DeleteClusterTaskOption for build delete cluster step
type DeleteClusterTaskOption struct {
	Cluster *proto.Cluster
}

// BuildCleanClusterDbInfoStep xxx
func (dn *DeleteClusterTaskOption) BuildCleanClusterDbInfoStep(task *proto.Task) {
	cleanStep := cloudprovider.InitTaskStep(cleanClusterDBInfoStep)

	cleanStep.Params[cloudprovider.ClusterIDKey.String()] = dn.Cluster.ClusterID
	cleanStep.Params[cloudprovider.CloudIDKey.String()] = dn.Cluster.Provider

	task.Steps[cleanClusterDBInfoStep.StepMethod] = cleanStep
	task.StepSequence = append(task.StepSequence, cleanClusterDBInfoStep.StepMethod)
}

// AddNodesTaskOption for build add cluster nodes step
type AddNodesTaskOption struct {
	Cluster *proto.Cluster
	Cloud   *proto.Cloud
	NodeIps []string
}

// BuildUpdateAddNodeDbInfoStep xxx
func (an *AddNodesTaskOption) BuildUpdateAddNodeDbInfoStep(task *proto.Task) {
	updateStep := cloudprovider.InitTaskStep(updateAddNodeDBInfoStep)

	updateStep.Params[cloudprovider.ClusterIDKey.String()] = an.Cluster.ClusterID
	updateStep.Params[cloudprovider.CloudIDKey.String()] = an.Cloud.CloudID
	updateStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(an.NodeIps, ",")

	task.Steps[updateAddNodeDBInfoStep.StepMethod] = updateStep
	task.StepSequence = append(task.StepSequence, updateAddNodeDBInfoStep.StepMethod)
}

// RemoveNodesTaskOption for build remove cluster nodes step
type RemoveNodesTaskOption struct {
	Cluster *proto.Cluster
	NodeIps []string
}

// BuildUpdateRemoveNodeDbInfoStep xxx
func (dn *RemoveNodesTaskOption) BuildUpdateRemoveNodeDbInfoStep(task *proto.Task) {
	removeStep := cloudprovider.InitTaskStep(updateRemoveNodeDBInfoStep)

	removeStep.Params[cloudprovider.ClusterIDKey.String()] = dn.Cluster.ClusterID
	removeStep.Params[cloudprovider.CloudIDKey.String()] = dn.Cluster.Provider
	removeStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(dn.NodeIps, ",")

	task.Steps[updateRemoveNodeDBInfoStep.StepMethod] = removeStep
	task.StepSequence = append(task.StepSequence, updateRemoveNodeDBInfoStep.StepMethod)
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

// BuildDeleteCloudNodeGroupStep xxx
func (dn *DeleteNodeGroupOption) BuildDeleteCloudNodeGroupStep(task *proto.Task) {
	deleteStep := cloudprovider.InitTaskStep(deleteNodePoolStep)

	deleteStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(dn.NodeIPs, ",")
	deleteStep.Params[cloudprovider.NodeGroupIDKey.String()] = dn.NodeGroup.NodeGroupID

	task.Steps[deleteNodePoolStep.StepMethod] = deleteStep
	task.StepSequence = append(task.StepSequence, deleteNodePoolStep.StepMethod)
}

// UpdateDesiredNodesTaskOption 扩容节点组节点
type UpdateDesiredNodesTaskOption struct {
	Cluster  *proto.Cluster
	Group    *proto.NodeGroup
	Cloud    *proto.Cloud
	Desired  uint32
	Operator string
}

// BuildApplyInstanceStep 在资源池中申请节点实例
func (ud *UpdateDesiredNodesTaskOption) BuildApplyInstanceStep(task *proto.Task) {
	applyInstanceStep := cloudprovider.InitTaskStep(applyNodesFromResourcePoolStep)

	applyInstanceStep.Params[cloudprovider.ClusterIDKey.String()] = ud.Group.ClusterID
	applyInstanceStep.Params[cloudprovider.NodeGroupIDKey.String()] = ud.Group.NodeGroupID
	applyInstanceStep.Params[cloudprovider.CloudIDKey.String()] = ud.Cluster.Provider
	applyInstanceStep.Params[cloudprovider.ScalingNodesNumKey.String()] = strconv.Itoa(int(ud.Desired))
	applyInstanceStep.Params[cloudprovider.OperatorKey.String()] = ud.Operator

	task.Steps[applyNodesFromResourcePoolStep.StepMethod] = applyInstanceStep
	task.StepSequence = append(task.StepSequence, applyNodesFromResourcePoolStep.StepMethod)
}

// BuildNodeAnnotationsStep set node annotations
func (ud *UpdateDesiredNodesTaskOption) BuildNodeAnnotationsStep(task *proto.Task) {
	if ud.Group == nil || ud.Group.NodeTemplate == nil || len(ud.Group.NodeTemplate.Annotations) == 0 {
		return
	}
	common.BuildNodeAnnotationsTaskStep(task, ud.Cluster.ClusterID, nil, ud.Group.NodeTemplate.Annotations)
}

// BuildNodeCommonLabelsStep set node common labels
func (ud *UpdateDesiredNodesTaskOption) BuildNodeCommonLabelsStep(task *proto.Task) {
	common.BuildNodeLabelsTaskStep(task, ud.Cluster.ClusterID, nil, cloudprovider.GetLabelsByNg(ud.Group))
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

// CleanNodesInGroupTaskOption for build CleanNodesInGroupTask step
type CleanNodesInGroupTaskOption struct {
	Group     *proto.NodeGroup
	Cluster   *proto.Cluster
	NodeIDs   []string
	NodeIPs   []string
	DeviceIDs []string
	Operator  string
}

// BuildReturnNodesToResourcePoolStep 归还节点到资源池
func (cn *CleanNodesInGroupTaskOption) BuildReturnNodesToResourcePoolStep(task *proto.Task) {
	returnNodesStep := cloudprovider.InitTaskStep(returnNodesToResourcePoolStep)

	returnNodesStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Group.ClusterID
	returnNodesStep.Params[cloudprovider.NodeGroupIDKey.String()] = cn.Group.NodeGroupID
	returnNodesStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider
	returnNodesStep.Params[cloudprovider.OperatorKey.String()] = cn.Operator

	returnNodesStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(cn.NodeIPs, ",")
	returnNodesStep.Params[cloudprovider.NodeIDsKey.String()] = strings.Join(cn.NodeIDs, ",")
	returnNodesStep.Params[cloudprovider.DeviceIDsKey.String()] = strings.Join(cn.DeviceIDs, ",")

	task.Steps[returnNodesToResourcePoolStep.StepMethod] = returnNodesStep
	task.StepSequence = append(task.StepSequence, returnNodesToResourcePoolStep.StepMethod)
}
