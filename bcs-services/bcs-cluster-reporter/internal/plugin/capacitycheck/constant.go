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

// Package capacitycheck xxx
package capacitycheck

const (
	initContent               = `interval: 600`
	pluginName                = "capacitycheck"
	ClusterCapacityMetricName = "cluster_capacity"

	// status
	NormalStatus = "ok"

	ServiceMaxNumCheckItemType = "service max num"
	ServiceNumCheckItemType    = "service available num"
	ServiceCidrCheckItemType   = "service cidr"
	NodeCidrNumCheckItemType   = "cidr max node num"
	NodeMaxPodCheckItemType    = "node max pod num"
	ObjectNumItemType          = "object num"
	MasterTarget               = "master node"
	MasterNumItemType          = "master num"
	MasterZoneItemType         = "master zone"
	NodeTypeItemType           = "node type"
	MasterNumDetailFormart     = "master num is %d, less than 3"
	MasterZoneDetailFormart    = "master num in zone %s is %d, larger than half"
	MasterCheckItemType        = "master"
	MasterZoneHAErrorStatus    = "zone ha error"
	MasterNumHAErrorStatus     = "num ha error"
)

var (
	ChinenseStringMap = map[string]string{
		pluginName:                 "集群容量检查",
		ServiceMaxNumCheckItemType: "service最大数检查",
		ServiceCidrCheckItemType:   "service cidr",
		MasterCheckItemType:        MasterCheckItemType,
		MasterTarget:               MasterTarget,
		MasterNumItemType:          "master节点数量",
		MasterZoneItemType:         "master节点可用区分布",
		MasterNumDetailFormart:     "master数为%d，少于3个",
		MasterZoneDetailFormart:    "%s 的节点数为%d, 超过半数",
		ObjectNumItemType:          "对象实例数",
		NormalStatus:               "正常",
		NodeTypeItemType:           "节点规格",
		NodeCidrNumCheckItemType:   NodeCidrNumCheckItemType,
		ServiceNumCheckItemType:    ServiceNumCheckItemType,
		NodeMaxPodCheckItemType:    NodeMaxPodCheckItemType,
		MasterZoneHAErrorStatus:    MasterZoneHAErrorStatus,
		MasterNumHAErrorStatus:     MasterNumHAErrorStatus,
	}

	EnglishStringMap = map[string]string{
		pluginName: pluginName,

		// status
		NormalStatus: NormalStatus,

		ServiceMaxNumCheckItemType: ServiceMaxNumCheckItemType,
		ServiceCidrCheckItemType:   ServiceCidrCheckItemType,
		MasterCheckItemType:        MasterCheckItemType,
		MasterTarget:               MasterTarget,
		MasterNumItemType:          MasterNumItemType,
		MasterZoneItemType:         MasterZoneItemType,
		MasterNumDetailFormart:     MasterNumDetailFormart,
		MasterZoneDetailFormart:    MasterZoneDetailFormart,
		ObjectNumItemType:          ObjectNumItemType,

		NodeTypeItemType:         NodeTypeItemType,
		NodeCidrNumCheckItemType: NodeCidrNumCheckItemType,
		ServiceNumCheckItemType:  ServiceNumCheckItemType,
		NodeMaxPodCheckItemType:  NodeMaxPodCheckItemType,
		MasterZoneHAErrorStatus:  MasterZoneHAErrorStatus,
		MasterNumHAErrorStatus:   MasterNumHAErrorStatus,
	}

	StringMap = ChinenseStringMap
)
