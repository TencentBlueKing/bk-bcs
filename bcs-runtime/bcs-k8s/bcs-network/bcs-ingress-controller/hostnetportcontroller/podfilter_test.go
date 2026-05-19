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
	"testing"

	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
)

func podWithAnnotation(name, ns string) *k8scorev1.Pod {
	return &k8scorev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Annotations: map[string]string{
				constant.AnnotationForHostNetPortPool: "pool-1",
			},
		},
	}
}

func plainPod(name, ns string) *k8scorev1.Pod {
	return &k8scorev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
}

func TestPodFilter_Create_WithAnnotation(t *testing.T) {
	f := NewHostNetPodFilter()
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer q.ShutDown()

	f.Create(event.CreateEvent{Object: podWithAnnotation("pod-1", "ns-1")}, q)
	if q.Len() != 1 {
		t.Errorf("expected 1 item in queue, got %d", q.Len())
	}
}

func TestPodFilter_Create_NoAnnotation(t *testing.T) {
	f := NewHostNetPodFilter()
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer q.ShutDown()

	f.Create(event.CreateEvent{Object: plainPod("pod-1", "ns-1")}, q)
	if q.Len() != 0 {
		t.Errorf("expected 0 items in queue, got %d", q.Len())
	}
}

func TestPodFilter_Create_NonPod(t *testing.T) {
	f := NewHostNetPodFilter()
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer q.ShutDown()

	node := &k8scorev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-1"}}
	f.Create(event.CreateEvent{Object: node}, q)
	if q.Len() != 0 {
		t.Errorf("expected 0 items for non-pod")
	}
}

func TestPodFilter_Delete_WithAnnotation(t *testing.T) {
	f := NewHostNetPodFilter()
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer q.ShutDown()

	f.Delete(event.DeleteEvent{Object: podWithAnnotation("pod-1", "ns-1")}, q)
	if q.Len() != 1 {
		t.Errorf("expected 1 item in queue")
	}
}

func TestPodFilter_Delete_NoAnnotation(t *testing.T) {
	f := NewHostNetPodFilter()
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer q.ShutDown()

	f.Delete(event.DeleteEvent{Object: plainPod("pod-1", "ns-1")}, q)
	if q.Len() != 0 {
		t.Errorf("expected 0 items")
	}
}

func TestPodFilter_Delete_NonPod(t *testing.T) {
	f := NewHostNetPodFilter()
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer q.ShutDown()

	f.Delete(event.DeleteEvent{Object: &k8scorev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n"}}}, q)
	if q.Len() != 0 {
		t.Errorf("expected 0 items for non-pod")
	}
}

func TestPodFilter_GenericWithAnno(t *testing.T) {
	f := NewHostNetPodFilter()
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer q.ShutDown()

	f.Generic(event.GenericEvent{Object: podWithAnnotation("pod-1", "ns-1")}, q)
	if q.Len() != 1 {
		t.Errorf("expected 1 item in queue")
	}
}

func TestPodFilter_Generic_NoAnnotation(t *testing.T) {
	f := NewHostNetPodFilter()
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer q.ShutDown()

	f.Generic(event.GenericEvent{Object: plainPod("pod-1", "ns-1")}, q)
	if q.Len() != 0 {
		t.Errorf("expected 0 items")
	}
}

func TestPodFilter_Generic_NonPod(t *testing.T) {
	f := NewHostNetPodFilter()
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer q.ShutDown()

	f.Generic(event.GenericEvent{Object: &k8scorev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n"}}}, q)
	if q.Len() != 0 {
		t.Errorf("expected 0 items")
	}
}

func TestPodFilter_Update_Relevant(t *testing.T) {
	f := NewHostNetPodFilter()
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer q.ShutDown()

	oldPod := podWithAnnotation("pod-1", "ns-1")
	newPod := podWithAnnotation("pod-1", "ns-1")
	newPod.Spec.NodeName = "node-1"

	f.Update(event.UpdateEvent{ObjectOld: oldPod, ObjectNew: newPod}, q)
	if q.Len() != 1 {
		t.Errorf("expected 1 item in queue for relevant change")
	}
}

func TestPodFilter_Update_NoChange(t *testing.T) {
	f := NewHostNetPodFilter()
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer q.ShutDown()

	pod := podWithAnnotation("pod-1", "ns-1")
	pod.Spec.NodeName = "node-1"
	pod.Status.Phase = k8scorev1.PodRunning

	f.Update(event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}, q)
	if q.Len() != 0 {
		t.Errorf("expected 0 items for no relevant change")
	}
}

func TestPodFilter_Update_NoAnnotation(t *testing.T) {
	f := NewHostNetPodFilter()
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer q.ShutDown()

	f.Update(event.UpdateEvent{
		ObjectOld: plainPod("p", "ns"),
		ObjectNew: plainPod("p", "ns"),
	}, q)
	if q.Len() != 0 {
		t.Errorf("expected 0 for plain pod update")
	}
}

func TestPodFilter_Update_NonPod(t *testing.T) {
	f := NewHostNetPodFilter()
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer q.ShutDown()

	node := &k8scorev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n"}}
	f.Update(event.UpdateEvent{ObjectOld: node, ObjectNew: node}, q)
	if q.Len() != 0 {
		t.Errorf("expected 0 for non-pod")
	}
}

func TestCheckHostNetPodReconcile(t *testing.T) {
	tests := []struct {
		name     string
		oldPod   *k8scorev1.Pod
		newPod   *k8scorev1.Pod
		expected bool
	}{
		{"nil old", nil, &k8scorev1.Pod{}, true},
		{"nil new", &k8scorev1.Pod{}, nil, true},
		{
			"node change",
			&k8scorev1.Pod{Spec: k8scorev1.PodSpec{NodeName: "a"}},
			&k8scorev1.Pod{Spec: k8scorev1.PodSpec{NodeName: "b"}},
			true,
		},
		{
			"phase change",
			&k8scorev1.Pod{Status: k8scorev1.PodStatus{Phase: k8scorev1.PodRunning}},
			&k8scorev1.Pod{Status: k8scorev1.PodStatus{Phase: k8scorev1.PodFailed}},
			true,
		},
		{
			"deletion timestamp set on new pod",
			&k8scorev1.Pod{},
			&k8scorev1.Pod{ObjectMeta: metav1.ObjectMeta{DeletionTimestamp: &metav1.Time{}}},
			false,
		},
		{
			"annotation change",
			&k8scorev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				constant.AnnotationForHostNetPortPool: "a",
			}}},
			&k8scorev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				constant.AnnotationForHostNetPortPool: "b",
			}}},
			true,
		},
		{
			"no change",
			&k8scorev1.Pod{
				Spec:   k8scorev1.PodSpec{NodeName: "node-1"},
				Status: k8scorev1.PodStatus{Phase: k8scorev1.PodRunning},
			},
			&k8scorev1.Pod{
				Spec:   k8scorev1.PodSpec{NodeName: "node-1"},
				Status: k8scorev1.PodStatus{Phase: k8scorev1.PodRunning},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkHostNetPortPodNeedReconcile(tt.oldPod, tt.newPod)
			if got != tt.expected {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHasHostNetPortPoolAnnotation(t *testing.T) {
	if hasHostNetPortPoolAnnotation(nil) {
		t.Error("nil annotations should return false")
	}
	if hasHostNetPortPoolAnnotation(map[string]string{}) {
		t.Error("empty annotations should return false")
	}
	if !hasHostNetPortPoolAnnotation(map[string]string{
		constant.AnnotationForHostNetPortPool: "pool",
	}) {
		t.Error("should return true")
	}
}

func TestGetAnnotation(t *testing.T) {
	if got := getAnnotation(nil, "key"); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
	if got := getAnnotation(map[string]string{"k": "v"}, "k"); got != "v" {
		t.Errorf("expected v, got %q", got)
	}
}

func TestIsHostNetPortPoolAnnotation(t *testing.T) {
	if !isHostNetPortPoolAnnotation("hostnetportpool.foo") {
		t.Error("should match")
	}
	if isHostNetPortPoolAnnotation("other.foo") {
		t.Error("should not match")
	}
}
