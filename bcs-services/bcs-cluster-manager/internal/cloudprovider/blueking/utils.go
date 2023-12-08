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

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
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
)

var (
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
