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

// Package types xxx
package types

// HandleHunBuNamespace 混部集群参数
type HandleHunBuNamespace struct {
	HandleNormalNamespace
	Labels map[string]string `json:"labels"`
}

// HandleNormalNamespace 普通集群参数
type HandleNormalNamespace struct {
	FedClusterId string            `json:"fed_cluster_id"`
	SubClusterId string            `json:"sub_cluster_id"`
	Namespace    string            `json:"namespace"`
	Annotations  map[string]string `json:"annotations"`
}
