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

// Operator condition type
type Operator string

const (
	// Tr Tree condition
	Tr Operator = "tr"

	// Query Selectors

	// ========== Comparison =============

	// Eq Matches values that are equal to a specified value
	Eq Operator = "eq"
	// Ne Matches all values that are not equal to a specified value
	Ne Operator = "ne"
	// Lt Matches values that are less than a specified value
	Lt Operator = "lt"
	// Lte Matches values that are less than or equal to a specified value
	Lte Operator = "lte"
	// Gt Matches values that are greater than a specified value
	Gt Operator = "gt"
	// Gte Matches values that are greater than or equal to a specified value
	Gte Operator = "gte"
	// In Matches any of the values specified in an array
	In Operator = "in"
	// Nin Matches none of the values specified in an array
	Nin Operator = "nin"
	// Con Matches values that contains some string
	Con Operator = "con"

	// ========== Logical =============

	// Not Inverts the effect of a query expression and returns documents that do not match the query expression
	Not Operator = "not"
	// And Joins query clauses with a logical AND returns all documents that match the conditions of both clauses
	And Operator = "and"
	// Or Joins query clauses with a logical OR returns all documents that match the conditions of either clause
	Or Operator = "or"
	// Nor Joins query clauses with a logical NOR returns all documents that fail to match both clauses
	Nor Operator = "nor"

	// ========== Element =============

	// Ext Matches documents that have the specified field
	Ext Operator = "exists"
	// Typ Selects documents if a field is of the specified type
	Typ Operator = "type"

	// ========== Pipeline in ChangeStream =========

	// Mat matches documents in change stream
	Mat Operator = "match"
)

var (
	// EmptyCondition just return {}
	EmptyCondition = NewLeafCondition(Tr, M{})
)

// M struct to store map data
type M map[string]interface{}

// Update update M
func (m M) Update(key string, value interface{}) M {
	m[key] = value
	return m
}

// Merge merge additional key-value pair into M
func (m M) Merge(additionM M) {
	for key, value := range additionM {
		m[key] = value
	}
}

// Condition the condition making is just the process of tree building
type Condition struct {
	Op       Operator
	Value    interface{}
	Children []*Condition
}

// NewLeafCondition create leaf condition
func NewLeafCondition(t Operator, v interface{}) *Condition {
	return &Condition{
		Op:    t,
		Value: v,
	}
}

// NewBranchCondition create branch condition
func NewBranchCondition(t Operator, cons ...*Condition) *Condition {
	newCondition := &Condition{
		Op: t,
	}
	newCondition.Children = append(newCondition.Children, cons...)
	return newCondition
}

// Combine combine
func (c *Condition) Combine(leafFunc func(Operator, interface{}) interface{},
	combineFunc func(Operator, []*Condition) interface{}) interface{} {
	// leaf node
	if len(c.Children) == 0 {
		return leafFunc(c.Op, c.Value)
	}
	return combineFunc(c.Op, c.Children)
}
