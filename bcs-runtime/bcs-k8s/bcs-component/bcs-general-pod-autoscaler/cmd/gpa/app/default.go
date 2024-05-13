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

// Package app xxx
package app

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/apis/config/v1alpha1"
)

// RecommendDefaultGPAControllerConfig 原方法名 RecommendedDefaultGPAControllerConfiguration
//
// RecommendDefaultGPAControllerConfig recommended default GPA controller configuration
func RecommendDefaultGPAControllerConfig(obj *v1alpha1.GPAControllerConfiguration) {
	zero := metav1.Duration{}
	if obj.GeneralPodAutoscalerSyncPeriod == zero {
		obj.GeneralPodAutoscalerSyncPeriod = metav1.Duration{Duration: 15 * time.Second}
	}
	if obj.GeneralPodAutoscalerUpscaleForbiddenWindow == zero {
		obj.GeneralPodAutoscalerUpscaleForbiddenWindow = metav1.Duration{Duration: 3 * time.Minute}
	}
	if obj.GeneralPodAutoscalerDownscaleStabilizationWindow == zero {
		obj.GeneralPodAutoscalerDownscaleStabilizationWindow = metav1.Duration{Duration: 5 * time.Minute}
	}
	if obj.GeneralPodAutoscalerCPUInitializationPeriod == zero {
		obj.GeneralPodAutoscalerCPUInitializationPeriod = metav1.Duration{Duration: 5 * time.Minute}
	}
	if obj.GeneralPodAutoscalerInitialReadinessDelay == zero {
		obj.GeneralPodAutoscalerInitialReadinessDelay = metav1.Duration{Duration: 30 * time.Second}
	}
	if obj.GeneralPodAutoscalerDownscaleForbiddenWindow == zero {
		obj.GeneralPodAutoscalerDownscaleForbiddenWindow = metav1.Duration{Duration: 5 * time.Minute}
	}
	if obj.GeneralPodAutoscalerTolerance <= 0 {
		obj.GeneralPodAutoscalerTolerance = 0.1
	}
	if obj.GeneralPodAutoscalerWorkers < 1 {
		obj.GeneralPodAutoscalerWorkers = 1
	}
}
