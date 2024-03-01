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

// Package metric xxx
package metric

import (
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricClient is interface for prometheus client
// nolint
type MetricClient interface {
	Start()
}

// SidecarMetricClient is implementation of MetricClient
type SidecarMetricClient struct {
	Host string
	Port string
}

// NewMetricClient create a new MetricClient
func NewMetricClient(host, port string) MetricClient {
	return &SidecarMetricClient{
		Host: host,
		Port: port,
	}
}

// Start starts the metric client
func (c *SidecarMetricClient) Start() {
	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))
	go func() {
		err := http.ListenAndServe(fmt.Sprintf("%s:%s", c.Host, c.Port), nil)
		if err != nil {
			blog.Errorf("Metric client stopped: %s", err.Error())
		}
	}()
	blog.Infof("Metric client start...")
}
