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

package custom

import (
	"strconv"

	"github.com/fatih/structs"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/common"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

// ParseHookTmpl xxx
func ParseHookTmpl(manifest map[string]interface{}) map[string]interface{} {
	tmpl := model.HookTmpl{}
	common.ParseMetadata(manifest, &tmpl.Metadata)
	ParseHookTmplSpec(manifest, &tmpl.Spec)
	return structs.Map(tmpl)
}

// ParseHookTmplSpec xxx
func ParseHookTmplSpec(manifest map[string]interface{}, spec *model.HookTmplSpec) {
	for _, arg := range mapx.GetList(manifest, "spec.args") {
		a := arg.(map[string]interface{})
		spec.Args = append(spec.Args, model.HookTmplArg{Key: a["name"].(string), Value: mapx.GetStr(a, "value")})
	}
	spec.ExecPolicy = mapx.Get(manifest, "spec.policy", resCsts.HookTmplPolicyParallel).(string)
	spec.DeletionProtectPolicy = mapx.Get(
		manifest,
		[]string{"metadata", "labels", resCsts.DeletionProtectLabelKey},
		resCsts.DeletionProtectPolicyNotAllow,
	).(string)
	for _, metric := range mapx.GetList(manifest, "spec.metrics") {
		spec.Metrics = append(spec.Metrics, genHookTmplMetric(metric.(map[string]interface{})))
	}
}

func genHookTmplMetric(raw map[string]interface{}) model.HookTmplMetric {
	// 表单创建的 interval 单位都是秒
	intervalStr, _ := stringx.Partition(mapx.Get(raw, "interval", "1s").(string), "s")
	interval, _ := strconv.Atoi(intervalStr)

	// 优先级 累计成功 > 连续成功
	successPolicy, successCnt := resCsts.HookTmplSuccessfulLimit, int64(0)
	if limit, ok := raw["successfulLimit"]; ok {
		successCnt = limit.(int64)
	} else if limit, ok = raw["consecutiveSuccessfulLimit"]; ok {
		successPolicy = resCsts.HookTmplConsecutiveSuccessfulLimit
		successCnt = limit.(int64)
	}

	metric := model.HookTmplMetric{
		Name:             mapx.GetStr(raw, "name"),
		Count:            mapx.GetInt64(raw, "count"),
		Interval:         interval,
		SuccessCondition: mapx.GetStr(raw, "successCondition"),
		SuccessPolicy:    successPolicy,
		SuccessCnt:       successCnt,
	}

	// provider 优先级 web > prometheus > kubernetes
	provider := raw["provider"].(map[string]interface{})
	if web, ok := provider["web"]; ok {
		// web 类型
		w := web.(map[string]interface{})
		metric.HookType = resCsts.HookTmplMetricTypeWeb
		metric.URL = mapx.GetStr(w, "url")
		metric.JSONPath = mapx.GetStr(w, "jsonPath")
		metric.TimeoutSecs = mapx.GetInt64(w, "timeoutSeconds")
	} else if prometheus, ok := provider["prometheus"]; ok {
		// prometheus 类型
		prom := prometheus.(map[string]interface{})
		metric.HookType = resCsts.HookTmplMetricTypeProm
		metric.Query = mapx.GetStr(prom, "query")
		metric.Address = mapx.GetStr(prom, "address")
	} else if kubernetes, ok := provider["kubernetes"]; ok {
		// kubernetes 类型
		k8s := kubernetes.(map[string]interface{})
		metric.HookType = resCsts.HookTmplMetricTypeK8S
		metric.Function = mapx.GetStr(k8s, "function")
		for _, field := range mapx.GetList(k8s, "fields") {
			f := field.(map[string]interface{})
			metric.Fields = append(metric.Fields, model.HookTmplField{
				Key: mapx.GetStr(f, "path"), Value: mapx.GetStr(f, "value"),
			})
		}
	}
	return metric
}
