/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// nolint
package metric_manager

import (
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestSendMessage(t *testing.T) {
	MM.RunPrometheusMetricsServer()
	vec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cluster_availability",
		Help: "cluster_availability, 1 means OK",
	}, []string{"target", "target_biz", "status"})
	Register(vec)
	SetMetric(vec, []*GaugeVecSet{
		{
			Labels: []string{"1", "2", "3"},
			Value:  1,
		},
	})

	// nolint
	vec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cluster_availability",
		Help: "cluster_availability, 1 means OK",
	}, []string{"target", "target_biz", "status"})
	MM.SetSeperatedMetric("123")

	_, ok := MM.registryMap["123"]
	fmt.Println(ok)
	time.Sleep(100 * time.Second)
}
