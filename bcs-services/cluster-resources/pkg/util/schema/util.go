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

package schema

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strings"
)

func isKind(what interface{}, kinds ...reflect.Kind) bool {
	target := what
	if isJSONNumber(what) {
		// JSON Numbers are strings!
		target = *mustBeNumber(what)
	}
	targetKind := reflect.ValueOf(target).Kind()
	for _, kind := range kinds {
		if targetKind == kind {
			return true
		}
	}
	return false
}

func isJSONNumber(what interface{}) bool {
	_, ok := what.(json.Number)
	return ok
}

func checkJSONInteger(what interface{}) (isInt bool) {
	jsonNumber, ok := what.(json.Number)
	if !ok {
		return false
	}
	bigFloat, isValidNumber := new(big.Rat).SetString(string(jsonNumber))
	return isValidNumber && bigFloat.IsInt()
}

func mustBeInteger(what interface{}) *int {
	if isJSONNumber(what) {
		number := what.(json.Number)
		isInt := checkJSONInteger(number)

		if isInt {
			int64Value, err := number.Int64()
			if err != nil {
				return nil
			}

			int32Value := int(int64Value)
			return &int32Value
		}
	}
	return nil
}

func mustBeNumber(what interface{}) *big.Rat {
	if isJSONNumber(what) {
		number := what.(json.Number)
		if float64Value, success := new(big.Rat).SetString(string(number)); success {
			return float64Value
		}
	}
	return nil
}

// 根据继承关系，生成指定节点的路径，形如 .properties.name.type
func genNodePaths(schema *subSchema, subPath string) string {
	// 检查确保 subPath 以 . 开头
	if !strings.HasPrefix(subPath, ".") {
		subPath = "." + subPath
	}

	paths := []string{subPath}
	for schema != nil && schema.Source != SchemaSourceRoot {
		switch schema.Source {
		case SchemaSourceItems:
			paths = append(paths, fmt.Sprintf(".%s", SchemaSourceItems))
		case SchemaSourceProperties:
			paths = append(paths, fmt.Sprintf(".%s.%s", SchemaSourceProperties, schema.Property))
		}
		schema = schema.Parent
	}
	var b strings.Builder
	for idx := len(paths) - 1; idx >= 0; idx-- {
		b.WriteString(paths[idx])
	}
	return strings.Trim(b.String(), ".")
}

// 生成 subPath
func genSubPath(prefix, key string) string {
	if strings.Contains(key, ".") {
		key = "(" + key + ")"
	}
	ret := fmt.Sprintf("%s.%s", prefix, key)
	if !strings.HasPrefix(ret, ".") {
		ret = "." + ret
	}
	return ret
}

// 生成带下标的 subPath
func genSubPathWithIdx(prefix, key string, idx int) string {
	return fmt.Sprintf("%s[%d]", genSubPath(prefix, key), idx)
}
