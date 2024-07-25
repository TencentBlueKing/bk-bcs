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

package service

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/metrics"
)

func initMetric() *metric {
	m := new(metric)
	m.currentUploadedFolderSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metrics.Namespace,
		Subsystem: metrics.FSConfigConsume,
		Name:      "upload_file_directory_size_bytes",
		Help:      "Size of the directory in bytes",
	}, []string{"bizID", "resourceID", "directory"})
	metrics.Register().MustRegister(m.currentUploadedFolderSize)

	return m
}

type metric struct {
	// currentUploadedFolderSize Record the current uploaded folder size
	currentUploadedFolderSize *prometheus.GaugeVec
}
