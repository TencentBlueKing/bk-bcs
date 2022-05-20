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

	"github.com/prometheus/client_golang/prometheus"
)

type PromServer struct{}

var (
	hrCreateDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "bkbcs",
		Subsystem: "gameworkload",
		Name:      "hookrun_create_duration_seconds",
		Help:      "create duration(seconds) of hookrun",
		Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
	}, []string{"namespace", "name", "status", "action", "objectKind"})
)

func init() {
	prometheus.MustRegister(hrCreateDuration)
}

func (p *PromServer) CollectHRCreateDurations(namespace, name, status, action, objectKind string, d time.Duration) {
	hrCreateDuration.WithLabelValues(namespace, name, status, action, objectKind).Observe(d.Seconds())
}
