/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package eop

import (
	"fmt"
	"strings"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

var (
	cloudName = "eopCloud"
)

const (
	// createClusterTaskTemplate bk-sops add task template
	createClusterTaskTemplate = "eck-create cluster: %s"
	// deleteClusterTaskTemplate bk-sops add task template
	deleteClusterTaskTemplate = "eck-delete cluster: %s"
)

var (
	createECKClusterStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CreateECKClusterTask", cloudName),
		StepName:   "创建集群",
	}
	checkECKClusterStatusStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CheckECKClusterStatusTask", cloudName),
		StepName:   "检测集群状态",
	}
	checkECKNodeGroupsStatusStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CheckECKNodeGroupsStatusTask", cloudName),
		StepName:   "检测集群节点池状态",
	}
	updateECKNodeGroupsToDBStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-UpdateECKNodesGroupToDBTask", cloudName),
		StepName:   "更新节点池信息",
	}
	checkCreateClusterNodeStatusStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CheckCreateClusterNodeStatusStep", cloudName),
		StepName:   "检测集群节点状态",
	}
	updateECKNodesToDBStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-UpdateECKNodesToDBTask", cloudName),
		StepName:   "更新节点信息",
	}
	registerManageClusterKubeConfigStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-RegisterManageClusterKubeConfigTask", cloudName),
		StepName:   "注册集群连接信息",
	}

	// delete cluster task
	deleteECKClusterStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-DeleteECKClusterTask", cloudName),
		StepName:   "删除集群",
	}
	cleanClusterDBInfoStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CleanClusterDBInfoTask", cloudName),
		StepName:   "清理集群数据",
	}
)

// CreateClusterTaskOption 创建集群构建step子任务
type CreateClusterTaskOption struct {
	Cluster      *proto.Cluster
	NodeGroupIDs []string
}

// BuildCreateClusterStep 创建集群任务
func (cn *CreateClusterTaskOption) BuildCreateClusterStep(task *proto.Task) {
	createStep := cloudprovider.InitTaskStep(createECKClusterStep)
	createStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	createStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider
	createStep.Params[cloudprovider.NodeGroupIDKey.String()] = strings.Join(cn.NodeGroupIDs, ",")

	task.Steps[createECKClusterStep.StepMethod] = createStep
	task.StepSequence = append(task.StepSequence, createECKClusterStep.StepMethod)
}

// BuildCheckClusterStatusStep 检测集群状态任务
func (cn *CreateClusterTaskOption) BuildCheckClusterStatusStep(task *proto.Task) {
	checkStep := cloudprovider.InitTaskStep(checkECKClusterStatusStep)
	checkStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	checkStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider

	task.Steps[checkECKClusterStatusStep.StepMethod] = checkStep
	task.StepSequence = append(task.StepSequence, checkECKClusterStatusStep.StepMethod)
}

// BuildCheckNodeGroupsStatusStep 检测集群节点池状态任务
func (cn *CreateClusterTaskOption) BuildCheckNodeGroupsStatusStep(task *proto.Task) {
	checkStep := cloudprovider.InitTaskStep(checkECKNodeGroupsStatusStep)
	checkStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	checkStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider
	checkStep.Params[cloudprovider.NodeGroupIDKey.String()] = strings.Join(cn.NodeGroupIDs, ",")

	task.Steps[checkECKNodeGroupsStatusStep.StepMethod] = checkStep
	task.StepSequence = append(task.StepSequence, checkECKNodeGroupsStatusStep.StepMethod)
}

// BuildUpdateNodeGroupsToDBStep 更新集群节点池信息任务
func (cn *CreateClusterTaskOption) BuildUpdateNodeGroupsToDBStep(task *proto.Task) {
	updateStep := cloudprovider.InitTaskStep(updateECKNodeGroupsToDBStep)
	updateStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	updateStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider

	task.Steps[updateECKNodeGroupsToDBStep.StepMethod] = updateStep
	task.StepSequence = append(task.StepSequence, updateECKNodeGroupsToDBStep.StepMethod)
}

// BuildCheckClusterNodesStatusStep 检测集群节点状态任务
func (cn *CreateClusterTaskOption) BuildCheckClusterNodesStatusStep(task *proto.Task) {
	checkStep := cloudprovider.InitTaskStep(checkCreateClusterNodeStatusStep)
	checkStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	checkStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider

	task.Steps[checkCreateClusterNodeStatusStep.StepMethod] = checkStep
	task.StepSequence = append(task.StepSequence, checkCreateClusterNodeStatusStep.StepMethod)
}

// BuildUpdateNodesToDBStep 更新集群节点信息任务
func (cn *CreateClusterTaskOption) BuildUpdateNodesToDBStep(task *proto.Task) {
	updateStep := cloudprovider.InitTaskStep(updateECKNodesToDBStep)
	updateStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	updateStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider

	task.Steps[updateECKNodesToDBStep.StepMethod] = updateStep
	task.StepSequence = append(task.StepSequence, updateECKNodesToDBStep.StepMethod)
}

// BuildRegisterClsKubeConfigStep 注册集群连接信息
func (cn *CreateClusterTaskOption) BuildRegisterClsKubeConfigStep(task *proto.Task) {
	registerStep := cloudprovider.InitTaskStep(registerManageClusterKubeConfigStep)
	registerStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	registerStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider

	task.Steps[registerManageClusterKubeConfigStep.StepMethod] = registerStep
	task.StepSequence = append(task.StepSequence, registerManageClusterKubeConfigStep.StepMethod)
}

// DeleteClusterTaskOption 删除集群
type DeleteClusterTaskOption struct {
	Cluster *proto.Cluster
}

// BuildDeleteECKClusterStep 删除集群
func (dc *DeleteClusterTaskOption) BuildDeleteECKClusterStep(task *proto.Task) {
	deleteStep := cloudprovider.InitTaskStep(deleteECKClusterStep)
	deleteStep.Params[cloudprovider.ClusterIDKey.String()] = dc.Cluster.ClusterID
	deleteStep.Params[cloudprovider.CloudIDKey.String()] = dc.Cluster.Provider

	task.Steps[deleteECKClusterStep.StepMethod] = deleteStep
	task.StepSequence = append(task.StepSequence, deleteECKClusterStep.StepMethod)
}

// BuildCleanClusterDBInfoStep 清理集群数据
func (dc *DeleteClusterTaskOption) BuildCleanClusterDBInfoStep(task *proto.Task) {
	updateStep := cloudprovider.InitTaskStep(cleanClusterDBInfoStep)
	updateStep.Params[cloudprovider.ClusterIDKey.String()] = dc.Cluster.ClusterID
	updateStep.Params[cloudprovider.CloudIDKey.String()] = dc.Cluster.Provider

	task.Steps[cleanClusterDBInfoStep.StepMethod] = updateStep
	task.StepSequence = append(task.StepSequence, cleanClusterDBInfoStep.StepMethod)
}
