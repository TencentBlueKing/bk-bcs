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
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"bscp.io/pkg/criteria/enumor"
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

func TestExpressionAnd(t *testing.T) {
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
		},
	}

	if err := expr.Validate(); err != nil {
		t.Errorf("validate expression failed, err: %v", err)
		return
	}

	opt := &SQLWhereOption{Priority: []string{"servers", "age", "name"}}
	sql, arg, err := expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate expression's sql where expression failed, err: %v", err)
		return
	}

	if sql != ` WHERE servers IN (?, ?) AND age > ? AND age < ? AND name = ?` {
		t.Errorf("expression's sql where is not expected, sql: %s", sql)
		return
	}
	for i := 0; i < len(arg); i++ {
		fmt.Println(arg[i])
	}

}

func TestExpressionOr(t *testing.T) {
	expr := &Expression{
		Op: Or,
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
		},
	}

	js, err := json.MarshalIndent(expr, "", "    ")
	if err != nil {
		t.Errorf("test expression failed, err: %v", err)
		return
	}

	fmt.Println(string(js))

	if err = expr.Validate(); err != nil {
		t.Errorf("validate expression failed, err: %v", err)
		return
	}

	opt := &SQLWhereOption{Priority: []string{"servers", "age", "name"}}
	sql, arg, err := expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate expression's sql where expression failed, err: %v", err)
		return
	}

	if sql != ` WHERE servers IN (?, ?) OR age > ? OR age < ? OR name = ?` {
		t.Errorf("expression's sql where is not expected, sql: %s", sql)
		return
	}

	fmt.Println(sql)
	fmt.Println(arg)
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

func TestCrownSQLWhereExpr(t *testing.T) {
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
		},
	}

	opt := &SQLWhereOption{
		Priority: []string{"biz_id", "age"},
		CrownedOption: &CrownedOption{
			CrownedOp: And,
			Rules: []RuleFactory{
				&AtomRule{
					Field: "biz_id",
					Op:    "eq",
					Value: 20,
				},
				&AtomRule{
					Field: "created_at",
					Op:    "gt",
					Value: "2021-01-01 08:09:10",
				},
			},
		},
	}

	where, arg, err := expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate SQL Where expression failed, err: %v", err)
		//return
	}

	fmt.Println("where AND-AND expr: ", where)
	if where != ` WHERE biz_id = ? AND age > ? AND name = ? AND created_at > ?` {
		t.Errorf("generate SQL AND-AND Where expression failed, err: %v", err)
		//return
	}
	fmt.Println("arg :", arg)
	expr.Op = And
	opt.CrownedOption.CrownedOp = Or
	where, arg, err = expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate SQL Where AND-OR expression failed, err: %v", err)
		//return
	}

	fmt.Println("where AND-OR expr: ", where)
	if where != ` WHERE (biz_id = ? AND created_at > ?) OR (age > ? AND name = ?)` {
		t.Errorf("generate SQL AND-OR Where expression failed, where: %v", where)
		//return
	}
	fmt.Println("and", arg)
	expr.Op = Or
	opt.CrownedOption.CrownedOp = Or
	where, arg, err = expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate SQL Where OR-OR expression failed, err: %v", err)
		return
	}

	fmt.Println("where OR-OR expr: ", where)
	if where != ` WHERE (biz_id = ? AND created_at > ?) OR age > ? OR name = ?` {
		t.Errorf("generate SQL OR-OR Where expression failed, where: %v", where)
		//return
	}
	fmt.Println("or", arg)
	opt.Priority = []string{"age", "biz_id"}
	where, arg, err = expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate SQL Where OR-OR expression failed, err: %v", err)
		return
	}

	// reverse the priority
	fmt.Println("where OR-OR-PRIORITY expr: ", where)
	if where != ` WHERE age > ? OR name = ? OR (biz_id = ? AND created_at > ?)` {
		t.Errorf("generate SQL OR-OR-PRIORITY Where expression failed, where: %v", where)
		//return
	}
	fmt.Println("aa", arg)
	expr.Op = Or
	opt.CrownedOption.CrownedOp = And
	opt.Priority = []string{"biz_id", "age"}
	where, arg, err = expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate SQL Where OR-AND expression failed, err: %v", err)
		return
	}

	fmt.Println("where OR-AND expr: ", where)
	if where != ` WHERE (biz_id = ? AND created_at > ?) AND (age > ? OR name = ?)` {
		t.Errorf("generate SQL OR-AND Where expression failed, where: %v", where)
		//return
	}
	fmt.Println("aaa", arg)
	// test NULL crown rules
	expr.Rules = []RuleFactory{
		&AtomRule{Field: "name", Op: "eq", Value: "bscp"},
		&AtomRule{Field: "age", Op: "gt", Value: 18}}
	opt.CrownedOption.Rules = make([]RuleFactory, 0)
	where, arg, err = expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate SQL Where NULL crown expr expression failed, err: %v", err)
		return
	}

	fmt.Println("where crown NULL expr: ", where)
	if where != ` WHERE age > ? OR name = ?` {
		t.Errorf("generate SQL Where NULL crown expr expression failed, where: %s", where)
		//return
	}
	fmt.Println("aaaa", arg)
	// test NULL Expression rules
	expr.Rules = make([]RuleFactory, 0)
	opt.CrownedOption.Rules = []RuleFactory{&AtomRule{
		Field: "age",
		Op:    "eq",
		Value: 8,
	}}
	where, arg, err = expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate SQL Where NULL crown expr expression failed, err: %v", err)
		return
	}

	fmt.Println("where crown NULL expr: ", where)
	if where != ` WHERE age = ?` {
		t.Errorf("generate SQL Where NULL crown expr expression failed, where: %s", where)
		//return
	}
	fmt.Println("aaaaa", arg)
	// test both Expression and crown rules is empty
	expr.Rules = make([]RuleFactory, 0)
	opt.CrownedOption.Rules = make([]RuleFactory, 0)
	where, arg, err = expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate SQL Where both rule is NULL failed, err: %v", err)
		return
	}

	fmt.Println("where both NULL expr: ", where)
	if where != "" {
		t.Errorf("generate SQL Where both rule is NULL failed, where: %s", where)
		//return
	}
	fmt.Println("bb", arg)
}
