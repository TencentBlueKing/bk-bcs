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
)

func TestNodeFilter_Create(t *testing.T) {
	f := NewHostNetNodeFilter()
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer q.ShutDown()

	f.Create(event.CreateEvent{Object: &k8scorev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n1"}}}, q)
	if q.Len() != 0 {
		t.Errorf("Create should not enqueue, got %d items", q.Len())
	}
}

func TestNodeFilter_Update(t *testing.T) {
	f := NewHostNetNodeFilter()
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer q.ShutDown()

	node := &k8scorev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n1"}}
	f.Update(event.UpdateEvent{ObjectOld: node, ObjectNew: node}, q)
	if q.Len() != 0 {
		t.Errorf("Update should not enqueue, got %d items", q.Len())
	}
}

func TestNodeFilter_Delete(t *testing.T) {
	f := NewHostNetNodeFilter()
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer q.ShutDown()

	f.Delete(event.DeleteEvent{Object: &k8scorev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-x"}}}, q)
	if q.Len() != 1 {
		t.Errorf("expected 1 item, got %d", q.Len())
	}

	item, _ := q.Get()
	defer q.Done(item)
	req, ok := item.(interface{ GetNamespacedName() (string, string) })
	_ = ok
	_ = req
}

func TestNodeFilter_Delete_NonNode(t *testing.T) {
	f := NewHostNetNodeFilter()
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer q.ShutDown()

	f.Delete(event.DeleteEvent{Object: &k8scorev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p"}}}, q)
	if q.Len() != 0 {
		t.Errorf("expected 0 for non-node, got %d", q.Len())
	}
}

func TestNodeFilter_Generic(t *testing.T) {
	f := NewHostNetNodeFilter()
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer q.ShutDown()

	f.Generic(event.GenericEvent{Object: &k8scorev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n1"}}}, q)
	if q.Len() != 0 {
		t.Errorf("Generic should not enqueue, got %d items", q.Len())
	}
}
