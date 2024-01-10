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
	"strings"
	"testing"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/enumor"
)

func TestUnmarshal(t *testing.T) {
	exprJson := `
{
	"op": "and",
	"rules": [{
			"field": "deploy_type",
			"op": "eq",
			"value": "common"
		},
		{
			"field": "creator",
			"op": "eq",
			"value": "tom"
		}
	]
}
`
	expr := new(Expression)
	err := expr.UnmarshalJSON([]byte(exprJson))
	if err != nil {
		t.Error(err)
		return
	}

	if !(expr.Op == And && len(expr.Rules) == 2 && (expr.Rules[0].(*AtomRule).Field == "deploy_type" &&
		expr.Rules[0].(*AtomRule).Op == "eq" && expr.Rules[0].(*AtomRule).Value == "common") &&
		(expr.Rules[1].(*AtomRule).Field == "creator" && expr.Rules[1].(*AtomRule).Op == "eq" &&
			expr.Rules[1].(*AtomRule).Value == "tom")) {
		t.Errorf("expression is not expected, op: %s, rules[0]: %v, rules[1]: %v", expr.Op,
			expr.Rules[0].(*AtomRule), expr.Rules[1].(*AtomRule))
		return
	}
}

func TestExpressionValidateOption(t *testing.T) {
	expr := &Expression{
		Op: And,
		Rules: []RuleFactory{
			&AtomRule{
				Field: "name",
				Op:    "eq",
				Value: "bscp",
			},
			&AtomRule{
				Field: "age",
				Op:    "gt",
				Value: 18,
			},
			&AtomRule{
				Field: "age",
				Op:    "lt",
				Value: 30,
			},
			&AtomRule{
				Field: "servers",
				Op:    "in",
				Value: []string{"api", "web"},
			},
			&AtomRule{
				Field: "asDefault",
				Op:    "eq",
				Value: true,
			},
			&AtomRule{
				Field: "created_at",
				Op:    "gt",
				Value: "2006-01-02 15:04:05",
			},
		},
	}

	opt := &ExprOption{
		RuleFields: map[string]enumor.ColumnType{
			"name":       enumor.String,
			"age":        enumor.Numeric,
			"servers":    enumor.String,
			"asDefault":  enumor.Boolean,
			"created_at": enumor.Time,
		},
		MaxInLimit:    0,
		MaxNotInLimit: 0,
		MaxRulesLimit: 10,
	}

	if err := expr.Validate(opt); err != nil {
		t.Errorf("validate expression failed, err: %v", err)
		return
	}

	// test invalidate scenario
	opt.RuleFields["name"] = enumor.Numeric
	if err := expr.Validate(opt); !strings.Contains(err.Error(), "value should be a numeric") {
		t.Errorf("validate numeric type failed, err: %v", err)
		return
	}
	opt.RuleFields["name"] = enumor.String

	opt.RuleFields["age"] = enumor.String
	if err := expr.Validate(opt); !strings.Contains(err.Error(), "value should be a string") {
		t.Errorf("validate string type failed, err: %v", err)
		return
	}
	opt.RuleFields["age"] = enumor.Numeric

	opt.RuleFields["asDefault"] = enumor.Time
	if err := expr.Validate(opt); !strings.Contains(err.Error(), "value should be a string time format") {
		t.Errorf("validate time type failed, err: %v", err)
		return
	}
	opt.RuleFields["asDefault"] = enumor.Boolean

	opt.RuleFields["created_at"] = enumor.Boolean
	if err := expr.Validate(opt); !strings.Contains(err.Error(), "value should be a boolean") {
		t.Errorf("validate boolean type failed, err: %v", err)
		return
	}

}
