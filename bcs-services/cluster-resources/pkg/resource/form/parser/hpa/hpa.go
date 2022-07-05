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

package hpa

import (
	"strconv"
	"strings"

	"github.com/fatih/structs"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/common"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ParseHPA ...
func ParseHPA(manifest map[string]interface{}) map[string]interface{} {
	hpa := model.HPA{}
	common.ParseMetadata(manifest, &hpa.Metadata)
	ParseHPASpec(manifest, &hpa.Spec)
	return structs.Map(hpa)
}

// ParseHPASpec ...
func ParseHPASpec(manifest map[string]interface{}, spec *model.HPASpec) {
	spec.Ref.APIVersion = mapx.GetStr(manifest, "spec.scaleTargetRef.apiVersion")
	spec.Ref.Kind = mapx.GetStr(manifest, "spec.scaleTargetRef.kind")
	spec.Ref.ResName = mapx.GetStr(manifest, "spec.scaleTargetRef.name")
	spec.Ref.MinReplicas = mapx.GetInt64(manifest, "spec.minReplicas")
	spec.Ref.MaxReplicas = mapx.GetInt64(manifest, "spec.maxReplicas")
	for _, metric := range mapx.GetList(manifest, "spec.metrics") {
		m := metric.(map[string]interface{})
		switch m["type"].(string) {
		case HPAMetricTypeRes:
			spec.Resource.Items = append(spec.Resource.Items, genResMetricItem(m))
		case HPAMetricTypeContainerRes:
			spec.ContainerRes.Items = append(spec.ContainerRes.Items, genContainerResMetricItem(m))
		case HPAMetricTypeExternal:
			spec.External.Items = append(spec.External.Items, genExternalMetricItem(m))
		case HPAMetricTypeObject:
			spec.Object.Items = append(spec.Object.Items, genObjectMetricItem(m))
		case HPAMetricTypePod:
			spec.Pod.Items = append(spec.Pod.Items, genPodMetricItem(m))
		}
	}
}

func genResMetricItem(m map[string]interface{}) model.ResourceMetricItem {
	ms := m["resource"].(map[string]interface{})
	return model.ResourceMetricItem{
		Name:  mapx.GetStr(ms, "name"),
		Type:  mapx.GetStr(ms, "target.type"),
		Value: getMetricValue(ms),
	}
}

func genContainerResMetricItem(m map[string]interface{}) model.ContainerResMetricItem {
	ms := m["containerResource"].(map[string]interface{})
	return model.ContainerResMetricItem{
		Name:          mapx.GetStr(m, "name"),
		ContainerName: mapx.GetStr(m, "container"),
		Type:          mapx.GetStr(m, "target.type"),
		Value:         getMetricValue(ms),
	}
}

func genExternalMetricItem(m map[string]interface{}) model.ExternalMetricItem {
	ms := m["external"].(map[string]interface{})
	return model.ExternalMetricItem{
		Name:     mapx.GetStr(ms, "metric.name"),
		Type:     mapx.GetStr(ms, "target.type"),
		Value:    getMetricValue(ms),
		Selector: genMetricSelector(ms),
	}
}

func genObjectMetricItem(m map[string]interface{}) model.ObjectMetricItem {
	ms := m["object"].(map[string]interface{})
	return model.ObjectMetricItem{
		Name:       mapx.GetStr(ms, "metric.name"),
		APIVersion: mapx.GetStr(ms, "describedObject.apiVersion"),
		Kind:       mapx.GetStr(ms, "describedObject.kind"),
		ResName:    mapx.GetStr(ms, "describedObject.name"),
		Type:       mapx.GetStr(ms, "target.type"),
		Value:      getMetricValue(ms),
		Selector:   genMetricSelector(ms),
	}
}

func genPodMetricItem(m map[string]interface{}) model.PodMetricItem {
	ms := m["pods"].(map[string]interface{})
	return model.PodMetricItem{
		Name:     mapx.GetStr(ms, "metric.name"),
		Type:     mapx.GetStr(ms, "target.type"),
		Value:    getMetricValue(ms),
		Selector: genMetricSelector(ms),
	}
}

// 通过 metricSource 获取 metric value
func getMetricValue(ms map[string]interface{}) string {
	switch mapx.GetStr(ms, "target.type") {
	case HPATargetTypeAverageValue:
		return mapx.GetStr(ms, "target.averageValue")
	case HPATargetTypeUtilization:
		return strconv.FormatInt(mapx.GetInt64(ms, "target.averageUtilization"), 10)
	case HPATargetTypeValue:
		return mapx.GetStr(ms, "target.value")
	default:
		return ""
	}
}

// 通过 metricSource 获取 metric selector
func genMetricSelector(ms map[string]interface{}) model.MetricSelector {
	selector := model.MetricSelector{}
	for _, exp := range mapx.GetList(ms, "metric.selector.matchExpressions") {
		e := exp.(map[string]interface{})
		values := []string{}
		for _, v := range mapx.GetList(e, "values") {
			values = append(values, v.(string))
		}
		selector.Expressions = append(selector.Expressions, model.ExpSelector{
			Key: e["key"].(string), Op: e["operator"].(string), Values: strings.Join(values, ","),
		})
	}
	for k, v := range mapx.GetMap(ms, "metric.selector.matchLabels") {
		selector.Labels = append(selector.Labels, model.LabelSelector{
			Key: k, Value: v.(string),
		})
	}
	return selector
}
