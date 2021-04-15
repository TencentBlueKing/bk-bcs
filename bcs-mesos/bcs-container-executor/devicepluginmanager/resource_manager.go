/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package devicepluginmanager

import (
	"fmt"

	comtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/extendedresource"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// ResourceManager manager for extended resources
type ResourceManager struct {
	devicePluginM          *DevicePluginManager
	extendedResourceDriver *extendedresource.Driver
}

// NewResourceManager create resource manager
func NewResourceManager(devicePluginM *DevicePluginManager,
	extendedResourceDriver *extendedresource.Driver) *ResourceManager {
	return &ResourceManager{
		devicePluginM:          devicePluginM,
		extendedResourceDriver: extendedResourceDriver,
	}
}

// ApplyExtendedResources apply extended resource, records to file, and call device plugin
// return envs after device plugin Allocate
func (rm *ResourceManager) ApplyExtendedResources(
	ex *comtypes.ExtendedResource, taskID string) (map[string]string, error) {
	if err := rm.extendedResourceDriver.Lock(); err != nil {
		return nil, err
	}
	defer rm.extendedResourceDriver.Unlock()

	allocateMap, err := rm.extendedResourceDriver.ListRecordByResourceType(ex.Name)
	if err != nil {
		return nil, err
	}

	devices, err := rm.devicePluginM.ListAndWatch(ex)
	if err != nil {
		return nil, err
	}
	deviceIDs := getDevicesIDList(devices)

	var allocateIDs []string
	var envs map[string]string
	if rm.deviceHasTopology(devices) {
		allocateIDs, err = getAllocateDeviceIDsByTopology(devices, allocateMap, int(ex.Value))
	} else {
		allocateIDs, err = getAllocateDeviceIDs(deviceIDs, allocateMap, int(ex.Value))
	}
	if err != nil {
		return nil, err
	}
	err = rm.extendedResourceDriver.AddRecord(ex.Name, taskID, allocateIDs)
	if err != nil {
		return nil, err
	}
	envs, err = rm.devicePluginM.Allocate(ex, allocateIDs)
	if err != nil {
		return nil, err
	}
	return envs, nil
}

// ReleaseExtendedResources release extended resource allocation in executor
func (rm *ResourceManager) ReleaseExtendedResources(exName, taskID string) error {
	if err := rm.extendedResourceDriver.Lock(); err != nil {
		return err
	}
	defer rm.extendedResourceDriver.Unlock()

	return rm.extendedResourceDriver.DelRecord(exName, taskID)
}

func getDevicesIDList(devices []*pluginapi.Device) []string {
	retList := make([]string, 0)
	for _, device := range devices {
		retList = append(retList, device.ID)
	}
	return retList
}

func getAllocateDeviceIDs(allIDs []string, existMap map[string]string, resourceNum int) ([]string, error) {
	var availableIDs []string
	for _, id := range allIDs {
		if _, ok := existMap[id]; !ok {
			availableIDs = append(availableIDs, id)
		}
	}
	if resourceNum > len(availableIDs) {
		return nil, fmt.Errorf("no enought devices for demand %d", resourceNum)
	}
	return availableIDs[0:resourceNum], nil
}

func getAllocateDeviceIDsByTopology(
	devices []*pluginapi.Device, existMap map[string]string, resourceNum int) ([]string, error) {
	deviceIDMap := make(map[int64][]string)
	for _, device := range devices {
		if _, ok := existMap[device.ID]; ok {
			continue
		}
		if device.Topology == nil {
			return nil, fmt.Errorf("device %v has no topology", device)
		}
		for _, node := range device.Topology.Nodes {
			if _, ok := deviceIDMap[node.ID]; !ok {
				deviceIDMap[node.ID] = make([]string, 0)
			}
			deviceIDMap[node.ID] = append(deviceIDMap[node.ID], device.ID)
		}
	}
	for _, idList := range deviceIDMap {
		if resourceNum > len(idList) {
			continue
		}
		return idList[0:resourceNum], nil
	}
	return nil, fmt.Errorf("no enough devices in numa")
}

func (rm *ResourceManager) deviceHasTopology(devices []*pluginapi.Device) bool {
	for _, device := range devices {
		if device.Topology != nil {
			return true
		}
	}
	return false
}
