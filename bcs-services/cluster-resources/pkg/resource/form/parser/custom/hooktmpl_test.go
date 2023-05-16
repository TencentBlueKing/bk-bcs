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
	"testing"

	"github.com/stretchr/testify/assert"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
)

var lightHookTmplManifest = map[string]interface{}{
	"apiVersion": "tkex.tencent.com/v1alpha1",
	"kind":       "HookTemplate",
	"metadata": map[string]interface{}{
		"annotations": map[string]interface{}{
			resCsts.EditModeAnnoKey: "form",
		},
		"labels": map[string]interface{}{
			"io.tencent.bcs.dev/deletion-allow": "Always",
		},
		"name":      "hooktemplate-pk7ef5cv",
		"namespace": "default",
	},
	"spec": map[string]interface{}{
		"args": []interface{}{
			map[string]interface{}{
				"name":  "customArg1",
				"value": "value1",
			},
			map[string]interface{}{
				"name":  "customArg2",
				"value": "",
			},
		},
		"metrics": []interface{}{
			map[string]interface{}{
				"consecutiveSuccessfulLimit": int64(1),
				"count":                      int64(2),
				"interval":                   "1s",
				"name":                       "web",
				"provider": map[string]interface{}{
					"web": map[string]interface{}{
						"jsonPath":       "{$.result}",
						"timeoutSeconds": int64(10),
						"url":            "http://1.1.1.1:80",
					},
				},
				"successCondition": "asInt(result) == 1",
			},
			map[string]interface{}{
				"count":    int64(0),
				"interval": "1s",
				"name":     "prom",
				"provider": map[string]interface{}{
					"prometheus": map[string]interface{}{
						"address": "http://prometheus.com",
						"query":   "query_test",
					},
				},
				"successCondition": "asInt(result) == 2",
				"successfulLimit":  int64(1),
			},
			map[string]interface{}{
				"consecutiveSuccessfulLimit": int64(3),
				"count":                      int64(0),
				"interval":                   "2s",
				"name":                       "k8s",
				"provider": map[string]interface{}{
					"kubernetes": map[string]interface{}{
						"fields": []interface{}{
							map[string]interface{}{
								"path":  "metadata.name",
								"value": "resName-xx",
							},
						},
						"function": "patch",
					},
				},
			},
		},
		"policy": "Ordered",
	},
}

var exceptedHookTmplSpec = model.HookTmplSpec{
	Args: []model.HookTmplArg{
		{
			Key:   "customArg1",
			Value: "value1",
		},
		{
			Key:   "customArg2",
			Value: "",
		},
	},
	ExecPolicy:            "Ordered",
	DeletionProtectPolicy: resCsts.DeletionProtectPolicyAlways,
	Metrics: []model.HookTmplMetric{
		{
			Name:             "web",
			HookType:         resCsts.HookTmplMetricTypeWeb,
			URL:              "http://1.1.1.1:80",
			JSONPath:         "{$.result}",
			TimeoutSecs:      10,
			Count:            2,
			Interval:         1,
			SuccessCondition: "asInt(result) == 1",
			SuccessPolicy:    resCsts.HookTmplConsecutiveSuccessfulLimit,
			SuccessCnt:       1,
		},
		{
			Name:             "prom",
			HookType:         resCsts.HookTmplMetricTypeProm,
			Address:          "http://prometheus.com",
			Query:            "query_test",
			Count:            0,
			Interval:         1,
			SuccessCondition: "asInt(result) == 2",
			SuccessPolicy:    resCsts.HookTmplSuccessfulLimit,
			SuccessCnt:       1,
		},
		{
			Name:     "k8s",
			HookType: resCsts.HookTmplMetricTypeK8S,
			Function: "patch",
			Fields: []model.HookTmplField{
				{
					Key:   "metadata.name",
					Value: "resName-xx",
				},
			},
			Count:         0,
			Interval:      2,
			SuccessPolicy: resCsts.HookTmplConsecutiveSuccessfulLimit,
			SuccessCnt:    3,
		},
	},
}

func TestParseHookTmplSpec(t *testing.T) {
	actualHookTmplSpec := model.HookTmplSpec{}
	ParseHookTmplSpec(lightHookTmplManifest, &actualHookTmplSpec)
	assert.Equal(t, exceptedHookTmplSpec, actualHookTmplSpec)
}
