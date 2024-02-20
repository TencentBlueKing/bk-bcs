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

// Package selector provides instance selector related structs and functions.
//
//nolint:goconst
package selector

import (
	"encoding/json"
	"testing"

	pbstruct "github.com/golang/protobuf/ptypes/struct"
)

func TestUnmarshalEqualElement(t *testing.T) {

	const eqJSON = `
	{
		"key": "bscp.biz",
		"op": "eq",
		"value": "lol"
	}`

	eqElement := new(Element)
	if err := json.Unmarshal([]byte(eqJSON), eqElement); err != nil {
		t.Errorf("test eq operator, failed, err: %v", err)
		return
	}

	if eqElement.Key != "bscp.biz" {
		t.Errorf("test eq operator, invalid key: %v", eqElement.Key)
		return
	}

	if eqElement.Op != &EqualOperator {
		t.Errorf("test eq operator, invalid op: %v", eqElement.Op)
		return
	}

	if eqElement.Value != "lol" {
		t.Errorf("test eq operator, invalid value: %v", eqElement.Value)
		return
	}

	labels := map[string]string{
		"bscp.biz": "lol",
	}

	matched, err := eqElement.Match(labels)
	if err != nil {
		t.Errorf("test eq operator, match failed, err: %v", err)
		return
	}

	if !matched {
		t.Error("test eq operator, not matched")
		return
	}

}

func TestUnmarshalNotEqualElement(t *testing.T) {

	const neJSON = `
	{
		"key": "bscp.biz",
		"op": "ne",
		"value": "lol"
	}`

	neElement := new(Element)
	if err := json.Unmarshal([]byte(neJSON), neElement); err != nil {
		t.Errorf("test ne operator, failed, err: %v", err)
		return
	}

	if neElement.Key != "bscp.biz" {
		t.Errorf("test ne operator, invalid key: %v", neElement.Key)
		return
	}

	if neElement.Op != &NotEqualOperator {
		t.Errorf("test ne operator, invalid op: %v", neElement.Op)
		return
	}

	if neElement.Value != "lol" {
		t.Errorf("test ne operator, invalid value: %v", neElement.Value)
		return
	}

	labels := map[string]string{
		"bscp.biz": "not-lol",
	}

	matched, err := neElement.Match(labels)
	if err != nil {
		t.Errorf("test ne operator, match failed, err: %v", err)
		return
	}

	if !matched {
		t.Error("test ne operator, but matched")
		return
	}

}

func TestUnmarshalGreaterThanElement(t *testing.T) {

	const gtJSON = `
	{
		"key": "bscp.qps",
		"op": "gt",
		"value": 10000
	}`

	gtElement := new(Element)
	if err := json.Unmarshal([]byte(gtJSON), gtElement); err != nil {
		t.Errorf("test gt operator, failed, err: %v", err)
		return
	}

	if gtElement.Key != "bscp.qps" {
		t.Errorf("test gt operator, invalid key: %v", gtElement.Key)
		return
	}

	if gtElement.Op != &GreaterThanOperator {
		t.Errorf("test gt operator, invalid op: %v", gtElement.Op)
		return
	}

	if mustFloat64(gtElement.Value) != float64(10000) {
		t.Errorf("test gt operator, invalid value: %v", gtElement.Value)
		return
	}

	labels := map[string]string{
		"bscp.qps": "20000",
	}

	matched, err := gtElement.Match(labels)
	if err != nil {
		t.Errorf("test gt operator, match failed, err: %v", err)
		return
	}

	if !matched {
		t.Error("test gt operator, but matched")
		return
	}

}

// nolint
// NOCC:tosa/fn_length(ignore)
func TestUnmarshalGreaterThanEqualElement(t *testing.T) {

	const geJSON = `
	{
		"key": "bscp.qps",
		"op": "ge",
		"value": 10000
	}`

	geElement := new(Element)
	if err := json.Unmarshal([]byte(geJSON), geElement); err != nil {
		t.Errorf("test ge operator, failed, err: %v", err)
		return
	}

	if geElement.Key != "bscp.qps" {
		t.Errorf("test ge operator, invalid key: %v", geElement.Key)
		return
	}

	if geElement.Op != &GreaterThanEqualOperator {
		t.Errorf("test ge operator, invalid op: %v", geElement.Op)
		return
	}

	if mustFloat64(geElement.Value) != float64(10000) {
		t.Errorf("test ge operator, invalid value: %v", geElement.Value)
		return
	}

	// test >
	labels := map[string]string{
		"bscp.qps": "20000",
	}

	matched, err := geElement.Match(labels)
	if err != nil {
		t.Errorf("test ge operator, match failed, err: %v", err)
		return
	}

	if !matched {
		t.Error("test ge operator, but matched")
		return
	}

	// test =
	labels["bscp.qps"] = "10000"

	matched, err = geElement.Match(labels)
	if err != nil {
		t.Errorf("test ge operator, match failed, err: %v", err)
		return
	}

	if !matched {
		t.Error("test ge operator, but matched")
		return
	}

}

func TestUnmarshalLessThanElement(t *testing.T) {

	const ltJSON = `
	{
		"key": "bscp.qps",
		"op": "lt",
		"value": 10000
	}`

	ltElement := new(Element)
	if err := json.Unmarshal([]byte(ltJSON), ltElement); err != nil {
		t.Errorf("test lt operator, failed, err: %v", err)
		return
	}

	if ltElement.Key != "bscp.qps" {
		t.Errorf("test lt operator, invalid key: %v", ltElement.Key)
		return
	}

	if ltElement.Op != &LessThanOperator {
		t.Errorf("test lt operator, invalid op: %v", ltElement.Op)
		return
	}

	if mustFloat64(ltElement.Value) != float64(10000) {
		t.Errorf("test lt operator, invalid value: %v", ltElement.Value)
		return
	}

	labels := map[string]string{
		"bscp.qps": "9000",
	}

	matched, err := ltElement.Match(labels)
	if err != nil {
		t.Errorf("test lt operator, match failed, err: %v", err)
		return
	}

	if !matched {
		t.Error("test lt operator, but matched")
		return
	}
}

func TestUnmarshalLessThanEqualElement(t *testing.T) {

	const leJSON = `
	{
		"key": "bscp.qps",
		"op": "le",
		"value": 10000
	}`

	leElement := new(Element)
	if err := json.Unmarshal([]byte(leJSON), leElement); err != nil {
		t.Errorf("test le operator, failed, err: %v", err)
		return
	}

	if leElement.Key != "bscp.qps" {
		t.Errorf("test le operator, invalid key: %v", leElement.Key)
		return
	}

	if leElement.Op != &LessThanEqualOperator {
		t.Errorf("test le operator, invalid op: %v", leElement.Op)
		return
	}

	if mustFloat64(leElement.Value) != float64(10000) {
		t.Errorf("test le operator, invalid value: %v", leElement.Value)
		return
	}

	labels := map[string]string{
		"bscp.qps": "9000",
	}

	matched, err := leElement.Match(labels)
	if err != nil {
		t.Errorf("test le operator, match failed, err: %v", err)
		return
	}

	if !matched {
		t.Error("test le operator, but matched")
		return
	}

	labels["bscp.qps"] = "10000"

	matched, err = leElement.Match(labels)
	if err != nil {
		t.Errorf("test le operator, match failed, err: %v", err)
		return
	}

	if !matched {
		t.Error("test le operator, but matched")
		return
	}

}

func TestUnmarshalInElement(t *testing.T) {

	const inJSON = `
	{
		"key": "bscp.modules",
		"op": "in",
		"value": ["sidecar", "controller"]
	}`

	inElement := new(Element)
	if err := json.Unmarshal([]byte(inJSON), inElement); err != nil {
		t.Errorf("test in operator, failed, err: %v", err)
		return
	}

	if inElement.Key != "bscp.modules" {
		t.Errorf("test in operator, invalid key: %v", inElement.Key)
		return
	}

	if inElement.Op != &InOperator {
		t.Errorf("test in operator, invalid op: %v", inElement.Op)
		return
	}

	val := inElement.Value.([]interface{})
	if val[0] != "sidecar" {
		t.Errorf("test in operator, invalid value: %v", inElement.Value)
		return
	}

	if val[1] != "controller" {
		t.Errorf("test in operator, invalid value: %v", inElement.Value)
		return
	}

	labels := map[string]string{
		"bscp.modules": "sidecar",
	}

	matched, err := inElement.Match(labels)
	if err != nil {
		t.Errorf("test in operator, match failed, err: %v", err)
		return
	}

	if !matched {
		t.Error("test in operator, but matched")
		return
	}

}

func TestUnmarshalNotInElement(t *testing.T) {

	const ninJSON = `
	{
		"key": "bscp.modules",
		"op": "nin",
		"value": ["sidecar", "controller"]
	}`

	ninElement := new(Element)
	if err := json.Unmarshal([]byte(ninJSON), ninElement); err != nil {
		t.Errorf("test nin operator, failed, err: %v", err)
		return
	}

	if ninElement.Key != "bscp.modules" {
		t.Errorf("test nin operator, invalid key: %v", ninElement.Key)
		return
	}

	if ninElement.Op != &NotInOperator {
		t.Errorf("test nin operator, invalid op: %v", ninElement.Op)
		return
	}

	val := ninElement.Value.([]interface{})
	if val[0] != "sidecar" {
		t.Errorf("test nin operator, invalid value: %v", ninElement.Value)
		return
	}

	if val[1] != "controller" {
		t.Errorf("test nin operator, invalid value: %v", ninElement.Value)
		return
	}

	labels := map[string]string{
		"bscp.modules": "template",
	}

	matched, err := ninElement.Match(labels)
	if err != nil {
		t.Errorf("test nin operator, match failed, err: %v", err)
		return
	}

	if !matched {
		t.Error("test nin operator, but matched")
		return
	}
}

func TestUnmarshalRegexElement(t *testing.T) {

	const reJSON = `
	{
		"key": "bscp.modules",
		"op": "re",
		"value": "sidecar"
	}`

	reElement := new(Element)
	if err := json.Unmarshal([]byte(reJSON), reElement); err != nil {
		t.Errorf("test re operator, failed, err: %v", err)
		return
	}

	if reElement.Key != "bscp.modules" {
		t.Errorf("test re operator, invalid key: %v", reElement.Key)
		return
	}

	if reElement.Op != &RegexOperator {
		t.Errorf("test re operator, invalid op: %v", reElement.Op)
		return
	}

	val := reElement.Value.(string)
	if val != "sidecar" {
		t.Errorf("test re operator, invalid value: %v", reElement.Value)
		return
	}

	labels := map[string]string{
		"bscp.modules": "template",
	}

	matched, err := reElement.Match(labels)
	if err != nil {
		t.Errorf("test re operator, match failed, err: %v", err)
		return
	}

	if matched {
		t.Error("test re operator, but matched")
		return
	}

	labels = map[string]string{
		"bscp.modules": "sidecar",
	}

	matched, err = reElement.Match(labels)
	if err != nil {
		t.Errorf("test re operator, match failed, err: %v", err)
		return
	}

	if !matched {
		t.Error("test re operator, but not matched")
		return
	}
}

func TestUnmarshalNotRegexElement(t *testing.T) {

	const nreJSON = `
	{
		"key": "bscp.modules",
		"op": "nre",
		"value": "sidecar"
	}`

	nreElement := new(Element)
	if err := json.Unmarshal([]byte(nreJSON), nreElement); err != nil {
		t.Errorf("test nre operator, failed, err: %v", err)
		return
	}

	if nreElement.Key != "bscp.modules" {
		t.Errorf("test nre operator, invalid key: %v", nreElement.Key)
		return
	}

	if nreElement.Op != &NotRegexOperator {
		t.Errorf("test nre operator, invalid op: %v", nreElement.Op)
		return
	}

	val := nreElement.Value.(string)
	if val != "sidecar" {
		t.Errorf("test nre operator, invalid value: %v", nreElement.Value)
		return
	}

	labels := map[string]string{
		"bscp.modules": "template",
	}

	matched, err := nreElement.Match(labels)
	if err != nil {
		t.Errorf("test nre operator, match failed, err: %v", err)
		return
	}

	if !matched {
		t.Error("test nre operator, but not matched")
		return
	}

	labels = map[string]string{
		"bscp.modules": "sidecar",
	}

	matched, err = nreElement.Match(labels)
	if err != nil {
		t.Errorf("test nre operator, match failed, err: %v", err)
		return
	}

	if matched {
		t.Error("test nre operator, but matched")
		return
	}

}

func TestUnmarshalLabelOr(t *testing.T) {

	const labelOrJSON = `
	{
	"match_all": false,
	"labels_or": [
		{
			"key": "version",
			"op": "eq",
			"value": "2.0.0"
		},
		{
			"key": "operator",
			"op": "eq",
			"value": "tom"
		},
		{
			"key": "count",
			"op": "gt",
			"value": 3
		}
	]
}`

	st := new(pbstruct.Struct)
	if err := st.UnmarshalJSON([]byte(labelOrJSON)); err != nil {
		t.Errorf("test labelor strategy, failed, err: %v", err)
		return
	}

	strategy, err := UnmarshalStrategyFromPbStruct(st)
	if err != nil {
		t.Errorf("test labelor strategy, failed, err: %v", err)
		return
	}

	labels := map[string]string{
		"version": "2.0.0",
		"count":   "2",
	}

	matched, err := strategy.MatchLabels(labels)
	if err != nil {
		t.Errorf("test labelor strategy, match failed, err: %v", err)
		return
	}

	if !matched {
		t.Error("test labelor strategy, not matched")
		return
	}
}

func TestUnmarshalLabelAnd(t *testing.T) {

	const labelAndJSON = `
	{
	"match_all": false,
	"labels_and": [
		{
			"key": "version",
			"op": "eq",
			"value": "2.0.0"
		},
		{
			"key": "operator",
			"op": "eq",
			"value": "tom"
		},
		{
			"key": "count",
			"op": "gt",
			"value": 3
		}
	]
}`

	st := new(pbstruct.Struct)
	if err := st.UnmarshalJSON([]byte(labelAndJSON)); err != nil {
		t.Errorf("test labeland strategy, failed, err: %v", err)
		return
	}

	strategy, err := UnmarshalStrategyFromPbStruct(st)
	if err != nil {
		t.Errorf("test labeland strategy, failed, err: %v", err)
		return
	}

	labels := map[string]string{
		"version":  "2.0.0",
		"operator": "tom",
		"count":    "4",
	}

	matched, err := strategy.MatchLabels(labels)
	if err != nil {
		t.Errorf("test labeland strategy, match failed, err: %v", err)
		return
	}

	if !matched {
		t.Error("test labeland strategy, not matched")
		return
	}
}
