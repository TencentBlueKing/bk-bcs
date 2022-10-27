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

package feature

import (
	"github.com/spf13/cast"
	"k8s.io/apimachinery/pkg/version"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

// k8s 特性门控  https://kubernetes.io/zh-cn/docs/reference/command-line-tools-reference/feature-gates/#feature-gates
const (
	// ImmutableEphemeralVolumes 允许将各个 Secret 和 ConfigMap 标记为不可变更的，以提高安全性和性能。
	ImmutableEphemeralVolumes = "ImmutableEphemeralVolumes"
)

// k8s 特性门控与版本关联
var featureGateVerMap = map[string]version.Info{
	ImmutableEphemeralVolumes: {Major: "1", Minor: "19"},
}

// 通过比较集群版本，判断特性是否可用
// 注意：集群获取到的 Minor 版本可能不完全是数字，如 20+, 这里截取前缀做判断
func isFeatureEnabled(cVer, fVer *version.Info) bool {
	return cast.ToInt(cVer.Major) >= cast.ToInt(fVer.Major) &&
		cast.ToInt(stringx.ExtractNumberPrefix(cVer.Minor)) >= cast.ToInt(fVer.Minor)
}

// GenFeatureGates 根据指定集群版本，获取特性门控表（通过比较版本大小）
func GenFeatureGates(ver *version.Info) map[string]bool {
	gates := map[string]bool{}
	for k, v := range featureGateVerMap {
		gates[k] = isFeatureEnabled(ver, &v)
	}
	return gates
}
