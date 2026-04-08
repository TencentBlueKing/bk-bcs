/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package generator provides utilities to parse rendered Helm YAML and generate
// DRPlan, DRWorkflow, and DRPlanExecution resources.
package generator

import (
	"sort"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Classify categorizes a list of Unstructured resources into hooks and main resources
// based on Helm hook annotations.
func Classify(resources []unstructured.Unstructured) ChartAnalysis {
	analysis := ChartAnalysis{
		Hooks: make(map[string][]HookResource),
	}

	for _, res := range resources {
		annotations := res.GetAnnotations()
		rawHookType, isHook := annotations[HookAnnotation]
		if !isHook {
			analysis.MainResources = append(analysis.MainResources, MainResource{Resource: res})
			continue
		}

		weight := parseWeight(annotations[HookWeightAnnotation])
		deletePolicy := annotations[HookDeletePolicy]
		hookTypes := splitHookTypes(rawHookType)
		added := false
		for _, hookType := range hookTypes {
			if hookType == HookTest || hookType == HookTestSuccess {
				continue
			}
			hook := HookResource{
				Resource:     res,
				HookType:     hookType,
				Weight:       weight,
				DeletePolicy: deletePolicy,
			}
			analysis.Hooks[hookType] = append(analysis.Hooks[hookType], hook)
			added = true
		}
		if !added {
			analysis.SkippedResources = append(analysis.SkippedResources, res)
		}
	}

	for hookType := range analysis.Hooks {
		sort.Slice(analysis.Hooks[hookType], func(i, j int) bool {
			return analysis.Hooks[hookType][i].Weight < analysis.Hooks[hookType][j].Weight
		})
	}

	return analysis
}

func splitHookTypes(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		result = append(result, p)
	}
	return result
}

func parseWeight(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	w, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return w
}
