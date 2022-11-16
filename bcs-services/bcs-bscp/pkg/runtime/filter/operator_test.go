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
	"testing"
)

func TestEqualSQLExpr(t *testing.T) {
	// test eq
	eq := EqualOp(Equal)
	eqExpr, err := eq.SQLExpr("name", "bscp")
	if err != nil {
		t.Errorf("test eq operator failed, err: %v", err)
		return
	}

	if eqExpr != `name = 'bscp'` {
		t.Errorf("test eq operator got wrong expr: %s", eqExpr)
		return
	}

}

func TestNotEqualSQLExpr(t *testing.T) {
	// test neq
	ne := NotEqualOp(NotEqual)
	neExpr, err := ne.SQLExpr("name", "bscp")
	if err != nil {
		t.Errorf("test ne operator failed, err: %v", err)
		return
	}

	if neExpr != `name != 'bscp'` {
		t.Errorf("test ne operator got wrong expr: %s", neExpr)
		return
	}

}

func TestGreaterThanSQLExpr(t *testing.T) {

	// test gt
	gt := GreaterThanOp(GreaterThan)
	gtExpr, err := gt.SQLExpr("count", 10)
	if err != nil {
		t.Errorf("test gt operator failed, err: %v", err)
		return
	}

	if gtExpr != `count > 10` {
		t.Errorf("test gt operator got wrong expr: %s", gtExpr)
		return
	}

	// test time scenario
	gtExpr, err = gt.SQLExpr("create_at", "2022-01-02 15:04:05")
	if err != nil {
		t.Errorf("test gt operator with time failed, err: %v", err)
		return
	}

	if gtExpr != `create_at > '2022-01-02 15:04:05'` {
		t.Errorf("test gt operator with time got wrong expr: %s", gtExpr)
		return
	}

}

func TestGreaterThanEqualSQLExpr(t *testing.T) {

	// test gte
	gte := GreaterThanEqualOp(GreaterThanEqual)
	gteExpr, err := gte.SQLExpr("count", 10)
	if err != nil {
		t.Errorf("test gte operator failed, err: %v", err)
		return
	}

	if gteExpr != `count >= 10` {
		t.Errorf("test gte operator got wrong expr: %s", gteExpr)
		return
	}

	// test with time scenario
	gteExpr, err = gte.SQLExpr("create_at", "2022-01-02 15:04:05")
	if err != nil {
		t.Errorf("test gte operator with time failed, err: %v", err)
		return
	}

	if gteExpr != `create_at >= '2022-01-02 15:04:05'` {
		t.Errorf("test gte operator with time got wrong expr: %s", gteExpr)
		return
	}

}

func TestLessThanSQLExpr(t *testing.T) {

	// test lt
	lt := LessThanOp(LessThan)
	ltExpr, err := lt.SQLExpr("count", 10)
	if err != nil {
		t.Errorf("test lt operator failed, err: %v", err)
		return
	}

	if ltExpr != `count < 10` {
		t.Errorf("test lt operator got wrong expr: %s", ltExpr)
		return
	}

	// test time scenario
	ltExpr, err = lt.SQLExpr("create_at", "2022-01-02 15:04:05")
	if err != nil {
		t.Errorf("test lt operator with time failed, err: %v", err)
		return
	}

	if ltExpr != `create_at < '2022-01-02 15:04:05'` {
		t.Errorf("test lt operator with time got wrong expr: %s", ltExpr)
		return
	}

}

func TestLessThanEqualSQLExpr(t *testing.T) {

	// test lte
	lte := LessThanEqualOp(LessThanEqual)
	lteExpr, err := lte.SQLExpr("count", 10)
	if err != nil {
		t.Errorf("test lte operator failed, err: %v", err)
		return
	}

	if lteExpr != `count <= 10` {
		t.Errorf("test lte operator got wrong expr: %s", lteExpr)
		return
	}

	// test time scenario
	lteExpr, err = lte.SQLExpr("create_at", "2022-01-02 15:04:05")
	if err != nil {
		t.Errorf("test lte operator with time failed, err: %v", err)
		return
	}

	if lteExpr != `create_at <= '2022-01-02 15:04:05'` {
		t.Errorf("test lte operator with time got wrong expr: %s", lteExpr)
		return
	}

}

func TestInSQLExpr(t *testing.T) {

	// test in
	in := InOp(In)
	sinExpr, err := in.SQLExpr("servers", []string{"api", "web"})
	if err != nil {
		t.Errorf("test in operator failed, err: %v", err)
		return
	}

	if sinExpr != `servers IN ('api', 'web')` {
		t.Errorf("test in operator got wrong expr: %s", sinExpr)
		return
	}

	intInExpr, err := in.SQLExpr("ages", []int{18, 30})
	if err != nil {
		t.Errorf("test in operator failed, err: %v", err)
		return
	}

	if intInExpr != `ages IN (18, 30)` {
		t.Errorf("test in operator got wrong expr: %s", sinExpr)
		return
	}

}

func TestNotInSQLExpr(t *testing.T) {

	// test nin
	nin := NotInOp(NotIn)
	sinExpr, err := nin.SQLExpr("servers", []string{"api", "web"})
	if err != nil {
		t.Errorf("test nin operator failed, err: %v", err)
		return
	}

	if sinExpr != `servers NOT IN ('api', 'web')` {
		t.Errorf("test nin operator got wrong expr: %s", sinExpr)
		return
	}

	intInExpr, err := nin.SQLExpr("ages", []int{18, 30})
	if err != nil {
		t.Errorf("test nin operator failed, err: %v", err)
		return
	}

	if intInExpr != `ages NOT IN (18, 30)` {
		t.Errorf("test nin operator got wrong expr: %s", sinExpr)
		return
	}

}

func TestContainsSensitiveSQLExpr(t *testing.T) {

	// test cs
	cs := ContainsSensitiveOp(ContainsSensitive)
	csExpr, err := cs.SQLExpr("name", "bscp-")
	if err != nil {
		t.Errorf("test cis operator failed, err: %v", err)
		return
	}

	if csExpr != `name LIKE BINARY '%bscp-%'` {
		t.Errorf("test cis operator got wrong expr: %s", csExpr)
		return
	}

}

func TestContainsInsensitiveSQLExpr(t *testing.T) {

	// test cis
	cis := ContainsInsensitiveOp(ContainsInsensitive)
	cisExpr, err := cis.SQLExpr("name", "bscp-")
	if err != nil {
		t.Errorf("test cis operator failed, err: %v", err)
		return
	}

	if cisExpr != `name LIKE '%bscp-%'` {
		t.Errorf("test cis operator got wrong expr: %s", cisExpr)
		return
	}

}
