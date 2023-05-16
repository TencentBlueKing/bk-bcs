/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package tools

import (
	"encoding/json"
	"reflect"
)

// IsNumeric test if an interface is a numeric value or not.
func IsNumeric(val interface{}) bool {
	switch val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, json.Number:
		return true
	}
	return false
}

// IsBasicValue test if an interface is the basic supported
// golang type or not.
func IsBasicValue(value interface{}) bool {
	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64,
		reflect.String:
		return true
	default:
		return false
	}
}

// IsString test if an interface is the string type.
func IsString(value interface{}) bool {
	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.String:
		return true
	default:
		return false
	}
}

// IsBoolean test if an interface is the boolean type.
func IsBoolean(value interface{}) bool {
	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Bool:
		return true
	default:
		return false
	}
}
