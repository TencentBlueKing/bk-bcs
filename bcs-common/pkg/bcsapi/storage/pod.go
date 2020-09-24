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
 *
 */

package storage

import (
	mesostype "github.com/Tencent/bk-bcs/bcs-common/common/types"
	core "k8s.io/api/core/v1"
)

// Pod information from
// https://bcs-api-gateway.bk.tencent.com:8443/bcsapi/v4/storage/query/k8s/dynamic/clusters/BCS-K8S-40026/pod
// data structure difinitions are refered to store structures in mongodb. they are not the the same with
// original kubernetes data structure. Module bcs-storage add some additional control information for
// multiple cluster management.

// Pod definition in mongodb
type Pod struct {
	ID           string    `json:"_id"`
	ResourceName string    `json:"resourceName"`
	ResourceType string    `json:"resourceType"`
	Namespace    string    `json:"namespace"`
	ClusterID    string    `json:"clusterId"`
	CreateTime   string    `json:"createTime"`
	UpdateTime   string    `json:"updateTime"`
	Data         *core.Pod `json:"data"`
}

// Taskgroup bcs-storage taskgroup data of mesos
type Taskgroup struct {
	ID           string                  `json:"_id"`
	ResourceName string                  `json:"resourceName"`
	ResourceType string                  `json:"resourceType"`
	Namespace    string                  `json:"namespace"`
	ClusterID    string                  `json:"clusterId"`
	CreateTime   string                  `json:"create_time"`
	UpdateTime   string                  `json:"update_time"`
	Data         *mesostype.BcsPodStatus `json:"data"`
}
