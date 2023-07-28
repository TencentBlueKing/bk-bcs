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

package tasks

import (
	"context"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"
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
		successInstances []model.Node
	)
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	client, err := api.NewCceClient(info.CmOption)
	if err != nil {
		return err
	}

	instaces, err := client.ListClusterNodePoolNodes(info.Cluster.SystemID, info.NodeGroup.CloudNodeGroupID)
	if err != nil {
		return err
	}

	for _, v := range instaces {
		if v.Status.Phase.Value() == model.GetNodeStatusPhaseEnum().ACTIVE.Value() {
			successInstances = append(successInstances, v)
		}
	}

	// rollback desired num
	if len(successInstances) != nodeNum {
		_ = cloudprovider.UpdateNodeGroupDesiredSize(info.NodeGroup.NodeGroupID, nodeNum-len(successInstances), true)
	}

	// record instanceIDs to task common
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	// remove existed instanceID
	var newInstances []model.Node
	for _, n := range successInstances {
		if existNode, _ := cloudprovider.GetStorageModel().GetNode(ctx, *n.Metadata.Uid); existNode != nil &&
			existNode.InnerIP != "" {
			continue
		}
		newInstances = append(newInstances, n)
	}

	// record successNodes to cluster manager DB
	nodeIPs, err := transInstancesToNode(ctx, newInstances, info)
	if err != nil {
		blog.Errorf("recordClusterInstanceToDB[%s] failed: %v", taskID, err)
	}

	if len(nodeIPs) > 0 {
		state.Task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")
	}

	return nil
}

// transInstancesToNode record success nodes to cm DB
func transInstancesToNode(ctx context.Context, instances []model.Node, info *cloudprovider.CloudDependBasicInfo) (
	[]string, error) {
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
