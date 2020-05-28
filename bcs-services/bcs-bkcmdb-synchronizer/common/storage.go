/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"encoding/json"
)

type EventType int32

const (
	Nop EventType = iota
	Add
	Del
	Chg
	SChg
	Brk EventType = -1
)

const (
	ResourceTypeTaskgroup = "taskgroup"
	ResourceTypePod       = "Pod"
)

// GetResTypeByClusterType get resource type by cluster type
func GetResTypeByClusterType(clusterType string) string {
	switch clusterType {
	case ClusterTypeK8S:
		return ResourceTypePod
	case ClusterTypeMesos:
		return ResourceTypeTaskgroup
	}
	return ""
}

// StorageEvent event for storage
type StorageEvent struct {
	Type  EventType        `json:"type"`
	Value *StorageResource `json:"value"`
}

// StorageResource storage resource
type StorageResource struct {
	ID           string          `json:"_id"`
	ClusterID    string          `json:"clusterId"`
	CreateTime   string          `json:"createTime"`
	UpdateTime   string          `json:"updateTime"`
	Data         json.RawMessage `json:"data"`
	Namespace    string          `json:"namespace"`
	ResourceType string          `json:"resourceType"`
	ResourceName string          `json:"resourceName"`
}

// ListStorageResourceResult result for list storage resource
type ListStorageResourceResult struct {
	Code    int64              `json:"code"`
	Data    []*StorageResource `json:"data"`
	Message string             `json:"message"`
	Result  bool               `json:"result"`
}
