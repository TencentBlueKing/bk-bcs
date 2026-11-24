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

package core

import (
	"fmt"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"
	"k8s.io/klog"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/scalingconfig"
)

const (
	// NodeUpcomingAnnotation is the annotation marking upcoming nodes
	NodeUpcomingAnnotation = "cluster-autoscaler.kubernetes.io/upcoming"
)

// NodeGroupSetManager 管理 NodeGroupSet 的分组和资源计算
type NodeGroupSetManager struct {
	mu sync.RWMutex

	// setConfig 配置的所有 NodeGroupSet
	setConfig *scalingconfig.NodeGroupSetList

	// nodeGroupToSet 快速查找 NodeGroup 属于哪个 Set
	// key: NodeGroup ID, value: Set Name
	nodeGroupToSet map[string]string

	// setNodeGroups 每个 Set 包含的 NodeGroup 对象
	// key: Set Name, value: NodeGroup 列表
	setNodeGroups map[string][]cloudprovider.NodeGroup
}

// NewNodeGroupSetManager 创建 NodeGroupSetManager
func NewNodeGroupSetManager(config *scalingconfig.NodeGroupSetList) *NodeGroupSetManager {
	return &NodeGroupSetManager{
		setConfig:      config,
		nodeGroupToSet: make(map[string]string),
		setNodeGroups:  make(map[string][]cloudprovider.NodeGroup),
	}
}

// RefreshNodeGroups 刷新 NodeGroup 到 Set 的映射关系
// 在每次 CA 主循环开始时调用,确保动态变化的 NodeGroup 能正确分组
func (m *NodeGroupSetManager) RefreshNodeGroups(allNodeGroups []cloudprovider.NodeGroup) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 重置映射
	m.nodeGroupToSet = make(map[string]string)
	m.setNodeGroups = make(map[string][]cloudprovider.NodeGroup)

	// 为每个配置的 Set 初始化
	for _, set := range m.setConfig.Sets {
		m.setNodeGroups[set.SetName] = make([]cloudprovider.NodeGroup, 0)
	}

	// 遍历所有 NodeGroup,分配到对应的 Set
	for _, ng := range allNodeGroups {
		setName := m.findMatchingSet(ng)
		m.nodeGroupToSet[ng.Id()] = setName
		m.setNodeGroups[setName] = append(m.setNodeGroups[setName], ng)
	}

	// 记录分组结果
	for setName, ngs := range m.setNodeGroups {
		klog.V(4).Infof("NodeGroupSet %s contains %d node groups", setName, len(ngs))
	}

	return nil
}

const (
	// GlobalSetName 是一个特殊的 Set 名称，表示使用全局集群级别的配置
	// 当 NodeGroup 不属于任何显式配置的 Set 时使用
	GlobalSetName = "__global__"
)

// findMatchingSet 查找 NodeGroup 匹配的 Set
func (m *NodeGroupSetManager) findMatchingSet(ng cloudprovider.NodeGroup) string {
	// 1. 检查显式配置的 NodeGroupIDs
	for _, set := range m.setConfig.Sets {
		for _, id := range set.NodeGroupIDs {
			if ng.Id() == id {
				return set.SetName
			}
		}
	}

	// 2. 如果未匹配到任何 Set，返回 GlobalSetName
	// 表示该 NodeGroup 应该使用全局集群级别的 buffer ratio 和资源计算
	return GlobalSetName
}

// GetSetConfig 获取指定 Set 的配置
func (m *NodeGroupSetManager) GetSetConfig(setName string) (*scalingconfig.NodeGroupSet, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// GlobalSetName 表示使用全局配置，不是一个真实的 Set
	if setName == GlobalSetName {
		return nil, fmt.Errorf("set %s is a special marker for global configuration, not a real set", setName)
	}

	for _, set := range m.setConfig.Sets {
		if set.SetName == setName {
			return &set, nil
		}
	}

	return nil, fmt.Errorf("set %s not found", setName)
}

// GetNodeGroupsInSet 获取指定 Set 包含的所有 NodeGroup
func (m *NodeGroupSetManager) GetNodeGroupsInSet(setName string) []cloudprovider.NodeGroup {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// GlobalSetName 不是一个真实的 Set，不应返回任何 NodeGroup
	if setName == GlobalSetName {
		return nil
	}

	return m.setNodeGroups[setName]
}

// GetAllSets 获取所有 Set 名称
func (m *NodeGroupSetManager) GetAllSets() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]string, 0, len(m.setConfig.Sets))
	for _, set := range m.setConfig.Sets {
		result = append(result, set.SetName)
	}

	return result
}

// GetNodeGroupSet 获取 NodeGroup 所属的 Set 名称
func (m *NodeGroupSetManager) GetNodeGroupSet(nodeGroupID string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	setName, exists := m.nodeGroupToSet[nodeGroupID]
	if !exists {
		return ""
	}
	return setName
}

// NodeGroupSetResourceInfo 单个 Set 的资源使用信息
type NodeGroupSetResourceInfo struct {
	SetName          string
	TotalCPU         int64   // 总 CPU(毫核)
	AvailableCPU     int64   // 可用 CPU(毫核)
	TotalMemory      int64   // 总内存(字节)
	AvailableMemory  int64   // 可用内存(字节)
	CPUUsageRatio    float64 // CPU 使用率
	MemUsageRatio    float64 // 内存使用率
	TriggeredScaleUp bool    // 是否触发扩容
}

// SetResourceSnapshot 资源快照,用于性能优化
// 避免在 Scale-Down 时对每个节点都重新计算整个 Set 的资源状态
type SetResourceSnapshot struct {
	Timestamp time.Time
	// key: Set Name, value: 该 Set 的资源信息
	Resources map[string]*NodeGroupSetResourceInfo
	// key: Node Name, value: 该节点的资源容量
	NodeCapacities map[string]*NodeResourceCapacity
	// key: Node Name, value: NodeGroup ID
	// 预构建的映射,避免重复调用 CloudProvider.NodeGroupForNode()
	NodeToGroup map[string]string
}

// IsStale 检查快照是否过期(TTL: 10秒)
func (s *SetResourceSnapshot) IsStale() bool {
	return time.Since(s.Timestamp) > 10*time.Second
}

// NodeResourceCapacity 单个节点的资源容量
type NodeResourceCapacity struct {
	NodeName          string
	AllocatableCPU    int64 // 可分配 CPU(毫核)
	AllocatableMemory int64 // 可分配内存(字节)
	UsedCPU           int64 // 已使用 CPU(毫核)
	UsedMemory        int64 // 已使用内存(字节)
}

// CheckScaleUpNeeded 检查指定 Set 是否需要扩容
func (m *NodeGroupSetManager) CheckScaleUpNeeded(
	setName string,
	resourceInfo *NodeGroupSetResourceInfo,
) bool {
	setConfig, err := m.GetSetConfig(setName)
	if err != nil {
		klog.Errorf("Failed to get set config for %s: %v", setName, err)
		return false
	}

	config := setConfig.BufferedRatios

	// 当 Set 没有任何节点时(冷启动场景)
	// 只有当配置了非零 buffer ratio 时，才需要保留至少1个节点
	// 如果所有 ratio 都为 0，则允许为空
	if resourceInfo.TotalCPU == 0 && resourceInfo.TotalMemory == 0 {
		// 检查是否配置了任何非零 buffer ratio
		if config.CPURatio > 0 || config.MemRatio > 0 || config.ResourceRatio > 0 {
			klog.V(2).Infof("Set %s has no nodes but buffer ratio configured, scale up needed",
				setName)
			return true
		}
		// 所有 ratio 都为 0，允许 Set 为空
		klog.V(2).Infof("Set %s has no nodes and all buffer ratios are 0, scale up not needed",
			setName)
		return false
	}

	// 计算可用率(而非使用率)
	cpuAvailableRatio := 1.0 - resourceInfo.CPUUsageRatio
	memAvailableRatio := 1.0 - resourceInfo.MemUsageRatio

	// 检查是否低于缓冲比例
	if config.CPURatio > 0 && cpuAvailableRatio < config.CPURatio {
		klog.V(2).Infof("Set %s CPU available ratio %.2f < threshold %.2f, scale up needed",
			setName, cpuAvailableRatio, config.CPURatio)
		return true
	}

	if config.MemRatio > 0 && memAvailableRatio < config.MemRatio {
		klog.V(2).Infof("Set %s Memory available ratio %.2f < threshold %.2f, scale up needed",
			setName, memAvailableRatio, config.MemRatio)
		return true
	}

	return false
}

// TakeResourceSnapshot 创建资源快照,用于性能优化
// 一次性计算所有 Set 的资源状态和每个节点的资源容量
// 扩容快照包括 upcoming nodes,避免重复扩容
// 缩容快照排除 upcoming nodes,避免缩容依赖虚拟容量
// NOCC:golint/fnsize(设计如此)
// nolint
func (m *NodeGroupSetManager) TakeResourceSnapshot(
	allNodeInfos map[string]*schedulernodeinfo.NodeInfo,
	cloudProvider cloudprovider.CloudProvider,
) (*SetResourceSnapshot, *SetResourceSnapshot, error) {
	scaleUpSnapshot := newSetResourceSnapshot()
	scaleDownSnapshot := newSetResourceSnapshot()

	klog.V(4).Infof("[NodeToGroup] Building node-to-nodegroup mapping for %d nodes", len(allNodeInfos))

	// allNodeInfos 来自 ClusterSnapshot,已包含真实节点 + upcoming 虚拟节点
	// 计算每个节点的资源容量(一次遍历)
	for nodeName, nodeInfo := range allNodeInfos {
		node := nodeInfo.Node()
		if node == nil {
			continue
		}
		if node.Spec.Unschedulable {
			continue
		}
		if deletetaint.HasToBeDeletedTaint(node) {
			continue
		}
		// 跳过特殊节点
		if node.Labels["node.kubernetes.io/instance-type"] == "eklet" {
			continue
		}
		if node.Annotations[filterNodeResourceAnnoKey] == valueTrue {
			continue
		}
		if node.Labels["node-role.kubernetes.io/controlplane"] == valueTrue {
			continue
		}
		if node.Labels["node-role.kubernetes.io/master"] == "true" {
			continue
		}

		isUpcomingNode := node.Annotations != nil &&
			node.Annotations[NodeUpcomingAnnotation] == valueTrue

		// 构建 Node → NodeGroup 映射
		var nodeGroupID string
		if isUpcomingNode {
			// Upcoming 节点:直接读取 NodeGroup ID Annotation
			nodeGroupID = node.Annotations[bcsNodeGroupAnnoKey]
		} else {
			// 真实节点:调用 CloudProvider API (只调用一次)
			nodeGroup, err := cloudProvider.NodeGroupForNode(node)
			if err == nil && nodeGroup != nil {
				nodeGroupID = nodeGroup.Id()
			}
		}

		usedCPU := int64(0)
		usedMemory := int64(0)
		for _, podInfo := range nodeInfo.Pods() {
			for _, container := range podInfo.Spec.Containers {
				if cpu := container.Resources.Requests.Cpu(); cpu != nil {
					usedCPU += cpu.MilliValue()
				}
				if mem := container.Resources.Requests.Memory(); mem != nil {
					usedMemory += mem.Value()
				}
			}
		}

		allocatable := nodeInfo.AllocatableResource()
		nodeCapacity := &NodeResourceCapacity{
			NodeName:          nodeName,
			AllocatableCPU:    allocatable.MilliCPU,
			AllocatableMemory: allocatable.Memory,
			UsedCPU:           usedCPU,
			UsedMemory:        usedMemory,
		}

		addNodeToResourceSnapshot(scaleUpSnapshot, nodeName, nodeGroupID, nodeCapacity)
		if !isUpcomingNode {
			addNodeToResourceSnapshot(scaleDownSnapshot, nodeName, nodeGroupID, nodeCapacity)
		}
	}

	klog.V(4).Infof("[NodeToGroup] Built scale-up mapping for %d nodes, scale-down mapping for %d nodes",
		len(scaleUpSnapshot.NodeToGroup), len(scaleDownSnapshot.NodeToGroup))

	m.fillResourceSnapshot(scaleUpSnapshot)
	m.fillResourceSnapshot(scaleDownSnapshot)

	return scaleUpSnapshot, scaleDownSnapshot, nil
}

func newSetResourceSnapshot() *SetResourceSnapshot {
	return &SetResourceSnapshot{
		Timestamp:      time.Now(),
		Resources:      make(map[string]*NodeGroupSetResourceInfo),
		NodeCapacities: make(map[string]*NodeResourceCapacity),
		NodeToGroup:    make(map[string]string),
	}
}

func addNodeToResourceSnapshot(snapshot *SetResourceSnapshot, nodeName, nodeGroupID string,
	nodeCapacity *NodeResourceCapacity,
) {
	if nodeGroupID != "" {
		snapshot.NodeToGroup[nodeName] = nodeGroupID
	}
	snapshot.NodeCapacities[nodeName] = &NodeResourceCapacity{
		NodeName:          nodeCapacity.NodeName,
		AllocatableCPU:    nodeCapacity.AllocatableCPU,
		AllocatableMemory: nodeCapacity.AllocatableMemory,
		UsedCPU:           nodeCapacity.UsedCPU,
		UsedMemory:        nodeCapacity.UsedMemory,
	}
}

func (m *NodeGroupSetManager) fillResourceSnapshot(snapshot *SetResourceSnapshot) {
	// 获取所有 Set 名称(不包括 GlobalSetName)
	allSets := m.GetAllSets()

	// 计算每个 Set 的资源总量(基于快照数据)
	// 传入 NodeToGroup 映射,避免重复调用 CloudProvider API
	for _, setName := range allSets {
		resourceInfo, err := m.calculateSetResourcesFromSnapshot(setName,
			snapshot.NodeCapacities, snapshot.NodeToGroup)
		if err != nil {
			klog.V(3).Infof("Failed to calculate resources for set %s: %v", setName, err)
			continue
		}
		snapshot.Resources[setName] = resourceInfo
	}

	// 计算 GlobalSet 的资源（包含所有节点）
	// 这样 GlobalSet 节点也可以使用 SimulateNodeRemoval 进行统一的缩容检查
	globalInfo := &NodeGroupSetResourceInfo{
		SetName: GlobalSetName,
	}
	for _, nodeCap := range snapshot.NodeCapacities {
		globalInfo.TotalCPU += nodeCap.AllocatableCPU
		globalInfo.TotalMemory += nodeCap.AllocatableMemory
		globalInfo.AvailableCPU += (nodeCap.AllocatableCPU - nodeCap.UsedCPU)
		globalInfo.AvailableMemory += (nodeCap.AllocatableMemory - nodeCap.UsedMemory)
	}
	// 计算使用率
	if globalInfo.TotalCPU > 0 {
		globalInfo.CPUUsageRatio = 1.0 - float64(globalInfo.AvailableCPU)/float64(globalInfo.TotalCPU)
	}
	if globalInfo.TotalMemory > 0 {
		globalInfo.MemUsageRatio = 1.0 - float64(globalInfo.AvailableMemory)/float64(globalInfo.TotalMemory)
	}
	snapshot.Resources[GlobalSetName] = globalInfo
	klog.V(4).Infof("[GlobalSet] Total CPU: %dm, Total Memory: %dMi, Available CPU: %dm, Available Memory: %dMi",
		globalInfo.TotalCPU, globalInfo.TotalMemory/(1024*1024),
		globalInfo.AvailableCPU, globalInfo.AvailableMemory/(1024*1024))
}

// calculateSetResourcesFromSnapshot 从快照计算 Set 资源
// 内部方法,由 TakeResourceSnapshot 调用
func (m *NodeGroupSetManager) calculateSetResourcesFromSnapshot(
	setName string,
	nodeCapacities map[string]*NodeResourceCapacity,
	nodeToGroup map[string]string,
) (*NodeGroupSetResourceInfo, error) {
	nodeGroups := m.GetNodeGroupsInSet(setName)
	if len(nodeGroups) == 0 {
		return nil, fmt.Errorf("no node groups in set %s", setName)
	}

	// 构建 NodeGroup ID 集合
	nodeGroupIDSet := make(map[string]bool)
	for _, ng := range nodeGroups {
		nodeGroupIDSet[ng.Id()] = true
	}

	info := &NodeGroupSetResourceInfo{
		SetName: setName,
	}

	// 遍历所有节点(真实 + upcoming),过滤出属于该 Set 的节点
	for nodeName, nodeCap := range nodeCapacities {
		// 使用预构建的映射查询 NodeGroup ID
		nodeGroupID, exists := nodeToGroup[nodeName]
		if !exists {
			continue
		}

		// 检查该节点是否属于当前 Set
		if !nodeGroupIDSet[nodeGroupID] {
			continue
		}

		// 累加资源
		info.TotalCPU += nodeCap.AllocatableCPU
		info.TotalMemory += nodeCap.AllocatableMemory
		info.AvailableCPU += (nodeCap.AllocatableCPU - nodeCap.UsedCPU)
		info.AvailableMemory += (nodeCap.AllocatableMemory - nodeCap.UsedMemory)
	}

	// 计算使用率
	if info.TotalCPU > 0 {
		info.CPUUsageRatio = 1.0 - float64(info.AvailableCPU)/float64(info.TotalCPU)
	}
	if info.TotalMemory > 0 {
		info.MemUsageRatio = 1.0 - float64(info.AvailableMemory)/float64(info.TotalMemory)
	}

	return info, nil
}

// SimulateNodeRemoval 模拟移除节点后的资源状态(增量计算)
// 用于 Scale-Down 时判断是否可以安全删除节点
func (m *NodeGroupSetManager) SimulateNodeRemoval(
	setName string,
	snapshot *SetResourceSnapshot,
	nodeName string,
	podsToReschedule []*apiv1.Pod,
) (*NodeGroupSetResourceInfo, error) {
	// 获取当前 Set 的资源状态
	currentInfo, exists := snapshot.Resources[setName]
	if !exists {
		return nil, fmt.Errorf("set %s not found in snapshot", setName)
	}

	// 获取要移除节点的资源容量
	nodeCap, exists := snapshot.NodeCapacities[nodeName]
	if !exists {
		return nil, fmt.Errorf("node %s not found in snapshot", nodeName)
	}

	// 计算需要重调度的 Pods 的资源总量
	podsResourceCPU := int64(0)
	podsResourceMemory := int64(0)
	if len(podsToReschedule) > 0 {
		for _, pod := range podsToReschedule {
			for _, container := range pod.Spec.Containers {
				if cpu := container.Resources.Requests.Cpu(); cpu != nil {
					podsResourceCPU += cpu.MilliValue()
				}
				if mem := container.Resources.Requests.Memory(); mem != nil {
					podsResourceMemory += mem.Value()
				}
			}
		}
		klog.V(5).Infof("[SimulateRemoval] Node %s has %d pods to reschedule, requiring CPU=%dm, Memory=%dMi",
			nodeName, len(podsToReschedule), podsResourceCPU, podsResourceMemory/(1024*1024))
	}

	// 创建新的资源信息(深拷贝)
	newInfo := &NodeGroupSetResourceInfo{
		SetName:         currentInfo.SetName,
		TotalCPU:        currentInfo.TotalCPU - nodeCap.AllocatableCPU,
		TotalMemory:     currentInfo.TotalMemory - nodeCap.AllocatableMemory,
		AvailableCPU:    currentInfo.AvailableCPU - (nodeCap.AllocatableCPU - nodeCap.UsedCPU) - podsResourceCPU,
		AvailableMemory: currentInfo.AvailableMemory - (nodeCap.AllocatableMemory - nodeCap.UsedMemory) - podsResourceMemory,
	}

	// 重新计算使用率
	if newInfo.TotalCPU > 0 {
		newInfo.CPUUsageRatio = 1.0 - float64(newInfo.AvailableCPU)/float64(newInfo.TotalCPU)
	} else {
		newInfo.CPUUsageRatio = 1.0
	}
	if newInfo.TotalMemory > 0 {
		newInfo.MemUsageRatio = 1.0 - float64(newInfo.AvailableMemory)/float64(newInfo.TotalMemory)
	} else {
		newInfo.MemUsageRatio = 1.0
	}

	klog.V(5).Infof("[SimulateRemoval] Set %s after removing %s: TotalCPU=%dm, AvailCPU=%dm, Usage=%.2f%%",
		setName, nodeName, newInfo.TotalCPU, newInfo.AvailableCPU, newInfo.CPUUsageRatio*100)

	return newInfo, nil
}

// DeepCopy 深拷贝资源快照
func (s *SetResourceSnapshot) DeepCopy() *SetResourceSnapshot {
	if s == nil {
		return nil
	}

	snapshot := &SetResourceSnapshot{
		Timestamp:      s.Timestamp,
		Resources:      make(map[string]*NodeGroupSetResourceInfo),
		NodeCapacities: make(map[string]*NodeResourceCapacity),
		NodeToGroup:    make(map[string]string),
	}

	// 深拷贝 Resources
	for setName, info := range s.Resources {
		snapshot.Resources[setName] = &NodeGroupSetResourceInfo{
			SetName:          info.SetName,
			TotalCPU:         info.TotalCPU,
			AvailableCPU:     info.AvailableCPU,
			TotalMemory:      info.TotalMemory,
			AvailableMemory:  info.AvailableMemory,
			CPUUsageRatio:    info.CPUUsageRatio,
			MemUsageRatio:    info.MemUsageRatio,
			TriggeredScaleUp: info.TriggeredScaleUp,
		}
	}

	// 深拷贝 NodeCapacities
	for nodeName, cap := range s.NodeCapacities {
		snapshot.NodeCapacities[nodeName] = &NodeResourceCapacity{
			NodeName:          cap.NodeName,
			AllocatableCPU:    cap.AllocatableCPU,
			AllocatableMemory: cap.AllocatableMemory,
			UsedCPU:           cap.UsedCPU,
			UsedMemory:        cap.UsedMemory,
		}
	}

	// 拷贝 NodeToGroup
	for nodeName, groupID := range s.NodeToGroup {
		snapshot.NodeToGroup[nodeName] = groupID
	}

	return snapshot
}

// UpdateSetResourceInfo 更新快照中指定 Set 的资源信息
func (s *SetResourceSnapshot) UpdateSetResourceInfo(setName string, newInfo *NodeGroupSetResourceInfo) {
	if s == nil || newInfo == nil {
		return
	}

	s.Resources[setName] = &NodeGroupSetResourceInfo{
		SetName:          newInfo.SetName,
		TotalCPU:         newInfo.TotalCPU,
		AvailableCPU:     newInfo.AvailableCPU,
		TotalMemory:      newInfo.TotalMemory,
		AvailableMemory:  newInfo.AvailableMemory,
		CPUUsageRatio:    newInfo.CPUUsageRatio,
		MemUsageRatio:    newInfo.MemUsageRatio,
		TriggeredScaleUp: newInfo.TriggeredScaleUp,
	}
}

// RemoveNodeFromSnapshot 从快照中移除节点的容量信息
func (s *SetResourceSnapshot) RemoveNodeFromSnapshot(nodeName string) {
	if s == nil {
		return
	}

	delete(s.NodeCapacities, nodeName)
	delete(s.NodeToGroup, nodeName)
}
