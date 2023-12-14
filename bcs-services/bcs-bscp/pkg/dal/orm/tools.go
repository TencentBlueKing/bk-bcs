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

package orm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// RearrangeSQLDataWithOption parse a *struct into a sql expression, and
// returned with the update sql expression and the to be updated data.
//  1. the input FieldOption only works for the returned 'expr', not controls
//     the returned 'toUpdate', so the returned 'toUpdate' contains all the
//     flatted tagged 'db' field and value.
//  2. Obviously, a data field need to be updated if the field value
//     is not blank(as is not "ZERO"),
//  3. If the field is defined in the blank options deliberately, then
//     update it to blank value as required.
//  4. see the test case to know the exact data returned.
func RearrangeSQLDataWithOption(data interface{}, opts *FieldOption) (
	expr string, toUpdate map[string]interface{}, err error) {

	if data == nil {
		return "", nil, errors.New("parse sql expr fields, but data is nil")
	}

	if opts == nil {
		return "", nil, errors.New("parse sql expr fields, but field options is nil")
	}

	var setFields []string
	toUpdate = make(map[string]interface{})
	taggedKV, err := RecursiveGetTaggedFieldValues(data)
	if err != nil {
		return "", nil, fmt.Errorf("get recursively tagged db kv faield, err: %v", err)
	}

	for tag, value := range taggedKV {
		// all the field's value is saved, no matter it's field need to be
		// blanked or ignored.
		toUpdate[tag] = value

		if opts.NeedIgnored(tag) {
			// this is a field which is need to be ignored,
			// which means do not need to be updated.
			continue
		}

		if !isBlank(reflect.ValueOf(value)) || opts.NeedBlanked(tag) {
			setFields = append(setFields, fmt.Sprintf("%s = :%s", tag, tag))
		}
	}

	expr = strings.Join(setFields, ", ")

	return expr, toUpdate, nil
}

// RecursiveGetTaggedFieldValues get all the tagged db kv
// in the struct to a flat map except ptr and struct tag.
// Note:
//  1. if the embedded tag is same, then it will be overlapped.
//  2. use this function carefully, it not supports all the type,
//     such as array, slice, map is not supported.
//  3. see the test case to know the output data example.
func RecursiveGetTaggedFieldValues(v interface{}) (map[string]interface{}, error) {
	if v == nil {
		return map[string]interface{}{}, nil
	}

	value := reflect.ValueOf(v)
	switch value.Kind() {
	case reflect.Ptr:
		if value.IsNil() {
			return map[string]interface{}{}, nil
		}

		return RecursiveGetTaggedFieldValues(value.Elem().Interface())

	case reflect.Struct:
		kv := make(map[string]interface{})

		for i := 0; i < value.NumField(); i++ {
			name := value.Type().Field(i).Name
			tag := value.Type().Field(i).Tag.Get("db")
			if tag == "" {
				return nil, fmt.Errorf("field: %s do not have a 'db' tag", name)
			}

			value := value.FieldByName(name).Interface()

			// this is a special treatment for scenarios where the entire
			// struct is treated as a field, such as strategy's scope,group's selector
			if tag == "scope" || tag == "selector" {
				kv[tag] = value
				continue
			}

			if isBasicValue(value) {
				// this is a basic value type
				kv[tag] = value

				// handle next field.
				continue
			}

			// this is not a basic value, then do get tags again recursively.
			mapper, err := RecursiveGetTaggedFieldValues(value)
			if err != nil {
				return nil, err
			}

			for k, v := range mapper {
				kv[k] = v
			}

		}

		return kv, nil

	default:
		return nil, fmt.Errorf("unsupported struct db tagged value type: %s", value.Kind())
	}
}

var timeType = reflect.TypeOf(time.Time{})

func isBasicValue(value interface{}) bool {
	v := reflect.ValueOf(value)
	if v.Type() == timeType {
		return true
	}

	return tools.IsBasicValue(value)
}

func isBlank(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return value.IsNil()
	}

	return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
}

// FieldOption is to define which field need to be:
//  1. updated to blank(as is ZERO) value.
//  2. be ignored, which means not be updated even its value
//     is not blank(as is not ZERO).
//
// NOTE:
// 1. A field can not in the blanked and ignore fields at the
// same time. if a field does, then it will be ignored without
// being updated.
// 2. The map's key is the structs' 'db' tag of that field.
type FieldOption struct {
	blanked map[string]struct{}
	ignored map[string]struct{}
}

// NewFieldOptions create a blank option instances for add keys
// to be updated when update data.
func NewFieldOptions() *FieldOption {
	return &FieldOption{
		blanked: make(map[string]struct{}),
		ignored: make(map[string]struct{}),
	}
}

// NeedBlanked check if this field need to be updated with blank
func (f *FieldOption) NeedBlanked(field string) bool {
	_, ok := f.blanked[field]
	return ok
}

// NeedIgnored check if this field does not need to be updated.
func (f *FieldOption) NeedIgnored(field string) bool {
	_, ok := f.ignored[field]
	return ok
}

// AddBlankedFields add fields to be updated to blank values.
func (f *FieldOption) AddBlankedFields(fields ...string) *FieldOption {
	for _, one := range fields {
		f.blanked[one] = struct{}{}
	}

	return f
}

// AddIgnoredFields add fields which do not need to be updated even it
// do has a value.
func (f *FieldOption) AddIgnoredFields(fields ...string) *FieldOption {
	for _, one := range fields {
		f.ignored[one] = struct{}{}
	}

	return f
}

// GetNamedSelectColumns get 'all' the named field with 'db' tag and
// embedded if it's an embedded struct.
// Use table.App as an example, the returned expr should as follows:
// id, biz_id, name as 'spec.name', memo as 'spec.memo'
// Note: define the embedded columns in the table.go manually.
// Deprecated: GetNamedSelectColumns will panic if there is a nil value.
func GetNamedSelectColumns(table interface{}) (expr string, err error) {
	if table == nil {
		return "", errors.New("invalid input table, is nil")
	}

	namedTags, err := recursiveGetNestedNamedTags(table)
	if err != nil {
		return "", err
	}

	// format the sql expr
	for k, v := range namedTags {
		if len(v) == 0 {
			expr = fmt.Sprintf("%s, %s", expr, k)
			continue
		}

		expr = fmt.Sprintf("%s, %s as '%s'", expr, k, v)
	}

	expr = strings.Trim(expr, ",")

	return expr, nil
}

func recursiveGetNestedNamedTags(table interface{}) (map[string]string, error) {

	value := reflect.ValueOf(table)
	switch value.Kind() {
	case reflect.Ptr:
		return recursiveGetNestedNamedTags(value.Elem().Interface())
	case reflect.Struct:

		nestedNamedTags := make(map[string]string)
		for i := 0; i < value.NumField(); i++ {
			name := value.Type().Field(i).Name
			tag := value.Type().Field(i).Tag.Get("db")
			if tag == "" {
				return nil, fmt.Errorf("field: %s do not have a 'db' tag", name)
			}

			value := value.FieldByName(name).Interface()
			if isBasicValue(value) {
				// this is a basic value type
				nestedNamedTags[tag] = ""

				// handle next field.
				continue
			}

			// this is not a basic value, then do get tags again recursively.
			nestedTags, err := recursiveGetNestedNamedTags(value)
			if err != nil {
				return nil, err
			}

			for k, v := range nestedTags {
				// add tag prefix with dot.
				if len(v) == 0 {
					// this k is a basic type
					nestedNamedTags[k] = tag + "." + k
				} else {
					// this k is a nested type
					nestedNamedTags[k] = tag + "." + v
				}

			}

		}

		return nestedNamedTags, nil

	default:
		return nil, fmt.Errorf("unsupported struct db tagged value type: %s", value.Kind())
	}
}

// RecursiveGetDBTags get a table's all the db tag recursively.
func RecursiveGetDBTags(table interface{}) ([]string, error) {

	value := reflect.ValueOf(table)
	switch value.Kind() {
	case reflect.Ptr:
		return RecursiveGetDBTags(value.Elem().Interface())
	case reflect.Struct:

		allTags := make([]string, 0)
		for i := 0; i < value.NumField(); i++ {
			name := value.Type().Field(i).Name
			tag := value.Type().Field(i).Tag.Get("db")
			if tag == "" {
				return nil, fmt.Errorf("field: %s do not have a 'db' tag", name)
			}

			value := value.FieldByName(name).Interface()
			if isBasicValue(value) {
				// this is a basic value type
				allTags = append(allTags, tag)

				// handle next field.
				continue
			}

			// this is not a basic value, then do get nested tags again recursively.
			nestedTags, err := RecursiveGetDBTags(value)
			if err != nil {
				return nil, err
			}

			allTags = append(allTags, nestedTags...)

		}

		return allTags, nil

	default:
		return nil, fmt.Errorf("unsupported struct db tagged value type: %s", value.Kind())
	}
}
