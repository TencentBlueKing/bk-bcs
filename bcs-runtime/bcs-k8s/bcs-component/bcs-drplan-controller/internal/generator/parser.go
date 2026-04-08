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

package generator

import (
	"fmt"
	"io"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	sigyaml "sigs.k8s.io/yaml"
)

// ParseYAML reads multi-document YAML from r and returns a slice of
// Unstructured objects. Empty documents and comment-only documents are skipped.
func ParseYAML(r io.Reader) ([]unstructured.Unstructured, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	docs := splitYAMLDocuments(raw)

	var resources []unstructured.Unstructured
	for i, doc := range docs {
		trimmed := strings.TrimSpace(doc)
		if trimmed == "" || isCommentOnly(trimmed) {
			continue
		}

		jsonBytes, err := sigyaml.YAMLToJSON([]byte(trimmed))
		if err != nil {
			return nil, fmt.Errorf("document %d: converting YAML to JSON: %w", i, err)
		}

		if string(jsonBytes) == "null" {
			continue
		}

		var obj unstructured.Unstructured
		if err := obj.UnmarshalJSON(jsonBytes); err != nil {
			return nil, fmt.Errorf("document %d: unmarshalling: %w", i, err)
		}

		if obj.GetKind() == "" {
			continue
		}

		resources = append(resources, obj)
	}

	if len(resources) == 0 {
		return nil, fmt.Errorf("no valid Kubernetes resources found in input")
	}

	return resources, nil
}

// splitYAMLDocuments splits raw bytes by YAML document separator "---".
func splitYAMLDocuments(data []byte) []string {
	return strings.Split(string(data), "\n---")
}

func isCommentOnly(s string) bool {
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			return false
		}
	}
	return true
}
