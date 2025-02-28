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

package podmonitor

// SampleLimitMax xxx
const SampleLimitMax = 100000

// SampleLimitMin xxx
const SampleLimitMin = 1

// ListPodMonitorsReq xxx
type ListPodMonitorsReq struct {
	ProjectCode   string `json:"projectCode" in:"path=projectCode" validate:"required"`
	ClusterId     string `json:"clusterId" in:"path=clusterId" validate:"required"`
	PathNamespace string `json:"namespace" in:"path=namespace"`
	Limit         string `json:"limit" in:"query=limit"`
	Offset        string `json:"offset" in:"query=offset"`
	Namespace     string `json:"namespace" in:"query=namespace"` // nolint:govet
}

// GetPodMonitorReq xxx
type GetPodMonitorReq struct {
	ProjectCode string `json:"projectCode" in:"path=projectCode" validate:"required"`
	ClusterId   string `json:"clusterId" in:"path=clusterId" validate:"required"`
	Namespace   string `json:"namespace" in:"path=namespace" validate:"required"`
	Name        string `json:"name" in:"path=name" validate:"required"`
}

// CreatePodMonitorReq create pod monitor req
type CreatePodMonitorReq struct {
	ProjectCode   string            `json:"projectCode" in:"path=projectCode" validate:"required"`
	ClusterId     string            `json:"clusterId" in:"path=clusterId" validate:"required"`
	PathNamespace string            `json:"namespace" in:"path=namespace" validate:"required"`
	ServiceName   string            `json:"service_name"`
	Path          string            `json:"path"`
	Selector      map[string]string `json:"selector"`
	Interval      string            `json:"interval"`
	Port          string            `json:"port"`
	SampleLimit   int               `json:"sample_limit"`
	Namespace     string            `json:"namespace"` // nolint:govet
	Name          string            `json:"name"`
	Params        map[string]string `json:"params"`
}

// Validate validate
func (r CreatePodMonitorReq) Validate() bool {
	if validateName(r.Name) && validateSelector(r.Selector) && validateSampleLimit(r.SampleLimit) {
		return true
	}
	return false
}

// BatchDeletePodMonitorReq batch delete pod monitor req
type BatchDeletePodMonitorReq struct {
	ProjectCode string       `json:"projectCode" in:"path=projectCode" validate:"required"`
	ClusterId   string       `json:"clusterId" in:"path=clusterId" validate:"required"`
	PodMonitors []PodMonitor `json:"pod_monitors"`
}

// PodMonitor pod monitor
type PodMonitor struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

// UpdatePodMonitorReq update pod monitor req
type UpdatePodMonitorReq struct {
	ProjectCode   string            `json:"projectCode" in:"path=projectCode" validate:"required"`
	ClusterId     string            `json:"clusterId" in:"path=clusterId" validate:"required"`
	PathNamespace string            `json:"namespace" in:"path=namespace" validate:"required"`
	PathName      string            `json:"name" in:"path=name" validate:"required"`
	ServiceName   string            `json:"service_name"`
	Path          string            `json:"path"`
	Selector      map[string]string `json:"selector"`
	Interval      string            `json:"interval"`
	Port          string            `json:"port"`
	SampleLimit   int               `json:"sample_limit"`
	Namespace     string            `json:"namespace"` // nolint:govet
	Name          string            `json:"name"`      // nolint:govet
	Params        map[string]string `json:"params"`
}

// Validate 校验参数
func (r UpdatePodMonitorReq) Validate() bool {
	if validateName(r.Name) && validateSelector(r.Selector) {
		return true
	}
	return false
}

// DeletePodMonitorReq delete pod monitor req
type DeletePodMonitorReq struct {
	ProjectCode string `json:"projectCode" in:"path=projectCode" validate:"required"`
	ClusterId   string `json:"clusterId" in:"path=clusterId" validate:"required"`
	Namespace   string `json:"namespace" in:"path=namespace" validate:"required"`
	Name        string `json:"name" in:"path=name" validate:"required"`
}
