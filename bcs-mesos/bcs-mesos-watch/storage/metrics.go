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

package storage

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	storageTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bkbcs_datawatch",
		Subsystem: "mesos",
		Name:      "storage_total",
		Help:      "The total number of storage synchronization operation.",
	}, []string{"datatype", "action", "status"})
	storageLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "bkbcs_datawatch",
		Subsystem: "mesos",
		Name:      "storage_latency_seconds",
		Help:      "BCS mesos datawatch storage operation latency statistic.",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"datatype", "action", "status"})

	dataTypeApp                = "Application"
	dataTypeTaskGroup          = "TaskGroup"
	dataTypeCfg                = "Configmap"
	dataTypeSecret             = "Secret"
	dataTypeDeploy             = "Deployment"
	dataTypeSvr                = "Service"
	dataTypeExpSVR             = "ExportService"
	dataTypeEp                 = "Endpoint"
	dataTypeIPPoolStatic       = "IPPoolStatic"
	dataTypeIPPoolStaticDetail = "IPPoolStaticDetail"

	actionDelete = "DELETE"
	actionPut    = "PUT"
	// actionPost   = "POST"

	statusFailure = "FAILURE"
	statusSuccess = "SUCCESS"
)

func reportStorageMetrics(datatype, action, status string, started time.Time) {
	storageTotal.WithLabelValues(datatype, action, status).Inc()
	storageLatency.WithLabelValues(datatype, action, status).Observe(time.Since(started).Seconds())
}

func init() {
	//add golang basic metrics
	prometheus.MustRegister(storageTotal)
	prometheus.MustRegister(storageLatency)
}
