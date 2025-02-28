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

package quota

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/time"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// TransStore2ProtoQuota trans store quota to proto quota
func TransStore2ProtoQuota(pQuota *ProjectQuota) *proto.ProjectQuota {
	return &proto.ProjectQuota{
		QuotaId:      pQuota.QuotaId,
		QuotaName:    pQuota.QuotaName,
		ProjectID:    pQuota.ProjectId,
		ProjectCode:  pQuota.ProjectCode,
		ClusterId:    pQuota.ClusterId,
		ClusterName:  "",
		NameSpace:    pQuota.Namespace,
		BusinessID:   pQuota.BusinessId,
		BusinessName: "",
		Description:  pQuota.Description,
		IsDeleted:    pQuota.IsDeleted,
		QuotaType:    pQuota.QuotaType.String(),
		Quota:        TransStore2ProtoQuotaResource(pQuota.Quota),
		Status:       pQuota.Status.String(),
		Message:      "",
		CreateTime:   time.TransTsToStr(pQuota.CreateTime),
		UpdateTime:   time.TransTsToStr(pQuota.UpdateTime),
		Creator:      pQuota.Creator,
		Updater:      pQuota.Updater,
		Provider:     pQuota.Provider,
		Labels:       pQuota.Labels,
		Annotations:  pQuota.Annotations,
	}
}

func dataDiskToDeviceDisk(disk *proto.DataDisk) DeviceDisk {
	return DeviceDisk{
		Type: disk.GetDiskType(),
		Size: disk.GetDiskSize(),
	}
}

func dataDisksToDeviceDisks(disks []*proto.DataDisk) []DeviceDisk {
	deviceDisks := make([]DeviceDisk, 0)

	for i := range disks {
		deviceDisks = append(deviceDisks, dataDiskToDeviceDisk(disks[i]))
	}
	return deviceDisks
}

func deviceDiskToDataDisk(disk DeviceDisk) *proto.DataDisk {
	return &proto.DataDisk{
		DiskType: disk.Type,
		DiskSize: disk.Type,
	}
}

func deviceDisksToDataDisks(disks []DeviceDisk) []*proto.DataDisk {
	dataDisks := make([]*proto.DataDisk, 0)

	for i := range disks {
		if disks[i].Type == "" || disks[i].Size == "" {
			continue
		}

		dataDisks = append(dataDisks, deviceDiskToDataDisk(disks[i]))
	}
	return dataDisks
}

// TransPorto2StoreQuota trans quota resource to store quota resource
func TransPorto2StoreQuota(quota *proto.QuotaResource) *QuotaResource {
	storeQuotaResource := &QuotaResource{}

	if quota.GetZoneResources() != nil {
		storeQuotaResource.HostResources = &HostConfig{
			Region:       quota.GetZoneResources().GetRegion(),
			InstanceType: quota.GetZoneResources().GetInstanceType(),
			Cpu:          quota.GetZoneResources().GetCpu(),
			Mem:          quota.GetZoneResources().GetMem(),
			Gpu:          quota.GetZoneResources().GetGpu(),
			ZoneId:       quota.GetZoneResources().GetZoneId(),
			ZoneName:     quota.GetZoneResources().GetZoneName(),
			QuotaNum:     quota.GetZoneResources().GetQuotaNum(),
			SystemDisk:   dataDiskToDeviceDisk(quota.GetZoneResources().GetSystemDisk()),
			DataDisks:    dataDisksToDeviceDisks(quota.GetZoneResources().GetDataDisks()),
		}
	}

	if quota.GetCpu() != nil {
		storeQuotaResource.Cpu = &DeviceInfo{
			DeviceType:  quota.GetCpu().GetDeviceType(),
			DeviceQuota: quota.GetCpu().GetDeviceQuota(),
			Attributes:  quota.GetCpu().GetAttributes(),
		}
	}

	if quota.GetMem() != nil {
		storeQuotaResource.Mem = &DeviceInfo{
			DeviceType:  quota.GetMem().GetDeviceType(),
			DeviceQuota: quota.GetMem().GetDeviceQuota(),
			Attributes:  quota.GetMem().GetAttributes(),
		}
	}

	if quota.GetGpu() != nil {
		storeQuotaResource.Gpu = &DeviceInfo{
			DeviceType:  quota.GetGpu().GetDeviceType(),
			DeviceQuota: quota.GetGpu().GetDeviceQuota(),
			Attributes:  quota.GetGpu().GetAttributes(),
		}
	}

	return storeQuotaResource
}

// TransStore2ProtoQuotaResource trans store quota resource to proto quota resource
func TransStore2ProtoQuotaResource(quota *QuotaResource) *proto.QuotaResource {
	protoQuotaResource := &proto.QuotaResource{}

	if quota.HostResources != nil {
		protoQuotaResource.ZoneResources = &proto.InstanceTypeConfig{
			Region:       quota.HostResources.Region,
			InstanceType: quota.HostResources.InstanceType,
			Cpu:          quota.HostResources.Cpu,
			Mem:          quota.HostResources.Mem,
			Gpu:          quota.HostResources.Gpu,
			ZoneId:       quota.HostResources.ZoneId,
			ZoneName:     quota.HostResources.ZoneName,
			QuotaNum:     quota.HostResources.QuotaNum,
			SystemDisk:   deviceDiskToDataDisk(quota.HostResources.SystemDisk),
			DataDisks:    deviceDisksToDataDisks(quota.HostResources.DataDisks),
		}
	}

	if quota.Cpu != nil {
		protoQuotaResource.Cpu = &proto.DeviceInfo{
			DeviceType:      quota.Cpu.DeviceType,
			DeviceQuota:     quota.Cpu.DeviceQuota,
			DeviceQuotaUsed: quota.Cpu.DeviceQuotaUsed,
			Attributes:      quota.Cpu.Attributes,
		}
	}

	if quota.Mem != nil {
		protoQuotaResource.Mem = &proto.DeviceInfo{
			DeviceType:      quota.Mem.DeviceType,
			DeviceQuota:     quota.Mem.DeviceQuota,
			DeviceQuotaUsed: quota.Mem.DeviceQuotaUsed,
			Attributes:      quota.Mem.Attributes,
		}
	}

	if quota.Gpu != nil {
		protoQuotaResource.Gpu = &proto.DeviceInfo{
			DeviceType:      quota.Gpu.DeviceType,
			DeviceQuota:     quota.Gpu.DeviceQuota,
			DeviceQuotaUsed: quota.Gpu.DeviceQuotaUsed,
			Attributes:      quota.Gpu.Attributes,
		}
	}

	return protoQuotaResource
}
