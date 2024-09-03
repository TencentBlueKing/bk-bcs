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

package tasks

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	cmutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// ApplyInstanceMachinesTask 扩容节点 - update desired nodes task
func ApplyInstanceMachinesTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		return nil
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// extract parameter && check validate
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	desiredNodes := step.Params[cloudprovider.ScalingNodesNumKey.String()]
	nodeNum, _ := strconv.Atoi(desiredNodes)

	manual := state.Task.CommonParams[cloudprovider.ManualKey.String()]

	if len(clusterID) == 0 || len(nodeGroupID) == 0 || len(cloudID) == 0 || len(desiredNodes) == 0 {
		blog.Errorf("ApplyInstanceMachinesTask[%s]: check parameter validate failed", taskID)
		retErr := fmt.Errorf("ApplyInstanceMachinesTask check parameters failed")
		_ = cloudprovider.DeleteVirtualNodes(clusterID, nodeGroupID, taskID)
		_ = cloudprovider.UpdateNodeGroupDesiredSize(nodeGroupID, nodeNum, true)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("ApplyInstanceMachinesTask[%s]: call GetClusterDependBasicInfo failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("ApplyInstanceMachinesTask call GetClusterDependBasicInfo failed")
		if manual == common.True {
			_ = cloudprovider.UpdateVirtualNodeStatus(clusterID, nodeGroupID, taskID)
		} else {
			_ = cloudprovider.UpdateNodeGroupDesiredSize(nodeGroupID, nodeNum, true)
		}
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	err = applyInstanceMachines(ctx, dependInfo, int64(nodeNum))
	if err != nil {
		blog.Errorf("ApplyInstanceMachinesTask[%s]: applyInstanceMachines failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("ApplyInstanceMachinesTask applyInstanceMachines failed")

		if manual == common.True {
			_ = cloudprovider.UpdateVirtualNodeStatus(clusterID, nodeGroupID, taskID)
		} else {
			_ = cloudprovider.UpdateNodeGroupDesiredSize(nodeGroupID, nodeNum, true)
		}
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// trans success nodes to cm DB and record common paras, not handle error
	err = recordClusterInstanceToDB(ctx, state, dependInfo, uint64(nodeNum))
	if err != nil {
		blog.Errorf("ApplyInstanceMachinesTask[%s]: recordClusterInstanceToDB failed: %s",
			taskID, err.Error())
		retErr := fmt.Errorf("ApplyInstanceMachinesTask applyInstanceMachines failed %s", err.Error())
		if manual == common.True {
			_ = cloudprovider.UpdateVirtualNodeStatus(clusterID, nodeGroupID, taskID)
		} else {
			_ = cloudprovider.UpdateNodeGroupDesiredSize(nodeGroupID, nodeNum, true)
		}
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// destroy virtual nodes
	if manual == common.True {
		blog.Infof("ApplyInstanceMachinesTask[%s] begin DeleteVirtualNodes", taskID)
		_ = cloudprovider.DeleteVirtualNodes(clusterID, nodeGroupID, taskID)
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		return errors.Wrapf(err, "ApplyInstanceMachinesTask[%s] task %s %s update to storage fatal", taskID,
			taskID, stepName)
	}
	return nil
}

func applyInstanceMachines(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, nodeNum int64) error {
	taskId := cloudprovider.GetTaskIDFromContext(ctx)

	client, err := api.NewAksServiceImplWithCommonOption(info.CmOption) // new client
	if err != nil {
		blog.Errorf("applyInstanceMachines[%s] init azure client failed", taskId)
		return errors.Wrapf(err, "new azure client failed")
	}

	// pre check agentPool & vmScaleSet
	agentPool, err := preCheckAgentPool(ctx, client, info) // 前置检查
	if err != nil {
		blog.Errorf("applyInstanceMachines[%s] preCheckAgentPool failed: %v", taskId, err)
		return err
	}

	// scale nodes to desired size
	agentPool.Properties.Count = to.Ptr(*agentPool.Properties.Count + int32(nodeNum))

	// update agent pool desired size
	if err = scaleUpNodePool(ctx, client, info, agentPool); err != nil {
		blog.Errorf("applyInstanceMachines[%s] failed: %v", taskId, err)
		return err
	}
	blog.Infof("applyInstanceMachines[%s] success", taskId)

	// check agent
	err = checkScaleUp(ctx, client, info)
	if err != nil {
		// rollout instances
		_ = ScaleAgentPoolToDesiredSize(ctx, info, *agentPool.Properties.Count-int32(nodeNum))
		blog.Errorf("applyInstanceMachines[%s] checkScaleUp failed: %v", taskId, err)
		return err
	}

	return nil
}

// ScaleAgentPoolToDesiredSize scale agent pool to desired size nodes
func ScaleAgentPoolToDesiredSize(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, desired int32) error {
	taskId := cloudprovider.GetTaskIDFromContext(ctx)

	client, err := api.NewAksServiceImplWithCommonOption(info.CmOption) // new client
	if err != nil {
		return errors.Wrapf(err, "new azure client failed")
	}

	agentPool, err := client.GetPoolAndReturn(ctx, cloudprovider.GetClusterResourceGroup(info.Cluster),
		info.Cluster.SystemID, info.NodeGroup.CloudNodeGroupID)
	if err != nil {
		blog.Errorf("ScaleAgentPoolToDesiredSize[%s] GetPoolAndReturn failed: %v", taskId, err)
		return err
	}
	agentPool.Properties.Count = &desired

	// update agent pool desired size
	if err = scaleUpNodePool(ctx, client, info, agentPool); err != nil {
		blog.Errorf("ScaleAgentPoolToDesiredSize[%s] scaleUpNodePool[%s] failed: %v", taskId, *agentPool.Name, err)
		return err
	}

	return nil
}

// preCheckAgentPool 前置检查
func preCheckAgentPool(rootCtx context.Context, client api.AksService, info *cloudprovider.CloudDependBasicInfo) (
	*armcontainerservice.AgentPool, error) {
	var (
		group       = info.NodeGroup
		asg         = group.AutoScaling
		agentPool   *armcontainerservice.AgentPool
		taskID      = cloudprovider.GetTaskIDFromContext(rootCtx)
		ctx, cancel = context.WithTimeout(rootCtx, 30*time.Second)
	)
	defer cancel()

	// 检查 vmScaleSet 是否存在: vmScaleSet 的 AutoScalingName 资源组
	err := retry.Do(
		func() error {
			if _, err := client.GetSetWithName(ctx, asg.AutoScalingName, asg.AutoScalingID); err != nil {
				return err
			}
			return nil
		},
		retry.Context(ctx), retry.Attempts(3),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "preCheckAgentPool[%s] GetSetWithName failed", taskID)
	}
	// 检查 agentPool 是否存在
	ctx, cancel = context.WithTimeout(rootCtx, 30*time.Second)
	defer cancel()
	err = retry.Do(
		func() error {
			if agentPool, err = client.GetPoolAndReturn(ctx, cloudprovider.GetClusterResourceGroup(info.Cluster),
				info.Cluster.SystemID, group.CloudNodeGroupID); err != nil {
				return err
			}
			return nil
		},
		retry.Context(ctx), retry.Attempts(3),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "preCheckAgentPool[%s] GetPoolAndReturn failed", taskID)
	}

	return agentPool, nil
}

// scaleUpNodePool 扩容
func scaleUpNodePool(rootCtx context.Context, client api.AksService, info *cloudprovider.CloudDependBasicInfo,
	targetPool *armcontainerservice.AgentPool) error {
	var (
		cluster     = info.Cluster
		taskID      = cloudprovider.GetTaskIDFromContext(rootCtx)
		ctx, cancel = context.WithTimeout(rootCtx, 20*time.Minute)
	)
	defer cancel()

	err := loop.LoopDoFunc(ctx, func() error {
		pool, err := client.UpdatePoolAndReturn(ctx, targetPool, cloudprovider.GetClusterResourceGroup(info.Cluster),
			cluster.SystemID, *targetPool.Name)
		// 扩容完成
		if err == nil {
			targetPool.Properties = pool.Properties
			return loop.EndLoop
		}
		if strings.Contains(err.Error(), "missing error information") { // 如果节点池正在扩容中，此时再次扩容，则会失败
			return errors.Errorf("scaleUpNodePool[%s] continuous scale up fails", taskID)
		}

		// 扩容失败
		return errors.Wrapf(err, "scaleUpNodePool[%s] UpdatePoolAndReturn failed(scale up)", taskID)
	}, loop.LoopInterval(30*time.Second))

	if err != nil {
		return errors.Wrapf(err, "scaleUpNodePool[%s] UpdatePoolAndReturn failed(scale up)", taskID)
	}
	blog.Infof("scaleUpNodePool[%s] successfully", taskID)

	return nil
}

// checkScaleUp 检查扩容状态
func checkScaleUp(rootCtx context.Context, client api.AksService, info *cloudprovider.CloudDependBasicInfo) error {
	var (
		group       = info.NodeGroup
		taskID      = cloudprovider.GetTaskIDFromContext(rootCtx)
		ctx, cancel = context.WithTimeout(rootCtx, 5*time.Second)
	)
	defer cancel()

	err := loop.LoopDoFunc(ctx, func() error {
		agentPool, err := client.GetPoolAndReturn(ctx, cloudprovider.GetClusterResourceGroup(info.Cluster),
			info.Cluster.SystemID, group.CloudNodeGroupID)
		if err != nil {
			return errors.Wrapf(err, "checkScaleUp[%s] call GetPoolAndReturn failed", taskID)
		}
		// 打印状态
		status := *agentPool.Properties.ProvisioningState
		blog.Infof("checkScaleUp[%s] check scale up state is %s", taskID, status)
		if status != api.NormalState {
			return nil
		}
		// 扩容完成
		return loop.EndLoop
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		return errors.Wrapf(err, "taskID[%s] checkScaleUp[%s][%s] failed", taskID, group.CloudNodeGroupID,
			group.Name)
	}

	return nil
}

// recordClusterInstanceToDB already auto build instances to cluster, thus not handle error
func recordClusterInstanceToDB(ctx context.Context, state *cloudprovider.TaskState,
	info *cloudprovider.CloudDependBasicInfo, nodeNum uint64) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx) // 获取 taskID

	client, err := api.NewAksServiceImplWithCommonOption(info.CmOption) // new client
	if err != nil {
		blog.Errorf("recordClusterInstanceToDB[%s] init azure client failed", taskID)
		return errors.Wrapf(err, "new azure client failed")
	}

	successInstance, err := differentInstance(ctx, client, info) // 与db中的node对比，筛选出被扩容出来的机器
	if err != nil {
		return errors.Wrapf(err, "recordClusterInstanceToDB[%s] call differentInstance failed", taskID)
	}

	// 回滚期望数量
	if len(successInstance) != int(nodeNum) {
		_ = cloudprovider.UpdateNodeGroupDesiredSize(info.NodeGroup.NodeGroupID, int(nodeNum)-len(successInstance),
			true)
	}

	//  无失败机器的情况 instanceIDs 保存到 队列common中
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	if len(successInstance) > 0 {
		successInstanceID := make([]string, len(successInstance))
		for i := range successInstance {
			successInstanceID[i] = *successInstance[i].InstanceID
		}
		state.Task.CommonParams[cloudprovider.NodeIDsKey.String()] = strings.Join(successInstanceID, ",")
		state.Task.CommonParams[cloudprovider.SuccessNodeIDsKey.String()] = strings.Join(successInstanceID, ",")
	}

	// record successNodes to cluster manager DB
	nodeIPs, err := transInstancesToNode(ctx, info, client, successInstance)
	if err != nil {
		blog.Errorf("recordClusterInstanceToDB[%s] transInstancesToNode failed: %v", taskID, err)
	}
	if len(nodeIPs) > 0 {
		state.Task.NodeIPList = nodeIPs
		state.Task.CommonParams[cloudprovider.OriginNodeIPsKey.String()] = strings.Join(nodeIPs, ",")
		state.Task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")
	}

	return nil
}

// differentInstance 对比 - 查找出被扩容出来的 vm node
func differentInstance(rootCtx context.Context, client api.AksService, info *cloudprovider.CloudDependBasicInfo) (
	[]*armcompute.VirtualMachineScaleSetVM, error) {
	ctx, cancel := context.WithTimeout(rootCtx, 30*time.Second)
	defer cancel()

	var (
		asg    = info.NodeGroup.AutoScaling
		taskID = cloudprovider.GetTaskIDFromContext(rootCtx)
		res    = make([]*armcompute.VirtualMachineScaleSetVM, 0)
	)

	// 获取 vmScaleSet node list
	vmList, err := client.ListInstanceAndReturn(ctx, asg.AutoScalingName, asg.AutoScalingID)
	if err != nil {
		return nil, errors.Wrapf(err, "differentInstance[%s] ListInstanceAndReturn failed", taskID)
	}

	// 获取 node map
	nodeMap, err := getNodeMap(rootCtx, taskID, info)
	if err != nil {
		return nil, errors.Wrapf(err, "differentInstance[%s] getNodeMap failed", taskID)
	}

	// 比对
	for i, vm := range vmList {
		nodeID := fmt.Sprintf("%s/%s/%s", *vm.Name, *vm.InstanceID, asg.AutoScalingName)
		if _, ok := nodeMap[nodeID]; !ok {
			// 如果当前vm不存在于nodeMap中，则为扩容出来的机器
			res = append(res, vmList[i])
		}
	}
	return res, nil
}

// getNodeMap node map
func getNodeMap(ctx context.Context, taskID string, info *cloudprovider.CloudDependBasicInfo) (map[string]bool,
	error) {
	group := info.NodeGroup
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"nodegroupid": group.NodeGroupID,
	})
	// get node list
	nodes, err := cloudprovider.GetStorageModel().ListNode(ctx, cond, &storeopt.ListOption{All: true})
	if err != nil {
		return nil, errors.Wrapf(err,
			"getNodeMap[%s] list group nodes in nodegroup %s for Cluster %s failed", taskID,
			group.NodeGroupID, info.Cluster.ClusterID)
	}
	// list to map
	nodeMap := make(map[string]bool)
	for _, node := range nodes {
		nodeMap[node.NodeID] = true
	}
	return nodeMap, nil
}

// transInstancesToNode record success nodes to cm DB
func transInstancesToNode(rootCtx context.Context, info *cloudprovider.CloudDependBasicInfo, client api.AksService,
	vmList []*armcompute.VirtualMachineScaleSetVM) ([]string, error) {
	var (
		err           error
		nodeIPs       = make([]string, 0)
		nodes         []*proto.Node
		asg           = info.NodeGroup.AutoScaling
		interfaceList = make([]*armnetwork.Interface, 0)
		taskID        = cloudprovider.GetTaskIDFromContext(rootCtx)
		ctx, cancel   = context.WithTimeout(rootCtx, 30*time.Second)
	)
	defer cancel()

	// 获取 interface list
	err = retry.Do(func() error {
		interfaceList, err = client.ListSetInterfaceAndReturn(ctx, asg.AutoScalingName, asg.AutoScalingID)
		if err != nil {
			return errors.Wrapf(err, "transInstancesToNode[%s] ListSetInterfaceAndReturn failed", taskID)
		}
		return nil
	}, retry.Context(ctx), retry.Attempts(3))
	if err != nil {
		return nil, errors.Wrapf(err, "transInstancesToNode[%s] get vm network interface failed", taskID)
	}

	// vm to bcs node
	nodes, err = vmToNode(client, info, vmList, interfaceList)
	if err != nil {
		return nil, errors.Wrapf(err, "transInstancesToNode[%s] call vmToNode failed", taskID)
	}

	// save node
	for _, node := range nodes {
		nodeIPs = append(nodeIPs, node.InnerIP)
		node.Status = common.StatusInitialization
		node.Passwd = info.NodeGroup.LaunchTemplate.InitLoginPassword

		blog.Infof("transInstancesToNode save node:%s", utils.ObjToJson(node))
		if err = cloudprovider.SaveNodeInfoToDB(context.Background(), node, true); err != nil {
			blog.Errorf("transInstancesToNode[%s] SaveNodeInfoToDB[%s] failed: %v", taskID, node.InnerIP, err)
		}
	}
	return nodeIPs, nil
}

// vmToNode vm to node
func vmToNode(client api.AksService, info *cloudprovider.CloudDependBasicInfo,
	vmList []*armcompute.VirtualMachineScaleSetVM, interfaceList []*armnetwork.Interface) ([]*proto.Node, error) { // nolint
	var (
		node    *proto.Node
		cluster = info.Cluster
		group   = info.NodeGroup
		resp    = make([]*proto.Node, 0)
		ipMap   = api.VmMatchInterface(vmList, interfaceList)
	)

	for _, vm := range vmList { // 字段对齐
		node = new(proto.Node)
		_ = client.VmToNode(vm, node)
		node.CPU = group.LaunchTemplate.CPU
		node.Mem = group.LaunchTemplate.Mem
		node.GPU = group.LaunchTemplate.GPU
		if ip, ok := ipMap[*vm.Name]; ok && len(ip) != 0 {
			node.InnerIP = ip[0]
		}
		node.NodeGroupID = group.NodeGroupID
		node.ClusterID = cluster.ClusterID
		resp = append(resp, node)
	}
	return resp, nil
}

// CheckClusterNodesStatusTask check update desired nodes status task. nodes already add to cluster,
// thus not rollback desiredNum and only record status
func CheckClusterNodesStatusTask(taskID string, stepName string) error { // nolint
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		return nil
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	// extract parameter && check validate
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	manual := state.Task.CommonParams[cloudprovider.ManualKey.String()]

	successInstanceID := strings.Split(state.Task.CommonParams[cloudprovider.SuccessNodeIDsKey.String()], ",")
	if len(clusterID) == 0 || len(nodeGroupID) == 0 || len(cloudID) == 0 || len(successInstanceID) == 0 {
		blog.Errorf("CheckClusterNodesStatusTask[%s]: check parameter validate failed", taskID)
		retErr := fmt.Errorf("CheckClusterNodesStatusTask check parameters failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("CheckClusterNodesStatusTask[%s]: GetClusterDependBasicInfo failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("CheckClusterNodesStatusTask GetClusterDependBasicInfo failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// 无失败机器的情况
	successInstances, failureInstances, err := checkClusterInstanceStatus(ctx, dependInfo, successInstanceID)
	if err != nil || len(successInstances) == 0 {
		if manual != common.True {
			// rollback failed nodes
			_ = returnAzureInstancesAndCleanNodes(ctx, dependInfo, failureInstances)
		}
		blog.Errorf("CheckClusterNodesStatusTask[%s]: checkClusterInstanceStatus failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("CheckClusterNodesStatusTask checkClusterInstanceStatus failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// rollback abnormal nodes
	if len(failureInstances) > 0 {
		blog.Errorf("CheckClusterNodesStatusTask[%s] handle failedNodes[%v]", taskID, failureInstances)
		errMsg := returnAzureInstancesAndCleanNodes(ctx, dependInfo, failureInstances)
		if errMsg != nil {
			blog.Errorf("CheckClusterNodesStatusTask[%s] returnInstancesAndCleanNodes failed %v", taskID, errMsg)
		}
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	// inject instances nodeIds
	if len(successInstances) > 0 {
		state.Task.CommonParams[cloudprovider.SuccessClusterNodeIDsKey.String()] = strings.Join(successInstances, ",")
	}
	if len(failureInstances) > 0 {
		state.Task.CommonParams[cloudprovider.FailedClusterNodeIDsKey.String()] = strings.Join(failureInstances, ",")
	}

	nodes := cloudprovider.GetNodesByInstanceIDs(successInstances)
	var (
		ipList    = make([]string, 0)
		nodeNames = make([]string, 0)
	)
	for i := range nodes {
		ipList = append(ipList, nodes[i].InnerIP)
		nodeNames = append(nodeNames, nodes[i].NodeName)
	}

	if len(ipList) > 0 {
		state.Task.CommonParams[cloudprovider.NodeNamesKey.String()] = strings.Join(nodeNames, ",")
		// dynamic inject paras
		state.Task.CommonParams[cloudprovider.DynamicNodeIPListKey.String()] = strings.Join(ipList, ",")
		state.Task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(ipList, ",")
		state.Task.NodeIPList = ipList
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckClusterNodesStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

func checkClusterInstanceStatus(rootCtx context.Context, info *cloudprovider.CloudDependBasicInfo,
	instanceIDs []string) ([]string, []string, error) {
	var (
		addSuccessNodes = make([]string, 0)
		addFailureNodes = make([]string, 0)
		asg             = info.NodeGroup.AutoScaling
		taskID          = cloudprovider.GetTaskIDFromContext(rootCtx)
		ctx, cancel     = context.WithTimeout(context.TODO(), 5*time.Minute)
		instanceList    []*armcompute.VirtualMachineScaleSetVM
	)
	defer cancel()

	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())
	client, err := api.NewAksServiceImplWithCommonOption(info.CmOption) // new client
	if err != nil {
		return nil, nil, errors.Wrapf(err, "checkClusterInstanceStatus[%s] new client failed", taskID)
	}

	// wait all nodes to be ready
	errLoop := loop.LoopDoFunc(ctx, func() error {
		running := make([]string, 0)
		nodes, err2 := k8sOperator.ListClusterNodes(context.Background(), info.Cluster.ClusterID)
		if err2 != nil {
			blog.Errorf("checkClusterInstanceStatus[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, err)
			return nil
		}

		var k8sNodeMap = make(map[string]*corev1.Node, 0)
		for i := range nodes {
			k8sNodeMap[nodes[i].Name] = nodes[i]
		}

		instanceList, err = client.ListInstanceByIDAndReturn(ctx, asg.AutoScalingName, asg.AutoScalingID, instanceIDs)
		if err != nil {
			return errors.Wrapf(err, "checkClusterInstanceStatus[%s] ListInstanceByIDAndReturn failed", taskID)
		}
		for _, vm := range instanceList {
			id := api.VmIDToNodeID(vm)
			if n, ok := k8sNodeMap[*vm.Properties.OSProfile.ComputerName]; ok {
				if ok && cmutils.CheckNodeIfReady(n) {
					blog.Infof("checkClusterInstanceStatus[%s] node[%s] ready", taskID, id)
					running = append(running, id)
				}
			}
		}

		blog.Infof("checkClusterInstanceStatus[%s] ready nodes[%+v]", taskID, running)
		if len(running) == len(instanceIDs) {
			addSuccessNodes = running
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(30*time.Second))
	// other error
	if errLoop != nil && !errors.Is(errLoop, context.DeadlineExceeded) {
		blog.Errorf("checkClusterInstanceStatus[%s] check nodes status failed: %v", taskID, errLoop)
		return nil, nil, errLoop
	}

	// timeout error
	if errors.Is(errLoop, context.DeadlineExceeded) {
		addSuccessNodes, addFailureNodes, err = getVmStatus(k8sOperator, info, client, instanceIDs, taskID)
		if err != nil {
			return nil, nil, err
		}
	}
	blog.Infof("checkClusterInstanceStatus[%s] success[%v] failure[%v]", taskID, addSuccessNodes, addFailureNodes)

	for _, n := range addFailureNodes { // set cluster node status
		err = cloudprovider.UpdateNodeStatusByInstanceID(n, common.StatusAddNodesFailed)
		if err != nil {
			blog.Errorf("checkClusterInstanceStatus[%s] UpdateNodeStatusByInstanceID[%s] failed: %v", taskID, n, err)
		}
	}

	return addSuccessNodes, addFailureNodes, nil
}

func getVmStatus(k8sOperator *clusterops.K8SOperator, info *cloudprovider.CloudDependBasicInfo,
	client api.AksService, instanceIDs []string, taskID string) (
	[]string, []string, error) {
	running, failure := make([]string, 0), make([]string, 0)
	nodes, err := k8sOperator.ListClusterNodes(context.Background(), info.Cluster.ClusterID) // nolint
	if err != nil {
		blog.Errorf("checkClusterInstanceStatus[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, err)
		return nil, nil, err
	}

	var k8sNodeMap = make(map[string]*corev1.Node, 0)
	for i := range nodes {
		k8sNodeMap[nodes[i].Name] = nodes[i]
	}
	instanceList, err := client.ListInstanceByIDAndReturn(context.Background(), info.NodeGroup.AutoScaling.AutoScalingName,
		info.NodeGroup.AutoScaling.AutoScalingID, instanceIDs)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "checkClusterInstanceStatus[%s] ListInstanceByIDAndReturn failed", taskID)
	}

	for _, ins := range instanceList {
		id := api.VmIDToNodeID(ins)
		n, ok := k8sNodeMap[*ins.Properties.OSProfile.ComputerName]
		if ok && cmutils.CheckNodeIfReady(n) {
			running = append(running, id)
		} else {
			failure = append(failure, id)
		}
	}

	return running, failure, nil
}

// returnAzureInstancesAndCleanNodes return nodes and clean nodes data
func returnAzureInstancesAndCleanNodes(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	nodeIds []string) error { // nolint
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	if len(nodeIds) == 0 {
		blog.Infof("returnAzureInstancesAndCleanNodes[%s] nodeIds empty", taskID)
		return nil
	}

	// delete db data record
	for _, id := range nodeIds {
		err := cloudprovider.GetStorageModel().DeleteClusterNode(context.Background(),
			info.Cluster.ClusterID, id)
		if err != nil {
			blog.Errorf("returnAzureInstancesAndCleanNodes[%s] DeleteClusterNode[%s] failed: %v",
				taskID, id, err)
		} else {
			blog.Infof("returnAzureInstancesAndCleanNodes[%s] DeleteClusterNode success[%+v]", taskID, id)
		}
	}

	// delete instances
	err := removeVMSSsInstances(ctx, info, nodeIds)
	if err != nil {
		blog.Errorf("returnAzureInstancesAndCleanNodes[%s] removeVMSSsInstances[%+v] "+
			"failed: %v", taskID, nodeIds, err)
	} else {
		blog.Infof("returnAzureInstancesAndCleanNodes[%s] removeVMSSsInstances[%+v] success", taskID, nodeIds)
	}

	// rollback nodeGroup desired size
	err = cloudprovider.UpdateNodeGroupDesiredSize(info.NodeGroup.NodeGroupID, len(nodeIds), true)
	if err != nil {
		blog.Errorf("returnAzureInstancesAndCleanNodes[%s] UpdateNodeGroupDesiredSize failed: %v", taskID, err)
	} else {
		blog.Infof("returnAzureInstancesAndCleanNodes[%s] UpdateNodeGroupDesiredSize success[%v]",
			taskID, len(nodeIds))
	}

	return nil
}
