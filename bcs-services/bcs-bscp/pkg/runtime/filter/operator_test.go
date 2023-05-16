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
	"fmt"
	"testing"
)

func TestEqualSQLExpr(t *testing.T) {
	// test eq
	eq := EqualOp(Equal)
	eqExpr, argList, err := eq.SQLExpr("name", "bscp")
	if err != nil {
		t.Errorf("test eq operator failed, err: %v", err)
		return
	}

	if eqExpr != `name = ?` {
		t.Errorf("test eq operator got wrong expr: %s", eqExpr)
		return
	}
	fmt.Println(argList)
}

func TestNotEqualSQLExpr(t *testing.T) {
	// test neq
	ne := NotEqualOp(NotEqual)
	neExpr, argList, err := ne.SQLExpr("name", "bscp")
	if err != nil {
		t.Errorf("test ne operator failed, err: %v", err)
		return
	}

	if neExpr != `name != ?` {
		t.Errorf("test ne operator got wrong expr: %s", neExpr)
		return
	}
	fmt.Println(argList)
}

func TestGreaterThanSQLExpr(t *testing.T) {

	// test gt
	gt := GreaterThanOp(GreaterThan)
	gtExpr, argList, err := gt.SQLExpr("count", 10)
	if err != nil {
		t.Errorf("test gt operator failed, err: %v", err)
		return
	}

	if gtExpr != `count > ?` {
		t.Errorf("test gt operator got wrong expr: %s", gtExpr)
		return
	}
	fmt.Println(argList)
	// test time scenario
	gtExpr, argList, err = gt.SQLExpr("create_at", "2022-01-02 15:04:05")
	if err != nil {
		t.Errorf("test gt operator with time failed, err: %v", err)
		return
	}

	if gtExpr != `create_at > ?` {
		t.Errorf("test gt operator with time got wrong expr: %s", gtExpr)
		return
	}
	fmt.Println(argList)
}

func TestGreaterThanEqualSQLExpr(t *testing.T) {

	// test gte
	gte := GreaterThanEqualOp(GreaterThanEqual)
	gteExpr, argList, err := gte.SQLExpr("count", 10)
	if err != nil {
		t.Errorf("test gte operator failed, err: %v", err)
		return
	}

	if gteExpr != `count >= ?` {
		t.Errorf("test gte operator got wrong expr: %s", gteExpr)
		return
	}
	fmt.Println(argList)
	// test with time scenario
	gteExpr, argList, err = gte.SQLExpr("create_at", "2022-01-02 15:04:05")
	if err != nil {
		t.Errorf("test gte operator with time failed, err: %v", err)
		return
	}

	if gteExpr != `create_at >= ?` {
		t.Errorf("test gte operator with time got wrong expr: %s", gteExpr)
		return
	}
	fmt.Println(argList)
}

func TestLessThanSQLExpr(t *testing.T) {

	// test lt
	lt := LessThanOp(LessThan)
	ltExpr, argList, err := lt.SQLExpr("count", 10)
	if err != nil {
		t.Errorf("test lt operator failed, err: %v", err)
		return
	}

	if ltExpr != `count < ?` {
		t.Errorf("test lt operator got wrong expr: %s", ltExpr)
		return
	}
	fmt.Println(argList)
	// test time scenario
	ltExpr, argList, err = lt.SQLExpr("create_at", "2022-01-02 15:04:05")
	if err != nil {
		t.Errorf("test lt operator with time failed, err: %v", err)
		return
	}

	if ltExpr != `create_at < ?` {
		t.Errorf("test lt operator with time got wrong expr: %s", ltExpr)
		return
	}
	fmt.Println(argList)
}

func TestLessThanEqualSQLExpr(t *testing.T) {

	// test lte
	lte := LessThanEqualOp(LessThanEqual)
	lteExpr, argList, err := lte.SQLExpr("count", 10)
	if err != nil {
		t.Errorf("test lte operator failed, err: %v", err)
		return
	}

	if lteExpr != `count <= ?` {
		t.Errorf("test lte operator got wrong expr: %s", lteExpr)
		return
	}
	fmt.Println(argList)
	// test time scenario
	lteExpr, argList, err = lte.SQLExpr("create_at", "2022-01-02 15:04:05")
	if err != nil {
		t.Errorf("test lte operator with time failed, err: %v", err)
		return
	}

	if lteExpr != `create_at <= ?` {
		t.Errorf("test lte operator with time got wrong expr: %s", lteExpr)
		return
	}
	fmt.Println(argList)
}

func TestInSQLExpr(t *testing.T) {

	// test in
	in := InOp(In)
	sinExpr, argList, err := in.SQLExpr("servers", []string{"api", "web"})
	if err != nil {
		t.Errorf("test in operator failed, err: %v", err)
		return
	}

	if sinExpr != `servers IN (?, ?)` {
		t.Errorf("test in operator got wrong expr: %s", sinExpr)
		return
	}
	fmt.Println(argList)
	intInExpr, argList, err := in.SQLExpr("ages", []int{18, 30})
	if err != nil {
		t.Errorf("test in operator failed, err: %v", err)
		return
	}

	if intInExpr != `ages IN (?, ?)` {
		t.Errorf("test in operator got wrong expr: %s", sinExpr)
		return
	}
	fmt.Println(argList)
}

func TestNotInSQLExpr(t *testing.T) {

	// test nin
	nin := NotInOp(NotIn)
	sinExpr, argList, err := nin.SQLExpr("servers", []string{"api", "web"})
	if err != nil {
		t.Errorf("test nin operator failed, err: %v", err)
		return
	}

	if sinExpr != `servers NOT IN (?, ?)` {
		t.Errorf("test nin operator got wrong expr: %s", sinExpr)
		return
	}
	fmt.Println(argList)
	intInExpr, argList, err := nin.SQLExpr("ages", []int{18, 30})
	if err != nil {
		t.Errorf("test nin operator failed, err: %v", err)
		return
	}

	if intInExpr != `ages NOT IN (?, ?)` {
		t.Errorf("test nin operator got wrong expr: %s", sinExpr)
		return
	}
	fmt.Println(argList)
}

func TestContainsSensitiveSQLExpr(t *testing.T) {

	// test cs
	cs := ContainsSensitiveOp(ContainsSensitive)
	csExpr, argList, err := cs.SQLExpr("name", "bscp-")
	if err != nil {
		t.Errorf("test cis operator failed, err: %v", err)
		return
	}

	if csExpr != `name LIKE BINARY %?%` {
		t.Errorf("test cis operator got wrong expr: %s", csExpr)
		return
	}
	fmt.Println(argList)
}

func TestContainsInsensitiveSQLExpr(t *testing.T) {

	// test cis
	cis := ContainsInsensitiveOp(ContainsInsensitive)
	cisExpr, argList, err := cis.SQLExpr("name", "bscp-")
	if err != nil {
		t.Errorf("test cis operator failed, err: %v", err)
		return
	}

	if cisExpr != `name LIKE %?%` {
		t.Errorf("test cis operator got wrong expr: %s", cisExpr)
		return
	}
	fmt.Println(argList)
}
