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

// Package tasks xxx
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

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/blueking/business"
	providerutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource"
)

// ApplyNodesFromResourcePoolTask apply instance from resource
func ApplyNodesFromResourcePoolTask(taskID, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		return err
	}

	// extract valid parameter
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	desiredNodes := step.Params[cloudprovider.ScalingNodesNumKey.String()]
	scalingNum, _ := strconv.Atoi(desiredNodes)
	operator := step.Params[cloudprovider.OperatorKey.String()]

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("ApplyNodesFromResourcePoolTask[%s] GetClusterDependBasicInfo for NodeGroup %s to clean Node in task %s "+
			"step %s failed, %s", taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("getClusterDependBasicInfo failed, %s", err.Error())
		blog.Infof("ApplyNodesFromResourcePoolTask[%s] begin DeleteVirtualNodes", taskID)
		_ = cloudprovider.UpdateNodeGroupDesiredSize(nodeGroupID, scalingNum, true)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// apply Instance from ResourcePool, get instance ipList and device idList
	recordInstanceList, err := applyNodesFromResourcePool(ctx, dependInfo, scalingNum, operator)
	if err != nil {
		blog.Errorf("ApplyNodesFromResourcePoolTask[%s] requestInstancesFromPool for NodeGroup "+
			"%s step %s failed, %s", taskID, nodeGroupID, stepName, err.Error())
		retErr := fmt.Errorf("requestInstancesFromPool failed: %s", err.Error())
		_ = cloudprovider.UpdateNodeGroupDesiredSize(nodeGroupID, scalingNum, true)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	err = saveNodesToDB(ctx, dependInfo, state.Task, &NodeOptions{
		InstanceIPs: recordInstanceList.InstanceIPList,
		DeviceIDs:   recordInstanceList.DeviceIDList,
	})
	if err != nil {
		blog.Errorf("ApplyNodesFromResourcePoolTask[%s] saveClusterNodes for NodeGroup %s step %s failed, %s",
			taskID, nodeGroupID, stepName, err.Error())
		retErr := fmt.Errorf("ApplyDesiredNodesTask failed, %s", err.Error())
		_, _ = providerutils.DestroyDeviceList(ctx, dependInfo, recordInstanceList.DeviceIDList, operator)
		_ = cloudprovider.UpdateNodeGroupDesiredSize(nodeGroupID, scalingNum, true)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("ApplyNodesFromResourcePoolTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}

	return nil
}

// applyInstanceFromResourcePool 申请机器
func applyNodesFromResourcePool(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	desired int, operator string) (*providerutils.RecordInstanceList, error) {
	orderID, err := providerutils.ConsumeDevicesFromResourcePool(ctx, info.NodeGroup,
		resource.IDC.String(), desired, operator)
	if err != nil {
		return nil, err
	}

	record, err := providerutils.CheckOrderStateFromResourcePool(ctx, orderID)
	if err != nil {
		return nil, err
	}
	record.OrderID = orderID

	return record, nil
}

// NodeOptions node data
type NodeOptions struct {
	Password    string
	InstanceIDs []string
	InstanceIPs []string
	DeviceIDs   []string
}

// saveNodesToDB 存储集群节点数据
func saveNodesToDB(ctx context.Context,
	info *cloudprovider.CloudDependBasicInfo, task *proto.Task, opt *NodeOptions) error {
	var (
		nodes = make([]*proto.Node, 0)
		err   error
	)

	// deviceID Map To InstanceIP
	instanceToDeviceID := make(map[string]string)
	for i := range opt.InstanceIPs {
		if _, ok := instanceToDeviceID[opt.InstanceIPs[i]]; !ok {
			instanceToDeviceID[opt.InstanceIPs[i]] = opt.DeviceIDs[i]
		}
	}

	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	err = retry.Do(func() error {
		nodes, err = business.ListNodesByInstanceIP(opt.InstanceIPs)
		if err != nil {
			return err
		}
		return nil
	}, retry.Attempts(10))
	if err != nil {
		blog.Errorf("saveClusterNodesToDB[%s] failed: %v", taskID, err)
		return err
	}
	if len(nodes) == 0 {
		blog.Errorf("saveClusterNodesToDB[%s] cmdb sync nodes failed: %v", taskID, opt.InstanceIPs)
		return errors.New("cmdb sync nodes failed")
	}

	// update response information to task common params
	if task.CommonParams == nil {
		task.CommonParams = make(map[string]string)
	}
	if len(opt.InstanceIPs) > 0 && len(opt.DeviceIDs) > 0 {
		task.CommonParams[cloudprovider.DeviceIDsKey.String()] = strings.Join(opt.DeviceIDs, ",")
		// Job Result parameter
		task.NodeIPList = opt.InstanceIPs
		task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(opt.InstanceIPs, ",")
		task.CommonParams[cloudprovider.DynamicNodeIPListKey.String()] = strings.Join(opt.InstanceIPs, ",")
		task.CommonParams[cloudprovider.NodeIDsKey.String()] = strings.Join(opt.InstanceIDs, ",")
	}

	for _, n := range nodes {
		n.ClusterID = info.NodeGroup.ClusterID
		n.NodeGroupID = info.NodeGroup.NodeGroupID
		n.Passwd = opt.Password
		n.Status = common.StatusInitialization
		n.DeviceID = instanceToDeviceID[n.InnerIP]
		err = cloudprovider.SaveNodeInfoToDB(ctx, n, true)
		if err != nil {
			blog.Errorf("saveClusterNodesToDB[%s] SaveNodeInfoToDB[%s] failed: %v", taskID, n.InnerIP, err)
		}
	}

	return nil
}
