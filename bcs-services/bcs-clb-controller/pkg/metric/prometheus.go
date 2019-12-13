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

package metric

import (
	"github.com/emicklei/go-restful"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	testDurations = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "test_datas_seconds",
			Help:       "test data distributions.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"service"},
	)
)

func init() {
	prometheus.MustRegister(testDurations)
}

// PromMetric metric for prometheus
type PromMetric struct {
}

// NewPromMetric create prometheus metric
func NewPromMetric() Resource {
	return &PromMetric{}
}

// Register implements Resource interface
func (p *PromMetric) Register(container *restful.Container) {
	container.Handle("/metrics", promhttp.Handler())
}
