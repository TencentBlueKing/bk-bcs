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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	// ResourceTypePod is pod resource type.
	ResourceTypePod = "Pod"
	// ResourceTypeReplicaSet is replicaset resource type.
	ResourceTypeReplicaSet = "ReplicaSet"
)

const (
	// AnnotationWorkloadKindKey is annotation key for the workload kind a pod belongs to.
	AnnotationWorkloadKindKey = "workload.bkbcs.tencent.com/kind"
	// AnnotationWorkloadNameKey is annotation key for the workload name a pod belongs to.
	AnnotationWorkloadNameKey = "workload.bkbcs.tencent.com/name"
)

// getControllerOwner returns the controller owner reference from the given
// owner references. It prefers the one with Controller == true, and falls
// back to the first owner reference when none is marked as controller.
func getControllerOwner(owners []metav1.OwnerReference) *metav1.OwnerReference {
	if len(owners) == 0 {
		return nil
	}
	for i := range owners {
		if owners[i].Controller != nil && *owners[i].Controller {
			return &owners[i]
		}
	}
	return &owners[0]
}

// resolveWorkload resolves the top-level workload (kind, name) that the given
// pod belongs to by walking up the ownerReferences chain. For pods owned by a
// ReplicaSet, it looks up the ReplicaSet from the shared local cache to find
// its owning Deployment. The resolved result is false while the ReplicaSet
// cache is still performing its initial sync.
func (w *Watcher) resolveWorkload(pod *unstructured.Unstructured) (kind, name string, resolved bool) {
	owner := getControllerOwner(pod.GetOwnerReferences())
	if owner == nil {
		return "", "", true
	}

	// pods managed by a Deployment are owned by a ReplicaSet, walk one more
	// hop to resolve the top-level Deployment.
	if owner.Kind == ResourceTypeReplicaSet {
		dep, resolved := w.resolveReplicaSetOwner(pod.GetNamespace(), owner.Name)
		if !resolved {
			return "", "", false
		}
		if dep != nil {
			return dep.Kind, dep.Name, true
		}
		// Fall back only after the ReplicaSet cache is synced or when the
		// ReplicaSet resource is not watched.
		return owner.Kind, owner.Name, true
	}

	// other kinds (StatefulSet/DaemonSet/Job/GameDeployment/GameStatefulSet ...)
	// are already the top-level workload.
	return owner.Kind, owner.Name, true
}

// resolveReplicaSetOwner looks up the ReplicaSet from the shared local cache and
// returns its controller owner (typically a Deployment). The resolved result is
// false while the ReplicaSet cache has not completed its initial sync.
func (w *Watcher) resolveReplicaSetOwner(namespace, name string) (*metav1.OwnerReference, bool) {
	sw, ok := w.sharedWatchers[ResourceTypeReplicaSet]
	if !ok || sw == nil {
		return nil, true
	}
	if !sw.HasSynced() {
		return nil, false
	}

	key := name
	if namespace != "" {
		key = namespace + "/" + name
	}
	obj, exist, err := sw.GetByKey(key)
	if err != nil {
		return nil, false
	}
	if !exist {
		return nil, true
	}
	rs, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, true
	}
	return getControllerOwner(rs.GetOwnerReferences()), true
}
