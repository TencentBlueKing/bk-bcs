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

package qcloud

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

var (
	cloudName = "qcloud"
)

// qcloud taskName
const (
	// createClusterTaskTemplate bk-sops add task template
	createClusterTaskTemplate = "tke-create cluster: %s"
	// deleteClusterTaskTemplate bk-sops add task template
	deleteClusterTaskTemplate = "tke-delete cluster: %s"
	// tkeAddNodeTaskTemplate bk-sops add task template
	tkeAddNodeTaskTemplate = "tke-add node: %s"
	// tkeCleanNodeTaskTemplate bk-sops add task template
	tkeCleanNodeTaskTemplate = "tke-remove node: %s"
	// createNodeGroupTaskTemplate bk-sops add task template
	createNodeGroupTaskTemplate = "tke-create node group: %s/%s"
	// deleteNodeGroupTaskTemplate bk-sops add task template
	deleteNodeGroupTaskTemplate = "tke-delete node group: %s/%s"
	// updateNodeGroupDesiredNode bk-sops add task template
	updateNodeGroupDesiredNodeTemplate = "tke-update node group desired node: %s/%s"
	// cleanNodeGroupNodesTaskTemplate bk-sops add task template
	cleanNodeGroupNodesTaskTemplate = "tke-remove node group nodes: %s/%s"
	// moveNodesToNodeGroupTaskTemplate bk-sops add task template
	moveNodesToNodeGroupTaskTemplate = "tke-move nodes to node group: %s/%s"
	// switchNodeGroupAutoScalingTaskTemplate bk-sops add task template
	switchNodeGroupAutoScalingTaskTemplate = "tke-switch node group auto scaling: %s/%s"
)

var (
	// import cluster task
	importClusterNodesTask        = fmt.Sprintf("%s-ImportClusterNodesTask", cloudName)
	registerClusterKubeConfigTask = fmt.Sprintf("%s-RegisterClusterKubeConfigTask", cloudName)

	// create cluster task
	createClusterShieldAlarmTask  = fmt.Sprintf("%s-CreateClusterShieldAlarmTask", cloudName)
	createTKEClusterTask          = fmt.Sprintf("%s-CreateTKEClusterTask", cloudName)
	checkTKEClusterStatusTask     = fmt.Sprintf("%s-CheckTKEClusterStatusTask", cloudName)
	enableTkeClusterVpcCniTask    = fmt.Sprintf("%s-EnableTkeClusterVpcCniTask", cloudName)
	updateCreateClusterDBInfoTask = fmt.Sprintf("%s-UpdateCreateClusterDBInfoTask", cloudName)

	// delete cluster task
	deleteTKEClusterTask   = fmt.Sprintf("%s-DeleteTKEClusterTask", cloudName)
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
	deleteNodeGroupTask             = fmt.Sprintf("%s-DeleteNodeGroupTask", cloudName)
	uninstallAutoScalerTask         = fmt.Sprintf("%s-UninstallAutoScalerTask", cloudName)
	updateDeleteNodeGroupDBInfoTask = fmt.Sprintf("%s-UpdateDeleteNodeGroupDBInfoTask", cloudName)

	// clean node in nodeGroup task
	cleanNodeGroupNodesTask             = fmt.Sprintf("%s-CleanNodeGroupNodesTask", cloudName)
	checkCleanNodeGroupNodesStatusTask  = fmt.Sprintf("%s-CheckCleanNodeGroupNodesStatusTask", cloudName)
	updateCleanNodeGroupNodesDBInfoTask = fmt.Sprintf("%s-UpdateCleanNodeGroupNodesDBInfoTask", cloudName)

	// update desired nodes task
	applyInstanceMachinesTask    = fmt.Sprintf("%s-%s", cloudName, cloudprovider.ApplyInstanceMachinesTask)
	checkClusterNodesStatusTask  = fmt.Sprintf("%s-CheckClusterNodesStatusTask", cloudName)
	updateDesiredNodesDBInfoTask = fmt.Sprintf("%s-UpdateDesiredNodesDBInfoTask", cloudName)

	// auto scale task
	ensureAutoScalerTask             = fmt.Sprintf("%s-EnsureAutoScalerTask", cloudName)
	deleteAutoScalerTask             = fmt.Sprintf("%s-DeleteAutoScalerTask", cloudName)
	updateNodeGroupAutoScalingDBTask = fmt.Sprintf("%s-UpdateNodeGroupAutoScalingDBTask", cloudName)

	// move nodes to nodeGroup task
)
