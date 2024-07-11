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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/avast/retry-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/google/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// ApplyInstanceMachinesTask update desired nodes task
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

	// extract parameter && check validate
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	desiredNodes := step.Params[cloudprovider.ScalingNodesNumKey.String()]
	nodeNum, _ := strconv.Atoi(desiredNodes)

	manual := state.Task.CommonParams[cloudprovider.ManualKey.String()]

	operator := step.Params[cloudprovider.OperatorKey.String()]
	if len(clusterID) == 0 || len(nodeGroupID) == 0 || len(cloudID) == 0 || len(desiredNodes) == 0 || len(operator) == 0 {
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
		blog.Errorf("ApplyInstanceMachinesTask[%s]: GetClusterDependBasicInfo failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("ApplyInstanceMachinesTask GetClusterDependBasicInfo failed")
		if manual == common.True {
			_ = cloudprovider.UpdateVirtualNodeStatus(clusterID, nodeGroupID, taskID)
		} else {
			_ = cloudprovider.UpdateNodeGroupDesiredSize(nodeGroupID, nodeNum, true)
		}
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	instanceNames, err := applyInstanceMachines(ctx, dependInfo, uint64(nodeNum))
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
	err = recordClusterInstanceToDB(ctx, state, dependInfo, instanceNames)
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
		blog.Errorf("ApplyInstanceMachinesTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

func generateInstanceName(migBaseName string, delta uint64) []string {
	instanceNames := make([]string, 0)
	for i := uint64(0); i < delta; i++ {
		newName := fmt.Sprintf("%v-%v", migBaseName, utils.RandomHexString(4))
		instanceNames = append(instanceNames, newName)
	}

	return instanceNames
}

// applyInstanceMachines apply machines from MIG
func applyInstanceMachines(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	nodeNum uint64) ([]string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	if len(info.NodeGroup.AutoScaling.Zones) != 0 {
		info.CmOption.Region = info.NodeGroup.AutoScaling.Zones[0]
	}

	client, err := api.NewComputeServiceClient(info.CmOption)
	if err != nil {
		return nil, err
	}
	mig, err := api.GetInstanceGroupManager(client, info.NodeGroup.AutoScaling.AutoScalingID)
	if err != nil {
		blog.Errorf("applyInstanceMachines[%s] GetInstanceGroupManager[%s] failed: %v",
			taskID, info.NodeGroup.AutoScaling.AutoScalingID, err)
		return nil, err
	}
	instanceNames := generateInstanceName(mig.BaseInstanceName, nodeNum)

	operation, err := api.CreateInstanceForGroupManager(client, info.NodeGroup.AutoScaling.AutoScalingID, instanceNames)
	if err != nil {
		blog.Errorf("applyInstanceMachines[%s] CreateInstanceForGroupManager[%s : %+v] failed: %v",
			taskID, info.NodeGroup.AutoScaling.AutoScalingID, instanceNames, err)
		return nil, err
	}

	blog.Infof("applyInstanceMachines[%s] CreateInstanceForGroupManager success[%s:%s]",
		taskID, operation.Name, operation.SelfLink)

	// get operation ID
	err = checkOperationStatus(client, operation.SelfLink, taskID, 5*time.Second)
	if err != nil {
		// rollout instances
		removeMigInstances(ctx, info, mig.Name, instanceNames) // nolint
		return nil, fmt.Errorf("applyInstanceMachines[%s] GetOperation failed: %v", taskID, err)
	}

	return instanceNames, nil
}

// removeMigInstances destroy mig instances
func removeMigInstances(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	migName string, instances []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	igmInfo, err := api.GetGCEResourceInfo(info.NodeGroup.AutoScaling.AutoScalingID)
	if err != nil {
		return fmt.Errorf("deleteIgmInstances[%s] get igm info failed: %v", taskID, err)
	}
	if migName == "" {
		migName = igmInfo[(len(igmInfo) - 1)]
	}

	client, err := api.NewComputeServiceClient(info.CmOption)
	if err != nil {
		blog.Errorf("deleteIgmInstances[%s] get gce client failed: %v", taskID, err.Error())
		return err
	}

	blog.Infof("removeMigInstances[%s] mig[%s:%s] instances[%+v]", taskID, igmInfo[3], migName, instances)
	operation, err := client.DeleteMigInstances(ctx, igmInfo[3], migName, instances)
	if err != nil {
		blog.Errorf("removeMigInstances[%s] mig[%s:%s] failed: %v", taskID, igmInfo[3], migName, err)
		return err
	}

	// async to check deleteMigInstances operation status
	go checkOperationStatus(client, operation.SelfLink, taskID, time.Second*10) // nolint

	return nil
}

// recordClusterInstanceToDB already auto build instances to cluster, thus not handle error
func recordClusterInstanceToDB(ctx context.Context, state *cloudprovider.TaskState,
	info *cloudprovider.CloudDependBasicInfo, instancesNames []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	if len(info.NodeGroup.AutoScaling.Zones) != 0 {
		info.CmOption.Region = info.NodeGroup.AutoScaling.Zones[0]
	}

	client, err := api.NewComputeServiceClient(info.CmOption)
	if err != nil {
		return err
	}

	// check instance Group node status
	igm, _ := api.GetInstanceGroupManager(client, info.NodeGroup.AutoScaling.AutoScalingID)
	instances, err := api.ListInstanceGroupsInstances(client, igm.InstanceGroup)
	if err != nil {
		blog.Errorf("recordClusterInstanceToDB[%s]: ListInstanceGroupsInstances failed: %s", taskID, err.Error())
		return err
	}
	blog.Infof("recordClusterInstanceToDB[%s]: ListInstanceGroupsInstances got %d instances, %#v", taskID,
		len(instances), instances)

	for _, ins := range instancesNames {
		for _, cloudIns := range instances {
			if strings.Contains(cloudIns.Instance, ins) {
				blog.Infof("recordClusterInstanceToDB[%s] instance[%s:%s] status", taskID, ins,
					cloudIns.Instance, cloudIns.Status)
				continue
			}
		}
	}

	// record instanceIDs to task common
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	if len(instancesNames) > 0 {
		state.Task.CommonParams[cloudprovider.SuccessNodeIDsKey.String()] = strings.Join(instancesNames, ",")
		state.Task.CommonParams[cloudprovider.NodeIDsKey.String()] = strings.Join(instancesNames, ",")
	}

	// record successNodes to cluster manager DB
	nodeIPs, err := transInstancesToNode(ctx, instancesNames, info)
	if err != nil {
		blog.Errorf("recordClusterInstanceToDB[%s] failed: %v", taskID, err)
	}
	if len(nodeIPs) > 0 {
		state.Task.NodeIPList = nodeIPs
		state.Task.CommonParams[cloudprovider.OriginNodeIPsKey.String()] = strings.Join(nodeIPs, ",")
		state.Task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")
	}

	return nil
}

func checkInstance(client *api.ComputeServiceClient, ids []string) error {
	timeCtx, cancel := context.WithTimeout(context.TODO(), 5*time.Minute)
	defer cancel()
	err := loop.LoopDoFunc(timeCtx, func() error {
		insList, err := client.ListZoneInstanceWithFilter(context.Background(), api.InstanceNameFilter(ids))
		if err != nil {
			blog.Errorf("checkInstance ListZoneInstanceWithFilter failed, %s", err.Error())
			return err
		}

		// check response data
		if len(insList.Items) != len(ids) {
			blog.Warnf("checkInstance desired %d, but got %d", len(ids), len(insList.Items))
			return nil
		}

		blog.Infof("checkInstance desired %d, response %d", len(ids), len(insList.Items))

		for _, in := range insList.Items {
			if len(in.NetworkInterfaces[0].NetworkIP) == 0 {
				blog.Warnf("checkInstance[%s] IP is still not distributed", in.Name)
				return nil
			}
		}

		return loop.EndLoop
	})
	if err != nil {
		return err
	}

	return nil
}

// transInstancesToNode record success nodes to cm DB
func transInstancesToNode(ctx context.Context, instanceNames []string, info *cloudprovider.CloudDependBasicInfo) (
	[]string, error) {
	var (
		nodeCli = api.NodeManager{}
		nodes   = make([]*proto.Node, 0)
		nodeIPs = make([]string, 0)
		err     error
	)
	client, err := api.NewComputeServiceClient(info.CmOption)
	if err != nil {
		blog.Errorf("transInstanceIDsToNodes create ComputeServiceClient failed, %s", err.Error())
		return nil, err
	}

	err = checkInstance(client, instanceNames)
	if err != nil {
		blog.Errorf("transInstanceIDsToNodes checkInstance failed, %s", err.Error())
		return nil, err
	}

	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	err = retry.Do(func() error {
		nodes, err = nodeCli.ListNodesByInstanceID(instanceNames, &cloudprovider.ListNodesOption{
			Common:       info.CmOption,
			ClusterVPCID: info.Cluster.VpcID,
		})
		if err != nil {
			return err
		}
		return nil
	}, retry.Attempts(3))
	if err != nil {
		blog.Errorf("transInstancesToNode[%s] failed: %v", taskID, err)
		return nil, err
	}

	for _, n := range nodes {
		nodeIPs = append(nodeIPs, n.InnerIP)
		n.ClusterID = info.NodeGroup.ClusterID
		n.NodeGroupID = info.NodeGroup.NodeGroupID
		n.Status = common.StatusInitialization
		n.Region = func() string {
			regionList := strings.Split(info.Cluster.Region, "-")
			if len(regionList) <= 2 {
				return info.Cluster.Region
			}

			return fmt.Sprintf("%s-%s", regionList[0], regionList[1])
		}()
		n.CPU = info.NodeGroup.GetLaunchTemplate().GetCPU()
		n.Mem = info.NodeGroup.GetLaunchTemplate().GetMem()
		n.GPU = info.NodeGroup.GetLaunchTemplate().GetGPU()

		err = cloudprovider.SaveNodeInfoToDB(ctx, n, true)
		if err != nil {
			blog.Errorf("transInstancesToNode[%s] SaveNodeInfoToDB[%s] failed: %v", taskID, n.InnerIP, err)
		}
	}

	return nodeIPs, nil
}

func checkClusterInstanceStatus(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	instanceNames []string) ([]string, []string, error) {
	var (
		addSuccessNodes = make([]string, 0)
		addFailureNodes = make([]string, 0)
	)

	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())

	// wait node group state to normal
	timeCtx, cancel := context.WithTimeout(context.TODO(), 10*time.Minute)
	defer cancel()

	// wait all nodes to be ready
	err := loop.LoopDoFunc(timeCtx, func() error {
		running := make([]string, 0)

		nodes, err := k8sOperator.ListClusterNodes(context.Background(), info.Cluster.ClusterID)
		if err != nil {
			blog.Errorf("checkClusterInstanceStatus[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, err)
			return nil
		}

		var nodeNameMap = make(map[string]*corev1.Node, 0)
		for i := range nodes {
			nodeNameMap[nodes[i].Name] = nodes[i]
		}

		for _, ins := range instanceNames {
			n, ok := nodeNameMap[ins]
			if ok && utils.CheckNodeIfReady(n) {
				blog.Infof("checkClusterInstanceStatus[%s] node[%s] ready", taskID, ins)
				running = append(running, ins)
			}
		}

		blog.Infof("checkClusterInstanceStatus[%s] ready nodes[%+v]", taskID, running)
		if len(running) == len(instanceNames) {
			addSuccessNodes = running
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(30*time.Second))
	// other error
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		blog.Errorf("checkClusterInstanceStatus[%s] check nodes status failed: %v", taskID, err)
		return nil, nil, err
	}

	// timeout error
	if errors.Is(err, context.DeadlineExceeded) {
		running, failure := make([]string, 0), make([]string, 0)

		nodes, err := k8sOperator.ListClusterNodes(context.Background(), info.Cluster.ClusterID) // nolint
		if err != nil {
			blog.Errorf("checkClusterInstanceStatus[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, err)
			return nil, nil, err
		}

		var nodeNameMap = make(map[string]*corev1.Node, 0)
		for i := range nodes {
			nodeNameMap[nodes[i].Name] = nodes[i]
		}

		for _, ins := range instanceNames {
			n, ok := nodeNameMap[ins]
			if ok && utils.CheckNodeIfReady(n) {
				running = append(running, ins)
			} else {
				failure = append(failure, ins)
			}
		}

		addSuccessNodes = running
		addFailureNodes = failure
	}
	blog.Infof("checkClusterInstanceStatus[%s] success[%v] failure[%v]", taskID, addSuccessNodes, addFailureNodes)

	// set cluster node status
	for _, n := range addFailureNodes {
		err = cloudprovider.UpdateNodeStatusByInstanceID(n, common.StatusAddNodesFailed)
		if err != nil {
			blog.Errorf("checkClusterInstanceStatus[%s] UpdateNodeStatusByInstanceID[%s] failed: %v", taskID, n, err)
		}
	}

	return addSuccessNodes, addFailureNodes, nil
}

// CheckClusterNodesStatusTask check update desired nodes status task. nodes already add to cluster,
// thus not rollback desiredNum and only record status
func CheckClusterNodesStatusTask(taskID string, stepName string) error {
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

	// step login started here
	// extract parameter && check validate
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	successInstanceNames := strings.Split(state.Task.CommonParams[cloudprovider.NodeIDsKey.String()], ",")
	manual := state.Task.CommonParams[cloudprovider.ManualKey.String()]

	if len(clusterID) == 0 || len(nodeGroupID) == 0 || len(cloudID) == 0 || len(successInstanceNames) == 0 {
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

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	successInstances, failureInstances, err := checkClusterInstanceStatus(ctx, dependInfo, successInstanceNames)
	if err != nil || len(successInstances) == 0 {
		if manual != common.True {
			// rollback failed nodes
			_ = returnGkeInstancesAndCleanNodes(ctx, dependInfo, successInstances)
		}
		blog.Errorf("CheckClusterNodesStatusTask[%s]: checkClusterInstanceStatus failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("CheckClusterNodesStatusTask checkClusterInstanceStatus failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// rollback abnormal nodes
	if len(failureInstances) > 0 {
		blog.Errorf("CheckClusterNodesStatusTask[%s] handle failedNodes[%v]", taskID, failureInstances)
		errMsg := returnGkeInstancesAndCleanNodes(ctx, dependInfo, failureInstances)
		if errMsg != nil {
			blog.Errorf("CheckClusterNodesStatusTask[%s] returnInstancesAndCleanNodes failed %v", taskID, errMsg)
		}
	}

	// trans instanceIDs to ipList
	ipList := cloudprovider.GetInstanceIPsByName(ctx, clusterID, successInstances)

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	if len(successInstances) > 0 {
		state.Task.CommonParams[cloudprovider.SuccessClusterNodeIDsKey.String()] = strings.Join(successInstances, ",")
	}
	if len(failureInstances) > 0 {
		state.Task.CommonParams[cloudprovider.FailedClusterNodeIDsKey.String()] = strings.Join(failureInstances, ",")
	}

	// successInstance ip list
	if len(ipList) > 0 {
		// dynamic inject paras
		state.Task.CommonParams[cloudprovider.DynamicNodeIPListKey.String()] = strings.Join(ipList, ",")
		state.Task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(ipList, ",")
		state.Task.NodeIPList = ipList
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckClusterNodesStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// RemoveClusterNodesTaintTask removes cluster nodes taint
func RemoveClusterNodesTaintTask(taskID string, stepName string) error {
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

	// step login started here
	// extract parameter && check validate
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	successInstanceID := strings.Split(state.Task.CommonParams[cloudprovider.SuccessNodeIDsKey.String()], ",")

	if len(clusterID) == 0 || len(nodeGroupID) == 0 || len(cloudID) == 0 || len(successInstanceID) == 0 {
		blog.Errorf("RemoveClusterNodesTaintTask[%s]: check parameter validate failed", taskID)
		retErr := fmt.Errorf("RemoveClusterNodesTaintTask check parameters failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("RemoveClusterNodesTaintTask[%s]: GetClusterDependBasicInfo failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("RemoveClusterNodesTaintTask GetClusterDependBasicInfo failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = removeClusterNodesTaint(ctx, dependInfo.Cluster.ClusterID, successInstanceID)
	if err != nil {
		blog.Errorf("RemoveClusterNodesTaintTask[%s]: removeClusterNodesTaint failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("RemoveClusterNodesTaintTask removeClusterNodesTaint failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("RemoveClusterNodesTaintTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

func removeClusterNodesTaint(ctx context.Context, clusterID string, successInstanceID []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())
	kubeCli, err := k8sOperator.GetClusterClient(clusterID)
	if err != nil {
		return err
	}

	for _, ins := range successInstanceID {
		node, errLocal := kubeCli.CoreV1().Nodes().Get(context.Background(), ins, metav1.GetOptions{})
		if errLocal != nil {
			blog.Errorf("removeClusterNodesTaint[%s] nodeName[%s] failed: %v", taskID, ins, err)
			continue
		}

		newTaints := make([]corev1.Taint, 0)
		for _, taint := range node.Spec.Taints {
			if taint.Key != api.BCSNodeGroupTaintKey {
				newTaints = append(newTaints, taint)
			}
		}
		node.Spec.Taints = newTaints

		_, errLocal = kubeCli.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
		if errLocal != nil {
			blog.Errorf("removeClusterNodesTaint[%s] nodeName[%s] failed: %v", taskID, ins, err)
			continue
		}

		blog.Errorf("removeClusterNodesTaint[%s] nodeName[%s] success", taskID, ins)
	}

	return nil
}

func returnGkeInstancesAndCleanNodes(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	instanceNames []string) error { // nolint
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	if len(instanceNames) == 0 {
		blog.Infof("returnGkeInstancesAndCleanNodes[%s] instanceNames empty", taskID)
		return nil
	}

	// delete db data record
	for _, name := range instanceNames {
		err := cloudprovider.GetStorageModel().DeleteClusterNodeByName(context.Background(),
			info.Cluster.ClusterID, name)
		if err != nil {
			blog.Errorf("returnGkeInstancesAndCleanNodes[%s] DeleteClusterNodeByName[%s] failed: %v",
				taskID, name, err)
		} else {
			blog.Infof("returnGkeInstancesAndCleanNodes[%s] DeleteClusterNodeByName success[%+v]", taskID, name)
		}
	}

	// delete instances
	err := removeMigInstances(ctx, info, "", instanceNames)
	if err != nil {
		blog.Errorf("returnGkeInstancesAndCleanNodes[%s] removeMigInstances[%+v] "+
			"failed: %v", taskID, instanceNames, err)
	} else {
		blog.Infof("returnGkeInstancesAndCleanNodes[%s] removeMigInstances[%+v] success", taskID, instanceNames)
	}

	// rollback nodeGroup desired size
	err = cloudprovider.UpdateNodeGroupDesiredSize(info.NodeGroup.NodeGroupID, len(instanceNames), true)
	if err != nil {
		blog.Errorf("returnGkeInstancesAndCleanNodes[%s] UpdateNodeGroupDesiredSize failed: %v", taskID, err)
	} else {
		blog.Infof("returnGkeInstancesAndCleanNodes[%s] UpdateNodeGroupDesiredSize success[%v]",
			taskID, len(instanceNames))
	}

	return nil
}
