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

package operator

import (
	"context"
)

type MockTank struct {
	Value      []interface{}
	Length     int
	ChangeInfo *ChangeInfo
	Err        error
}

// implements type GetNewTank, return a mock function which will return the provided mock tank
func GetMockTankNewFunc(mt *MockTank) func() Tank {
	return func() Tank {
		return mt
	}
}

func (mt *MockTank) Close() {
}

func (mt *MockTank) GetValue() []interface{} {
	return mt.Value
}

func (mt *MockTank) GetLen() int {
	return mt.Length
}

func (mt *MockTank) GetChangeInfo() *ChangeInfo {
	return mt.ChangeInfo
}

func (mt *MockTank) GetError() error {
	return mt.Err
}

func (mt *MockTank) Databases() Tank {
	return mt
}

func (mt *MockTank) Using(name string) Tank {
	return mt
}

func (mt *MockTank) Tables() Tank {
	return mt
}

func (mt *MockTank) SetTableV(data interface{}) Tank {
	return mt
}

func (mt *MockTank) GetTableV() Tank {
	return mt
}

func (mt *MockTank) From(name string) Tank {
	return mt
}

func (mt *MockTank) Distinct(key string) Tank {
	return mt
}

func (mt *MockTank) OrderBy(key ...string) Tank {
	return mt
}

func (mt *MockTank) Select(key ...string) Tank {
	return mt
}

func (mt *MockTank) Offset(n int) Tank {
	return mt
}

func (mt *MockTank) Limit(n int) Tank {
	return mt
}

func (mt *MockTank) Index(key ...string) Tank {
	return mt
}

func (mt *MockTank) Filter(cond *Condition, args ...interface{}) Tank {
	return mt
}

func (mt *MockTank) Count() Tank {
	return mt
}

func (mt *MockTank) Query(args ...interface{}) Tank {
	return mt
}

func (mt *MockTank) Insert(data ...M) Tank {
	return mt
}

func (mt *MockTank) Upsert(data M, args ...interface{}) Tank {
	return mt
}

func (mt *MockTank) Update(data M, args ...interface{}) Tank {
	return mt
}

func (mt *MockTank) UpdateAll(data M, args ...interface{}) Tank {
	return mt
}

func (mt *MockTank) Remove(args ...interface{}) Tank {
	return mt
}

func (mt *MockTank) RemoveAll(args ...interface{}) Tank {
	return mt
}

func (mt *MockTank) Watch(opts *WatchOptions) (chan *Event, context.CancelFunc) {
	return nil, nil
}

// return the result of Condition Combine with mockLeafFunc and mockBranchFunc
func MockCombineCondition(c *Condition) interface{} {
	return c.Combine(mockLeafFunc, mockBranchFunc)
}

func mockLeafFunc(c *Condition) interface{} {
	switch c.Type {
	case Tr:
		return M{}
	case Eq:
		return c.Value
	default:
		return M{string(c.Type): c.Value}
	}
}

func mockBranchFunc(t ConditionType, cl []interface{}) interface{} {
	length := len(cl)
	if length == 0 {
		return nil
	}
	switch t {
	case And:
		if length == 1 {
			return cl[0]
		}
		return M{"and": cl}
	case Or:
		if length == 1 {
			return cl[0]
		}
		return M{"or": cl}
	case Not:
		return M{"not": cl[0]}
	}
	return nil
}
