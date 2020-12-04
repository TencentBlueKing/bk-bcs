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

package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-hook-operator/pkg/util/evaluate"
	metricutil "github.com/Tencent/bk-bcs/bcs-k8s/bcs-hook-operator/pkg/util/metric"
	templateutil "github.com/Tencent/bk-bcs/bcs-k8s/bcs-hook-operator/pkg/util/template"
	"github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/jsonpath"
	"k8s.io/klog"
)

type Provider struct {
	client     *http.Client
	jsonParser *jsonpath.JSONPath
}

func (p *Provider) Run(run *v1alpha1.HookRun, metric v1alpha1.Metric) v1alpha1.Measurement {
	startTime := metav1.Now()

	// Measurement to pass back
	measurement := v1alpha1.Measurement{
		StartedAt: &startTime,
	}

	// Create request
	request := &http.Request{
		Method: "GET",
	}
	urlStr, err := templateutil.ResolveArgs(metric.Provider.Web.URL, run.Spec.Args)
	if err != nil {
		return metricutil.MarkMeasurementError(measurement, err)
	}

	url, err := url.Parse(urlStr)
	if err != nil {
		return metricutil.MarkMeasurementError(measurement, err)
	}

	request.URL = url

	request.Header = make(http.Header)

	for _, header := range metric.Provider.Web.Headers {
		value, err := templateutil.ResolveArgs(header.Value, run.Spec.Args)
		if err != nil {
			return metricutil.MarkMeasurementError(measurement, err)
		}
		request.Header.Set(header.Key, value)
	}

	// Send Request
	response, err := p.client.Do(request)
	if err != nil {
		return metricutil.MarkMeasurementError(measurement, err)
	} else if response.StatusCode < 200 || response.StatusCode >= 300 {
		return metricutil.MarkMeasurementError(measurement, fmt.Errorf("received non 2xx response code: %v", response.StatusCode))
	}

	value, status, err := p.parseResponse(metric, response)
	if err != nil {
		return metricutil.MarkMeasurementError(measurement, err)
	}

	measurement.Value = value
	measurement.Phase = status
	finishedTime := metav1.Now()
	measurement.FinishedAt = &finishedTime

	return measurement
}

// Resume should not be used the WebMetric provider since all the work should occur in the Run method
func (p *Provider) Resume(run *v1alpha1.HookRun, metric v1alpha1.Metric, measurement v1alpha1.Measurement) v1alpha1.Measurement {
	klog.Warningf("HookRun: %s/%s, metric: %s. WebMetric provider should not execute the Resume method", run.Namespace, run.Name, metric.Name)
	return measurement
}

// Terminate should not be used the WebMetric provider since all the work should occur in the Run method
func (p *Provider) Terminate(run *v1alpha1.HookRun, metric v1alpha1.Metric, measurement v1alpha1.Measurement) v1alpha1.Measurement {
	klog.Warningf("HookRun: %s/%s, metric: %s. WebMetric provider should not execute the Terminate method", run.Namespace, run.Name, metric.Name)
	return measurement
}

// GarbageCollect is a no-op for the WebMetric provider
func (p *Provider) GarbageCollect(run *v1alpha1.HookRun, metric v1alpha1.Metric, limit int) error {
	return nil
}

func (p *Provider) parseResponse(metric v1alpha1.Metric, response *http.Response) (string, v1alpha1.HookPhase, error) {
	var data interface{}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", v1alpha1.HookPhaseError, fmt.Errorf("Received no bytes in response: %v", err)
	}

	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return "", v1alpha1.HookPhaseError, fmt.Errorf("Could not parse JSON body: %v", err)
	}

	buf := new(bytes.Buffer)
	err = p.jsonParser.Execute(buf, data)
	if err != nil {
		return "", v1alpha1.HookPhaseError, fmt.Errorf("Could not find JsonPath in body: %s", err)
	}
	out := buf.String()

	status := evaluate.EvaluateResult(out, metric)
	return out, status, nil
}

func NewWebMetricHttpClient(metric v1alpha1.Metric) *http.Client {
	var timeout time.Duration

	// Using a default timeout of 10 seconds
	if metric.Provider.Web.TimeoutSeconds <= 0 {
		timeout = time.Duration(10) * time.Second
	} else {
		timeout = time.Duration(metric.Provider.Web.TimeoutSeconds) * time.Second
	}

	c := &http.Client{
		Timeout: timeout,
	}
	return c
}

func NewWebMetricJsonParser(metric v1alpha1.Metric) (*jsonpath.JSONPath, error) {
	jsonParser := jsonpath.New("metrics")

	err := jsonParser.Parse(metric.Provider.Web.JsonPath)

	return jsonParser, err
}

func NewWebMetricProvider(client *http.Client, jsonParser *jsonpath.JSONPath) *Provider {
	return &Provider{
		client:     client,
		jsonParser: jsonParser,
	}
}
