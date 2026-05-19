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
	"fmt"
	"strings"
	"sync"
	"testing"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func makePool(name, ns string, start, end, segLen uint32) *networkextensionv1.HostNetPortPool {
	return &networkextensionv1.HostNetPortPool{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: networkextensionv1.HostNetPortPoolSpec{
			StartPort:     start,
			EndPort:       end,
			SegmentLength: segLen,
		},
	}
}

func poolKey(ns, name string) string {
	return fmt.Sprintf("%s/%s", ns, name)
}

func TestAddPoolAndRemovePool(t *testing.T) {
	c := NewHostNetPortPoolCache()
	pool := makePool("test-pool", "ns1", 30000, 30100, 10)
	c.AddPool(pool)

	entry := c.GetPoolEntry("ns1/test-pool")
	if entry == nil {
		t.Fatal("expected pool entry to exist")
	}
	if entry.StartPort != 30000 || entry.EndPort != 30100 || entry.SegmentLength != 10 {
		t.Fatalf("unexpected pool entry: %+v", entry)
	}

	c.RemovePool("ns1/test-pool")
	if c.GetPoolEntry("ns1/test-pool") != nil {
		t.Fatal("expected pool entry to be removed")
	}
}

func TestAllocateContiguousSingleSegment(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30100, 10))

	start, end, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if start != 30000 || end != 30009 {
		t.Fatalf("expected 30000-30009, got %d-%d", start, end)
	}
}

func TestAllocateContiguousMultiSegment(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30100, 10))

	start, end, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if start != 30000 || end != 30029 {
		t.Fatalf("expected 30000-30029, got %d-%d", start, end)
	}
}

func TestAllocateContiguousFirstFit(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30100, 10))

	// Allocate first segment
	_, _, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Next allocation should get the second segment (first-fit)
	start, end, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-2", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if start != 30010 || end != 30019 {
		t.Fatalf("expected 30010-30019, got %d-%d", start, end)
	}
}

func TestAllocContigFailNoTotal(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30020, 10)) // only 2 segments

	// Allocate both segments
	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 1)
	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-2", 1)

	// Third allocation should fail
	_, _, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-3", 1)
	if err == nil {
		t.Fatal("expected error for insufficient segments")
	}
	if !strings.Contains(err.Error(), "totalFree=0") {
		t.Fatalf("expected diagnostic info in error, got: %v", err)
	}
}

func TestAllocContigFailFragment(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30050, 10)) // 5 segments

	// Allocate segments 0, 2, 4 to create fragmentation
	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 1) // seg 0
	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-2", 1) // seg 1
	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-3", 1) // seg 2

	// Release seg 1 to create a gap
	c.Release("ns/p", "node-1", 30010, 30019)

	// Try to allocate 2 contiguous: seg 3 and seg 4 are free and contiguous
	start, end, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-4", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if start != 30030 || end != 30049 {
		t.Fatalf("expected 30030-30049, got %d-%d", start, end)
	}
}

func TestAllocContigFailNoFit(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30050, 10)) // 5 segments

	// Allocate segments 0, 2, 4 leaving gaps of size 1
	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 1) // seg 0
	c.Lock()
	entry := c.pools["ns/p"]
	alloc := entry.NodeAllocators["node-1"]
	alloc.Segments[2].Allocated = true
	alloc.Segments[2].PodKey = "ns/pod-x"
	alloc.AllocatedCount++
	alloc.Segments[4].Allocated = true
	alloc.Segments[4].PodKey = "ns/pod-y"
	alloc.AllocatedCount++
	c.Unlock()

	// Try to allocate 2 contiguous - should fail because max contiguous is 1
	_, _, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-5", 2)
	if err == nil {
		t.Fatal("expected error for fragmentation")
	}
	if !strings.Contains(err.Error(), "maxContiguousFree=1") {
		t.Fatalf("expected maxContiguousFree diagnostic, got: %v", err)
	}
}

func TestAllocateContiguousPoolNotFound(t *testing.T) {
	c := NewHostNetPortPoolCache()
	_, _, err := c.AllocateContiguous("ns/nonexistent", "node-1", "ns/pod-1", 1)
	if err == nil {
		t.Fatal("expected error for nonexistent pool")
	}
}

func TestReleaseAndReuse(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30020, 10))

	start1, end1, _ := c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 1)
	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-2", 1)

	c.Release("ns/p", "node-1", start1, end1)

	// Should reuse the released segment (first-fit)
	start, end, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-3", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if start != start1 || end != end1 {
		t.Fatalf("expected reuse of %d-%d, got %d-%d", start1, end1, start, end)
	}
}

func TestReleaseByPodKey(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30100, 10))

	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 3) // 3 segments

	affected := c.ReleaseByPodKey("ns/pod-1")

	if len(affected) != 1 {
		t.Fatalf("expected 1 affected pool, got %d", len(affected))
	}
	if affected[0].PoolName != "p" || affected[0].PoolNamespace != "ns" {
		t.Fatalf("expected pool ns/p, got %s/%s", affected[0].PoolNamespace, affected[0].PoolName)
	}

	// All 3 segments should be free; allocating 3 should start from 30000
	start, _, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-2", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if start != 30000 {
		t.Fatalf("expected 30000, got %d", start)
	}
}

func TestUpdatePoolExpand(t *testing.T) {
	c := NewHostNetPortPoolCache()
	pool := makePool("p", "ns", 30000, 30050, 10)
	c.AddPool(pool)

	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 1)

	// Expand the pool
	expandedPool := makePool("p", "ns", 30000, 30100, 10)
	conflicts := c.UpdatePool(expandedPool)
	if len(conflicts) > 0 {
		t.Fatalf("unexpected conflicts on expand: %+v", conflicts)
	}

	entry := c.GetPoolEntry("ns/p")
	if entry.EndPort != 30100 {
		t.Fatalf("expected end port 30100, got %d", entry.EndPort)
	}

	// Old allocation should be preserved
	c.Lock()
	alloc := entry.NodeAllocators["node-1"]
	if !alloc.Segments[0].Allocated || alloc.Segments[0].PodKey != "ns/pod-1" {
		t.Fatal("expected first segment to remain allocated")
	}
	if alloc.TotalCount != 10 {
		t.Fatalf("expected 10 total segments, got %d", alloc.TotalCount)
	}
	c.Unlock()
}

func TestUpdatePoolShrinkConflict(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30100, 10))

	// Allocate in the 30050-30059 range
	c.Lock()
	alloc := c.getOrCreateNodeAllocator(c.pools["ns/p"], "node-1")
	alloc.Segments[5].Allocated = true
	alloc.Segments[5].PodKey = "ns/pod-x"
	alloc.AllocatedCount++
	c.Unlock()

	// Shrink to 30050 → should conflict
	shrunkPool := makePool("p", "ns", 30000, 30050, 10)
	conflicts := c.UpdatePool(shrunkPool)
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0].PodKey != "ns/pod-x" {
		t.Fatalf("expected conflict pod ns/pod-x, got %s", conflicts[0].PodKey)
	}

	// Pool should NOT have been updated
	entry := c.GetPoolEntry("ns/p")
	if entry.EndPort != 30100 {
		t.Fatalf("expected pool endPort to remain 30100, got %d", entry.EndPort)
	}
}

func TestUpdatePoolShrinkNoConflict(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30100, 10))

	// Allocate in the first segment only
	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 1)

	// Shrink to 30050 → no conflict (allocated segment is at 30000-30009)
	shrunkPool := makePool("p", "ns", 30000, 30050, 10)
	conflicts := c.UpdatePool(shrunkPool)
	if len(conflicts) > 0 {
		t.Fatalf("unexpected conflicts: %+v", conflicts)
	}

	entry := c.GetPoolEntry("ns/p")
	if entry.EndPort != 30050 {
		t.Fatalf("expected endPort 30050, got %d", entry.EndPort)
	}
}

func TestCleanupNode(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30100, 10))

	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 1)
	_, _, _ = c.AllocateContiguous("ns/p", "node-2", "ns/pod-2", 1)

	c.CleanupNode("node-1")

	c.Lock()
	entry := c.pools["ns/p"]
	if _, ok := entry.NodeAllocators["node-1"]; ok {
		t.Fatal("expected node-1 allocator to be removed")
	}
	if _, ok := entry.NodeAllocators["node-2"]; !ok {
		t.Fatal("expected node-2 allocator to still exist")
	}
	c.Unlock()
}

func TestRebuildFromPod(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30100, 10))

	c.RebuildFromPod("ns/p", "node-1", "ns/pod-1", 30000, 30009)
	c.RebuildFromPod("ns/p", "node-1", "ns/pod-2", 30020, 30029)

	// Should not allocate the rebuilt segments
	start, _, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-3", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if start != 30010 {
		t.Fatalf("expected 30010, got %d (rebuilt segments should be occupied)", start)
	}
}

func TestLazyCreationOnAllocate(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30050, 10))

	// No explicit node creation — allocate should lazily create
	start, end, err := c.AllocateContiguous("ns/p", "new-node", "ns/pod-1", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if start != 30000 || end != 30009 {
		t.Fatalf("expected 30000-30009, got %d-%d", start, end)
	}
}

func TestGetNodeAllocations(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30100, 10))

	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 2)
	_, _, _ = c.AllocateContiguous("ns/p", "node-2", "ns/pod-2", 1)

	allocs := c.GetNodeAllocations("ns/p")
	if len(allocs) != 2 {
		t.Fatalf("expected 2 node allocations, got %d", len(allocs))
	}

	allocMap := make(map[string]*networkextensionv1.NodeHostNetPortPoolStatus)
	for _, a := range allocs {
		allocMap[a.NodeName] = a
	}

	if allocMap["node-1"].AllocatedCount != 2 || allocMap["node-1"].TotalSegments != 10 {
		t.Fatalf("unexpected node-1 allocation: %+v", allocMap["node-1"])
	}
	if allocMap["node-2"].AllocatedCount != 1 || allocMap["node-2"].TotalSegments != 10 {
		t.Fatalf("unexpected node-2 allocation: %+v", allocMap["node-2"])
	}
}

func TestGetNodeAllocNoPool(t *testing.T) {
	c := NewHostNetPortPoolCache()
	allocs := c.GetNodeAllocations("ns/nonexistent")
	if allocs != nil {
		t.Fatalf("expected nil for nonexistent pool, got %+v", allocs)
	}
}

func TestGetAllocatedSegments(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30050, 10))

	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 2)

	segs := c.GetAllocatedSegments()
	if len(segs) != 2 {
		t.Fatalf("expected 2 allocated segments, got %d", len(segs))
	}
	for _, s := range segs {
		if s.PodKey != "ns/pod-1" {
			t.Fatalf("expected pod key ns/pod-1, got %s", s.PodKey)
		}
	}
}

func TestConcurrentAllocateRelease(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 31000, 10)) // 100 segments

	var wg sync.WaitGroup
	allocErrors := make(chan error, 50)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			podKey := fmt.Sprintf("ns/pod-%d", idx)
			_, _, err := c.AllocateContiguous("ns/p", "node-1", podKey, 1)
			if err != nil {
				allocErrors <- err
			}
		}(i)
	}

	wg.Wait()
	close(allocErrors)

	for err := range allocErrors {
		t.Fatalf("concurrent allocation error: %v", err)
	}

	// Verify 50 segments allocated
	c.Lock()
	alloc := c.pools["ns/p"].NodeAllocators["node-1"]
	if alloc.AllocatedCount != 50 {
		t.Fatalf("expected 50 allocated, got %d", alloc.AllocatedCount)
	}
	c.Unlock()

	// Concurrent release
	var wg2 sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg2.Add(1)
		go func(idx int) {
			defer wg2.Done()
			podKey := fmt.Sprintf("ns/pod-%d", idx)
			c.ReleaseByPodKey(podKey)
		}(i)
	}
	wg2.Wait()

	c.Lock()
	if alloc.AllocatedCount != 0 {
		t.Fatalf("expected 0 allocated after release, got %d", alloc.AllocatedCount)
	}
	c.Unlock()
}

func TestReleaseNonexistentPool(t *testing.T) {
	c := NewHostNetPortPoolCache()
	// Should not panic
	c.Release("ns/nonexistent", "node-1", 30000, 30009)
}

func TestReleaseNonexistentNode(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30100, 10))
	// Should not panic
	c.Release("ns/p", "nonexistent-node", 30000, 30009)
}

func TestReleaseByPodKeyNoMatch(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30100, 10))
	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 1)
	// Release a different pod key — should not affect existing allocation
	affected := c.ReleaseByPodKey("ns/pod-nonexistent")

	if len(affected) != 0 {
		t.Fatalf("expected 0 affected pools for non-matching pod, got %d", len(affected))
	}

	c.Lock()
	alloc := c.pools["ns/p"].NodeAllocators["node-1"]
	if alloc.AllocatedCount != 1 {
		t.Fatalf("expected 1 allocated, got %d", alloc.AllocatedCount)
	}
	c.Unlock()
}

func TestDifferentNodesGetSamePorts(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30050, 10))

	start1, end1, _ := c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 1)
	start2, end2, _ := c.AllocateContiguous("ns/p", "node-2", "ns/pod-2", 1)

	if start1 != start2 || end1 != end2 {
		t.Fatalf("different nodes should get same port range, got %d-%d vs %d-%d",
			start1, end1, start2, end2)
	}
}

func TestUpdatePoolNewPool(t *testing.T) {
	c := NewHostNetPortPoolCache()
	pool := makePool("p", "ns", 30000, 30100, 10)
	conflicts := c.UpdatePool(pool)
	if len(conflicts) > 0 {
		t.Fatalf("unexpected conflicts for new pool: %+v", conflicts)
	}
	entry := c.GetPoolEntry("ns/p")
	if entry == nil {
		t.Fatal("expected pool entry to be created by UpdatePool")
	}
}

func TestRebuildFromPodNonexistentPool(t *testing.T) {
	c := NewHostNetPortPoolCache()
	// Should not panic
	c.RebuildFromPod("ns/nonexistent", "node-1", "ns/pod-1", 30000, 30009)
}

func TestCleanupNodeMultiplePools(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p1", "ns", 30000, 30050, 10))
	c.AddPool(makePool("p2", "ns", 40000, 40050, 10))

	_, _, _ = c.AllocateContiguous("ns/p1", "node-1", "ns/pod-1", 1)
	_, _, _ = c.AllocateContiguous("ns/p2", "node-1", "ns/pod-2", 1)

	c.CleanupNode("node-1")

	c.Lock()
	if _, ok := c.pools["ns/p1"].NodeAllocators["node-1"]; ok {
		t.Fatal("expected node-1 removed from p1")
	}
	if _, ok := c.pools["ns/p2"].NodeAllocators["node-1"]; ok {
		t.Fatal("expected node-1 removed from p2")
	}
	c.Unlock()
}

// --- Edge case tests ---

func TestAddPoolOverwrite(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30100, 10))
	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 1)

	c.AddPool(makePool("p", "ns", 30000, 30050, 5))

	entry := c.GetPoolEntry("ns/p")
	if entry.EndPort != 30050 || entry.SegmentLength != 5 {
		t.Fatalf("expected overwritten pool config, got endPort=%d segLen=%d",
			entry.EndPort, entry.SegmentLength)
	}

	c.Lock()
	if len(entry.NodeAllocators) != 0 {
		t.Fatal("overwrite should reset allocators")
	}
	c.Unlock()
}

func TestAllocateContiguousZeroSegments(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30050, 10))

	_, _, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 0)
	if err == nil {
		t.Fatal("allocate 0 segments should return error, got nil")
	}
}

func TestAllocContigExceedAvail(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30030, 10)) // 3 segments

	_, _, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 5)
	if err == nil {
		t.Fatal("expected error when requesting more segments than exist")
	}
	if !strings.Contains(err.Error(), "maxContiguousFree=3") {
		t.Fatalf("expected maxContiguousFree=3 in error, got: %v", err)
	}
}

func TestPoolNotDivisibleBySegmentLength(t *testing.T) {
	c := NewHostNetPortPoolCache()
	// 30000-30035 with segLen=10 → 3 full segments, 5 ports truncated
	c.AddPool(makePool("p", "ns", 30000, 30035, 10))

	entry := c.GetPoolEntry("ns/p")
	c.Lock()
	alloc := c.getOrCreateNodeAllocator(entry, "node-1")
	if alloc.TotalCount != 3 {
		t.Fatalf("expected 3 segments (truncated), got %d", alloc.TotalCount)
	}
	c.Unlock()

	// All 3 segments should be allocatable
	_, _, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 4th should fail
	_, _, err = c.AllocateContiguous("ns/p", "node-1", "ns/pod-2", 1)
	if err == nil {
		t.Fatal("expected error: all segments exhausted")
	}
}

func TestPoolWithSegmentLengthOne(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30005, 1))

	for i := 0; i < 5; i++ {
		start, end, err := c.AllocateContiguous("ns/p", "node-1",
			fmt.Sprintf("ns/pod-%d", i), 1)
		if err != nil {
			t.Fatalf("allocate %d failed: %v", i, err)
		}
		if start != 30000+i || end != 30000+i {
			t.Fatalf("expected %d-%d, got %d-%d", 30000+i, 30000+i, start, end)
		}
	}

	_, _, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-overflow", 1)
	if err == nil {
		t.Fatal("expected error: pool exhausted")
	}
}

func TestPoolSegLenExceedsRange(t *testing.T) {
	c := NewHostNetPortPoolCache()
	// range=10 but segLen=20 → 0 segments
	c.AddPool(makePool("p", "ns", 30000, 30010, 20))

	_, _, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 1)
	if err == nil {
		t.Fatal("expected error: 0 segments in pool")
	}
}

func TestReleasePartialMultiSegment(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30050, 10))

	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 3) // 30000-30029

	// Release only the middle segment
	c.Release("ns/p", "node-1", 30010, 30019)

	c.Lock()
	alloc := c.pools["ns/p"].NodeAllocators["node-1"]
	if alloc.AllocatedCount != 2 {
		t.Fatalf("expected 2 allocated after partial release, got %d", alloc.AllocatedCount)
	}
	if !alloc.Segments[0].Allocated || !alloc.Segments[2].Allocated {
		t.Fatal("expected segments 0 and 2 to remain allocated")
	}
	if alloc.Segments[1].Allocated {
		t.Fatal("expected segment 1 to be released")
	}
	c.Unlock()
}

func TestReleaseNonAlignedRange(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30050, 10))
	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 2) // 30000-30019

	// Release with a range that doesn't align to segment boundaries
	c.Release("ns/p", "node-1", 30005, 30025)

	c.Lock()
	alloc := c.pools["ns/p"].NodeAllocators["node-1"]
	// Segment 0 (30000-30009): StartPort=30000 >= 30005? No → not released
	// Segment 1 (30010-30019): StartPort=30010 >= 30005 and EndPort=30019 <= 30025 → released
	if alloc.AllocatedCount != 1 {
		t.Fatalf("expected 1 remaining, got %d", alloc.AllocatedCount)
	}
	if !alloc.Segments[0].Allocated {
		t.Fatal("segment 0 should remain allocated (partial overlap)")
	}
	c.Unlock()
}

func TestRebuildFromPodDuplicate(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30050, 10))

	c.RebuildFromPod("ns/p", "node-1", "ns/pod-1", 30000, 30009)
	c.RebuildFromPod("ns/p", "node-1", "ns/pod-1", 30000, 30009)

	c.Lock()
	alloc := c.pools["ns/p"].NodeAllocators["node-1"]
	if alloc.AllocatedCount != 1 {
		t.Fatalf("duplicate rebuild should be idempotent, got allocatedCount=%d", alloc.AllocatedCount)
	}
	c.Unlock()
}

func TestRebuildFromPodMultipleSegments(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30050, 10))

	c.RebuildFromPod("ns/p", "node-1", "ns/pod-1", 30000, 30029)

	c.Lock()
	alloc := c.pools["ns/p"].NodeAllocators["node-1"]
	if alloc.AllocatedCount != 3 {
		t.Fatalf("expected 3 segments rebuilt, got %d", alloc.AllocatedCount)
	}
	for i := 0; i < 3; i++ {
		if !alloc.Segments[i].Allocated || alloc.Segments[i].PodKey != "ns/pod-1" {
			t.Fatalf("segment %d should be allocated to ns/pod-1", i)
		}
	}
	c.Unlock()
}

func TestUpdatePoolSegmentLengthChange(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30100, 10))
	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 1) // seg 0: 30000-30009

	// Change segmentLength from 10 to 5 → 20 segments
	newPool := makePool("p", "ns", 30000, 30100, 5)
	conflicts := c.UpdatePool(newPool)
	if len(conflicts) > 0 {
		t.Fatalf("unexpected conflicts: %+v", conflicts)
	}

	c.Lock()
	alloc := c.pools["ns/p"].NodeAllocators["node-1"]
	if alloc.TotalCount != 20 {
		t.Fatalf("expected 20 segments with segLen=5, got %d", alloc.TotalCount)
	}
	if alloc.AllocatedCount != 1 {
		t.Fatalf("expected 1 allocated (30000-30004 preserved), got %d", alloc.AllocatedCount)
	}
	c.Unlock()
}

func TestUpdatePoolStartPortChange(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30050, 10))
	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 1) // 30000-30009

	// Shift start port → conflict because 30000-30009 is outside new range
	newPool := makePool("p", "ns", 30010, 30050, 10)
	conflicts := c.UpdatePool(newPool)
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0].StartPort != 30000 {
		t.Fatalf("expected conflict at 30000, got %d", conflicts[0].StartPort)
	}
}

func TestReleaseByPodKeyMultiplePools(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p1", "ns", 30000, 30050, 10))
	c.AddPool(makePool("p2", "ns", 40000, 40050, 10))

	_, _, _ = c.AllocateContiguous("ns/p1", "node-1", "ns/pod-shared", 1)
	_, _, _ = c.AllocateContiguous("ns/p2", "node-2", "ns/pod-shared", 1)

	affected := c.ReleaseByPodKey("ns/pod-shared")

	if len(affected) != 2 {
		t.Fatalf("expected 2 affected pools, got %d", len(affected))
	}

	segs := c.GetAllocatedSegments()
	for _, seg := range segs {
		if seg.PodKey == "ns/pod-shared" {
			t.Fatalf("expected all segments for ns/pod-shared to be released, found in pool %s", seg.PoolKey)
		}
	}
}

func TestGetAllocSegsMultiPoolNode(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p1", "ns", 30000, 30050, 10))
	c.AddPool(makePool("p2", "ns", 40000, 40050, 10))

	_, _, _ = c.AllocateContiguous("ns/p1", "node-1", "ns/pod-1", 2)
	_, _, _ = c.AllocateContiguous("ns/p1", "node-2", "ns/pod-2", 1)
	_, _, _ = c.AllocateContiguous("ns/p2", "node-1", "ns/pod-3", 3)

	segs := c.GetAllocatedSegments()
	if len(segs) != 6 {
		t.Fatalf("expected 6 total allocated segments (2+1+3), got %d", len(segs))
	}

	poolCounts := make(map[string]int)
	nodeCounts := make(map[string]int)
	for _, s := range segs {
		poolCounts[s.PoolKey]++
		nodeCounts[s.NodeName]++
	}
	if poolCounts["ns/p1"] != 3 || poolCounts["ns/p2"] != 3 {
		t.Fatalf("unexpected pool counts: %+v", poolCounts)
	}
	if nodeCounts["node-1"] != 5 || nodeCounts["node-2"] != 1 {
		t.Fatalf("unexpected node counts: %+v", nodeCounts)
	}
}

func TestRemovePoolWithAllocations(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30050, 10))
	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 2)

	c.RemovePool("ns/p")

	if c.GetPoolEntry("ns/p") != nil {
		t.Fatal("pool should be removed even with allocations")
	}
	segs := c.GetAllocatedSegments()
	if len(segs) != 0 {
		t.Fatalf("expected 0 segments after pool removal, got %d", len(segs))
	}
}

func TestRemoveNonexistentPool(t *testing.T) {
	c := NewHostNetPortPoolCache()
	// Should not panic
	c.RemovePool("ns/nonexistent")
}

func TestCleanupNonexistentNode(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30050, 10))
	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 1)

	// Should not panic or affect existing allocations, returns empty slice
	affected := c.CleanupNode("nonexistent-node")
	if len(affected) != 0 {
		t.Fatalf("expected 0 affected pools, got %d", len(affected))
	}

	segs := c.GetAllocatedSegments()
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
}

func TestCleanupNodeReturnsAffectedPools(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("pool-a", "ns-1", 30000, 30050, 10))
	c.AddPool(makePool("pool-b", "ns-2", 40000, 40050, 10))
	_, _, _ = c.AllocateContiguous("ns-1/pool-a", "node-1", "ns-1/pod-1", 1)
	_, _, _ = c.AllocateContiguous("ns-2/pool-b", "node-1", "ns-2/pod-2", 1)

	affected := c.CleanupNode("node-1")
	if len(affected) != 2 {
		t.Fatalf("expected 2 affected pools, got %d", len(affected))
	}

	found := map[string]bool{}
	for _, a := range affected {
		found[a.PoolNamespace+"/"+a.PoolName] = true
	}
	if !found["ns-1/pool-a"] || !found["ns-2/pool-b"] {
		t.Fatalf("unexpected affected pools: %+v", affected)
	}
}

func TestAllocExhaustReleaseReuse(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30030, 10)) // 3 segments

	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 1) // seg 0
	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-2", 1) // seg 1
	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-3", 1) // seg 2

	// Pool fully exhausted
	_, _, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-4", 1)
	if err == nil {
		t.Fatal("expected error: pool exhausted")
	}

	// Release middle segment
	c.ReleaseByPodKey("ns/pod-2")

	// Should be able to allocate again
	start, end, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-4", 1)
	if err != nil {
		t.Fatalf("unexpected error after release: %v", err)
	}
	if start != 30010 || end != 30019 {
		t.Fatalf("expected reused segment 30010-30019, got %d-%d", start, end)
	}
}

func TestConcAllocDiffNodesExhaust(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30100, 10)) // 10 segments

	var wg sync.WaitGroup
	errCh := make(chan error, 30)

	for node := 0; node < 3; node++ {
		for pod := 0; pod < 10; pod++ {
			wg.Add(1)
			go func(n, p int) {
				defer wg.Done()
				nodeName := fmt.Sprintf("node-%d", n)
				podKey := fmt.Sprintf("ns/pod-%d-%d", n, p)
				_, _, err := c.AllocateContiguous("ns/p", nodeName, podKey, 1)
				if err != nil {
					errCh <- fmt.Errorf("node=%s pod=%s: %v", nodeName, podKey, err)
				}
			}(node, pod)
		}
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("concurrent allocation error: %v", err)
	}

	segs := c.GetAllocatedSegments()
	if len(segs) != 30 {
		t.Fatalf("expected 30 segments (10 per node * 3 nodes), got %d", len(segs))
	}
}

func TestGetPoolEntryNonexistent(t *testing.T) {
	c := NewHostNetPortPoolCache()
	if c.GetPoolEntry("ns/nonexistent") != nil {
		t.Fatal("expected nil for nonexistent pool")
	}
}

func TestGetPoolRange(t *testing.T) {
	c := NewHostNetPortPoolCache()

	// Non-existent pool
	_, _, exists := c.GetPoolRange("ns/no-pool")
	if exists {
		t.Fatal("expected false for non-existent pool")
	}

	// Add pool and verify range
	c.AddPool(makePool("p", "ns", 30000, 30100, 10))
	start, end, exists := c.GetPoolRange("ns/p")
	if !exists {
		t.Fatal("expected true after AddPool")
	}
	if start != 30000 || end != 30100 {
		t.Errorf("expected 30000-30100, got %d-%d", start, end)
	}

	// After UpdatePool with shrink (no conflicts), range should update
	shrunk := makePool("p", "ns", 30000, 30050, 10)
	conflicts := c.UpdatePool(shrunk)
	if len(conflicts) > 0 {
		t.Fatalf("unexpected conflicts: %+v", conflicts)
	}
	start, end, exists = c.GetPoolRange("ns/p")
	if !exists || start != 30000 || end != 30050 {
		t.Errorf("expected 30000-30050, got %d-%d (exists=%v)", start, end, exists)
	}

	// After RemovePool, should not exist
	c.RemovePool("ns/p")
	_, _, exists = c.GetPoolRange("ns/p")
	if exists {
		t.Fatal("expected false after RemovePool")
	}
}

func TestPoolRangeOnShrinkConflict(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30020, 10))
	c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 1) // 30000-30009
	c.AllocateContiguous("ns/p", "node-1", "ns/pod-2", 1) // 30010-30019

	// Shrink to exclude pod-2's segment → conflict
	shrunk := makePool("p", "ns", 30000, 30010, 10)
	conflicts := c.UpdatePool(shrunk)
	if len(conflicts) == 0 {
		t.Fatal("expected conflicts")
	}

	// Range should still be the OLD range (30000-30020) since shrink was rejected
	start, end, _ := c.GetPoolRange("ns/p")
	if start != 30000 || end != 30020 {
		t.Errorf("expected original range 30000-30020 during conflict, got %d-%d", start, end)
	}
}

func TestAllocFullExpandRealloc(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30020, 10)) // 2 segments

	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 1)
	_, _, _ = c.AllocateContiguous("ns/p", "node-1", "ns/pod-2", 1)

	// Pool full
	_, _, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-3", 1)
	if err == nil {
		t.Fatal("expected error: pool full")
	}

	// Expand pool
	expanded := makePool("p", "ns", 30000, 30050, 10)
	conflicts := c.UpdatePool(expanded)
	if len(conflicts) > 0 {
		t.Fatalf("unexpected conflicts: %+v", conflicts)
	}

	// Should now succeed
	start, end, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-3", 1)
	if err != nil {
		t.Fatalf("expected success after expand: %v", err)
	}
	if start != 30020 || end != 30029 {
		t.Fatalf("expected 30020-30029, got %d-%d", start, end)
	}
}

func TestAllocateContiguousIdempotent(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30100, 10)) // 10 segments

	s1, e1, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 2)
	if err != nil {
		t.Fatalf("first allocation failed: %v", err)
	}
	if s1 != 30000 || e1 != 30019 {
		t.Fatalf("expected 30000-30019, got %d-%d", s1, e1)
	}

	s2, e2, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-1", 2)
	if err != nil {
		t.Fatalf("idempotent allocation failed: %v", err)
	}
	if s2 != s1 || e2 != e1 {
		t.Fatalf("expected idempotent result %d-%d, got %d-%d", s1, e1, s2, e2)
	}

	allocs := c.GetNodeAllocations("ns/p")
	if len(allocs) != 1 || allocs[0].AllocatedCount != 2 {
		t.Fatalf("expected 2 allocated segments, got %+v", allocs)
	}
}

func TestAllocContigIdempotentConc(t *testing.T) {
	c := NewHostNetPortPoolCache()
	c.AddPool(makePool("p", "ns", 30000, 30100, 10))

	const goroutines = 10
	results := make([][2]int, goroutines)
	errs := make([]error, goroutines)
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			s, e, err := c.AllocateContiguous("ns/p", "node-1", "ns/pod-race", 1)
			results[idx] = [2]int{s, e}
			errs[idx] = err
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d failed: %v", i, err)
		}
	}

	expected := results[0]
	for i, r := range results {
		if r != expected {
			t.Fatalf("goroutine %d got %v, expected %v (idempotent violation)", i, r, expected)
		}
	}

	allocs := c.GetNodeAllocations("ns/p")
	if len(allocs) != 1 || allocs[0].AllocatedCount != 1 {
		t.Fatalf("expected exactly 1 allocated segment, got %+v", allocs)
	}
}
