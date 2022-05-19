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
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

// SetDefaults_GeneralPodAutoscaler sets default values
func SetDefaults_GeneralPodAutoscaler(obj *GeneralPodAutoscaler) {
	if obj.Spec.Behavior == nil {
		obj.Spec.Behavior = new(GeneralPodAutoscalerBehavior)
	}

	// set defautls for scale down behavior
	if obj.Spec.Behavior.ScaleDown == nil {
		obj.Spec.Behavior.ScaleDown = new(GPAScalingRules)
	}
	if obj.Spec.Behavior.ScaleDown.StabilizationWindowSeconds == nil {
		obj.Spec.Behavior.ScaleDown.StabilizationWindowSeconds = new(int32)
		*obj.Spec.Behavior.ScaleDown.StabilizationWindowSeconds = 300
	}
	if obj.Spec.Behavior.ScaleDown.SelectPolicy == nil {
		obj.Spec.Behavior.ScaleDown.SelectPolicy = new(ScalingPolicySelect)
		*obj.Spec.Behavior.ScaleDown.SelectPolicy = MaxPolicySelect

	}
	if len(obj.Spec.Behavior.ScaleDown.Policies) == 0 {
		obj.Spec.Behavior.ScaleDown.Policies = append(obj.Spec.Behavior.ScaleDown.Policies,
			GPAScalingPolicy{
				Type:          PercentScalingPolicy,
				Value:         100,
				PeriodSeconds: 15,
			})
	}

	// set defautls for scale up behavior
	if obj.Spec.Behavior.ScaleUp == nil {
		obj.Spec.Behavior.ScaleUp = new(GPAScalingRules)
	}
	if obj.Spec.Behavior.ScaleUp.StabilizationWindowSeconds == nil {
		obj.Spec.Behavior.ScaleUp.StabilizationWindowSeconds = new(int32)
		*obj.Spec.Behavior.ScaleUp.StabilizationWindowSeconds = 0
	}
	if obj.Spec.Behavior.ScaleUp.SelectPolicy == nil {
		obj.Spec.Behavior.ScaleUp.SelectPolicy = new(ScalingPolicySelect)
		*obj.Spec.Behavior.ScaleUp.SelectPolicy = MaxPolicySelect

	}
	if len(obj.Spec.Behavior.ScaleUp.Policies) == 0 {
		obj.Spec.Behavior.ScaleUp.Policies = append(obj.Spec.Behavior.ScaleUp.Policies,
			GPAScalingPolicy{
				Type:          PercentScalingPolicy,
				Value:         100,
				PeriodSeconds: 15,
			})
		obj.Spec.Behavior.ScaleUp.Policies = append(obj.Spec.Behavior.ScaleUp.Policies,
			GPAScalingPolicy{
				Type:          PodsScalingPolicy,
				Value:         4,
				PeriodSeconds: 15,
			})
	}
}
