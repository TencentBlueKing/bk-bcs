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

package providers

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-hook-operator/pkg/providers/prometheus"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-hook-operator/pkg/providers/web"
	"github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	"k8s.io/client-go/kubernetes"
)

// Provider methods to query a external systems and generate a measurement
type Provider interface {
	// Run start a new external system call for a measurement
	// Should be idempotent and do nothing if a call has already been started
	Run(*v1alpha1.HookRun, v1alpha1.Metric) v1alpha1.Measurement
	// Checks if the external system call is finished and returns the current measurement
	Resume(*v1alpha1.HookRun, v1alpha1.Metric, v1alpha1.Measurement) v1alpha1.Measurement
	// Terminate will terminate an in-progress measurement
	Terminate(*v1alpha1.HookRun, v1alpha1.Metric, v1alpha1.Measurement) v1alpha1.Measurement
	// GarbageCollect is used to garbage collect completed measurements to the specified limit
	GarbageCollect(*v1alpha1.HookRun, v1alpha1.Metric, int) error
}

type ProviderFactory struct {
	KubeClient kubernetes.Interface
}

// NewProvider creates the correct provider based on the provider type of the Metric
func (f *ProviderFactory) NewProvider(metric v1alpha1.Metric) (Provider, error) {
	if metric.Provider.Web != nil {
		c := web.NewWebMetricHttpClient(metric)
		p, err := web.NewWebMetricJsonParser(metric)
		if err != nil {
			return nil, err
		}
		return web.NewWebMetricProvider(c, p), nil
	} else if metric.Provider.Prometheus != nil {
		api, err := prometheus.NewPrometheusAPI(metric)
		if err != nil {
			return nil, err
		}
		return prometheus.NewPrometheusProvider(api), nil
	}
	return nil, fmt.Errorf("no valid provider in metric '%s'", metric.Name)
}
