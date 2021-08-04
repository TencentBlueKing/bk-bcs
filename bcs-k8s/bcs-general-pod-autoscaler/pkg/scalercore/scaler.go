// Copyright 2021 The BCS Authors.
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

package scalercore

import autoscalingv1 "github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-general-pod-autoscaler/pkg/apis/autoscaling/v1alpha1"

type ScalerName string

const (
	Webhook = "Webhook"
	Event   = "Event"
	Cron    = "Cron"
)

type Scaler interface {
	GetReplicas(*autoscalingv1.GeneralPodAutoscaler, int32) (int32, error)
	ScalerName() string
}

type LongRunScaler interface {
	Scaler
	Run(<-chan struct{}) error
}
