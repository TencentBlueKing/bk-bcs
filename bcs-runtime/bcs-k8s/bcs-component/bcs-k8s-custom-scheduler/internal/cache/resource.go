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

package cache

// Resource pod resource
type Resource struct {
	PodName      string
	PodNamespace string
	Node         string
	ResourceKind string
	Value        int
}

// GetPodName get PodName field
func (r *Resource) GetPodName() string {
	return r.PodName
}

// GetPodNamespace get PodNamespace field
func (r *Resource) GetPodNamespace() string {
	return r.PodNamespace
}

// Key get key of resource
func (r *Resource) Key() string {
	return GetMetaKey(r.PodName, r.PodNamespace)
}

// GetNodeName get node name
func (r *Resource) GetNodeName() string {
	return r.Node
}

// DeepCopy deep copy to a new resource object
func (r *Resource) DeepCopy() *Resource {
	return &Resource{
		PodName:      r.PodName,
		PodNamespace: r.PodNamespace,
		Node:         r.Node,
		ResourceKind: r.ResourceKind,
		Value:        r.Value,
	}
}
