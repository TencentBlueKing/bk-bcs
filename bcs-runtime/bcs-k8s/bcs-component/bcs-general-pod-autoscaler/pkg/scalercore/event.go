/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package scalercore

import autoscalingv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/apis/autoscaling/v1alpha1"

var _ Scaler = &EventScaler{}

// EventScaler event scaler
type EventScaler struct {
	schedule string
	name     string
}

// NewEventScaler new event scaler
func NewEventScaler(schedule string) Scaler {
	return &EventScaler{schedule: schedule, name: Event}
}

// Run  run
func (e *EventScaler) Run(stopCh <-chan struct{}) error {
	return nil
}

// GetReplicas get replicas
func (e *EventScaler) GetReplicas(*autoscalingv1.GeneralPodAutoscaler, int32) (int32, error) {
	return 0, nil
}

// ScalerName scaler name
func (s *EventScaler) ScalerName() string {
	return s.name
}
