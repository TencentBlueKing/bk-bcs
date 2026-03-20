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
	"bytes"
	"fmt"
	"text/template"

	"k8s.io/klog/v2"
)

// TemplateData holds data for template rendering
type TemplateData struct {
	Params   map[string]interface{}
	PlanName string
	Outputs  map[string]interface{}
}

// RenderTemplate renders a template string with the provided data
// Supports syntax: {{ .params.xxx }}, {{ .planName }}, {{ .outputs.xxx }}
func RenderTemplate(tmpl string, data *TemplateData) (string, error) {
	if data == nil {
		data = &TemplateData{
			Params:  make(map[string]interface{}),
			Outputs: make(map[string]interface{}),
		}
	}

	// Create template with custom functions
	t, err := template.New("template").Funcs(template.FuncMap{
		"default": func(defaultVal, val interface{}) interface{} {
			if val == nil || val == "" {
				return defaultVal
			}
			return val
		},
	}).Parse(tmpl)
	if err != nil {
		klog.V(4).Infof("Failed to parse template: %v, template: %s", err, tmpl)
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Render template
	var buf bytes.Buffer
	if err := t.Execute(&buf, map[string]interface{}{
		"params":   data.Params,
		"planName": data.PlanName,
		"outputs":  data.Outputs,
	}); err != nil {
		klog.V(4).Infof("Failed to execute template: %v, template: %s, data: %+v", err, tmpl, data)
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	result := buf.String()
	klog.V(4).Infof("Template rendered: %s -> %s", tmpl, result)
	return result, nil
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
