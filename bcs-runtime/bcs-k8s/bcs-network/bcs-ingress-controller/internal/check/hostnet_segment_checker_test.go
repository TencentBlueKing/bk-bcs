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

package check

import (
	"testing"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/hostnetportpoolcache"
)

func newCheckScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = k8scorev1.AddToScheme(s)
	_ = networkextensionv1.AddToScheme(s)
	return s
}

func TestSegChecker_ReleaseOrphanedPod(t *testing.T) {
	scheme := newCheckScheme()
	cli := k8sfake.NewFakeClientWithScheme(scheme)

	cache := hostnetportpoolcache.NewHostNetPortPoolCache()
	pool := &networkextensionv1.HostNetPortPool{
		ObjectMeta: metav1.ObjectMeta{Name: "pool-1", Namespace: "default"},
		Spec: networkextensionv1.HostNetPortPoolSpec{
			StartPort: 10000, EndPort: 10010, SegmentLength: 5,
		},
	}
	cache.AddPool(pool)
	cache.AllocateContiguous("default/pool-1", "node-1", "default/orphan-pod", 1)

	segs := cache.GetAllocatedSegments()
	if len(segs) != 1 {
		t.Fatalf("expected 1 allocated segment, got %d", len(segs))
	}

	checker := NewHostNetSegmentChecker(cli, cache)
	checker.Run()

	segs = cache.GetAllocatedSegments()
	if len(segs) != 0 {
		t.Errorf("expected 0 after leak check, got %d", len(segs))
	}
}

func TestSegChecker_KeepRunningPod(t *testing.T) {
	scheme := newCheckScheme()
	pod := &k8scorev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "running-pod", Namespace: "default"},
		Status:     k8scorev1.PodStatus{Phase: k8scorev1.PodRunning},
	}
	cli := k8sfake.NewFakeClientWithScheme(scheme, pod)

	cache := hostnetportpoolcache.NewHostNetPortPoolCache()
	pool := &networkextensionv1.HostNetPortPool{
		ObjectMeta: metav1.ObjectMeta{Name: "pool-1", Namespace: "default"},
		Spec: networkextensionv1.HostNetPortPoolSpec{
			StartPort: 10000, EndPort: 10010, SegmentLength: 5,
		},
	}
	cache.AddPool(pool)
	cache.AllocateContiguous("default/pool-1", "node-1", "default/running-pod", 1)

	checker := NewHostNetSegmentChecker(cli, cache)
	checker.Run()

	segs := cache.GetAllocatedSegments()
	if len(segs) != 1 {
		t.Errorf("expected 1 (running pod should be kept), got %d", len(segs))
	}
}

func TestSegChecker_ReleaseTerminatedPod(t *testing.T) {
	scheme := newCheckScheme()
	pod := &k8scorev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "done-pod", Namespace: "default"},
		Status:     k8scorev1.PodStatus{Phase: k8scorev1.PodSucceeded},
	}
	cli := k8sfake.NewFakeClientWithScheme(scheme, pod)

	cache := hostnetportpoolcache.NewHostNetPortPoolCache()
	pool := &networkextensionv1.HostNetPortPool{
		ObjectMeta: metav1.ObjectMeta{Name: "pool-1", Namespace: "default"},
		Spec: networkextensionv1.HostNetPortPoolSpec{
			StartPort: 10000, EndPort: 10010, SegmentLength: 5,
		},
	}
	cache.AddPool(pool)
	cache.AllocateContiguous("default/pool-1", "node-1", "default/done-pod", 1)

	checker := NewHostNetSegmentChecker(cli, cache)
	checker.Run()

	segs := cache.GetAllocatedSegments()
	if len(segs) != 0 {
		t.Errorf("expected 0 after releasing terminated pod, got %d", len(segs))
	}
}

func TestSegChecker_InvalidPodKey(t *testing.T) {
	scheme := newCheckScheme()
	cli := k8sfake.NewFakeClientWithScheme(scheme)

	cache := hostnetportpoolcache.NewHostNetPortPoolCache()
	pool := &networkextensionv1.HostNetPortPool{
		ObjectMeta: metav1.ObjectMeta{Name: "pool-1", Namespace: "default"},
		Spec: networkextensionv1.HostNetPortPoolSpec{
			StartPort: 10000, EndPort: 10010, SegmentLength: 5,
		},
	}
	cache.AddPool(pool)
	cache.AllocateContiguous("default/pool-1", "node-1", "bad-key", 1)

	checker := NewHostNetSegmentChecker(cli, cache)
	checker.Run()

	segs := cache.GetAllocatedSegments()
	if len(segs) != 1 {
		t.Errorf("invalid key should be skipped, expected 1, got %d", len(segs))
	}
}

func TestSegChecker_NoAllocations(t *testing.T) {
	scheme := newCheckScheme()
	cli := k8sfake.NewFakeClientWithScheme(scheme)

	cache := hostnetportpoolcache.NewHostNetPortPoolCache()
	checker := NewHostNetSegmentChecker(cli, cache)
	checker.Run()
}

// --- Edge case tests ---

func TestSegChecker_MultiSegmentLeak(t *testing.T) {
	scheme := newCheckScheme()
	cli := k8sfake.NewFakeClientWithScheme(scheme)

	cache := hostnetportpoolcache.NewHostNetPortPoolCache()
	pool := &networkextensionv1.HostNetPortPool{
		ObjectMeta: metav1.ObjectMeta{Name: "pool-1", Namespace: "default"},
		Spec: networkextensionv1.HostNetPortPoolSpec{
			StartPort: 10000, EndPort: 10030, SegmentLength: 5,
		},
	}
	cache.AddPool(pool)
	cache.AllocateContiguous("default/pool-1", "node-1", "default/multi-seg-pod", 3)

	segs := cache.GetAllocatedSegments()
	if len(segs) != 3 {
		t.Fatalf("expected 3 allocated segments, got %d", len(segs))
	}

	checker := NewHostNetSegmentChecker(cli, cache)
	checker.Run()

	segs = cache.GetAllocatedSegments()
	if len(segs) != 0 {
		t.Errorf("expected all 3 segments released for orphaned multi-seg pod, got %d", len(segs))
	}
}

func TestSegChecker_PendingPodKept(t *testing.T) {
	scheme := newCheckScheme()
	pod := &k8scorev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pending-pod", Namespace: "default"},
		Status:     k8scorev1.PodStatus{Phase: k8scorev1.PodPending},
	}
	cli := k8sfake.NewFakeClientWithScheme(scheme, pod)

	cache := hostnetportpoolcache.NewHostNetPortPoolCache()
	pool := &networkextensionv1.HostNetPortPool{
		ObjectMeta: metav1.ObjectMeta{Name: "pool-1", Namespace: "default"},
		Spec: networkextensionv1.HostNetPortPoolSpec{
			StartPort: 10000, EndPort: 10010, SegmentLength: 5,
		},
	}
	cache.AddPool(pool)
	cache.AllocateContiguous("default/pool-1", "node-1", "default/pending-pod", 1)

	checker := NewHostNetSegmentChecker(cli, cache)
	checker.Run()

	segs := cache.GetAllocatedSegments()
	if len(segs) != 1 {
		t.Errorf("pending pod should be kept, expected 1, got %d", len(segs))
	}
}

func TestSegChecker_MixedPodStates(t *testing.T) {
	scheme := newCheckScheme()
	runningPod := &k8scorev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "running-pod", Namespace: "default"},
		Status:     k8scorev1.PodStatus{Phase: k8scorev1.PodRunning},
	}
	failedPod := &k8scorev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "failed-pod", Namespace: "default"},
		Status:     k8scorev1.PodStatus{Phase: k8scorev1.PodFailed},
	}
	// orphan-pod does not exist in the fake client
	cli := k8sfake.NewFakeClientWithScheme(scheme, runningPod, failedPod)

	cache := hostnetportpoolcache.NewHostNetPortPoolCache()
	pool := &networkextensionv1.HostNetPortPool{
		ObjectMeta: metav1.ObjectMeta{Name: "pool-1", Namespace: "default"},
		Spec: networkextensionv1.HostNetPortPoolSpec{
			StartPort: 10000, EndPort: 10030, SegmentLength: 5,
		},
	}
	cache.AddPool(pool)
	cache.AllocateContiguous("default/pool-1", "node-1", "default/running-pod", 1)
	cache.AllocateContiguous("default/pool-1", "node-1", "default/failed-pod", 1)
	cache.AllocateContiguous("default/pool-1", "node-1", "default/orphan-pod", 1)

	checker := NewHostNetSegmentChecker(cli, cache)
	checker.Run()

	segs := cache.GetAllocatedSegments()
	if len(segs) != 1 {
		t.Errorf("expected 1 segment (running pod only), got %d", len(segs))
	}
	if segs[0].PodKey != "default/running-pod" {
		t.Errorf("expected running-pod, got %s", segs[0].PodKey)
	}
}

func TestSegChecker_MultiPoolOrphan(t *testing.T) {
	scheme := newCheckScheme()
	cli := k8sfake.NewFakeClientWithScheme(scheme)

	cache := hostnetportpoolcache.NewHostNetPortPoolCache()
	pool1 := &networkextensionv1.HostNetPortPool{
		ObjectMeta: metav1.ObjectMeta{Name: "pool-1", Namespace: "default"},
		Spec: networkextensionv1.HostNetPortPoolSpec{
			StartPort: 10000, EndPort: 10010, SegmentLength: 5,
		},
	}
	pool2 := &networkextensionv1.HostNetPortPool{
		ObjectMeta: metav1.ObjectMeta{Name: "pool-2", Namespace: "default"},
		Spec: networkextensionv1.HostNetPortPoolSpec{
			StartPort: 20000, EndPort: 20010, SegmentLength: 5,
		},
	}
	cache.AddPool(pool1)
	cache.AddPool(pool2)
	cache.AllocateContiguous("default/pool-1", "node-1", "default/orphan-pod", 1)
	cache.AllocateContiguous("default/pool-2", "node-1", "default/orphan-pod", 1)

	checker := NewHostNetSegmentChecker(cli, cache)
	checker.Run()

	segs := cache.GetAllocatedSegments()
	if len(segs) != 0 {
		t.Errorf("expected all segments released across pools, got %d", len(segs))
	}
}

func TestSegChecker_DifferentNodesOrphan(t *testing.T) {
	scheme := newCheckScheme()
	cli := k8sfake.NewFakeClientWithScheme(scheme)

	cache := hostnetportpoolcache.NewHostNetPortPoolCache()
	pool := &networkextensionv1.HostNetPortPool{
		ObjectMeta: metav1.ObjectMeta{Name: "pool-1", Namespace: "default"},
		Spec: networkextensionv1.HostNetPortPoolSpec{
			StartPort: 10000, EndPort: 10020, SegmentLength: 5,
		},
	}
	cache.AddPool(pool)
	cache.AllocateContiguous("default/pool-1", "node-1", "default/orphan-pod", 1)
	cache.AllocateContiguous("default/pool-1", "node-2", "default/orphan-pod", 1)

	checker := NewHostNetSegmentChecker(cli, cache)
	checker.Run()

	segs := cache.GetAllocatedSegments()
	if len(segs) != 0 {
		t.Errorf("expected both nodes' segments released, got %d", len(segs))
	}
}

func TestSegChecker_PodKeyWithSlash(t *testing.T) {
	scheme := newCheckScheme()
	cli := k8sfake.NewFakeClientWithScheme(scheme)

	cache := hostnetportpoolcache.NewHostNetPortPoolCache()
	pool := &networkextensionv1.HostNetPortPool{
		ObjectMeta: metav1.ObjectMeta{Name: "pool-1", Namespace: "default"},
		Spec: networkextensionv1.HostNetPortPoolSpec{
			StartPort: 10000, EndPort: 10010, SegmentLength: 5,
		},
	}
	cache.AddPool(pool)
	// Unusual podKey with extra slashes
	cache.AllocateContiguous("default/pool-1", "node-1", "ns/sub/pod-name", 1)

	checker := NewHostNetSegmentChecker(cli, cache)
	checker.Run()

	segs := cache.GetAllocatedSegments()
	// SplitN with 2 → ns="ns", name="sub/pod-name", pod lookup will fail → release
	if len(segs) != 0 {
		t.Errorf("expected segment released for pod with unusual key, got %d", len(segs))
	}
}

func TestSegChecker_RunMultipleTimes(t *testing.T) {
	scheme := newCheckScheme()
	pod := &k8scorev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "stable-pod", Namespace: "default"},
		Status:     k8scorev1.PodStatus{Phase: k8scorev1.PodRunning},
	}
	cli := k8sfake.NewFakeClientWithScheme(scheme, pod)

	cache := hostnetportpoolcache.NewHostNetPortPoolCache()
	pool := &networkextensionv1.HostNetPortPool{
		ObjectMeta: metav1.ObjectMeta{Name: "pool-1", Namespace: "default"},
		Spec: networkextensionv1.HostNetPortPoolSpec{
			StartPort: 10000, EndPort: 10010, SegmentLength: 5,
		},
	}
	cache.AddPool(pool)
	cache.AllocateContiguous("default/pool-1", "node-1", "default/stable-pod", 1)

	checker := NewHostNetSegmentChecker(cli, cache)

	for i := 0; i < 5; i++ {
		checker.Run()
	}

	segs := cache.GetAllocatedSegments()
	if len(segs) != 1 {
		t.Errorf("running checker multiple times should be idempotent, got %d segments", len(segs))
	}
}
