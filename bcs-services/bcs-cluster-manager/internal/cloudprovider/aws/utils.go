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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

var (
	cloudName = "aws"
)

// aws taskName
const (
	// createClusterTaskTemplate bk-sops add task template
	createClusterTaskTemplate = "eks-create cluster: %s"
	// deleteClusterTaskTemplate bk-sops add task template
	deleteClusterTaskTemplate = "eks-delete cluster: %s"
	// eksAddNodeTaskTemplate bk-sops add task template
	eksAddNodeTaskTemplate = "eks-add node: %s"
	// eksCleanNodeTaskTemplate bk-sops add task template
	eksCleanNodeTaskTemplate = "eks-remove node: %s"
	// createNodeGroupTaskTemplate bk-sops add task template
	createNodeGroupTaskTemplate = "eks-create node group: %s/%s"
	// deleteNodeGroupTaskTemplate bk-sops add task template
	deleteNodeGroupTaskTemplate = "eks-delete node group: %s/%s"
	// updateNodeGroupDesiredNode bk-sops add task template
	updateNodeGroupDesiredNodeTemplate = "eks-update node group desired node: %s/%s"
	// cleanNodeGroupNodesTaskTemplate bk-sops add task template
	cleanNodeGroupNodesTaskTemplate = "eks-remove node group nodes: %s/%s"
	// moveNodesToNodeGroupTaskTemplate bk-sops add task template
	moveNodesToNodeGroupTaskTemplate = "eks-move nodes to node group: %s/%s"
	// switchNodeGroupAutoScalingTaskTemplate bk-sops add task template
	switchNodeGroupAutoScalingTaskTemplate = "eks-switch node group auto scaling: %s/%s"
	// updateAutoScalingOptionTemplate bk-sops add task template
	updateAutoScalingOptionTemplate = "eks-update auto scaling option: %s"
	// switchAutoScalingOptionStatusTemplate bk-sops add task template
	switchAutoScalingOptionStatusTemplate = "eks-switch auto scaling option status: %s"
)

var (
	// import cluster task
	importClusterNodesTask        = fmt.Sprintf("%s-ImportClusterNodesTask", cloudName)
	registerClusterKubeConfigTask = fmt.Sprintf("%s-RegisterClusterKubeConfigTask", cloudName)
	installWatchComponentTask     = fmt.Sprintf("%s-InstallWatchComponentTask", cloudName)

	// create cluster task
	createClusterShieldAlarmTask  = fmt.Sprintf("%s-CreateClusterShieldAlarmTask", cloudName)
	createEKSClusterTask          = fmt.Sprintf("%s-CreateEKSClusterTask", cloudName)
	checkEKSClusterStatusTask     = fmt.Sprintf("%s-CheckEKSClusterStatusTask", cloudName)
	enableEKSClusterVpcCniTask    = fmt.Sprintf("%s-EnableEKSClusterVpcCniTask", cloudName)
	updateCreateClusterDBInfoTask = fmt.Sprintf("%s-UpdateCreateClusterDBInfoTask", cloudName)

	// delete cluster task
	deleteEKSClusterTask   = fmt.Sprintf("%s-DeleteEKSClusterTask", cloudName)
	cleanClusterDBInfoTask = fmt.Sprintf("%s-CleanClusterDBInfoTask", cloudName)

	// add node to cluster
	addNodesShieldAlarmTask = fmt.Sprintf("%s-AddNodesShieldAlarmTask", cloudName)
	addNodesToClusterTask   = fmt.Sprintf("%s-AddNodesToClusterTask", cloudName)
	checkAddNodesStatusTask = fmt.Sprintf("%s-CheckAddNodesStatusTask", cloudName)
	updateAddNodeDBInfoTask = fmt.Sprintf("%s-UpdateAddNodeDBInfoTask", cloudName)

	// remove node from cluster
	removeNodesFromClusterTask = fmt.Sprintf("%s-RemoveNodesFromClusterTask", cloudName)
	updateRemoveNodeDBInfoTask = fmt.Sprintf("%s-UpdateRemoveNodeDBInfoTask", cloudName)

	// create nodeGroup task
	createCloudNodeGroupTask        = fmt.Sprintf("%s-CreateCloudNodeGroupTask", cloudName)
	checkCloudNodeGroupStatusTask   = fmt.Sprintf("%s-CheckCloudNodeGroupStatusTask", cloudName)
	updateCreateNodeGroupDBInfoTask = fmt.Sprintf("%s-UpdateCreateNodeGroupDBInfoTask", cloudName)

	// delete nodeGroup task
	deleteNodeGroupTask = fmt.Sprintf("%s-DeleteNodeGroupTask", cloudName)

	// clean node in nodeGroup task
	cleanNodeGroupNodesTask             = fmt.Sprintf("%s-CleanNodeGroupNodesTask", cloudName)
	checkCleanNodeGroupNodesStatusTask  = fmt.Sprintf("%s-CheckCleanNodeGroupNodesStatusTask", cloudName)
	updateCleanNodeGroupNodesDBInfoTask = fmt.Sprintf("%s-UpdateCleanNodeGroupNodesDBInfoTask", cloudName)

	// update desired nodes task
	applyInstanceMachinesTask    = fmt.Sprintf("%s-%s", cloudName, cloudprovider.ApplyInstanceMachinesTask)
	checkClusterNodesStatusTask  = fmt.Sprintf("%s-CheckClusterNodesStatusTask", cloudName)
	updateDesiredNodesDBInfoTask = fmt.Sprintf("%s-UpdateDesiredNodesDBInfoTask", cloudName)

	// auto scale task
	ensureAutoScalerTask = fmt.Sprintf("%s-EnsureAutoScalerTask", cloudName)
)
