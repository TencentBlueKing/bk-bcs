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

// Package cache xxx
package cache

import (
	"fmt"
	"sort"
	"sync"
)

// Node node info
type Node struct {
	NodeName  string
	Resources map[string]*Resource
}

// ResourceCache cache of node resource
type ResourceCache struct {
	lock      sync.Mutex
	Resources map[string]*Resource
	Nodes     map[string]*Node
}

// NewResourceCache new
func NewResourceCache() *ResourceCache {
	return &ResourceCache{
		Resources: make(map[string]*Resource),
		Nodes:     make(map[string]*Node),
	}
}

// GetResource get resource
func (rc *ResourceCache) GetResource(key string) *Resource {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	r, ok := rc.Resources[key]
	if !ok {
		return nil
	}
	return r.DeepCopy()
}

// GetNodes get node name list
func (rc *ResourceCache) GetNodes() []string {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	var retList []string
	for nodeName := range rc.Nodes {
		retList = append(retList, nodeName)
	}
	return retList
}

// GetNodeResources get node resource
func (rc *ResourceCache) GetNodeResources(nodeName string) []*Resource {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	node, ok := rc.Nodes[nodeName]
	if !ok {
		return nil
	}
	retRes := make([]*Resource, 0)
	for _, r := range node.Resources {
		retRes = append(retRes, r.DeepCopy())
	}
	sort.Slice(retRes, func(i, j int) bool {
		return retRes[i].Key() < retRes[j].Key()
	})
	return retRes
}

// GetAllResources get all resources
func (rc *ResourceCache) GetAllResources() []*Resource {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	retRes := make([]*Resource, 0)
	for _, node := range rc.Nodes {
		for _, r := range node.Resources {
			retRes = append(retRes, r.DeepCopy())
		}
	}
	sort.Slice(retRes, func(i, j int) bool {
		return retRes[i].Key() < retRes[j].Key()
	})
	return retRes
}

// UpdateResource update resource
func (rc *ResourceCache) UpdateResource(r *Resource) error {
	if r == nil {
		return fmt.Errorf("resource to update cannot be empty")
	}
	if len(r.GetNodeName()) == 0 || len(r.GetPodName()) == 0 || len(r.GetPodNamespace()) == 0 {
		return fmt.Errorf("PodName, NodeName, PodNamespace cannot be empty")
	}
	rc.lock.Lock()
	defer rc.lock.Unlock()
	// add to node resource
	node, okNode := rc.Nodes[r.GetNodeName()]
	if !okNode {
		rc.Nodes[r.GetNodeName()] = &Node{
			NodeName:  r.GetNodeName(),
			Resources: make(map[string]*Resource),
		}
		node = rc.Nodes[r.GetNodeName()]
	}
	node.Resources[r.Key()] = r

	// update resources map
	rc.Resources[r.Key()] = r
	return nil
}

// DeleteResource delete resource
func (rc *ResourceCache) DeleteResource(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("key cannot be empty")
	}
	rc.lock.Lock()
	defer rc.lock.Unlock()
	r, okRes := rc.Resources[key]
	if !okRes {
		return nil
	}
	delete(rc.Resources, key)

	node, okNode := rc.Nodes[r.GetNodeName()]
	if !okNode {
		return nil
	}
	delete(node.Resources, key)
	return nil
}
