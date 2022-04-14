/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"strconv"
	"strings"
)

const (
	// UnitCnt 单位：个
	UnitCnt = "cnt"
	// UnitPercent 单位：%
	UnitPercent = "percent"
)

// AnalyzeIntStr 分析可能是 int 或者 string 的字段，比如 spec.strategy.rollingUpdate.maxSurge，返回值与单位
// 规则：以 % 为结尾的，单位是 %，否则单位为个数
func AnalyzeIntStr(raw interface{}) (int64, string) {
	switch r := raw.(type) {
	case int64:
		return r, UnitCnt
	case string:
		val, _ := strconv.ParseInt(r[:len(r)-1], 10, 64)
		return val, UnitPercent
	}
	return 0, UnitCnt
}

// ConvertCPUUnit 将 resource 中定义的 CPU 配置统一为 mCpus 为单位
// 支持示例：1000m / 1，500m / 0.5
func ConvertCPUUnit(raw string) int {
	if raw == "" {
		return 0
	}
	if strings.Contains(raw, "m") {
		val, _ := strconv.Atoi(strings.Replace(raw, "m", "", 1))
		return val
	}
	val, _ := strconv.ParseFloat(raw, 64)
	return int(val * 1000)
}

// ConvertMemoryUnit 将 resource 中定义的 Memory 配置统一为 Mi 为单位
// 支持示例：10Mi，10M，1Gi，2G
func ConvertMemoryUnit(raw string) int {
	if strings.Contains(raw, "M") {
		raw = strings.Replace(raw, "Mi", "", 1)
		raw = strings.Replace(raw, "M", "", 1)
		val, _ := strconv.Atoi(raw)
		return val
	} else if strings.Contains(raw, "G") {
		raw = strings.Replace(raw, "Gi", "", 1)
		raw = strings.Replace(raw, "G", "", 1)
		val, _ := strconv.Atoi(raw)
		return val * 1024
	}
	return 0
}
