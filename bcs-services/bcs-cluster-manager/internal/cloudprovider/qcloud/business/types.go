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
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// NodeAdvancedOptions node advanced options
type NodeAdvancedOptions struct {
	NodeScheduler         bool
	SetPreStartUserScript bool
	CreateCluster         bool
	Advance               *proto.NodeAdvancedInfo
}

// ClusterCommonLabels cluster common labels
func ClusterCommonLabels(cluster *proto.Cluster) map[string]string {
	labels := make(map[string]string)
	if len(cluster.Region) > 0 {
		regions := strings.Split(cluster.Region, "-")
		if len(regions) >= 1 {
			labels[utils.RegionLabelKey] = regions[1]
		}
	}

	return labels
}

// instance nodes disk info

// InstanceDisk xxx
type InstanceDisk struct {
	InstanceID string
	InstanceIP string
	DiskCount  int
}

// GetNodeInstanceDataDiskInfo get node instance dataDisks
func GetNodeInstanceDataDiskInfo(
	instanceIDs []string, opt *cloudprovider.CommonOption) (map[string]InstanceDisk, error) {
	client, err := api.GetCVMClient(opt)
	if err != nil {
		return nil, err
	}

	instanceList, err := client.GetInstancesById(instanceIDs)
	if err != nil {
		blog.Errorf("GetNodeInstanceDataDiskInfo[%+v] failed: %v", instanceIDs, err)
		return nil, err
	}

	instances := make(map[string]InstanceDisk, 0)
	for _, cvm := range instanceList {
		instances[*cvm.InstanceId] = InstanceDisk{
			InstanceID: *cvm.InstanceId,
			InstanceIP: *cvm.PrivateIpAddresses[0],
			DiskCount:  len(cvm.DataDisks),
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

// nolint : type `subnetIpNum` is unused
type subnetIpNum struct {
	subnetId string
	cnt      uint64
}
