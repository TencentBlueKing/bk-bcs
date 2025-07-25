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

package business

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
)

// NodeAdvancedOptions node advanced options
type NodeAdvancedOptions struct {
	NodeScheduler bool
	Disks         []*proto.CloudDataDisk
}

// instance nodes disk info

// InstanceInfo instance info
type InstanceInfo struct {
	InstanceID string
	InstanceIP string
}

// GetInstanceIPs get instance ip
func GetInstanceIPs(ins []InstanceInfo) []string {
	ips := make([]string, 0)
	for i := range ins {
		ips = append(ips, ins[i].InstanceIP)
	}
	return ips
}

// GetInstanceIDs get instance id
func GetInstanceIDs(ins []InstanceInfo) []string {
	ids := make([]string, 0)
	for i := range ins {
		ids = append(ids, ins[i].InstanceIP)
	}
	return ids
}

// InstanceDisk xxx
type InstanceDisk struct {
	InstanceInfo
	DiskCount int
}

// GetNodeInstanceDataDiskInfo get node instance dataDisks
func GetNodeInstanceDataDiskInfo(
	instanceIDs []string, opt *cloudprovider.CommonOption) (map[string]InstanceDisk, error) {
	client, err := api.GetCVMClient(opt)
	if err != nil {
		return nil, err
	}

	instanceList, err := client.GetInstancesByID(instanceIDs)
	if err != nil {
		blog.Errorf("GetNodeInstanceDataDiskInfo[%+v] failed: %v", instanceIDs, err)
		return nil, err
	}

	instances := make(map[string]InstanceDisk, 0)
	for _, cvm := range instanceList {
		instances[*cvm.InstanceId] = InstanceDisk{
			InstanceInfo: InstanceInfo{
				InstanceID: *cvm.InstanceId,
				InstanceIP: *cvm.PrivateIpAddresses[0],
			},
			DiskCount: len(cvm.DataDisks),
		}
	}

	return instances, nil
}

// FilterInstanceByDataDisk xxx
type FilterInstanceByDataDisk struct {
	SingleDiskInstance   []string
	SingleDiskInstanceIP []string
	ManyDiskInstance     []string
	ManyDiskInstanceIP   []string
}

// FilterNodesByDataDisk filter instance by data disks
func FilterNodesByDataDisk(instanceIDs []string, opt *cloudprovider.CommonOption) (*FilterInstanceByDataDisk, error) {
	instanceDisk, err := GetNodeInstanceDataDiskInfo(instanceIDs, opt)
	if err != nil {
		blog.Errorf("FilterNodesByDataDisk GetNodeInstanceDataDiskInfo failed: %v", err)
		return nil, err
	}

	filter := &FilterInstanceByDataDisk{
		SingleDiskInstance:   make([]string, 0),
		SingleDiskInstanceIP: make([]string, 0),
		ManyDiskInstance:     make([]string, 0),
		ManyDiskInstanceIP:   make([]string, 0),
	}

	for i := range instanceDisk {
		if instanceDisk[i].DiskCount <= 1 {
			filter.SingleDiskInstance = append(filter.SingleDiskInstance, instanceDisk[i].InstanceID)
			filter.SingleDiskInstanceIP = append(filter.SingleDiskInstanceIP, instanceDisk[i].InstanceIP)
			continue
		}
		filter.ManyDiskInstance = append(filter.ManyDiskInstance, instanceDisk[i].InstanceID)
		filter.ManyDiskInstanceIP = append(filter.ManyDiskInstanceIP, instanceDisk[i].InstanceIP)
	}

	return filter, nil
}

// InternetConnect for cluster connection kubeconfig
type InternetConnect struct {
	// InternetAccessible xxx
	InternetAccessible struct {
		InternetChargeType      string `json:"InternetChargeType"`
		InternetMaxBandwidthOut int    `json:"InternetMaxBandwidthOut"`
	} `json:"InternetAccessible"`
	// VipIsp xxx
	VipIsp string `json:"VipIsp,omitempty"`
	// BandwidthPackageId xxx
	BandwidthPackageId string `json:"BandwidthPackageId,omitempty"`
}
