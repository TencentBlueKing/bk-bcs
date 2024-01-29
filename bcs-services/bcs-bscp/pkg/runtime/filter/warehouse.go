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
	"bytes"
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// isNumericOrTime test if a value is a standard time
// format string or a numeric value.
func isNumericOrTime(v interface{}) (bool, bool) { //nolint:unparam
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
