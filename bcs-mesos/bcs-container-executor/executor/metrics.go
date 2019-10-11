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

package executor

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics the scheduler info
var (
	executorSlaveConnection = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "bkbcs",
		Subsystem: "executor",
		Name:      "slave_connection",
		Help:      "executor slave connection",
	})

	taskgroupReportTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bkbcs",
		Subsystem: "executor",
		Name:      "taskgroup_status_report_total",
		Help:      "report taskgroup status total",
	}, []string{"taskgroup"})

	taskgroupAckTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bkbcs",
		Subsystem: "executor",
		Name:      "taskgroup_status_ack_total",
		Help:      "ack  taskgroup status total",
	}, []string{"taskgroup"})
)

func init() {
	prometheus.MustRegister(executorSlaveConnection, taskgroupAckTotal, taskgroupReportTotal)
}
