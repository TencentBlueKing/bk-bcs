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

package azure

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/pkg/errors"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

var (
	cloudName = "azure"
)

// task template name
const (
	// deleteClusterTaskTemplate bk-sops add task template
	deleteClusterTaskTemplate = "aks-delete cluster: %s"
	// createNodeGroupTaskTemplate bk-sops add task template
	createNodeGroupTaskTemplate = "aks-create node group: %s/%s"
	// switchNodeGroupAutoScalingTaskTemplate bk-sops add task template
	switchNodeGroupAutoScalingTaskTemplate = "aks-switch node group auto scaling: %s/%s"
	// deleteNodeGroupTaskTemplate bk-sops add task template
	deleteNodeGroupTaskTemplate = "aks-delete node group: %s/%s"
	// updateNodeGroupDesiredNode bk-sops add task template
	updateNodeGroupDesiredNodeTemplate = "aks-update node group desired node: %s/%s"
	// updateAutoScalingOptionTemplate bk-sops add task template
	updateAutoScalingOptionTemplate = "aks-update auto scaling option: %s"
	// cleanNodeGroupNodesTaskTemplate bk-sops add task template
	cleanNodeGroupNodesTaskTemplate = "aks-remove node group nodes: %s/%s"
	// switchAutoScalingOptionStatusTemplate bk-sops add task template
	switchAutoScalingOptionStatusTemplate = "aks-switch auto scaling option status: %s"
)

// tasks
var (
	// import cluster task
	importClusterNodesTask        = fmt.Sprintf("%s-ImportClusterNodesTask", cloudName)
	registerClusterKubeConfigTask = fmt.Sprintf("%s-RegisterClusterKubeConfigTask", cloudName)

	// delete cluster task
	deleteAKSKEClusterTask = fmt.Sprintf("%s-deleteAKSKEClusterTask", cloudName)
	cleanClusterDBInfoTask = fmt.Sprintf("%s-CleanClusterDBInfoTask", cloudName)

	// create nodeGroup task
	createCloudNodeGroupTask        = fmt.Sprintf("%s-CreateCloudNodeGroupTask", cloudName)
	checkCloudNodeGroupStatusTask   = fmt.Sprintf("%s-CheckCloudNodeGroupStatusTask", cloudName)
	updateCreateNodeGroupDBInfoTask = fmt.Sprintf("%s-UpdateCreateNodeGroupDBInfoTask", cloudName)

	// delete nodeGroup task
	deleteNodeGroupTask = fmt.Sprintf("%s-DeleteNodeGroupTask", cloudName)

	// clean node in nodeGroup task
	cleanNodeGroupNodesTask             = fmt.Sprintf("%s-CleanNodeGroupNodesTask", cloudName)
	removeHostFromCMDBTask              = fmt.Sprintf("%s-RemoveHostFromCMDBTask", cloudName)
	checkCleanNodeGroupNodesStatusTask  = fmt.Sprintf("%s-CheckCleanNodeGroupNodesStatusTask", cloudName)
	updateCleanNodeGroupNodesDBInfoTask = fmt.Sprintf("%s-UpdateCleanNodeGroupNodesDBInfoTask", cloudName)

	// update desired nodes task
	applyInstanceMachinesTask    = fmt.Sprintf("%s-%s", cloudName, cloudprovider.ApplyInstanceMachinesTask)
	checkClusterNodesStatusTask  = fmt.Sprintf("%s-CheckClusterNodesStatusTask", cloudName)
	installGSEAgentTask          = fmt.Sprintf("%s-InstallGSEAgentTask", cloudName)
	transferHostModuleTask       = fmt.Sprintf("%s-TransferHostModuleTask", cloudName)
	updateDesiredNodesDBInfoTask = fmt.Sprintf("%s-UpdateDesiredNodesDBInfoTask", cloudName)

	// auto scale task
	ensureAutoScalerTask = fmt.Sprintf("%s-EnsureAutoScalerTask", cloudName)
)

// errors
var (
	nodePoolScaleUpErr  = errors.New("the status of the aks node pool is scale up, and currently no operations can be performed on it")
	nodePoolUpdatingErr = errors.New("the aks node pool status is in the process of being updating and no operations can be performed on it right now")
)

// transAksNodeToNode 节点转换
func transAksNodeToNode(node *armcompute.VirtualMachineScaleSetVM, vmIPMap map[string][]string) *proto.NodeGroupNode {
	n := &proto.NodeGroupNode{NodeID: *node.InstanceID}
	// azure 默认为节点，无法获取master
	properties := node.Properties
	if properties != nil && properties.ProvisioningState != nil {
		switch *properties.ProvisioningState {
		case api.NormalState:
			n.Status = common.StatusRunning
		case api.CreatingState:
			n.Status = common.StatusInitialization
		//case "failed":
		//	n.Status = "FAILED"
		default:
			n.Status = *properties.ProvisioningState
		}
	}
	if list, ok := vmIPMap[*node.Name]; ok && len(list) != 0 {
		n.InnerIP = list[0]
	}
	return n
}

// updateAgentPoolProperties 更新 AKS 代理节点池 - update agent pool
func (ng *NodeGroup) updateAgentPoolProperties(client api.AksService, cluster *proto.Cluster,
	group *proto.NodeGroup) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := client.GetPoolAndReturn(ctx, cluster.SystemID, group.CloudNodeGroupID)
	if err != nil {
		return errors.Wrapf(err, "UpdateNodeGroup: call GetAgentPool api failed")
	}
	if err = checkPoolState(pool); err != nil { // 更新前检查节点池的状态
		return errors.Wrapf(err, "nodeGroupID: %s unable to update agent pool", group.NodeGroupID)
	}

	// 更新 pool
	api.SetAgentPoolFromNodeGroup(group, pool)

	// update agent pool
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if _, err = client.UpdatePoolAndReturn(ctx, pool, cluster.SystemID, *pool.Name); err != nil {
		return errors.Wrapf(err, "UpdateNodeGroup: call UpdateAgentPool api failed")
	}

	return nil
}

// updateVMSSProperties 更新虚拟机规模集 - update virtual machine scale set
func (ng *NodeGroup) updateVMSSProperties(client api.AksService, group *proto.NodeGroup) error {
	asg := group.AutoScaling
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	set, err := client.GetSetWithName(ctx, asg.AutoScalingName, asg.AutoScalingID)
	if err != nil {
		return errors.Wrapf(err, "UpdateNodeGroup: call GetSetWithName api failed")
	}

	if group.LaunchTemplate != nil && len(group.LaunchTemplate.UserData) != 0 {
		set.Properties.VirtualMachineProfile.UserData = to.Ptr(group.LaunchTemplate.UserData)
	}
	// 镜像引用-暂时置空处理，若不置空会导致无法更新set
	api.SetImageReferenceNull(set)

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if _, err = client.UpdateSetWithName(ctx, set, asg.AutoScalingName, asg.AutoScalingID); err != nil {
		return errors.Wrapf(err, "UpdateNodeGroup: call UpdateSetWithName api failed")
	}

	return nil
}

// scaleUpPreCheck 扩容前置检查
func (ng *NodeGroup) scaleUpPreCheck(clusterID, cloudID, nodeGroupID string) error {
	info, err := cloudprovider.GetClusterDependBasicInfo(clusterID, cloudID, nodeGroupID)
	if err != nil {
		return errors.Wrapf(err, "call GetClusterDependBasicInfo failed")
	}

	client, err := api.NewAksServiceImplWithCommonOption(info.CmOption)
	if err != nil {
		return errors.Wrapf(err, "new azure client failed")
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()
	pool, err := client.GetPoolAndReturn(ctx, info.Cluster.SystemID, info.NodeGroup.CloudNodeGroupID)
	if err != nil {
		return errors.Wrapf(err, "call GetPoolAndReturn failed")
	}

	return checkPoolState(pool)
}

// checkPoolState 更新前，检查节点池的状态
// 如果节点池正在 "更新中" 或 "扩容中"，将无法对其进行操作
func checkPoolState(pool *armcontainerservice.AgentPool) error {
	state := *pool.Properties.ProvisioningState
	if state == api.UpdatingState {
		return errors.Wrapf(nodePoolUpdatingErr, "cloudNodeGroupID: %s", *pool.Name)
	}
	if state == api.ScalingState {
		return errors.Wrapf(nodePoolScaleUpErr, "cloudNodeGroupID: %s", *pool.Name)
	}
	return nil
}
