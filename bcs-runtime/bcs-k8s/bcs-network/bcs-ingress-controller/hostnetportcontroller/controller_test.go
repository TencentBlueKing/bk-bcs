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

package hostnetportcontroller

import (
	"context"
	"encoding/json"
	"testing"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	k8sfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/hostnetportpoolcache"
)

func newTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = k8scorev1.AddToScheme(s)
	_ = networkextensionv1.AddToScheme(s)
	return s
}

func newPool(name, ns string, start, end, segLen uint32) *networkextensionv1.HostNetPortPool {
	return &networkextensionv1.HostNetPortPool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: networkextensionv1.HostNetPortPoolSpec{
			StartPort:     start,
			EndPort:       end,
			SegmentLength: segLen,
		},
	}
}

func newNode(name string) *k8scorev1.Node {
	return &k8scorev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
}

func newPod(name, ns, nodeName, poolAnnotation string) *k8scorev1.Pod {
	return &k8scorev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Annotations: map[string]string{
				constant.AnnotationForHostNetPortPool: poolAnnotation,
			},
		},
		Spec: k8scorev1.PodSpec{
			NodeName: nodeName,
		},
		Status: k8scorev1.PodStatus{
			Phase: k8scorev1.PodRunning,
		},
	}
}

func newPoolReconciler(objs ...runtime.Object) (*HostNetPortPoolReconciler, *hostnetportpoolcache.HostNetPortPoolCache) {
	scheme := newTestScheme()
	cli := k8sfake.NewFakeClientWithScheme(scheme, objs...)
	cache := hostnetportpoolcache.NewHostNetPortPoolCache()
	eventer := record.NewFakeRecorder(100)
	r := NewHostNetPortPoolReconciler(context.Background(), cli, cache, eventer)
	return r, cache
}

func newPodReconciler(objs ...runtime.Object) (*HostNetPodReconciler, *hostnetportpoolcache.HostNetPortPoolCache) {
	scheme := newTestScheme()
	cli := k8sfake.NewFakeClientWithScheme(scheme, objs...)
	cache := hostnetportpoolcache.NewHostNetPortPoolCache()
	cache.MarkSynced()
	eventer := record.NewFakeRecorder(100)
	r := NewHostNetPodReconciler(context.Background(), cli, cache, eventer)
	return r, cache
}

func newNodeReconciler(objs ...runtime.Object) (*HostNetNodeReconciler, *hostnetportpoolcache.HostNetPortPoolCache) {
	scheme := newTestScheme()
	cli := k8sfake.NewFakeClientWithScheme(scheme, objs...)
	cache := hostnetportpoolcache.NewHostNetPortPoolCache()
	cache.MarkSynced()
	r := NewHostNetNodeReconciler(context.Background(), cli, cache)
	return r, cache
}

// --- Cache sync gate tests ---

func TestReconcilePod_CacheNotSynced(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10010, 5)
	pod := newPod("test-pod", "default", "node-1", "test-pool")

	scheme := newTestScheme()
	cli := k8sfake.NewFakeClientWithScheme(scheme, pool, pod)
	cache := hostnetportpoolcache.NewHostNetPortPoolCache()
	// deliberately NOT calling cache.MarkSynced()
	eventer := record.NewFakeRecorder(100)
	r := NewHostNetPodReconciler(context.Background(), cli, cache, eventer)
	cache.AddPool(pool)

	result, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pod"},
	})
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	if result.RequeueAfter == 0 {
		t.Fatal("expected requeue when cache not synced")
	}
	if cache.IsPodAllocated("default/test-pod") {
		t.Fatal("pod should NOT be allocated before cache is synced")
	}
}

func TestReconcileNode_CacheNotSynced(t *testing.T) {
	scheme := newTestScheme()
	cli := k8sfake.NewFakeClientWithScheme(scheme)
	cache := hostnetportpoolcache.NewHostNetPortPoolCache()
	// deliberately NOT calling cache.MarkSynced()
	r := NewHostNetNodeReconciler(context.Background(), cli, cache)

	result, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "node-1"},
	})
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	if result.RequeueAfter == 0 {
		t.Fatal("expected requeue when cache not synced")
	}
}

// --- Pod tests ---

func TestReconcilePod_Allocate(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10010, 5)
	pod := newPod("test-pod", "default", "node-1", "test-pool")

	r, cache := newPodReconciler(pool, pod)
	cache.AddPool(pool)

	result, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pod"},
	})
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Errorf("expected no requeue, got %v", result.RequeueAfter)
	}

	updatedPod := &k8scorev1.Pod{}
	if err := r.client.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: "test-pod",
	}, updatedPod); err != nil {
		t.Fatalf("get pod failed: %v", err)
	}

	resultStr, ok := updatedPod.Annotations[constant.AnnotationForHostNetPortPoolBindingResult]
	if !ok {
		t.Fatal("expected binding result annotation")
	}

	var bindingResult hostnetportpoolcache.HostNetPortPoolBindingResult
	if err := json.Unmarshal([]byte(resultStr), &bindingResult); err != nil {
		t.Fatalf("parse binding result: %v", err)
	}

	if bindingResult.PoolName != "test-pool" {
		t.Errorf("expected pool name test-pool, got %s", bindingResult.PoolName)
	}
	if bindingResult.StartPort != 10000 {
		t.Errorf("expected start port 10000, got %d", bindingResult.StartPort)
	}
	if bindingResult.EndPort != 10004 {
		t.Errorf("expected end port 10004, got %d", bindingResult.EndPort)
	}

	status := updatedPod.Annotations[constant.AnnotationForHostNetPortPoolBindingStatus]
	if status != "Ready" {
		t.Errorf("expected status Ready, got %s", status)
	}
}

func TestReconcilePod_AlreadyAllocated(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10010, 5)
	pod := newPod("test-pod", "default", "node-1", "test-pool")

	r, cache := newPodReconciler(pool, pod)
	cache.AddPool(pool)
	// Pre-allocate in cache (simulates prior successful allocation)
	cache.AllocateContiguous("default/test-pool", "node-1", "default/test-pod", 1)

	result, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pod"},
	})
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Errorf("expected no requeue for already allocated")
	}
}

func TestReconcilePod_Unscheduled(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10010, 5)
	pod := newPod("test-pod", "default", "", "test-pool")

	r, cache := newPodReconciler(pool, pod)
	cache.AddPool(pool)

	result, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pod"},
	})
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Errorf("expected no requeue for unscheduled")
	}
}

func TestReconcilePod_TerminalPhase(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10010, 5)
	pod := newPod("test-pod", "default", "node-1", "test-pool")
	pod.Status.Phase = k8scorev1.PodSucceeded

	r, cache := newPodReconciler(pool, pod)
	cache.AddPool(pool)
	cache.AllocateContiguous("default/test-pool", "node-1", "default/test-pod", 1)

	_, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pod"},
	})
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	segs := cache.GetAllocatedSegments()
	if len(segs) != 0 {
		t.Errorf("expected segments released for terminal pod, got %d", len(segs))
	}
}

func TestReconcilePod_PoolNotFound(t *testing.T) {
	pod := newPod("test-pod", "default", "node-1", "nonexistent-pool")

	r, _ := newPodReconciler(pod)

	result, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pod"},
	})
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	updatedPod := &k8scorev1.Pod{}
	_ = r.client.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: "test-pod",
	}, updatedPod)

	status := updatedPod.Annotations[constant.AnnotationForHostNetPortPoolBindingStatus]
	if status != "Failed" {
		t.Errorf("expected status Failed for missing pool, got %s", status)
	}
	if result.RequeueAfter != 0 {
		t.Errorf("expected no requeue")
	}
}

func TestReconcilePod_AllocExhaust(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10005, 5)
	pod1 := newPod("pod-1", "default", "node-1", "test-pool")
	pod2 := newPod("pod-2", "default", "node-1", "test-pool")

	r, cache := newPodReconciler(pool, pod1, pod2)
	cache.AddPool(pool)

	_, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "pod-1"},
	})
	if err != nil {
		t.Fatalf("Reconcile pod-1 failed: %v", err)
	}

	_, err = r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "pod-2"},
	})
	if err != nil {
		t.Fatalf("Reconcile pod-2 failed: %v", err)
	}

	updatedPod2 := &k8scorev1.Pod{}
	_ = r.client.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: "pod-2",
	}, updatedPod2)

	status := updatedPod2.Annotations[constant.AnnotationForHostNetPortPoolBindingStatus]
	if status != "Failed" {
		t.Errorf("expected Failed for exhausted pool, got %s", status)
	}
}

func TestReconcilePod_DeleteNotFound(t *testing.T) {
	r, cache := newPodReconciler()

	pool := newPool("test-pool", "default", 10000, 10010, 5)
	cache.AddPool(pool)
	cache.AllocateContiguous("default/test-pool", "node-1", "default/ghost-pod", 1)

	_, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "ghost-pod"},
	})
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	segs := cache.GetAllocatedSegments()
	for _, seg := range segs {
		if seg.PodKey == "default/ghost-pod" {
			t.Error("expected ghost-pod segments to be released")
		}
	}
}

func TestReconcilePod_NoAnnotation(t *testing.T) {
	pod := &k8scorev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "plain-pod",
			Namespace: "default",
		},
		Spec: k8scorev1.PodSpec{
			NodeName: "node-1",
		},
		Status: k8scorev1.PodStatus{Phase: k8scorev1.PodRunning},
	}

	r, _ := newPodReconciler(pod)

	result, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "plain-pod"},
	})
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Errorf("expected no requeue for plain pod")
	}
}

func TestReconcilePod_Terminating(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10010, 5)
	now := metav1.Now()
	pod := newPod("test-pod", "default", "node-1", "test-pool")
	pod.DeletionTimestamp = &now
	pod.Finalizers = []string{"keep"}

	r, cache := newPodReconciler(pool, pod)
	cache.AddPool(pool)

	result, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pod"},
	})
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Errorf("expected no requeue for terminating pod")
	}
}

func TestReconcilePod_InvalidPortCount(t *testing.T) {
	tests := []struct {
		name      string
		portCount string
	}{
		{"non-numeric", "abc"},
		{"zero", "0"},
		{"negative", "-5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := newPool("test-pool", "default", 10000, 10020, 5)
			pod := newPod("test-pod", "default", "node-1", "test-pool")
			pod.Annotations[constant.AnnotationForHostNetPortPoolPortCount] = tt.portCount

			r, cache := newPodReconciler(pool, pod)
			cache.AddPool(pool)

			_, err := r.Reconcile(ctrl.Request{
				NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pod"},
			})
			if err != nil {
				t.Fatalf("Reconcile failed: %v", err)
			}

			updatedPod := &k8scorev1.Pod{}
			_ = r.client.Get(context.Background(), types.NamespacedName{
				Namespace: "default", Name: "test-pod",
			}, updatedPod)

			status := updatedPod.Annotations[constant.AnnotationForHostNetPortPoolBindingStatus]
			if status != "Failed" {
				t.Errorf("expected Failed status for invalid portcount %q, got %s", tt.portCount, status)
			}

			// Should NOT have a binding result
			if _, ok := updatedPod.Annotations[constant.AnnotationForHostNetPortPoolBindingResult]; ok {
				t.Errorf("invalid portcount should not produce a binding result")
			}
		})
	}
}

func TestReconcilePod_EmptyPortCount(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10020, 5)
	pod := newPod("test-pod", "default", "node-1", "test-pool")
	pod.Annotations[constant.AnnotationForHostNetPortPoolPortCount] = ""

	r, cache := newPodReconciler(pool, pod)
	cache.AddPool(pool)

	_, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pod"},
	})
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	updatedPod := &k8scorev1.Pod{}
	_ = r.client.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: "test-pod",
	}, updatedPod)

	// Empty portcount is also invalid (Sscanf will fail)
	status := updatedPod.Annotations[constant.AnnotationForHostNetPortPoolBindingStatus]
	if status != "Failed" {
		t.Errorf("expected Failed status for empty portcount, got %s", status)
	}
}

func TestReconcilePod_FailRelease(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10010, 5)
	pod := newPod("test-pod", "default", "node-1", "test-pool")
	pod.Status.Phase = k8scorev1.PodFailed

	r, cache := newPodReconciler(pool, pod)
	cache.AddPool(pool)
	cache.AllocateContiguous("default/test-pool", "node-1", "default/test-pod", 1)

	_, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pod"},
	})
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	segs := cache.GetAllocatedSegments()
	if len(segs) != 0 {
		t.Errorf("expected segments released for Failed pod, got %d", len(segs))
	}
}

func TestReconcilePod_IdempotAlloc(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10020, 5)
	pod := newPod("test-pod", "default", "node-1", "test-pool")

	r, cache := newPodReconciler(pool, pod)
	cache.AddPool(pool)

	_, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pod"},
	})
	if err != nil {
		t.Fatalf("first Reconcile failed: %v", err)
	}

	segsAfterFirst := cache.GetAllocatedSegments()

	_, err = r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pod"},
	})
	if err != nil {
		t.Fatalf("second Reconcile failed: %v", err)
	}

	segsAfterSecond := cache.GetAllocatedSegments()
	if len(segsAfterSecond) != len(segsAfterFirst) {
		t.Errorf("idempotent reconcile should not double-allocate: first=%d second=%d",
			len(segsAfterFirst), len(segsAfterSecond))
	}
}

func TestReconcilePod_MultiSegAlloc(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10015, 5)
	pod := newPod("test-pod", "default", "node-1", "test-pool")
	pod.Annotations[constant.AnnotationForHostNetPortPoolPortCount] = "8"

	r, cache := newPodReconciler(pool, pod)
	cache.AddPool(pool)

	_, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pod"},
	})
	if err != nil {
		t.Fatalf("Reconcile multi-segment failed: %v", err)
	}

	updatedPod := &k8scorev1.Pod{}
	_ = r.client.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: "test-pod",
	}, updatedPod)

	resultStr := updatedPod.Annotations[constant.AnnotationForHostNetPortPoolBindingResult]
	var result hostnetportpoolcache.HostNetPortPoolBindingResult
	if err := json.Unmarshal([]byte(resultStr), &result); err != nil {
		t.Fatalf("parse result: %v", err)
	}

	portRange := result.EndPort - result.StartPort + 1
	if portRange != 10 {
		t.Errorf("expected 10 ports (2 segments * 5), got %d", portRange)
	}
}

func TestReconcilePod_CrossNamespacePool(t *testing.T) {
	pool := newPool("shared-pool", "infra", 20000, 20010, 5)
	pod := newPod("test-pod", "app-ns", "node-1", "shared-pool")
	pod.Annotations[constant.AnnotationForHostNetPortPoolNamespace] = "infra"

	r, cache := newPodReconciler(pool, pod)
	cache.AddPool(pool)

	_, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "app-ns", Name: "test-pod"},
	})
	if err != nil {
		t.Fatalf("Reconcile cross-namespace failed: %v", err)
	}

	updatedPod := &k8scorev1.Pod{}
	_ = r.client.Get(context.Background(), types.NamespacedName{
		Namespace: "app-ns", Name: "test-pod",
	}, updatedPod)

	resultStr := updatedPod.Annotations[constant.AnnotationForHostNetPortPoolBindingResult]
	var result hostnetportpoolcache.HostNetPortPoolBindingResult
	if err := json.Unmarshal([]byte(resultStr), &result); err != nil {
		t.Fatalf("parse result: %v", err)
	}

	if result.PoolNamespace != "infra" {
		t.Errorf("expected pool namespace infra, got %s", result.PoolNamespace)
	}
}

func TestReconcilePod_PortCountExact(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10015, 5)
	pod := newPod("test-pod", "default", "node-1", "test-pool")
	pod.Annotations[constant.AnnotationForHostNetPortPoolPortCount] = "5"

	r, cache := newPodReconciler(pool, pod)
	cache.AddPool(pool)

	_, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pod"},
	})
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	updatedPod := &k8scorev1.Pod{}
	_ = r.client.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: "test-pod",
	}, updatedPod)

	var result hostnetportpoolcache.HostNetPortPoolBindingResult
	_ = json.Unmarshal([]byte(
		updatedPod.Annotations[constant.AnnotationForHostNetPortPoolBindingResult]), &result)

	portRange := result.EndPort - result.StartPort + 1
	if portRange != 5 {
		t.Errorf("expected exactly 5 ports, got %d", portRange)
	}
}

func TestReconcilePod_ExceedPool(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10010, 5)
	pod := newPod("test-pod", "default", "node-1", "test-pool")
	pod.Annotations[constant.AnnotationForHostNetPortPoolPortCount] = "100"

	r, cache := newPodReconciler(pool, pod)
	cache.AddPool(pool)

	_, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pod"},
	})
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	updatedPod := &k8scorev1.Pod{}
	_ = r.client.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: "test-pod",
	}, updatedPod)

	status := updatedPod.Annotations[constant.AnnotationForHostNetPortPoolBindingStatus]
	if status != "Failed" {
		t.Errorf("expected Failed for port count exceeding pool, got %s", status)
	}
}

func TestReconcilePod_MultiPodNodes(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10010, 5)
	pod1 := newPod("pod-1", "default", "node-1", "test-pool")
	pod2 := newPod("pod-2", "default", "node-2", "test-pool")

	r, cache := newPodReconciler(pool, pod1, pod2)
	cache.AddPool(pool)

	_, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "pod-1"},
	})
	if err != nil {
		t.Fatalf("Reconcile pod-1 failed: %v", err)
	}

	_, err = r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "pod-2"},
	})
	if err != nil {
		t.Fatalf("Reconcile pod-2 failed: %v", err)
	}

	get := func(name string) hostnetportpoolcache.HostNetPortPoolBindingResult {
		p := &k8scorev1.Pod{}
		_ = r.client.Get(context.Background(), types.NamespacedName{
			Namespace: "default", Name: name}, p)
		var res hostnetportpoolcache.HostNetPortPoolBindingResult
		_ = json.Unmarshal([]byte(p.Annotations[constant.AnnotationForHostNetPortPoolBindingResult]), &res)
		return res
	}

	r1 := get("pod-1")
	r2 := get("pod-2")

	if r1.StartPort != r2.StartPort || r1.EndPort != r2.EndPort {
		t.Errorf("pods on different nodes should get same ports: pod-1=%d-%d pod-2=%d-%d",
			r1.StartPort, r1.EndPort, r2.StartPort, r2.EndPort)
	}
	if r1.NodeName == r2.NodeName {
		t.Error("pods should be on different nodes")
	}
}

func TestReconcilePod_TermRecreate(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10020, 5)
	pod := newPod("test-pod", "default", "node-1", "test-pool")
	pod.Status.Phase = k8scorev1.PodSucceeded

	r, cache := newPodReconciler(pool, pod)
	cache.AddPool(pool)
	cache.AllocateContiguous("default/test-pool", "node-1", "default/test-pod", 1)

	_, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pod"},
	})
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	segs := cache.GetAllocatedSegments()
	if len(segs) != 0 {
		t.Fatalf("expected 0 segments after terminal release, got %d", len(segs))
	}

	newRunningPod := newPod("new-pod", "default", "node-1", "test-pool")
	_ = r.client.Create(context.Background(), newRunningPod)

	_, err = r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "new-pod"},
	})
	if err != nil {
		t.Fatalf("Reconcile new pod failed: %v", err)
	}

	segs = cache.GetAllocatedSegments()
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment for new pod, got %d", len(segs))
	}
}

// --- Node tests ---

func TestReconcileNodeDelete(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10010, 5)

	r, cache := newNodeReconciler(pool)
	cache.AddPool(pool)
	cache.AllocateContiguous("default/test-pool", "dying-node", "default/pod1", 1)

	_, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "dying-node"},
	})
	if err != nil {
		t.Fatalf("Reconcile node delete failed: %v", err)
	}

	segs := cache.GetAllocatedSegments()
	for _, seg := range segs {
		if seg.NodeName == "dying-node" {
			t.Error("expected dying-node allocations to be cleaned up")
		}
	}
}

// --- Pool tests ---

func TestReconcilePool_AddFinalizer(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10010, 5)

	r, _ := newPoolReconciler(pool)
	r.cacheSynced = true

	_, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pool"},
	})
	if err != nil {
		t.Fatalf("Reconcile pool failed: %v", err)
	}

	updatedPool := &networkextensionv1.HostNetPortPool{}
	if err := r.client.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: "test-pool",
	}, updatedPool); err != nil {
		t.Fatalf("get pool failed: %v", err)
	}

	if !controllerutil.ContainsFinalizer(updatedPool, constant.FinalizerNameHostNetPortPool) {
		t.Error("expected finalizer to be added")
	}
}

func TestReconcilePool_UpdateStatus(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10010, 5)
	pool.Finalizers = []string{constant.FinalizerNameHostNetPortPool}

	r, cache := newPoolReconciler(pool)
	cache.AddPool(pool)
	r.cacheSynced = true

	_, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pool"},
	})
	if err != nil {
		t.Fatalf("Reconcile pool failed: %v", err)
	}

	updatedPool := &networkextensionv1.HostNetPortPool{}
	if err := r.client.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: "test-pool",
	}, updatedPool); err != nil {
		t.Fatalf("get pool failed: %v", err)
	}

	if updatedPool.Status.Status != "Ready" {
		t.Errorf("expected pool status Ready, got %s", updatedPool.Status.Status)
	}
}

func TestReconcilePool_Delete(t *testing.T) {
	now := metav1.Now()
	pool := newPool("test-pool", "default", 10000, 10010, 5)
	pool.Finalizers = []string{constant.FinalizerNameHostNetPortPool}
	pool.DeletionTimestamp = &now

	r, cache := newPoolReconciler(pool)
	cache.AddPool(pool)
	r.cacheSynced = true

	_, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pool"},
	})
	if err != nil {
		t.Fatalf("Reconcile pool delete failed: %v", err)
	}
}

func TestReconcilePool_DelBlocked(t *testing.T) {
	now := metav1.Now()
	pool := newPool("test-pool", "default", 10000, 10010, 5)
	pool.Finalizers = []string{constant.FinalizerNameHostNetPortPool}
	pool.DeletionTimestamp = &now

	r, cache := newPoolReconciler(pool)
	cache.AddPool(pool)
	cache.AllocateContiguous("default/test-pool", "node-1", "default/still-running-pod", 1)
	r.cacheSynced = true

	result, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pool"},
	})
	if err != nil {
		t.Fatalf("Reconcile pool delete failed: %v", err)
	}

	if result.RequeueAfter == 0 {
		t.Error("expected requeue when pool deletion is blocked by allocated segments")
	}

	updatedPool := &networkextensionv1.HostNetPortPool{}
	if err := r.client.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: "test-pool",
	}, updatedPool); err != nil {
		t.Fatalf("get pool failed: %v", err)
	}

	if !controllerutil.ContainsFinalizer(updatedPool, constant.FinalizerNameHostNetPortPool) {
		t.Error("finalizer should remain when deletion is blocked")
	}
}

func TestReconcilePool_ShrinkEvent(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10020, 5)
	pool.Finalizers = []string{constant.FinalizerNameHostNetPortPool}

	r, cache := newPoolReconciler(pool)
	cache.AddPool(pool)
	r.cacheSynced = true

	_, _, err := cache.AllocateContiguous("default/test-pool", "node-1", "default/high-port-pod", 4)
	if err != nil {
		t.Fatalf("pre-allocate failed: %v", err)
	}

	pool.Spec.EndPort = 10010
	if err := r.client.Update(context.Background(), pool); err != nil {
		t.Fatalf("update pool: %v", err)
	}

	result, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pool"},
	})
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	if result.RequeueAfter == 0 {
		t.Error("expected requeue for shrink conflict")
	}
}

func TestReconcilePool_ResolveEvent(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10020, 5)
	pool.Finalizers = []string{constant.FinalizerNameHostNetPortPool}

	r, cache := newPoolReconciler(pool)
	cache.AddPool(pool)
	r.cacheSynced = true

	_, _, err := cache.AllocateContiguous("default/test-pool", "node-1", "default/high-pod", 4)
	if err != nil {
		t.Fatalf("pre-allocate failed: %v", err)
	}

	pool.Spec.EndPort = 10010
	if err := r.client.Update(context.Background(), pool); err != nil {
		t.Fatalf("update pool: %v", err)
	}

	result, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pool"},
	})
	if err != nil {
		t.Fatalf("first reconcile failed: %v", err)
	}
	if result.RequeueAfter == 0 {
		t.Fatal("expected requeue for shrink conflict")
	}

	cache.ReleaseByPodKey("default/high-pod")

	result, err = r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pool"},
	})
	if err != nil {
		t.Fatalf("second reconcile failed: %v", err)
	}

	updatedPool := &networkextensionv1.HostNetPortPool{}
	_ = r.client.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: "test-pool",
	}, updatedPool)

	if updatedPool.Status.Status != "Ready" {
		t.Errorf("expected Ready status, got %s", updatedPool.Status.Status)
	}

	start, end, exists := cache.GetPoolRange("default/test-pool")
	if !exists {
		t.Fatal("pool not found in cache after resolved shrink")
	}
	if start != 10000 || end != 10010 {
		t.Errorf("expected cache range 10000-10010, got %d-%d", start, end)
	}
}

func TestReconcilePool_NoResolveExp(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10010, 5)
	pool.Finalizers = []string{constant.FinalizerNameHostNetPortPool}

	r, cache := newPoolReconciler(pool)
	cache.AddPool(pool)
	r.cacheSynced = true

	pool.Spec.EndPort = 10030
	if err := r.client.Update(context.Background(), pool); err != nil {
		t.Fatalf("update pool: %v", err)
	}

	_, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pool"},
	})
	if err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}

	start, end, exists := cache.GetPoolRange("default/test-pool")
	if !exists {
		t.Fatal("pool not found in cache")
	}
	if start != 10000 || end != 10030 {
		t.Errorf("expected cache range 10000-10030, got %d-%d", start, end)
	}
}

func TestReconcilePool_MultiStatus(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10020, 5)
	pool.Finalizers = []string{constant.FinalizerNameHostNetPortPool}

	r, cache := newPoolReconciler(pool)
	cache.AddPool(pool)
	r.cacheSynced = true

	_, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pool"},
	})
	if err != nil {
		t.Fatalf("first Reconcile failed: %v", err)
	}

	cache.AllocateContiguous("default/test-pool", "node-1", "default/pod-1", 2)

	_, err = r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pool"},
	})
	if err != nil {
		t.Fatalf("second Reconcile failed: %v", err)
	}

	updatedPool := &networkextensionv1.HostNetPortPool{}
	_ = r.client.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: "test-pool",
	}, updatedPool)

	if updatedPool.Status.Status != "Ready" {
		t.Errorf("expected Ready status, got %s", updatedPool.Status.Status)
	}
}

func TestReconcilePool_SkipDupStatus(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10010, 5)
	pool.Finalizers = []string{constant.FinalizerNameHostNetPortPool}
	pool.Status.Status = "Ready"

	r, cache := newPoolReconciler(pool)
	cache.AddPool(pool)
	r.cacheSynced = true

	result, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pool"},
	})
	if err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Errorf("expected no requeue (event-driven), got %v", result.RequeueAfter)
	}

	prePool := &networkextensionv1.HostNetPortPool{}
	_ = r.client.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: "test-pool",
	}, prePool)
	preRV := prePool.ResourceVersion

	result, err = r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test-pool"},
	})
	if err != nil {
		t.Fatalf("second reconcile failed: %v", err)
	}

	postPool := &networkextensionv1.HostNetPortPool{}
	_ = r.client.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: "test-pool",
	}, postPool)

	if postPool.ResourceVersion != preRV {
		t.Errorf("resourceVersion changed (%s → %s); redundant status update was NOT skipped",
			preRV, postPool.ResourceVersion)
	}
}

// --- InitCache tests ---

func TestInitCache(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10010, 5)
	pod := newPod("test-pod", "default", "node-1", "test-pool")
	pod.Annotations[constant.AnnotationForHostNetPortPoolBindingResult] = `{"poolName":"test-pool","poolNamespace":"default","nodeName":"node-1","startPort":10000,"endPort":10004,"segmentLength":5}`

	r, _ := newPoolReconciler(pool, pod, newNode("node-1"))

	if err := r.initCache(); err != nil {
		t.Fatalf("initCache failed: %v", err)
	}

	r.cacheSynced = true
}

func TestInitCache_SkipsTerminalPods(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10020, 5)

	runningPod := newPod("running-pod", "default", "node-1", "test-pool")
	runningPod.Annotations[constant.AnnotationForHostNetPortPoolBindingResult] =
		`{"poolName":"test-pool","poolNamespace":"default","nodeName":"node-1","startPort":10000,"endPort":10004,"segmentLength":5}`

	failedPod := newPod("failed-pod", "default", "node-1", "test-pool")
	failedPod.Status.Phase = k8scorev1.PodFailed
	failedPod.Annotations[constant.AnnotationForHostNetPortPoolBindingResult] =
		`{"poolName":"test-pool","poolNamespace":"default","nodeName":"node-1","startPort":10005,"endPort":10009,"segmentLength":5}`

	succeededPod := newPod("succeeded-pod", "default", "node-1", "test-pool")
	succeededPod.Status.Phase = k8scorev1.PodSucceeded
	succeededPod.Annotations[constant.AnnotationForHostNetPortPoolBindingResult] =
		`{"poolName":"test-pool","poolNamespace":"default","nodeName":"node-1","startPort":10010,"endPort":10014,"segmentLength":5}`

	r, _ := newPoolReconciler(pool, runningPod, failedPod, succeededPod, newNode("node-1"))

	if err := r.initCache(); err != nil {
		t.Fatalf("initCache failed: %v", err)
	}

	segs := r.cache.GetAllocatedSegments()
	if len(segs) != 1 {
		t.Fatalf("expected only 1 recovered segment (running pod), got %d", len(segs))
	}
	if segs[0].PodKey != "default/running-pod" {
		t.Errorf("expected running-pod, got %s", segs[0].PodKey)
	}
}

func TestInitCache_CorruptedJSON(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10010, 5)
	pod := newPod("bad-pod", "default", "node-1", "test-pool")
	pod.Annotations[constant.AnnotationForHostNetPortPoolBindingResult] = `{invalid json}`

	r, _ := newPoolReconciler(pool, pod, newNode("node-1"))

	if err := r.initCache(); err != nil {
		t.Fatalf("initCache should not fail on bad JSON, got: %v", err)
	}

	segs := r.cache.GetAllocatedSegments()
	if len(segs) != 0 {
		t.Errorf("corrupted JSON pod should be skipped, got %d segments", len(segs))
	}
}

func TestInitCache_NoBindResult(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10010, 5)
	pod := newPod("pending-pod", "default", "node-1", "test-pool")

	r, _ := newPoolReconciler(pool, pod, newNode("node-1"))

	if err := r.initCache(); err != nil {
		t.Fatalf("initCache failed: %v", err)
	}

	segs := r.cache.GetAllocatedSegments()
	if len(segs) != 0 {
		t.Errorf("pod without binding result should be skipped, got %d segments", len(segs))
	}
}

func TestInitCache_CleansStaleNodes(t *testing.T) {
	pool := newPool("test-pool", "default", 10000, 10020, 5)

	podOnAlive := newPod("pod-alive", "default", "alive-node", "test-pool")
	podOnAlive.Annotations[constant.AnnotationForHostNetPortPoolBindingResult] =
		`{"poolName":"test-pool","poolNamespace":"default","nodeName":"alive-node","startPort":10000,"endPort":10004,"segmentLength":5}`

	podOnStale := newPod("pod-stale", "default", "deleted-node", "test-pool")
	podOnStale.Annotations[constant.AnnotationForHostNetPortPoolBindingResult] =
		`{"poolName":"test-pool","poolNamespace":"default","nodeName":"deleted-node","startPort":10005,"endPort":10009,"segmentLength":5}`

	// Only alive-node exists; deleted-node is gone.
	r, cache := newPoolReconciler(pool, podOnAlive, podOnStale, newNode("alive-node"))

	if err := r.initCache(); err != nil {
		t.Fatalf("initCache failed: %v", err)
	}

	segs := cache.GetAllocatedSegments()
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment (alive-node only), got %d", len(segs))
	}
	if segs[0].NodeName != "alive-node" {
		t.Errorf("expected alive-node segment, got %s", segs[0].NodeName)
	}
}

// --- Helper function tests ---

func TestNodeAllocationsEqual(t *testing.T) {
	tests := []struct {
		name string
		a    []*networkextensionv1.NodeHostNetPortPoolStatus
		b    []*networkextensionv1.NodeHostNetPortPoolStatus
		want bool
	}{
		{
			name: "both nil",
			a:    nil,
			b:    nil,
			want: true,
		},
		{
			name: "nil vs empty",
			a:    nil,
			b:    []*networkextensionv1.NodeHostNetPortPoolStatus{},
			want: true,
		},
		{
			name: "different length",
			a: []*networkextensionv1.NodeHostNetPortPoolStatus{
				{NodeName: "n1", AllocatedCount: 1, TotalSegments: 10},
			},
			b:    nil,
			want: false,
		},
		{
			name: "same content",
			a: []*networkextensionv1.NodeHostNetPortPoolStatus{
				{NodeName: "n1", AllocatedCount: 2, TotalSegments: 10},
				{NodeName: "n2", AllocatedCount: 0, TotalSegments: 10},
			},
			b: []*networkextensionv1.NodeHostNetPortPoolStatus{
				{NodeName: "n2", AllocatedCount: 0, TotalSegments: 10},
				{NodeName: "n1", AllocatedCount: 2, TotalSegments: 10},
			},
			want: true,
		},
		{
			name: "different allocated count",
			a: []*networkextensionv1.NodeHostNetPortPoolStatus{
				{NodeName: "n1", AllocatedCount: 1, TotalSegments: 10},
			},
			b: []*networkextensionv1.NodeHostNetPortPoolStatus{
				{NodeName: "n1", AllocatedCount: 2, TotalSegments: 10},
			},
			want: false,
		},
		{
			name: "different node name",
			a: []*networkextensionv1.NodeHostNetPortPoolStatus{
				{NodeName: "n1", AllocatedCount: 1, TotalSegments: 10},
			},
			b: []*networkextensionv1.NodeHostNetPortPoolStatus{
				{NodeName: "n2", AllocatedCount: 1, TotalSegments: 10},
			},
			want: false,
		},
		{
			name: "different total segments",
			a: []*networkextensionv1.NodeHostNetPortPoolStatus{
				{NodeName: "n1", AllocatedCount: 1, TotalSegments: 10},
			},
			b: []*networkextensionv1.NodeHostNetPortPoolStatus{
				{NodeName: "n1", AllocatedCount: 1, TotalSegments: 20},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nodeAllocationsEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("nodeAllocationsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
