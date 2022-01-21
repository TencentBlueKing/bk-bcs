/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package formatter

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"k8s.io/api/autoscaling/v2beta2"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
)

const (
	// 至多展示的 HPA 指标数量
	HPAMetricMaxDisplayNum = 3
	// HPA Metric Current 默认值
	HPAMetricCurrentDefaultVal = "<unknown>"
	// HPA Metric Target 默认值
	HPAMetricTargetDefaultVal = "<auto>"
)

// FormatHPA ...
func FormatHPA(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonFormatRes(manifest)
	ref, _ := util.GetItems(manifest, "spec.scaleTargetRef")
	ret["reference"] = fmt.Sprintf("%s/%s", ref.(map[string]interface{})["kind"], ref.(map[string]interface{})["name"])
	parser := hpaTargetsParser{manifest: manifest}
	ret["targets"] = parser.Parse()
	ret["minPods"] = util.GetWithDefault(manifest, "spec.minReplicas", "<unset>")
	ret["maxPods"] = util.GetWithDefault(manifest, "spec.maxReplicas", "<unset>")
	ret["replicas"] = util.GetWithDefault(manifest, "status.currentReplicas", "--")
	return ret
}

// HPA targets 解析器
type hpaTargetsParser struct {
	manifest map[string]interface{}
	specs    []LightHPAMetricSpec
	statuses []LightHPAMetricStatus
	metrics  []string
}

// targets 解析逻辑，参考来源：https://github.com/kubernetes/kubernetes/blob/master/pkg/printers/internalversion/printers.go#L2025
func (p *hpaTargetsParser) Parse() string {
	specs := util.GetWithDefault(p.manifest, "spec.metrics", []interface{}{})
	if err := mapstructure.Decode(specs, &p.specs); err != nil {
		return "--"
	}
	statuses := util.GetWithDefault(p.manifest, "status.currentMetrics", []interface{}{})
	if err := mapstructure.Decode(statuses, &p.statuses); err != nil {
		return "--"
	}

	if len(p.specs) == 0 {
		return "<none>"
	}

	tooManyItems2Display := false
	for idx, spec := range p.specs {
		// 超过展示数量限制则不再解析，直接截断
		if len(p.metrics) >= HPAMetricMaxDisplayNum {
			tooManyItems2Display = true
			break
		}
		switch spec.Type {
		case v2beta2.ExternalMetricSourceType:
			p.metrics = append(p.metrics, p.parseExternalMetric(idx, spec))
		case v2beta2.PodsMetricSourceType:
			p.metrics = append(p.metrics, p.parsePodMetric(idx, spec))
		case v2beta2.ObjectMetricSourceType:
			p.metrics = append(p.metrics, p.parseObjectMetric(idx, spec))
		case v2beta2.ResourceMetricSourceType:
			p.metrics = append(p.metrics, p.parseResourceMetric(idx, spec))
		case v2beta2.ContainerResourceMetricSourceType:
			p.metrics = append(p.metrics, p.parseContainerResourceMetric(idx, spec))
		default:
			p.metrics = append(p.metrics, "<unknown type>")
		}
	}

	ret := strings.Join(p.metrics, ", ")
	if tooManyItems2Display {
		// 带上 more... 标识有数据被截断未展示
		return ret + fmt.Sprintf(" + %d more...", len(p.specs)-len(p.metrics))
	}
	return ret
}

// 解析来源自 External 的指标信息
func (p *hpaTargetsParser) parseExternalMetric(idx int, spec LightHPAMetricSpec) string {
	current := HPAMetricCurrentDefaultVal
	if spec.External.Target.AverageValue != "" {
		if len(p.statuses) > idx && p.statuses[idx].External != nil && p.statuses[idx].External.Current.AverageValue != "" {
			current = p.statuses[idx].External.Current.AverageValue
		}
		return fmt.Sprintf("%s/%s (avg)", current, spec.External.Target.AverageValue)
	}
	if len(p.statuses) > idx && p.statuses[idx].External != nil {
		current = p.statuses[idx].External.Current.Value
	}
	return fmt.Sprintf("%s/%s", current, spec.External.Target.Value)
}

// 解析来源自 Pods 的指标信息
func (p *hpaTargetsParser) parsePodMetric(idx int, spec LightHPAMetricSpec) string {
	current := HPAMetricCurrentDefaultVal
	if len(p.statuses) > idx && p.statuses[idx].Pods != nil {
		current = p.statuses[idx].Pods.Current.AverageValue
	}
	return fmt.Sprintf("%s/%s", current, spec.Pods.Target.AverageValue)
}

// 解析来源自 Object 的指标信息
func (p *hpaTargetsParser) parseObjectMetric(idx int, spec LightHPAMetricSpec) string {
	current := HPAMetricCurrentDefaultVal
	if spec.Object.Target.AverageValue != "" {
		if len(p.statuses) > idx && p.statuses[idx].Object != nil && p.statuses[idx].Object.Current.AverageValue != "" {
			current = p.statuses[idx].Object.Current.AverageValue
		}
		return fmt.Sprintf("%s/%s (avg)", current, spec.Object.Target.AverageValue)
	}
	if len(p.statuses) > idx && p.statuses[idx].Object != nil {
		current = p.statuses[idx].Object.Current.Value
	}
	return fmt.Sprintf("%s/%s", current, spec.Object.Target.Value)
}

// 解析来源自 Resource 的指标信息
func (p *hpaTargetsParser) parseResourceMetric(idx int, spec LightHPAMetricSpec) string {
	current := HPAMetricCurrentDefaultVal
	if spec.Resource.Target.AverageValue != "" {
		if len(p.statuses) > idx && p.statuses[idx].Resource != nil {
			current = p.statuses[idx].Resource.Current.AverageValue
		}
		return fmt.Sprintf("%s/%s", current, spec.Resource.Target.AverageValue)
	}
	if len(p.statuses) > idx && p.statuses[idx].Resource != nil && p.statuses[idx].Resource.Current.AverageUtilization != 0 { // nolint:lll
		current = fmt.Sprintf("%d%%", p.statuses[idx].Resource.Current.AverageUtilization)
	}
	target := HPAMetricTargetDefaultVal
	if spec.Resource.Target.AverageUtilization != 0 {
		target = fmt.Sprintf("%d%%", spec.Resource.Target.AverageUtilization)
	}
	return fmt.Sprintf("%s/%s", current, target)
}

// 解析来源自 ContainerResource 的指标信息
func (p *hpaTargetsParser) parseContainerResourceMetric(idx int, spec LightHPAMetricSpec) string {
	current := HPAMetricCurrentDefaultVal
	if spec.ContainerResource.Target.AverageValue != "" {
		if len(p.statuses) > idx && p.statuses[idx].ContainerResource != nil {
			current = p.statuses[idx].ContainerResource.Current.AverageValue
		}
		return fmt.Sprintf("%s/%s", current, spec.ContainerResource.Target.AverageValue)
	}
	if len(p.statuses) > idx && p.statuses[idx].ContainerResource != nil && p.statuses[idx].ContainerResource.Current.AverageUtilization != 0 { // nolint:lll
		current = fmt.Sprintf("%d%%", p.statuses[idx].ContainerResource.Current.AverageUtilization)
	}
	target := HPAMetricTargetDefaultVal
	if spec.ContainerResource.Target.AverageUtilization != 0 {
		target = fmt.Sprintf("%d%%", spec.ContainerResource.Target.AverageUtilization)
	}
	return fmt.Sprintf("%s/%s", current, target)
}
