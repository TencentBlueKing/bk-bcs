/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mongo

import (
	"fmt"
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
)

// Handle leaf node of Condition while combining
func leafNodeProcessor(op operator.Operator, value interface{}) interface{} {
	v := bson.M{}
	switch op {
	case operator.Tr:
		return v
	case operator.Eq:
		originValue := value.(operator.M)
		v = bson.M(originValue)
	case operator.Ne:
		originValue := value.(operator.M)
		v = convertOrigin2Bson("$ne", originValue)
	case operator.Lt:
		originValue := value.(operator.M)
		v = convertOrigin2Bson("$lt", originValue)
	case operator.Lte:
		originValue := value.(operator.M)
		v = convertOrigin2Bson("$lte", originValue)
	case operator.Gt:
		originValue := value.(operator.M)
		v = convertOrigin2Bson("$gt", originValue)
	case operator.Gte:
		originValue := value.(operator.M)
		v = convertOrigin2Bson("$gte", originValue)
	case operator.In:
		originValue := value.(operator.M)
		v = convertOrigin2Bson("$in", originValue)
	case operator.Nin:
		originValue := value.(operator.M)
		v = convertOrigin2Bson("$nin", originValue)
	case operator.Con:
		originValue := value.(operator.M)
		v = convertContains2Bson(originValue)
	case operator.Typ:
		originValue := value.(operator.M)
		v = convertOrigin2Bson("$type", originValue)
	case operator.Ext:
		originValue := value.(operator.M)
		v = convertOrigin2Bson("$exists", originValue)
	default:
		return nil
	}
	return v
}

// handle branch node of Condition while combining
func branchNodeProcessor(op operator.Operator, cons []*operator.Condition) interface{} {
	if len(cons) == 0 {
		return nil
	}
	var v bson.M
	switch op {
	case operator.And:
		var condList []interface{}
		for _, c := range cons {
			conRet := c.Combine(leafNodeProcessor, branchNodeProcessor)
			condList = append(condList, conRet)
		}
		v = bson.M{"$and": condList}
	case operator.Or:
		var condList []interface{}
		for _, c := range cons {
			conRet := c.Combine(leafNodeProcessor, branchNodeProcessor)
			condList = append(condList, conRet)
		}
		v = bson.M{"$or": condList}
	case operator.Nor:
		var condList []interface{}
		for _, c := range cons {
			conRet := c.Combine(leafNodeProcessor, branchNodeProcessor)
			condList = append(condList, conRet)
		}
		v = bson.M{"$nor": condList}
	case operator.Not:
		conRet := cons[0].Combine(leafNodeProcessor, branchNodeProcessor)
		v = bson.M{"$not": conRet}
	case operator.Mat:
		conRet := cons[0].Combine(leafNodeProcessor, branchNodeProcessor)
		v = bson.M{"$match": conRet}
	}
	return v
}

// Convert drivers.M of leafNode to bSon for mongodb
func convertOrigin2Bson(symbol string, originValue operator.M) bson.M {
	r := make(bson.M)
	for k, v := range originValue {
		r[k] = bson.M{symbol: v}
	}
	return r
}

// Handle the contains condition
func convertContains2Bson(originValue operator.M) bson.M {
	r := make(bson.M)
	for k, v := range originValue {
		if s, ok := v.(string); ok {
			r[k] = primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", regexp.QuoteMeta(s))}
			continue
		}
		// support primitive.Regex
		if s, ok := v.(primitive.Regex); ok {
			r[k] = s
		}
	}
	return r
}
