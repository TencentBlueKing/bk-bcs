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

package zookeeper

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"
)

type search struct {
	tank      *zkTank
	condition *operator.Condition
	rawCond   operator.M

	orders   []string
	distinct string
	offset   int
	limit    int
	selector []string
}

func (s *search) clone() *search {
	ns := &search{
		tank:     s.tank,
		orders:   s.orders,
		distinct: s.distinct,
		offset:   s.offset,
		limit:    s.limit,
		selector: s.selector,
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
	s.selector = key
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

func (s *search) getRawCond() operator.M {
	if s.rawCond == nil {
		return s.rawCond
	}
	raw := s.condition.Combine(
		leafNodeProcessor,
		branchNodeProcessor,
	)

	s.rawCond = raw.(operator.M)
	return s.rawCond
}

// Handle leaf node of Condition while combining
func leafNodeProcessor(cond *operator.Condition) (v interface{}) {
	switch cond.Type {
	case operator.Tr:
		v = operator.M{}
	case operator.Eq:
		v = operator.M{string(operator.Eq): cond.Value.(operator.M)}
	case operator.Ne:
		v = operator.M{string(operator.Ne): cond.Value.(operator.M)}
	case operator.Lt:
		v = operator.M{string(operator.Lt): cond.Value.(operator.M)}
	case operator.Lte:
		v = operator.M{string(operator.Lte): cond.Value.(operator.M)}
	case operator.Gt:
		v = operator.M{string(operator.Gt): cond.Value.(operator.M)}
	case operator.Gte:
		v = operator.M{string(operator.Gte): cond.Value.(operator.M)}
	case operator.In:
		v = operator.M{}
	case operator.Nin:
		v = operator.M{}
	case operator.Con:
		v = operator.M{string(operator.Con): cond.Value.(operator.M)}
	case operator.Ext:
		v = operator.M{}
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

	if t == operator.Or || t == operator.And {
		if length == 1 {
			return condList[0]
		}
		v = operator.M{string(t): condList}
	}
	if t == operator.Not {
		v = operator.M{string(t): condList[0]}
	}
	return
}
