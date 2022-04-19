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

package bcs

import (
	"fmt"
	"sync"
	"time"

	"k8s.io/klog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/cloudprovider/bcs/clustermanager"
)

// GetNodes get nodes in the node group
type GetNodes func(ng string) ([]*clustermanager.Node, error)

// NodeGroupCache store some information about node groups.
type NodeGroupCache struct {
	registeredGroups       []*NodeGroup
	instanceToGroup        map[InstanceRef]*NodeGroup
	instanceToCreationType map[InstanceRef]CreationType
	cacheMutex             sync.Mutex
	lastUpdateTime         time.Time
	getNodes               GetNodes
}

const (
	// CreationTypeManual the node is created by manual attaching
	CreationTypeManual = "MANUAL_ATTACHING"
	// CreationTypeAuto the node is created by auto
	CreationTypeAuto = "AUTO_CREATION"

	// ScalingTypeClassic classical scaling
	ScalingTypeClassic = "CLASSIC_SCALING"
	// ScalingTypeWakeUpStopped wake up stopped scaling
	ScalingTypeWakeUpStopped = "WAKE_UP_STOPPED_SCALING"
)

// CreationType is the type of node creation
type CreationType string

// NewNodeGroupCache news node group cache
func NewNodeGroupCache(getNodes GetNodes) *NodeGroupCache {
	registry := &NodeGroupCache{
		registeredGroups:       make([]*NodeGroup, 0),
		instanceToGroup:        make(map[InstanceRef]*NodeGroup),
		instanceToCreationType: make(map[InstanceRef]CreationType),
		getNodes:               getNodes,
	}

	return registry
}

// Register registers group in node group cache
func (m *NodeGroupCache) Register(group *NodeGroup) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	m.registeredGroups = append(m.registeredGroups, group)
}

// GetRegisteredNodeGroups get all the registered node group in node group cache
func (m *NodeGroupCache) GetRegisteredNodeGroups() []*NodeGroup {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	return m.registeredGroups
}

// FindForInstance returns NodeGroup of the given Instance
func (m *NodeGroupCache) FindForInstance(instance *InstanceRef) (*NodeGroup, error) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	klog.V(5).Infof("instance :%v", instance)
	if config, found := m.instanceToGroup[*instance]; found {
		return config, nil
	}

	if err := m.regenerateCache(); err != nil {
		return nil, fmt.Errorf("Error while looking for NodeGroup for instance %+v, error: %v", *instance, err)
	}

	if config, found := m.instanceToGroup[*instance]; found {
		return config, nil
	}

	return nil, nil
}

// CheckInstancesTerminateByAs checks the instances is terminated by as or not
func (m *NodeGroupCache) CheckInstancesTerminateByAs(instances []*InstanceRef) bool {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	for _, ins := range instances {
		if m.instanceToCreationType[*ins] != CreationTypeAuto {
			klog.V(4).Infof("ins %s, CreationType %s,terminate by as direct", (*ins).Name, CreationTypeManual)
			return true
		}
	}

	return false
}

func (m *NodeGroupCache) regenerateCache() error {
	return m.regenerateCacheForInternal()
}

func (m *NodeGroupCache) regenerateCacheForInternal() error {

	now := time.Now()
	if m.lastUpdateTime.Add(3 * time.Minute).After(time.Now()) {
		klog.V(5).Infof("Refresh RegenerateCache latest updateTime %s, now %s, return",
			m.lastUpdateTime.Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05"))
		return nil
	}

	newCache := make(map[InstanceRef]*NodeGroup)
	newTypeCache := make(map[InstanceRef]CreationType)
	//groupIds := make([]*string, 0)

	for _, group := range m.registeredGroups {
		klog.V(4).Infof("Refresh Regenerating NodeGroup information for %s", group.nodeGroupID)
		groupID := group.nodeGroupID
		//groupIds = append(groupIds, &groupID)

		ins, err := m.getNodes(groupID)
		if err != nil {
			return fmt.Errorf("Failed to get nodes: %v", err)
		}

		for _, instance := range ins {
			klog.V(4).Infof("Instance %+v", instance.InnerIP)
			if instance.NodeGroupID != groupID {
				continue
			}
			ref := InstanceRef{Name: instance.NodeID}
			newCache[ref] = group
			newTypeCache[ref] = CreationType(CreationTypeAuto)
		}
		apigroup, err := group.GetNodeGroup()
		if err != nil {
			return err
		}
		ngc := apigroup.AutoScaling
		if ngc == nil {
			return fmt.Errorf("GetNodeConfig %v result is zore", groupID)
		}
		klog.V(4).Infof("AutoScaling info %+v", ngc)

		if ngc.MaxSize == 0 {
			return fmt.Errorf("do m.client.Get response max size is 0, groupId %v", groupID)
		}

		group.lk.Lock()
		if group.minSize != int(ngc.MinSize) {
			group.minSize = int(ngc.MinSize)
		}
		if group.maxSize != int(ngc.MaxSize) {
			group.maxSize = int(ngc.MaxSize)
		}
		group.lk.Unlock()
	}

	m.instanceToGroup = newCache
	m.instanceToCreationType = newTypeCache
	m.lastUpdateTime = time.Now()
	klog.V(4).Infof("Refresh RegenerateCache set latest updateTime %s", m.lastUpdateTime.Format("2006-01-02 15:04:05"))

	return nil
}

// SetNodeGroupMinSize sets minsize of the nodegroup
func (m *NodeGroupCache) SetNodeGroupMinSize(groupID string, num int) error {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	changed := false
	for i, ng := range m.registeredGroups {
		if ng.nodeGroupID == groupID {
			m.registeredGroups[i].minSize = num
			changed = true
			break
		}
	}
	if !changed {
		return fmt.Errorf("Cannot find the nodegroup %s in cache", groupID)
	}
	return nil
}
