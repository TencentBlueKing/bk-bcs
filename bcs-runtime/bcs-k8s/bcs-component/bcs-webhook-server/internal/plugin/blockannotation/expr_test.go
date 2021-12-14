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
