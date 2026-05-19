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

// Package hostnetportpoolcache provides in-memory cache for HostNetPortPool segment allocation.
package hostnetportpoolcache

// HostNetPortPoolBindingResult is serialized as JSON into Pod annotation to convey allocation results.
type HostNetPortPoolBindingResult struct {
	PoolName      string `json:"poolName"`
	PoolNamespace string `json:"poolNamespace"`
	NodeName      string `json:"nodeName"`
	StartPort     int    `json:"startPort"`
	EndPort       int    `json:"endPort"`
	SegmentLength int    `json:"segmentLength"`
}

// ConflictSegment describes a segment that conflicts with a pool shrink operation.
type ConflictSegment struct {
	NodeName  string
	StartPort int
	EndPort   int
	PodKey    string
}
