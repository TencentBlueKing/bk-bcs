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

package metric

import (
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/prometheus/client_golang/prometheus"
)

func newMetricController(conf Config, metrics ...*MetricContructor) (*MetricController, error) {
	metricController := new(MetricController)
	meta := MetaData{
		Module:     strings.Replace(conf.ModuleName, "-", "_", -1),
		IP:         conf.IP,
		MetricPort: conf.MetricPort,
		ClusterID:  conf.ClusterID,
	}
	if err := meta.Valid(); nil != err {
		return nil, err
	}
	metricController.Meta = meta

	// initial metrics
	var ms []*MetricContructor
	for _, metric := range metrics {
		ms = append(ms, metric)
	}
	metricController.Metrics = ms
	return metricController, nil
}

// MetricController controller implementation
type MetricController struct {
	Meta    MetaData
	Metrics []*MetricContructor
}

func (m MetricController) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range m.Metrics {
		blog.V(5).Infof("describe metric name- > %s\n", metric.GetMeta().Name)
		m.initGaugeMetric(metric).Describe(ch)
	}

	newVersionMetric(m.Meta).Describe(ch)
	newModuleMetric(m.Meta).Describe(ch)
	newRuntimeMetric(m.Meta).Describe(ch)
}

func (m MetricController) Collect(ch chan<- prometheus.Metric) {
	for _, metric := range m.Metrics {
		blog.V(5).Infof("collect metric - > %s\n", metric.GetMeta().Name)
		done := make(chan struct{})

		go func(mtc *MetricContructor) {
			result, err := mtc.GetResult()
			if err != nil {
				blog.Errorf("get metric result failed. err: %v", err)
				return
			}
			base := metric.GetMeta()
			// add special labels
			if base.ConstLables == nil {
				base.ConstLables = make(map[string]string)
			}
			base.ConstLables[module_ip_label] = m.Meta.IP
			base.ConstLables[module_cluster_id_label] = m.Meta.ClusterID

			var varLabelsKey, varLablesValue []string
			for k, v := range result.VariableLabels {
				varLabelsKey = append(varLabelsKey, k)
				varLablesValue = append(varLablesValue, v)
			}
			var g *prometheus.GaugeVec
			switch result.Value.Type {
			case Float:
				g = prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Namespace:   m.Meta.Module,
					Name:        base.Name,
					Help:        base.Help,
					ConstLabels: prometheus.Labels(base.ConstLables),
				}, varLabelsKey)
				g.WithLabelValues(varLablesValue...).Set(result.Value.Float)
			case String:
				varLabelsKey = append(varLabelsKey, "bcs_metric_value")
				varLablesValue = append(varLablesValue, result.Value.String)
				g = prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Namespace:   m.Meta.Module,
					Name:        base.Name,
					Help:        base.Help,
					ConstLabels: prometheus.Labels(base.ConstLables),
				}, varLabelsKey)
				g.WithLabelValues(varLablesValue...).Set(1)
			default:
				blog.Errorf("unsupported metric value type: %s", result.Value.Type)
				done <- struct{}{}
				return
			}
			g.Collect(ch)
			done <- struct{}{}
		}(metric)

		timeout := time.After(10 * time.Second)
		select {
		case <-timeout:
			blog.Errorf("get metric %s timeout, skip.", metric.GetMeta().Name)
			continue
		case <-done:
			close(done)
		}
	}

	newVersionMetric(m.Meta).Collect(ch)
	newModuleMetric(m.Meta).Collect(ch)
	newRuntimeMetric(m.Meta).Collect(ch)
}

func (m *MetricController) initGaugeMetric(metric *MetricContructor) prometheus.Gauge {
	base := metric.GetMeta()
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   m.Meta.Module,
		Subsystem:   "",
		Name:        base.Name,
		Help:        base.Help,
		ConstLabels: prometheus.Labels(base.ConstLables),
	})

	return gauge
}
