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

package tspider

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/utils"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

// Public public store for tspider db
type Public struct {
	TableName string
	DB        *sqlx.DB
}

// QueryxToStructpb query data and return struct
func (p *Public) QueryxToStructpb(builder sq.SelectBuilder) ([]*structpb.Struct, error) {
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	if sql == "" {
		return nil, fmt.Errorf("sql should not be empty string")
	}

	response := make([]*structpb.Struct, 0)
	rows, err := p.DB.Queryx(sql, args...)
	if err != nil {
		blog.Errorf("query data error, sql:%s, args: %v, err:%s", sql, args, err.Error())
		return nil, fmt.Errorf("query data error, err: %s", err.Error())
	}
	for rows.Next() {
		r := make(map[string]interface{})
		if err := rows.MapScan(r); err != nil {
			blog.Errorf("map data to interface{} error, sql:%s, args: %v, err: %s", sql, args, err.Error())
			return nil, fmt.Errorf("map data to interface{} error, err: %s ", err.Error())
		}

		structData, err := structpb.NewStruct(utils.Bytes2String(r))
		if err != nil {
			return nil, err
		}
		response = append(response, structData)
	}
	return response, nil
}

// QueryxToStruct query data and return struct
func (p *Public) QueryxToAny(builder sq.SelectBuilder) ([]*any.Any, error) {
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	if sql == "" {
		return nil, fmt.Errorf("sql should not be empty string")
	}
	blog.Infof("sql: [%s], args: [%v]", sql, args)

	response := make([]*any.Any, 0)
	rows, err := p.DB.Queryx(sql, args...)
	if err != nil {
		blog.Errorf("query data error, sql:%s, args: %v, err:%s", sql, args, err.Error())
		return nil, fmt.Errorf("query data error, err: %s", err.Error())
	}
	for rows.Next() {
		r := make(map[string]interface{})
		if err := rows.MapScan(r); err != nil {
			blog.Errorf("map data to interface{} error, sql:%s, args: %v, err: %s", sql, args, err.Error())
			return nil, fmt.Errorf("map data to interface{} error, err: %s ", err.Error())
		}

		structData, err := structpb.NewStruct(utils.Bytes2String(r))
		if err != nil {
			return nil, err
		}

		anyData, err := anypb.New(structData)
		if err != nil {
			return nil, err
		}

		response = append(response, anyData)
	}
	return response, nil
}

// QueryxToAny query data and return map[string]interface{}
func (p *Public) QueryxToMap(builder sq.SelectBuilder) ([]map[string]interface{}, error) {
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	if sql == "" {
		return nil, fmt.Errorf("sql should not be empty string")
	}

	response := make([]map[string]interface{}, 0)
	rows, err := p.DB.Queryx(sql, args...)
	if err != nil {
		blog.Errorf("query data error, sql:%s, args: %v, err:%s", sql, args, err.Error())
		return nil, fmt.Errorf("query data error, err: %s", err.Error())
	}
	for rows.Next() {
		r := make(map[string]interface{})
		if err := rows.MapScan(r); err != nil {
			blog.Errorf("map data to interface{} error, sql:%s, args: %v, err: %s", sql, args, err.Error())
			return nil, fmt.Errorf("map data to interface{} error, err: %s ", err.Error())
		}

		response = append(response, utils.Bytes2String(r))
	}
	return response, nil
}

// QueryxToStruct query and map result into given obj
func (p *Public) QueryxToStruct(builder sq.SelectBuilder, obj interface{}) error {
	sql, args, err := builder.ToSql()
	if err != nil {
		return err
	}
	if sql == "" {
		return fmt.Errorf("sql should not be empty string")
	}

	if err := p.DB.Select(obj, sql, args...); err != nil {
		blog.Errorf("select data to obj error, sql: %s, args: %v, err: %s", sql, args, err.Error())
		return fmt.Errorf("select data to obj error, err: %s", err.Error())
	}
	return nil
}

// Countx count data and return total
func (p *Public) Countx(builder sq.SelectBuilder) (int, error) {
	sql, args, err := builder.ToSql()
	if err != nil {
		return 0, err
	}
	if sql == "" {
		return 0, fmt.Errorf("sql should not be empty string")
	}

	var count int
	if err := p.DB.QueryRowx(sql, args...).Scan(&count); err != nil {
		blog.Errorf("countx data error, sql:%s, args: %v, err:%s", sql, args, err.Error())
		return 0, fmt.Errorf("countx Error, err: %s", err.Error())
	}

	return count, nil
}

// GetMax get Max dtEventTimeStamp
func (p *Public) GetMax(table string, key string, value interface{}) error {
	builder := sq.Select(fmt.Sprintf("max(%s)", key)).
		From(table)

	sql, args, err := builder.ToSql()
	if err != nil {
		return err
	}
	if sql == "" {
		return fmt.Errorf("sql should not be empty string")
	}

	if err := p.DB.QueryRowx(sql, args...).Scan(value); err != nil {
		blog.Errorf("get max key(%s) error, sql:%s, args: %v, err:%s", key, sql, args, err.Error())
		return fmt.Errorf("get max key(%s) error, err: %s", key, err.Error())
	}
	return nil
}
