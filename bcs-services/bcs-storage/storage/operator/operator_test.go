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
	"reflect"
	"testing"
)

func TestMUpdate(t *testing.T) {
	m := M{
		"foo1": "bar1",
		"foo2": M{"foo3": "bar3"},
	}
	expect := M{
		"foo1": "bar1",
		"foo2": M{"foo3": "bar3"},
		"foo4": "bar4",
		"foo5": M{"foo6": 123},
	}

	if !reflect.DeepEqual(m.Update("foo4", "bar4").Update("foo5", M{"foo6": 123}), expect) {
		t.Errorf("M Update() do not work as expected! \nresult:\n%v\nexpect:\n%v\n", m, expect)
	}
}

func TestCondition(t *testing.T) {
	c0 := NewCondition(Tr, M(nil))
	if c0.list.Front().Value != c0 {
		t.Errorf("NewCondition() init list failed!")
	}

	c1 := c0.AddOp(Eq, "foo1", "bar1").AddOp(Lte, "foo2", 123).AddOp(In, "foo3", []bool{true, false, false})
	c1e := c1.list.Front().Next()
	if c2 := c1e.Value.(*Condition); c2.Type != Eq || c2.Value.(M)["foo1"] != "bar1" {
		t.Errorf("Condition AddOp() failed! \nresult:\n%v\nexpect:\n%v\n", *c2, Condition{Type: Eq, Value: M{"foo1": "bar"}})
	}
	if c3 := c1e.Next().Value.(*Condition); c3.Type != Lte || c3.Value.(M)["foo2"] != 123 {
		t.Errorf("Condition AddOp() failed! \nresult:\n%v\nexpect:\n%v\n", *c3, Condition{Type: Lte, Value: M{"foo2": 123}})
	}
	if c4 := c1e.Next().Next().Value.(*Condition); c4.Type != In || !reflect.DeepEqual(c4.Value.(M)["foo3"], []bool{true, false, false}) {
		t.Errorf("Condition AddOp() failed! \nresult:\n%v\nexpect:\n%v\n", *c4, Condition{Type: In, Value: M{"foo3": []bool{true, false, false}}})
	}

	c5 := BaseCondition.AddOp(Eq, "foo4", "bar4").AddOp(Ne, "foo5", "bar5")
	c6 := c1.And(c5)
	c7 := c6.Or(c0)
	c8 := c7.Not()
	if c8.Type != Not || c8.Value != c7 {
		t.Errorf("Condition Not() failed!")
	}
	if c7e := c7.Value.(*ConditionPair); c7.Type != Or || c7e.First != c6 || c7e.Second != c0 {
		t.Errorf("Condition Or() failed!")
	}
	if c6e := c6.Value.(*ConditionPair); c6.Type != And || c6e.First != c1 || c6e.Second != c5 {
		t.Errorf("Condition And() failed!")
	}

	rc := c8.Combine(mockLeafFunc, mockBranchFunc)
	expect := M{
		"not": M{
			"or": []interface{}{
				M{"and": []interface{}{
					M{"and": []interface{}{
						M{},
						M{"foo1": "bar1"},
						M{"lte": M{"foo2": 123}},
						M{"in": M{"foo3": []bool{true, false, false}}},
					}},
					M{"and": []interface{}{
						M{"foo4": "bar4"},
						M{"ne": M{"foo5": "bar5"}},
					}},
				}},
				M{},
			},
		},
	}
	if c, ok := rc.(M); !ok || !reflect.DeepEqual(c, expect) {
		t.Errorf("Condition Combine() failed! \nresult:\n%v\nexpect:\n%v\n", c, expect)
	}
}
