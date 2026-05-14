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

package hostnetportpoolcache

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// ErrPoolNotInCache indicates the requested pool has not been synced into the cache yet.
var ErrPoolNotInCache = errors.New("pool not found in cache")

// HostNetPortPoolSegment represents a single port segment in a node allocator.
type HostNetPortPoolSegment struct {
	StartPort int
	EndPort   int
	Allocated bool
	PodKey    string
}

// NodeSegmentAllocator manages segment allocation on a specific node.
type NodeSegmentAllocator struct {
	NodeName       string
	Segments       []*HostNetPortPoolSegment
	TotalCount     int
	AllocatedCount int
}

// HostNetPortPoolEntry holds pool configuration and per-node allocators.
type HostNetPortPoolEntry struct {
	PoolKey        string
	StartPort      int
	EndPort        int
	SegmentLength  int
	NodeAllocators map[string]*NodeSegmentAllocator
}

// PoolChangedEvent carries the pool key that had an allocation change.
type PoolChangedEvent struct {
	PoolNamespace string
	PoolName      string
}

// HostNetPortPoolCache is the top-level thread-safe cache for all HostNetPortPools.
type HostNetPortPoolCache struct {
	sync.Mutex
	pools      map[string]*HostNetPortPoolEntry
	notifyCh   chan PoolChangedEvent
	syncedOnce sync.Once
	syncedCh   chan struct{}
}

// NewHostNetPortPoolCache creates a new empty cache.
func NewHostNetPortPoolCache() *HostNetPortPoolCache {
	return &HostNetPortPoolCache{
		pools:    make(map[string]*HostNetPortPoolEntry),
		notifyCh: make(chan PoolChangedEvent, 64),
		syncedCh: make(chan struct{}),
	}
}

// MarkSynced signals that the initial cache rebuild is complete.
// Safe to call multiple times; only the first call takes effect.
func (c *HostNetPortPoolCache) MarkSynced() {
	c.syncedOnce.Do(func() { close(c.syncedCh) })
}

// IsSynced returns true after MarkSynced has been called.
func (c *HostNetPortPoolCache) IsSynced() bool {
	select {
	case <-c.syncedCh:
		return true
	default:
		return false
	}
}

// NotifyCh returns the channel that receives pool change notifications.
// Pool controller should watch this channel to trigger status sync.
func (c *HostNetPortPoolCache) NotifyCh() <-chan PoolChangedEvent {
	return c.notifyCh
}

// notifyPoolChanged sends a non-blocking notification for the given pool key.
// Must NOT be called with c.Mutex held to avoid blocking on a full channel.
func (c *HostNetPortPoolCache) notifyPoolChanged(poolKey string) {
	parts := strings.SplitN(poolKey, "/", 2)
	if len(parts) != 2 {
		return
	}
	select {
	case c.notifyCh <- PoolChangedEvent{PoolNamespace: parts[0], PoolName: parts[1]}:
	default:
	}
}

// AddPool registers a new HostNetPortPool into the cache.
func (c *HostNetPortPoolCache) AddPool(pool *networkextensionv1.HostNetPortPool) {
	c.Lock()
	defer c.Unlock()

	key := fmt.Sprintf("%s/%s", pool.Namespace, pool.Name)
	c.pools[key] = &HostNetPortPoolEntry{
		PoolKey:        key,
		StartPort:      int(pool.Spec.StartPort),
		EndPort:        int(pool.Spec.EndPort),
		SegmentLength:  int(pool.Spec.SegmentLength),
		NodeAllocators: make(map[string]*NodeSegmentAllocator),
	}
}

// RemovePool removes a HostNetPortPool from the cache.
func (c *HostNetPortPoolCache) RemovePool(poolKey string) {
	c.Lock()
	defer c.Unlock()
	delete(c.pools, poolKey)
}

// GetPoolRange returns the cached port range for a pool.
// Returns (startPort, endPort, true) if the pool exists, or (0, 0, false) otherwise.
func (c *HostNetPortPoolCache) GetPoolRange(poolKey string) (int, int, bool) {
	c.Lock()
	defer c.Unlock()
	entry, ok := c.pools[poolKey]
	if !ok {
		return 0, 0, false
	}
	return entry.StartPort, entry.EndPort, true
}

// UpdatePool updates a pool's configuration and detects shrink conflicts.
// Returns conflicting segments if the new range is smaller and allocated segments exist outside it.
func (c *HostNetPortPoolCache) UpdatePool(pool *networkextensionv1.HostNetPortPool) []ConflictSegment {
	c.Lock()
	defer c.Unlock()

	key := fmt.Sprintf("%s/%s", pool.Namespace, pool.Name)
	entry, ok := c.pools[key]
	if !ok {
		entry = &HostNetPortPoolEntry{
			PoolKey:        key,
			StartPort:      int(pool.Spec.StartPort),
			EndPort:        int(pool.Spec.EndPort),
			SegmentLength:  int(pool.Spec.SegmentLength),
			NodeAllocators: make(map[string]*NodeSegmentAllocator),
		}
		c.pools[key] = entry
		return nil
	}

	newStart := int(pool.Spec.StartPort)
	newEnd := int(pool.Spec.EndPort)
	newSegLen := int(pool.Spec.SegmentLength)

	var conflicts []ConflictSegment
	for nodeName, alloc := range entry.NodeAllocators {
		for _, seg := range alloc.Segments {
			if seg.Allocated && (seg.StartPort < newStart || seg.EndPort >= newEnd) {
				conflicts = append(conflicts, ConflictSegment{
					NodeName:  nodeName,
					StartPort: seg.StartPort,
					EndPort:   seg.EndPort,
					PodKey:    seg.PodKey,
				})
			}
		}
	}
	if len(conflicts) > 0 {
		return conflicts
	}

	entry.StartPort = newStart
	entry.EndPort = newEnd
	entry.SegmentLength = newSegLen

	for nodeName, alloc := range entry.NodeAllocators {
		entry.NodeAllocators[nodeName] = rebuildAllocator(nodeName, newStart, newEnd, newSegLen, alloc)
	}

	return nil
}

func rebuildAllocator(nodeName string, startPort, endPort, segLen int,
	old *NodeSegmentAllocator) *NodeSegmentAllocator {

	if segLen <= 0 {
		blog.Errorf("hostnetport cache: rebuildAllocator called with invalid segLen=%d for node %s (port %d-%d), "+
			"returning empty allocator", segLen, nodeName, startPort, endPort)
		return &NodeSegmentAllocator{NodeName: nodeName}
	}
	totalSegs := (endPort - startPort) / segLen
	newAlloc := &NodeSegmentAllocator{
		NodeName:   nodeName,
		Segments:   make([]*HostNetPortPoolSegment, totalSegs),
		TotalCount: totalSegs,
	}

	oldByStart := make(map[int]*HostNetPortPoolSegment)
	if old != nil {
		for _, seg := range old.Segments {
			oldByStart[seg.StartPort] = seg
		}
	}

	for i := 0; i < totalSegs; i++ {
		sp := startPort + i*segLen
		ep := sp + segLen - 1
		seg := &HostNetPortPoolSegment{StartPort: sp, EndPort: ep}
		if oldSeg, exists := oldByStart[sp]; exists && oldSeg.Allocated {
			seg.Allocated = true
			seg.PodKey = oldSeg.PodKey
			newAlloc.AllocatedCount++
		}
		newAlloc.Segments[i] = seg
	}
	return newAlloc
}

// AllocateContiguous allocates segmentsNeeded contiguous free segments on the given node.
// Returns (startPort, endPort, error). endPort is inclusive (last port of last segment).
//
// Idempotent: if podKey already owns segments in this pool (on any node), returns
// the existing allocation. This prevents double-allocation when the informer cache
// is stale and multiple Reconcile calls race for the same pod.
func (c *HostNetPortPoolCache) AllocateContiguous(poolKey, nodeName, podKey string,
	segmentsNeeded int) (int, int, error) {

	startPort, endPort, err := c.allocateContiguousLocked(poolKey, nodeName, podKey, segmentsNeeded)
	if err != nil {
		return 0, 0, err
	}
	c.notifyPoolChanged(poolKey)
	return startPort, endPort, nil
}

func (c *HostNetPortPoolCache) allocateContiguousLocked(poolKey, nodeName, podKey string,
	segmentsNeeded int) (int, int, error) {
	c.Lock()
	defer c.Unlock()

	if segmentsNeeded <= 0 {
		return 0, 0, fmt.Errorf("segmentsNeeded must be positive, got %d", segmentsNeeded)
	}

	entry, ok := c.pools[poolKey]
	if !ok {
		return 0, 0, fmt.Errorf("pool %s: %w", poolKey, ErrPoolNotInCache)
	}

	// Idempotent check: return existing allocation if podKey already owns segments.
	if sp, ep, found := c.findExistingAllocation(entry, podKey); found {
		return sp, ep, nil
	}

	alloc := c.getOrCreateNodeAllocator(entry, nodeName)

	bestStart := -1
	currentRun := 0
	maxContiguousFree := 0

	for i, seg := range alloc.Segments {
		if !seg.Allocated {
			currentRun++
			if currentRun > maxContiguousFree {
				maxContiguousFree = currentRun
			}
			if currentRun >= segmentsNeeded && bestStart == -1 {
				bestStart = i - segmentsNeeded + 1
			}
		} else {
			currentRun = 0
		}
	}

	if bestStart == -1 {
		totalFree := alloc.TotalCount - alloc.AllocatedCount
		return 0, 0, fmt.Errorf(
			"no %d contiguous free segments on node %s in pool %s "+
				"(totalFree=%d, maxContiguousFree=%d)",
			segmentsNeeded, nodeName, poolKey, totalFree, maxContiguousFree)
	}

	lastIdx := bestStart + segmentsNeeded - 1
	if bestStart < 0 || lastIdx >= len(alloc.Segments) {
		return 0, 0, fmt.Errorf(
			"segment index out of range: bestStart=%d, lastIdx=%d, len=%d",
			bestStart, lastIdx, len(alloc.Segments))
	}

	startPort := alloc.Segments[bestStart].StartPort
	endPort := alloc.Segments[lastIdx].EndPort
	for i := bestStart; i <= lastIdx; i++ {
		alloc.Segments[i].Allocated = true
		alloc.Segments[i].PodKey = podKey
	}
	alloc.AllocatedCount += segmentsNeeded

	return startPort, endPort, nil
}

// findExistingAllocation scans all node allocators for segments already owned by podKey.
// Returns (startPort, endPort, true) if found, or (0, 0, false) otherwise.
// Must be called with c.Mutex held.
func (c *HostNetPortPoolCache) findExistingAllocation(
	entry *HostNetPortPoolEntry, podKey string) (int, int, bool) {

	for _, alloc := range entry.NodeAllocators {
		minPort, maxPort := -1, -1
		for _, seg := range alloc.Segments {
			if seg.Allocated && seg.PodKey == podKey {
				if minPort == -1 || seg.StartPort < minPort {
					minPort = seg.StartPort
				}
				if seg.EndPort > maxPort {
					maxPort = seg.EndPort
				}
			}
		}
		if minPort != -1 {
			return minPort, maxPort, true
		}
	}
	return 0, 0, false
}

// Release frees a specific port range on a node within a pool.
func (c *HostNetPortPoolCache) Release(poolKey, nodeName string, startPort, endPort int) {
	c.Lock()
	released := false
	entry, ok := c.pools[poolKey]
	if ok {
		alloc, aok := entry.NodeAllocators[nodeName]
		if aok {
			for _, seg := range alloc.Segments {
				if seg.StartPort >= startPort && seg.EndPort <= endPort && seg.Allocated {
					seg.Allocated = false
					seg.PodKey = ""
					alloc.AllocatedCount--
					released = true
				}
			}
		}
	}
	c.Unlock()
	if released {
		c.notifyPoolChanged(poolKey)
	}
}

// ReleaseByPodKeyResult contains the pool identity affected by a release operation.
type ReleaseByPodKeyResult struct {
	PoolName      string
	PoolNamespace string
}

// ReleaseByPodKey releases all segments owned by a given pod key across all pools and nodes.
// Returns the list of affected pools so callers can update pool status immediately.
func (c *HostNetPortPoolCache) ReleaseByPodKey(podKey string) []ReleaseByPodKeyResult {
	c.Lock()
	seen := make(map[string]bool)
	var affected []ReleaseByPodKeyResult
	var changedPoolKeys []string
	for _, entry := range c.pools {
		released := false
		for _, alloc := range entry.NodeAllocators {
			for _, seg := range alloc.Segments {
				if seg.Allocated && seg.PodKey == podKey {
					seg.Allocated = false
					seg.PodKey = ""
					alloc.AllocatedCount--
					released = true
				}
			}
		}
		if released && !seen[entry.PoolKey] {
			seen[entry.PoolKey] = true
			changedPoolKeys = append(changedPoolKeys, entry.PoolKey)
			parts := strings.SplitN(entry.PoolKey, "/", 2)
			if len(parts) == 2 {
				affected = append(affected, ReleaseByPodKeyResult{
					PoolName:      parts[1],
					PoolNamespace: parts[0],
				})
			}
		}
	}
	c.Unlock()
	for _, pk := range changedPoolKeys {
		c.notifyPoolChanged(pk)
	}
	return affected
}

// GetNodeAllocations returns status for all nodes of a given pool.
func (c *HostNetPortPoolCache) GetNodeAllocations(
	poolKey string) []*networkextensionv1.NodeHostNetPortPoolStatus {
	c.Lock()
	defer c.Unlock()

	entry, ok := c.pools[poolKey]
	if !ok {
		return nil
	}

	result := make([]*networkextensionv1.NodeHostNetPortPoolStatus, 0, len(entry.NodeAllocators))
	for _, alloc := range entry.NodeAllocators {
		result = append(result, &networkextensionv1.NodeHostNetPortPoolStatus{
			NodeName:       alloc.NodeName,
			AllocatedCount: alloc.AllocatedCount,
			TotalSegments:  alloc.TotalCount,
		})
	}
	return result
}

// RebuildFromPod restores a segment allocation from an existing pod during cache rebuild.
func (c *HostNetPortPoolCache) RebuildFromPod(poolKey, nodeName, podKey string, startPort, endPort int) {
	c.Lock()
	defer c.Unlock()

	entry, ok := c.pools[poolKey]
	if !ok {
		return
	}

	alloc := c.getOrCreateNodeAllocator(entry, nodeName)
	for _, seg := range alloc.Segments {
		if seg.StartPort >= startPort && seg.EndPort <= endPort && !seg.Allocated {
			seg.Allocated = true
			seg.PodKey = podKey
			alloc.AllocatedCount++
		}
	}
}

// CleanupNodeResult contains the name and namespace of a pool affected by node cleanup.
type CleanupNodeResult struct {
	PoolName      string
	PoolNamespace string
}

// CleanupNode removes all allocator state for a node across every pool.
// Returns a list of affected pools so callers can clean up per-node metrics.
func (c *HostNetPortPoolCache) CleanupNode(nodeName string) []CleanupNodeResult {
	c.Lock()
	var affected []CleanupNodeResult
	var changedPoolKeys []string
	for _, entry := range c.pools {
		if _, ok := entry.NodeAllocators[nodeName]; ok {
			changedPoolKeys = append(changedPoolKeys, entry.PoolKey)
			parts := strings.SplitN(entry.PoolKey, "/", 2)
			if len(parts) == 2 {
				affected = append(affected, CleanupNodeResult{
					PoolName:      parts[1],
					PoolNamespace: parts[0],
				})
			}
			delete(entry.NodeAllocators, nodeName)
		}
	}
	c.Unlock()
	for _, pk := range changedPoolKeys {
		c.notifyPoolChanged(pk)
	}
	return affected
}

// CleanupStaleNodes removes NodeAllocators for nodes not in the given valid set.
// Returns the total number of stale node entries removed.
func (c *HostNetPortPoolCache) CleanupStaleNodes(validNodes map[string]struct{}) int {
	c.Lock()
	defer c.Unlock()

	removed := 0
	for _, entry := range c.pools {
		for nodeName := range entry.NodeAllocators {
			if _, ok := validNodes[nodeName]; !ok {
				delete(entry.NodeAllocators, nodeName)
				removed++
			}
		}
	}
	return removed
}

// GetAllocatedSegments returns all allocated segments across all pools and nodes.
// Used by the leak checker to iterate over allocated state.
func (c *HostNetPortPoolCache) GetAllocatedSegments() []AllocatedSegmentInfo {
	c.Lock()
	defer c.Unlock()

	var result []AllocatedSegmentInfo
	for _, entry := range c.pools {
		for nodeName, alloc := range entry.NodeAllocators {
			for _, seg := range alloc.Segments {
				if seg.Allocated {
					result = append(result, AllocatedSegmentInfo{
						PoolKey:   entry.PoolKey,
						NodeName:  nodeName,
						StartPort: seg.StartPort,
						EndPort:   seg.EndPort,
						PodKey:    seg.PodKey,
					})
				}
			}
		}
	}
	return result
}

// AllocatedSegmentInfo holds info about an allocated segment for external iteration.
type AllocatedSegmentInfo struct {
	PoolKey   string
	NodeName  string
	StartPort int
	EndPort   int
	PodKey    string
}

// IsPodAllocated checks whether a pod already owns allocated segments in any pool.
func (c *HostNetPortPoolCache) IsPodAllocated(podKey string) bool {
	c.Lock()
	defer c.Unlock()

	for _, entry := range c.pools {
		for _, alloc := range entry.NodeAllocators {
			for _, seg := range alloc.Segments {
				if seg.Allocated && seg.PodKey == podKey {
					return true
				}
			}
		}
	}
	return false
}

// GetPoolEntry returns the pool entry for a given key (nil if not found). Caller must hold no lock.
func (c *HostNetPortPoolCache) GetPoolEntry(poolKey string) *HostNetPortPoolEntry {
	c.Lock()
	defer c.Unlock()
	return c.pools[poolKey]
}

func (c *HostNetPortPoolCache) getOrCreateNodeAllocator(
	entry *HostNetPortPoolEntry, nodeName string) *NodeSegmentAllocator {

	alloc, ok := entry.NodeAllocators[nodeName]
	if ok {
		return alloc
	}

	if entry.SegmentLength <= 0 {
		blog.Errorf("hostnetport cache: getOrCreateNodeAllocator called with invalid SegmentLength=%d "+
			"for pool %s node %s, returning empty allocator", entry.SegmentLength, entry.PoolKey, nodeName)
		return &NodeSegmentAllocator{NodeName: nodeName}
	}
	totalSegs := (entry.EndPort - entry.StartPort) / entry.SegmentLength
	alloc = &NodeSegmentAllocator{
		NodeName:   nodeName,
		Segments:   make([]*HostNetPortPoolSegment, totalSegs),
		TotalCount: totalSegs,
	}
	for i := 0; i < totalSegs; i++ {
		sp := entry.StartPort + i*entry.SegmentLength
		ep := sp + entry.SegmentLength - 1
		alloc.Segments[i] = &HostNetPortPoolSegment{StartPort: sp, EndPort: ep}
	}
	entry.NodeAllocators[nodeName] = alloc
	return alloc
}
