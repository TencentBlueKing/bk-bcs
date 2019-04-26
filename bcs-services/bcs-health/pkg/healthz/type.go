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

package healthz

type HealthResponse struct {
	Code    int    `json:"code"`
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	Data    Status `json:"data"`
}

type Status struct {
	// whether the platform and clusters is healthy or not.
	Healthy        HealthStatus     `json:"healthy"`
	PlatformStatus *PlatformStatus  `json:"platform_status"`
	ClustersStatus []*ClusterStatus `json:"clusters_status"`
}

type PlatformStatus struct {
	// whether the platform is healthy or not
	Healthy HealthStatus `json:"healthy"`
	// health status of each platform component.
	Status map[string]*HealthResult `json:"status"`
}

type ClusterStatus struct {
	// could be one of k8s or mesos.
	Type string `json:"type"`
	// cluster id of container cluster platform
	ClusterID string `json:"cluster_id"`
	// whether this cluster is healthy or not.
	Healthy HealthStatus `json:"healthy"`
	// health status of each component
	Status map[string]*HealthResult `json:"status"`
}

type HealthResult struct {
	Status  HealthStatus `json:"status"`
	Message MsgDetail    `json:"message, omitempty"`
}

type MsgDetail []string

func (d *MsgDetail) Append(msg string) {
	*d = append(*d, msg)
}

type HealthStatus string

const (
	Healthy   HealthStatus = "healthy"
	Unhealthy HealthStatus = "unhealthy"
	Unknown   HealthStatus = "unknown"
)

func AggregateHealthStatus(status ...HealthStatus) HealthStatus {
	if len(status) == 0 {
		return Unknown
	}

	base := status[0]
	for i := 1; i < len(status); i = i + 1 {
		if base != Unknown {
			if status[i] == Unknown {
				base = Unknown
			} else if status[i] == Unhealthy {
				base = Unhealthy
			}
		}
	}

	return base
}
