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

package filter

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

var opFactory map[OpFactory]Operator

func init() {
	opFactory = make(map[OpFactory]Operator)

	eq := EqualOp(Equal)
	opFactory[OpFactory(eq.Name())] = &eq

	neq := NotEqualOp(NotEqual)
	opFactory[OpFactory(neq.Name())] = &neq

	gt := GreaterThanOp(GreaterThan)
	opFactory[OpFactory(gt.Name())] = &gt

	gte := GreaterThanEqualOp(GreaterThanEqual)
	opFactory[OpFactory(gte.Name())] = &gte

	lt := LessThanOp(LessThan)
	opFactory[OpFactory(lt.Name())] = &lt

	lte := LessThanEqualOp(LessThanEqual)
	opFactory[OpFactory(lte.Name())] = &lte

	in := InOp(In)
	opFactory[OpFactory(in.Name())] = &in

	nin := NotInOp(NotIn)
	opFactory[OpFactory(nin.Name())] = &nin

	cs := ContainsSensitiveOp(ContainsSensitive)
	opFactory[OpFactory(cs.Name())] = &cs

	cis := ContainsInsensitiveOp(ContainsInsensitive)
	opFactory[OpFactory(cis.Name())] = &cis

}

const (
	// And logic operator
	And LogicOperator = "and"
	// Or logic operator
	Or LogicOperator = "or"
)

// LogicOperator defines the logic operator
type LogicOperator string

// Validate the logic operator is valid or not.
func (lo LogicOperator) Validate() error {
	switch lo {
	case And:
	case Or:
	default:
		return fmt.Errorf("unsupported expression's logic operator: %s", lo)
	}

	return nil
}

// OpFactory defines the operator's factory type.
type OpFactory string

// Operator return this operator factory's Operator
func (of OpFactory) Operator() Operator {
	op, exist := opFactory[of]
	if !exist {
		unknown := UnknownOp(Unknown)
		return &unknown
	}

	return op
}

// Validate this operator factory is valid or not.
func (of OpFactory) Validate() error {
	typ := OpType(of)
	return typ.Validate()
}

const (
	// Unknown is an unsupported operator
	Unknown OpType = "unknown"
	// Equal operator
	Equal OpType = "eq"
	// NotEqual operator
	NotEqual OpType = "neq"
	// GreaterThan operator
	GreaterThan OpType = "gt"
	// GreaterThanEqual operator
	GreaterThanEqual OpType = "gte"
	// LessThan operator
	LessThan OpType = "lt"
	// LessThanEqual operator
	LessThanEqual OpType = "lte"
	// In operator
	In OpType = "in"
	// NotIn operator
	NotIn OpType = "nin"
	// ContainsSensitive operator match the value with
	// regular expression with case-sensitive.
	ContainsSensitive OpType = "cs"
	// ContainsInsensitive operator match the value with
	// regular expression with case-insensitive.
	ContainsInsensitive OpType = "cis"
)

// OpType defines the operators supported by mysql.
type OpType string

// Validate test the operator is valid or not.
func (op OpType) Validate() error {
	switch op {
	case Equal, NotEqual,
		GreaterThan, GreaterThanEqual,
		LessThan, LessThanEqual,
		In, NotIn,
		ContainsSensitive, ContainsInsensitive:
	default:
		return fmt.Errorf("unsupported operator: %s", op)
	}

	return nil
}

// Factory return opType's factory type.
func (op OpType) Factory() OpFactory {
	return OpFactory(op)
}

// Operator is a collection of supported query operators.
type Operator interface {
	// Name is the operator's name
	Name() OpType
	// ValidateValue validate the operator's value is valid or not
	ValidateValue(v interface{}, opt *ExprOption) error
	// SQLExpr generate an operator's SQL expression with its filed
	// and value.
	SQLExpr(field string, value interface{}) (string, []interface{}, error)
}

// UnknownOp is unknown operator
type UnknownOp OpType

// Name is equal operator
func (uo UnknownOp) Name() OpType {
	return Unknown
}

// ValidateValue validate equal's value
func (uo UnknownOp) ValidateValue(_ interface{}, _ *ExprOption) error {
	return errors.New("unknown operator")
}

// SQLExpr convert this operator's field and value to a mysql's sub
// query expression.
func (uo UnknownOp) SQLExpr(_ string, _ interface{}) (string, []interface{}, error) {
	return "", []interface{}{}, errors.New("unknown operator, can not gen sql expression")
}

// EqualOp is equal operator type
type EqualOp OpType

// Name is equal operator
func (eo EqualOp) Name() OpType {
	return Equal
}

// ValidateValue validate equal's value
func (eo EqualOp) ValidateValue(v interface{}, opt *ExprOption) error {
	if !tools.IsBasicValue(v) {
		return errors.New("invalid value field")
	}
	return nil
}

// SQLExpr convert this operator's field and value to a mysql's sub
// query expression.
func (eo EqualOp) SQLExpr(field string, value interface{}) (string, []interface{}, error) {
	var argList []interface{}
	if len(field) == 0 {
		return "", argList, errors.New("field is empty")
	}

	if !tools.IsBasicValue(value) {
		return "", argList, errors.New("invalid value field")
	}

	argList = append(argList, value)
	return fmt.Sprintf(`%s = ?`, field), argList, nil
}

// NotEqualOp is not equal operator type
type NotEqualOp OpType

// Name is not equal operator
func (ne NotEqualOp) Name() OpType {
	return NotEqual
}

// ValidateValue validate not equal's value
func (ne NotEqualOp) ValidateValue(v interface{}, opt *ExprOption) error {
	if !tools.IsBasicValue(v) {
		return errors.New("invalid ne operator's value field")
	}
	return nil
}

// SQLExpr convert this operator's field and value to a mysql's sub
// query expression.
func (ne NotEqualOp) SQLExpr(field string, value interface{}) (string, []interface{}, error) {
	var argList []interface{}
	if len(field) == 0 {
		return "", argList, errors.New("field is empty")
	}

	if !tools.IsBasicValue(value) {
		return "", argList, errors.New("invalid ne operator's value field")
	}

	argList = append(argList, value)
	return fmt.Sprintf(`%s != ?`, field), argList, nil
}

// GreaterThanOp is greater than operator
type GreaterThanOp OpType

// Name is greater than operator
func (gt GreaterThanOp) Name() OpType {
	return GreaterThan
}

// ValidateValue validate greater than value
func (gt GreaterThanOp) ValidateValue(v interface{}, opt *ExprOption) error {
	if _, hit := isNumericOrTime(v); !hit {
		return errors.New("invalid gt operator's value, should be a numeric or time format string value")
	}
	return nil
}

// SQLExpr convert this operator's field and value to a mysql's sub
// query expression.
func (gt GreaterThanOp) SQLExpr(field string, value interface{}) (string, []interface{}, error) {
	var argList []interface{}
	if len(field) == 0 {
		return "", argList, errors.New("field is empty")
	}

	_, hit := isNumericOrTime(value)
	if !hit {
		return "", argList, errors.New("invalid gt operator's value field, should be a numeric value")
	}

	argList = append(argList, value)
	return fmt.Sprintf(`%s > ?`, field), argList, nil
}

// GreaterThanEqualOp is greater than equal operator
type GreaterThanEqualOp OpType

// Name is greater than operator
func (gte GreaterThanEqualOp) Name() OpType {
	return GreaterThanEqual
}

// ValidateValue validate greater than value
func (gte GreaterThanEqualOp) ValidateValue(v interface{}, opt *ExprOption) error {
	if _, hit := isNumericOrTime(v); !hit {
		return errors.New("invalid gte operator's value, should be a numeric or time format string value")
	}
	return nil
}

// SQLExpr convert this operator's field and value to a mysql's sub
// query expression.
func (gte GreaterThanEqualOp) SQLExpr(field string, value interface{}) (string, []interface{}, error) {
	var argList []interface{}
	if len(field) == 0 {
		return "", argList, errors.New("field is empty")
	}

	_, hit := isNumericOrTime(value)
	if !hit {
		return "", argList, errors.New("invalid gte operator's value field, should be a numeric value")
	}

	argList = append(argList, value)

	return fmt.Sprintf(`%s >= ?`, field), argList, nil
}

// LessThanOp is less than operator
type LessThanOp OpType

// Name is less than equal operator
func (lt LessThanOp) Name() OpType {
	return LessThan
}

// ValidateValue validate less than equal value
func (lt LessThanOp) ValidateValue(v interface{}, opt *ExprOption) error {
	if _, hit := isNumericOrTime(v); !hit {
		return errors.New("invalid lt operator's value, should be a numeric or time format string value")
	}
	return nil
}

// SQLExpr convert this operator's field and value to a mysql's sub
// query expression.
func (lt LessThanOp) SQLExpr(field string, value interface{}) (string, []interface{}, error) {
	var argList []interface{}
	if len(field) == 0 {
		return "", argList, errors.New("field is empty")
	}

	_, hit := isNumericOrTime(value)
	if !hit {
		return "", argList, errors.New("invalid lt operator's value field, should be a numeric value")
	}

	argList = append(argList, value)
	return fmt.Sprintf(`%s < ?`, field), argList, nil
}

// LessThanEqualOp is less than equal operator
type LessThanEqualOp OpType

// Name is less than equal operator
func (lte LessThanEqualOp) Name() OpType {
	return LessThanEqual
}

// ValidateValue validate less than equal value
func (lte LessThanEqualOp) ValidateValue(v interface{}, opt *ExprOption) error {
	if _, hit := isNumericOrTime(v); !hit {
		return errors.New("invalid lte operator's value, should be a numeric or time format string value")
	}
	return nil
}

// SQLExpr convert this operator's field and value to a mysql's sub
// query expression.
func (lte LessThanEqualOp) SQLExpr(field string, value interface{}) (string, []interface{}, error) {
	var argList []interface{}
	if len(field) == 0 {
		return "", argList, errors.New("field is empty")
	}

	_, hit := isNumericOrTime(value)
	if !hit {
		return "", argList, errors.New("invalid lte operator's value field, should be a numeric value")
	}

	argList = append(argList, value)
	return fmt.Sprintf(`%s <= ?`, field), argList, nil
}

// InOp is in operator
type InOp OpType

// Name is in operator
func (io InOp) Name() OpType {
	return In
}

// ValidateValue validate in operator's value
func (io InOp) ValidateValue(v interface{}, opt *ExprOption) error {

	switch reflect.TypeOf(v).Kind() {
	case reflect.Array:
	case reflect.Slice:
	default:
		return errors.New("in operator's value should be an array")
	}

	value := reflect.ValueOf(v)
	length := value.Len()
	if length == 0 {
		return errors.New("invalid in operator's value, at least have one element")
	}

	maxInV := DefaultMaxInLimit
	if opt != nil {
		if opt.MaxInLimit > 0 {
			maxInV = opt.MaxInLimit
		}
	}

	if length > int(maxInV) {
		return fmt.Errorf("invalid in operator's value, at most have %d elements", maxInV)
	}

	// each element in the array or slice should be a basic type.
	for i := 0; i < length; i++ {
		if !tools.IsBasicValue(value.Index(i).Interface()) {
			return fmt.Errorf("invalid in operator's value: %v, each element's value should be a basic type",
				value.Index(i).Interface())
		}
	}

	return nil
}

// SQLExpr convert this operator's field and value to a mysql's sub
// query expression.
func (io InOp) SQLExpr(field string, value interface{}) (string, []interface{}, error) {
	var argList []interface{}
	if len(field) == 0 {
		return "", argList, errors.New("field is empty")
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.Array:
	case reflect.Slice:
	default:
		return "", argList, errors.New("in operator's value should be an array")
	}

	valOf := reflect.ValueOf(value)
	length := valOf.Len()
	if length == 0 {
		return "", argList, errors.New("invalid in operator's value, at least have one element")
	}

	var joined string
	for i := 0; i < length; i++ {
		ele := valOf.Index(i).Interface()
		if !tools.IsBasicValue(ele) {
			return "", []interface{}{},
				fmt.Errorf("invalid in operator's value: %v, each element's value should be a basic type", ele)
		}

		joined = fmt.Sprintf("%s, ?", joined)
		argList = append(argList, ele)
	}

	joined = strings.Trim(joined, ",")
	joined = strings.TrimSpace(joined)

	return fmt.Sprintf(`%s IN (%s)`, field, joined), argList, nil
}

// NotInOp is not in operator
type NotInOp OpType

// Name is not in operator
func (nio NotInOp) Name() OpType {
	return NotIn
}

// ValidateValue validate not in value
func (nio NotInOp) ValidateValue(v interface{}, opt *ExprOption) error {

	switch reflect.TypeOf(v).Kind() {
	case reflect.Array:
	case reflect.Slice:
	default:
		return errors.New("nin operator's value should be an array")
	}

	value := reflect.ValueOf(v)
	length := value.Len()
	if length == 0 {
		return errors.New("invalid nin operator's value, at least have one element")
	}

	maxNotInV := DefaultMaxNotInLimit
	if opt != nil {
		if opt.MaxNotInLimit > 0 {
			maxNotInV = opt.MaxNotInLimit
		}
	}

	if length > int(maxNotInV) {
		return fmt.Errorf("invalid nin operator's value, at most have %d elements", maxNotInV)
	}

	// each element in the array or slice should be a basic type.
	for i := 0; i < length; i++ {
		if !tools.IsBasicValue(value.Index(i).Interface()) {
			return fmt.Errorf("invalid nin operator's value: %v, each element's value should be a basic type",
				value.Index(i).Interface())
		}
	}

	return nil
}

// SQLExpr convert this operator's field and value to a mysql's sub
// query expression.
func (nio NotInOp) SQLExpr(field string, value interface{}) (string, []interface{}, error) {
	var argList []interface{}
	if len(field) == 0 {
		return "", argList, errors.New("field is empty")
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.Array:
	case reflect.Slice:
	default:
		return "", argList, errors.New("nin operator's value should be an array")
	}

	valOf := reflect.ValueOf(value)
	length := valOf.Len()
	if length == 0 {
		return "", argList, errors.New("invalid nin operator's value, at least have one element")
	}

	var joined string
	for i := 0; i < length; i++ {
		ele := valOf.Index(i).Interface()
		if !tools.IsBasicValue(ele) {
			return "", []interface{}{},
				fmt.Errorf("invalid nin operator's value: %v, each element's value should be a basic type", ele)
		}

		joined = fmt.Sprintf("%s, ?", joined)
		argList = append(argList, ele)

	}

	joined = strings.Trim(joined, ",")
	joined = strings.TrimSpace(joined)

	return fmt.Sprintf(`%s NOT IN (%s)`, field, joined), argList, nil
}

// ContainsSensitiveOp is contains sensitive operator
type ContainsSensitiveOp OpType

// Name is 'like' expression with camel sensitive operator
func (cso ContainsSensitiveOp) Name() OpType {
	return ContainsSensitive
}

// ValidateValue validate 'like' operator's value
func (cso ContainsSensitiveOp) ValidateValue(v interface{}, opt *ExprOption) error {

	if reflect.TypeOf(v).Kind() != reflect.String {
		return errors.New("cs operator's value should be an string")
	}

	value, ok := v.(string)
	if !ok {
		return errors.New("cs operator's value should be an string")
	}

	if len(value) == 0 {
		return errors.New("cs operator's value can not be a empty string")
	}

	return nil
}

// SQLExpr convert this operator's field and value to a mysql's sub
// query expression.
func (cso ContainsSensitiveOp) SQLExpr(field string, value interface{}) (string, []interface{}, error) {
	var argList []interface{}
	if len(field) == 0 {
		return "", argList, errors.New("field is empty")
	}

	if reflect.TypeOf(value).Kind() != reflect.String {
		return "", argList, errors.New("cs operator's value should be an string")
	}

	s, ok := value.(string)
	if !ok {
		return "", argList, errors.New("cs operator's value should be an string")
	}

	if len(s) == 0 {
		return "", argList, errors.New("cs operator's value can not be a empty string")
	}
	argList = append(argList, fmt.Sprintf("%%%v%%", value))
	return fmt.Sprintf(`%s LIKE BINARY ?`, field), argList, nil
}

// ContainsInsensitiveOp is contains insensitive operator
type ContainsInsensitiveOp OpType

// Name is 'like' expression with camel insensitive operator
func (cio ContainsInsensitiveOp) Name() OpType {
	return ContainsInsensitive
}

// ValidateValue validate 'like' operator's value
func (cio ContainsInsensitiveOp) ValidateValue(v interface{}, opt *ExprOption) error {

	if reflect.TypeOf(v).Kind() != reflect.String {
		return errors.New("cis operator's value should be an string")
	}

	value, ok := v.(string)
	if !ok {
		return errors.New("cis operator's value should be an string")
	}

	if len(value) == 0 {
		return errors.New("cis operator's value can not be a empty string")
	}

	return nil
}

// SQLExpr convert this operator's field and value to a mysql's sub
// query expression.
func (cio ContainsInsensitiveOp) SQLExpr(field string, value interface{}) (string, []interface{}, error) {
	var argList []interface{}
	argList = append(argList, fmt.Sprintf("%%%v%%", value))
	return fmt.Sprintf(`%s LIKE ?`, field), argList, nil
}
