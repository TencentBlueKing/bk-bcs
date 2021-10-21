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

package metrics

import (
	"time"

	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
)

// TaskgroupMetric contains pod metric value (the metric values are expected to be the metric as a milli-value)
type TaskgroupMetric struct {
	Timestamp time.Time
	Window    int //seconds
	Value     float32
}

const (
	TaskgroupResourcesCpuMetricsName    = "cpu"
	TaskgroupResourcesMemoryMetricsName = "memory"
)

// TaskgroupMetricsInfo contains taskgroup metrics as a map from pod names to TaskgroupMetricsInfo
type TaskgroupMetricsInfo map[string]TaskgroupMetric

// MetricsController collect external metrics or taskgroup resource metrics
type MetricsController interface {
	//start to collect scaler metrics
	StartScalerMetrics(scaler *commtypes.BcsAutoscaler)

	//stop to collect scaler metrics
	StopScalerMetrics(scaler *commtypes.BcsAutoscaler)

	// GetResourceMetric gets the given resource metric (and an associated oldest timestamp)
	// for all taskgroup matching the specified uuid
	GetResourceMetric(resourceName, uuid string) (TaskgroupMetricsInfo, error)
}
