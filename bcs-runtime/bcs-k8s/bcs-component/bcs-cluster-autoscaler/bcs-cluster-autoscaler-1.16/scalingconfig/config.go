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

// Package scalingconfig xxx
package scalingconfig

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/config"
)

// Options are the option of autoscaler
type Options struct {
	config.AutoscalingOptions
	BufferedCPURatio      float64
	BufferedMemRatio      float64
	BufferedResourceRatio float64
	WebhookMode           string
	WebhookModeConfig     string
	WebhookModeToken      string
	MaxBulkScaleUpCount   int
	ScanInterval          time.Duration
	EvictLatest           bool
}
