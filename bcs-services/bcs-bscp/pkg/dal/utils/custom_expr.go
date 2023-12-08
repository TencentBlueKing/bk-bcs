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

// Package utils xxx
package utils

import (
	"database/sql"
	"database/sql/driver"
	"reflect"

	"gorm.io/gen/field"
	"gorm.io/gorm/clause"
)

// NewCustomExpr custom expression
func NewCustomExpr(sql string, value []interface{}) CustomExpr {
	return CustomExpr{
		SQL:                sql,
		Vars:               value,
		WithoutParentheses: false,
	}
}

// CustomExpr raw expression
type CustomExpr struct {
	SQL                string
	Vars               []interface{}
	WithoutParentheses bool
	field.Expr
}

// Build build raw expression
func (expr CustomExpr) Build(builder clause.Builder) {
	var (
		afterParenthesis bool
		idx              int
	)

	for _, v := range []byte(expr.SQL) {
		if v == '?' && len(expr.Vars) > idx {
			if afterParenthesis || expr.WithoutParentheses {
				if _, ok := expr.Vars[idx].(driver.Valuer); ok {
					builder.AddVar(builder, expr.Vars[idx])
				} else {
					switch rv := reflect.ValueOf(expr.Vars[idx]); rv.Kind() {
					case reflect.Slice, reflect.Array:
						if rv.Len() == 0 {
							builder.AddVar(builder, nil)
						} else {
							for i := 0; i < rv.Len(); i++ {
								if i > 0 {
									_ = builder.WriteByte(',')
								}
								builder.AddVar(builder, rv.Index(i).Interface())
							}
						}
					default:
						builder.AddVar(builder, expr.Vars[idx])
					}
				}
			} else {
				builder.AddVar(builder, expr.Vars[idx])
			}

			idx++
		} else {
			if v == '(' {
				afterParenthesis = true
			} else {
				afterParenthesis = false
			}
			_ = builder.WriteByte(v)
		}
	}

	if idx < len(expr.Vars) {
		for _, v := range expr.Vars[idx:] {
			builder.AddVar(builder, sql.NamedArg{Value: v})
		}
	}
}
