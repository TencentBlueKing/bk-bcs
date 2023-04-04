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

package google

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

var (
	cloudName = "google"
)

// googleCloud taskName
const (
	// deleteClusterTaskTemplate bk-sops add task template
	deleteClusterTaskTemplate = "gke-delete cluster: %s"
	// gkeAddNodeTaskTemplate bk-sops add task template
	gkeAddNodeTaskTemplate = "gke-add node: %s"
	// gkeCleanNodeTaskTemplate bk-sops add task template
	gkeCleanNodeTaskTemplate = "gke-remove node: %s"
	// createNodeGroupTaskTemplate bk-sops add task template
	createNodeGroupTaskTemplate = "gke-create node group: %s/%s"
	// deleteNodeGroupTaskTemplate bk-sops add task template
	deleteNodeGroupTaskTemplate = "gke-delete node group: %s/%s"
	// updateNodeGroupDesiredNode bk-sops add task template
	updateNodeGroupDesiredNodeTemplate = "gke-update node group desired node: %s/%s"
	// cleanNodeGroupNodesTaskTemplate bk-sops add task template
	cleanNodeGroupNodesTaskTemplate = "gke-remove node group nodes: %s/%s"
	// moveNodesToNodeGroupTaskTemplate bk-sops add task template
	moveNodesToNodeGroupTaskTemplate = "gke-move nodes to node group: %s/%s"
	// switchNodeGroupAutoScalingTaskTemplate bk-sops add task template
	switchNodeGroupAutoScalingTaskTemplate = "gke-switch node group auto scaling: %s/%s"
	// updateAutoScalingOptionTemplate bk-sops add task template
	updateAutoScalingOptionTemplate = "gke-update auto scaling option: %s"
	// switchAutoScalingOptionStatusTemplate bk-sops add task template
	switchAutoScalingOptionStatusTemplate = "gke-switch auto scaling option status: %s"
)

// tasks
var (
	// import cluster task
	importClusterNodesTask        = fmt.Sprintf("%s-ImportClusterNodesTask", cloudName)
	registerClusterKubeConfigTask = fmt.Sprintf("%s-RegisterClusterKubeConfigTask", cloudName)

	// delete cluster task
	deleteGKEClusterTask   = fmt.Sprintf("%s-deleteGKEClusterTask", cloudName)
	cleanClusterDBInfoTask = fmt.Sprintf("%s-CleanClusterDBInfoTask", cloudName)

	// create nodeGroup task
	createCloudNodeGroupTask      = fmt.Sprintf("%s-CreateCloudNodeGroupTask", cloudName)
	checkCloudNodeGroupStatusTask = fmt.Sprintf("%s-CheckCloudNodeGroupStatusTask", cloudName)

	// delete nodeGroup task
	deleteNodeGroupTask = fmt.Sprintf("%s-DeleteNodeGroupTask", cloudName)

	// auto scale task
	ensureAutoScalerTask = fmt.Sprintf("%s-EnsureAutoScalerTask", cloudName)

	// update desired nodes task
	applyInstanceMachinesTask    = fmt.Sprintf("%s-%s", cloudName, cloudprovider.ApplyInstanceMachinesTask)
	checkClusterNodesStatusTask  = fmt.Sprintf("%s-CheckClusterNodesStatusTask", cloudName)
	updateDesiredNodesDBInfoTask = fmt.Sprintf("%s-UpdateDesiredNodesDBInfoTask", cloudName)

	// clean node in nodeGroup task
	cleanNodeGroupNodesTask             = fmt.Sprintf("%s-CleanNodeGroupNodesTask", cloudName)
	checkCleanNodeGroupNodesStatusTask  = fmt.Sprintf("%s-CheckCleanNodeGroupNodesStatusTask", cloudName)
	updateCleanNodeGroupNodesDBInfoTask = fmt.Sprintf("%s-UpdateCleanNodeGroupNodesDBInfoTask", cloudName)
)
