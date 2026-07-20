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

// Package k8s xxx
package k8s

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// fakeWatcher is a minimal WatcherInterface implementation for tests, backed by a map.
type fakeWatcher struct {
	store map[string]interface{}
}

func (f *fakeWatcher) Run(stopCh <-chan struct{}) {}

func (f *fakeWatcher) AddEvent(obj interface{}) {}

func (f *fakeWatcher) DeleteEvent(obj interface{}) {}

func (f *fakeWatcher) UpdateEvent(oldObj, newObj interface{}) {}

func (f *fakeWatcher) GetByKey(key string) (interface{}, bool, error) {
	obj, ok := f.store[key]
	return obj, ok, nil
}

func (f *fakeWatcher) ListKeys() []string {
	keys := make([]string, 0, len(f.store))
	for k := range f.store {
		keys = append(keys, k)
	}
	return keys
}

func boolPtr(b bool) *bool { return &b }

func podWithOwners(namespace, name string, owners []metav1.OwnerReference) *unstructured.Unstructured {
	pod := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
		},
	}
	pod.SetOwnerReferences(owners)
	return pod
}

func TestGetControllerOwner(t *testing.T) {
	tests := []struct {
		desc     string
		owners   []metav1.OwnerReference
		wantNil  bool
		wantKind string
		wantName string
	}{
		{
			desc:    "empty owners returns nil",
			owners:  nil,
			wantNil: true,
		},
		{
			desc: "prefer controller owner",
			owners: []metav1.OwnerReference{
				{Kind: "ReplicaSet", Name: "rs-a", Controller: boolPtr(false)},
				{Kind: "ReplicaSet", Name: "rs-b", Controller: boolPtr(true)},
			},
			wantKind: "ReplicaSet",
			wantName: "rs-b",
		},
		{
			desc: "fall back to first when no controller flag",
			owners: []metav1.OwnerReference{
				{Kind: "ReplicaSet", Name: "rs-a"},
			},
			wantKind: "ReplicaSet",
			wantName: "rs-a",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := getControllerOwner(test.owners)
			if test.wantNil {
				if got != nil {
					t.Fatalf("got %+v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("got nil, want kind=%s name=%s", test.wantKind, test.wantName)
			}
			if got.Kind != test.wantKind || got.Name != test.wantName {
				t.Errorf("got kind=%s name=%s, want kind=%s name=%s", got.Kind, got.Name, test.wantKind, test.wantName)
			}
		})
	}
}

func TestResolveWorkload(t *testing.T) {
	// ReplicaSet owned by a Deployment, cached in the shared ReplicaSet watcher.
	rs := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "ReplicaSet",
			"metadata": map[string]interface{}{
				"name":      "web-6d8f",
				"namespace": "ns1",
			},
		},
	}
	rs.SetOwnerReferences([]metav1.OwnerReference{
		{Kind: "Deployment", Name: "web", Controller: boolPtr(true)},
	})

	rsWatcher := &fakeWatcher{store: map[string]interface{}{
		"ns1/web-6d8f": rs,
	}}

	w := &Watcher{
		resourceType: ResourceTypePod,
		sharedWatchers: map[string]WatcherInterface{
			ResourceTypeReplicaSet: rsWatcher,
		},
	}

	tests := []struct {
		desc     string
		pod      *unstructured.Unstructured
		wantKind string
		wantName string
	}{
		{
			desc: "pod -> replicaset -> deployment",
			pod: podWithOwners("ns1", "web-6d8f-abc", []metav1.OwnerReference{
				{Kind: "ReplicaSet", Name: "web-6d8f", Controller: boolPtr(true)},
			}),
			wantKind: "Deployment",
			wantName: "web",
		},
		{
			desc: "replicaset not cached, fall back to replicaset",
			pod: podWithOwners("ns1", "orphan-pod", []metav1.OwnerReference{
				{Kind: "ReplicaSet", Name: "missing-rs", Controller: boolPtr(true)},
			}),
			wantKind: "ReplicaSet",
			wantName: "missing-rs",
		},
		{
			desc: "pod owned by statefulset directly",
			pod: podWithOwners("ns1", "sts-0", []metav1.OwnerReference{
				{Kind: "StatefulSet", Name: "db", Controller: boolPtr(true)},
			}),
			wantKind: "StatefulSet",
			wantName: "db",
		},
		{
			desc:     "pod without owner returns empty",
			pod:      podWithOwners("ns1", "bare-pod", nil),
			wantKind: "",
			wantName: "",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			kind, name := w.resolveWorkload(test.pod)
			if kind != test.wantKind || name != test.wantName {
				t.Errorf("got kind=%q name=%q, want kind=%q name=%q", kind, name, test.wantKind, test.wantName)
			}
		})
	}
}
