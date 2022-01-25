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
	"testing"
)

// TestBlockUnit test block uint function
func TestBlockUnit(t *testing.T) {
	testCases := []struct {
		message    string
		refer      string
		toMatch    string
		op         string
		failPolicy string
		isBlock    bool
	}{
		{
			message:    "string-equal block",
			refer:      "test1",
			toMatch:    "test1",
			op:         OperatorStringEqual,
			failPolicy: FailPolicyBlock,
			isBlock:    true,
		},
		{
			message:    "string-equal allow",
			refer:      "test1",
			toMatch:    "test2",
			op:         OperatorStringEqual,
			failPolicy: FailPolicyBlock,
			isBlock:    false,
		},
		{
			message:    "string-not-equal block",
			refer:      "test1",
			toMatch:    "test2",
			op:         OperatorStringNotEqual,
			failPolicy: FailPolicyBlock,
			isBlock:    true,
		},
		{
			message:    "string-not-equal allow",
			refer:      "test1",
			toMatch:    "test1",
			op:         OperatorStringNotEqual,
			failPolicy: FailPolicyBlock,
			isBlock:    false,
		},
		{
			message:    "json-equal block",
			refer:      "{\"test1\":\"value1\"}",
			toMatch:    "{\"test1\":\"value1\",\"test2\":\"value2\"}",
			op:         OperatorJSONEqual,
			failPolicy: FailPolicyBlock,
			isBlock:    true,
		},
		{
			message:    "json-equal allow",
			refer:      "{\"test1\":\"value11\"}",
			toMatch:    "{\"test1\":\"value1\",\"test2\":\"value2\"}",
			op:         OperatorJSONEqual,
			failPolicy: FailPolicyBlock,
			isBlock:    false,
		},
		{
			message:    "json-equal block",
			refer:      "test1",
			toMatch:    "{\"test1\":\"value1\",\"test2\":\"value2\"}",
			op:         OperatorJSONEqual,
			failPolicy: FailPolicyBlock,
			isBlock:    true,
		},
		{
			message:    "json-equal allow",
			refer:      "test1",
			toMatch:    "{\"test1\":\"value1\",\"test2\":\"value2\"}",
			op:         OperatorJSONEqual,
			failPolicy: FailPolicyAllow,
			isBlock:    false,
		},
		{
			message:    "json-not-equal block",
			refer:      "{\"test1\":\"value11\"}",
			toMatch:    "{\"test1\":\"value1\",\"test2\":\"value2\"}",
			op:         OperatorJSONNotEqual,
			failPolicy: FailPolicyBlock,
			isBlock:    true,
		},
		{
			message:    "json-not-equal allow",
			refer:      "{\"test1\":\"value1\"}",
			toMatch:    "{\"test1\":\"value1\",\"test2\":\"value2\"}",
			op:         OperatorJSONNotEqual,
			failPolicy: FailPolicyBlock,
			isBlock:    false,
		},
	}

	for index, test := range testCases {
		t.Logf("test %d: %v", index, test)
		newBlock := NewBlockUnit(test.refer, test.op, test.failPolicy)
		actualIsBlock := newBlock.IsBlock(test.toMatch)
		if actualIsBlock != test.isBlock {
			t.Errorf("expect %v, but get %v", test.isBlock, actualIsBlock)
			t.Fail()
		}
	}
}
