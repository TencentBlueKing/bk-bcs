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

package observer

import (
	"github.com/prometheus/client_golang/prometheus"
	prm "github.com/prometheus/client_golang/prometheus"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/metrics"
)

func initMetric(name string) *metric {
	m := new(metric)
	labels := prm.Labels{"name": name}

	m.lastCursor = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.FSObserver,
		Name:        "last_cursor",
		Help:        "record the last consumed event cursor id by feed server",
		ConstLabels: labels,
	}, []string{})
	metrics.Register().MustRegister(m.lastCursor)

	return m
}

type metric struct {
	// lastCursor record the last consumed cursor id.
	lastCursor *prometheus.GaugeVec
}
