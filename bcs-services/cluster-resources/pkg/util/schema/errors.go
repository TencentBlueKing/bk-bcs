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
	"errors"
	"fmt"
)

// NewSchemaInvalidErr Schema 格式不合法
func NewSchemaInvalidErr() error {
	return errors.New("expected: valid schema, given: Invalid JSON")
}

// NewInvalidTypeErr 类型不合法
func NewInvalidTypeErr(schema *subSchema, subPath string, except string) error {
	return fmt.Errorf("Paths: %s has invalid type, excepted: %s", genNodePaths(schema, subPath), except)
}

// NewRequiredErr 字段必须存在
func NewRequiredErr(schema *subSchema, subPath string) error {
	return fmt.Errorf("Paths: %s is required", genNodePaths(schema, subPath))
}

// NewNotAValidTypeErr 非 Schema 支持的类型
func NewNotAValidTypeErr(schema *subSchema, subPath string, _type string) error {
	return fmt.Errorf(
		"Paths: %s has a type that is NOT VALID -- given: %s, expected: %v",
		genNodePaths(schema, subPath), _type, SchemaTypes,
	)
}

// NewNotAValidCompErr 非 Schema 支持的组件
func NewNotAValidCompErr(schema *subSchema, subPath string, compName string) error {
	return fmt.Errorf(
		"Paths: %s has a compoent that is NOT VALID -- given: %s, expected: %v",
		genNodePaths(schema, subPath), compName, SchemaComps,
	)
}

// NewMustBeGTEZeroErr 值必须大于等于 0
func NewMustBeGTEZeroErr(schema *subSchema, subPath string) error {
	return fmt.Errorf("Paths: %s must be greater than or equal to 0", genNodePaths(schema, subPath))
}

// NewMustBeGTEOneErr 值必须大于等于 1
func NewMustBeGTEOneErr(schema *subSchema, subPath string) error {
	return fmt.Errorf("Paths: %s must be greater than or equal to 1", genNodePaths(schema, subPath))
}

// NewMustBeOfAErr 值必须为某类型
func NewMustBeOfAErr(schema *subSchema, subPath string, _type string) error {
	return fmt.Errorf("Paths: %s must be of a %s", genNodePaths(schema, subPath), _type)
}

// NewMustBeOfAnErr 值必须为某类型
func NewMustBeOfAnErr(schema *subSchema, subPath string, _type string) error {
	return fmt.Errorf("Paths: %s must be of an %s", genNodePaths(schema, subPath), _type)
}

// NewEmptyMapErr 键存在，但是值为空 Map
func NewEmptyMapErr(schema *subSchema, subPath string) error {
	return fmt.Errorf("Paths: %s can't be empty map", genNodePaths(schema, subPath))
}
