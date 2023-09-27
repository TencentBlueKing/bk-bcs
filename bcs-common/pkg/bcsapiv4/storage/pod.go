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

package storage

import (
	core "k8s.io/api/core/v1"

	mesostype "github.com/Tencent/bk-bcs/bcs-common/common/types"
)

// Pod information from
// https://bcs-api-gateway.bk.tencent.com:8443/bcsapi/v4/storage/query/k8s/dynamic/clusters/BCS-K8S-40026/pod
// data structure difinitions are referred to store structures in mongodb. they are not the the same with
// original kubernetes data structure. Module bcs-storage add some additional control information for
// multiple cluster management.

// Pod definition in mongodb
type Pod struct {
	CommonDataHeader
	Data *core.Pod `json:"data"`
}

// Taskgroup bcs-storage taskgroup data of mesos
type Taskgroup struct {
	CommonDataHeader
	Data *mesostype.BcsPodStatus `json:"data"`
}

// PodList is response for storage pod list operation
type PodList struct {
	CommonResponseHeader
	Data []Pod `json:"data"`
}

// TaskgroupList is response for storage taskgroup list operation
type TaskgroupList struct {
	CommonResponseHeader
	Data []Taskgroup `json:"data"`
}
