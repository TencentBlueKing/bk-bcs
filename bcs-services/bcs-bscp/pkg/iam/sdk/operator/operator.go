/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package operator NOTES
package operator

import (
	"errors"
	"reflect"
	"strings"
)

// factory defines all the supported operators
var factory map[string]Operator

func init() {
	factory = make(map[string]Operator)

	equal := EqualOp("")
	factory[equal.Name()] = &equal

	notEqual := NotEqualOp("")
	factory[notEqual.Name()] = &notEqual

	in := InOp("")
	factory[in.Name()] = &in

	notIn := NotInOp("")
	factory[notIn.Name()] = &notIn

	contains := ContainsOp("")
	factory[contains.Name()] = &contains

	notContains := NotContainsOp("")
	factory[notContains.Name()] = &notContains

	startWith := StartsWithOp("")
	factory[startWith.Name()] = &startWith

	notStartWith := NotStartsWithOp("")
	factory[notStartWith.Name()] = &notStartWith

	endWith := EndsWithOp("")
	factory[endWith.Name()] = &endWith

	notEndWith := NotEndsWithOp("")
	factory[notEndWith.Name()] = &notEndWith

	lessThan := LessThanOp("")
	factory[lessThan.Name()] = &lessThan

	lessThanEqual := LessThanEqualOp("")
	factory[lessThanEqual.Name()] = &lessThanEqual

	greaterThan := GreaterThanOp("")
	factory[greaterThan.Name()] = &greaterThan

	greaterThanEqual := GreaterThanEqualOp("")
	factory[greaterThanEqual.Name()] = &greaterThanEqual

	any := AnyOp("")
	factory[any.Name()] = &any

}

// Operator defines all the operators required operations.
type Operator interface {
	// Name of the operator
	Name() string

	// Match is used to check if "match" is "logical equal" to the "with"
	// with different OpType, different OpType has different definition
	// of "logical equal", if "logical equal" then return bool "true" value.
	//
	// match: the value to test
	// with: the value to compare to, which is also the template
	Match(match interface{}, with interface{}) (bool, error)
}

const (
	// Unknown is Unknown operator
	Unknown = "unknown"
	// Equal is Equal operator
	Equal = "eq"
	// NEqual is NEqual operator
	NEqual = "not_eq"
	// Any is Any operator
	Any = "any"
	// In is In operator
	In = "in"
	// Nin is Nin operator
	Nin = "not_in"
	// Contains is Contains operator
	Contains = "contains"
	// NContains is NContains operator
	NContains = "not_contains"
	// StartWith is StartWith operator
	StartWith = "starts_with"
	// NStartWith is NStartWith operator
	NStartWith = "not_starts_with"
	// EndWith is EndWith operator
	EndWith = "ends_with"
	// NEndWith is NEndWith operator
	NEndWith = "not_ends_with"
	// LessThan is LessThan operator
	LessThan = "lt"
	// LessThanEqual is LessThanEqual operator
	LessThanEqual = "lte"
	// GreaterThan is GreaterThan operator
	GreaterThan = "gt"
	// GreaterThanEqual is GreaterThanEqual operator
	GreaterThanEqual = "gte"
)

// OpType defines the operator types.
type OpType string

const (
	// And operator
	And = OpType("AND")
	// Or operator
	Or = OpType("OR")
)

// Operator returns one operator's Operator interface.
func (o *OpType) Operator() Operator {
	if o == nil {
		unknown := UnknownOp("")
		return &unknown
	}

	op, support := factory[string(*o)]
	if !support {
		unknown := UnknownOp("")
		return &unknown
	}

	return op
}

// UnknownOp is the unknown or unsupported operator
type UnknownOp OpType

// Name return the name of UnknownOp
func (u *UnknownOp) Name() string {
	return Unknown
}

// Match the unknown operator
func (u *UnknownOp) Match(_ interface{}, _ interface{}) (bool, error) {
	return false, errors.New("unknown type, can not do match")
}

// EqualOp is equal type
type EqualOp OpType

// Name return the name of EqualOp
func (e *EqualOp) Name() string {
	return Equal
}

// Match the equal operator
func (e *EqualOp) Match(match interface{}, with interface{}) (bool, error) {
	mType := reflect.TypeOf(match)
	wType := reflect.TypeOf(with)
	if mType.Kind() != wType.Kind() {
		return false, errors.New("mismatch type")
	}

	return reflect.DeepEqual(match, with), nil
}

// NotEqualOp is the not equal operator.
type NotEqualOp OpType

// Name return the name of NotEqualOp
func (e *NotEqualOp) Name() string {
	return NEqual
}

// Match the not equal operator
func (e *NotEqualOp) Match(match interface{}, with interface{}) (bool, error) {
	mType := reflect.TypeOf(match)
	wType := reflect.TypeOf(with)
	if mType.Kind() != wType.Kind() {
		return false, errors.New("mismatch type")
	}

	return !reflect.DeepEqual(match, with), nil
}

// InOp is the in operator.
type InOp OpType

// Name return the name of InOp
func (e *InOp) Name() string {
	return In
}

// Match the in operator
func (e *InOp) Match(match interface{}, with interface{}) (bool, error) {
	if match == nil || with == nil {
		return false, errors.New("invalid parameter")
	}

	if !reflect.ValueOf(match).IsValid() || !reflect.ValueOf(with).IsValid() {
		return false, errors.New("invalid parameter value")
	}

	mKind := reflect.TypeOf(match).Kind()
	if mKind == reflect.Slice || mKind == reflect.Array {
		return false, errors.New("invalid type, can not be array or slice")
	}

	wKind := reflect.TypeOf(with).Kind()
	if !(wKind == reflect.Slice || wKind == reflect.Array) {
		return false, errors.New("invalid type, should be array or slice")
	}

	// compare string if it's can
	if m, ok := match.(string); ok {
		valWith := reflect.ValueOf(with)
		for i := 0; i < valWith.Len(); i++ {
			v, ok := valWith.Index(i).Interface().(string)
			if !ok {
				return false, errors.New("unsupported compare with type")
			}
			if m == v {
				return true, nil
			}
		}
		return false, nil
	}

	// compare bool if it's can
	if m, ok := match.(bool); ok {
		valWith := reflect.ValueOf(with)
		for i := 0; i < valWith.Len(); i++ {
			v, ok := valWith.Index(i).Interface().(bool)
			if !ok {
				return false, errors.New("unsupported compare with type")
			}
			if m == v {
				return true, nil
			}
		}
		return false, nil
	}

	// compare numeric value if it's can
	if !isNumeric(match) {
		return false, errors.New("unsupported compare type")
	}

	// with value is slice or array, so we need to compare it one by one.
	hit := false
	valWith := reflect.ValueOf(with)

	for i := 0; i < valWith.Len(); i++ {
		if !isNumeric(valWith.Index(i).Interface()) {
			return false, errors.New("unsupported compare with type")
		}
		if toFloat64(match) == toFloat64(valWith.Index(i).Interface()) {
			hit = true
			break
		}
	}

	return hit, nil

}

// NotInOp is the not in operator.
type NotInOp OpType

// Name return the name of NotInOp
func (n *NotInOp) Name() string {
	return Nin
}

// Match the not in operator
func (n *NotInOp) Match(match interface{}, with interface{}) (bool, error) {
	inOp := InOp("in")
	hit, err := inOp.Match(match, with)
	if err != nil {
		return false, err
	}

	return !hit, nil
}

// ContainsOp is the contains operator.
type ContainsOp OpType

// Name return the name of ContainsOp
func (c *ContainsOp) Name() string {
	return Contains
}

// Match the contains operator
func (c *ContainsOp) Match(match interface{}, with interface{}) (bool, error) {
	m, ok := match.(string)
	if !ok {
		return false, errors.New("invalid parameter")
	}

	w, ok := with.(string)
	if !ok {
		return false, errors.New("invalid parameter")
	}

	return strings.Contains(m, w), nil
}

// NotContainsOp is the not contains operator.
type NotContainsOp OpType

// Name return the name of NotContainsOp
func (c *NotContainsOp) Name() string {
	return NContains
}

// Match the not contains operator
func (c *NotContainsOp) Match(match interface{}, with interface{}) (bool, error) {
	m, ok := match.(string)
	if !ok {
		return false, errors.New("invalid parameter")
	}

	w, ok := with.(string)
	if !ok {
		return false, errors.New("invalid parameter")
	}

	return !strings.Contains(m, w), nil
}

// StartsWithOp is the start with operator.
type StartsWithOp OpType

// Name return the name of StartsWithOp
func (s *StartsWithOp) Name() string {
	return StartWith
}

// Match the start with operator
func (s *StartsWithOp) Match(match interface{}, with interface{}) (bool, error) {
	m, ok := match.(string)
	if !ok {
		return false, errors.New("invalid parameter")
	}

	w, ok := with.(string)
	if !ok {
		return false, errors.New("invalid parameter")
	}

	return strings.HasPrefix(m, w), nil
}

// NotStartsWithOp is the not start with operator.
type NotStartsWithOp OpType

// Name return the name of NotStartsWithOp
func (n *NotStartsWithOp) Name() string {
	return NStartWith
}

// Match the not start with operator
func (n *NotStartsWithOp) Match(match interface{}, with interface{}) (bool, error) {
	m, ok := match.(string)
	if !ok {
		return false, errors.New("invalid parameter")
	}

	w, ok := with.(string)
	if !ok {
		return false, errors.New("invalid parameter")
	}

	return !strings.HasPrefix(m, w), nil
}

// EndsWithOp is the end with operator.
type EndsWithOp OpType

// Name return the name of EndsWithOp
func (e *EndsWithOp) Name() string {
	return EndWith
}

// Match the end with operator
func (e *EndsWithOp) Match(match interface{}, with interface{}) (bool, error) {
	m, ok := match.(string)
	if !ok {
		return false, errors.New("invalid parameter")
	}

	w, ok := with.(string)
	if !ok {
		return false, errors.New("invalid parameter")
	}

	return strings.HasSuffix(m, w), nil
}

// NotEndsWithOp is the not end with operator.
type NotEndsWithOp OpType

// Name return the name of NotEndsWithOp
func (e *NotEndsWithOp) Name() string {
	return NEndWith
}

// Match the not end with operator
func (e *NotEndsWithOp) Match(match interface{}, with interface{}) (bool, error) {
	m, ok := match.(string)
	if !ok {
		return false, errors.New("invalid parameter")
	}

	w, ok := with.(string)
	if !ok {
		return false, errors.New("invalid parameter")
	}

	return !strings.HasSuffix(m, w), nil
}

// LessThanOp is the less than operator.
type LessThanOp OpType

// Name return the name of LessThanOp
func (l *LessThanOp) Name() string {
	return LessThan
}

// Match the less than operator
func (l *LessThanOp) Match(match interface{}, with interface{}) (bool, error) {
	if !isNumeric(match) || !isNumeric(with) {
		return false, errors.New("invalid parameter")
	}

	return toFloat64(match) < toFloat64(with), nil
}

// LessThanEqualOp is the less than equal operator.
type LessThanEqualOp OpType

// Name return the name of LessThanEqualOp
func (l *LessThanEqualOp) Name() string {
	return LessThanEqual
}

// Match the less than equal operator
func (l *LessThanEqualOp) Match(match interface{}, with interface{}) (bool, error) {
	if !isNumeric(match) || !isNumeric(with) {
		return false, errors.New("invalid parameter")
	}

	return toFloat64(match) <= toFloat64(with), nil
}

// GreaterThanOp is the greater than operator.
type GreaterThanOp OpType

// Name return the name of GreaterThanOp
func (gt *GreaterThanOp) Name() string {
	return GreaterThan
}

// Match the greater than operator
func (gt *GreaterThanOp) Match(match interface{}, with interface{}) (bool, error) {
	if !isNumeric(match) || !isNumeric(with) {
		return false, errors.New("invalid parameter")
	}

	return toFloat64(match) > toFloat64(with), nil
}

// GreaterThanEqualOp is the greater than equal operator.
type GreaterThanEqualOp OpType

// Name return the name of GreaterThanEqualOp
func (gte *GreaterThanEqualOp) Name() string {
	return GreaterThanEqual
}

// Match the greater than equal operator
func (gte *GreaterThanEqualOp) Match(match interface{}, with interface{}) (bool, error) {
	if !isNumeric(match) || !isNumeric(with) {
		return false, errors.New("invalid parameter")
	}

	return toFloat64(match) > toFloat64(with), nil
}

// AnyOp is the any operator.
type AnyOp OpType

// Name return the name of AnyOp
func (a *AnyOp) Name() string {
	return Any
}

// Match the any operator
func (a *AnyOp) Match(match interface{}, _ interface{}) (bool, error) {
	return true, nil
}
