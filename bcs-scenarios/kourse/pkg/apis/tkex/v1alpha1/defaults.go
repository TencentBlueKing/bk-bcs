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
 *
 */

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

// SetDefaults_GameStatefulSet sets defaults for gamestatefulset
func SetDefaults_GameStatefulSet(obj *GameStatefulSet) {
	if len(obj.Spec.PodManagementPolicy) == 0 {
		obj.Spec.PodManagementPolicy = OrderedReadyPodManagement
	}

	if obj.Spec.UpdateStrategy.Type == "" {
		obj.Spec.UpdateStrategy.Type = OnDeleteGameStatefulSetStrategyType
	}

	if obj.Spec.UpdateStrategy.Type != OnDeleteGameStatefulSetStrategyType &&
		obj.Spec.UpdateStrategy.RollingUpdate == nil {
		rollingUpate := RollingUpdateStatefulSetStrategy{}
		obj.Spec.UpdateStrategy.RollingUpdate = &rollingUpate
	}

	if obj.Spec.UpdateStrategy.Type != OnDeleteGameStatefulSetStrategyType &&
		obj.Spec.UpdateStrategy.RollingUpdate != nil &&
		obj.Spec.UpdateStrategy.RollingUpdate.Partition == nil {
		partition := intstr.FromInt(0)
		obj.Spec.UpdateStrategy.RollingUpdate.Partition = &partition
	}

	if obj.Spec.UpdateStrategy.Type != OnDeleteGameStatefulSetStrategyType &&
		obj.Spec.UpdateStrategy.RollingUpdate != nil &&
		obj.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable == nil {
		maxUnavailable := intstr.FromString("25%")
		obj.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable = &maxUnavailable
	}

	if obj.Spec.UpdateStrategy.Type != OnDeleteGameStatefulSetStrategyType &&
		obj.Spec.UpdateStrategy.RollingUpdate != nil &&
		obj.Spec.UpdateStrategy.RollingUpdate.MaxSurge == nil {
		maxSurge := intstr.FromInt(0)
		obj.Spec.UpdateStrategy.RollingUpdate.MaxSurge = &maxSurge
	}

	if obj.Spec.Replicas == nil {
		obj.Spec.Replicas = new(int32)
		*obj.Spec.Replicas = 1
	}
	if obj.Spec.RevisionHistoryLimit == nil {
		obj.Spec.RevisionHistoryLimit = new(int32)
		*obj.Spec.RevisionHistoryLimit = 10
	}
}
