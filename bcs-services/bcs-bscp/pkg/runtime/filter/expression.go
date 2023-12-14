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

// Package filter provides expression filter.
package filter

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	pbstruct "github.com/golang/protobuf/ptypes/struct"
	"github.com/tidwall/gjson"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/enumor"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

const (
	// DefaultMaxInLimit defines the default max in limit
	DefaultMaxInLimit = uint(3000)
	// DefaultMaxNotInLimit defines the default max nin limit
	DefaultMaxNotInLimit = uint(3000)
	// DefaultMaxRuleLimit defines the default max number of rules limit
	DefaultMaxRuleLimit = uint(5)
)

// ExprOption defines how to validate an
// expression.
type ExprOption struct {
	// RuleFields:
	// 1. used to test if all the expression rule's field
	//    is in the RuleFields' key restricts.
	// 2. all the expression's rule filed should be a sub-set
	//    of the RuleFields' key.
	RuleFields map[string]enumor.ColumnType
	// MaxInLimit defines the max element of the in operator
	// If not set, then use default value: DefaultMaxInLimit
	MaxInLimit uint
	// MaxNotInLimit defines the max element of the nin operator
	// If not set, then use default value: DefaultMaxNotInLimit
	MaxNotInLimit uint
	// MaxRulesLimit defines the max number of rules an expression allows.
	// If not set, then use default value: DefaultMaxRuleLimit
	MaxRulesLimit uint
}

// Expression is to build a query expression
type Expression struct {
	Op    LogicOperator `json:"op"`
	Rules []RuleFactory `json:"rules"`
}

// Validate the expression is valid or not.
func (exp Expression) Validate(opts ...*ExprOption) (hitErr error) {
	defer func() {
		if hitErr != nil {
			hitErr = errf.New(errf.InvalidParameter, hitErr.Error())
		}
	}()

	if len(opts) > 1 {
		return errors.New("expression's validate option only support at most one")
	}

	if err := exp.Op.Validate(); err != nil {
		return err
	}

	if len(exp.Rules) == 0 {
		return nil
	}

	maxRules := DefaultMaxRuleLimit
	if len(opts) != 0 {
		if opts[0].MaxRulesLimit > 0 {
			maxRules = opts[0].MaxRulesLimit
		}
	}

	if len(exp.Rules) > int(maxRules) {
		return fmt.Errorf("rules elements number is overhead, it at most have %d rules", maxRules)
	}

	fieldsReminder := make(map[string]bool)
	for _, r := range exp.Rules {
		fieldsReminder[r.RuleField()] = true
	}

	if len(fieldsReminder) == 0 {
		return errors.New("invalid expression, no field is found to query")
	}

	if len(opts) != 0 {
		reminder := make(map[string]bool)
		for col := range opts[0].RuleFields {
			reminder[col] = true
		}

		// all the rule's filed should exist in the reminder.
		for one := range fieldsReminder {
			if exist := reminder[one]; !exist {
				return fmt.Errorf("expression rules filed(%s) should not exist(not supported)", one)
			}
		}
	}

	var valOpt *ExprOption
	if len(opts) != 0 {
		valOpt = opts[0]
	}

	for _, one := range exp.Rules {
		if err := one.Validate(valOpt); err != nil {
			return err
		}
	}

	return nil
}

// UnmarshalJSON unmarshal a json raw to this expression
func (exp *Expression) UnmarshalJSON(raw []byte) error {
	parsed := gjson.GetManyBytes(raw, "op", "rules")
	op := LogicOperator(parsed[0].String())
	if err := op.Validate(); err != nil {
		return err
	}
	exp.Op = op

	rules := parsed[1]

	typ, err := ruleType(rules)
	if err != nil {
		return err
	}

	switch typ {
	case AtomType:
		atoms := make([]*AtomRule, 0)
		if err := json.Unmarshal([]byte(rules.Raw), &atoms); err != nil {
			return err
		}

		for idx := range atoms {
			exp.Rules = append(exp.Rules, atoms[idx])
		}

	case EmptyType:
		exp.Rules = make([]RuleFactory, 0)

	default:
		return errors.New("unknown expression rule type")
	}

	return nil
}

// MarshalPB marshal Expression to pb struct.
func (exp *Expression) MarshalPB() (*pbstruct.Struct, error) {
	if exp == nil {
		return nil, errf.New(errf.InvalidParameter, "expression is nil")
	}

	marshal, err := json.Marshal(exp)
	if err != nil {
		return nil, err
	}

	st := new(pbstruct.Struct)
	if err = st.UnmarshalJSON(marshal); err != nil {
		return nil, err
	}

	return st, nil
}

// RuleFactory defines an expression's basic rule.
// which is used to filter the resources.
type RuleFactory interface {
	// WithType get a rule's type
	WithType() RuleType
	// Validate this rule is valid or not
	Validate(opt *ExprOption) error
	// RuleField get this rule's filed
	RuleField() string
	// SQLExpr convert this rule to a mysql's sub
	// query expression
	SQLExpr() (string, []interface{}, error)
}

var _ RuleFactory = new(AtomRule)

// AtomRule is the basic query rule.
type AtomRule struct {
	Field string      `json:"field"`
	Op    OpFactory   `json:"op"`
	Value interface{} `json:"value"`
}

// WithType return this atom rule's tye.
func (ar AtomRule) WithType() RuleType {
	return AtomType
}

// Validate this atom rule is valid or not
// Note: opt can be nil, check it before use it.
func (ar AtomRule) Validate(opt *ExprOption) error {
	if len(ar.Field) == 0 {
		return errors.New("filed is empty")
	}

	// validate operator
	if err := ar.Op.Validate(); err != nil {
		return err
	}

	if ar.Value == nil {
		return errors.New("rule value can not be nil")
	}

	if opt != nil {
		typ, exist := opt.RuleFields[ar.Field]
		if !exist {
			return fmt.Errorf("rule field: %s is not exist in the expr option", ar.Field)
		}

		if err := validateFieldValue(ar.Value, typ); err != nil {
			return fmt.Errorf("invalid %s's value, %v", ar.Field, err)
		}
	}

	// validate the operator's value
	if err := ar.Op.Operator().ValidateValue(ar.Value, opt); err != nil {
		return fmt.Errorf("%s validate failed, %v", ar.Field, err)
	}

	return nil
}

func validateFieldValue(v interface{}, typ enumor.ColumnType) error {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Array, reflect.Slice:
		return validateSliceElements(v, typ)
	default:
	}

	switch typ {
	case enumor.String:
		if reflect.ValueOf(v).Type().Kind() != reflect.String {
			return errors.New("value should be a string")
		}

	case enumor.Numeric:
		if !tools.IsNumeric(v) {
			return errors.New("value should be a numeric")
		}

	case enumor.Boolean:
		if reflect.ValueOf(v).Type().Kind() != reflect.Bool {
			return errors.New("value should be a boolean")
		}

	case enumor.Time:
		valOf := reflect.ValueOf(v)
		if valOf.Type().Kind() != reflect.String {
			return fmt.Errorf("value should be a string time format like: %s", constant.TimeStdFormat)
		}

		if !constant.TimeStdRegexp.MatchString(valOf.String()) {
			return fmt.Errorf("invalid time format, should be like: %s", constant.TimeStdFormat)
		}

		_, err := time.Parse(constant.TimeStdFormat, valOf.String())
		if err != nil {
			return fmt.Errorf("parse time from value failed, err: %v", err)
		}

	default:
		return fmt.Errorf("unsupported value type format: %s", typ)
	}

	return nil
}

func validateSliceElements(v interface{}, typ enumor.ColumnType) error {
	value := reflect.ValueOf(v)
	length := value.Len()
	if length == 0 {
		return nil
	}

	// validate each slice's element data type
	for i := 0; i < length; i++ {
		if err := validateFieldValue(value.Index(i).Interface(), typ); err != nil {
			return err
		}
	}

	return nil
}

// RuleField get atom rule's filed
func (ar AtomRule) RuleField() string {
	return ar.Field
}

// SQLExpr convert this atom rule to a mysql's sub
// query expression.
func (ar AtomRule) SQLExpr() (string, []interface{}, error) {
	return ar.Op.Operator().SQLExpr(ar.Field, ar.Value)
}

type broker struct {
	Field string          `json:"field"`
	Op    OpFactory       `json:"op"`
	Value json.RawMessage `json:"value"`
}

// UnmarshalJSON unmarshal the json raw to AtomRule
func (ar *AtomRule) UnmarshalJSON(raw []byte) error {
	br := new(broker)
	err := json.Unmarshal(raw, br)
	if err != nil {
		return err
	}

	ar.Field = br.Field
	ar.Op = br.Op
	if br.Op == OpFactory(In) || br.Op == OpFactory(NotIn) {
		// in and nin operator's value should be an array.
		array := make([]interface{}, 0)
		if err := json.Unmarshal(br.Value, &array); err != nil {
			return err
		}

		ar.Value = array

		return nil
	}

	to := new(interface{})
	if err := json.Unmarshal(br.Value, to); err != nil {
		return err
	}
	ar.Value = *to

	return nil
}
