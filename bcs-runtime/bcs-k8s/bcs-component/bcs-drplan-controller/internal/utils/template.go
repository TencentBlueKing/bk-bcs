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

// Package utils provides utility functions for template rendering and retry logic
package utils

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/klog/v2"
)

// paramPattern matches $(params.xxx), $(planName), $(outputs.xxx.yyy) etc.
var paramPattern = regexp.MustCompile(`\$\(([a-zA-Z][a-zA-Z0-9]*(?:\.[a-zA-Z][a-zA-Z0-9]*)*)\)`)

// legacyParamPattern matches legacy Go-template-like placeholders such as
// {{ .params.xxx }}, {{ .planName }}, {{ .outputs.step.phase }}.
// Only the documented DR placeholders are supported; unrelated Helm templates
// like {{ .Release.Name }} are left untouched.
var legacyParamPattern = regexp.MustCompile(
	`\{\{\s*\.(planName|params(?:\.[a-zA-Z][a-zA-Z0-9]*)+|outputs(?:\.[a-zA-Z][a-zA-Z0-9]*)+)\s*\}\}`,
)

// TemplateData holds data for template rendering
type TemplateData struct {
	Params   map[string]interface{}
	PlanName string
	Outputs  map[string]interface{}
}

// RenderTemplate renders a template string with the provided data.
// Recommended syntax: $(params.xxx), $(planName), $(outputs.xxx).
// Legacy syntax {{ .params.xxx }} is still supported for backward compatibility.
func RenderTemplate(tmpl string, data *TemplateData) (string, error) {
	if data == nil {
		data = &TemplateData{
			Params:  make(map[string]interface{}),
			Outputs: make(map[string]interface{}),
		}
	}

	lookup := map[string]interface{}{
		"params":   data.Params,
		"planName": data.PlanName,
		"outputs":  data.Outputs,
	}

	result, err := renderWithPattern(tmpl, paramPattern, func(match string) string {
		return match[2 : len(match)-1] // strip "$(" and ")"
	}, lookup)
	if err != nil {
		klog.V(4).Infof("Failed to render template: %v, template: %s, data: %+v", err, tmpl, data)
		return "", err
	}

	result, err = renderWithPattern(result, legacyParamPattern, func(match string) string {
		submatches := legacyParamPattern.FindStringSubmatch(match)
		if len(submatches) != 2 {
			return ""
		}
		return submatches[1]
	}, lookup)
	if err != nil {
		klog.V(4).Infof("Failed to render legacy template: %v, template: %s, data: %+v", err, tmpl, data)
		return "", err
	}

	klog.V(4).Infof("Template rendered: %s -> %s", tmpl, result)
	return result, nil
}

func renderWithPattern(
	tmpl string,
	pattern *regexp.Regexp,
	pathExtractor func(string) string,
	lookup map[string]interface{},
) (string, error) {
	var renderErr error
	result := pattern.ReplaceAllStringFunc(tmpl, func(match string) string {
		path := pathExtractor(match)
		if path == "" {
			return match
		}
		val, err := resolveVarPath(path, lookup)
		if err != nil {
			renderErr = fmt.Errorf("failed to resolve %q: %w", match, err)
			return match
		}
		return fmt.Sprintf("%v", val)
	})
	if renderErr != nil {
		return "", renderErr
	}
	return result, nil
}

// resolveVarPath resolves a dot-separated path like "params.feedNamespace" against a nested map.
func resolveVarPath(path string, root map[string]interface{}) (interface{}, error) {
	parts := strings.SplitN(path, ".", 2) //nolint:mnd // split into key and remainder
	val, ok := root[parts[0]]
	if !ok {
		return nil, fmt.Errorf("key %q not found", parts[0])
	}
	if len(parts) == 1 {
		return val, nil
	}
	nested, ok := val.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("key %q is not a map, cannot resolve %q", parts[0], parts[1])
	}
	return resolveVarPath(parts[1], nested)
}

// RenderTemplateMap renders all string values in a map
func RenderTemplateMap(m map[string]string, data *TemplateData) (map[string]string, error) {
	result := make(map[string]string, len(m))
	for k, v := range m {
		rendered, err := RenderTemplate(v, data)
		if err != nil {
			return nil, fmt.Errorf("failed to render key %s: %w", k, err)
		}
		result[k] = rendered
	}
	return result, nil
}

// BuildParamsMap builds a parameter map from workflow parameters, global parameters, and stage parameters
// Priority: stageParams > globalParams > workflowDefaults
func BuildParamsMap(workflowDefaults, globalParams, stageParams map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Start with workflow defaults
	for k, v := range workflowDefaults {
		result[k] = v
	}

	// Override with global params
	for k, v := range globalParams {
		result[k] = v
	}

	// Override with stage params (highest priority)
	for k, v := range stageParams {
		result[k] = v
	}

	klog.V(4).Infof("Built params map: %+v", result)
	return result
}
