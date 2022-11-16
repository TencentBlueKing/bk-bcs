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

package types

import (
	"testing"
)

func TestBasePageCount(t *testing.T) {
	bp := &BasePage{
		Count: true,
		Start: 1,
		Order: "ASC",
	}

	err := bp.Validate()
	if err == nil {
		t.Error("validate base page, but got invalid validate, an error should be occurred.")
		return
	}

	bp = &BasePage{
		Count: true,
		Sort:  "id",
	}

	err = bp.Validate()
	if err == nil {
		t.Error("validate base page, but got invalid validate, an error should be occurred.")
		return
	}

	bp = &BasePage{
		Count: true,
		Order: "ASC",
	}

	err = bp.Validate()
	if err == nil {
		t.Error("validate base page, but got invalid validate, an error should be occurred.")
		return
	}

	bp = &BasePage{
		Count: true,
		Limit: 10,
	}

	err = bp.Validate()
	if err == nil {
		t.Error("validate base page, but got invalid validate, an error should be occurred.")
		return
	}

}

func TestBasePageOption(t *testing.T) {
	bp := BasePage{
		Count: false,
	}

	// test limit
	bp.Limit = 300
	opt := DefaultPageOption
	// opt := &PageOption{
	// 	EnableUnlimitedLimit: false,
	// 	MaxLimit:             0,
	// 	DisabledSort:         false,
	// }

	err := bp.Validate(opt)
	if err == nil {
		t.Error("validate base page limit, an over limit error should occur, but not.")
		return
	}
	bp.Limit = 0

	bp.Start = 0
	err = bp.Validate(opt)
	if err == nil {
		t.Error("validate base page limit, an limit should > 0 error should occur, but not.")
		return
	}

	opt.EnableUnlimitedLimit = true
	err = bp.Validate(opt)
	if err != nil {
		t.Errorf("validate base page limit, test query all scenario failed. err: %v", err)
		return
	}

	bp.Limit = 300
	opt.MaxLimit = 500
	err = bp.Validate(opt)
	if err != nil {
		t.Errorf("validate base page limit, test config max limit scenario failed. err: %v", err)
		return
	}

	// test sort
	opt.DisabledSort = true
	bp.Sort = "self-defined-sort"
	err = bp.Validate(opt)
	if err == nil {
		t.Errorf("validate base page limit, test disable user configed sort scenario failed. err: %v", err)
		return
	}

	opt.DisabledSort = false
	bp.Sort = ""
	err = bp.Validate(opt)
	if err != nil {
		t.Errorf("validate base page limit, test disable user configed sort scenario failed. err: %v", err)
		return
	}

}

func TestBasePageSQLExpr(t *testing.T) {
	bp := BasePage{
		Count: true,
		Start: 0,
		Limit: 0,
		Sort:  "memo",
		Order: "",
	}

	opt := &PageSQLOption{
		Sort: SortOption{},
	}

	_, err := bp.SQLExpr(opt)
	if err == nil {
		t.Errorf("test generate page SQL expression failed, an count error should occur")
		return
	}

	// test query all SQL
	bp.Count = false
	expr, err := bp.SQLExpr(opt)
	if err != nil {
		t.Errorf("test generate page SQL expression failed, err: %v", err)
		return
	}

	if expr != "ORDER BY memo ASC" {
		t.Errorf("test generate page SQL expression failed, unexpected SQL expr: %s", expr)
		return
	}

	// test if not present sort scenario
	opt.Sort = SortOption{
		Sort:         "name",
		IfNotPresent: true,
	}
	expr, err = bp.SQLExpr(opt)
	if err != nil {
		t.Errorf("test generate page SQL expression with overlapped sort failed, err: %v", err)
		return
	}

	if expr != "ORDER BY memo ASC" {
		t.Errorf("test generate page SQL expression with overlapped sort failed, unexpected SQL expr: %s", expr)
		return
	}

	bp.Sort = ""
	opt.Sort = SortOption{
		Sort:         "name",
		IfNotPresent: true,
	}
	expr, err = bp.SQLExpr(opt)
	if err != nil {
		t.Errorf("test generate page SQL expression with overlapped sort failed, err: %v", err)
		return
	}

	if expr != "ORDER BY name ASC" {
		t.Errorf("test generate page SQL expression with overlapped sort failed, unexpected SQL expr: %s", expr)
		return
	}

	// test force overlapped sort scenario
	bp.Sort = "memo"
	opt.Sort = SortOption{
		Sort:         "name",
		ForceOverlap: true,
	}
	expr, err = bp.SQLExpr(opt)
	if err != nil {
		t.Errorf("test generate page SQL expression with overlapped sort failed, err: %v", err)
		return
	}

	if expr != "ORDER BY name ASC" {
		t.Errorf("test generate page SQL expression with overlapped sort failed, unexpected SQL expr: %s", expr)
		return
	}

	// test query with start and limit
	bp.Start = 100
	bp.Limit = 50
	expr, err = bp.SQLExpr(opt)
	if err != nil {
		t.Errorf("test generate page SQL expression with start and limit failed, err: %v", err)
		return
	}

	if expr != "ORDER BY name ASC LIMIT 50 OFFSET 100" {
		t.Errorf("test generate page SQL expression with start and limit failed, unexpected SQL expr: %s", expr)
		return
	}

	// test direction
	bp.Order = "DESC"
	expr, err = bp.SQLExpr(opt)
	if err != nil {
		t.Errorf("test generate page SQL expression with direction failed, err: %v", err)
		return
	}

	if expr != "ORDER BY name DESC LIMIT 50 OFFSET 100" {
		t.Errorf("test generate page SQL expression with direction failed, unexpected SQL expr: %s", expr)
		return
	}

}
