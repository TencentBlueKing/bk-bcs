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

package servermetric

import (
	"bk-bcs/bcs-common/common/metric"
	"sync"
)

var (
	Metrics    ServerMetric
	metricLock sync.Mutex
)

type ServerMetric struct {
	role              metric.RoleType
	dcStatus          bool
	clusterStatus     bool
	clusterStatusDesc string
}

func SetRole(role metric.RoleType) {
	metricLock.Lock()
	defer metricLock.Unlock()

	Metrics.role = role
}

func SetDCStatus(dcStatus bool) {
	metricLock.Lock()
	defer metricLock.Unlock()

	Metrics.dcStatus = dcStatus
}

func SetClusterStatus(clusterStatus bool, msg string) {
	metricLock.Lock()
	defer metricLock.Unlock()

	Metrics.clusterStatus = clusterStatus
	Metrics.clusterStatusDesc = msg
}

func GetRole() metric.RoleType {

	metricLock.Lock()
	defer metricLock.Unlock()

	return Metrics.role
}

func IsHealthy() (bool, string) {

	metricLock.Lock()
	defer metricLock.Unlock()

	if Metrics.dcStatus != true {
		return false, "storage error"
	}

	if Metrics.clusterStatus != true {
		return false, Metrics.clusterStatusDesc
	}

	return true, "run ok"
}
