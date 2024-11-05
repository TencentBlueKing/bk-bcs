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

// Package pluginmanager xxx
package pluginmanager

import "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"

const (
	RISKLevel    = "RISK"
	WARNLevel    = "WARN"
	SERIOUSLevel = "SERIOUS"

	ClusterAvailabilityCheckItemName = "ClusterAvailabilityCheckItemName"
	ClusterAvailabilityOkStatus      = "ClusterAvailabilityOkStatus"
	ClusterAvailabilityPanicStatus   = "ClusterAvailabilityPanicStatus"

	// checkitemname
	NormalStatus = "ok"

	ClusterID       = "ClusterID"
	CheckItemName   = "CheckItemName"
	CheckItemType   = "CheckItemType"
	CheckItemTarget = "CheckItemTarget"
	CheckItemResult = "CheckItemResult"
	CheckItemLevel  = "CheckItemLevel"
	CheckItemDetail = "CheckItemDetail"

	CheckItemSolution = "CheckItemSolution"

	promotStrFormat = "%s %s result is %s, detail: %s"
)

var (
	LevelColor = map[string]util.Color{
		RISKLevel:    {Red: 60, Green: 17, Blue: 14},
		WARNLevel:    {Red: 98, Green: 85, Blue: 29},
		SERIOUSLevel: {Red: 98, Green: 29, Blue: 41},
	}
	ChinenseStringMap = map[string]string{
		ClusterAvailabilityCheckItemName: "集群可用性",
		ClusterAvailabilityOkStatus:      "ok",
		ClusterAvailabilityPanicStatus:   "panic",
		NormalStatus:                     "ok",

		CheckItemLevel:  "问题等级",
		CheckItemDetail: "检测详情",
		CheckItemName:   "检测项",
		CheckItemType:   "检测类型",
		CheckItemTarget: "检测对象",
		CheckItemResult: "检测结果",

		ClusterID: "集群ID",

		promotStrFormat: "%s针对%s的检查结果为%s, 检查详情:%s",

		CheckItemSolution: "检测详情",
	}

	EnglishStringMap = map[string]string{
		ClusterAvailabilityCheckItemName: "cluster availability",
		ClusterAvailabilityOkStatus:      "ok",
		ClusterAvailabilityPanicStatus:   "ok",
		NormalStatus:                     "ok",

		CheckItemName:   "check item",
		CheckItemDetail: "detail",
		CheckItemType:   "check item type",
		CheckItemTarget: "check item target",
		CheckItemResult: "check item result",

		ClusterID: "clusterID",

		promotStrFormat: promotStrFormat,

		CheckItemSolution: "check item solution",
	}

	StringMap = ChinenseStringMap
)
