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
 *
 */

package aws

import (
	"fmt"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

var (
	cloudName = "aws"
)

// awsCloud taskName
const (
	// importClusterTaskTemplate bk-sops add task template
	importClusterTaskTemplate = "aws-import cluster: %s"

	// createNodeGroupTaskTemplate bk-sops add task template
	createNodeGroupTaskTemplate = "aws-create node group: %s/%s"
	// deleteNodeGroupTaskTemplate bk-sops add task template
	deleteNodeGroupTaskTemplate = "aws-delete node group: %s/%s"
)

// tasks
var (
	// import cluster task
	importClusterNodesStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-ImportClusterNodesTask", cloudName),
		StepName:   "导入集群节点",
	}
	registerClusterKubeConfigStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-RegisterClusterKubeConfigTask", cloudName),
		StepName:   "注册集群kubeConfig认证",
	}

	// create nodeGroup task
	createCloudNodeGroupStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CreateCloudNodeGroupTask", cloudName),
		StepName:   "创建云节点组",
	}
	checkCloudNodeGroupStatusStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CheckCloudNodeGroupStatusTask", cloudName),
		StepName:   "检测云节点组状态",
	}

	// delete nodeGroup task
	deleteNodeGroupStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-DeleteNodeGroupTask", cloudName),
		StepName:   "删除云节点组",
	}
)

// ImportClusterTaskOption 纳管集群
type ImportClusterTaskOption struct {
	Cluster *proto.Cluster
}

// BuildRegisterKubeConfigStep 注册集群kubeConfig
func (ic *ImportClusterTaskOption) BuildRegisterKubeConfigStep(task *proto.Task) {
	registerKubeConfigStep := cloudprovider.InitTaskStep(registerClusterKubeConfigStep)
	registerKubeConfigStep.Params[cloudprovider.ClusterIDKey.String()] = ic.Cluster.ClusterID
	registerKubeConfigStep.Params[cloudprovider.CloudIDKey.String()] = ic.Cluster.Provider

	task.Steps[registerClusterKubeConfigStep.StepMethod] = registerKubeConfigStep
	task.StepSequence = append(task.StepSequence, registerClusterKubeConfigStep.StepMethod)
}

// BuildImportClusterNodesStep 纳管集群节点
func (ic *ImportClusterTaskOption) BuildImportClusterNodesStep(task *proto.Task) {
	importNodesStep := cloudprovider.InitTaskStep(importClusterNodesStep)
	importNodesStep.Params[cloudprovider.ClusterIDKey.String()] = ic.Cluster.ClusterID
	importNodesStep.Params[cloudprovider.CloudIDKey.String()] = ic.Cluster.Provider

	task.Steps[importClusterNodesStep.StepMethod] = importNodesStep
	task.StepSequence = append(task.StepSequence, importClusterNodesStep.StepMethod)
}

// CreateNodeGroupTaskOption 创建节点组
type CreateNodeGroupTaskOption struct {
	Group *proto.NodeGroup
}

// BuildCreateCloudNodeGroupStep 通过云接口创建节点组
func (cn *CreateNodeGroupTaskOption) BuildCreateCloudNodeGroupStep(task *proto.Task) {
	createStep := cloudprovider.InitTaskStep(createCloudNodeGroupStep)

	createStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Group.ClusterID
	createStep.Params[cloudprovider.NodeGroupIDKey.String()] = cn.Group.NodeGroupID
	createStep.Params[cloudprovider.CloudIDKey.String()] = cn.Group.Provider

	task.Steps[createCloudNodeGroupStep.StepMethod] = createStep
	task.StepSequence = append(task.StepSequence, createCloudNodeGroupStep.StepMethod)
}

// BuildCheckCloudNodeGroupStatusStep 检测节点组状态
func (cn *CreateNodeGroupTaskOption) BuildCheckCloudNodeGroupStatusStep(task *proto.Task) {
	checkStep := cloudprovider.InitTaskStep(checkCloudNodeGroupStatusStep)

	checkStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Group.ClusterID
	checkStep.Params[cloudprovider.NodeGroupIDKey.String()] = cn.Group.NodeGroupID
	checkStep.Params[cloudprovider.CloudIDKey.String()] = cn.Group.Provider

	task.Steps[checkCloudNodeGroupStatusStep.StepMethod] = checkStep
	task.StepSequence = append(task.StepSequence, checkCloudNodeGroupStatusStep.StepMethod)
}

// DeleteNodeGroupTaskOption 删除节点组
type DeleteNodeGroupTaskOption struct {
	Group *proto.NodeGroup
}

// BuildDeleteNodeGroupStep 删除云节点组
func (dn *DeleteNodeGroupTaskOption) BuildDeleteNodeGroupStep(task *proto.Task) {
	deleteStep := cloudprovider.InitTaskStep(deleteNodeGroupStep)

	deleteStep.Params[cloudprovider.ClusterIDKey.String()] = dn.Group.ClusterID
	deleteStep.Params[cloudprovider.NodeGroupIDKey.String()] = dn.Group.NodeGroupID
	deleteStep.Params[cloudprovider.CloudIDKey.String()] = dn.Group.Provider

	task.Steps[deleteNodeGroupStep.StepMethod] = deleteStep
	task.StepSequence = append(task.StepSequence, deleteNodeGroupStep.StepMethod)
}
