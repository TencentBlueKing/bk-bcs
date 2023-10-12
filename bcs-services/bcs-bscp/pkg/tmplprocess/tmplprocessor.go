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

// Package tmplprocess is used for template process
package tmplprocess

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var re = regexp.MustCompile(`{{\s*\.(?i)(bk_bscp_[A-Za-z0-9_]*)\s*}}`)

// TmplProcessor is the interface for search
type TmplProcessor interface {
	ExtractVariables(template []byte) []string
	Render(template []byte, variablesKV map[string]interface{}) []byte
}

// processor implements the TmplProcessor interface
type processor struct{}

// NewTmplProcessor new a TmplProcessor
func NewTmplProcessor() TmplProcessor {
	return &processor{}
}

// ExtractVariables extracts variables from template
func (p *processor) ExtractVariables(template []byte) []string {
	if len(template) == 0 {
		return []string{}
	}

	matches := re.FindAllStringSubmatch(string(template), -1)

	nameMap := make(map[string]struct{})
	for _, match := range matches {
		nameMap[match[1]] = struct{}{}
	}

	varNames := make([]string, 0, len(nameMap))
	for name := range nameMap {
		varNames = append(varNames, name)
	}

	// Sort in ascending order
	sort.Strings(varNames)

	return varNames
}

// Render renders template with variables key value map
func (p *processor) Render(template []byte, variablesKV map[string]interface{}) []byte {
	ret := re.ReplaceAllStringFunc(string(template), func(match string) string {
		varName := strings.TrimPrefix(match, "{{")
		varName = strings.TrimSuffix(varName, "}}")
		varName = strings.TrimSpace(varName)
		varName = strings.TrimPrefix(varName, ".")

		value, ok := variablesKV[varName]
		if !ok {
			return ""
		}
		return fmt.Sprint(value)
	})
	return []byte(ret)
}
