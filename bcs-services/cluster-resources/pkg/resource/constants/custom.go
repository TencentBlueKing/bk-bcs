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

package constants

const (
	// DeletionProtectLabelKey 删除保护标记使用的 Label 键名
	DeletionProtectLabelKey = "io.tencent.bcs.dev/deletion-allow"
)

const (
	// DeletionProtectPolicyCascading 实例数量为 0 才可以删除
	DeletionProtectPolicyCascading = "Cascading"
	// DeletionProtectPolicyAlways 不限制，任意时候可删除
	DeletionProtectPolicyAlways = "Always"
	// DeletionProtectPolicyNotAllow 任意时候都无法删除
	DeletionProtectPolicyNotAllow = "NotAllow"
)

const (
	// HookTmplPolicyParallel metrics 并行执行
	HookTmplPolicyParallel = "Parallel"
	// HookTmplPolicyOrdered metrics 顺序执行
	HookTmplPolicyOrdered = "Ordered"
)

const (
	// HookTmplSuccessfulLimit 累计成功数
	HookTmplSuccessfulLimit = "successfulLimit"
	// HookTmplConsecutiveSuccessfulLimit 连续成功数
	HookTmplConsecutiveSuccessfulLimit = "consecutiveSuccessfulLimit"
)

const (
	// HookTmplMetricTypeWeb ...
	HookTmplMetricTypeWeb = "web"
	// HookTmplMetricTypeProm ...
	HookTmplMetricTypeProm = "prometheus"
	// HookTmplMetricTypeK8S ...
	HookTmplMetricTypeK8S = "kubernetes"
)

const (
	// DefaultGWorkloadMaxSurge ...
	DefaultGWorkloadMaxSurge = 0
	// DefaultGWorkloadMaxUnavailable ...
	DefaultGWorkloadMaxUnavailable = 20
	// DefaultGWorkloadGracePeriodSecs 默认优雅更新时间
	DefaultGWorkloadGracePeriodSecs = 30
)
