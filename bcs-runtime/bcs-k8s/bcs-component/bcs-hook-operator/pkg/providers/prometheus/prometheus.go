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

package prometheus

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/pkg/util/evaluate"
	metricutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/pkg/util/metric"
	templateutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/pkg/util/template"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"

	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const (
	//ProviderType indicates the provider is prometheus
	ProviderType = "Prometheus"
)

// Provider contains all the required components to run a prometheus query
type Provider struct {
	api v1.API
}

// Type incidates provider is a prometheus provider
func (p *Provider) Type() string {
	return ProviderType
}

// Run queries prometheus for the metric
func (p *Provider) Run(run *v1alpha1.HookRun, metric v1alpha1.Metric) v1alpha1.Measurement {
	startTime := metav1.Now()
	newMeasurement := v1alpha1.Measurement{
		StartedAt: &startTime,
	}

	//TODO(dthomson) make timeout configuriable
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query, err := templateutil.ResolveArgs(metric.Provider.Prometheus.Query, run.Spec.Args)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	response, _, err := p.api.Query(ctx, query, time.Now())
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	newValue, newStatus, err := p.processResponse(metric, response)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)

	}
	newMeasurement.Value = newValue

	newMeasurement.Phase = newStatus
	finishedTime := metav1.Now()
	newMeasurement.FinishedAt = &finishedTime
	return newMeasurement
}

// Resume should not be used the prometheus provider since all the work should occur in the Run method
func (p *Provider) Resume(run *v1alpha1.HookRun, metric v1alpha1.Metric, measurement v1alpha1.Measurement) v1alpha1.Measurement {
	klog.Warningf("HookRun: %s/%s, metric: %s. Prometheus provider should not execute the Resume method", run.Namespace, run.Name, metric.Name)
	return measurement
}

// Terminate should not be used the prometheus provider since all the work should occur in the Run method
func (p *Provider) Terminate(run *v1alpha1.HookRun, metric v1alpha1.Metric, measurement v1alpha1.Measurement) v1alpha1.Measurement {
	klog.Warningf("HookRun: %s/%s, metric: %s. Prometheus provider should not execute the Terminate method", run.Namespace, run.Name, metric.Name)
	return measurement
}

// GarbageCollect is a no-op for the prometheus provider
func (p *Provider) GarbageCollect(run *v1alpha1.HookRun, metric v1alpha1.Metric, limit int) error {
	return nil
}

func (p *Provider) evaluateResult(result interface{}, metric v1alpha1.Metric) v1alpha1.HookPhase {
	successCondition := false
	failCondition := false
	var err error

	if metric.SuccessCondition != "" {
		successCondition, err = evaluate.EvalCondition(result, metric.SuccessCondition)
		if err != nil {
			klog.Warningf(err.Error())
			return v1alpha1.HookPhaseError
		}
	}
	if metric.FailureCondition != "" {
		failCondition, err = evaluate.EvalCondition(result, metric.FailureCondition)
		if err != nil {
			return v1alpha1.HookPhaseError
		}
	}

	switch {
	case metric.SuccessCondition == "" && metric.FailureCondition == "":
		//Always return success unless there is an error
		return v1alpha1.HookPhaseSuccessful
	case metric.SuccessCondition != "" && metric.FailureCondition == "":
		// Without a failure condition, a measurement is considered a failure if the measurement's success condition is not true
		failCondition = !successCondition
	case metric.SuccessCondition == "" && metric.FailureCondition != "":
		// Without a success condition, a measurement is considered a successful if the measurement's failure condition is not true
		successCondition = !failCondition
	}

	if failCondition {
		return v1alpha1.HookPhaseFailed
	}

	if !failCondition && !successCondition {
		return v1alpha1.HookPhaseInconclusive
	}

	// If we reach this code path, failCondition is false and successCondition is true
	return v1alpha1.HookPhaseSuccessful
}

func (p *Provider) processResponse(metric v1alpha1.Metric, response model.Value) (string, v1alpha1.HookPhase, error) {
	switch value := response.(type) {
	case *model.Scalar:
		valueStr := value.Value.String()
		result := float64(value.Value)
		if math.IsNaN(result) {
			return valueStr, v1alpha1.HookPhaseInconclusive, nil
		}
		newStatus := p.evaluateResult(result, metric)
		return valueStr, newStatus, nil
	case model.Vector:
		results := make([]float64, 0, len(value))
		valueStr := "["
		for _, s := range value {
			if s != nil {
				valueStr = valueStr + s.Value.String() + ","
				results = append(results, float64(s.Value))
			}
		}
		// if we appended to the string, we should remove the last comma on the string
		if len(valueStr) > 1 {
			valueStr = valueStr[:len(valueStr)-1]
		}
		valueStr = valueStr + "]"
		for _, result := range results {
			if math.IsNaN(result) {
				return valueStr, v1alpha1.HookPhaseInconclusive, nil
			}
		}
		newStatus := p.evaluateResult(results, metric)
		return valueStr, newStatus, nil
	default:
		return "", v1alpha1.HookPhaseError, fmt.Errorf("Prometheus metric type not supported")
	}
}

// NewPrometheusProvider Creates a new Prometheus client
func NewPrometheusProvider(api v1.API) *Provider {
	return &Provider{
		api: api,
	}
}

// NewPrometheusAPI generates a prometheus API from the metric configuration
func NewPrometheusAPI(metric v1alpha1.Metric) (v1.API, error) {
	client, err := api.NewClient(api.Config{
		Address: metric.Provider.Prometheus.Address,
	})
	if err != nil {
		return nil, err
	}

	return v1.NewAPI(client), nil
}
