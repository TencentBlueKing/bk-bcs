// Copyright 2021 The BCS Authors.
// Copyright 2021 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package app

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-general-pod-autoscaler/pkg/apis/config/v1alpha1"
)

func RecommendedDefaultGPAControllerConfiguration(obj *v1alpha1.GPAControllerConfiguration) {
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
	if obj.GeneralPodAutoscalerTolerance == 0 {
		obj.GeneralPodAutoscalerTolerance = 0.1
	}
}
