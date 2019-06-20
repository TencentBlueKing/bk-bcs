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

package collector

import (
	"encoding/json"
	"fmt"
	"io"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func ParsePromMetric(metric io.Reader, cLabels map[string]string) ([]byte, error) {
	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(metric)
	if err != nil {
		return []byte{}, fmt.Errorf("new parser failed, err: %v", err)
	}

	var result []*Family
	for _, m := range mf {
		result = append(result, NewFamily(m, cLabels))
	}

	return json.Marshal(MetricWrapper{result})
}

type MetricWrapper struct {
	Data []*Family `json:"data"`
}

type Family struct {
	//Time    time.Time
	Name    string        `json:"name"`
	Help    string        `json:"help"`
	Type    string        `json:"type"`
	Metrics []interface{} `json:"metrics,omitempty"` // Either metric or summary.
}

// Metric is for all "single value" metrics, i.e. Counter, Gauge, and Untyped.
type Metric struct {
	Labels map[string]string `json:"labels,omitempty"`
	Value  string            `json:"value"`
}

// Summary mirrors the Summary proto message.
type Summary struct {
	Labels    map[string]string `json:"labels,omitempty"`
	Quantiles map[string]string `json:"quantiles,omitempty"`
	Count     string            `json:"count"`
	Sum       string            `json:"sum"`
}

// Histogram mirrors the Histogram proto message.
type Histogram struct {
	Labels  map[string]string `json:"labels,omitempty"`
	Buckets map[string]string `json:"buckets,omitempty"`
	Count   string            `json:"count"`
	Sum     string            `json:"sum"`
}

// NewFamily consumes a MetricFamily and transforms it to the local Family type.
func NewFamily(dtoMF *dto.MetricFamily, cLabels map[string]string) *Family {
	mf := &Family{
		//Time:    time.Now(),
		Name:    dtoMF.GetName(),
		Help:    dtoMF.GetHelp(),
		Type:    dtoMF.GetType().String(),
		Metrics: make([]interface{}, len(dtoMF.Metric)),
	}
	for i, m := range dtoMF.Metric {
		if dtoMF.GetType() == dto.MetricType_SUMMARY {
			mf.Metrics[i] = Summary{
				Labels:    makeLabels(m, cLabels),
				Quantiles: makeQuantiles(m),
				Count:     fmt.Sprint(m.GetSummary().GetSampleCount()),
				Sum:       fmt.Sprint(m.GetSummary().GetSampleSum()),
			}
		} else if dtoMF.GetType() == dto.MetricType_HISTOGRAM {
			mf.Metrics[i] = Histogram{
				Labels:  makeLabels(m, cLabels),
				Buckets: makeBuckets(m),
				Count:   fmt.Sprint(m.GetHistogram().GetSampleCount()),
				Sum:     fmt.Sprint(m.GetSummary().GetSampleSum()),
			}
		} else {
			mf.Metrics[i] = Metric{
				Labels: makeLabels(m, cLabels),
				Value:  fmt.Sprint(getValue(m)),
			}
		}
	}
	return mf
}

func getValue(m *dto.Metric) float64 {
	if m.Gauge != nil {
		return m.GetGauge().GetValue()
	}
	if m.Counter != nil {
		return m.GetCounter().GetValue()
	}
	if m.Untyped != nil {
		return m.GetUntyped().GetValue()
	}
	return 0.
}

func makeLabels(m *dto.Metric, cLabels map[string]string) map[string]string {
	result := map[string]string{}
	for _, lp := range m.Label {
		result[lp.GetName()] = lp.GetValue()
	}

	// add user defined const labels,
	// which is also a relabel opeation, which will replace the existence label's value.
	for k, v := range cLabels {
		result[k] = v
	}

	return result
}

func makeQuantiles(m *dto.Metric) map[string]string {
	result := map[string]string{}
	for _, q := range m.GetSummary().Quantile {
		result[fmt.Sprint(q.GetQuantile())] = fmt.Sprint(q.GetValue())
	}
	return result
}

func makeBuckets(m *dto.Metric) map[string]string {
	result := map[string]string{}
	for _, b := range m.GetHistogram().Bucket {
		result[fmt.Sprint(b.GetUpperBound())] = fmt.Sprint(b.GetCumulativeCount())
	}
	return result
}
