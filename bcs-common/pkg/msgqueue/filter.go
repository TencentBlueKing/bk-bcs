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

package msgqueue

// EventKind event type
type EventKind string

const (
	// EventTypeAdd for add event
	EventTypeAdd EventKind = "add"
	// EventTypeUpdate for update event
	EventTypeUpdate EventKind = "update"
	// EventTypeDelete for delete event
	EventTypeDelete EventKind = "delete"
	// EventTypeUnknown for unknown event type
	EventTypeUnknown EventKind = "other"
)

// FilterType cluster resource meta
type FilterType string

var (
	// ClusterID for meta clusterId
	ClusterID FilterType = "clusterId"
	// Namespace for meta namespace
	Namespace FilterType = "namespace"
	// ResourceType for meta k8s resourceType
	ResourceType FilterType = "resourceType"
	// ResourceKind for meta resourceKind
	ResourceKind FilterType = "resourceKind"
	// ResourceName for meta resourceName
	ResourceName FilterType = "resourceName"
	// EventLevel for event level
	EventLevel FilterType = "level"
	// EventType for resource event
	EventType FilterType = "event"
)

// Filter data filter
type Filter interface {
	Filter(meta map[string]string) bool
}

// DefaultClusterFilter subscribe specific kind data by ID
type DefaultClusterFilter struct {
	FilterKind FilterType
	ClusterID  string
}

// Filter filter clusterID resource by meta
func (filter *DefaultClusterFilter) Filter(meta map[string]string) bool {
	if id, ok := meta[string(filter.FilterKind)]; ok && id == filter.ClusterID {
		return true
	}
	return false
}

// DefaultNamespaceFilter subscribe specific kind data by Namespace
type DefaultNamespaceFilter struct {
	FilterKind FilterType
	Namespace  string
}

// Filter filter namespace resource by meta
func (filter *DefaultNamespaceFilter) Filter(meta map[string]string) bool {
	if namespace, ok := meta[string(filter.FilterKind)]; ok && namespace == filter.Namespace {
		return true
	}
	return false
}

// DefaultResourceTypeFilter subscribe specific kind data by ResourceType
type DefaultResourceTypeFilter struct {
	FilterKind   FilterType
	ResourceType string
}

// Filter filter resourceType resource by meta
func (filter *DefaultResourceTypeFilter) Filter(meta map[string]string) bool {
	if resourceType, ok := meta[string(filter.FilterKind)]; ok && resourceType == filter.ResourceType {
		return true
	}
	return false
}

// DefaultResourceNameFilter subscribe specific kind data by ResourceName
type DefaultResourceNameFilter struct {
	FilterKind   FilterType
	ResourceName string
}

// Filter filter resourceName resource by meta
func (filter *DefaultResourceNameFilter) Filter(meta map[string]string) bool {
	if name, ok := meta[string(filter.FilterKind)]; ok && name == filter.ResourceName {
		return true
	}
	return false
}

// DefaultEventFilter subscribe specific kind data by Event
type DefaultEventFilter struct {
	FilterKind FilterType
	EventType  string
}

// Filter filter resourceEvent resource by meta
func (filter *DefaultEventFilter) Filter(meta map[string]string) bool {
	if event, ok := meta[string(filter.FilterKind)]; ok && event == filter.EventType {
		return true
	}
	return false
}
