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

package custom

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
	// HookTmplMetricTypeWeb TODO
	HookTmplMetricTypeWeb = "web"
	// HookTmplMetricTypeProm TODO
	HookTmplMetricTypeProm = "prometheus"
	// HookTmplMetricTypeK8S TODO
	HookTmplMetricTypeK8S = "kubernetes"
)

const (
	// DefaultGWorkloadMaxSurge TODO
	DefaultGWorkloadMaxSurge = 0
	// DefaultGWorkloadMaxUnavailable TODO
	DefaultGWorkloadMaxUnavailable = 20
	// DefaultGWorkloadGracePeriodSecs 默认优雅更新时间
	DefaultGWorkloadGracePeriodSecs = 30
)
