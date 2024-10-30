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

// Package nodeinfocheck xxx
package nodeinfocheck

const (
	pluginName           = "nodeinfocheck"
	normalStatus         = "ok"
	errorStatus          = "error"
	initContent          = `interval: 600`
	nodeItemTarget       = "node"
	ZoneItemType         = "zone"
	RegionItemType       = "region"
	InstanceTypeItemType = "instance type"
)

var (
	ChinenseStringMap = map[string]string{
		pluginName:           "节点信息检查",
		normalStatus:         normalStatus,
		nodeItemTarget:       "节点",
		ZoneItemType:         "可用区",
		RegionItemType:       "地域",
		InstanceTypeItemType: "机型",
	}

	EnglishStringMap = map[string]string{
		pluginName:           pluginName,
		normalStatus:         normalStatus,
		nodeItemTarget:       nodeItemTarget,
		ZoneItemType:         ZoneItemType,
		RegionItemType:       RegionItemType,
		InstanceTypeItemType: InstanceTypeItemType,
	}

	StringMap = ChinenseStringMap
)
