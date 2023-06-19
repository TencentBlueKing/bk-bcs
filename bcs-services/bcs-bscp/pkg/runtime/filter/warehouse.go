/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package filter

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/tools"

	"github.com/tidwall/gjson"
)

// isNumericOrTime test if a value is a standard time
// format string or a numeric value.
func isNumericOrTime(v interface{}) (numeric bool, hit bool) {
	valOf := reflect.ValueOf(v)
	if valOf.Type().Kind() == reflect.String {
		// test if this value is a standard time string
		if !constant.TimeStdRegexp.MatchString(valOf.String()) {
			return false, false
		}

		_, err := time.Parse(constant.TimeStdFormat, valOf.String())
		if err != nil {
			return false, false
		}

		return false, true
	}

	if tools.IsNumeric(v) {
		return true, true
	}

	return false, false
}

// Priority defines the SQL Where option's priority.
type Priority []string

// SQLWhereOption defines how to generate the SQL expression with Expression.
type SQLWhereOption struct {
	// Priority defines the ordered expression rule's fields to generate the SQL
	// expression.
	// The lower the index of the priority's array, the higher priority of the
	// field during query.
	Priority      Priority
	CrownedOption *CrownedOption
}

// Validate the options is valid or not
func (sop SQLWhereOption) Validate() error {
	if len(sop.Priority) == 0 {
		return errors.New("priority fields can not be empty, should be the resource table's index")
	}

	if sop.CrownedOption == nil {
		return nil
	}

	if len(sop.CrownedOption.Rules) == 0 {
		return nil
	}

	if err := sop.CrownedOption.CrownedOp.Validate(); err != nil {
		return err
	}

	for _, one := range sop.CrownedOption.Rules {
		if err := one.Validate(nil); err != nil {
			return err
		}
	}

	return nil
}

// CrownedOption defines to be crowned options with its parent expression.
// This CrownedOption.CrownedOp defines the logic operator with its parent
// 'Expression'.
// The generated crowned expression rest on the top of its parent 'Expression'.
type CrownedOption struct {
	// CrownedOp is the logic operator to operate with its fully parent 'Expression'.
	CrownedOp LogicOperator
	// Rules defines all the rules to be appended to its parent 'Expression'.
	// Note: all these rules is operator with logic 'AND'.
	Rules []RuleFactory
}

type hitType string

const (
	exprType  hitType = "expr"
	crownType hitType = "crown"
	// no expr rules and crown rules  at the same time.
	anyType hitType = "any"
)

func rearrangeMixedRulesWithPriority(exprRules []RuleFactory, crownRules []RuleFactory, priority []string) (
	reExprRules []RuleFactory, reCrownedRules []RuleFactory, typ hitType) {

	if len(exprRules) == 0 && len(crownRules) == 0 {
		return exprRules, crownRules, anyType
	}

	exprHitIndexes := make(map[int]bool)
	rearrangedExpr := make([]RuleFactory, 0)

	crownHitIndexes := make(map[int]bool)
	rearrangedCrown := make([]RuleFactory, 0)

	var firstHitType hitType
	for _, col := range priority {
		for idx, one := range exprRules {
			if col == one.RuleField() {
				rearrangedExpr = append(rearrangedExpr, one)
				exprHitIndexes[idx] = true

				if len(firstHitType) == 0 {
					firstHitType = exprType
				}
			}
		}

		for idx, one := range crownRules {
			if col == one.RuleField() {
				rearrangedCrown = append(rearrangedCrown, one)
				crownHitIndexes[idx] = true

				if len(firstHitType) == 0 {
					firstHitType = crownType
				}
			}
		}
	}

	// append the not hit index rules
	for idx, one := range exprRules {
		if exprHitIndexes[idx] {
			continue
		}

		rearrangedExpr = append(rearrangedExpr, one)
	}

	for idx, one := range crownRules {
		if crownHitIndexes[idx] {
			continue
		}

		rearrangedCrown = append(rearrangedCrown, one)
	}

	if len(exprRules) == 0 {
		firstHitType = crownType
	}

	if len(crownRules) == 0 {
		firstHitType = exprType
	}

	return rearrangedExpr, rearrangedCrown, firstHitType
}

// doMixedSQLWhereExpr generated mixed SQL WHERE expression with mixed priority rules.
func doMixedSQLWhereExpr(exprOp LogicOperator, exprRules []RuleFactory,
	crownOp LogicOperator, crownRules []RuleFactory, priority []string) (string, []interface{}, error) {

	var sqlSentence []string

	exprRules, crownRules, typ := rearrangeMixedRulesWithPriority(exprRules, crownRules, priority)

	exprExpr, argOp, err := genMixedSQLWhereExpr(exprOp, exprRules)
	if err != nil {
		return "", []interface{}{}, fmt.Errorf("gen mixed expr failed, %v", err)
	}

	// crowned rules is always operate with 'AND' logic operator.
	crownExpr, argAnd, err := genMixedSQLWhereExpr(And, crownRules)
	if err != nil {
		return "", []interface{}{}, fmt.Errorf("gen mixed crown expr failed, %v", err)
	}

	switch {
	case len(exprExpr) == 0 && len(crownExpr) == 0:
		// both is empty, return "" without prefixed 'WHERE'
		return "", []interface{}{}, nil

	case len(exprExpr) == 0 && len(crownExpr) != 0:
		// only have crowned rules, then return its expr and prefixed with 'WHERE'
		sqlSentence = append(sqlSentence, " WHERE ", crownExpr)
		return SqlJoint(sqlSentence), argAnd, nil

	case len(exprExpr) != 0 && len(crownExpr) == 0:
		// only have Expression rules, then return its expr and prefixed with 'WHERE'
		sqlSentence = append(sqlSentence, " WHERE ", exprExpr)
		return SqlJoint(sqlSentence), argOp, nil

	default:
		// generate SQL Where expression as follows:01
	}

	if exprOp == Or && crownOp == Or {
		// generate SQL where expression with mixed priority.
		switch typ {
		case exprType:
			// return fmt.Sprintf("WHERE %s %s (%s)", exprExpr, strings.ToUpper(string(crownOp)), crownExpr), nil
			argOp = append(argOp, argAnd...)
			sqlSentence = append(sqlSentence, " WHERE ", exprExpr, " ", strings.ToUpper(string(crownOp)), " (", crownExpr, ")")
			return SqlJoint(sqlSentence), argOp, nil
		case crownType:
			argAnd = append(argAnd, argOp...)
			// return fmt.Sprintf("WHERE (%s) %s %s", crownExpr, strings.ToUpper(string(crownOp)), exprExpr), nil
			sqlSentence = append(sqlSentence, " WHERE (", crownExpr, ") ", strings.ToUpper(string(crownOp)), " ", exprExpr)
			return SqlJoint(sqlSentence), argAnd, nil
		case anyType:
			// no expr rules and crown rules  at the same time.
			return "", []interface{}{}, nil
		default:
			return "", []interface{}{}, fmt.Errorf("unsupported expr type: %s", typ)
		}

	}

	// generate SQL where expression with mixed priority.
	switch typ {
	case exprType:
		argOp = append(argOp, argAnd...)
		// return fmt.Sprintf("WHERE %s %s (%s)", exprExpr, strings.ToUpper(string(crownOp)), crownExpr), nil
		sqlSentence = append(sqlSentence, " WHERE (", exprExpr, ") ", strings.ToUpper(string(crownOp)),
			" (", crownExpr, ")")
		return SqlJoint(sqlSentence), argOp, nil
	case crownType:
		argAnd = append(argAnd, argOp...)
		// return fmt.Sprintf("WHERE (%s) %s %s", crownExpr, strings.ToUpper(string(crownOp)), exprExpr), nil
		sqlSentence = append(sqlSentence, " WHERE (", crownExpr, ") ", strings.ToUpper(string(crownOp)),
			" (", exprExpr, ")")
		return SqlJoint(sqlSentence), argAnd, nil
	case anyType:
		// no expr rules and crown rules  at the same time.
		return "", []interface{}{}, nil
	default:
		return "", []interface{}{}, fmt.Errorf("unsupported expr type: %s", typ)
	}
}

func genMixedSQLWhereExpr(op LogicOperator, rules []RuleFactory) (string, []interface{}, error) {
	if len(rules) == 0 {
		return "", []interface{}{}, nil
	}

	// generate all the sub-expressions which is described by each rule.
	subExpr := make([]string, 0)
	argList := make([]interface{}, 0)
	for _, one := range rules {
		expr, arg, err := one.SQLExpr()
		if err != nil {
			return "", []interface{}{}, err
		}

		subExpr = append(subExpr, expr)
		argList = append(argList, arg...)
	}

	if len(subExpr) == 0 {
		return "", []interface{}{}, errors.New("invalid expression with 0 rules to query")
	}

	switch op {
	case And:
		return strings.Join(subExpr, " AND "), argList, nil

	case Or:
		return strings.Join(subExpr, " OR "), argList, nil

	default:
		return "", []interface{}{}, fmt.Errorf("unsupported expression's logic operator: %s", op)
	}
}

func doSoloSQLWhereExpr(op LogicOperator, rules []RuleFactory, priority []string) (
	where string, args []interface{}, err error) {

	var sqlSentence []string

	if len(rules) == 0 {
		return "", []interface{}{}, nil
	}

	// rearrange the rules with priority so that the query expression can
	// match the db indexes.
	rearrangedRules := rearrangeSoloRulesWithPriority(rules, priority)

	// generate all the sub-expressions which is described by each rule.
	subExpr := make([]string, 0)
	argList := make([]interface{}, 0)
	for _, one := range rearrangedRules {
		expr, arg, err := one.SQLExpr()
		if err != nil {
			return "", []interface{}{}, err
		}
		argList = append(argList, arg...)
		subExpr = append(subExpr, expr)
	}

	if len(subExpr) == 0 {
		return "", []interface{}{}, errors.New("invalid expression with 0 rules to query")
	}

	switch op {
	case And:
		sqlSentence = append(sqlSentence, " WHERE ", strings.Join(subExpr, " AND "))
		return SqlJoint(sqlSentence), argList, nil

	case Or:
		sqlSentence = append(sqlSentence, " WHERE ", strings.Join(subExpr, " OR "))
		return SqlJoint(sqlSentence), argList, nil

	default:
		return "", []interface{}{}, fmt.Errorf("unsupported expression's logic operator: %s", op)
	}
}

// rearrangeSoloRulesWithPriority rearrange the query rules with priority, the lower the
// index of the priority's array, the higher priority of the field during query.
func rearrangeSoloRulesWithPriority(rules []RuleFactory, priority []string) []RuleFactory {
	if len(priority) == 0 {
		return rules
	}

	arranged := make([]RuleFactory, 0)

	hitIndexes := make(map[int]bool, 0)
	for _, pri := range priority {
		// loop all the rules one by one to test if one of
		// it can match the priority field.
		for idx, one := range rules {
			if pri != one.RuleField() {
				continue
			}

			// this rule's filed matched the priority field,
			// then put it to the rules head.
			arranged = append(arranged, one)
			hitIndexes[idx] = true
			// Note:
			// do not break here, because a filed may occur
			// multiple times. such as '1< count <3'.
		}
	}

	// add the not matched rules to the tail.
	for idx := range rules {
		if hitIndexes[idx] {
			// this rule has already been put to head.
			continue
		}

		arranged = append(arranged, rules[idx])
	}

	return arranged
}

// RuleType is the expression rule's rule type.
type RuleType string

const (
	// EmptyType means the rules is empty
	EmptyType RuleType = "Empty"
	// AtomType means it's a AtomRule
	AtomType RuleType = "AtomRule"
	// UnknownType means it's an unknown type.
	UnknownType RuleType = "Unknown"
)

func ruleType(rules gjson.Result) (RuleType, error) {
	if !rules.IsArray() {
		return UnknownType, errors.New("rules should be an array")
	}

	if strings.TrimSpace(rules.Raw) == "[]" {
		return EmptyType, nil
	}

	allHit := true
	rules.ForEach(func(_, value gjson.Result) bool {
		parsed := gjson.GetMany(value.Raw, "field", "op", "value")
		if !parsed[0].Exists() || !parsed[1].Exists() || !parsed[2].Exists() {
			// if one of the field, op, value is not exist, then it's not a
			// valid AtomRule, then end the ForEach iterator.
			allHit = false
			return false
		}

		return true
	})

	if !allHit {
		return UnknownType, errors.New("invalid rules")
	}

	return AtomType, nil
}

// SqlJoint ..
func SqlJoint(sql []string) string {
	buff := bytes.NewBuffer([]byte{})
	for _, value := range sql {
		buff.WriteString(value)
	}
	return buff.String()
}
