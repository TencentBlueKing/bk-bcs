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

package table

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/enumor"
)

func TestMergeNamedColumns(t *testing.T) {
	namedA := ColumnDescriptors{{Column: "id", NamedC: "id", Type: enumor.Numeric}}
	namedB := mergeColumnDescriptors("nested",
		ColumnDescriptors{
			{Column: "name", NamedC: "name", Type: enumor.String},
			{Column: "memo", NamedC: "memo", Type: enumor.String}},
	)

	merged := mergeColumns(namedA, namedB)
	compared := mergeColumns(ColumnDescriptors{
		{Column: "id", NamedC: "id", Type: enumor.Numeric},
		{Column: "name", NamedC: "nested.name", Type: enumor.String},
		{Column: "memo", NamedC: "nested.memo", Type: enumor.String},
	})

	fmt.Println("columns: ", merged.Columns())
	if !reflect.DeepEqual(merged.Columns(), compared.Columns()) {
		t.Errorf("test merged columns failed, not equal")
		return
	}

	fmt.Println("column expr: ", merged.ColumnExpr())
	if merged.ColumnExpr() != compared.ColumnExpr() {
		t.Errorf("test merged columns expr failed, not equal")
		return
	}

	fmt.Println("named columns expr: ", merged.NamedExpr())
	if merged.NamedExpr() != compared.NamedExpr() {
		t.Errorf("test merged columns named expr failed,  not equal")
		return
	}

	fmt.Println("colon named columns expr: ", merged.ColonNameExpr())
	if merged.ColonNameExpr() != compared.ColonNameExpr() {
		t.Errorf("test merged columns named expr failed,  not equal")
		return
	}

	fmt.Println("without column: ", merged.WithoutColumn("id"))
	if !reflect.DeepEqual(merged.WithoutColumn("id"), map[string]enumor.ColumnType{"name": enumor.String,
		"memo": enumor.String}) {
		t.Errorf("test merged without columns failed,  not equal")
		return
	}
}
