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

package orm

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
)

func TestRearrangeSQLDataWithOption(t *testing.T) {
	app := &table.App{
		ID:    1,
		BizID: 0,
		Spec: &table.AppSpec{
			Name:       "demo",
			ConfigType: table.File,
			Memo:       "",
		},
		Revision: &table.Revision{
			Creator:   "bscp",
			CreatedAt: time.Now(),
		},
	}

	opts := NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields("id")
	expr, toUpdate, err := RearrangeSQLDataWithOption(app, opts)
	if err != nil {
		t.Errorf("parse data field, err: %v", err)
		return
	}

	fmt.Println("expr: ", expr)
	js, _ := json.MarshalIndent(toUpdate, "", "    ")
	fmt.Printf("to update: %s\n", js)
}

type deepEmbedded struct {
	Age      int       `db:"age"`
	Birthday time.Time `db:"birthday"`
}

type embedded struct {
	Name string `db:"name"`
	// test not pointer struct case.
	Deep deepEmbedded `db:"deep"`
}

type cases struct {
	// test flat field
	ID int `db:"id"`
	// test embedded struct
	Embedded *embedded `db:"embedded"`
	// test interface
	Iter interface{} `db:"iter"`
	// test time.
	Time time.Time `db:"time"`
}

func TestRearrangeSQLDataWithOptionFully(t *testing.T) {

	c := cases{
		// test ignored field
		ID: 20,
		Embedded: &embedded{
			Name: "demo",
			Deep: deepEmbedded{
				// test blanked field
				Age: 0,
				// test not blanked field.
				Birthday: time.Time{},
			},
		},
		Iter: "123456789",
		Time: time.Now(),
	}

	opts := NewFieldOptions().
		AddBlankedFields("age").
		AddIgnoredFields("id")
	expr, toUpdate, err := RearrangeSQLDataWithOption(c, opts)
	if err != nil {
		t.Errorf("parse data field, err: %v", err)
		return
	}

	// validate result
	if strings.Contains(expr, ":id") {
		t.Errorf("test ignored field failed")
		return
	}

	if !strings.Contains(expr, ":name") {
		t.Errorf("test embedded *struct field failed")
		return
	}

	if !strings.Contains(expr, ":age") {
		t.Errorf("test deep embedded *struct and blanked field failed")
		return
	}

	if strings.Contains(expr, ":birthday") {
		t.Errorf("test deep embedded *struct and *NOT* blanked field failed")
		return
	}

	if !strings.Contains(expr, ":iter") {
		t.Errorf("test interface field failed")
		return
	}

	if !strings.Contains(expr, ":time") {
		t.Errorf("test not blanked time field failed")
		return
	}

	fmt.Println("expr: ", expr)
	js, _ := json.MarshalIndent(toUpdate, "", "    ")
	fmt.Printf("to update: %s\n", js)

	// test result should be like this:
	// expr:  iter = :iter, time = :time, name = :name, age = :age
	// to update: {
	//    "age": 0,
	//    "birthday": "0001-01-01T00:00:00Z",
	//    "id": 20,
	//    "iter": "123456789",
	//    "name": "demo",
	//    "time": "2022-01-02T11:00:53.692535+08:00"
	// }
}

func TestRecursiveGetTaggedFieldValues(t *testing.T) {
	c := cases{
		// test ignored field
		ID: 20,
		Embedded: &embedded{
			Name: "demo",
			Deep: deepEmbedded{
				// test blanked field
				Age: 0,
				// test not blanked field.
				Birthday: time.Time{},
			},
		},
		Iter: "123456789",
		Time: time.Now(),
	}

	kv, err := RecursiveGetTaggedFieldValues(c)
	if err != nil {
		t.Errorf("get kv failed, err: %v", err)
		return
	}

	js, _ := json.MarshalIndent(kv, "", "    ")
	fmt.Printf("kv json: %s\n", js)

}

func TestGetTaggedDbField(t *testing.T) {
	now := time.Now()

	app := &table.App{
		ID:    1,
		BizID: 0,
		Spec: &table.AppSpec{
			Name:       "demo",
			ConfigType: table.File,
			Memo:       "",
		},
		Revision: &table.Revision{
			Creator:   "bscp",
			CreatedAt: now,
		},
	}

	kv, err := RecursiveGetTaggedFieldValues(app)
	if err != nil {
		t.Errorf("recursively get tagged db field failed, err: %v", err)
		return
	}

	// test out layer, not embedded.
	if !reflect.DeepEqual(kv["id"], uint32(1)) {
		t.Errorf("recursively get tagged db id field failed, not equal")
		return
	}

	// test embedded string
	if !reflect.DeepEqual(kv["name"], "demo") {
		t.Errorf("recursively get tagged db name field failed, not equal")
		return
	}

	// test embedded time
	if !reflect.DeepEqual(kv["create_time"], now) {
		t.Errorf("recursively get tagged db create_name field failed, not equal")
		return
	}

	// test embedded empty time
	if !reflect.DeepEqual(kv["update_time"], time.Time{}) {
		t.Errorf("recursively get tagged db update_name field failed, not equal")
		return
	}

	js, _ := json.MarshalIndent(kv, "", "    ")
	fmt.Printf("kv: %s\n", js)
}

func TestRecursiveGetNestedNamedTags(t *testing.T) {

	c := &cases{
		// test ignored field
		ID: 20,
		Embedded: &embedded{
			Name: "demo",
			Deep: deepEmbedded{
				// test blanked field
				Age: 0,
				// test not blanked field.
				Birthday: time.Time{},
			},
		},
		Iter: "123456789",
		Time: time.Now(),
	}

	namedTags, err := recursiveGetNestedNamedTags(c)
	if err != nil {
		t.Errorf("get nested tags failed, err: %v", err)
		return
	}

	// validate the result
	v, exist := namedTags["age"]
	if !exist {
		t.Error("test embedded age failed")
		return
	}

	if v != "embedded.deep.age" {
		t.Error("test embedded age failed")
		return
	}

	v, exist = namedTags["id"]
	if !exist {
		t.Error("test flat id failed")
		return
	}

	if v != "" {
		t.Error("test flat id failed")
		return
	}

	js, _ := json.MarshalIndent(namedTags, "", "    ")
	fmt.Println(string(js))
	// test result should be like this:
	// {
	//    "age": "embedded.deep.age",
	//    "birthday": "embedded.deep.birthday",
	//    "id": "",
	//    "iter": "",
	//    "name": "embedded.name",
	//    "time": ""
	// }
}

func TestGetNamedSelectExpr(t *testing.T) {
	c := &cases{
		// test ignored field
		ID: 20,
		Embedded: &embedded{
			Name: "demo",
			Deep: deepEmbedded{
				// test blanked field
				Age: 0,
				// test not blanked field.
				Birthday: time.Time{},
			},
		},
		Iter: "123456789",
		Time: time.Now(),
	}

	expr, err := GetNamedSelectColumns(c)
	if err != nil {
		t.Errorf("test get named select expr failed, err: %v", err)
		return
	}

	if !strings.Contains(expr, `age as 'embedded.deep.age'`) {
		t.Errorf("get named age expr failed, expr: %s", expr)
		return
	}

	if !strings.Contains(expr, "time") || strings.Contains(expr, "time as") {
		t.Errorf("get named time expr failed, expr: %s", expr)
		return
	}

	if !strings.Contains(expr, "iter") || strings.Contains(expr, "iter as") {
		t.Errorf("get named iter expr failed, expr: %s", expr)
		return
	}

	if !strings.Contains(expr, "id") || strings.Contains(expr, "id as") {
		t.Errorf("get named id expr failed, expr: %s", expr)
		return
	}

	fmt.Println(expr)
	// test result should be like:
	// iter, time, id, name as 'embedded.name', age as 'embedded.deep.age', birthday as 'embedded.deep.birthday'
}

func TestRecursiveGetDBTags(t *testing.T) {
	c := &cases{
		// test ignored field
		ID: 20,
		Embedded: &embedded{
			Name: "demo",
			Deep: deepEmbedded{
				// test blanked field
				Age: 0,
				// test not blanked field.
				Birthday: time.Time{},
			},
		},
		Iter: "123456789",
		Time: time.Now(),
	}

	tags, err := RecursiveGetDBTags(c)
	if err != nil {
		t.Errorf("test get named db tags failed, err: %v", err)
		return
	}

	if len(tags) != 6 {
		t.Error("test get named db tags failed, not enough")
		return
	}

	fmt.Println(tags)
	// test result should be like:
	// [id name age birthday iter time]
}
