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
 *
 */

package mongodb

import (
	"fmt"
	"regexp"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type search struct {
	tank       *mongoTank
	collection *mgo.Collection
	condition  *operator.Condition
	rawCond    bson.M

	orders   []string
	distinct string
	offset   int
	limit    int
	selector bson.M
}

func (s *search) clone() *search {
	ns := &search{
		tank:       s.tank,
		collection: s.collection,
		orders:     s.orders,
		distinct:   s.distinct,
		offset:     s.offset,
		limit:      s.limit,
		selector:   s.selector,
	}
	if s.condition == nil {
		ns.condition = operator.BaseCondition
	} else {
		ns.condition = s.condition
	}
	return ns
}

func (s *search) combineCondition(cond *operator.Condition) *search {
	if s.condition == operator.BaseCondition {
		s.condition = cond
	} else {
		s.condition = s.condition.And(cond)
	}
	return s
}

func (s *search) setDistinct(key string) *search {
	s.distinct = key
	return s
}

func (s *search) setOrder(key ...string) *search {
	s.orders = key
	return s
}

func (s *search) setSelector(key ...string) *search {
	tmp := make(bson.M)
	for _, v := range key {
		if v == "" {
			continue
		}
		tmp[v] = 1
	}
	if len(tmp) > 0 {
		s.selector = tmp
	}
	return s
}

func (s *search) setOffset(offset int) *search {
	s.offset = offset
	return s
}

func (s *search) setLimit(limit int) *search {
	s.limit = limit
	return s
}

func (s *search) getRawCond() bson.M {
	if s.rawCond != nil {
		return s.rawCond
	}
	raw := s.condition.Combine(
		leafNodeProcessor,
		branchNodeProcessor,
	)

	s.rawCond = raw.(bson.M)
	return s.rawCond
}

// Handle leaf node of Condition while combining
func leafNodeProcessor(cond *operator.Condition) (v interface{}) {
	switch cond.Type {
	case operator.Tr:
		v = bson.M{}
	case operator.Eq:
		originValue := cond.Value.(operator.M)
		v = bson.M(originValue)
	case operator.Ne:
		originValue := cond.Value.(operator.M)
		v = convertOrigin2BSon("$ne", originValue)
	case operator.Lt:
		originValue := cond.Value.(operator.M)
		v = convertOrigin2BSon("$lt", originValue)
	case operator.Lte:
		originValue := cond.Value.(operator.M)
		v = convertOrigin2BSon("$lte", originValue)
	case operator.Gt:
		originValue := cond.Value.(operator.M)
		v = convertOrigin2BSon("$gt", originValue)
	case operator.Gte:
		originValue := cond.Value.(operator.M)
		v = convertOrigin2BSon("$gte", originValue)
	case operator.In:
		originValue := cond.Value.(operator.M)
		v = convertOrigin2BSon("$in", originValue)
	case operator.Nin:
		originValue := cond.Value.(operator.M)
		v = convertOrigin2BSon("$nin", originValue)
	case operator.Con:
		originValue := cond.Value.(operator.M)
		v = convertContains2BSon(originValue)
	case operator.Ext:
		originValue := cond.Value.(operator.M)
		v = convertOrigin2BSon("$exists", originValue)
	default:
	}
	return
}

// Handle branch node of Condition while combining
func branchNodeProcessor(t operator.ConditionType, condList []interface{}) (v interface{}) {
	length := len(condList)
	if length == 0 {
		return nil
	}
	switch t {
	case operator.And:
		if length == 1 {
			return condList[0]
		}
		v = bson.M{"$and": condList}
	case operator.Or:
		if length == 1 {
			return condList[0]
		}
		v = bson.M{"$or": condList}
	case operator.Not:
		tmp := make(bson.M)
		cond := condList[0].(bson.M)
		for condK, condV := range cond {
			tmp[condK] = bson.M{"$not": condV}
		}
		v = tmp
	}
	return
}

// Convert operator.M of leafNode to bSon for mongodb
func convertOrigin2BSon(symbol string, originValue operator.M) bson.M {
	r := make(bson.M)
	for k, v := range originValue {
		r[k] = bson.M{symbol: v}
	}
	return r
}

// Handle the contains condition
func convertContains2BSon(originValue operator.M) bson.M {
	r := make(bson.M)
	for k, v := range originValue {
		if s, ok := v.(string); ok {
			r[k] = bson.RegEx{Pattern: fmt.Sprintf(".*%s.*", regexp.QuoteMeta(s))}
		}
	}
	return r
}
