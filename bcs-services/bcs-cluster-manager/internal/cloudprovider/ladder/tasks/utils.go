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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/avast/retry-go"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/business"
	providerutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource"
)

// 申请机器
func applyInstanceFromResourcePool(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	state *cloudprovider.TaskState, oldOrderId string, desired int, operator string) (
	*providerutils.RecordInstanceList, string, error) {
	var (
		orderID string
		err     error
	)

	getOrderId := func() string {
		if len(orderID) > 0 {
			return orderID
		}

		return oldOrderId
	}

	// check if already submit old task
	if len(oldOrderId) == 0 {
		orderID, err = providerutils.ConsumeDevicesFromResourcePool(ctx, info.NodeGroup, resource.CVM.String(),
			desired, operator)
		if err != nil {
			return nil, orderID, err
		}
	}

	if len(orderID) > 0 {
		state.Task.CommonParams[cloudprovider.DeviceRecordIDKey.String()] = orderID
		_ = cloudprovider.GetStorageModel().UpdateTask(context.Background(), state.Task)
	}

	record, err := providerutils.CheckOrderStateFromResourcePool(ctx, getOrderId())
	if err != nil {
		return nil, getOrderId(), err
	}
	record.OrderID = getOrderId()

	return record, getOrderId(), nil
}

// RecordInstanceToDBOption xxx
type RecordInstanceToDBOption struct {
	Password    string
	InstanceIDs []string
	DeviceIDs   []string
}

// 录入机器
func recordClusterCVMInfoToDB(
	ctx context.Context, info *cloudprovider.CloudDependBasicInfo, opt *RecordInstanceToDBOption) error {

	var (
		nodes = make([]*proto.Node, 0)
		err   error
	)

	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	blog.Infof("recordClusterCVMInfoToDB[%s] devices[%+v] instanceIDs[%+v]", taskID, opt.DeviceIDs, opt.InstanceIDs)

	// deviceID Map To InstanceID
	instanceToDeviceID := make(map[string]string)
	for i := range opt.InstanceIDs {
		if _, ok := instanceToDeviceID[opt.InstanceIDs[i]]; !ok {
			instanceToDeviceID[opt.InstanceIDs[i]] = opt.DeviceIDs[i]
		}
	}

	err = retry.Do(func() error {
		nodes, err = business.ListNodesByInstanceID(opt.InstanceIDs, &cloudprovider.ListNodesOption{
			Common:       info.CmOption,
			ClusterVPCID: info.Cluster.VpcID,
		})
		if err != nil {
			return err
		}
		return nil
	}, retry.Attempts(3))
	if err != nil {
		blog.Errorf("recordClusterCVMInfoToDB[%s] ListNodesByInstanceID failed: %v", taskID, err)
		return err
	}
	blog.Infof("recordClusterCVMInfoToDB[%s] ListNodesByInstanceID[%+v] success", taskID, opt.InstanceIDs)

	for _, n := range nodes {
		blog.Infof("recordClusterCVMInfoToDB[%s] clusterID[%s] nodeGroupID[%s] "+
			"nodeID[%s] nodeIP[%s]", taskID, info.NodeGroup.ClusterID, info.NodeGroup.NodeGroupID, n.NodeID, n.InnerIP)
		n.ClusterID = info.NodeGroup.ClusterID
		n.NodeGroupID = info.NodeGroup.NodeGroupID
		n.Passwd = opt.Password
		n.Status = common.StatusInitialization
		n.DeviceID = instanceToDeviceID[n.NodeID]
		err = cloudprovider.SaveNodeInfoToDB(ctx, n, false)
		if err != nil {
			blog.Errorf("transInstancesToNode[%s] SaveNodeInfoToDB[%s:%s] failed: %v",
				taskID, n.InnerIP, n.NodeID, err)
			continue
		}

		blog.Infof("transInstancesToNode[%s] SaveNodeInfoToDB[%s:%s] success",
			taskID, n.InnerIP, n.NodeID)
	}

	return nil
}

// returnDevicesToRMAndCleanNodes need to handle some operations when failure
// 1. clean cluster-manager data / update desired value
// 2. delete cluster nodes
// 3. return devices to resourceManager module
func returnDevicesToRMAndCleanNodes(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, instanceIDs []string,
	delInstance bool, operator string) error { // nolint
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	if info == nil || len(instanceIDs) == 0 {
		blog.Infof("returnDevicesToRMAndCleanNodes[%s] info null or instanceIDs empty", taskID)
		return nil
	}

	nodes := cloudprovider.GetNodesByInstanceIDs(instanceIDs)
	var deviceIDs []string
	for _, n := range nodes {
		deviceIDs = append(deviceIDs, n.DeviceID)
	}

	// delete db data record
	for _, instanceID := range instanceIDs {
		err := cloudprovider.GetStorageModel().DeleteClusterNode(context.Background(), info.Cluster.ClusterID,
			instanceID)
		if err != nil {
			blog.Errorf("returnDevicesToRMAndCleanNodes[%s] DeleteClusterNode[%s] failed: %v", taskID,
				instanceID, err)
		} else {
			blog.Infof("returnDevicesToRMAndCleanNodes[%s] DeleteClusterNode success[%+v]", taskID, instanceID)
		}
	}

	// delete cluster instances
	if delInstance {
		successIDs, err := business.DeleteClusterInstance(ctx, info, instanceIDs, true)
		if err != nil {
			blog.Errorf("returnDevicesToRMAndCleanNodes[%s] DeleteClusterInstance failed: %v", taskID, err)
		} else {
			blog.Infof("returnDevicesToRMAndCleanNodes[%s] DeleteClusterInstance success[%v]", taskID, successIDs)
		}
	}

	// destroy device to resource manager
	orderID, err := providerutils.DestroyDeviceList(ctx, info, deviceIDs, operator)
	if err != nil {
		blog.Errorf("returnDevicesToRMAndCleanNodes[%s] destroyDeviceList failed: %v", taskID, err)
	} else {
		blog.Infof("returnDevicesToRMAndCleanNodes[%s] successful[%v] orderID[%v]", taskID, instanceIDs, orderID)
	}

	// rollback nodeGroup desired size
	err = cloudprovider.UpdateNodeGroupDesiredSize(info.NodeGroup.NodeGroupID, len(instanceIDs), true)
	if err != nil {
		blog.Errorf("returnDevicesToRMAndCleanNodes[%s] UpdateNodeGroupDesiredSize failed: %v", taskID, err)
	} else {
		blog.Infof("returnDevicesToRMAndCleanNodes[%s] UpdateNodeGroupDesiredSize success[%v]", taskID, len(instanceIDs))
	}

	return nil
}
