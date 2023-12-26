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

package selector

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/validator"
)

// Operator is a factory which defines the operators which is supported to match labels
type Operator interface {
	// Name is the name of operator
	Name() OperatorType

	// Validate is to validate if a element with this operator is valid or not.
	Validate(match *Element) error

	// Match is used to check if "match" is "logical equal" to the "labels"
	// with different OperatorType, different OperatorType has different definition
	// of "logical equal", if "logical equal" then return bool "true" value.

	// Match the value to test labels.
	// labels: the labels to be matched, which the compared match key
	// should exist at first. otherwise, it's always not matched with this label.
	Match(match *Element, labels map[string]string) (bool, error)
}

// OperatorType defines a operator type to describe every supported operator.
type OperatorType string

// defines all the supported operator type
const (
	Equal            OperatorType = "eq"
	NotEqual         OperatorType = "ne"
	GreaterThan      OperatorType = "gt"
	GreaterThanEqual OperatorType = "ge"
	LessThan         OperatorType = "lt"
	LessThanEqual    OperatorType = "le"
	In               OperatorType = "in"
	NotIn            OperatorType = "nin"
	Regex            OperatorType = "re"
	NotRegex         OperatorType = "nre"
)

// supported default operators
var (
	EqualOperator            = EqualType(Equal)
	NotEqualOperator         = NotEqualType(NotEqual)
	GreaterThanOperator      = GreaterThanType(GreaterThan)
	GreaterThanEqualOperator = GreaterThanEqualType(GreaterThanEqual)
	LessThanOperator         = LessThanType(LessThan)
	LessThanEqualOperator    = LessThanEqualType(LessThanEqual)
	InOperator               = InType(In)
	NotInOperator            = NotInType(NotIn)
	RegexOperator            = RegexType(Regex)
	NotRegexOperator         = NotRegexType(NotRegex)
)

// OperatorEnums enum all the supported operators.
var OperatorEnums = map[OperatorType]Operator{
	Equal:            &EqualOperator,
	NotEqual:         &NotEqualOperator,
	GreaterThan:      &GreaterThanOperator,
	GreaterThanEqual: &GreaterThanEqualOperator,
	LessThan:         &LessThanOperator,
	LessThanEqual:    &LessThanEqualOperator,
	In:               &InOperator,
	NotIn:            &NotInOperator,
	Regex:            &RegexOperator,
	NotRegex:         &NotRegexOperator,
}

var _ Operator = new(EqualType)

// EqualType is a equal operator
type EqualType OperatorType

// Name is the name of equal operator
func (eq *EqualType) Name() OperatorType {
	return Equal
}

// Validate valid the match element is valid to equal operator or not
func (eq *EqualType) Validate(match *Element) error {
	v, ok := match.Value.(string)
	if !ok {
		return fmt.Errorf("invalid eq oper with value: %v, should be string", match.Value)
	}

	if err := validator.ValidateLabelValue(v); err != nil {
		return fmt.Errorf("invalid label key's value, key: %s value: %s, %v", match.Key, v, err)
	}

	return nil
}

// Match matched only when the match key is exist and test value is equal with it's target value.
func (eq *EqualType) Match(match *Element, labels map[string]string) (bool, error) {
	val, ok := match.Value.(string)
	if !ok {
		return false, fmt.Errorf("invalid eq oper with value: %v, should be string", match.Value)
	}

	to, exists := labels[match.Key]
	if !exists {
		return false, nil
	}

	return val == to, nil
}

var _ Operator = new(NotEqualType)

// NotEqualType is a not equal operator
type NotEqualType OperatorType

// Name is the name of not equal operator
func (ne *NotEqualType) Name() OperatorType {
	return NotEqual
}

// Validate valid the match element is valid to not equal operator or not
func (ne *NotEqualType) Validate(match *Element) error {
	v, ok := match.Value.(string)
	if !ok {
		return fmt.Errorf("invalid ne oper with value: %v, should be string", match.Value)
	}

	if err := validator.ValidateLabelValue(v); err != nil {
		return fmt.Errorf("invalid label key's value, key: %s value: %s, %v", match.Key, v, err)
	}

	return nil
}

// Match matched only when the match key is exist and value is not equal with it's target value.
func (ne *NotEqualType) Match(match *Element, with map[string]string) (bool, error) {
	val, ok := match.Value.(string)
	if !ok {
		return false, fmt.Errorf("invalid ne oper with value: %v, should be string", match.Value)
	}

	to, exists := with[match.Key]
	if !exists {
		return false, nil
	}

	return val != to, nil
}

var _ Operator = new(GreaterThanType)

// GreaterThanType is a greater than operator
type GreaterThanType OperatorType

// Name is the name of greater than operator
func (gt *GreaterThanType) Name() OperatorType {
	return GreaterThan
}

// Validate valid the match element is valid to greater than operator or not
func (gt *GreaterThanType) Validate(match *Element) error {
	if !isNumeric(match.Value) {
		return fmt.Errorf("invalid gt oper with value: %v, should be number", match.Value)
	}
	return nil
}

// Match matched only when the match key is exist and value is greater than it's target value.
func (gt *GreaterThanType) Match(match *Element, labels map[string]string) (bool, error) {
	if !isNumeric(match.Value) {
		return false, fmt.Errorf("invalid gt oper with value: %v, should be number", match.Value)
	}

	from := mustFloat64(match.Value)

	compare, exists := labels[match.Key]
	if !exists {
		return false, nil
	}

	to, err := strconv.ParseFloat(compare, 32)
	if err != nil {
		return false, fmt.Errorf("parse gt oper's target label value: %s to float failed, err: %v", compare, err)
	}

	return to > from, nil
}

var _ Operator = new(GreaterThanEqualType)

// GreaterThanEqualType is a greater than equal operator
type GreaterThanEqualType OperatorType

// Name is the name of greater than equal operator
func (ge *GreaterThanEqualType) Name() OperatorType {
	return GreaterThanEqual
}

// Validate valid the match element is valid to greater than equal operator or not
func (ge *GreaterThanEqualType) Validate(match *Element) error {
	if !isNumeric(match.Value) {
		return fmt.Errorf("invalid ge oper with value: %v, should be number", match.Value)
	}
	return nil
}

// Match matched only when the match key is exist and value is greater than equal with it's target value.
func (ge *GreaterThanEqualType) Match(match *Element, with map[string]string) (bool, error) {
	if !isNumeric(match.Value) {
		return false, fmt.Errorf("invalid ge oper with value: %v, should be number", match.Value)
	}

	from := mustFloat64(match.Value)

	compare, exists := with[match.Key]
	if !exists {
		return false, nil
	}

	to, err := strconv.ParseFloat(compare, 32)
	if err != nil {
		return false, fmt.Errorf("parse ge oper's target label value: %s to float failed, err: %v", compare, err)
	}

	return to >= from, nil
}

var _ Operator = new(LessThanEqualType)

// LessThanType is a less than operator
type LessThanType OperatorType

// Name is the name of less than operator
func (lt *LessThanType) Name() OperatorType {
	return LessThan
}

// Validate valid the match element is valid to less than operator or not
func (lt *LessThanType) Validate(match *Element) error {
	if !isNumeric(match.Value) {
		return fmt.Errorf("invalid lt oper with value: %v, should be number", match.Value)
	}
	return nil
}

// Match matched only when the match key is exist and value is less than with it's target value.
func (lt *LessThanType) Match(match *Element, labels map[string]string) (bool, error) {
	if !isNumeric(match.Value) {
		return false, fmt.Errorf("invalid lt oper with value: %v, should be number", match.Value)
	}

	from := mustFloat64(match.Value)

	compare, exists := labels[match.Key]
	if !exists {
		return false, nil
	}

	to, err := strconv.ParseFloat(compare, 32)
	if err != nil {
		return false, fmt.Errorf("parse lt oper's target label value: %s to float failed, err: %v", compare, err)
	}

	return to < from, nil
}

var _ Operator = new(LessThanEqualType)

// LessThanEqualType is a less than equal operator
type LessThanEqualType OperatorType

// Name is the name of less than equal operator
func (le *LessThanEqualType) Name() OperatorType {
	return LessThanEqual
}

// Validate valid the match element is valid to less than equal operator or not
func (le *LessThanEqualType) Validate(match *Element) error {
	if !isNumeric(match.Value) {
		return fmt.Errorf("invalid le oper with value: %v, should be number", match.Value)
	}
	return nil
}

// Match matched only when the match key is exist and value is less than equal with it's target value.
func (le *LessThanEqualType) Match(match *Element, labels map[string]string) (bool, error) {
	if !isNumeric(match.Value) {
		return false, fmt.Errorf("invalid le oper with value: %v, should be number", match.Value)
	}

	from := mustFloat64(match.Value)

	compare, exists := labels[match.Key]
	if !exists {
		return false, nil
	}

	to, err := strconv.ParseFloat(compare, 32)
	if err != nil {
		return false, fmt.Errorf("parse le oper's target label value: %s to float failed, err: %v", compare, err)
	}

	return to <= from, nil
}

var _ Operator = new(InType)

// InType is an in operator
type InType OperatorType

// Name is the name of in operator
func (in *InType) Name() OperatorType {
	return In
}

// MaxInTypeElementSize is the max element number of an in type.
const MaxInTypeElementSize = 10

// Validate the match element is valid to in operator or not
func (in *InType) Validate(match *Element) error {
	values, ok := match.Value.([]interface{})
	if !ok {
		return fmt.Errorf("invalid in oper with value: %v, should be array string", match.Value)
	}

	if len(values) > 10 {
		return fmt.Errorf("oversize the in operator's value size, max size: %d ", MaxInTypeElementSize)
	}

	for i := range values {
		v, ok := values[i].(string)
		if !ok {
			return fmt.Errorf("invalid in oper with value: %v, should be array string", match.Value)
		}

		if err := validator.ValidateLabelValue(v); err != nil {
			return fmt.Errorf("invalid label key's value, key: %s value: %s, %v", match.Key, v, err)
		}
	}

	return nil
}

// Match matched only when the match key is exist and value is in it's target values.
func (in *InType) Match(match *Element, labels map[string]string) (bool, error) {
	values, ok := match.Value.([]interface{})
	if !ok {
		return false, fmt.Errorf("invalid in oper with value: %v, should be array string", match.Value)
	}

	to, exists := labels[match.Key]
	if !exists {
		return false, nil
	}

	for _, val := range values {
		if val == to {
			return true, nil
		}
	}

	return false, nil
}

var _ Operator = new(NotInType)

// NotInType is a not in operator
type NotInType OperatorType

// Name is the name of in operator
func (nin *NotInType) Name() OperatorType {
	return NotIn
}

// MaxNotInTypeElementSize is the max element number of a nin type.
const MaxNotInTypeElementSize = 10

// Validate valid the match element is valid to not in operator or not
func (nin *NotInType) Validate(match *Element) error {
	values, ok := match.Value.([]interface{})
	if !ok {
		return fmt.Errorf("invalid nin oper with value: %v, should be array string", match.Value)
	}

	if len(values) > 10 {
		return fmt.Errorf("oversize the nin operator's value size, max size: %d ", MaxNotInTypeElementSize)
	}

	for i := range values {
		v, ok := values[i].(string)
		if !ok {
			return fmt.Errorf("invalid nin oper with value: %v, should be array string", match.Value)
		}

		if err := validator.ValidateLabelValue(v); err != nil {
			return fmt.Errorf("invalid label key's value, key: %s value: %s, %v", match.Key, v, err)
		}
	}
	return nil
}

// Match matched only when the match key is exist and value is not in it's target values.
func (nin *NotInType) Match(match *Element, labels map[string]string) (bool, error) {
	values, ok := match.Value.([]interface{})
	if !ok {
		return false, fmt.Errorf("invalid nin oper with value: %v, should be array string", match.Value)
	}

	to, exists := labels[match.Key]
	if !exists {
		return false, nil
	}

	for _, val := range values {
		if val == to {
			return false, nil
		}
	}

	return true, nil
}

var _ Operator = new(RegexType)

// RegexType is a regex operator
type RegexType OperatorType

// Name is the name of in operator
func (re *RegexType) Name() OperatorType {
	return Regex
}

// Validate valid the match element is valid to match regex operator or not
func (re *RegexType) Validate(match *Element) error {
	v, ok := match.Value.(string)
	if !ok {
		return fmt.Errorf("invalid re oper with value: %v, should be string", match.Value)
	}

	if err := validator.ValidateLabelValue(v); err != nil {
		return fmt.Errorf("invalid label key's value, key: %s value: %s, %v", match.Key, v, err)
	}
	if err := validator.ValidateLabelValueRegex(v); err != nil {
		return fmt.Errorf("invalid label key's value, key: %s value: %s, %v", match.Key, v, err)
	}
	return nil
}

// Match matched only when the match key is exist and test value is matched with regular expression.
func (re *RegexType) Match(match *Element, labels map[string]string) (bool, error) {
	val, ok := match.Value.(string)
	if !ok {
		return false, fmt.Errorf("invalid re oper with value: %v, should be string", match.Value)
	}

	to, exists := labels[match.Key]
	if !exists {
		return false, nil
	}

	return regexp.MatchString(val, to)
}

var _ Operator = new(NotRegexType)

// NotRegexType is a not regex operator
type NotRegexType OperatorType

// Name is the name of not regex operator
func (nre *NotRegexType) Name() OperatorType {
	return NotRegex
}

// Validate valid the match element is valid to match not regex operator or not
func (nre *NotRegexType) Validate(match *Element) error {
	v, ok := match.Value.(string)
	if !ok {
		return fmt.Errorf("invalid nre oper with value: %v, should be string", match.Value)
	}

	if err := validator.ValidateLabelValue(v); err != nil {
		return fmt.Errorf("invalid label key's value, key: %s value: %s, %v", match.Key, v, err)
	}
	if err := validator.ValidateLabelValueRegex(v); err != nil {
		return fmt.Errorf("invalid label key's value, key: %s value: %s, %v", match.Key, v, err)
	}

	return nil
}

// Match matched only when the match key is exist and test value is not matched with regular expression.
func (nre *NotRegexType) Match(match *Element, labels map[string]string) (bool, error) {
	val, ok := match.Value.(string)
	if !ok {
		return false, fmt.Errorf("invalid nre oper with value: %v, should be string", match.Value)
	}

	to, exists := labels[match.Key]
	if !exists {
		return false, nil
	}

	matched, err := regexp.MatchString(val, to)
	if err != nil {
		return false, err
	}
	return !matched, nil
}

func isNumeric(val interface{}) bool {
	switch val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, json.Number:
		return true
	}
	return false
}

// mustFloat64 convert a interface to float64, if not, it will be panic.
// so the interface should be numeric.
func mustFloat64(val interface{}) float64 {
	switch v := val.(type) {
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case json.Number:
		val, _ := val.(json.Number).Float64()
		return val
	case float64:
		return v
	case float32:
		return float64(v)
	default:
		panic(fmt.Sprintf("unsupported type, value: %v", val))
	}
}
