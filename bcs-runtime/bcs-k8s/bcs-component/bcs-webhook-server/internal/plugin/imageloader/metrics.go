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

package imageloader

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/metrics"
)

var (
	actionCreate = "Create"
	actionDelete = "Delete"
	actionRun    = "Run"

	statusSuccess = "Success"
	statusFailure = "Failure"
)

var (
	// operation time
	imageloaderJobDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metrics.BkBcsWebhookServer,
		Name:      "image_loader_job_duration",
		Help:      "The duration of jobs for image loader",
	}, []string{"name", "action", "status"})

	// status
	imageloaderJobStatus = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metrics.BkBcsWebhookServer,
		Name:      "image_loader_job_status",
		Help:      "The status of jobs for image loader",
	}, []string{"name", "action", "status"})
)

func init() {
	prometheus.MustRegister(imageloaderJobDuration)
	prometheus.MustRegister(imageloaderJobStatus)
}

func collectJobDuration(name, action, status string, t time.Duration) {
	duration := t.Seconds()
	imageloaderJobDuration.WithLabelValues(name, action, status).Observe(duration)
}

func collectJobStatus(name, action, status string) {
	imageloaderJobStatus.WithLabelValues(name, action, status).Inc()
}
