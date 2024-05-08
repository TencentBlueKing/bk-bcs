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

package utils

import (
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm/clause"
)

// Tabler xxx
type Tabler interface {
	Alias() string
	TableName() string
}

// Field xxx
type Field struct {
	Field field.Expr
	Table Tabler
}
type rawCond struct {
	field.Field
	sql  string
	args []interface{}
}

// BeCond xxx
func (m rawCond) BeCond() interface{} {
	args := []interface{}{}
	for _, v := range m.args {
		switch arg := v.(type) {
		case Field:
			column := clause.Column{
				Name: arg.Field.ColumnName().String(),
				Raw:  false,
			}
			if arg.Table != nil {
				if arg.Table.Alias() != "" {
					column.Table = arg.Table.Alias()
				} else {
					column.Table = arg.Table.TableName()
				}
			}
			args = append(args, column)
		case field.Expr:
			column := clause.Column{
				Name: arg.ColumnName().String(),
				Raw:  false,
			}
			args = append(args, column)
		default:
			args = append(args, v)
		}
	}

	expr := clause.NamedExpr{SQL: m.sql}
	expr.Vars = append(expr.Vars, args...)
	return expr
}

// CondError xxx
func (rawCond) CondError() error { return nil }

// RawCond 自定义sql语句，支持所有mysql逻辑运算符以及mysql函数
// RawCond("JSON_UNQUOTE(column_name) = xxx ")
// SELECT * FROM `table_name` JSON_UNQUOTE(column_name) = "xxx"
//
// RawCond("JSON_UNQUOTE(?) = ?", "column_name", "xxx")
// SELECT * FROM `table_name` JSON_UNQUOTE("column_name") = "xxx"
//
// RawCond("JSON_UNQUOTE(?) = ?", Field{Field: "column_name", Table: &q} , "xxx")
// SELECT * FROM `table_name` JSON_UNQUOTE("table_name"."column_name") = "xxx"
//
// RawCond("column_name != xxx ")
// SELECT * FROM `table_name` column_name != "xxx"
//
// RawCond("column_name in (?)",[]uint32{1,2,4})
// SELECT * FROM `table_name` column_name in (1,2,4)
//
// RawCond("column_name > ?", 3)
// SELECT * FROM `table_name` WHERE column_name > 3
//
// RawCond("column_name like ?", "%s%")
// SELECT * FROM `table_name` WHERE column_name like '%s%'
func RawCond(sql string, args ...interface{}) gen.Condition {
	return &rawCond{
		sql:  sql,
		args: args,
	}
}
