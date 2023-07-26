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

package huawei

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

var (
	cloudName = "huawei"
)

// googleCloud taskName
const (
	// createNodeGroupTaskTemplate bk-sops add task template
	createNodeGroupTaskTemplate = "cce-create node group: %s/%s"
	// deleteNodeGroupTaskTemplate bk-sops add task template
	deleteNodeGroupTaskTemplate = "cce-delete node group: %s/%s"
	// cleanNodeGroupNodesTaskTemplate bk-sops add task template
	cleanNodeGroupNodesTaskTemplate = "cce-remove node group nodes: %s/%s"
	// updateNodeGroupDesiredNode bk-sops add task template
	updateNodeGroupDesiredNodeTemplate = "cce-update node group desired node: %s/%s"
	// switchNodeGroupAutoScalingTaskTemplate bk-sops add task template
	switchNodeGroupAutoScalingTaskTemplate = "cce-switch node group auto scaling: %s/%s"
	// updateAutoScalingOptionTemplate bk-sops add task template
	updateAutoScalingOptionTemplate = "cce-update auto scaling option: %s"
	// switchAutoScalingOptionStatusTemplate bk-sops add task template
	switchAutoScalingOptionStatusTemplate = "cce-switch auto scaling option status: %s"
)

// tasks
var (
	// import cluster task
	importClusterNodesTask        = fmt.Sprintf("%s-ImportClusterNodesTask", cloudName)
	registerClusterKubeConfigTask = fmt.Sprintf("%s-RegisterClusterKubeConfigTask", cloudName)

	// create nodeGroup task
	createCloudNodeGroupTask      = fmt.Sprintf("%s-CreateCloudNodeGroupTask", cloudName)
	checkCloudNodeGroupStatusTask = fmt.Sprintf("%s-CheckCloudNodeGroupStatusTask", cloudName)

	// delete nodeGroup task
	deleteNodeGroupTask = fmt.Sprintf("%s-DeleteNodeGroupTask", cloudName)

	// clean node in nodeGroup task
	cleanNodeGroupNodesTask = fmt.Sprintf("%s-CleanNodeGroupNodesTask", cloudName)

	// auto scale task
	ensureAutoScalerTask  = fmt.Sprintf("%s-EnsureAutoScalerTask", cloudName)
	ensureAutoScalingTask = fmt.Sprintf("%s-EnsureAutoScalingTask", cloudName)
	// update desired nodes task
	applyInstanceMachinesTask   = fmt.Sprintf("%s-%s", cloudName, cloudprovider.ApplyInstanceMachinesTask)
	checkClusterNodesStatusTask = fmt.Sprintf("%s-CheckClusterNodesStatusTask", cloudName)
)
