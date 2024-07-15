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

package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ResourceTypeEvent is event resource type.,
	ResourceTypeEvent = "Event"
)

// WatcherInterface describes a resource watcher.
type WatcherInterface interface {
	// Run starts the watcher.
	Run(stopCh <-chan struct{})

	// AddEvent is event to sync new resource.
	AddEvent(obj interface{})

	// DeleteEvent is event to delete old resource.
	DeleteEvent(obj interface{})

	// UpdateEvent is event to update old resource.
	UpdateEvent(oldObj, newObj interface{})

	// GetByKey returns object data by target key.
	GetByKey(key string) (interface{}, bool, error)

	// ListKeys returns all keys in local store.
	ListKeys() []string
}

// CustomResourceDefinitionCore save core info for different version CRDs
type CustomResourceDefinitionCore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec CustomResourceDefinitionCoreSpec
}

// ResourceScope is an enum defining the different scopes available to a custom resource
type ResourceScope string

const (
	// ClusterScoped ResourceScope Cluster
	ClusterScoped ResourceScope = "Cluster" // nolint
	// NamespaceScoped ResourceScope Cluster
	NamespaceScoped ResourceScope = "Namespaced" // nolint
)

// CustomResourceDefinitionCoreSpec save Spec info for CustomResourceDefinitionCore
type CustomResourceDefinitionCoreSpec struct {
	Group       string
	Version     string
	NamesPlural string
	NamesKind   string
	Scope       ResourceScope
}
