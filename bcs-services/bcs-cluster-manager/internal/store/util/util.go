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

// Package util xxx
package util

import (
	"context"
	"errors"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	// ErrDecryptCloudCredential decrypt cloud error
	ErrDecryptCloudCredential = errors.New("decrypt credential error")
)

const (
	// DataTableNamePrefix is prefix of data table name
	DataTableNamePrefix = "bcsclustermanagerv2_"

	// DefaultLimit table default limit
	DefaultLimit = 5000
)

// EnsureTable ensure object database table and table indexes
func EnsureTable(ctx context.Context, db drivers.DB, tableName string, indexes []drivers.Index) error {
	hasTable, err := db.HasTable(ctx, tableName)
	if err != nil {
		return err
	}
	if !hasTable {
		tErr := db.CreateTable(ctx, tableName)
		if tErr != nil {
			return tErr
		}
	}
	// only ensure index when index name is not empty
	for _, idx := range indexes {
		hasIndex, iErr := db.Table(tableName).HasIndex(ctx, idx.Name)
		if iErr != nil {
			return iErr
		}
		if !hasIndex {
			if iErr = db.Table(tableName).CreateIndex(ctx, idx); iErr != nil {
				return iErr
			}
		}
	}
	return nil
}

// SliceInterface2String to string slice
func SliceInterface2String(interfaceSlice []interface{}) []string {
	stringSlice := make([]string, 0)

	for _, v := range interfaceSlice {
		str, ok := v.(string)
		if !ok {
			continue
		}
		stringSlice = append(stringSlice, str)
	}

	return stringSlice
}

// MapInt2MapIf convert map[string]int to map[string]interface{}
func MapInt2MapIf(m map[string]int) map[string]interface{} {
	newM := make(map[string]interface{})
	for k, v := range m {
		newM[k] = v
	}
	return newM
}

const (
	// Regex xxx
	Regex = "regex"
	// Range xxx
	Range = "range"
)

// Condition xxx
func Condition(ope operator.Operator, src string, values []string) bson.E { // nolint
	if len(values) == 0 {
		return bson.E{}
	}

	switch ope {
	case operator.Eq:
		return bson.E{
			Key: src,
			Value: bson.M{
				"$eq": values[0],
			},
		}
	case operator.Ne:
		return bson.E{
			Key: src,
			Value: bson.M{
				"$ne": values[0],
			},
		}
	case Regex:
		return bson.E{
			Key: src,
			Value: bson.M{
				"$regex": values[0],
			},
		}
	case Range:
		if len(values) <= 1 {
			return bson.E{}
		}
		return bson.E{
			Key: src,
			Value: bson.M{
				"$gte": values[0],
				"$lte": values[1],
			},
		}
	case operator.Lte:
		return bson.E{
			Key: src,
			Value: bson.M{
				"$lte": values[0],
			},
		}
	case operator.Gte:
		return bson.E{
			Key: src,
			Value: bson.M{
				"$gte": values[0],
			},
		}
	case operator.In:
		return bson.E{
			Key: src,
			Value: bson.M{
				"$in": values,
			},
		}
	}

	return bson.E{}
}

// UnionTable body
type UnionTable struct {
	DstTable   string
	FromFields string
	DstFields  string
	AsField    string
}

// BuildLookUpCond build lookUp cond
func BuildLookUpCond(t UnionTable) map[string]interface{} {
	return map[string]interface{}{
		"$lookup": BuildLookUpValue(t),
	}
}

// BuildLookUpValue build lookup value
func BuildLookUpValue(table UnionTable) map[string]interface{} {
	return map[string]interface{}{
		"from":         table.DstTable,
		"localField":   table.FromFields,
		"foreignField": table.DstFields,
		"as":           table.AsField,
	}
}

// BuildUnWindCond build unWind cond
func BuildUnWindCond(asField string) map[string]interface{} {
	return map[string]interface{}{
		"$unwind": map[string]interface{}{
			"path":                       "$" + asField,
			"preserveNullAndEmptyArrays": true,
		},
	}
}

// BuildMatchExprCond build match/expr cond
func BuildMatchExprCond(cond interface{}) map[string]interface{} {
	return map[string]interface{}{
		"$match": map[string]interface{}{
			"expr": cond,
		},
	}
}

// BuildMatchCond build match cond
func BuildMatchCond(cond map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"$match": cond,
	}
}

// TransBsonEToMap trans to map interface
func TransBsonEToMap(condE []bson.E) map[string]interface{} {
	condM := make(map[string]interface{}, 0)
	for i := range condE {
		condM[condE[i].Key] = condE[i].Value
	}

	return condM
}

// BuildAndManyConds and conditions
func BuildAndManyConds(conds []bson.E) map[string]interface{} {
	return map[string]interface{}{
		"$and": bson.D(conds),
	}
}

// BuildProjectOutput build union table output
func BuildProjectOutput(project interface{}) map[string]interface{} {
	return map[string]interface{}{
		"$project": project,
	}
}

// BuildTaskOperationLogProject build task operation log
func BuildTaskOperationLogProject() map[string]interface{} {
	return map[string]interface{}{
		"resourcetype": "$resourcetype",
		"resourceid":   "$resourceid",
		"resourcename": "$resourcename",
		"taskid":       "$taskid",
		"message":      "$message",
		"opuser":       "$opuser",
		"createtime":   "$createtime",
		"tasktype":     "$task.tasktype",
		"status":       "$task.status",
		"clusterid":    "$clusterid",
		"projectid":    "$projectid",
		"nodeiplist":   "$task.nodeiplist",
	}
}

// TransStrToUTCStr trans time string to utc RFC3339 time string
func TransStrToUTCStr(timeType, input string) string {
	switch timeType {
	case time.RFC3339Nano:
		t, err := time.Parse(time.RFC3339Nano, input)
		if err != nil {
			return input
		}
		return t.UTC().Format(time.RFC3339)
	case time.DateTime:
		t, err := time.Parse(time.DateTime, input)
		if err != nil {
			return input
		}
		return t.UTC().Format(time.RFC3339)
	default:
		return input
	}
}
