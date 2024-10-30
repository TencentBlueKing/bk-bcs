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

// Package plugin xxx
package plugin

import "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"

const (
	RISKLevel    = "RISK"
	WARNLevel    = "WARN"
	SERIOUSLevel = "SERIOUS"

	CheckFlagNotSetDetailFormat = "%s not found, recommand set as %.2f"
	CheckFlagLeDetailFormat     = "%s value is %s, recommand >= %.2f"
)

var (
	LevelColor = map[string]util.Color{
		RISKLevel:    {Red: 60, Green: 17, Blue: 14},
		WARNLevel:    {Red: 98, Green: 85, Blue: 29},
		SERIOUSLevel: {Red: 98, Green: 29, Blue: 41},
	}
	ChinenseStringMap = map[string]string{
		CheckFlagNotSetDetailFormat: "未找到参数%s, 推荐设置该参数为%.2f",
		CheckFlagLeDetailFormat:     "%s 的值为%s, 推荐设置为%.2f",
	}

	EnglishStringMap = map[string]string{
		CheckFlagNotSetDetailFormat: CheckFlagNotSetDetailFormat,
		CheckFlagLeDetailFormat:     CheckFlagLeDetailFormat,
	}

	StringMap = ChinenseStringMap
)
