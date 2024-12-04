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
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
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
}

// Metadata defines what the structure of the metadata of a manifest file
type Metadata struct {
	Name   string `yaml:"name"`
	Labels map[string]interface{}
}

// GetManifestMetadata 获取 Manifest metadata
func GetManifestMetadata(manifest string) SimpleHead {
	var entry SimpleHead
	if err := yaml.Unmarshal([]byte(manifest), &entry); err != nil {
		return entry
	}
	return entry
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

// GetLablesFromManifest get labels from manifest
func GetLablesFromManifest(manifest string, kind, associateName string) map[string]interface{} {
	resp := make([]map[string]interface{}, 0)
	manifests := SplitManifests(manifest)
	for _, v := range manifests {
		metadata := GetManifestMetadata(v)
		if kind != metadata.Kind {
			continue
		}
		// 筛选关联资源名称列表
		if associateName == "" {
			resp = append(resp, map[string]interface{}{
				"label":    metadata.Metadata.Name,
				"value":    metadata.Metadata.Name,
				"disabled": false,
				"tips":     "",
			})
			continue
		}
		// 筛选关联应用的labels
		if associateName == metadata.Metadata.Name {
			for kk, vv := range metadata.Metadata.Labels {
				resp = append(resp, map[string]interface{}{
					"key":   kk,
					"value": vv,
				})
			}
		}
	}

	if associateName != "" {
		return map[string]interface{}{action.FormDataFormat: resp}
	}
	return map[string]interface{}{action.SelectItemsFormat: resp}
}
