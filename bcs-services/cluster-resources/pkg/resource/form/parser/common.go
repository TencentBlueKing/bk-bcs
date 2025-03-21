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
 */

// Package parser xxx
package parser

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/samber/lo"
	"gopkg.in/yaml.v2"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
)

const (
	// NameRegex 仅支持小写字母、数字、- 以及小写字母、数字组合
	NameRegex = `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
)

// GetResParseFunc 获取资源对应 Parser
// Parser 目前为函数实现，可以考虑抽成 interface，如 DeployParser，STSParser etc
func GetResParseFunc(
	ctx context.Context, kind string,
) (func(manifest map[string]interface{}) map[string]interface{}, error) {
	parseFunc, exists := Kind2ParseFuncMap[kind]
	if !exists {
		return nil, errorx.New(errcode.Unsupported, i18n.GetMsg(ctx, "资源类型 `%s` 不支持表单化"), kind)
	}
	return parseFunc, nil
}

var sep = regexp.MustCompile("(?:^|\\s*\n)---\\s*")

// SplitManifests takes a string of manifest and returns a map contains individual manifests
func SplitManifests(bigFile string) []string {
	// Basically, we're quickly splitting a stream of YAML documents into an
	// array of YAML docs. The file name is just a place holder, but should be
	// integer-sortable so that manifests get output in the same order as the
	// input (see `BySplitManifestsOrder`).
	res := make([]string, 0)
	// Making sure that any extra whitespace in YAML stream doesn't interfere in splitting documents correctly.
	bigFileTmp := strings.TrimSpace(bigFile)
	docs := sep.Split(bigFileTmp, -1)
	for _, d := range docs {
		if d == "" {
			continue
		}

		d = strings.TrimSpace(d)
		res = append(res, d)
	}
	return res
}

// SimpleHead defines what the structure of the head of a manifest file
type SimpleHead struct {
	Version  string   `yaml:"apiVersion"`
	Kind     string   `yaml:"kind,omitempty"`
	Metadata Metadata `yaml:"metadata"`
	Spec     Spec     `yaml:"spec"`
}

// Spec defines what the structure of the Spec of a manifest file
type Spec struct {
	Template Template `yaml:"template"`
}

// Template defines what the structure of the Template of a manifest file
type Template struct {
	Spec TemplateSpec `yaml:"spec"`
}

// TemplateSpec defines what the structure of the TemplateSpec of a manifest file
type TemplateSpec struct {
	Containers []Containers `yaml:"containers"`
}

// Containers defines what the structure of the Containers of a manifest file
type Containers struct {
	Ports []Ports `yaml:"ports"`
}

// Ports defines what the structure of the Ports of a manifest file
type Ports struct {
	Protocol      string `yaml:"protocol"`
	ContainerPort string `yaml:"containerPort"`
}

// Metadata defines what the structure of the metadata of a manifest file
type Metadata struct {
	Name   string `yaml:"name"`
	Labels map[string]interface{}
}

// GetManifestMetadata 获取 Manifest metadata
func GetManifestMetadata(manifest string) SimpleHead {
	var entry SimpleHead

	if strings.Contains(manifest, "{{") || strings.Contains(manifest, "}}") {
		manifest = FilterLines(manifest)
	}

	if err := yaml.Unmarshal([]byte(manifest), &entry); err != nil {
		return entry
	}
	return entry
}

// FilterLines 过滤文本中的行
func FilterLines(text string) string {
	// 按行分割文本
	lines := strings.Split(text, "\n")

	// 用于存储过滤后的行
	var filteredLines []string

	// 遍历每一行
	for _, line := range lines {
		// 如果行中不包含 {{ 和 }}，保留该行
		if !strings.Contains(line, "{{") && !strings.Contains(line, "}}") {
			filteredLines = append(filteredLines, line)
		}
	}

	// 将过滤后的行重新组合为字符串
	return strings.Join(filteredLines, "\n")
}

// GetResourceTypesFromManifest get resourceTypes from manifest
func GetResourceTypesFromManifest(manifest string) []string {
	resourceType := make([]string, 0)
	manifests := SplitManifests(manifest)
	for _, v := range manifests {
		metadata := GetManifestMetadata(v)
		if metadata.Kind != "" {
			resourceType = append(resourceType, metadata.Kind)
		}
	}
	return slice.RemoveDuplicateValues(resourceType)
}

// ParseResourcesFromManifest parse resources from manifest
func ParseResourcesFromManifest(versions []*entity.TemplateVersion, kind string) map[string]interface{} {
	resp := make([]map[string]interface{}, 0)
	for _, ver := range versions {
		manifests := SplitManifests(ver.Content)
		for _, v := range manifests {
			head := GetManifestMetadata(v)
			name := head.Metadata.Name
			if kind != head.Kind || !regexp.MustCompile(NameRegex).MatchString(name) {
				continue
			}

			resp = append(resp, map[string]interface{}{
				"label":    fmt.Sprintf("%s/%s", ver.TemplateName, name),
				"value":    fmt.Sprintf("%s/%s", ver.TemplateName, name),
				"disabled": false,
				"tips":     "",
			})
		}
	}

	return map[string]interface{}{action.SelectItemsFormat: resp}
}

// ParseLablesFromManifest parse labels from manifest
func ParseLablesFromManifest(versions []*entity.TemplateVersion, kind, name string) map[string]interface{} {
	resp := make([]map[string]interface{}, 0)
	for _, ver := range versions {
		manifests := SplitManifests(ver.Content)
		for _, v := range manifests {
			metadata := GetManifestMetadata(v)
			if kind != metadata.Kind {
				continue
			}
			if name != metadata.Metadata.Name {
				continue
			}

			// 筛选关联应用的labels
			for kk, vv := range metadata.Metadata.Labels {
				resp = append(resp, map[string]interface{}{
					"label":    kk,
					"value":    vv,
					"disabled": false,
					"tips":     "",
				})
			}
			break
		}
	}

	return map[string]interface{}{action.SelectItemsFormat: resp}
}

// ParsePortsFromManifest parse ports from manifest
func ParsePortsFromManifest(versions []*entity.TemplateVersion, kind, name, protocol string) map[string]interface{} {
	resp := make([]map[string]interface{}, 0)
	for _, ver := range versions {
		manifests := SplitManifests(ver.Content)
		for _, v := range manifests {
			metadata := GetManifestMetadata(v)
			if kind != metadata.Kind {
				continue
			}
			if name != metadata.Metadata.Name {
				continue
			}

			resp = append(resp, getPortsValue(protocol, metadata.Spec.Template.Spec.Containers)...)
		}
	}

	// 去重
	resp = lo.UniqBy(resp, func(item map[string]any) string {
		return item["label"].(string)
	})

	return map[string]interface{}{action.SelectItemsFormat: resp}
}

// 筛选关联应用的ports
func getPortsValue(protocol string, containers []Containers) []map[string]interface{} {
	if protocol != "TCP" && protocol != "UDP" {
		protocol = ""
	}
	resp := make([]map[string]interface{}, 0)
	for _, v := range containers {
		for _, port := range v.Ports {
			// 默认tcp
			if port.Protocol == "" {
				port.Protocol = "TCP"
			}
			// protocol没给默认返回所有端口
			if port.Protocol == protocol || protocol == "" {
				resp = append(resp, map[string]interface{}{
					"label":    port.ContainerPort,
					"value":    port.ContainerPort,
					"disabled": false,
					"tips":     "",
				})
			}
		}

	}
	return resp
}
