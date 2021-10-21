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

package v1

import (
	metricTypes "github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
)

// MetricList contains the metric resource which order by namespaces
type MetricList []*metricTypes.Metric

func (m MetricList) Len() int           { return len(m) }
func (m MetricList) Less(i, j int) bool { return m[i].Namespace < m[j].Namespace }
func (m MetricList) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

type listMetricQuery struct {
	ClusterID []string `json:"clusterID"`
	Name      string   `json:"name"`
}

type MetricTaskList []*metricTypes.MetricTask

func (m MetricTaskList) Len() int           { return len(m) }
func (m MetricTaskList) Less(i, j int) bool { return m[i].Namespace < m[j].Namespace }
func (m MetricTaskList) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
