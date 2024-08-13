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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/avast/retry-go"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	corev1 "k8s.io/api/core/v1"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/aws/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
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
	operator := step.Params[cloudprovider.OperatorKey.String()]
	manual := state.Task.CommonParams[cloudprovider.ManualKey.String()]

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
	err = applyInstanceMachines(ctx, dependInfo, uint64(nodeNum))
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
		blog.Errorf("ApplyInstanceMachinesTask[%s]: recordClusterInstanceToDB failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("ApplyInstanceMachinesTask recordClusterInstanceToDB failed")
		if manual == common.True {
			_ = cloudprovider.UpdateVirtualNodeStatus(clusterID, nodeGroupID, taskID)
		} else {
			_ = cloudprovider.UpdateNodeGroupDesiredSize(nodeGroupID, nodeNum, true)
		}
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("ApplyInstanceMachinesTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	// destroy virtual nodes
	if manual == common.True {
		blog.Infof("ApplyInstanceMachinesTask[%s] begin DeleteVirtualNodes", taskID)
		_ = cloudprovider.DeleteVirtualNodes(clusterID, nodeGroupID, taskID)
	}

	return nil
}

// applyInstanceMachines apply machines from asg
func applyInstanceMachines(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, nodeNum uint64) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	asgName, err := getAsgNameByNodeGroup(ctx, info)
	if err != nil {
		return fmt.Errorf("applyInstanceMachines[%s] getAsgNameByNodeGroup failed: %v", taskID, err)
	}
	asCli, err := api.NewAutoScalingClient(info.CmOption)
	if err != nil {
		return err
	}

	err = loop.LoopDoFunc(context.Background(), func() error {
		err = asCli.SetDesiredCapacity(asgName, int64(nodeNum))
		if err != nil {
			if strings.Contains(err.Error(), autoscaling.ErrCodeScalingActivityInProgressFault) {
				blog.Infof("applyInstanceMachines[%s] ScaleOutInstances: %v", taskID,
					autoscaling.ErrCodeScalingActivityInProgressFault)
				return nil
			}
			blog.Errorf("applyInstanceMachines[%s] ScaleOutInstances failed: %v", taskID, err)
			return err
		}
		return loop.EndLoop
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		return fmt.Errorf("applyInstanceMachines[%s] SetDesiredCapacity failed: %v", taskID, err)
	}

	return nil
}

func getInstancesFromAsg(asCli *api.AutoScalingClient, asgName string) ([]*autoscaling.Instance, error) {
	asgInfo, err := asCli.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{&asgName}})
	if err != nil {
		blog.Errorf("getInstancesFromAsg DescribeAutoScalingGroups[%s] failed: %v", asgName, err)
		return nil, err
	}
	var instances []*autoscaling.Instance
	if asgInfo != nil {
		instances = asgInfo[0].Instances
	}

	return instances, nil
}

// recordClusterInstanceToDB already auto build instances to cluster, thus not handle error
func recordClusterInstanceToDB(ctx context.Context, state *cloudprovider.TaskState,
	info *cloudprovider.CloudDependBasicInfo, nodeNum uint64) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	successInstance, err := differentInstance(ctx, info, nodeNum) // 与db中的node对比，筛选出被扩容出来的机器
	if err != nil {
		return fmt.Errorf("recordClusterInstanceToDB[%s] call differentInstance failed", taskID)
	}

	// rollback desired num
	if len(successInstance) != int(nodeNum) {
		_ = cloudprovider.UpdateNodeGroupDesiredSize(info.NodeGroup.NodeGroupID, int(nodeNum)-len(successInstance),
			true)
	}

	// record instanceIDs to task common
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	successInstanceID := make([]string, 0)
	for _, v := range successInstance {
		successInstanceID = append(successInstanceID, *v.InstanceId)
	}

	successIns, failureIns, err := checkInstance(info, successInstanceID)
	if err != nil {
		_ = applyInstanceMachines(ctx, info, 0)
		blog.Errorf("recordClusterInstanceToDB[%s] checkInstance failed, %v, successInstances[%+v],"+
			" failureInstances[%+v]", taskID, err, successIns, failureIns)
		state.Task.CommonParams[cloudprovider.SuccessNodeIDsKey.String()] = strings.Join(successIns, ",")
		state.Task.CommonParams[cloudprovider.FailureNodeIDsKey.String()] = strings.Join(failureIns, ",")
		state.Task.CommonParams[cloudprovider.FailureReason.String()] = err.Error()
		return fmt.Errorf("checkInstance failed, %v, successInstances[%+v], failureInstances[%+v]",
			err, successIns, failureIns)
	}

	if len(successIns) > 0 {
		state.Task.CommonParams[cloudprovider.SuccessNodeIDsKey.String()] = strings.Join(successIns, ",")
		state.Task.CommonParams[cloudprovider.NodeNamesKey.String()] = strings.Join(successIns, ",")
		state.Task.CommonParams[cloudprovider.NodeIDsKey.String()] = strings.Join(successIns, ",")
	}

	// record successNodes to cluster manager DB
	nodeIPs, err := transInstancesToNode(ctx, state, successIns, info)
	if err != nil {
		blog.Errorf("recordClusterInstanceToDB[%s] failed: %v", taskID, err)
	}
	if len(nodeIPs) > 0 {
		state.Task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")
	}

	return nil
}

func checkInstance(info *cloudprovider.CloudDependBasicInfo, nodeIDs []string) ([]string, []string, error) {
	client, err := api.NewAWSClientSet(info.CmOption)
	if err != nil {
		blog.Errorf("create aws client failed, %v", err)
		return nil, nil, err
	}

	successIns, failureIns := make([]string, 0), make([]string, 0)
	timeCtx, cancel := context.WithTimeout(context.TODO(), 5*time.Minute)
	defer cancel()

	err = loop.LoopDoFunc(timeCtx, func() error {
		running, failed := make([]string, 0), make([]string, 0)
		instances, errDes := client.DescribeInstancesPages(
			&ec2.DescribeInstancesInput{InstanceIds: aws.StringSlice(nodeIDs)})
		if errDes != nil {
			blog.Errorf("checkInstance DescribeInstances[%+v] failed, %s", nodeIDs, errDes.Error())
			return nil
		}

		for _, ins := range instances {
			running = append(running, *ins.InstanceId)
		}

		for _, id := range nodeIDs {
			if !utils.StringInSlice(id, running) {
				failed = append(failed, id)
				blog.Infof("checkInstance instance[%s] not found", id)
				continue
			}
		}

		blog.Infof("checkInstance desired %d, response %d", len(nodeIDs), len(instances))

		successIns = running
		failureIns = failed
		if len(successIns) == len(nodeIDs) {
			return loop.EndLoop
		}

		return nil
	})
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return nil, nil, err
	}
	if errors.Is(err, context.DeadlineExceeded) {
		instances, errDes := client.DescribeInstancesPages(
			&ec2.DescribeInstancesInput{InstanceIds: aws.StringSlice(nodeIDs)})
		if errDes != nil {
			blog.Errorf("checkInstance DescribeInstances[%+v] failed, %s", nodeIDs, errDes.Error())
			return nil, nil, errDes
		}
		for _, ins := range instances {
			successIns = append(successIns, *ins.InstanceId)
		}

		for _, id := range nodeIDs {
			if !utils.StringInSlice(id, successIns) {
				failureIns = append(failureIns, id)
				blog.Infof("checkInstance instance[%s] not found", id)
				continue
			}
		}

		if len(failureIns) > 0 {
			return successIns, failureIns, fmt.Errorf("failed to get instances[%+v]", failureIns)
		}
	}

	return successIns, failureIns, nil
}

// differentInstance 对比 - 查找出被扩容出来的 vm node
func differentInstance(rootCtx context.Context, info *cloudprovider.CloudDependBasicInfo, nodeNum uint64) (
	[]*autoscaling.Instance, error) {
	taskID := cloudprovider.GetTaskIDFromContext(rootCtx)
	res := make([]*autoscaling.Instance, 0)
	asgName, err := getAsgNameByNodeGroup(rootCtx, info)
	if err != nil {
		return nil, err
	}

	asCli, err := api.NewAutoScalingClient(info.CmOption)
	if err != nil {
		return nil, err
	}

	timeCtx, cancel := context.WithTimeout(context.TODO(), 10*time.Minute)
	defer cancel()
	instances := make([]*autoscaling.Instance, 0)
	err = loop.LoopDoFunc(timeCtx, func() error {
		instances, err = getInstancesFromAsg(asCli, asgName)
		if err != nil {
			return err
		}

		blog.Infof("differentInstance[%s] instances[%d], desired[%d]", taskID, len(instances), nodeNum)
		if len(instances) == int(nodeNum) {
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(30*time.Second))
	if err != nil {
		return nil, err
	}

	// 获取 node map
	nodeMap, err := getNodeMap(rootCtx, info)
	if err != nil {
		return nil, err
	}

	// 比对
	for _, vm := range instances {
		nodeID := *vm.InstanceId
		if _, ok := nodeMap[nodeID]; !ok {
			// 如果当前vm不存在于nodeMap中，则为扩容出来的机器
			res = append(res, vm)
		}
	}

	return res, nil
}

// getNodeMap node map
func getNodeMap(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (map[string]bool,
	error) {
	group := info.NodeGroup
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"nodegroupid": group.NodeGroupID,
	})
	// get node list
	nodes, err := cloudprovider.GetStorageModel().ListNode(ctx, cond, &storeopt.ListOption{All: true})
	if err != nil {
		return nil, err
	}
	// list to map
	nodeMap := make(map[string]bool)
	for _, node := range nodes {
		nodeMap[node.NodeID] = true
	}
	return nodeMap, nil
}

// transInstancesToNode record success nodes to cm DB
func transInstancesToNode(ctx context.Context, state *cloudprovider.TaskState, successInstanceID []string,
	info *cloudprovider.CloudDependBasicInfo) ([]string, error) {
	var (
		client  = api.NodeManager{}
		nodes   = make([]*proto.Node, 0)
		nodeIPs = make([]string, 0)
		err     error
	)

	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	err = retry.Do(func() error {
		nodes, err = client.ListNodesByInstanceID(successInstanceID, &cloudprovider.ListNodesOption{
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
		n.CPU = info.NodeGroup.LaunchTemplate.CPU
		n.Mem = info.NodeGroup.LaunchTemplate.Mem
		n.Status = common.StatusInitialization
		err = cloudprovider.SaveNodeInfoToDB(ctx, n, true)
		if err != nil {
			blog.Errorf("transInstancesToNode[%s] SaveNodeInfoToDB[%s] failed: %v", taskID, n.InnerIP, err)
		}
	}

	if len(nodes) > 0 {
		successNodeNames := make([]string, len(successInstanceID))
		for i := range nodes {
			successNodeNames[i] = nodes[i].NodeName
		}
		state.Task.CommonParams[cloudprovider.NodeIDsKey.String()] = strings.Join(successNodeNames, ",")
		state.Task.CommonParams[cloudprovider.SuccessNodeIDsKey.String()] = strings.Join(successNodeNames, ",")
	}

	return nodeIPs, nil
}

func getAsgNameByNodeGroup(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	ngName := info.NodeGroup.CloudNodeGroupID
	eksCli, err := api.NewEksClient(info.CmOption)
	if err != nil {
		return "", err
	}

	var ng *eks.Nodegroup
	err = retry.Do(
		func() error {
			ng, err = eksCli.DescribeNodegroup(&ngName, &info.Cluster.SystemID)
			if err != nil {
				return err
			}

			return nil
		},
		retry.Context(ctx), retry.Attempts(3),
	)
	if err != nil {
		blog.Errorf("ApplyInstanceMachinesTask[%s] getAsgNameByNodeGroup[%s] failed: %v", taskID, ngName, err)
		return "", err
	}
	if ng.Resources != nil && ng.Resources.AutoScalingGroups != nil {
		return *ng.Resources.AutoScalingGroups[0].Name, nil
	}

	return "", fmt.Errorf("ApplyInstanceMachinesTask[%s] getAsgNameByNodeGroup[%s] failed: %v", taskID, ngName, err)
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
	successNodeNames := strings.Split(state.Task.CommonParams[cloudprovider.SuccessNodeIDsKey.String()], ",")

	if len(clusterID) == 0 || len(nodeGroupID) == 0 || len(cloudID) == 0 || len(successNodeNames) == 0 {
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
	successInstances, failureInstances, err := checkClusterInstanceStatus(ctx, dependInfo, successNodeNames)
	if err != nil {
		blog.Errorf("CheckClusterNodesStatusTask[%s]: checkClusterInstanceStatus failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("CheckClusterNodesStatusTask checkClusterInstanceStatus failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	if len(successInstances) > 0 {
		state.Task.CommonParams[cloudprovider.SuccessClusterNodeIDsKey.String()] = strings.Join(successInstances, ",")
		// dynamic inject paras
		state.Task.CommonParams[cloudprovider.DynamicNodeIPListKey.String()] = strings.Join(successInstances, ",")
	}
	if len(failureInstances) > 0 {
		state.Task.CommonParams[cloudprovider.FailedClusterNodeIDsKey.String()] = strings.Join(failureInstances, ",")
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckClusterNodesStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

func checkClusterInstanceStatus(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	instanceNames []string) ([]string, []string, error) { //  nolint
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
