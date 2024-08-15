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

package utils

import (
	"context"
	"fmt"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource/tresource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// resourceManager resource pool operation

// CreateResourcePoolAction create nodeGroup resource pool
func CreateResourcePoolAction(ctx context.Context, data *cloudprovider.CloudDependBasicInfo,
	pool cloudprovider.ResourcePoolData) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	consumerID, err := createResourcePool(ctx, data, pool)
	if err != nil {
		blog.Errorf("createNodeGroupAction[%s] failed: %v", taskID, err)
		return err
	}

	err = cloudprovider.UpdateNodeGroupCloudAndModuleInfo(data.NodeGroup.NodeGroupID, consumerID,
		true, data.Cluster.BusinessID)
	if err != nil {
		blog.Errorf("createNodeGroupAction[%s] UpdateNodeGroupCloudAndModuleInfo failed: %v", taskID, err)
		return err
	}

	blog.Infof("createNodeGroupAction[%s] successful", taskID)
	return nil
}

func createResourcePool(ctx context.Context, data *cloudprovider.CloudDependBasicInfo,
	pool cloudprovider.ResourcePoolData) (string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	consumerID, err := tresource.GetResourceManagerClient().CreateResourcePool(ctx, resource.ResourcePoolInfo{
		Name:      data.NodeGroup.NodeGroupID,
		Provider:  pool.Provider,
		ClusterID: data.Cluster.ClusterID,
		RelativeDevicePool: func() []string {
			if pool.ResourcePoolID == "" {
				return nil
			}
			return strings.Split(pool.ResourcePoolID, ",")
		}(),
		PoolID:   []string{pool.ResourcePoolID},
		Operator: common.ClusterManager,
	})
	if err != nil {
		blog.Errorf("task[%s] createResourcePool failed: %v", taskID, err)
		return "", err
	}

	blog.Infof("task[%s] createResourcePool successful[%s]", taskID, consumerID)
	return consumerID, nil
}

// DeleteResourcePoolAction delete nodeGroup resource pool
func DeleteResourcePoolAction(ctx context.Context, consumerId string) error {
	return tresource.GetResourceManagerClient().DeleteResourcePool(ctx, consumerId)
}

// DestroyDeviceList 销毁节点池归还机器
func DestroyDeviceList(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, deviceList []string,
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

// buildApplyCvmInstanceRequest build resource request
func buildApplyCvmInstanceRequest(group *proto.NodeGroup, operator string) *resource.ApplyInstanceReq {
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

// buildApplyIdcNodesRequest build resource request
func buildApplyIdcNodesRequest(group *proto.NodeGroup, operator string) *resource.ApplyInstanceReq {
	return &resource.ApplyInstanceReq{
		NodeType: resource.IDC,

		CPU:          group.GetLaunchTemplate().GetCPU(),
		Memory:       group.GetLaunchTemplate().GetMem(),
		Gpu:          group.GetLaunchTemplate().GetGPU(),
		Region:       group.GetRegion(),
		VpcID:        group.GetAutoScaling().GetVpcID(),
		ZoneList:     group.GetAutoScaling().GetZones(),
		SubnetList:   group.GetAutoScaling().GetSubnetIDs(),
		InstanceType: group.GetLaunchTemplate().GetInstanceType(),
		PoolID:       group.GetConsumerID(),
		Operator:     operator,
		Selector:     group.GetLaunchTemplate().GetSelector(),
	}
}

// ConsumeDevicesFromResourcePool apply cvm instances to generate orderID form resource pool
func ConsumeDevicesFromResourcePool(
	ctx context.Context, group *proto.NodeGroup, resourceType string, nodeNum int, operator string) (string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	ctx = utils.WithTraceIDForContext(ctx, taskID)

	req := &resource.ApplyInstanceReq{}
	switch resourceType {
	case resource.CVM.String():
		req = buildApplyCvmInstanceRequest(group, operator)
	case resource.IDC.String():
		req = buildApplyIdcNodesRequest(group, operator)
	default:
		return "", fmt.Errorf("task[%s] not support resourceType[%s]", taskID, resourceType)
	}

	resp, err := tresource.GetResourceManagerClient().ApplyInstances(ctx, nodeNum, req)
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

// CheckOrderStateFromResourcePool check order status from resource pool
func CheckOrderStateFromResourcePool(ctx context.Context, orderID string) (*RecordInstanceList, error) {
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
