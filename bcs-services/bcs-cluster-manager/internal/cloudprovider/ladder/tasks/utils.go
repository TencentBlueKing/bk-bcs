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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/avast/retry-go"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource/tresource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// buildApplyInstanceRequest build resource request
func buildApplyInstanceRequest(group *proto.NodeGroup, operator string) *resource.ApplyInstanceReq {
	return &resource.ApplyInstanceReq{
		NodeType: resource.CVM,

		Region:             group.GetRegion(),
		VpcID:              group.GetAutoScaling().GetVpcID(),
		ZoneList:           group.GetAutoScaling().GetZones(),
		SubnetList:         group.GetAutoScaling().GetSubnetIDs(),
		InstanceType:       group.GetLaunchTemplate().GetInstanceType(),
		CPU:                group.GetLaunchTemplate().GetCPU(),
		Memory:             group.GetLaunchTemplate().GetMem(),
		Gpu:                group.GetLaunchTemplate().GetGPU(),
		InstanceChargeType: group.GetLaunchTemplate().GetInstanceChargeType(),
		SystemDisk: resource.DataDisk{
			DiskType: group.GetLaunchTemplate().GetSystemDisk().GetDiskType(),
			DiskSize: group.GetLaunchTemplate().GetSystemDisk().GetDiskSize(),
		},
		DataDisks: func() []resource.DataDisk {
			if len(group.GetLaunchTemplate().GetDataDisks()) > 0 {
				disks := make([]resource.DataDisk, 0)
				for _, disk := range group.GetLaunchTemplate().GetDataDisks() {
					if disk == nil {
						continue
					}
					disks = append(disks, resource.DataDisk{
						DiskType: disk.GetDiskType(),
						DiskSize: disk.GetDiskSize(),
					})
				}
				return disks
			}

			return nil
		}(),
		Image: func() *resource.ImageInfo {
			var (
				imageId   = ""
				imageName = ""
			)
			if group.GetLaunchTemplate() != nil && group.GetLaunchTemplate().GetImageInfo() != nil {
				imageId = group.GetLaunchTemplate().GetImageInfo().ImageID
				imageName = group.GetLaunchTemplate().GetImageInfo().ImageName
			}
			return &resource.ImageInfo{
				ImageID:   imageId,
				ImageName: imageName,
			}
		}(),
		LoginInfo:        &resource.LoginSettings{Password: group.GetLaunchTemplate().GetInitLoginPassword()},
		SecurityGroupIds: group.GetLaunchTemplate().GetSecurityGroupIDs(),
		EnhancedService: &resource.EnhancedService{
			SecurityService: group.GetLaunchTemplate().GetIsSecurityService(),
			MonitorService:  group.GetLaunchTemplate().GetIsMonitorService(),
		},
		PoolID:   group.GetConsumerID(),
		Operator: operator,
	}
}

// 申请机器
func applyInstanceFromResourcePool(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	desired int, operator string) (*RecordInstanceList, string, error) {
	orderID, err := consumeDevicesFromResourcePool(ctx, info.NodeGroup, desired, operator)
	if err != nil {
		return nil, orderID, err
	}

	record, err := checkOrderStateFromResourcePool(ctx, orderID)
	if err != nil {
		return nil, orderID, err
	}
	record.OrderID = orderID

	return record, orderID, nil
}

// consumeDevicesFromResourcePool apply cvm instances to generate orderID form resource pool
func consumeDevicesFromResourcePool(
	ctx context.Context, group *proto.NodeGroup, nodeNum int, operator string) (string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	ctx = utils.WithTraceIDForContext(ctx, taskID)
	resp, err := tresource.GetResourceManagerClient().ApplyInstances(ctx, nodeNum,
		buildApplyInstanceRequest(group, operator))
	if err != nil {
		blog.Errorf("consumeDevicesFromResourcePool[%s] ApplyInstances failed: %v", taskID, err)
		return "", err
	}

	blog.Infof("consumeDevicesFromResourcePool[%s] success", taskID)
	return resp.OrderID, nil
}

// RecordInstanceList xxx
type RecordInstanceList struct {
	OrderID        string
	InstanceIPList []string
	InstanceIDList []string
	DeviceIDList   []string
}

func checkOrderStateFromResourcePool(ctx context.Context, orderID string) (*RecordInstanceList, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	ctx = utils.WithTraceIDForContext(ctx, taskID)
	result, err := tresource.GetResourceManagerClient().CheckOrderStatus(ctx, orderID)
	if err != nil {
		blog.Errorf("checkOrderStateFromResourcePool[%s] CheckOrderStatus[%s] failed: %v", taskID, orderID, err)
		return nil, err
	}

	// get device instanceIDs & instanceIPs
	if len(result.InstanceIDs) == 0 || len(result.InstanceIPs) == 0 {
		retErr := fmt.Errorf("checkOrderStateFromResourcePool[%s] return instance empty", taskID)
		blog.Errorf(retErr.Error())
		return nil, retErr
	}

	return &RecordInstanceList{
		InstanceIPList: result.InstanceIPs,
		InstanceIDList: result.InstanceIDs,
		DeviceIDList:   result.ExtraIDs,
	}, nil
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

// 销毁归还机器
func destroyDeviceList(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, deviceList []string,
	operator string) (string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	if info == nil || info.NodeGroup == nil || info.Cluster == nil || len(deviceList) == 0 {
		return "", fmt.Errorf("destroyDeviceList[%s] lost validate info", taskID)
	}

	ctx = utils.WithTraceIDForContext(ctx, taskID)
	resp, err := tresource.GetResourceManagerClient().DestroyInstances(ctx, &resource.DestroyInstanceReq{
		PoolID:      info.NodeGroup.GetConsumerID(),
		SystemID:    info.Cluster.GetSystemID(),
		InstanceIDs: deviceList,
		Operator:    operator,
	})
	if err != nil {
		blog.Errorf("destroyDeviceList[%s] DestroyInstances failed: %v", taskID, err)
		return "", err
	}

	blog.Infof("destroyDeviceList[%s] call DestroyInstances successfully, orders %v.", resp.OrderID)
	return resp.OrderID, nil
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
	orderID, err := destroyDeviceList(ctx, info, deviceIDs, operator)
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
