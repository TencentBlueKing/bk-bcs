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

package formatter

import v1 "k8s.io/api/core/v1"

type LightPodCondition struct {
	Type   v1.PodConditionType
	Status v1.ConditionStatus
}

type LightContainerStateWaiting struct {
	Reason string
}

type LightContainerStateRunning struct {
	StartedAt string
}

type LightContainerStateTerminated struct {
	ExitCode int32
	Signal   int32
	Reason   string
}

type LightContainerState struct {
	Waiting    *LightContainerStateWaiting
	Running    *LightContainerStateRunning
	Terminated *LightContainerStateTerminated
}

type LightContainerStatus struct {
	State LightContainerState
	Ready bool
}

// 轻量化的 PodStatus，主要用于解析 Pod Status 信息
type LightPodStatus struct {
	Phase                 v1.PodPhase
	Conditions            []LightPodCondition
	Reason                string
	InitContainerStatuses []LightContainerStatus
	ContainerStatuses     []LightContainerStatus
}
