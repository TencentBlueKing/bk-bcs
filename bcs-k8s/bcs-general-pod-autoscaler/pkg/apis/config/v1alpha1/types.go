/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GPAControllerConfiguration contains elements describing GPAController.
type GPAControllerConfiguration struct {
	metav1.TypeMeta `json:",inline"`
	// generalPodAutoscalerSyncPeriod is the period for syncing the number of
	// pods in general pod autoscaler.
	GeneralPodAutoscalerSyncPeriod metav1.Duration
	// generalPodAutoscalerUpscaleForbiddenWindow is a period after which next upscale allowed.
	GeneralPodAutoscalerUpscaleForbiddenWindow metav1.Duration
	// generalPodAutoscalerDownscaleForbiddenWindow is a period after which next downscale allowed.
	GeneralPodAutoscalerDownscaleForbiddenWindow metav1.Duration
	// GeneralPodAutoscalerDowncaleStabilizationWindow is a period for which autoscaler will look
	// backwards and not scale down below any recommendation it made during that period.
	GeneralPodAutoscalerDownscaleStabilizationWindow metav1.Duration
	// generalPodAutoscalerTolerance is the tolerance for when
	// resource usage suggests upscaling/downscaling
	GeneralPodAutoscalerTolerance float64
	// GeneralPodAutoscalerUseRESTClients causes the GPA controller to use REST clients
	// through the kube-aggregator when enabled, instead of using the legacy metrics client
	// through the API server proxy.
	GeneralPodAutoscalerUseRESTClients bool
	// GeneralPodAutoscalerCPUInitializationPeriod is the period after pod start when CPU samples
	// might be skipped.
	GeneralPodAutoscalerCPUInitializationPeriod metav1.Duration
	// GeneralPodAutoscalerInitialReadinessDelay is period after pod start during which readiness
	// changes are treated as readiness being set for the first time. The only effect of this is that
	// GPA will disregard CPU samples from unready pods that had last readiness change during that
	// period.
	GeneralPodAutoscalerInitialReadinessDelay metav1.Duration
}
