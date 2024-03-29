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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"
	"github.com/pkg/errors"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	cmutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
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

	if len(clusterID) == 0 || len(nodeGroupID) == 0 || len(cloudID) == 0 ||
		len(desiredNodes) == 0 || len(operator) == 0 {
		blog.Errorf("ApplyInstanceMachinesTask[%s]: check parameter validate failed", taskID)
		retErr := fmt.Errorf("ApplyInstanceMachinesTask check parameters failed")
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
		_ = cloudprovider.UpdateNodeGroupDesiredSize(nodeGroupID, nodeNum, true)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = applyInstanceMachines(ctx, dependInfo, int32(nodeNum))
	if err != nil {
		blog.Errorf("ApplyInstanceMachinesTask[%s]: applyInstanceMachines failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("ApplyInstanceMachinesTask applyInstanceMachines failed")
		_ = cloudprovider.UpdateNodeGroupDesiredSize(nodeGroupID, nodeNum, true)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// trans success nodes to cm DB and record common paras, not handle error
	_ = recordClusterInstanceToDB(ctx, state, dependInfo, nodeNum)

	blog.Infof("ApplyInstanceMachinesTask[%s]: call updateDesiredNodes successful", taskID)

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("ApplyInstanceMachinesTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

// applyInstanceMachines apply machines from MIG
func applyInstanceMachines(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, nodeNum int32) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	client, err := api.NewCceClient(info.CmOption)
	if err != nil {
		return err
	}

	_, err = client.UpdateDesiredNodes(info.Cluster.SystemID, info.NodeGroup.CloudNodeGroupID, nodeNum)
	if err != nil {
		return err
	}

	var nodePool *model.ShowNodePoolResponse
	err = loop.LoopDoFunc(context.Background(), func() error {
		nodePool, err = client.GetClusterNodePool(info.Cluster.SystemID, info.NodeGroup.CloudNodeGroupID)

		if *nodePool.Status.CurrentNode == nodeNum && nodePool.Status.Phase.Value() == "" {
			return loop.EndLoop
		} else if nodePool.Status.Phase.Value() == model.GetNodePoolStatusPhaseEnum().ERROR.Value() ||
			nodePool.Status.Phase.Value() == model.GetNodePoolStatusPhaseEnum().SOLD_OUT.Value() {
			return fmt.Errorf("applyInstanceMachines[%s] GetOperation failed: %v", taskID, nodePool.Status.Phase.Value())
		}

		blog.Infof("taskID[%s] operation %s still running", taskID, nodePool.Status.Phase.Value())
		return nil
	}, loop.LoopInterval(3*time.Second))

	if err != nil {
		return fmt.Errorf("applyInstanceMachines[%s] GetOperation failed: %v", taskID, err)
	}

	return nil
}

// recordClusterInstanceToDB already auto build instances to cluster, thus not handle error
func recordClusterInstanceToDB(ctx context.Context, state *cloudprovider.TaskState,
	info *cloudprovider.CloudDependBasicInfo, nodeNum int) error {
	var (
		successInstance []model.Node
	)
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	client, err := api.NewCceClient(info.CmOption)
	if err != nil {
		return err
	}

	successInstance, err = differentInstance(ctx, info, client)
	if err != nil {
		return err
	}

	// 回滚期望数量
	if len(successInstance) != nodeNum {
		_ = cloudprovider.UpdateNodeGroupDesiredSize(info.NodeGroup.NodeGroupID, nodeNum-len(successInstance),
			true)
	}

	//  无失败机器的情况
	// instanceIDs 保存到 队列common中
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	if len(successInstance) > 0 {
		successInstanceID := make([]string, len(successInstance))
		for i := range successInstance {
			successInstanceID[i] = *successInstance[i].Metadata.Uid
		}
		state.Task.CommonParams[cloudprovider.NodeIDsKey.String()] = strings.Join(successInstanceID, ",")
		state.Task.CommonParams[cloudprovider.SuccessNodeIDsKey.String()] = strings.Join(successInstanceID, ",")
	}

	// record successNodes to cluster manager DB
	nodeIPs, err := transInstancesToNode(ctx, successInstance, info)
	if err != nil {
		blog.Errorf("recordClusterInstanceToDB[%s] transInstancesToNode failed: %v", taskID, err)
	}
	if len(nodeIPs) > 0 {
		state.Task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")
	}

	return nil
}

func differentInstance(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	client *api.CceClient) ([]model.Node, error) {
	var (
		taskID = cloudprovider.GetTaskIDFromContext(ctx)
		res    = make([]model.Node, 0)
	)
	instaces, err := client.ListClusterNodePoolNodes(info.Cluster.SystemID, info.NodeGroup.CloudNodeGroupID)
	if err != nil {
		return nil, err
	}

	// 获取 node map
	nodeMap, err := getNodeMap(ctx, taskID, info)
	if err != nil {
		return nil, errors.Wrapf(err, "scaleUpNodePool[%s] getNodeMap failed", taskID)
	}

	// 比对
	for i, vm := range instaces {
		nodeID := *vm.Metadata.Uid
		if _, ok := nodeMap[nodeID]; !ok {
			// 如果当前vm不存在于nodeMap中，则为扩容出来的机器
			res = append(res, instaces[i])
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
			"scaleUpNodePool[%s] list group nodes in nodegroup %s for Cluster %s failed", taskID,
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
func transInstancesToNode(ctx context.Context, instances []model.Node, info *cloudprovider.CloudDependBasicInfo) (
	[]string, error) { // nolint
	var (
		nodeIPs = make([]string, 0)
		err     error
	)
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	for _, v := range instances {
		node := proto.Node{}
		node.NodeID = *v.Metadata.Uid
		node.InstanceType = v.Spec.Flavor
		node.Region = info.CmOption.Region
		node.InnerIP = *v.Status.PrivateIP

		node.ClusterID = info.NodeGroup.ClusterID
		node.NodeGroupID = info.NodeGroup.NodeGroupID
		node.Status = common.StatusInitialization

		blog.Infof("ApplyInstanceMachinesTask[%s]: call transInstancesToNode successful. node: %#v", node)
		blog.Infof("ApplyInstanceMachinesTask[%s]: call transInstancesToNode successful. node.server: %#v", *v.Status)

		err = cloudprovider.SaveNodeInfoToDB(ctx, &node, false)
		if err != nil {
			blog.Errorf("transInstancesToNode[%s] SaveNodeInfoToDB[%s] failed: %v", taskID, node.InnerIP, err)
		}

		nodeIPs = append(nodeIPs, node.InnerIP)
	}

	return nodeIPs, nil
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

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	// extract parameter && check validate
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
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

func checkClusterInstanceStatus(rootCtx context.Context, info *cloudprovider.CloudDependBasicInfo,
	instanceIDs []string) ([]string, []string, error) {
	var (
		addSuccessNodes = make([]string, 0)
		addFailureNodes = make([]string, 0)
		taskID          = cloudprovider.GetTaskIDFromContext(rootCtx)
		ctx, cancel     = context.WithTimeout(context.TODO(), 10*time.Minute)
	)
	defer cancel()

	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())

	// wait all nodes to be ready
	err := loop.LoopDoFunc(ctx, func() error {
		running := make([]string, 0)
		nodes, lnErr := k8sOperator.ListClusterNodes(context.Background(), info.Cluster.ClusterID)
		if lnErr != nil {
			blog.Errorf("checkClusterInstanceStatus[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, lnErr)
			return nil
		}

		for _, v := range nodes {
			for _, id := range instanceIDs {
				if v.Spec.ProviderID == id && cmutils.CheckNodeIfReady(v) {
					blog.Infof("checkClusterInstanceStatus[%s] node[%s] ready", taskID, id)
					running = append(running, id)
				}
			}
		}

		blog.Infof("checkClusterInstanceStatus[%s] ready nodes[%+v]", taskID, addSuccessNodes)
		if len(running) == len(instanceIDs) {
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
		addSuccessNodes, addFailureNodes, err = getVmStatus(k8sOperator, info, instanceIDs, taskID)
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

func getVmStatus(k8sOperator *clusterops.K8SOperator, info *cloudprovider.CloudDependBasicInfo, instanceIDs []string,
	taskID string) ([]string, []string, error) {
	running, failure := make([]string, 0), make([]string, 0)
	nodes, err := k8sOperator.ListClusterNodes(context.Background(), info.Cluster.ClusterID)
	if err != nil {
		blog.Errorf("checkClusterInstanceStatus[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, err)
		return nil, nil, err
	}

	for _, v := range nodes {
		for _, id := range instanceIDs {
			if v.Spec.ProviderID == id && cmutils.CheckNodeIfReady(v) {
				blog.Infof("checkClusterInstanceStatus[%s] node[%s] ready", taskID, id)
				running = append(running, id)
			} else {
				failure = append(failure, id)
			}
		}
	}

	return running, failure, nil
}
