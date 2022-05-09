/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package blockannotation

import (
	"encoding/json"
	"fmt"
	"reflect"
)

const (
	// for string type

	// OperatorStringEqual string a equals string b
	OperatorStringEqual = "str-equals"
	// OperatorStringNotEqual string a is not equals string b
	OperatorStringNotEqual = "str-notequals"

	// for json type

	// OperatorJSONEqual json a has all the kvs in json b
	OperatorJSONEqual = "json-equals"
	// OperatorJSONNotEqual json a does not has all the kvs in json b
	OperatorJSONNotEqual = "json-notequals"

	// FailPolicyAllow allow request when block failed
	FailPolicyAllow = "allow"
	// FailPolicyBlock reject request when compare failed
	FailPolicyBlock = "block"
)

var Operators = []string{OperatorStringEqual, OperatorStringNotEqual, OperatorJSONEqual, OperatorJSONNotEqual}

// BlockUnit expression for blocker
type BlockUnit struct {
	ReferenceContent string
	Operator         string
	FailPolicy       string
}

// NewBlockUnit create block unit
func NewBlockUnit(refer, op, failPolicy string) *BlockUnit {
	return &BlockUnit{
		ReferenceContent: refer,
		Operator:         op,
		FailPolicy:       failPolicy,
	}
}

// IsBlock do compare, result true means that object is blocked, result false means not blocked
func (cu *BlockUnit) IsBlock(toMatch string) bool {
	isBlocked, err := cu.doMatch(toMatch)
	if err != nil {
		if cu.FailPolicy == FailPolicyAllow {
			return false
		}
		return true
	}
	return isBlocked
}

func (cu *BlockUnit) doMatch(toMatch string) (bool, error) {
	switch cu.Operator {
	case OperatorStringEqual:
		return strEqual(cu.ReferenceContent, toMatch), nil
	case OperatorStringNotEqual:
		return !strEqual(cu.ReferenceContent, toMatch), nil
	case OperatorJSONEqual:
		return jsonEqual(cu.ReferenceContent, toMatch)
	case OperatorJSONNotEqual:
		return jsonNotEqual(cu.ReferenceContent, toMatch)
	default:
		return false, fmt.Errorf("unknown operator")
	}
}

func strEqual(op1, op2 string) bool {
	return op1 == op2
}

func jsonEqual(op1, op2 string) (bool, error) {
	obj1 := make(map[string]interface{})
	if err := json.Unmarshal([]byte(op1), &obj1); err != nil {
		return false, err
	}
	obj2 := make(map[string]interface{})
	if err := json.Unmarshal([]byte(op2), &obj2); err != nil {
		return false, err
	}
	for k, v := range obj1 {
		tmpV, ok := obj2[k]
		if !ok {
			return false, nil
		}
		if !reflect.DeepEqual(tmpV, v) {
			return false, nil
		}
	}
	return true, nil
}

func jsonNotEqual(op1, op2 string) (bool, error) {
	ret, err := jsonEqual(op1, op2)
	if err != nil {
		return false, err
	}
	return !ret, nil
}
